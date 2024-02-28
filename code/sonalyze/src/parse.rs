use crate::format;
use crate::{MetaArgs, ParsePrintArgs};

use anyhow::Result;
use sonarlog::{gpuset_to_string, is_empty_gpuset, GpuStatus, LogEntry};
use std::boxed::Box;
use std::collections::HashMap;
use std::io;

pub fn print_parsed_data(
    output: &mut dyn io::Write,
    print_args: &ParsePrintArgs,
    meta_args: &MetaArgs,
    entries: Vec<Box<LogEntry>>,
) -> Result<()> {
    if meta_args.verbose {
        println!("{} source records", entries.len());
        return Ok(());
    }

    let (formatters, aliases) = my_formatters();
    let spec = if let Some(ref fmt) = print_args.fmt {
        fmt
    } else {
        FMT_DEFAULTS
    };
    let (fields, others) = format::parse_fields(spec, &formatters, &aliases)?;
    let mut opts = format::standard_options(&others);
    // `parse` defaults to headerless un-named csv.  Would be more elegant to pass defaults to
    // standard_options, not hack it in afterwards.
    if !opts.fixed && !opts.csv && !opts.json {
        opts.csv = true;
        opts.header = false;
    }

    format::format_data(
        output,
        &fields,
        &formatters,
        &opts,
        entries,
        &opts.nodefaults,
    );
    Ok(())
}

pub fn fmt_help() -> format::Help {
    let (formatters, aliases) = my_formatters();
    format::Help {
        fields: formatters.keys().cloned().collect::<Vec<String>>(),
        aliases: aliases
            .iter()
            .map(|(k, v)| (k.clone(), v.clone()))
            .collect::<Vec<(String, Vec<String>)>>(),
        defaults: FMT_DEFAULTS.to_string(),
    }
}

const FMT_DEFAULTS: &str = "job,user,cmd";

fn my_formatters() -> (
    HashMap<String, &'static dyn Fn(LogDatum, LogCtx) -> String>,
    HashMap<String, Vec<String>>,
) {
    let mut formatters: HashMap<String, &'static dyn Fn(LogDatum, LogCtx) -> String> =
        HashMap::new();
    let mut aliases: HashMap<String, Vec<String>> = HashMap::new();
    formatters.insert("version".to_string(), &format_version);
    formatters.insert("v".to_string(), &format_version);
    formatters.insert("localtime".to_string(), &format_localtime);
    formatters.insert("time".to_string(), &format_time);
    formatters.insert("host".to_string(), &format_host);
    formatters.insert("cores".to_string(), &format_cores);
    formatters.insert("memtotal".to_string(), &format_memtotal);
    formatters.insert("user".to_string(), &format_user);
    formatters.insert("pid".to_string(), &format_pid);
    formatters.insert("job".to_string(), &format_job);
    formatters.insert("cmd".to_string(), &format_cmd);
    formatters.insert("cpu_pct".to_string(), &format_cpu_pct);
    formatters.insert("cpu%".to_string(), &format_cpu_pct);
    formatters.insert("mem_gb".to_string(), &format_mem_gb);
    formatters.insert("res_gb".to_string(), &format_res_gb);
    formatters.insert("cpukib".to_string(), &format_cpukib);
    formatters.insert("gpus".to_string(), &format_gpus);
    formatters.insert("gpu_pct".to_string(), &format_gpu_pct);
    formatters.insert("gpu%".to_string(), &format_gpu_pct);
    formatters.insert("gpumem_pct".to_string(), &format_gpumem_pct);
    formatters.insert("gpumem%".to_string(), &format_gpumem_pct);
    formatters.insert("gpumem_gb".to_string(), &format_gpumem_gb);
    formatters.insert("gpukib".to_string(), &format_gpukib);
    formatters.insert("gpu_status".to_string(), &format_gpu_status);
    formatters.insert("gpufail".to_string(), &format_gpu_status);
    formatters.insert("cputime_sec".to_string(), &format_cputime_sec);
    formatters.insert("rolledup".to_string(), &format_rolledup);
    formatters.insert("cpu_util_pct".to_string(), &format_cpu_util_pct);

    aliases.insert(
        "all".to_string(),
        vec![
            "version".to_string(),
            "localtime".to_string(),
            "host".to_string(),
            "cores".to_string(),
            "memtotal".to_string(),
            "user".to_string(),
            "pid".to_string(),
            "job".to_string(),
            "cmd".to_string(),
            "cpu_pct".to_string(),
            "mem_gb".to_string(),
            "res_gb".to_string(),
            "gpus".to_string(),
            "gpu_pct".to_string(),
            "gpumem_pct".to_string(),
            "gpumem_gb".to_string(),
            "gpu_status".to_string(),
            "cputime_sec".to_string(),
            "rolledup".to_string(),
            "cpu_util_pct".to_string(),
        ],
    );

    aliases.insert(
        "roundtrip".to_string(),
        vec![
            "v".to_string(),
            "time".to_string(),
            "host".to_string(),
            "cores".to_string(),
            "user".to_string(),
            "job".to_string(),
            "pid".to_string(),
            "cmd".to_string(),
            "cpu%".to_string(),
            "cpukib".to_string(),
            "gpus".to_string(),
            "gpu%".to_string(),
            "gpumem%".to_string(),
            "gpukib".to_string(),
            "gpufail".to_string(),
            "cputime_sec".to_string(),
            "rolledup".to_string(),
        ],
    );

    (formatters, aliases)
}

type LogDatum<'a> = &'a Box<LogEntry>;
type LogCtx<'a> = &'a bool;

fn format_version(d: LogDatum, _: LogCtx) -> String {
    format!("{}.{}.{}", d.major, d.minor, d.bugfix)
}

fn format_time(d: LogDatum, _: LogCtx) -> String {
    d.timestamp.format("%Y-%m-%dT%H:%M:%SZ").to_string()
}

fn format_localtime(d: LogDatum, _: LogCtx) -> String {
    d.timestamp.format("%Y-%m-%d %H:%M").to_string()
}

fn format_host(d: LogDatum, _: LogCtx) -> String {
    d.hostname.to_string()
}

fn format_cores(d: LogDatum, _: LogCtx) -> String {
    d.num_cores.to_string()
}

fn format_memtotal(d: LogDatum, _: LogCtx) -> String {
    (d.memtotal_gb.round() as isize).to_string()
}

fn format_user(d: LogDatum, _: LogCtx) -> String {
    d.user.to_string()
}

fn format_pid(d: LogDatum, nodefaults: LogCtx) -> String {
    if *nodefaults && d.pid == 0 {
        "*skip*".to_string()
    } else {
        d.pid.to_string()
    }
}

fn format_job(d: LogDatum, _: LogCtx) -> String {
    d.job_id.to_string()
}

fn format_cmd(d: LogDatum, _: LogCtx) -> String {
    d.command.to_string()
}

fn format_cpu_pct(d: LogDatum, _: LogCtx) -> String {
    d.cpu_pct.to_string()
}

fn format_mem_gb(d: LogDatum, _: LogCtx) -> String {
    (d.mem_gb.round() as isize).to_string()
}

fn format_res_gb(d: LogDatum, _: LogCtx) -> String {
    (d.rssanon_gb.round() as isize).to_string()
}

fn format_cpukib(d: LogDatum, _: LogCtx) -> String {
    ((d.mem_gb * 1024.0 * 1024.0).round() as isize).to_string()
}

fn format_gpus(d: LogDatum, nodefaults: LogCtx) -> String {
    if *nodefaults && is_empty_gpuset(&d.gpus) {
        "*skip*".to_string()
    } else {
        gpuset_to_string(&d.gpus)
    }
}

fn format_gpu_pct(d: LogDatum, nodefaults: LogCtx) -> String {
    if *nodefaults && d.gpu_pct == 0.0 {
        "*skip*".to_string()
    } else {
        d.gpu_pct.to_string()
    }
}

fn format_gpumem_pct(d: LogDatum, nodefaults: LogCtx) -> String {
    if *nodefaults && d.gpumem_pct == 0.0 {
        "*skip*".to_string()
    } else {
        d.gpumem_pct.to_string()
    }
}

fn format_gpumem_gb(d: LogDatum, nodefaults: LogCtx) -> String {
    if *nodefaults && d.gpumem_gb.round() == 0.0 {
        "*skip*".to_string()
    } else {
        (d.gpumem_gb.round() as isize).to_string()
    }
}

fn format_gpukib(d: LogDatum, nodefaults: LogCtx) -> String {
    if *nodefaults && d.gpumem_gb.round() == 0.0 {
        "*skip*".to_string()
    } else {
        ((d.gpumem_gb * 1024.0 * 1024.0).round() as isize).to_string()
    }
}

fn format_gpu_status(d: LogDatum, nodefaults: LogCtx) -> String {
    if *nodefaults && d.gpu_status == GpuStatus::Ok {
        "*skip*".to_string()
    } else {
        (d.gpu_status as i32).to_string()
    }
}

fn format_cputime_sec(d: LogDatum, nodefaults: LogCtx) -> String {
    if *nodefaults && d.cputime_sec == 0.0 {
        "*skip*".to_string()
    } else {
        d.cputime_sec.to_string()
    }
}

fn format_rolledup(d: LogDatum, nodefaults: LogCtx) -> String {
    if *nodefaults && d.rolledup == 0 {
        "*skip*".to_string()
    } else {
        d.rolledup.to_string()
    }
}

fn format_cpu_util_pct(d: LogDatum, nodefaults: LogCtx) -> String {
    if *nodefaults && d.cpu_util_pct == 0.0 {
        "*skip*".to_string()
    } else {
        d.cpu_util_pct.to_string()
    }
}
