// Jobs printer
//
// Feature: One could imagine other sort orders for the output than least-recently-started-first.
// This only matters for the --numjobs switch.

use crate::configs;
use crate::format;
use crate::{JobPrintArgs, MetaArgs};
use crate::jobs::{JobAggregate, LIVE_AT_START, LIVE_AT_END};

use anyhow::{bail,Result};
use sonarlog::{self, now, LogEntry};
use std::boxed::Box;
use std::collections::{HashMap, HashSet};
use std::io;
use std::ops::Add;

pub fn print_jobs(
    output: &mut dyn io::Write,
    system_config: &Option<HashMap<String, configs::System>>,
    mut jobvec: Vec<(JobAggregate, Vec<Box<LogEntry>>)>,
    print_args: &JobPrintArgs,
    meta_args: &MetaArgs,
) -> Result<()> {
    // And sort ascending by lowest beginning timestamp, and if those are equal (which happens when
    // we start reading logs at some arbitrary date), by job number.
    jobvec.sort_by(|a, b| {
        if a.0.first == b.0.first {
            a.1[0].job_id.cmp(&b.1[0].job_id)
        } else {
            a.0.first.cmp(&b.0.first)
        }
    });

    // Select a number of jobs per user, if applicable.  This means working from the bottom up
    // in the vector and marking the n first per user.  We need a hashmap user -> count.

    if let Some(n) = print_args.numjobs {
        let mut counts: HashMap<&str, usize> = HashMap::new();
        jobvec.iter_mut().rev().for_each(|(aggregate, job)| {
            if let Some(c) = counts.get(&(*job[0].user)) {
                if *c < n {
                    counts.insert(&job[0].user, *c + 1);
                } else {
                    aggregate.selected = false;
                }
            } else {
                counts.insert(&job[0].user, 1);
            }
        })
    }

    let numselected = jobvec
        .iter()
        .map(
            |(aggregate, _)| {
                if aggregate.selected {
                    1i32
                } else {
                    0i32
                }
            },
        )
        .reduce(i32::add)
        .unwrap_or(0);
    if meta_args.verbose {
        eprintln!("Number of jobs after output filtering: {}", numselected);
    }

    // Now print.

    if meta_args.verbose {
        return Ok(());
    }

    // Unix user names are max 8 chars.
    // Linux pids are max 7 decimal digits.
    // We don't care about seconds in the timestamp, nor timezone.

    if meta_args.raw {
        jobvec.iter().for_each(|(aggregate, job)| {
            output
                .write(
                    format!(
                        "{} job records\n\n{:?}\n\n{:?}\n",
                        job.len(),
                        &job[0..std::cmp::min(5, job.len())],
                        aggregate
                    )
                    .as_bytes(),
                )
                .unwrap();
        });
    } else if numselected > 0 {
        // TODO: For multi-host jobs, we probably want the option of printing many of these data
        // per-host and not summed across all hosts necessarily.  I can imagine a keyword that
        // controls this, `per-host` say.

        let mut formatters: HashMap<String, &dyn Fn(LogDatum, LogCtx) -> String> = HashMap::new();
        formatters.insert("jobm".to_string(), &format_jobm_id);
        formatters.insert("job".to_string(), &format_job_id);
        formatters.insert("user".to_string(), &format_user);
        formatters.insert("duration".to_string(), &format_duration);
        formatters.insert("start".to_string(), &format_start);
        formatters.insert("end".to_string(), &format_end);
        formatters.insert("cpu-avg".to_string(), &format_cpu_avg);
        formatters.insert("cpu-peak".to_string(), &format_cpu_peak);
        formatters.insert("rcpu-avg".to_string(), &format_rcpu_avg);
        formatters.insert("rcpu-peak".to_string(), &format_rcpu_peak);
        formatters.insert("mem-avg".to_string(), &format_mem_avg);
        formatters.insert("mem-peak".to_string(), &format_mem_peak);
        formatters.insert("rmem-avg".to_string(), &format_rmem_avg);
        formatters.insert("rmem-peak".to_string(), &format_rmem_peak);
        formatters.insert("gpu-avg".to_string(), &format_gpu_avg);
        formatters.insert("gpu-peak".to_string(), &format_gpu_peak);
        formatters.insert("rgpu-avg".to_string(), &format_rgpu_avg);
        formatters.insert("rgpu-peak".to_string(), &format_rgpu_peak);
        formatters.insert("gpumem-avg".to_string(), &format_gpumem_avg);
        formatters.insert("gpumem-peak".to_string(), &format_gpumem_peak);
        formatters.insert("rgpumem-avg".to_string(), &format_rgpumem_avg);
        formatters.insert("rgpumem-peak".to_string(), &format_rgpumem_peak);
        formatters.insert("gpus".to_string(), &format_gpus);
        formatters.insert("cmd".to_string(), &format_command);
        formatters.insert("host".to_string(), &format_host);
        formatters.insert("now".to_string(), &format_now);

        let mut aliases: HashMap<String, Vec<String>> = HashMap::new();
        aliases.insert("std".to_string(), vec!["jobm".to_string(), "user".to_string(), "duration".to_string(), "host".to_string()]);
        aliases.insert("cpu".to_string(), vec!["cpu-avg".to_string(), "cpu-peak".to_string()]);
        aliases.insert("rcpu".to_string(), vec!["rcpu-avg".to_string(), "rcpu-peak".to_string()]);
        aliases.insert("mem".to_string(), vec!["mem-avg".to_string(), "mem-peak".to_string()]);
        aliases.insert("rmem".to_string(), vec!["rmem-avg".to_string(), "rmem-peak".to_string()]);
        aliases.insert("gpu".to_string(), vec!["gpu-avg".to_string(), "gpu-peak".to_string()]);
        aliases.insert("rgpu".to_string(), vec!["rgpu-avg".to_string(), "rgpu-peak".to_string()]);
        aliases.insert("gpumem".to_string(), vec!["gpumem-avg".to_string(), "gpumem-peak".to_string()]);
        aliases.insert("rgpumem".to_string(), vec!["rgpumem-avg".to_string(), "rgpumem-peak".to_string()]);

        let spec = if let Some(ref fmt) = print_args.fmt {
            fmt
        } else {
            "std,cpu,mem,gpu,gpumem,cmd"
        };
        let (fields, others) = format::parse_fields(spec, &formatters, &aliases);
        let opts = format::standard_options(&others);
        let relative = fields.iter().any(|x| match *x {
            "rcpu-avg" | "rcpu-peak" | "rmem-avg" | "rmem-peak" | "rgpu-avg" | "rgpu-peak" | "rgpumem-avg" | "rgpumem-peak" => true,
            _ => false,
        });
        if relative && system_config.is_none() {
            bail!("Relative values requested without config file");
        }

        if fields.len() > 0 {
            let selected = jobvec
                .drain(0..)
                .filter(|(aggregate, _)| aggregate.selected)
                .collect::<Vec<(JobAggregate, Vec<Box<LogEntry>>)>>();
            format::format_data(output, &fields, &formatters, &opts, selected, false);
        }
    }

    Ok(())
}

type LogDatum<'a> = &'a (JobAggregate, Vec<Box<LogEntry>>);
type LogCtx = bool; // Not used

fn format_user((_, job): LogDatum, _: LogCtx) -> String {
    job[0].user.clone()
}

fn format_jobm_id((a, job): LogDatum, _: LogCtx) -> String {
    format!(
        "{}{}",
        job[0].job_id,
        if a.classification & (LIVE_AT_START | LIVE_AT_END) == LIVE_AT_START | LIVE_AT_END {
            "!"
        } else if a.classification & LIVE_AT_START != 0 {
            "<"
        } else if a.classification & LIVE_AT_END != 0 {
            ">"
        } else {
            ""
        }
    )
}

fn format_job_id((_, job): LogDatum, _: LogCtx) -> String {
    format!("{}", job[0].job_id)
}

fn format_host((_, job): LogDatum, _: LogCtx) -> String {
    // The hosts are in the jobs only, we aggregate only for presentation
    let mut hosts = HashSet::<String>::new();
    for j in job {
        hosts.insert(j.hostname.split('.').next().unwrap().to_string());
    }
    format::combine_hosts(hosts.drain().collect::<Vec<String>>())
}

fn format_gpus((_, job): LogDatum, _: LogCtx) -> String {
    // As for hosts, it's OK to do this for presentation.
    let mut gpus = HashSet::<u32>::new();
    for j in job {
        if let Some(ref x) = j.gpus {
            gpus.extend(x);
        }
    }
    let mut term = "";
    let mut s = String::new();
    for x in gpus {
        s += term;
        term = ",";
        s += &x.to_string();
    }
    s
}

fn format_duration((a, _): LogDatum, _: LogCtx) -> String {
    format!(
        "{:}d{:2}h{:2}m",
        a.days, a.hours, a.minutes
    )
}

// An argument could be made that this should be ISO time, at least when the output is CSV, but
// for the time being I'm keeping it compatible with `start` and `end`.
fn format_now((_, _): LogDatum, _: LogCtx) -> String {
    now().format("%Y-%m-%d %H:%M").to_string()
}

fn format_start((a, _): LogDatum, _: LogCtx) -> String {
    a.first.format("%Y-%m-%d %H:%M").to_string()
}

fn format_end((a, _): LogDatum, _: LogCtx) -> String {
    a.last.format("%Y-%m-%d %H:%M").to_string()
}

fn format_cpu_avg((a, _): LogDatum, _: LogCtx) -> String {
    format!("{}", a.cpu_avg)
}

fn format_cpu_peak((a, _): LogDatum, _: LogCtx) -> String {
    format!("{}", a.cpu_peak)
}

fn format_rcpu_avg((a, _): LogDatum, _: LogCtx) -> String {
    format!("{}", a.rcpu_avg)
}

fn format_rcpu_peak((a, _): LogDatum, _: LogCtx) -> String {
    format!("{}", a.rcpu_peak)
}

fn format_mem_avg((a, _): LogDatum, _: LogCtx) -> String {
    format!("{}", a.mem_avg)
}

fn format_mem_peak((a, _): LogDatum, _: LogCtx) -> String {
    format!("{}", a.mem_peak)
}

fn format_rmem_avg((a, _): LogDatum, _: LogCtx) -> String {
    format!("{}", a.rmem_avg)
}

fn format_rmem_peak((a, _): LogDatum, _: LogCtx) -> String {
    format!("{}", a.rmem_peak)
}

fn format_gpu_avg((a, _): LogDatum, _: LogCtx) -> String {
    format!("{}", a.gpu_avg)
}

fn format_gpu_peak((a, _): LogDatum, _: LogCtx) -> String {
    format!("{}", a.gpu_peak)
}

fn format_rgpu_avg((a, _): LogDatum, _: LogCtx) -> String {
    format!("{}", a.rgpu_avg)
}

fn format_rgpu_peak((a, _): LogDatum, _: LogCtx) -> String {
    format!("{}", a.rgpu_peak)
}

fn format_gpumem_avg((a, _): LogDatum, _: LogCtx) -> String {
    format!("{}", a.gpumem_avg)
}

fn format_gpumem_peak((a, _): LogDatum, _: LogCtx) -> String {
    format!("{}", a.gpumem_peak)
}

fn format_rgpumem_avg((a, _): LogDatum, _: LogCtx) -> String {
    format!("{}", a.rgpumem_avg)
}

fn format_rgpumem_peak((a, _): LogDatum, _: LogCtx) -> String {
    format!("{}", a.rgpumem_peak)
}

fn format_command((_, job): LogDatum, _: LogCtx) -> String {
    let mut names = HashSet::new();
    let mut name = "".to_string();
    for entry in job {
        if names.contains(&entry.command) {
            continue;
        }
        if name != "" {
            name += ", ";
        }
        name += &entry.command;
        names.insert(&entry.command);
    }
    name
}


#[cfg(feature = "untagged_sonar_data")]
#[test]
fn test_format_jobs() {
    let filter = |e:&LogEntry| e.job_id <= 2447150;
    let (jobs, _numrec, earliest, latest) = sonarlog::compute_jobs(
        &vec![
            "../sonar_test_data0/2023/05/31/ml8.hpc.uio.no.csv".to_string(),
            "../sonar_test_data0/2023/06/01/ml8.hpc.uio.no.csv".to_string(),
        ],
        &filter,
        /* merge_across_hosts= */ false,
    )
    .unwrap();

    let mut filter_args = JobFilterAndAggregationArgs::default();
    // TODO: Annoying and does not scale - surely there's a better way?
    filter_args.max_cpu_avg = 100000000;
    filter_args.max_cpu_peak = 100000000;
    filter_args.max_rcpu_avg = 100000000;
    filter_args.max_rcpu_peak = 100000000;
    filter_args.max_gpu_avg = 100000000;
    filter_args.max_gpu_peak = 100000000;
    filter_args.max_rgpu_avg = 100000000;
    filter_args.max_rgpu_peak = 100000000;
    let print_args = JobPrintArgs::default();
    let meta_args = MetaArgs::default();
    let mut c = Collector::new();
    aggregate_and_print_jobs(
        &mut c,
        &None,
        &filter_args,
        &print_args,
        &meta_args,
        jobs,
        earliest,
        latest,
    )
    .unwrap();
    c.flush().unwrap();
    let contents = c.get();
    let expected =
"jobm      user      duration  cpu-avg  cpu-peak  mem-avg  mem-peak  gpu-avg  gpu-peak  gpumem-avg  gpumem-peak  host  cmd            
4079<     root      1d16h55m  4        4         1        1         0        0         0           0            ml8   tuned          
4093!     zabbix    1d17h 0m  5        5         1        1         0        0         0           0            ml8   zabbix_agentd  
585616<   larsbent  0d 0h45m  933      1273      194      199       72       84        16          26           ml8   python         
1649588<  riccarsi  0d 3h20m  141      141       127      155       38       44        2           2            ml8   python         
2381069<  einarvid  1d16h55m  2        2         4        4         0        0         0           0            ml8   mongod         
1592463   larsbent  0d 2h44m  594      1292      92       116       76       89        20          37           ml8   python         
1593746   larsbent  0d 2h44m  2701     2834      21       29        52       71        2           3            ml8   python         
1921146   riccarsi  0d20h50m  143      146       104      115       38       42        2           2            ml8   python         
1939269   larsbent  0d 2h59m  536      3095      116      132       79       92        19          33           ml8   python         
1940843   larsbent  0d 2h59m  260      888       47       62        46       58        2           3            ml8   python         
2126454   riccarsi  0d 6h44m  131      134       149      149       57       59        2           3            ml8   python         
2447150   larsbent  0d20h34m  163      178       18       19        0        0         1           1            ml8   python         
";
    assert!(expected == contents);
}


// Presumably there's something standard for this
#[cfg(all(feature = "untagged_sonar_data", test))]
struct Collector {
    storage: Vec<u8>,
}

#[cfg(all(feature = "untagged_sonar_data", test))]
impl Collector {
    fn new() -> Collector {
        Collector { storage: vec![] }
    }

    fn get(&mut self) -> String {
        String::from_utf8(self.storage.clone()).unwrap()
    }
}

#[cfg(all(feature = "untagged_sonar_data", test))]
impl io::Write for Collector {
    fn write(&mut self, buf: &[u8]) -> io::Result<usize> {
        self.storage.extend_from_slice(buf);
        Ok(buf.len())
    }

    fn flush(&mut self) -> io::Result<()> {
        Ok(())
    }
}
