/// Compute system load aggregates from a set of log entries.
///
/// TODO:
///
/// - For some listings it may be desirable to print a heading?

use crate::format;
use crate::{LoadFilterAndAggregationArgs, LoadPrintArgs, MetaArgs};

use anyhow::{bail, Result};
use sonarlog::{self, add_day, add_hour, empty_logentry, now, truncate_to_day, truncate_to_hour,
               HostFilter, InputStreamSet, LogEntry, Timestamp};
use std::boxed::Box;
use std::collections::HashMap;
use std::io;

#[derive(PartialEq, Clone, Copy)]
enum BucketOpt {
    None,
    Hourly,
    Daily,
}

#[derive(PartialEq, Clone, Copy)]
enum PrintOpt {
    All,
    Last,
}

struct PrintContext<'a> {
    sys: Option<&'a sonarlog::System>,
    t: Timestamp,
}

// We read and filter sonar records, bucket by host, sort by ascending timestamp, and then bucket by
// timestamp.  The buckets can then be aggregated into a "load" value for each time, which can in
// turn be averaged for a span of times.

pub fn aggregate_and_print_load(
    output: &mut dyn io::Write,
    system_config: &Option<HashMap<String, sonarlog::System>>,
    _include_hosts: &HostFilter,
    from: Timestamp,
    to: Timestamp,
    filter_args: &LoadFilterAndAggregationArgs,
    print_args: &LoadPrintArgs,
    meta_args: &MetaArgs,
    streams: InputStreamSet,
) -> Result<()> {

    if meta_args.verbose {
        return Ok(());
    }

    let bucket_opt = if filter_args.daily {
        BucketOpt::Daily
    } else if filter_args.none {
        BucketOpt::None
    } else {
        BucketOpt::Hourly // Default
    };

    let print_opt = if print_args.last {
        PrintOpt::Last
    } else {
        PrintOpt::All // Default
    };

    let mut formatters: HashMap<String, &dyn Fn(&Box<LogEntry>, LoadCtx) -> String> = HashMap::new();
    formatters.insert("datetime".to_string(), &format_datetime);
    formatters.insert("date".to_string(), &format_date);
    formatters.insert("time".to_string(), &format_time);
    formatters.insert("cpu".to_string(), &format_cpu);
    formatters.insert("rcpu".to_string(), &format_rcpu);
    formatters.insert("mem".to_string(), &format_mem);
    formatters.insert("rmem".to_string(), &format_rmem);
    formatters.insert("gpu".to_string(), &format_gpu);
    formatters.insert("rgpu".to_string(), &format_rgpu);
    formatters.insert("gpumem".to_string(), &format_gpumem);
    formatters.insert("rgpumem".to_string(), &format_rgpumem);
    formatters.insert("gpus".to_string(), &format_gpus);
    formatters.insert("now".to_string(), &format_now);
    formatters.insert("host".to_string(), &format_host);

    let spec = if let Some(ref fmt) = print_args.fmt {
        fmt
    } else {
        "date,time,cpu,mem,gpu,gpumem,gpumask"
    };
    let aliases = HashMap::new();
    let (fields, others) = format::parse_fields(spec, &formatters, &aliases);
    let opts = format::standard_options(&others);
    let relative = fields.iter().any(|x| match *x {
        "rcpu" | "rmem" | "rgpu" | "rgpumem" => true,
        _ => false,
    });

    if relative && system_config.is_none() {
        bail!("Relative values requested without config file");
    }

    // There one synthesized sample stream per host.  The samples will all have different
    // timestamps, and each stream will be sorted ascending by timestamp.

    let mut merged_streams = sonarlog::merge_by_host(streams);

    // Sort hosts lexicographically.  This is not ideal because hosts like c1-10 vs c1-5 are not in
    // the order we expect but at least it's predictable.

    merged_streams.sort_by(|a, b| a[0].hostname.cmp(&b[0].hostname));

    let explicit_host = fields.iter().any(|x| *x == "host");
    for stream in merged_streams {
        let hostname = stream[0].hostname.clone();
        if !opts.csv && !explicit_host {
            output
                .write(format!("HOST: {}\n", hostname).as_bytes())
                .unwrap();
        }

        let sysconf = if let Some(ref ht) = system_config {
            ht.get(&hostname)
        } else {
            None
        };

        let ctx = PrintContext {
            sys: sysconf,
            t: now()
        };

        if bucket_opt != BucketOpt::None {
            let by_timeslot =
                if bucket_opt == BucketOpt::Hourly {
                    sonarlog::fold_samples_hourly(stream)
                } else {
                    sonarlog::fold_samples_daily(stream)
                };
            if print_opt == PrintOpt::All {
                let by_timeslot2 =
                    if print_args.compact {
                        by_timeslot
                    } else {
                        insert_missing_records(by_timeslot, from, to, bucket_opt)
                    };
                format::format_data(output, &fields, &formatters, &opts, by_timeslot2, &ctx);
            } else {
                // Invariant: there's always at least one record
                // TODO: Really not happy about the clone() here
                let data = vec![by_timeslot[by_timeslot.len() - 1].clone()];
                format::format_data(output, &fields, &formatters, &opts, data, &ctx);
            }
        } else if print_opt == PrintOpt::All {
            // TODO: A question here about whether we should be inserting zero records.  I'm
            // inclined to think probably not but it's debatable.
            format::format_data(output, &fields, &formatters, &opts, stream, &ctx);
        } else {
            // Invariant: there's always at least one record
            // TODO: Really not happy about the clone() here
            let data = vec![stream[stream.len() - 1].clone()];
            format::format_data(output, &fields, &formatters, &opts, data, &ctx);
        }
    }

    Ok(())
}

fn insert_missing_records(
    mut records: Vec<Box<LogEntry>>,
    from: Timestamp,
    to: Timestamp,
    bucket_opt: BucketOpt,
) -> Vec<Box<LogEntry>> {
    let (trunc, step) : (fn(Timestamp) -> Timestamp, fn(Timestamp)->Timestamp) =
        if bucket_opt == BucketOpt::Hourly {
            (truncate_to_hour, add_hour)
        } else {
            (truncate_to_day, add_day)
        };
    let host = records[0].hostname.clone();
    let mut t = trunc(from);
    let mut result = vec![];

    for r in records.drain(0..) {
        while t < r.timestamp {
            result.push(empty_logentry(t, &host));
            t = step(t);
        }
        result.push(r);
        t = step(t);
    }
    let ending = trunc(to);
    while t <= ending {
        result.push(empty_logentry(t, &host));
        t = step(t);
    }
    result
}

type LoadDatum<'a> = &'a Box<LogEntry>;
type LoadCtx<'a> = &'a PrintContext<'a>;

// An argument could be made that this should be ISO time, at least when the output is CSV, but
// for the time being I'm keeping it compatible with `date` and `time`.
fn format_now(_: LoadDatum, ctx: LoadCtx) -> String {
    ctx.t.format("%Y-%m-%d %H:%M").to_string()
}

fn format_datetime(d: LoadDatum, _: LoadCtx) -> String {
    d.timestamp.format("%Y-%m-%d %H:%M").to_string()
}

fn format_date(d: LoadDatum, _: LoadCtx) -> String {
    d.timestamp.format("%Y-%m-%d").to_string()
}

fn format_time(d: LoadDatum, _: LoadCtx) -> String {
    d.timestamp.format("%H:%M").to_string()
}

fn format_cpu(d: LoadDatum, _: LoadCtx) -> String {
    format!("{}", d.cpu_util_pct as isize)
}

fn format_rcpu(d: LoadDatum, ctx: LoadCtx) -> String {
    let s = ctx.sys.unwrap();
    format!("{}", ((d.cpu_util_pct as f64) / (s.cpu_cores as f64)).round())
}

fn format_mem(d: LoadDatum, _: LoadCtx) -> String {
    format!("{}", d.mem_gb as isize)
}

fn format_rmem(d: LoadDatum, ctx: LoadCtx) -> String {
    let s = ctx.sys.unwrap();
    format!("{}", ((d.mem_gb as f64) / (s.mem_gb as f64) * 100.0).round())
}

fn format_gpu(d: LoadDatum, _: LoadCtx) -> String {
    format!("{}", d.gpu_pct as isize)
}

fn format_rgpu(d: LoadDatum, ctx: LoadCtx) -> String {
    let s = ctx.sys.unwrap();
    format!("{}", ((d.gpu_pct as f64) / (s.gpu_cards as f64)).round())
}

fn format_gpumem(d: LoadDatum, _: LoadCtx) -> String {
    format!("{}", d.gpumem_gb as isize)
}

fn format_rgpumem(d: LoadDatum, ctx: LoadCtx) -> String {
    let s = ctx.sys.unwrap();
    format!("{}", ((d.gpumem_gb as f64) / (s.gpumem_gb as f64) * 100.0).round())
}

fn format_gpus(d: LoadDatum, _: LoadCtx) -> String {
    if let Some(ref gpus) = d.gpus {
        if gpus.is_empty() {
            "none".to_string()
        } else {
            let mut gpunums = vec![];
            for x in gpus {
                gpunums.push(*x);
            }
            gpunums.sort();
            let mut s = "".to_string();
            for x in gpunums {
                if !s.is_empty() {
                    s += ",";
                }
                s += &format!("{}", x)
            }
            s
        }
    } else {
        "unknown".to_string()
    }
}

fn format_host(d: LoadDatum, _: LoadCtx) -> String {
    format!("{}", d.hostname)
}
