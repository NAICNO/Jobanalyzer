/// Jobs printer
///
/// Feature: One could imagine other sort orders for the output than least-recently-started-first.
/// This only matters for the --numjobs switch.
use crate::format;
use crate::jobs::{JobSummary, LIVE_AT_END, LIVE_AT_START};
/* BREAKDOWN
 * use crate::jobs::{LEVEL_MASK, LEVEL_SHIFT};
 */
use crate::{JobPrintArgs, MetaArgs};

use anyhow::{bail, Result};
use sonarlog::{self, empty_gpuset, gpuset_to_string, now, union_gpuset, Timestamp};
use std::collections::{HashMap, HashSet};
use std::io;
use std::ops::Add;
use ustr::Ustr;

pub fn print_jobs(
    output: &mut dyn io::Write,
    system_config: &Option<HashMap<String, sonarlog::System>>,
    mut hosts: HashSet<Ustr>,
    mut jobvec: Vec<JobSummary>,
    print_args: &JobPrintArgs,
    meta_args: &MetaArgs,
) -> Result<()> {
    // And sort ascending by lowest beginning timestamp, and if those are equal (which happens when
    // we start reading logs at some arbitrary date), by job number.
    jobvec.sort_by(|a, b| {
        if a.aggregate.first == b.aggregate.first {
            a.job[0].job_id.cmp(&b.job[0].job_id)
        } else {
            a.aggregate.first.cmp(&b.aggregate.first)
        }
    });

    // Select a number of jobs per user, if applicable.  This means working from the bottom up
    // in the vector and marking the n first per user.  We need a hashmap user -> count.

    if let Some(n) = print_args.numjobs {
        let mut counts: HashMap<&str, usize> = HashMap::new();
        jobvec
            .iter_mut()
            .rev()
            .for_each(|JobSummary { aggregate, job, .. }| {
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
            |JobSummary { aggregate, .. }| {
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
        println!("Number of jobs after output filtering: {}", numselected);
    }

    // Now print.

    if meta_args.verbose {
        return Ok(());
    }

    // Unix user names are max 8 chars.
    // Linux pids are max 7 decimal digits.
    // We don't care about seconds in the timestamp, nor timezone.

    if meta_args.raw {
        jobvec.iter().for_each(|JobSummary { aggregate, job, .. }| {
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
    } else {
        // Note, numselected may be zero, but for the JSON format we can't bail out early.
        let (/*mut*/ formatters, aliases) = my_formatters();
        /* BREAKDOWN
         * if print_args.breakdown.is_some() {
         *     formatters.insert("*prefix*".to_string(), &format_prefix);
         * }
         */
        let spec = if let Some(ref fmt) = print_args.fmt {
            fmt
        } else {
            FMT_DEFAULTS
        };
        let (fields, others) = format::parse_fields(spec, &formatters, &aliases)?;
        let opts = format::standard_options(&others);
        let relative = fields.iter().any(|x| match *x {
            "rcpu-avg" | "rcpu-peak" | "rmem-avg" | "rmem-peak" | "rgpu-avg" | "rgpu-peak"
            | "rgpumem-avg" | "rgpumem-peak" => true,
            _ => false,
        });
        if relative {
            if let Some(ref ht) = system_config {
                for host in hosts.drain() {
                    if ht.get(host.as_str()).is_none() {
                        // Note that system_config is not actually used during printing.  What we're
                        // doing here is making somebody (hopefully) aware that there are problems.
                        // We have generated nonsense/zero data for relative fields for anything to
                        // do with this host, already.  But it's only here that we have available
                        // information that we are asking for relative fields.
                        eprintln!("Warning: Missing host configuration for {}", &host);
                    }
                }
            } else {
                bail!("Relative values requested without config file");
            }
        }

        let mut selected = vec![];
        for
        /*mut*/
        job in jobvec.drain(0..) {
            if job.aggregate.selected {
                /* BREAKDOWN
                 * let breakdown = job.breakdown;
                 * job.breakdown = None;
                 */
                selected.push(job);
                /* BREAKDOWN
                 * expand_subjobs(1, breakdown, &mut selected);
                 */
            }
        }
        let c = Context {
            t: now(),
            fixed_format: !opts.json && !opts.csv,
        };
        format::format_data(output, &fields, &formatters, &opts, selected, &c);
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
        defaults: FMT_DEFAULTS.to_string(),
    }
}

const FMT_DEFAULTS: &str = "std,cpu,mem,gpu,gpumem,cmd";

fn my_formatters() -> (
    HashMap<String, &'static dyn Fn(LogDatum, LogCtx) -> String>,
    HashMap<String, Vec<String>>,
) {
    let mut formatters: HashMap<String, &dyn Fn(LogDatum, LogCtx) -> String> = HashMap::new();
    let mut aliases: HashMap<String, Vec<String>> = HashMap::new();

    formatters.insert("jobm".to_string(), &format_jobm_id);
    formatters.insert("job".to_string(), &format_job_id);
    formatters.insert("user".to_string(), &format_user);
    formatters.insert("duration".to_string(), &format_duration);
    formatters.insert("duration/sec".to_string(), &format_duration_sec);
    formatters.insert("start".to_string(), &format_start);
    formatters.insert("start/sec".to_string(), &format_start_sec);
    formatters.insert("end".to_string(), &format_end);
    formatters.insert("end/sec".to_string(), &format_end_sec);
    formatters.insert("cpu-avg".to_string(), &format_cpu_avg);
    formatters.insert("cpu-peak".to_string(), &format_cpu_peak);
    formatters.insert("rcpu-avg".to_string(), &format_rcpu_avg);
    formatters.insert("rcpu-peak".to_string(), &format_rcpu_peak);
    formatters.insert("mem-avg".to_string(), &format_mem_avg);
    formatters.insert("mem-peak".to_string(), &format_mem_peak);
    formatters.insert("rmem-avg".to_string(), &format_rmem_avg);
    formatters.insert("rmem-peak".to_string(), &format_rmem_peak);
    formatters.insert("res-avg".to_string(), &format_res_avg);
    formatters.insert("res-peak".to_string(), &format_res_peak);
    formatters.insert("rres-avg".to_string(), &format_rres_avg);
    formatters.insert("rres-peak".to_string(), &format_rres_peak);
    formatters.insert("gpu-avg".to_string(), &format_gpu_avg);
    formatters.insert("gpu-peak".to_string(), &format_gpu_peak);
    formatters.insert("rgpu-avg".to_string(), &format_rgpu_avg);
    formatters.insert("rgpu-peak".to_string(), &format_rgpu_peak);
    formatters.insert("gpumem-avg".to_string(), &format_gpumem_avg);
    formatters.insert("gpumem-peak".to_string(), &format_gpumem_peak);
    formatters.insert("rgpumem-avg".to_string(), &format_rgpumem_avg);
    formatters.insert("rgpumem-peak".to_string(), &format_rgpumem_peak);
    formatters.insert("gpus".to_string(), &format_gpus);
    formatters.insert("gpufail".to_string(), &format_gpufail);
    formatters.insert("cmd".to_string(), &format_command);
    formatters.insert("host".to_string(), &format_host);
    formatters.insert("now".to_string(), &format_now);
    formatters.insert("now/sec".to_string(), &format_now_sec);
    formatters.insert("classification".to_string(), &format_classification);
    formatters.insert("cputime".to_string(), &format_cputime);
    formatters.insert("cputime/sec".to_string(), &format_cputime_sec);
    formatters.insert("gputime".to_string(), &format_gputime);
    formatters.insert("gputime/sec".to_string(), &format_gputime_sec);

    aliases.insert(
        "std".to_string(),
        vec![
            "jobm".to_string(),
            "user".to_string(),
            "duration".to_string(),
            "host".to_string(),
        ],
    );
    aliases.insert(
        "cpu".to_string(),
        vec!["cpu-avg".to_string(), "cpu-peak".to_string()],
    );
    aliases.insert(
        "rcpu".to_string(),
        vec!["rcpu-avg".to_string(), "rcpu-peak".to_string()],
    );
    aliases.insert(
        "mem".to_string(),
        vec!["mem-avg".to_string(), "mem-peak".to_string()],
    );
    aliases.insert(
        "rmem".to_string(),
        vec!["rmem-avg".to_string(), "rmem-peak".to_string()],
    );
    aliases.insert(
        "res".to_string(),
        vec!["res-avg".to_string(), "res-peak".to_string()],
    );
    aliases.insert(
        "rres".to_string(),
        vec!["rres-avg".to_string(), "rres-peak".to_string()],
    );
    aliases.insert(
        "gpu".to_string(),
        vec!["gpu-avg".to_string(), "gpu-peak".to_string()],
    );
    aliases.insert(
        "rgpu".to_string(),
        vec!["rgpu-avg".to_string(), "rgpu-peak".to_string()],
    );
    aliases.insert(
        "gpumem".to_string(),
        vec!["gpumem-avg".to_string(), "gpumem-peak".to_string()],
    );
    aliases.insert(
        "rgpumem".to_string(),
        vec!["rgpumem-avg".to_string(), "rgpumem-peak".to_string()],
    );

    (formatters, aliases)
}

/*
 * fn expand_subjobs(
 *     level: u32,
 *     breakdown: Option<(String, Vec<JobSummary>)>,
 *     selected: &mut Vec<JobSummary>,
 * ) {
 *     if let Some((tag, mut subjobs)) = breakdown {
 *         match tag.as_str() {
 *             "host" => {
 *                 subjobs.sort_by(|a, b| a.job[0].hostname.cmp(&b.job[0].hostname));
 *             }
 *             "command" => {
 *                 subjobs.sort_by(|a, b| a.job[0].command.cmp(&b.job[0].command));
 *             }
 *             _ => {}
 *         }
 *         for mut subjob in subjobs {
 *             let sub_breakdown = subjob.breakdown;
 *             subjob.breakdown = None;
 *             subjob.aggregate.classification |= level << LEVEL_SHIFT;
 *             selected.push(subjob);
 *             expand_subjobs(level + 1, sub_breakdown, selected);
 *         }
 *     }
 * }
 */

struct Context {
    t: Timestamp,
    fixed_format: bool,
}

type LogDatum<'a> = &'a JobSummary;
type LogCtx<'a> = &'a Context;

fn format_user(JobSummary { job, .. }: LogDatum, _: LogCtx) -> String {
    job[0].user.to_string()
}

fn format_jobm_id(
    JobSummary {
        aggregate: a, job, ..
    }: LogDatum,
    _: LogCtx,
) -> String {
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

fn format_job_id(JobSummary { job, .. }: LogDatum, _: LogCtx) -> String {
    format!("{}", job[0].job_id)
}

/* BREAKDOWN
 * fn format_prefix(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
 *     let mut s = "".to_string();
 *     let mut level = (a.classification >> LEVEL_SHIFT) & LEVEL_MASK;
 *     while level > 0 {
 *         s += "*";
 *         level -= 1;
 *     }
 *     s
 * }
 */

fn format_host(JobSummary { job, .. }: LogDatum, c: LogCtx) -> String {
    // The hosts are in the jobs only, we aggregate only for presentation
    let mut hosts = HashSet::<Ustr>::new();
    if c.fixed_format {
        for j in job {
            hosts.insert(Ustr::from(j.hostname.split('.').next().unwrap()));
        }
    } else {
        for j in job {
            hosts.insert(j.hostname);
        }
    }
    sonarlog::combine_hosts(hosts.drain().collect::<Vec<Ustr>>()).to_string()
}

fn format_gpus(JobSummary { job, .. }: LogDatum, _: LogCtx) -> String {
    // As for hosts, it's OK to do create this set only for the presentation.
    //
    // If the gpu set is "unknown" in any of the job records then the result is also "unknown", this
    // is probably OK.  We could instead have kept it as "unknown+1,2,3" but this seems unnecessary.
    let mut gpus = empty_gpuset();
    for j in job {
        union_gpuset(&mut gpus, &j.gpus);
    }
    gpuset_to_string(&gpus)
}

fn format_duration(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    format!("{:}d{:2}h{:2}m", a.days, a.hours, a.minutes)
}

fn format_duration_sec(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    // We have no seconds, all job lengths are measured in minutes
    format!("{}", 60 * (a.minutes + 60 * (a.hours + (a.days * 24))))
}

// Note that this is frequently longer than the job duration b/c multicore.  I mention this only
// because it can be confusing.
fn format_cputime(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    // See below for a description of the computation.
    let duration = 60 * (a.minutes + 60 * (a.hours + (a.days * 24)));
    let mut cputime_sec = (a.cpu_avg * duration as f64 / 100.0).round() as i64;
    if cputime_sec % 60 >= 30 {
        cputime_sec += 30;
    }
    let _seconds = cputime_sec % 60;
    let minutes = (cputime_sec / 60) % 60;
    let hours = (cputime_sec / (60 * 60)) % 24;
    let days = cputime_sec / (60 * 60 * 24);
    format!("{:}d{:2}h{:2}m", days, hours, minutes)
}

fn format_cputime_sec(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    // The unit for average cpu utilization is core-seconds per second, we multiply this by duration
    // (whose units is second) to get total core-seconds for the job.  Finally scale by 100 because
    // the cpu_avg numbers are expressed in integer percentage point.
    let duration = 60 * (a.minutes + 60 * (a.hours + (a.days * 24)));
    let cputime = (a.cpu_avg * duration as f64 / 100.0).round() as i64;
    format!("{}", cputime)
}

fn format_gputime(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    // See below for a description of the computation.
    let duration = 60 * (a.minutes + 60 * (a.hours + (a.days * 24)));
    let mut gputime_sec = (a.gpu_avg * duration as f64 / 100.0).round() as i64;
    if gputime_sec % 60 >= 30 {
        gputime_sec += 30;
    }
    let _seconds = gputime_sec % 60;
    let minutes = (gputime_sec / 60) % 60;
    let hours = (gputime_sec / (60 * 60)) % 24;
    let days = gputime_sec / (60 * 60 * 24);
    format!("{:}d{:2}h{:2}m", days, hours, minutes)
}

fn format_gputime_sec(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    // The unit for average gpu utilization is card-seconds per second, we multiply this by duration
    // (whose units is second) to get total card-seconds for the job.  Finally scale by 100 because
    // the gpu_avg numbers are expressed in integer percentage point.
    let duration = 60 * (a.minutes + 60 * (a.hours + (a.days * 24)));
    let gputime = (a.gpu_avg * duration as f64 / 100.0).round() as i64;
    format!("{}", gputime)
}

// An argument could be made that this should be ISO time, at least when the output is CSV, but
// for the time being I'm keeping it compatible with `start` and `end`.
fn format_now(_: LogDatum, c: LogCtx) -> String {
    c.t.format("%Y-%m-%d %H:%M").to_string()
}

fn format_now_sec(_: LogDatum, c: LogCtx) -> String {
    c.t.format("%s").to_string()
}

fn format_start(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    a.first.format("%Y-%m-%d %H:%M").to_string()
}

fn format_start_sec(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    a.first.format("%s").to_string()
}

fn format_end(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    a.last.format("%Y-%m-%d %H:%M").to_string()
}

fn format_end_sec(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    a.last.format("%s").to_string()
}

fn format_cpu_avg(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    format!("{}", a.cpu_avg)
}

fn format_cpu_peak(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    format!("{}", a.cpu_peak)
}

fn format_rcpu_avg(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    format!("{}", a.rcpu_avg)
}

fn format_rcpu_peak(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    format!("{}", a.rcpu_peak)
}

fn format_mem_avg(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    format!("{}", a.mem_avg)
}

fn format_mem_peak(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    format!("{}", a.mem_peak)
}

fn format_rmem_avg(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    format!("{}", a.rmem_avg)
}

fn format_rmem_peak(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    format!("{}", a.rmem_peak)
}

fn format_res_avg(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    format!("{}", a.res_avg)
}

fn format_res_peak(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    format!("{}", a.res_peak)
}

fn format_rres_avg(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    format!("{}", a.rres_avg)
}

fn format_rres_peak(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    format!("{}", a.rres_peak)
}

fn format_gpu_avg(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    format!("{}", a.gpu_avg)
}

fn format_gpu_peak(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    format!("{}", a.gpu_peak)
}

fn format_rgpu_avg(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    format!("{}", a.rgpu_avg)
}

fn format_rgpu_peak(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    format!("{}", a.rgpu_peak)
}

fn format_gpumem_avg(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    format!("{}", a.gpumem_avg)
}

fn format_gpumem_peak(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    format!("{}", a.gpumem_peak)
}

fn format_rgpumem_avg(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    format!("{}", a.rgpumem_avg)
}

fn format_rgpumem_peak(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    format!("{}", a.rgpumem_peak)
}

fn format_gpufail(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    format!("{}", a.gpu_status as i32)
}

fn format_classification(JobSummary { aggregate: a, .. }: LogDatum, _: LogCtx) -> String {
    format!("0x{:x}", a.classification)
}

fn format_command(JobSummary { job, .. }: LogDatum, _: LogCtx) -> String {
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
