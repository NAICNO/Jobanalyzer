/// Generic formatting code for a set of data extracted from a data structure to be presented
/// columnar, as csv, or as json, and (except for json) with or without a header and with or without
/// named fields.

use csv;
use json;
use std::collections::{HashMap, HashSet};
use std::io;

pub struct Help {
    pub fields: Vec<String>,
    pub aliases: Vec<(String, Vec<String>)>,
    pub defaults: String,
}

pub fn maybe_help<F>(fmt: &Option<String>, f: F) -> bool
where
    F: Fn() -> Help
{
    if let Some(ref s) = fmt {
        if s.as_str() == "help" || s.starts_with("help") {
            let mut help = f();
            println!("Syntax:\n  --fmt=(field|alias|control),...");
            println!("\nFields:");
            help.fields.sort();
            for f in help.fields {
                println!("  {f}");
            }
            if help.aliases.len() > 0 {
                println!("\nAliases:");
                for (name, mut fields) in help.aliases {
                    fields.sort();
                    let explication = (&fields).join(",");
                    println!("  {name} --> {explication}");
                }
            }
            println!("\nDefaults:\n  {}", help.defaults);
            println!("\nControl:\n  csv\n  csvnamed  \n  fixed\n  json\n  header\n  noheader\n  tag:<tagvalue>");
            return true;
        }
    }
    return false;
}

/// Return a vector of the known fields in `spec` wrt the formatters, and a HashSet of any other
/// strings found in `spec`

pub fn parse_fields<'a, DataT, FmtT, CtxT>(
    spec: &'a str,
    formatters: &HashMap<String, FmtT>,
    aliases: &'a HashMap<String, Vec<String>>,
) -> (Vec<&'a str>, HashSet<&'a str>)
where
    FmtT: Fn(&DataT, CtxT) -> String,
    CtxT: Copy,
{
    let mut others = HashSet::new();
    let mut fields = vec![];
    for x in spec.split(',') {
        if formatters.get(x).is_some() {
            fields.push(x);
        } else if let Some(aliases) = aliases.get(x) {
            for alias in aliases {
                if formatters.get(alias).is_some() {
                    fields.push(alias.as_ref());
                } else {
                    others.insert(alias.as_ref());
                }
            }
        } else {
            others.insert(x);
        }
    }
    (fields, others)
}

pub struct FormatOptions {
    pub tag: Option<String>,
    pub json: bool,             // json explicitly requested
    pub csv: bool,              // csv or csvnamed explicitly requested
    pub fixed: bool,            // fixed output explicitly requested
    pub named: bool,            // csvnamed explicitly requested
    pub header: bool,           // true if nothing requested b/c fixed+header is default
}

pub fn standard_options(others: &HashSet<&str>) -> FormatOptions {
    let csvnamed = others.get("csvnamed").is_some();
    let csv = others.get("csv").is_some() || csvnamed;
    let json = others.get("json").is_some() && !csv;
    let fixed = others.get("fixed").is_some() && !csv &&!json;
    // json gets no header, even if one is requested
    let header =
        (!csv && !json && !others.get("noheader").is_some()) || (csv && others.get("header").is_some());
    let mut tag : Option<String> = None;
    for x in others {
        if x.starts_with("tag:") {
            tag = Some(x[4..].to_string());
            break;
        }
    }
    FormatOptions {
        csv,
        json,
        header,
        tag,
        fixed,
        named: csvnamed
    }
}

/// The `fields` are the names of formatting functions to get from the `formatters`, these are
/// applied to the `data`.  Set `opts.header` to true to print a first row with field names as a
/// header (independent of csv).  Set `opts.csv` to true to get CSV output instead of fixed-format.
/// Set `opts.tag` to Some(s) to print a tag=s field in the output.

pub fn format_data<'a, DataT, FmtT, CtxT>(
    output: &mut dyn io::Write,
    fields: &[&'a str],
    formatters: &HashMap<String, FmtT>,
    opts: &FormatOptions,
    data: Vec<DataT>,
    ctx: CtxT,
) where
    FmtT: Fn(&DataT, CtxT) -> String,
    CtxT: Copy,
{
    let mut cols = Vec::<Vec<String>>::new();
    cols.resize(fields.len(), vec![]);

    // TODO: For performance this could cache the results of the hash table lookups in a local
    // array, it's wasteful to perform a lookup for each field for each iteration.
    data.iter().for_each(|x| {
        let mut i = 0;
        for kwd in fields {
            let val = formatters.get(*kwd).unwrap()(x, ctx);
            if i == 0 {
                if let Some(f) = formatters.get("*prefix*") {
                    cols[i].push(f(x, ctx) + &val)
                } else {
                    cols[i].push(val)
                }
            } else {
                cols[i].push(val)
            }
            i += 1;
        }
    });

    if opts.csv {
        format_csv(output, fields, opts, cols);
    } else if opts.json {
        format_json(output, fields, opts, cols);
    } else {
        format_fixed_width(output, fields, opts, cols);
    }
}

fn format_fixed_width<'a>(
    output: &mut dyn io::Write,
    fields: &[&'a str],
    opts: &FormatOptions,
    cols: Vec<Vec<String>>,
) {
    // The column width is the max across all the entries in the column (including header,
    // if present).  If there's a tag, it is printed in the last column.
    let mut widths = vec![];
    widths.resize(fields.len() + if opts.tag.is_some() { 1 } else { 0 }, 0);

    if opts.header {
        let mut i = 0;
        for kwd in fields {
            widths[i] = usize::max(widths[i], kwd.len());
            i += 1;
        }
        if opts.tag.is_some() {
            widths[i] = usize::max(widths[i], "tag".len());
        }
    }

    let mut row = 0;
    while row < cols[0].len() {
        let mut col = 0;
        while col < fields.len() {
            widths[col] = usize::max(widths[col], cols[col][row].len());
            col += 1;
        }
        if let Some(ref tag) = opts.tag {
            widths[col] = usize::max(widths[col], tag.len());
        }
        row += 1;
    }

    // Header
    if opts.header {
        let mut i = 0;
        let mut s = "".to_string();
        for kwd in fields {
            let w = widths[i];
            s += format!("{:w$}  ", kwd).as_str();
            i += 1;
        }
        if opts.tag.is_some() {
            let w = widths[i];
            s += format!("{:w$}  ", "tag").as_str();
        }
        output.write(s.trim_end().as_bytes()).unwrap();
        output.write(b"\n").unwrap();
    }

    // Body
    let mut row = 0;
    while row < cols[0].len() {
        let mut col = 0;
        let mut s = "".to_string();
        while col < fields.len() {
            let w = widths[col];
            s += format!("{:w$}  ", cols[col][row]).as_str();
            col += 1;
        }
        if let Some(ref tag) = opts.tag {
            let w = widths[col];
            s += format!("{:w$}  ", tag).as_str();
        }
        output.write(s.trim_end().as_bytes()).unwrap();
        output.write(b"\n").unwrap();
        row += 1;
    }
}

fn format_csv<'a>(
    output: &mut dyn io::Write,
    fields: &[&'a str],
    opts: &FormatOptions,
    cols: Vec<Vec<String>>,
) {
    let mut writer = csv::Writer::from_writer(output);

    if opts.header {
        let mut out_fields = Vec::new();
        for kwd in fields {
            out_fields.push(kwd.to_string());
        }
        if opts.tag.is_some() {
            out_fields.push("tag".to_string());
        }
        writer.write_record(out_fields).unwrap();
    }

    let mut row = 0;
    while row < cols[0].len() {
        let mut out_fields = Vec::new();
        let mut col = 0;
        while col < fields.len() {
            if opts.named {
                out_fields.push(format!("{}={}", fields[col], cols[col][row]));
            } else {
                out_fields.push(format!("{}", cols[col][row]));
            }
            col += 1;
        }
        if let Some(ref tag) = opts.tag {
            if opts.named {
                out_fields.push(format!("tag={tag}"));
            } else {
                out_fields.push(tag.clone());
            }
        }
        writer.write_record(out_fields).unwrap();
        row += 1;
    }

    writer.flush().unwrap();
}

fn format_json<'a>(
    output: &mut dyn io::Write,
    fields: &[&'a str],
    opts: &FormatOptions,
    cols: Vec<Vec<String>>,
) {
    let mut row = 0;
    let mut objects = vec![];
    while row < cols[0].len() {
        let mut obj = json::JsonValue::new_object();
        let mut col = 0;
        while col < fields.len() {
            obj[fields[col]] = cols[col][row].clone().into();
            col += 1;
        }
        if let Some(ref tag) = opts.tag {
            obj["tag"] = tag.to_string().into();
        }
        objects.push(obj);
        row += 1;
    }
    output.write(json::stringify(objects).as_bytes()).unwrap();
}

