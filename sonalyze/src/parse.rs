use crate::format;
use crate::{ParsePrintArgs, MetaArgs};

use anyhow::Result;
use sonarlog::{gpuset_to_string, LogEntry};
use std::boxed::Box;
use std::collections::HashMap;
use std::io;

pub fn print_parsed_data(
    output: &mut dyn io::Write,
    print_args: &ParsePrintArgs,
    meta_args: &MetaArgs,
    entries: Vec<Box<LogEntry>>
) -> Result<()> {
    if meta_args.verbose {
        eprintln!("{} source records", entries.len());
        return Ok(())
    }

    let (formatters, aliases) = my_formatters();
    let spec = if let Some(ref fmt) = print_args.fmt {
        fmt
    } else {
        FMT_DEFAULTS
    };
    let (fields, others) = format::parse_fields(spec, &formatters, &aliases);
    let mut opts = format::standard_options(&others);
    // `parse` defaults to headerless un-named csv.  Would be more elegant to pass defaults to
    // standard_options, not hack it in afterwards.
    if !opts.fixed && !opts.csv && !opts.json {
        opts.csv = true;
        opts.header = false;
    }
    if fields.len() > 0 {
        format::format_data(output, &fields, &formatters, &opts, entries, &false);
    }
    Ok(())
}

pub fn fmt_help() -> format::Help {
    let (formatters, aliases) = my_formatters();
    format::Help {
        fields: formatters.iter().map(|(k, _)| k.clone()).collect::<Vec<String>>(),
        aliases: aliases.iter().map(|(k,v)| (k.clone(), v.clone())).collect::<Vec<(String, Vec<String>)>>(),
        defaults: FMT_DEFAULTS.to_string(),
    }
}

const FMT_DEFAULTS : &str = "job,user,cmd";

fn my_formatters() -> (HashMap<String, &'static dyn Fn(LogDatum, LogCtx) -> String>,
                       HashMap<String, Vec<String>>) {
    let mut formatters: HashMap<String, &'static dyn Fn(LogDatum, LogCtx) -> String> = HashMap::new();
    let mut aliases: HashMap<String, Vec<String>> = HashMap::new();
    formatters.insert("version".to_string(), &format_version);
    formatters.insert("time".to_string(), &format_time);
    formatters.insert("host".to_string(), &format_host);
    formatters.insert("cores".to_string(), &format_cores);
    formatters.insert("user".to_string(), &format_user);
    formatters.insert("pid".to_string(), &format_pid);
    formatters.insert("job".to_string(), &format_job);
    formatters.insert("cmd".to_string(), &format_cmd);
    formatters.insert("cpu_pct".to_string(), &format_cpu_pct);
    formatters.insert("mem_gb".to_string(), &format_mem_gb);
    formatters.insert("gpus".to_string(), &format_gpus);
    formatters.insert("gpu_pct".to_string(), &format_gpu_pct);
    formatters.insert("gpumem_pct".to_string(), &format_gpumem_pct);
    formatters.insert("gpumem_gb".to_string(), &format_gpumem_gb);
    formatters.insert("gpu_status".to_string(), &format_gpu_status);
    formatters.insert("cputime_sec".to_string(), &format_cputime_sec);
    formatters.insert("rolledup".to_string(), &format_rolledup);
    formatters.insert("cpu_util_pct".to_string(), &format_cpu_util_pct);

    aliases.insert("all".to_string(),
                   vec!["version".to_string(),"time".to_string(),"host".to_string(),
                        "cores".to_string(),"user".to_string(),"pid".to_string(),
                        "job".to_string(),"cmd".to_string(),"cpu_pct".to_string(),
                        "mem_gb".to_string(),"gpus".to_string(),"gpu_pct".to_string(),
                        "gpumem_pct".to_string(),"gpumem_gb".to_string(),
                        "gpu_status".to_string(),"cputime_sec".to_string(),
                        "rolledup".to_string(),"cpu_util_pct".to_string()]);

    (formatters, aliases)
}

type LogDatum<'a> = &'a Box<LogEntry>;
type LogCtx<'a> = &'a bool;

fn format_version(d: LogDatum, _: LogCtx) -> String {
    d.version.clone()
}

fn format_time(d: LogDatum, _: LogCtx) -> String {
    d.timestamp.format("%Y-%m-%d %H:%M").to_string()
}

fn format_host(d: LogDatum, _: LogCtx) -> String {
    d.hostname.clone()
}

fn format_cores(d: LogDatum, _: LogCtx) -> String {
    d.num_cores.to_string()
}

fn format_user(d: LogDatum, _: LogCtx) -> String {
    d.user.clone()
}

fn format_pid(d: LogDatum, _: LogCtx) -> String {
    d.pid.to_string()
}

fn format_job(d: LogDatum, _: LogCtx) -> String {
    d.job_id.to_string()
}

fn format_cmd(d: LogDatum, _: LogCtx) -> String {
    d.command.clone()
}

fn format_cpu_pct(d: LogDatum, _: LogCtx) -> String {
    d.cpu_pct.to_string()
}

fn format_mem_gb(d: LogDatum, _: LogCtx) -> String {
    ((d.mem_gb.round()) as isize).to_string()
}

fn format_gpus(d: LogDatum, _: LogCtx) -> String {
    gpuset_to_string(&d.gpus)
}

fn format_gpu_pct(d: LogDatum, _: LogCtx) -> String {
    d.gpu_pct.to_string()
}

fn format_gpumem_pct(d: LogDatum, _: LogCtx) -> String {
    d.gpumem_pct.to_string()
}

fn format_gpumem_gb(d: LogDatum, _: LogCtx) -> String {
    ((d.gpumem_gb.round()) as isize).to_string()
}

fn format_gpu_status(d: LogDatum, _: LogCtx) -> String {
    (d.gpu_status as i32).to_string()
}

fn format_cputime_sec(d: LogDatum, _: LogCtx) -> String {
    d.cputime_sec.to_string()
}

fn format_rolledup(d: LogDatum, _: LogCtx) -> String {
    d.rolledup.to_string()
}

fn format_cpu_util_pct(d: LogDatum, _: LogCtx) -> String {
    d.cpu_util_pct.to_string()
}

