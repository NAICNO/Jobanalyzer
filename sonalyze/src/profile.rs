// For one particular job, break it down in its component processes and print
// individual stats for each process for each time slot.
//
// The data fields are raw data items from LogEntry records.  There are no peaks because we only
// have samples at the start and beginning of the time slot.  There is a straightforward
// derivation from the LogEntry values to relative values should we need that (not currently
// implemented).
//
//   cpu: the cpu_util_pct field
//   mem: the mem_gb field
//   gpu: the gpu_pct field
//   gpumem: the gpumem_gb field
//   nproc: the rolledup field + 1, this is not printed if every process has rolledup=0
//   command: the command field
//
// TODO: Remove the single-host restriction.
// TODO: Support at least JSON output, it is sensible and a front-end could make use of it.

use crate::format;
use crate::{MetaArgs, ProfilePrintArgs};

use anyhow::{bail, Result};
use sonarlog::{InputStreamSet, LogEntry, Timestamp};

use std::collections::HashMap;
use std::io;

pub fn print(
    output: &mut dyn io::Write,
    print_args: &ProfilePrintArgs,
    meta_args: &MetaArgs,
    mut entries: InputStreamSet,
) -> Result<()> {
    // Each stream is already filtered and sorted ascending by time without time duplicates.

    // Simplify: Assert that we have only a single host.
    // Precompute: check whether we need the nproc field

    let mut h : Option<&str> = None;
    let mut has_rolledup = false;
    for ((k, _, _), es) in entries.iter() {
        for e in es {
            if e.rolledup > 0 {
                has_rolledup = true;
            }
        }
        if let Some(hn) = h {
            if k.as_str() != hn {
                bail!("profile only implemented for single-host jobs")
            }
        } else {
            h = Some(k.as_str());
        }
    }

    // `lists` has the event streams for the processes (or cluster of rolled-up processes).
    //
    // We want these sorted in the order in which they start being shown, so that there is a natural
    // feel to the list of processes for each timestamp.  Sorting ascending by first timestamp will
    // accomplish that.
    let mut lists = entries.drain().map(|(_,v)| v).collect::<Vec<Vec<Box<LogEntry>>>>();
    lists.sort_by(|a,b| a[0].timestamp.cmp(&b[0].timestamp));

    // Indices into those streams of the next record we want.
    let mut indices = vec![0; lists.len()];

    // Number of nonempty streams remaining, this is the termination condition.
    let mut nonempty = lists.iter().fold(0, |acc, l| acc + if l.len() > 0 { 1 } else { 0 });
    let initial_nonempty = nonempty;

    // The generated report structures.
    let mut reports = vec![];

    // Generate the reports.
    //
    // This loop is quadratic-ish but `lists` will tend (modulo non-rolled-up MPI jobs, TBD) to be
    // very short and it's not clear what's to be gained yet by doing something more complicated
    // here like a priority queue, say.
    while nonempty > 0 {
        // Compute current time: the minimum timestamp in the lists that are not exhausted
        let mut i = 0;
        let mut mintime = None;
        while i < lists.len() {
            if indices[i] < lists[i].len() {
                let candidate = lists[i][indices[i]].timestamp;
                if mintime == None || mintime.unwrap() > candidate {
                    mintime = Some(candidate);
                }
            }
            i += 1;
        }
        assert!(mintime.is_some());

        // The report is the current timestamp + a list of all LogEntries with that timestamp
        let mintime = mintime.unwrap();
        let mut pushed = false;

        i = 0;
        while i < lists.len() {
            if indices[i] < lists[i].len() {
                let r = lists[i][indices[i]].clone();
                if r.timestamp == mintime {
                    if pushed {
                        reports.push(ReportLine{ t: None, r });
                    } else {
                        reports.push(ReportLine{ t: Some(mintime), r });
                        pushed = true;
                    }
                    indices[i] += 1;
                    if indices[i] == lists[i].len() {
                        nonempty -= 1;
                    }
                }
            }
            i += 1;
        }
    }

    if meta_args.verbose {
        println!("Number of processes: {}", initial_nonempty);
        println!("Any rolled-up processes: {}", has_rolledup);
        println!("Number of time steps: {}", reports.len());
        return Ok(());
    }

    let (formatters, aliases) = my_formatters();
    let spec = if let Some(ref fmt) = print_args.fmt {
        fmt
    } else if has_rolledup {
        FMT_DEFAULTS_WITH_NPROC
    } else {
        FMT_DEFAULTS_WITHOUT_NPROC
    };
    let (fields, others) = format::parse_fields(spec, &formatters, &aliases);
    let mut opts = format::standard_options(&others);
    if opts.csv || opts.named {
        bail!("CSV output not supported");
    }
    if opts.json {
        bail!("JSON output not supported yet");
    }
    opts.fixed = true;
    opts.nodefaults = false;
    if fields.len() > 0 {
        format::format_data(output, &fields, &formatters, &opts, reports, &false);
    }

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
        defaults: FMT_DEFAULTS_WITH_NPROC.to_string(),
    }
}

const FMT_DEFAULTS_WITH_NPROC: &str = "time,cpu,mem,gpu,gpumem,nproc,cmd";
const FMT_DEFAULTS_WITHOUT_NPROC: &str = "time,cpu,mem,gpu,gpumem,cmd";

fn my_formatters() -> (
    HashMap<String, &'static dyn Fn(LogDatum, LogCtx) -> String>,
    HashMap<String, Vec<String>>,
) {
    let mut formatters: HashMap<String, &dyn Fn(LogDatum, LogCtx) -> String> = HashMap::new();
    let aliases: HashMap<String, Vec<String>> = HashMap::new();

    formatters.insert("time".to_string(), &format_time);
    formatters.insert("cpu".to_string(), &format_cpu);
    formatters.insert("mem".to_string(), &format_mem);
    formatters.insert("gpu".to_string(), &format_gpu);
    formatters.insert("gpumem".to_string(), &format_gpumem);
    formatters.insert("nproc".to_string(), &format_nproc);
    formatters.insert("cmd".to_string(), &format_cmd);

    (formatters, aliases)
}

struct ReportLine {
    t: Option<Timestamp>,
    r: Box<LogEntry>,
}

type LogDatum<'a> = &'a ReportLine;
type LogCtx<'a> = &'a bool;

fn format_time(d: LogDatum, _: LogCtx) -> String {
    if let Some(t) = d.t {
        t.format("%Y-%m-%d %H:%M").to_string()
    } else {
        "                ".to_string() // YYYY-MM-DD HH:MM
    }
}

fn format_cpu(d: LogDatum, _: LogCtx) -> String {
    format!("{}", d.r.cpu_util_pct.round() as isize)
}

fn format_mem(d: LogDatum, _: LogCtx) -> String {
    format!("{}", d.r.mem_gb.round() as isize)
}

fn format_gpu(d: LogDatum, _: LogCtx) -> String {
    format!("{}", d.r.gpu_pct.round() as isize)
}

fn format_gpumem(d: LogDatum, _: LogCtx) -> String {
    format!("{}", d.r.gpumem_gb.round() as isize)
}

fn format_nproc(d: LogDatum, _: LogCtx) -> String {
    if d.r.rolledup == 0 { "".to_string() } else { format!("{}", d.r.rolledup + 1) }
}

fn format_cmd(d: LogDatum, _: LogCtx) -> String {
    d.r.command.clone()
}
