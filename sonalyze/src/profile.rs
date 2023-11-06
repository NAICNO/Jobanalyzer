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
// TODO: Remove the single-host restriction, for multi-host jobs.  Will complicate everything.

use crate::format;
use crate::{MetaArgs, ProfilePrintArgs};

use anyhow::{bail, Result};
use json;
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

    let fixed_output;
    let mut json_output = false;
    let mut csv_output = false;
    if let Some(ref fmt) = print_args.fmt {
        for fld in fmt.split(',') {
            if fld == "json" {
                json_output = true;
            } else if fld == "csv" {
                csv_output = true;
            }
        }
    }
    if csv_output && json_output {
        bail!("one type of output at a time")
    }
    fixed_output = !json_output && !csv_output;

    // `lists` has the event streams for the processes (or cluster of rolled-up processes).
    //
    // We want these sorted in the order in which they start being shown, so that there is a natural
    // feel to the list of processes for each timestamp.  Sorting ascending by first timestamp will
    // accomplish that.
    let mut lists = entries.drain().map(|(_,v)| v).collect::<Vec<Vec<Box<LogEntry>>>>();
    lists.sort_by(|a,b| a[0].timestamp.cmp(&b[0].timestamp));

    // Print headings: Command name with PID, for now
    if csv_output {
        let mut commands = "time".to_string();
        for x in &lists {
            commands += &format!(",{} ({})", x[0].command, x[0].pid);
        }
        commands += "\n";
        output.write(commands.as_bytes())?;
    }

    // Indices into those streams of the next record we want.
    let mut indices = vec![0; lists.len()];

    // Number of nonempty streams remaining, this is the termination condition.
    let mut nonempty = lists.iter().fold(0, |acc, l| acc + if l.len() > 0 { 1 } else { 0 });
    let initial_nonempty = nonempty;

    // The generated report structures.
    let mut fixed_reports = vec![];
    let mut json_reports = vec![];
    let mut csv_reports = vec![];
    let mut timesteps = 0;

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
        let mut first = true;
        let mut curr_json_report = vec![];
        let mut curr_csv_report = vec![];

        let mut i = 0;
        while i < lists.len() {
            if indices[i] < lists[i].len() {
                let r = &lists[i][indices[i]];
                if r.timestamp == mintime {
                    // The cloning is dumb but we need RC to do better.
                    if fixed_output {
                        if first {
                            fixed_reports.push(ReportLine{ t: Some(mintime), r: r.clone() });
                        } else {
                            fixed_reports.push(ReportLine{ t: None, r: r.clone() });
                        }
                    } else if json_output {
                        curr_json_report.push(r.clone());
                    } else if csv_output {
                        curr_csv_report.push(Some(r.clone()))
                    }
                    indices[i] += 1;
                    if indices[i] == lists[i].len() {
                        nonempty -= 1;
                    }
                    if first {
                        timesteps += 1;
                        first = false;
                    }
                } else if csv_output {
                    curr_csv_report.push(None)
                }
            } else if csv_output {
                curr_csv_report.push(None)
            }
            i += 1;
        }
        if json_output {
            json_reports.push(curr_json_report);
        } else if csv_output {
            csv_reports.push((mintime, curr_csv_report));
        }
    }

    if meta_args.verbose {
        println!("Number of processes: {}", initial_nonempty);
        println!("Any rolled-up processes: {}", has_rolledup);
        println!("Number of time steps: {}", timesteps);
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
    let (fields, others) = format::parse_fields(spec, &formatters, &aliases)?;
    let mut opts = format::standard_options(&others);
    if opts.named  {
        bail!("Named fields are not supported")
    }
    opts.nodefaults = false;

    // The Formatter code does not support nested structures.  For fixed-format output this has been
    // solved with a clever encoding but for JSON and CSV we'll do some crude things.

    if json_output {
        write_json(output, &json_reports)?;
    } else if csv_output {
        write_csv(output, &fields, &csv_reports)?;
    } else {
        opts.fixed = true;
        format::format_data(output, &fields, &formatters, &opts, fixed_reports, &false);
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
    let mut aliases: HashMap<String, Vec<String>> = HashMap::new();

    formatters.insert("time".to_string(), &format_time);
    formatters.insert("cpu".to_string(), &format_cpu);
    formatters.insert("mem".to_string(), &format_mem);
    formatters.insert("gpu".to_string(), &format_gpu);
    formatters.insert("gpumem".to_string(), &format_gpumem);
    formatters.insert("nproc".to_string(), &format_nproc);
    formatters.insert("cmd".to_string(), &format_cmd);

    aliases.insert("all".to_string(),
                   vec!["time".to_string(), "cpu".to_string(), "mem".to_string(),
                        "gpu".to_string(), "gpumem".to_string(), "nproc".to_string(),
                        "cmd".to_string()]);

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

// Each of the inner vectors are commands, coherently sorted, for the same timestamp.  We write all
// fields.

fn write_json(output: &mut dyn io::Write, data: &[Vec<Box<LogEntry>>]) -> Result<()> {
    let mut objects: Vec<json::JsonValue> = vec![];
    for x in data {
        let mut obj = json::JsonValue::new_object();
        obj["time"] = x[0].timestamp.format("%Y-%m-%d %H:%M").to_string().into();
        obj["job"] = x[0].job_id.into();
        let mut points: Vec<json::JsonValue> = vec![];
        for y in x {
            let mut point = json::JsonValue::new_object();
            point["command"] = y.command.to_string().into();
            point["pid"] = y.pid.into();
            point["cpu"] = (y.cpu_util_pct.round() as isize).into();
            point["mem"] = (y.mem_gb.round() as isize).into();
            point["gpu"] = (y.gpu_pct.round() as isize).into();
            point["gpumem"] = (y.gpumem_gb.round() as isize).into();
            point["nproc"] = (y.rolledup + 1).into();
            points.push(point.into())
        }
        obj["points"] = points.into();
        objects.push(obj);
    }
    output.write(json::stringify(objects).as_bytes())?;
    Ok(())
}

fn write_csv(output: &mut dyn io::Write, fields: &[&str], data: &[(Timestamp, Vec<Option<Box<LogEntry>>>)]) -> Result<()> {
    let mut cpu_field = 0;
    let mut mem_field = 0;
    let mut gpu_field = 0;
    let mut gpumem_field = 0;
    for f in fields {
        match *f {
            "cpu" => { cpu_field += 1; }
            "gpu" => { gpu_field += 1; }
            "mem" => { mem_field += 1; }
            "gpumem" => { gpumem_field += 1; }
            "csv" => {}
            _ => { bail!("Not a known field: {f}") }
        }
    }
    if cpu_field + mem_field + gpu_field + gpumem_field != 1 {
        bail!("csv output needs exactly one valid field")
    }

    for (t, xs) in data {
        let mut s = "".to_string();
        s += t.format("%Y-%m-%d %H:%M").to_string().as_str();
        for x in xs {
            s += ",";
            if let Some(x) = x {
                if cpu_field > 0 {
                    s += (x.cpu_util_pct.round() as isize).to_string().as_str();
                } else if mem_field > 0 {
                    s += (x.mem_gb.round() as isize).to_string().as_str();
                } else if gpu_field > 0 {
                    s += (x.gpu_pct.round() as isize).to_string().as_str();
                } else if gpumem_field > 0 {
                    s += (x.gpumem_gb.round() as isize).to_string().as_str();
                } else {
                    panic!("Should not happen")
                }
            }
        }
        s += "\n";
        output.write(s.as_bytes())?;
    }

    Ok(())
}
