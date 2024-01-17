/// Compute system load aggregates from a set of log entries.
///
/// TODO:
///
/// - For some listings it may be desirable to print a heading?
use crate::format;
use crate::{LoadFilterAndAggregationArgs, LoadPrintArgs, MetaArgs};

use anyhow::{bail, Result};
use sonarlog::{
    self, add_day, add_half_day, add_half_hour, add_hour, empty_logentry, gpuset_to_string, now,
    truncate_to_day, truncate_to_half_day, truncate_to_half_hour, truncate_to_hour, HostFilter,
    InputStreamSet, LogEntry, MergedSampleStreams, Timestamp,
};
use std::boxed::Box;
use std::collections::HashMap;
use std::io;

#[derive(PartialEq, Clone, Copy)]
enum BucketOpt {
    None,
    HalfHourly,
    Hourly,
    HalfDaily,
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
    } else if filter_args.half_daily {
        BucketOpt::HalfDaily
    } else if filter_args.none {
        BucketOpt::None
    } else if filter_args.half_hourly {
        BucketOpt::HalfHourly
    } else {
        BucketOpt::Hourly // Default
    };

    if filter_args.group && bucket_opt == BucketOpt::None {
        bail!("Grouping across hosts requires first bucketing by time");
    }

    let print_opt = if print_args.last {
        PrintOpt::Last
    } else {
        PrintOpt::All // Default
    };

    let (formatters, aliases) = my_formatters();
    let spec = if let Some(ref fmt) = print_args.fmt {
        fmt
    } else {
        FMT_DEFAULTS
    };
    let (fields, others) = format::parse_fields(spec, &formatters, &aliases)?;
    let opts = format::standard_options(&others);
    let relative = fields.iter().any(|x| match *x {
        "rcpu" | "rmem" | "rgpu" | "rgpumem" => true,
        _ => false,
    });

    // After this, everyone can assume there will be a system_config for every host in the data.
    if relative {
        if let Some(ref ht) = system_config {
            for (_, s) in & streams {
                if ht.get(&s[0].hostname).is_none() {
                    bail!("Missing host configuration for {}", &s[0].hostname)
                }
            }
        } else {
            bail!("Relative values requested without config file");
        }
    }

    // There one synthesized sample stream per host.  The samples will all have different
    // timestamps, and each stream will be sorted ascending by timestamp.

    let mut merged_streams = sonarlog::merge_by_host(streams);

    // Bucket the data, if applicable

    if bucket_opt != BucketOpt::None {
        merged_streams = merged_streams
            .drain(0..)
            .map(|stream| match bucket_opt {
                BucketOpt::Hourly => sonarlog::fold_samples_hourly(stream),
                BucketOpt::HalfHourly => sonarlog::fold_samples_half_hourly(stream),
                BucketOpt::Daily => sonarlog::fold_samples_daily(stream),
                BucketOpt::HalfDaily => sonarlog::fold_samples_half_daily(stream),
                BucketOpt::None => panic!("Unexpected"),
            })
            .collect::<MergedSampleStreams>();
    }

    // If grouping, merge the streams across hosts and compute a system config that represents the
    // sum of the hosts in the group.

    let mut the_conf: sonarlog::System = Default::default();
    let mut merged_conf = None;
    if filter_args.group {
        if let Some(ref ht) = system_config {
            for stream in &merged_streams {
                let probe = ht.get(&stream[0].hostname).unwrap();
                if the_conf.description != "" {
                    the_conf.description += "|||"; // JSON-compatible separator
                }
                the_conf.description += &probe.description;
                the_conf.cpu_cores += probe.cpu_cores;
                the_conf.mem_gb += probe.mem_gb;
                the_conf.gpu_cards += probe.gpu_cards;
                the_conf.gpumem_gb += probe.gpumem_gb;
            }
            merged_conf = Some(&the_conf);
        }
        merged_streams = sonarlog::merge_across_hosts_by_time(merged_streams);
        assert!(merged_streams.len() <= 1)
    }

    // Sort hosts lexicographically.  This is not ideal because hosts like c1-10 vs c1-5 are not in
    // the order we expect but at least it's predictable.

    merged_streams.sort_by(|a, b| a[0].hostname.cmp(&b[0].hostname));

    // The handling of hostname is a hack.
    // The handling of JSON is also a hack.
    let explicit_host = fields.iter().any(|x| *x == "host");
    let mut first = true;
    if opts.json {
        output.write("[".as_bytes())?;
    }
    for stream in merged_streams {
        if opts.json {
            if !first {
                output.write(",".as_bytes())?;
            }
            first = false;
        }
        let hostname = stream[0].hostname.clone();
        if !opts.csv && !opts.json && !explicit_host {
            output
                .write(format!("HOST: {}\n", hostname).as_bytes())
                .unwrap();
        }

        let sysconf = if let Some(_) = merged_conf {
            merged_conf
        } else if let Some(ref ht) = system_config {
            ht.get(&hostname)
        } else {
            None
        };

        let ctx = PrintContext {
            sys: sysconf,
            t: now(),
        };

        // For JSON, add richer information about the host so that the client does not have to
        // synthesize this information itself.
        if opts.json {
            let (description, gpu_cards) = if let Some(ref s) = sysconf {
                (s.description.clone(), s.gpu_cards)
            } else {
                ("Unknown".to_string(), 0)
            };
            // TODO: This is not completely safe because the description could contain a non-JSON
            // compatible character such as \t or \n, or a double-quote.
            output.write(format!("{{\"system\":{{\"hostname\":\"{}\",\"description\":\"{}\",\"gpucards\":\"{}\"}},\"records\":",
                                 &hostname,
                                 &description,
                                 gpu_cards).as_bytes())?;
        }

        match bucket_opt {
            BucketOpt::Hourly | BucketOpt::HalfHourly | BucketOpt::Daily | BucketOpt::HalfDaily => {
                if print_opt == PrintOpt::All {
                    let stream = if print_args.compact {
                        stream
                    } else {
                        insert_missing_records(stream, from, to, bucket_opt)
                    };
                    format::format_data(output, &fields, &formatters, &opts, stream, &ctx);
                } else {
                    // Invariant: there's always at least one record
                    // TODO: Really not happy about the clone() here
                    let data = vec![stream[stream.len() - 1].clone()];
                    format::format_data(output, &fields, &formatters, &opts, data, &ctx);
                }
            }
            BucketOpt::None => {
                if print_opt == PrintOpt::All {
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
        }

        if opts.json {
            output.write("}".as_bytes())?;
        }
    }
    if opts.json {
        output.write("]".as_bytes())?;
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

const FMT_DEFAULTS: &str = "date,time,cpu,mem,gpu,gpumem,gpumask";

fn my_formatters() -> (
    HashMap<String, &'static dyn Fn(LoadDatum, LoadCtx) -> String>,
    HashMap<String, Vec<String>>,
) {
    let mut formatters: HashMap<String, &'static dyn Fn(LoadDatum, LoadCtx) -> String> =
        HashMap::new();
    let aliases: HashMap<String, Vec<String>> = HashMap::new();
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

    (formatters, aliases)
}

fn insert_missing_records(
    mut records: Vec<Box<LogEntry>>,
    from: Timestamp,
    to: Timestamp,
    bucket_opt: BucketOpt,
) -> Vec<Box<LogEntry>> {
    let (trunc, step): (fn(Timestamp) -> Timestamp, fn(Timestamp) -> Timestamp) = match bucket_opt {
        BucketOpt::Hourly => (truncate_to_hour, add_hour),
        BucketOpt::HalfHourly => (truncate_to_half_hour, add_half_hour),
        BucketOpt::HalfDaily => (truncate_to_half_day, add_half_day),
        BucketOpt::Daily | BucketOpt::None => (truncate_to_day, add_day),
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
    format!(
        "{}",
        ((d.cpu_util_pct as f64) / (s.cpu_cores as f64)).round()
    )
}

fn format_mem(d: LoadDatum, _: LoadCtx) -> String {
    format!("{}", d.mem_gb as isize)
}

fn format_rmem(d: LoadDatum, ctx: LoadCtx) -> String {
    let s = ctx.sys.unwrap();
    format!(
        "{}",
        ((d.mem_gb as f64) / (s.mem_gb as f64) * 100.0).round()
    )
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
    format!(
        "{}",
        ((d.gpumem_gb as f64) / (s.gpumem_gb as f64) * 100.0).round()
    )
}

fn format_gpus(d: LoadDatum, _: LoadCtx) -> String {
    gpuset_to_string(&d.gpus)
}

fn format_host(d: LoadDatum, _: LoadCtx) -> String {
    format!("{}", d.hostname)
}
