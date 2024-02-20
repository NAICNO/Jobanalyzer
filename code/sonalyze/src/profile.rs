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
//   res: the res_gb field
//   gpu: the gpu_pct field
//   gpumem: the gpumem_gb field
//   nproc: the rolledup field + 1, this is not printed if every process has rolledup=0
//   command: the command field
//
// TODO: Remove the single-host restriction, for multi-host jobs.  Will complicate everything.

use crate::format;
use crate::{MetaArgs, ProfileFilterAndAggregationArgs, ProfilePrintArgs};

use anyhow::{bail, Result};
use json;
use sonarlog::{InputStreamSet, LogEntry, Timestamp};
use ustr::Ustr;

use std::cmp::max;
use std::collections::HashMap;
use std::io;

pub fn print(
    output: &mut dyn io::Write,
    jobno: usize,
    filter_args: &ProfileFilterAndAggregationArgs,
    print_args: &ProfilePrintArgs,
    meta_args: &MetaArgs,
    mut entries: InputStreamSet,
) -> Result<()> {
    // Each stream is already filtered and sorted ascending by time without time duplicates.

    // Simplify: Assert that we have only a single host.
    // Precompute: check whether we need the nproc field

    let mut host: Option<&str> = None;
    let mut has_rolledup = false;
    for ((k, _, _), es) in entries.iter() {
        for e in es {
            if e.rolledup > 0 {
                has_rolledup = true;
            }
        }
        if let Some(hn) = host {
            if k.as_str() != hn {
                bail!("profile only implemented for single-host jobs")
            }
        } else {
            host = Some(k.as_str());
        }
    }
    let hostname = host.unwrap_or("unknown").to_string();

    let fixed_output;
    let mut json_output = false;
    let mut csv_output = false;
    let mut html_output = false;
    let mut outputs = 0;
    if let Some(ref fmt) = print_args.fmt {
        for fld in fmt.split(',') {
            if fld == "html" {
                html_output = true;
                outputs += 1;
            }
            if fld == "json" {
                json_output = true;
                outputs += 1;
            } else if fld == "csv" {
                csv_output = true;
                outputs += 1;
            }
        }
    }
    if outputs > 1 {
        bail!("one type of output at a time")
    }
    fixed_output = outputs == 0;

    // The input is a matrix of per-process-per-point-in-time data, with time running down the
    // column, process index running across the row, and where each datum can have one or more
    // measurements of interest for that process at that time (cpu, mem, gpu, gpumem, nproc).  THE
    // MATRIX IS SPARSE, as processes only have data at points in time when they are running, and
    // has a vector-of-columns representation.
    //
    // To make it printable, we make the matrix dense by filling in the missing elements with None
    // markers and we convert it to a vector-of-rows representation, as this is best for fixed, csv,
    // and json output (though it is suboptimal for the current html output code) and anyway is a
    // natural consequence of making it dense.
    //
    // We apply clamping to all the pertinent fields during the matrix conversion step.

    // `processes` has the event streams for the processes (or group of rolled-up processes).
    //
    // We want these sorted in the order in which they start being shown, so that there is a natural
    // feel to the list of processes for each timestamp.  Sorting ascending by first timestamp will
    // accomplish that.
    let mut processes = entries
        .drain()
        .map(|(_, v)| v)
        .collect::<Vec<Vec<Box<LogEntry>>>>();
    processes.sort_by(|a, b| a[0].timestamp.cmp(&b[0].timestamp));

    // Print headings: Command name with PID, for now
    if csv_output {
        let mut commands = "time".to_string();
        for process in &processes {
            commands += &format!(",{} ({})", process[0].command, process[0].pid);
        }
        commands += "\n";
        output.write(commands.as_bytes())?;
    }

    // Indices into those streams of the next record we want.
    let mut indices = vec![0; processes.len()];

    // Number of nonempty streams remaining, this is the termination condition.
    let mut nonempty = processes
        .iter()
        .fold(0, |acc, l| acc + if l.len() > 0 { 1 } else { 0 });
    let initial_nonempty = nonempty;

    // The generated report structures.
    let mut fixed_reports = vec![];
    let mut json_reports = vec![];
    let mut csv_reports = vec![];
    let mut timesteps = 0;

    // Generate the reports.
    //
    // This loop is quadratic-ish but `processes` will tend (modulo non-rolled-up MPI jobs, TBD) to
    // be very short and it's not clear what's to be gained yet by doing something more complicated
    // here like a priority queue, say.
    while nonempty > 0 {
        // Compute current time: the minimum timestamp in the lists that are not exhausted
        let mut i = 0;
        let mut mintime = None;
        while i < processes.len() {
            if indices[i] < processes[i].len() {
                let candidate = processes[i][indices[i]].timestamp;
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
        while i < processes.len() {
            if indices[i] < processes[i].len() {
                let r = &processes[i][indices[i]];
                if r.timestamp == mintime {
                    let newr = clamp_fields(&r, filter_args);
                    if fixed_output {
                        if first {
                            fixed_reports.push(ReportLine {
                                t: Some(mintime),
                                r: newr,
                            });
                        } else {
                            fixed_reports.push(ReportLine { t: None, r: newr });
                        }
                    } else if json_output {
                        curr_json_report.push(newr);
                    } else if csv_output || html_output {
                        curr_csv_report.push(Some(newr))
                    }
                    indices[i] += 1;
                    if indices[i] == processes[i].len() {
                        nonempty -= 1;
                    }
                    if first {
                        timesteps += 1;
                        first = false;
                    }
                } else if csv_output || html_output {
                    curr_csv_report.push(None)
                }
            } else if csv_output || html_output {
                curr_csv_report.push(None)
            }
            i += 1;
        }
        if json_output {
            json_reports.push(curr_json_report);
        } else if csv_output || html_output {
            csv_reports.push((mintime, curr_csv_report));
        }
    }

    // Bucketing will average consecutive records in the clamped record stream within the same
    // process.  We count only present entries in the divisor for the average.  The time value will
    // be the midpoint in the chunk.  As in the case of the HTML printing code, this gets a bit
    // complex due to the vector-of-rows representation of the matrix.

    if let Some(b) = filter_args.bucket {
        let nproc = processes.len();
        if b > 1 && (csv_output || html_output) {
            csv_reports = csv_reports
                .chunks(b)
                .map(|rows| {
                    let new_t = rows[rows.len() / 2].0;
                    let mut recs = vec![];
                    for i in 0..nproc {
                        let mut count = 0;
                        let mut cpu_util_pct = 0.0;
                        let mut mem_gb = 0.0;
                        let mut res_gb = 0.0;
                        let mut gpu_pct = 0.0;
                        let mut gpumem_gb = 0.0;
                        let mut avg = None;
                        for (_, row) in rows {
                            if let Some(ref proc) = row[i] {
                                if avg.is_none() {
                                    avg = Some(proc.clone())
                                }
                                count += 1;
                                cpu_util_pct += proc.cpu_util_pct;
                                mem_gb += proc.mem_gb;
                                res_gb += proc.rssanon_gb;
                                gpu_pct += proc.gpu_pct;
                                gpumem_gb += proc.gpumem_gb;
                            }
                        }
                        if let Some(ref mut avg) = avg {
                            avg.cpu_util_pct = cpu_util_pct / (count as f32);
                            avg.mem_gb = mem_gb / (count as f64);
                            avg.rssanon_gb = res_gb / (count as f32);
                            avg.gpu_pct = gpu_pct / (count as f32);
                            avg.gpumem_gb = gpumem_gb / (count as f64);
                        }
                        recs.push(avg);
                    }
                    (new_t, recs)
                })
                .collect::<Vec<(Timestamp, Vec<Option<Box<LogEntry>>>)>>();
        }
        if b > 1 && json_output {
            bail!("Bucketing not implemented for json output");
        }
        if b > 1 && fixed_output {
            bail!("Bucketing not implemented for fixed output");
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
    if opts.named {
        bail!("Named fields are not supported")
    }
    opts.nodefaults = false;

    // The Formatter code does not support nested structures.  For fixed-format output this has been
    // solved with a clever encoding but for JSON and CSV we'll do some crude things.

    if html_output {
        let bucketing = max(filter_args.bucket.or(Some(1)).unwrap(), 1);
        let html_labels = processes
            .iter()
            .map(|p| format!("{} ({})", p[0].command, p[0].pid))
            .collect::<Vec<String>>();
        write_html(
            output,
            &hostname,
            jobno,
            bucketing,
            &fields,
            &html_labels,
            processes.len(),
            &csv_reports,
        )?;
    } else if json_output {
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
    formatters.insert("res".to_string(), &format_res);
    formatters.insert("gpu".to_string(), &format_gpu);
    formatters.insert("gpumem".to_string(), &format_gpumem);
    formatters.insert("nproc".to_string(), &format_nproc);
    formatters.insert("cmd".to_string(), &format_cmd);

    aliases.insert(
        "all".to_string(),
        vec![
            "time".to_string(),
            "cpu".to_string(),
            "mem".to_string(),
            "res".to_string(),
            "gpu".to_string(),
            "gpumem".to_string(),
            "nproc".to_string(),
            "cmd".to_string(),
        ],
    );

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

fn format_res(d: LogDatum, _: LogCtx) -> String {
    format!("{}", d.r.rssanon_gb.round() as isize)
}

fn format_gpu(d: LogDatum, _: LogCtx) -> String {
    format!("{}", d.r.gpu_pct.round() as isize)
}

fn format_gpumem(d: LogDatum, _: LogCtx) -> String {
    format!("{}", d.r.gpumem_gb.round() as isize)
}

fn format_nproc(d: LogDatum, _: LogCtx) -> String {
    if d.r.rolledup == 0 {
        "".to_string()
    } else {
        format!("{}", d.r.rolledup + 1)
    }
}

fn format_cmd(d: LogDatum, _: LogCtx) -> String {
    d.r.command.to_string()
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
            point["res"] = (y.rssanon_gb.round() as isize).into();
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

fn check_fields(fields: &[&str]) -> Result<(bool, bool, bool, bool, bool)> {
    let mut cpu_field = 0;
    let mut mem_field = 0;
    let mut res_field = 0;
    let mut gpu_field = 0;
    let mut gpumem_field = 0;
    for f in fields {
        match *f {
            "cpu" => {
                cpu_field += 1;
            }
            "gpu" => {
                gpu_field += 1;
            }
            "mem" => {
                mem_field += 1;
            }
            "res" => {
                res_field += 1;
            }
            "gpumem" => {
                gpumem_field += 1;
            }
            "csv" | "html" => {}
            _ => {
                bail!("Not a known field: {f}")
            }
        }
    }
    if cpu_field + mem_field + res_field + gpu_field + gpumem_field != 1 {
        bail!("formatted output needs exactly one valid field")
    }
    Ok((
        cpu_field > 0,
        mem_field > 0,
        res_field > 0,
        gpu_field > 0,
        gpumem_field > 0,
    ))
}

// Each data record has a timestamp and then a vector of length = the number of processes.  The
// values in the record are None (no information - process not running) or the LogEntry for the
// process at that time.

fn write_csv(
    output: &mut dyn io::Write,
    fields: &[&str],
    data: &[(Timestamp, Vec<Option<Box<LogEntry>>>)],
) -> Result<()> {
    let (cpu_field, mem_field, res_field, gpu_field, gpumem_field) = check_fields(fields)?;
    for (t, xs) in data {
        let mut s = "".to_string();
        s += t.format("%Y-%m-%d %H:%M").to_string().as_str();
        for x in xs {
            s += ",";
            if let Some(x) = x {
                if cpu_field {
                    s += (x.cpu_util_pct.round() as isize).to_string().as_str();
                } else if mem_field {
                    s += (x.mem_gb.round() as isize).to_string().as_str();
                } else if res_field {
                    s += (x.rssanon_gb.round() as isize).to_string().as_str();
                } else if gpu_field {
                    s += (x.gpu_pct.round() as isize).to_string().as_str();
                } else if gpumem_field {
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

// Each data record has a timestamp and then a vector of length = the number of processes.  The
// values in the record are None (no information - process not running) or the LogEntry for the
// process at that time.
//
// The logic is a little knotty due to the vector-of-rows representation of the profile matrix.

fn write_html(
    output: &mut dyn io::Write,
    hostname: &str,
    jobno: usize,
    bucketing: usize,
    fields: &[&str],
    col_labels: &[String],
    numprocs: usize,
    data: &[(Timestamp, Vec<Option<Box<LogEntry>>>)],
) -> Result<()> {
    let (cpu_field, mem_field, res_field, gpu_field, gpumem_field) = check_fields(fields)?;
    let labels = data
        .iter()
        .map(|(t, _)| t.format("\"%Y-%m-%d %H:%M\"").to_string())
        .collect::<Vec<String>>()
        .join(",");
    // One dataset per process.
    let mut datasets = vec![];
    let mut username = Ustr::from("");
    for i in 0..numprocs {
        // s has the rendered data for process i.
        let mut s = "".to_string();
        let mut first = true;
        for (_, data) in data.iter() {
            if !first {
                s += ",";
            }
            let val = if let Some(ref x) = data[i] {
                if username.is_empty() {
                    username = x.user;
                }
                if cpu_field {
                    (x.cpu_util_pct.round() as isize).to_string()
                } else if mem_field {
                    (x.mem_gb.round() as isize).to_string()
                } else if res_field {
                    (x.rssanon_gb.round() as isize).to_string()
                } else if gpu_field {
                    (x.gpu_pct.round() as isize).to_string()
                } else if gpumem_field {
                    (x.gpumem_gb.round() as isize).to_string()
                } else {
                    panic!("Should not happen")
                }
            } else {
                "".to_string()
            };
            s += &val;
            first = false;
        }
        let label = &col_labels[i];
        datasets.push(format!("{{label: \"{label}\", data: [{s}]}}"));
    }
    let all_datasets = datasets.join(",");
    let mut text = "".to_string();
    let quant = if cpu_field {
        "cpu"
    } else if gpu_field {
        "gpu"
    } else if mem_field {
        "mem"
    } else if gpumem_field {
        "gpumem"
    } else {
        panic!("Should not happen")
    };
    let mut title = format!("`{quant}` profile of job {jobno} on `{hostname}`, user `{username}`");
    if bucketing > 1 {
        title += format!(", bucketing={bucketing}").as_str();
    }
    text += "
<html>
 <head>
  <title>";
    text += &title;
    text += "</title>
  <script src=\"https://cdn.jsdelivr.net/npm/chart.js\"></script>
  <script>
var LABELS = [";
    text += &labels;
    text += "];
var DATASETS = [";
    text += &all_datasets;
    text += "];
function render() {
  new Chart(document.getElementById(\"chart_node\"), {
    type: 'line',
    data: {
      labels: LABELS,
      datasets: DATASETS
    },
    options: { scales: { x: { beginAtZero: true }, y: { beginAtZero: true } } }
  })
}
  </script>
 </head>
 <body onload=\"render()\">
  <center><h1>";
    text += &title;
    text += "</h1></center>
  <div><canvas id=\"chart_node\"></canvas></div>
 </body>
<html>
";
    output.write(text.as_bytes())?;
    Ok(())
}

// Clamping is a hack but it works.

fn clamp_fields(r: &Box<LogEntry>, filter_args: &ProfileFilterAndAggregationArgs) -> Box<LogEntry> {
    let mut newr = r.clone();
    if filter_args.max.is_some() {
        newr.cpu_util_pct = clamp_max(newr.cpu_util_pct, filter_args.max);
        newr.mem_gb = clamp_max64(newr.mem_gb, filter_args.max);
        newr.rssanon_gb = clamp_max(newr.rssanon_gb, filter_args.max);
        newr.gpu_pct = clamp_max(newr.gpu_pct, filter_args.max);
        newr.gpumem_gb = clamp_max64(newr.gpumem_gb, filter_args.max);
    }
    newr
}

// Max clamping: If the value is greater than the clamp then return the clamp, except if it is more
// than twice the value of the clamp, in which case return 0 - the assumption is that it's a wild
// outlier / noise.

fn clamp_max(x: f32, c: Option<f64>) -> f32 {
    if let Some(d) = c {
        let c = d as f32;
        if x > c {
            if x > 2.0 * c {
                0.0
            } else {
                c
            }
        } else {
            x
        }
    } else {
        x
    }
}

fn clamp_max64(x: f64, c: Option<f64>) -> f64 {
    if let Some(c) = c {
        if x > c {
            if x > 2.0 * c {
                0.0
            } else {
                c
            }
        } else {
            x
        }
    } else {
        x
    }
}
