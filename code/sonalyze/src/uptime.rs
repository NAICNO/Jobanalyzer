// Compute uptime for a host or a host's GPUs.
//
// Given a list of LogEntries, including heartbeat records, the uptime for each host can be computed
// by looking at gaps in the timeline of observations for the host.  If a gap exceeds the threshold
// for the gap, we assume the system was down.
//
// Given a stretch of time - a set of LogEntries - when a host was up, the status of its GPUs can be
// determined by looking at the records' gpu_status fields.
//
// In addition to the LogEntries, we take as inputs the `from` and `to` timestamps defining the time
// window of interest.  A host is up at the start if its first LogEntry is within the gap-threshold
// of the `from` time, and ditto it is up at the end for its last LogEntry close to the `to` time.
// The gap-threshold is computed from the sampling interval provided as an argument to the program.
//
// The CSV form has five fields: device, host, state, start, end where
//
//  - device is `host` or `gpu`
//  - host is the name of the host (FQDN probably)
//  - state is `up` or `down`
//  - start is the inclusive start of the window when the device was in the given state, on the form
//    YYYY-MM-DD HH:MM (the same form used elsewhere)
//  - end is the exclusive end of the window, ditto
//
// `start` and `end` of hosts are computed so that windows overlap: the `end` of one record will
// equal the `start` of the next.  This is fine, and helps clients display the data.  `start` and
// `end` of gpus similarly form a complete timeline within the time that its host is up.
//
// For csvnamed the field names are as given above, and all the values are strings.
//
// Outputs are sorted by host name and then increasing time of the `start` field.  This means the
// report can be read top-to-bottom to get a chronological sense for the history of each host.
//
// TODO:
//
// - TODO: For nodes/hosts that don't have GPUs it would be nice not to print any GPU information.
//   We should be able to use the config data to drive that.
//
// - (Speculative) At the moment, the gpu_status is per-host, not per-card, because that's all sonar
//   is able to discern.  When that changes, the device field will be generalized so that its value
//   may be `gpu0`, `gpu1`, etc.  Most likely records for these will be in addition to the records
//   for plain `gpu`, which will plausibly retain its existing semantics.
//
// - (Speculative) As gpu_status is an enum it can take on other values than up or down; thus when
//   we improve state detection, the representation here of that value may change, or there may be
//   additional fields.

use crate::format;
use crate::{HostGlobber, MetaArgs, UptimePrintArgs};

use anyhow::Result;
use sonarlog::{GpuStatus, LogEntry, Timestamp};
use std::cmp::min;
use std::collections::HashMap;
use std::io;

struct Report {
    device: String,
    host: String,
    state: String,
    start: String,
    end: String,
}

fn new_report(device: &str, host: &str, state: &str, from: Timestamp, to: Timestamp) -> Report {
    Report {
        device: device.to_string(),
        host: host.to_string(),
        state: state.to_string(),
        start: from.format("%Y-%m-%d %H:%M").to_string(),
        end: to.format("%Y-%m-%d %H:%M").to_string(),
    }
}

fn delta_t_le(prev: Timestamp, next: Timestamp, duration: i64) -> bool {
    (next - prev).num_seconds() <= duration
}

pub fn aggregate_and_print_uptime(
    output: &mut dyn io::Write,
    include_hosts: &HostGlobber,
    from_incl: Timestamp,
    to_excl: Timestamp,
    print_args: &UptimePrintArgs,
    meta_args: &MetaArgs,
    mut entries: Vec<Box<LogEntry>>,
) -> Result<()> {
    // We work by sorting by hostname and timestamp and then scanning the sorted list
    entries.sort_by(|a, b| {
        if a.hostname != b.hostname {
            a.hostname.cmp(&b.hostname)
        } else {
            a.timestamp.cmp(&b.timestamp)
        }
    });

    let lim = entries.len();
    let cutoff = print_args.interval as i64 * 60 * 2;
    let mut reports = vec![];
    let mut i = 0;
    let mut host_up_windows = vec![];
    while i < lim {
        // Skip anything for the host before the window we're interested in.

        while i < lim && entries[i].timestamp < from_incl {
            i += 1;
        }
        if i == lim {
            break;
        }
        let host_start = i;
        let mut host_end = i;

        let host_first = &entries[host_start];

        // Scan along the list of entries while we have the same host, and update the index of
        // the last relevant record in the window as we do so.

        i += 1;
        while i < lim && entries[i].hostname == host_first.hostname {
            if entries[i].timestamp < to_excl {
                host_end = i;
            }
            i += 1;
        }

        if !include_hosts.is_empty() && !include_hosts.match_hostname(&host_first.hostname) {
            continue;
        }

        if meta_args.verbose {
            println!(
                "{}: {host_start}..{host_end} inclusive, i={i}",
                host_first.hostname
            );
        }

        // If the host is down at the start, push out a record saying so.  Then we start in the "up"
        // state always.
        if !delta_t_le(from_incl, host_first.timestamp, cutoff) {
            if meta_args.verbose {
                println!("  Down at start");
            }
            if !print_args.only_up {
                reports.push(new_report(
                    "host",
                    &host_first.hostname,
                    "down",
                    from_incl,
                    host_first.timestamp,
                ));
            }
        }

        let host_last = &entries[host_end];

        // If the host is down at the end, push out a record saying so. Note this is out of order;
        // we'll re-sort later anyway.
        if !delta_t_le(host_last.timestamp, to_excl, cutoff) {
            if meta_args.verbose {
                println!(
                    "  Down at end: {} {} {}",
                    host_last.timestamp, to_excl, cutoff
                );
            }
            if !print_args.only_up {
                reports.push(new_report(
                    "host",
                    &host_first.hostname,
                    "down",
                    host_last.timestamp,
                    to_excl,
                ));
            }
        }

        // Within the relevant window of the host's entries, we need to figure out when it might
        // have been down and push out up/down records.  It will be up at the beginning and end,
        // we've ensured that above.

        let mut window_start = host_start;
        loop {
            let mut prev_timestamp = entries[window_start].timestamp;
            let mut j = window_start + 1;

            // We're in an "up" window, scan to its end.
            while j <= host_end && delta_t_le(prev_timestamp, entries[j].timestamp, cutoff) {
                prev_timestamp = entries[j].timestamp;
                j += 1;
            }

            // Now j points past the last record in the up window.  There's a chance here that start
            // and end are the same value (only one sample between two "down" windows); nothing to
            // be done about that.
            if meta_args.verbose {
                println!("  Up window {window_start}..{} inclusive", j - 1);
            }
            if !print_args.only_down {
                reports.push(new_report(
                    "host",
                    &host_first.hostname,
                    "up",
                    entries[window_start].timestamp,
                    entries[j - 1].timestamp,
                ));
            }

            // Record this window, we'll need it for the GPU scans later.  (The scans could happen
            // here, but it just makes the code unreadable.)
            host_up_windows.push((window_start, j - 1));

            if j > host_end {
                break;
            }

            // System went down in the window.  The window in which it is down is entirely between
            // these two records.  The fact that there is a following record means it came up again.
            if meta_args.verbose {
                println!("  Down window {}..{} inclusive", j - 1, j);
            }
            if !print_args.only_up {
                reports.push(new_report(
                    "host",
                    &host_first.hostname,
                    "down",
                    prev_timestamp,
                    entries[j].timestamp,
                ));
            }

            window_start = j;
        }
    }

    // Now for each host "up" window, figure out the GPU status within that window.
    for (host_start, host_end) in host_up_windows {
        let mut i = host_start;
        while i <= host_end {
            let gpu_is_up = entries[i].gpu_status == GpuStatus::Ok;
            let start = i;
            while i <= host_end && (entries[i].gpu_status == GpuStatus::Ok) == gpu_is_up {
                i += 1;
            }
            let updown = if gpu_is_up { "up" } else { "down" };
            if !(updown == "up" && print_args.only_down)
                && !(updown == "down" && print_args.only_up)
            {
                reports.push(new_report(
                    "gpu",
                    &entries[host_start].hostname,
                    updown,
                    entries[start].timestamp,
                    entries[min(host_end, i)].timestamp,
                ));
            }
        }
    }

    // Sort the reports by hostname and start time.
    reports.sort_by(|a, b| {
        if a.host != b.host {
            a.host.cmp(&b.host)
        } else {
            a.start.cmp(&b.start)
        }
    });

    // And print.
    let (formatters, aliases) = my_formatters();
    let spec = if let Some(ref fmt) = print_args.fmt {
        fmt
    } else {
        FMT_DEFAULTS
    };
    let (fields, others) = format::parse_fields(spec, &formatters, &aliases)?;
    let opts = format::standard_options(&others);
    format::format_data(output, &fields, &formatters, &opts, reports, &false);

    Ok(())
}

pub fn fmt_help() -> format::Help {
    let (formatters, aliases) = my_formatters();
    format::Help {
        fields: formatters
            .iter()
            .map(|(k, _)| k.clone())
            .collect::<Vec<String>>(),
        aliases: aliases
            .iter()
            .map(|(k, v)| (k.clone(), v.clone()))
            .collect::<Vec<(String, Vec<String>)>>(),
        defaults: FMT_DEFAULTS.to_string(),
    }
}

const FMT_DEFAULTS: &str = "device,host,state,start,end";

fn my_formatters() -> (
    HashMap<String, &'static dyn Fn(LogDatum, LogCtx) -> String>,
    HashMap<String, Vec<String>>,
) {
    let mut formatters: HashMap<String, &dyn Fn(LogDatum, LogCtx) -> String> = HashMap::new();
    let mut aliases: HashMap<String, Vec<String>> = HashMap::new();
    formatters.insert("device".to_string(), &format_device);
    formatters.insert("host".to_string(), &format_host);
    formatters.insert("state".to_string(), &format_state);
    formatters.insert("start".to_string(), &format_start);
    formatters.insert("end".to_string(), &format_end);
    aliases.insert(
        "all".to_string(),
        vec![
            "device".to_string(),
            "host".to_string(),
            "state".to_string(),
            "start".to_string(),
            "end".to_string(),
        ],
    );

    (formatters, aliases)
}

type LogDatum<'a> = &'a Report;
type LogCtx<'a> = &'a bool;

fn format_device(r: LogDatum, _: LogCtx) -> String {
    r.device.clone()
}

fn format_host(r: LogDatum, _: LogCtx) -> String {
    r.host.clone()
}

fn format_state(r: LogDatum, _: LogCtx) -> String {
    r.state.clone()
}

fn format_start(r: LogDatum, _: LogCtx) -> String {
    r.start.clone()
}

fn format_end(r: LogDatum, _: LogCtx) -> String {
    r.end.clone()
}
