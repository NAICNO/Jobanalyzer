/// Simple parser / preprocessor / input filterer for the Sonar log file format.
///
/// For the definition of the input file format, see the README.md on the sonar repo.
///
/// NOTE:
///
/// - Tagged and untagged records can be mixed in a file in any order; this allows files to be
///   catenated and sonar to be updated at any time.
///
/// - It's an important feature of this program that a corrupted record is dropped silently.  (We
///   can add a switch to be noisy about it if that is useful for interactive log testing.)  The
///   reason is that appending-to-log is not atomic wrt reading-from-log and it is somewhat likely
///   that there will be situations where the analysis code runs into a partly-written
///   (corrupted-looking) record.
///
/// - There's an assumption here that if the CSV decoder encounters illegal UTF8 - or for that
///   matter any other parse error, but bad UTF8 is a special case - it will make progress to the
///   end of the record anyway (as CSV is line-oriented).  This is a reasonable assumption but I've
///   found no documentation that guarantees it.
use crate::{GpuStatus, LogEntry};

use anyhow::Result;
use rustutils::{
    empty_gpuset, gpuset_from_bitvector, gpuset_from_list, parse_timestamp, CsvToken, CsvTokenizer,
    GpuSet, Timestamp, CSV_EQ_SENTINEL,
};
use std::boxed::Box;
use std::str::FromStr;
use ustr::Ustr;

/// A sensible "zero" LogEntry for use when we need it.  The user name and command are "_zero_" so
/// that we can recognize this weird LogEntry as intentional and not some mistake.

pub fn empty_logentry(t: Timestamp, hostname: Ustr) -> Box<LogEntry> {
    Box::new(LogEntry {
        major: 0,
        minor: 0,
        bugfix: 0,
        timestamp: t,
        hostname,
        num_cores: 0,
        memtotal_gb: 0.0,
        user: Ustr::from("_zero_"),
        pid: 0,
        job_id: 0,
        command: Ustr::from("_zero_"),
        cpu_pct: 0.0,
        mem_gb: 0.0,
        rssanon_gb: 0.0,
        gpus: empty_gpuset(),
        gpu_pct: 0.0,
        gpumem_pct: 0.0,
        gpumem_gb: 0.0,
        gpu_status: GpuStatus::Ok,
        cputime_sec: 0.0,
        rolledup: 0,
        cpu_util_pct: 0.0,
    })
}

/// Parse a version string.  Avoid allocation here, we parse *a lot* of these.

pub fn parse_version(v1: &str) -> (u16, u16, u16) {
    let mut major = 0u16;
    let mut minor = 0u16;
    let mut bugfix = 0u16;
    if let Some(p1) = v1.find('.') {
        major = v1[0..p1].parse::<u16>().unwrap();
        let v2 = &v1[p1 + 1..];
        if let Some(p2) = v2.find('.') {
            minor = v2[0..p2].parse::<u16>().unwrap();
            let v3 = &v2[p2 + 1..];
            bugfix = v3.parse::<u16>().unwrap();
        }
    }
    (major, minor, bugfix)
}

/// Parse a log file into a set of LogEntry structures, and append to `entries` in the order
/// encountered.  Entries are boxed so that later processing won't copy these increasingly large
/// structures all the time.  Return an error in the case of I/O errors, but silently drop records
/// with parse errors.  Returns the number of discarded records.
///
/// TODO: This should possibly take a Path, not a &str filename.  See comments in logtree.rs.
///
/// TODO: Use Ustr to avoid allocating lots and lots of duplicate strings, both here and elsewhere.

pub fn parse_logfile(file_name: &str, entries: &mut Vec<Box<LogEntry>>) -> Result<usize> {
    let mut file = std::fs::File::open(file_name)?;
    let mut tokenizer = CsvTokenizer::new(&mut file);
    let mut discarded: usize = 0;
    let mut end_of_input = false;

    #[derive(PartialEq)]
    enum Format {
        Unknown,
        Untagged,
        Tagged,
    }

    'line_loop: while !end_of_input
    /* every line */
    {
        // Find the fields and then convert them.  Duplicates are not allowed.  Mandatory
        // fields are really required.
        let mut version: Option<(u16, u16, u16)> = None;
        let mut timestamp: Option<Timestamp> = None;
        let mut hostname: Option<Ustr> = None;
        let mut num_cores: Option<u16> = None;
        let mut memtotal_gb: Option<f32> = None;
        let mut user: Option<Ustr> = None;
        let mut pid: Option<u32> = None;
        let mut job_id: Option<u32> = None;
        let mut command: Option<Ustr> = None;
        let mut cpu_pct: Option<f32> = None;
        let mut mem_gb: Option<f64> = None;
        let mut rssanon_gb: Option<f32> = None;
        let mut gpus: Option<GpuSet> = None;
        let mut gpu_pct: Option<f32> = None;
        let mut gpumem_pct: Option<f32> = None;
        let mut gpumem_gb: Option<f64> = None;
        let mut gpu_status: Option<GpuStatus> = None;
        let mut cputime_sec: Option<f64> = None;
        let mut rolledup: Option<u32> = None;
        let mut untagged_position = 0;
        let mut format = Format::Unknown;
        let mut any_fields = false;

        'field_loop: loop
        /* every field on a line */
        {
            let t0 = tokenizer.get();
            let mut failed = false;
            let mut matched = false;
            match t0 {
                Err(e) => match e.downcast_ref::<std::io::Error>() {
                    Some(_) => {
                        return Err(e);
                    }
                    None => {
                        discarded += 1;
                        continue 'line_loop;
                    }
                },
                Ok(CsvToken::EOL) => {
                    break 'field_loop;
                }
                Ok(CsvToken::EOF) => {
                    end_of_input = true;
                    break 'field_loop;
                }
                Ok(CsvToken::Field(start, lim, eqloc)) => {
                    any_fields = true;
                    if format == Format::Unknown {
                        format = if eqloc == CSV_EQ_SENTINEL {
                            Format::Untagged
                        } else {
                            Format::Tagged
                        };
                    }
                    match format {
                        Format::Unknown => {
                            panic!("Unexpected");
                        }
                        Format::Untagged => {
                            // This is an untagged record.  It does not carry a version number and
                            // has evolved a bit over time.
                            //
                            // Old old format (current on Saga as of 2023-10-13)
                            // 0  timestamp
                            // 1  hostname
                            // 2  numcores
                            // 3  username
                            // 4  jobid
                            // 5  command
                            // 6  cpu_pct
                            // 7  mem_kib
                            //
                            // New old format (what was briefly deployed on the UiO ML nodes)
                            // 8  gpus bitvector
                            // 9  gpu_pct
                            // 10 gpumem_pct
                            // 11 gpumem_kib
                            //
                            // Newer old format (again briefly used on the UiO ML nodes)
                            // 12 cputime_sec

                            version = Some((0u16, 6u16, 0u16));
                            match untagged_position {
                                0 => match parse_timestamp(tokenizer.get_str(start, lim)) {
                                    Ok(t) => {
                                        timestamp = Some(t);
                                        matched = true;
                                    }
                                    Err(_) => {
                                        failed = true;
                                    }
                                },
                                1 => {
                                    hostname = Some(Ustr::from(tokenizer.get_str(start, lim)));
                                    matched = true;
                                }
                                2 => {
                                    (num_cores, failed) = get_u16(tokenizer.get_str(start, lim));
                                    matched = true;
                                }
                                3 => {
                                    user = Some(Ustr::from(tokenizer.get_str(start, lim)));
                                    matched = true;
                                }
                                4 => {
                                    (job_id, failed) = get_u32(tokenizer.get_str(start, lim));
                                    // Untagged data do not carry a PID, so use the job ID in its
                                    // place.  This should be mostly OK.  However, sometimes the job
                                    // ID is also zero, for root jobs.  Client code needs to either
                                    // filter those records or handle the problem.
                                    pid = job_id;
                                    matched = true;
                                }
                                5 => {
                                    command = Some(Ustr::from(tokenizer.get_str(start, lim)));
                                    matched = true;
                                }
                                6 => {
                                    (cpu_pct, failed) = get_f32(tokenizer.get_str(start, lim), 1.0);
                                    matched = true;
                                }
                                7 => {
                                    (mem_gb, failed) = get_f64(
                                        tokenizer.get_str(start, lim),
                                        1.0 / (1024.0 * 1024.0),
                                    );
                                    matched = true;
                                }
                                8 => {
                                    (gpus, failed) =
                                        gpuset_from_bitvector(tokenizer.get_str(start, lim));
                                    matched = true;
                                }
                                9 => {
                                    (gpu_pct, failed) = get_f32(tokenizer.get_str(start, lim), 1.0);
                                    matched = true;
                                }
                                10 => {
                                    (gpumem_pct, failed) =
                                        get_f32(tokenizer.get_str(start, lim), 1.0);
                                    matched = true;
                                }
                                11 => {
                                    (gpumem_gb, failed) = get_f64(
                                        tokenizer.get_str(start, lim),
                                        1.0 / (1024.0 * 1024.0),
                                    );
                                    matched = true;
                                }
                                12 => {
                                    (cputime_sec, failed) =
                                        get_f64(tokenizer.get_str(start, lim), 1.0);
                                    matched = true;
                                }
                                _ => {
                                    // Drop the field, we may learn about it later
                                    matched = true;
                                }
                            }
                            untagged_position += 1;
                        }
                        Format::Tagged => {
                            if eqloc == CSV_EQ_SENTINEL {
                                // Invalid field syntax: Drop the record on the floor
                                discarded += 1;
                                continue 'field_loop;
                            }

                            // The first two characters will always be present because eqloc >= 1.

                            match tokenizer.buf_at(start) {
                                b'c' => match tokenizer.buf_at(start + 1) {
                                    b'm' => {
                                        if tokenizer.match_tag(b"cmd", start, eqloc)
                                            && command.is_none()
                                        {
                                            command =
                                                Some(Ustr::from(tokenizer.get_str(eqloc, lim)));
                                            matched = true;
                                        }
                                    }
                                    b'o' => {
                                        if tokenizer.match_tag(b"cores", start, eqloc)
                                            && num_cores.is_none()
                                        {
                                            (num_cores, failed) =
                                                get_u16(tokenizer.get_str(eqloc, lim));
                                            matched = true;
                                        }
                                    }
                                    b'p' => {
                                        let field = tokenizer.get_str(eqloc, lim);
                                        if tokenizer.match_tag(b"cpu%", start, eqloc)
                                            && cpu_pct.is_none()
                                        {
                                            (cpu_pct, failed) = get_f32(field, 1.0);
                                            matched = true;
                                        } else if tokenizer.match_tag(b"cpukib", start, eqloc)
                                            && mem_gb.is_none()
                                        {
                                            (mem_gb, failed) =
                                                get_f64(field, 1.0 / (1024.0 * 1024.0));
                                            matched = true;
                                        } else if tokenizer.match_tag(b"cputime_sec", start, eqloc)
                                            && cputime_sec.is_none()
                                        {
                                            (cputime_sec, failed) = get_f64(field, 1.0);
                                            matched = true;
                                        }
                                    }
                                    _ => {}
                                },
                                b'g' => {
                                    let field = tokenizer.get_str(eqloc, lim);
                                    if tokenizer.match_tag(b"gpus", start, eqloc) && gpus.is_none()
                                    {
                                        (gpus, failed) = gpuset_from_list(field);
                                        matched = true;
                                    } else if tokenizer.match_tag(b"gpu%", start, eqloc)
                                        && gpu_pct.is_none()
                                    {
                                        (gpu_pct, failed) = get_f32(field, 1.0);
                                        matched = true;
                                    } else if tokenizer.match_tag(b"gpumem%", start, eqloc)
                                        && gpumem_pct.is_none()
                                    {
                                        (gpumem_pct, failed) = get_f32(field, 1.0);
                                        matched = true;
                                    } else if tokenizer.match_tag(b"gpukib", start, eqloc)
                                        && gpumem_gb.is_none()
                                    {
                                        (gpumem_gb, failed) =
                                            get_f64(field, 1.0 / (1024.0 * 1024.0));
                                        matched = true;
                                    } else if tokenizer.match_tag(b"gpufail", start, eqloc)
                                        && gpu_status.is_none()
                                    {
                                        let val;
                                        (val, failed) = get_u32(field);
                                        if !failed {
                                            match val {
                                                Some(0u32) => {}
                                                Some(1u32) => {
                                                    gpu_status = Some(GpuStatus::UnknownFailure)
                                                }
                                                _ => gpu_status = Some(GpuStatus::UnknownFailure),
                                            }
                                            matched = true;
                                        }
                                    }
                                }
                                b'h' => {
                                    if tokenizer.match_tag(b"host", start, eqloc)
                                        && hostname.is_none()
                                    {
                                        hostname = Some(Ustr::from(tokenizer.get_str(eqloc, lim)));
                                        matched = true;
                                    }
                                }
                                b'j' => {
                                    if tokenizer.match_tag(b"job", start, eqloc) && job_id.is_none()
                                    {
                                        (job_id, failed) = get_u32(tokenizer.get_str(eqloc, lim));
                                        matched = true;
                                    }
                                }
                                b'm' => {
                                    if tokenizer.match_tag(b"memtotalkib", start, eqloc)
                                        && memtotal_gb.is_none()
                                    {
                                        (memtotal_gb, failed) = get_f32(
                                            tokenizer.get_str(eqloc, lim),
                                            1.0 / (1024.0 * 1024.0),
                                        );
                                        matched = true;
                                    }
                                }
                                b'p' => {
                                    if tokenizer.match_tag(b"pid", start, eqloc) && pid.is_none() {
                                        (pid, failed) = get_u32(tokenizer.get_str(eqloc, lim));
                                        matched = true;
                                    }
                                }
                                b'r' => {
                                    let field = tokenizer.get_str(eqloc, lim);
                                    match tokenizer.buf_at(start + 1) {
                                        b's' => {
                                            if tokenizer.match_tag(b"rssanonkib", start, eqloc)
                                                && rssanon_gb.is_none()
                                            {
                                                (rssanon_gb, failed) =
                                                    get_f32(field, 1.0 / (1024.0 * 1024.0));
                                                matched = true;
                                            }
                                        }
                                        b'o' => {
                                            if tokenizer.match_tag(b"rolledup", start, eqloc)
                                                && rolledup.is_none()
                                            {
                                                (rolledup, failed) = get_u32(field);
                                                matched = true;
                                            }
                                        }
                                        _ => {}
                                    }
                                }
                                b't' => {
                                    if tokenizer.match_tag(b"time", start, eqloc)
                                        && timestamp.is_none()
                                    {
                                        if let Ok(t) =
                                            parse_timestamp(tokenizer.get_str(eqloc, lim))
                                        {
                                            timestamp = Some(t);
                                            matched = true;
                                        } else {
                                            failed = true;
                                        }
                                    }
                                }
                                b'u' => {
                                    if tokenizer.match_tag(b"user", start, eqloc) && user.is_none()
                                    {
                                        user = Some(Ustr::from(tokenizer.get_str(eqloc, lim)));
                                        matched = true;
                                    }
                                }
                                b'v' => {
                                    if tokenizer.match_tag(b"v", start, eqloc) && version.is_none()
                                    {
                                        version =
                                            Some(parse_version(tokenizer.get_str(eqloc, lim)));
                                        matched = true;
                                    }
                                }
                                _ => {
                                    // Unknown field, ignore it silently, this is benign (mostly - it
                                    // could be a field whose tag was chopped off, so maybe we should
                                    // look for `=`).
                                    matched = true;
                                }
                            }
                        }
                    }
                }
            }
            // Four cases:
            //
            //   matched && !failed - field matched a tag, value is good
            //   matched && failed - field matched a tag, value is bad
            //   !matched && !failed - field did not match any tag
            //   !matched && failed - impossible
            //
            // The second case suggests something bad, so discard the record in this case.  Note
            // this is actually the same as just `failed` due to the fourth case.
            if matched && failed {
                discarded += 1;
                continue 'line_loop;
            }
        } // Field loop

        if end_of_input && !any_fields {
            break 'line_loop;
        }

        // Check that untagged records have a sensible number of fields.

        if format == Format::Untagged
            && untagged_position != 8
            && untagged_position != 12
            && untagged_position != 13
        {
            discarded += 1;
            continue 'line_loop;
        }

        // Check that mandatory fields are present.

        if version.is_none()
            || timestamp.is_none()
            || hostname.is_none()
            || user.is_none()
            || command.is_none()
        {
            discarded += 1;
            continue 'line_loop;
        }

        // Fill in default data for optional fields.

        if num_cores.is_none() {
            num_cores = Some(0u16);
        }
        if memtotal_gb.is_none() {
            memtotal_gb = Some(0.0);
        }
        if job_id.is_none() {
            job_id = Some(0);
        }
        if pid.is_none() {
            pid = Some(0);
        }
        if cpu_pct.is_none() {
            cpu_pct = Some(0.0);
        }
        if mem_gb.is_none() {
            mem_gb = Some(0.0);
        }
        if rssanon_gb.is_none() {
            rssanon_gb = Some(0.0);
        }
        if gpus.is_none() {
            gpus = Some(empty_gpuset());
        }
        if gpu_pct.is_none() {
            gpu_pct = Some(0.0)
        }
        if gpumem_pct.is_none() {
            gpumem_pct = Some(0.0)
        }
        if gpumem_gb.is_none() {
            gpumem_gb = Some(0.0)
        }
        if gpu_status.is_none() {
            gpu_status = Some(GpuStatus::Ok)
        }
        if cputime_sec.is_none() {
            cputime_sec = Some(0.0);
        }
        if rolledup.is_none() {
            rolledup = Some(0);
        }

        // Ship it!

        let (major, minor, bugfix) = version.unwrap();
        entries.push(Box::new(LogEntry {
            major,
            minor,
            bugfix,
            timestamp: timestamp.unwrap(),
            hostname: hostname.unwrap(),
            num_cores: num_cores.unwrap(),
            memtotal_gb: memtotal_gb.unwrap(),
            user: user.unwrap(),
            pid: pid.unwrap(),
            job_id: job_id.unwrap(),
            command: command.unwrap(),
            cpu_pct: cpu_pct.unwrap(),
            mem_gb: mem_gb.unwrap(),
            rssanon_gb: rssanon_gb.unwrap(),
            gpus: gpus.unwrap(),
            gpu_pct: gpu_pct.unwrap(),
            gpumem_pct: gpumem_pct.unwrap(),
            gpumem_gb: gpumem_gb.unwrap(),
            gpu_status: gpu_status.unwrap(),
            cputime_sec: cputime_sec.unwrap(),
            rolledup: rolledup.unwrap(),
            // Computed fields
            cpu_util_pct: 0.0,
        }));
    } // Line loop

    Ok(discarded)
}

fn get_u32(s: &str) -> (Option<u32>, bool) {
    if let Ok(n) = u32::from_str(s) {
        (Some(n), false)
    } else {
        (None, true)
    }
}

fn get_u16(s: &str) -> (Option<u16>, bool) {
    if let Ok(n) = u16::from_str(s) {
        (Some(n), false)
    } else {
        (None, true)
    }
}

fn get_f64(s: &str, scale: f64) -> (Option<f64>, bool) {
    if let Ok(n) = f64::from_str(s) {
        if f64::is_infinite(n) {
            (None, true)
        } else {
            (Some(n * scale), false)
        }
    } else {
        (None, true)
    }
}

fn get_f32(s: &str, scale: f32) -> (Option<f32>, bool) {
    if let Ok(n) = f32::from_str(s) {
        if f32::is_infinite(n) {
            (None, true)
        } else {
            (Some(n * scale), false)
        }
    } else {
        (None, true)
    }
}
