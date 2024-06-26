use crate::format;
use crate::{MetaArgs, ParsePrintArgs};

use anyhow::Result;
use rustutils::Timestamp;
use sonarlog::Timebounds;
use std::collections::HashMap;
use std::io;
use ustr::Ustr;

struct Item {
    host: Ustr,
    earliest: Timestamp,
    latest: Timestamp,
}

pub fn print(
    output: &mut dyn io::Write,
    print_args: &ParsePrintArgs,
    meta_args: &MetaArgs,
    mut bounds: Timebounds,
) -> Result<()> {
    if meta_args.verbose {
        println!("{} source records", bounds.len());
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
    // `metadata` defaults to headerless un-named csv.  Would be more elegant to pass defaults to
    // standard_options, not hack it in afterwards.
    if !opts.fixed && !opts.csv && !opts.json && !opts.awk {
        opts.csv = true;
        opts.header = false;
    }
    let mut data = bounds
        .drain()
        .map(|(k, v)| Item {
            host: k,
            earliest: v.earliest,
            latest: v.latest,
        })
        .collect::<Vec<Item>>();
    data.sort_by(|a, b| {
        if a.host == b.host {
            a.earliest.cmp(&b.earliest)
        } else {
            a.host.cmp(&b.host)
        }
    });
    format::format_data(output, &fields, &formatters, &opts, data, &false);
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

const FMT_DEFAULTS: &str = "all";

fn my_formatters() -> (
    HashMap<String, &'static dyn Fn(LogDatum, LogCtx) -> String>,
    HashMap<String, Vec<String>>,
) {
    let mut formatters: HashMap<String, &'static dyn Fn(LogDatum, LogCtx) -> String> =
        HashMap::new();
    let mut aliases: HashMap<String, Vec<String>> = HashMap::new();
    formatters.insert("host".to_string(), &format_host);
    formatters.insert("earliest".to_string(), &format_earliest);
    formatters.insert("latest".to_string(), &format_latest);

    aliases.insert(
        "all".to_string(),
        vec![
            "host".to_string(),
            "earliest".to_string(),
            "latest".to_string(),
        ],
    );

    (formatters, aliases)
}

type LogDatum<'a> = &'a Item;
type LogCtx<'a> = &'a bool;

fn format_host(d: LogDatum, _: LogCtx) -> String {
    d.host.to_string()
}

fn format_earliest(d: LogDatum, _: LogCtx) -> String {
    d.earliest.format("%Y-%m-%d %H:%M").to_string()
}

fn format_latest(d: LogDatum, _: LogCtx) -> String {
    d.latest.format("%Y-%m-%d %H:%M").to_string()
}
