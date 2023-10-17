/// Simple parser / preprocessor / input filterer for the Sonar log file format.
///
/// For the definition of the input file format, see the README.md on the sonar repo.
///
/// NOTE:
///
/// - Tagged and untagged records can be mixed in a file in any order; this allows files to be
///   catenated and sonar to be updated at any time.
///
/// - It's an important feature of this program that a corrupted record is dropped silently.  (We can
///   add a switch to be noisy about it if that is useful for interactive log testing.)  The reason
///   is that appending-to-log is not atomic wrt reading-from-log and it is somewhat likely that
///   there will be situations where the analysis code runs into a partly-written (corrupted-looking)
///   record.
///
/// - There's an assumption here that if the CSV decoder encounters illegal UTF8 - or for that matter
///   any other parse error, but bad UTF8 is a special case - it will make progress to the end of the
///   record anyway (as CSV is line-oriented).  This is a reasonable assumption but I've found no
///   documentation that guarantees it.

use crate::{parse_timestamp, GpuStatus, LogEntry, Timestamp};

use anyhow::Result;
use serde::Deserialize;
use std::boxed::Box;
use std::collections::HashSet;
use std::str::FromStr;

/// The GpuSet has three states:
///
///  - the set is known to be empty, this is Some({})
///  - the set is known to be nonempty and have only known gpus in the set, this is Some({a,b,..})
///  - the set is known to be nonempty but have (some) unknown members, this is None
///
/// During processing, the set starts out as Some({}).  If a device reports "unknown" GPUs then the
/// set can transition from Some({}) to None or from Some({a,b,..}) to None.  Once in the None state,
/// the set will stay in that state.  There is no representation for some known + some unknown GPUs,
/// it is not believed to be worthwhile.

pub type GpuSet = Option<HashSet<u32>>;

pub fn empty_gpuset() -> GpuSet {
    Some(HashSet::new())
}

pub fn is_empty_gpuset(s: &GpuSet) -> bool {
    if let Some(set) = s {
        set.is_empty()
    } else {
        false
    }
}

pub fn unknown_gpuset() -> GpuSet {
    None
}

pub fn is_unknown_gpuset(s: &GpuSet) -> bool {
    s.is_none()
}

pub fn singleton_gpuset(maybe_device: Option<u32>) -> GpuSet {
    if let Some(dev) = maybe_device {
        let mut gpus = HashSet::new();
        gpus.insert(dev);
        Some(gpus)
    } else {
        None
    }
}

pub fn adjoin_gpuset(lhs: &mut GpuSet, rhs: u32) {
    if let Some(gpus) = lhs {
        gpus.insert(rhs);
    }
}

pub fn union_gpuset(lhs: &mut GpuSet, rhs: &GpuSet) {
    if lhs.is_none() {
        // The result is also None
    } else if rhs.is_none() {
        *lhs = None;
    } else {
        lhs.as_mut().unwrap().extend(rhs.as_ref().unwrap());
    }
}

// For testing, we need a predictable order, so accumulate as numbers and sort
pub fn gpuset_to_string(gpus: &GpuSet) -> String {
    if let Some(gpus) = gpus {
        if gpus.is_empty() {
            "none".to_string()
        } else {
            let mut cards = vec![];
            for g in gpus {
                cards.push(*g);
            }
            cards.sort();
            let mut term = "";
            let mut s = String::new();
            for x in cards {
                s += term;
                term = ",";
                s += &x.to_string();
            }
            s
        }
    } else {
        "unknown".to_string()
    }
}

pub fn merge_gpu_status(lhs: GpuStatus, rhs: GpuStatus) -> GpuStatus {
    match (lhs, rhs) {
        (v, w) if v == w => v,
        (v, GpuStatus::Ok) => v,
        (GpuStatus::Ok, v) => v,
        (_, _) => GpuStatus::UnknownFailure,
    }
}

/// Parse a log file into a set of LogEntry structures, and append to `entries` in the order
/// encountered.  Entries are boxed so that later processing won't copy these increasingly large
/// structures all the time.  Return an error in the case of I/O errors, but silently drop records
/// with parse errors.
///
/// TODO: This should possibly take a Path, not a &str filename.  See comments in logtree.rs.
///
/// TODO: Use Ustr to avoid allocating lots and lots of duplicate strings, both here and elsewhere.

pub fn parse_logfile(file_name: &str, entries: &mut Vec<Box<LogEntry>>) -> Result<()> {
    #[derive(Debug, Deserialize)]
    struct LogRecord {
        fields: Vec<String>,
    }

    // An error here is going to be an I/O error so always propagate it.
    let mut reader = csv::ReaderBuilder::new()
        .has_headers(false)
        .flexible(true)
        .from_path(file_name)?;

    'outer: for deserialized_record in reader.deserialize::<LogRecord>() {
        match deserialized_record {
            Err(e) => {
                if e.is_io_error() {
                    return Err(e.into());
                }
                // Otherwise drop the record
                continue 'outer;
            }
            Ok(record) => {
                // Find the fields and then convert them.  Duplicates are not allowed.  Mandatory
                // fields are really required.
                let mut version: Option<String> = None;
                let mut timestamp: Option<Timestamp> = None;
                let mut hostname: Option<String> = None;
                let mut num_cores: Option<u32> = None;
                let mut user: Option<String> = None;
                let mut pid: Option<u32> = None;
                let mut job_id: Option<u32> = None;
                let mut command: Option<String> = None;
                let mut cpu_pct: Option<f64> = None;
                let mut mem_gb: Option<f64> = None;
                let mut gpus: Option<GpuSet> = None;
                let mut gpu_pct: Option<f64> = None;
                let mut gpumem_pct: Option<f64> = None;
                let mut gpumem_gb: Option<f64> = None;
                let mut gpu_status: Option<GpuStatus> = None;
                let mut cputime_sec: Option<f64> = None;
                let mut rolledup: Option<u32> = None;

                if let Ok(t) = parse_timestamp(&record.fields[0]) {
                    // This is an untagged record.  It does not carry a version number and has
                    // evolved a bit over time.
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

                    if cfg!(feature = "untagged_sonar_data") {
                        if record.fields.len() != 8 && record.fields.len() != 12 && record.fields.len() != 13 {
                            continue 'outer;
                        }
                        let mut failed;
                        version = Some("0.6.0".to_string());
                        timestamp = Some(t);
                        hostname = Some(record.fields[1].to_string());
                        (num_cores, failed) = get_u32(&record.fields[2]);
                        if failed {
                            continue 'outer;
                        }
                        user = Some(record.fields[3].to_string());
                        (job_id, failed) = get_u32(&record.fields[4]);
                        if failed {
                            continue 'outer;
                        }
                        // Untagged data do not carry a PID, so use the job ID in its place.  This
                        // should be mostly OK.  However, sometimes the job ID is also zero, for
                        // root jobs.  Client code needs to either filter those records or handle
                        // the problem.
                        pid = job_id;
                        command = Some(record.fields[5].to_string());
                        (cpu_pct, failed) = get_f64(&record.fields[6], 1.0);
                        if failed {
                            continue 'outer;
                        }
                        (mem_gb, failed) = get_f64(&record.fields[7], 1.0 / (1024.0 * 1024.0));
                        if failed {
                            continue 'outer;
                        }
                        if record.fields.len() == 12 || record.fields.len() == 13 {
                            (gpus, failed) = get_gpus_from_bitvector(&record.fields[8]);
                            if failed {
                                continue 'outer;
                            }
                            (gpu_pct, failed) = get_f64(&record.fields[9], 1.0);
                            if failed {
                                continue 'outer;
                            }
                            (gpumem_pct, failed) = get_f64(&record.fields[10], 1.0);
                            if failed {
                                continue 'outer;
                            }
                            (gpumem_gb, failed) = get_f64(&record.fields[11], 1.0 / (1024.0 * 1024.0));
                            if failed {
                                continue 'outer;
                            }
                            if record.fields.len() == 13 {
                                (cputime_sec, failed) = get_f64(&record.fields[12], 1.0);
                                if failed {
                                    continue 'outer;
                                }
                            }
                        }
                    } else {
                        // Drop the record on the floor
                        continue 'outer;
                    }
                } else {
                    // This must be a tagged record
                    for field in record.fields {
                        // TODO: Performance: Would it be better to extract the keyword, hash
                        // it, extract a code for it from a hash table, and then switch on that?
                        // It's bad either way.  Or we could run a state machine across the
                        // string here, that would likely be best.
                        let mut failed = false;
                        if field.starts_with("v=") {
                            if version.is_some() {
                                continue 'outer;
                            }
                            version = Some(field[2..].to_string())
                        } else if field.starts_with("time=") {
                            if timestamp.is_some() {
                                continue 'outer;
                            }
                            if let Ok(t) = parse_timestamp(&field[5..]) {
                                timestamp = Some(t.into());
                            } else {
                                continue 'outer;
                            }
                        } else if field.starts_with("host=") {
                            if hostname.is_some() {
                                continue 'outer;
                            }
                            hostname = Some(field[5..].to_string())
                        } else if field.starts_with("cores=") {
                            if num_cores.is_some() {
                                continue 'outer;
                            }
                            (num_cores, failed) = get_u32(&field[6..]);
                        } else if field.starts_with("user=") {
                            if user.is_some() {
                                continue 'outer;
                            }
                            user = Some(field[5..].to_string())
                        } else if field.starts_with("pid=") {
                            if pid.is_some() {
                                continue 'outer;
                            }
                            (pid, failed) = get_u32(&field[4..]);
                        } else if field.starts_with("job=") {
                            if job_id.is_some() {
                                continue 'outer;
                            }
                            (job_id, failed) = get_u32(&field[4..]);
                        } else if field.starts_with("cmd=") {
                            if command.is_some() {
                                continue 'outer;
                            }
                            command = Some(field[4..].to_string())
                        } else if field.starts_with("cpu%=") {
                            if cpu_pct.is_some() {
                                continue 'outer;
                            }
                            (cpu_pct, failed) = get_f64(&field[5..], 1.0);
                        } else if field.starts_with("cpukib=") {
                            if mem_gb.is_some() {
                                continue 'outer;
                            }
                            (mem_gb, failed) = get_f64(&field[7..], 1.0 / (1024.0 * 1024.0));
                        } else if field.starts_with("gpus=") {
                            if gpus.is_some() {
                                continue 'outer;
                            }
                            (gpus, failed) = get_gpus_from_list(&field[5..]);
                        } else if field.starts_with("gpu%=") {
                            if gpu_pct.is_some() {
                                continue 'outer;
                            }
                            (gpu_pct, failed) = get_f64(&field[5..], 1.0);
                        } else if field.starts_with("gpumem%=") {
                            if gpumem_pct.is_some() {
                                continue 'outer;
                            }
                            (gpumem_pct, failed) = get_f64(&field[8..], 1.0);
                        } else if field.starts_with("gpukib=") {
                            if gpumem_gb.is_some() {
                                continue 'outer;
                            }
                            (gpumem_gb, failed) = get_f64(&field[7..], 1.0 / (1024.0 * 1024.0));
                        } else if field.starts_with("gpufail=") {
                            if gpu_status.is_some() {
                                continue 'outer;
                            }
                            let val;
                            (val, failed) = get_u32(&field[8..]);
                            if !failed {
                                match val {
                                    Some(0u32) => {}
                                    Some(1u32) => { gpu_status = Some(GpuStatus::UnknownFailure) }
                                    _ => { gpu_status = Some(GpuStatus::UnknownFailure) }
                                }
                            }
                        } else if field.starts_with("cputime_sec=") {
                            if cputime_sec.is_some() {
                                continue 'outer;
                            }
                            (cputime_sec, failed) = get_f64(&field[12..], 1.0);
                        } else if field.starts_with("rolledup=") {
                            if rolledup.is_some() {
                                continue 'outer;
                            }
                            (rolledup, failed) = get_u32(&field[9..]);
                        } else {
                            // Unknown field, ignore it silently, this is benign (mostly - it could
                            // be a field whose tag was chopped off, so maybe we should look for
                            // `=`).
                        }
                        if failed {
                            continue 'outer;
                        }
                    }
                }

                // Check that mandatory fields are present.

                if version.is_none()
                    || timestamp.is_none()
                    || hostname.is_none()
                    || user.is_none()
                    || command.is_none()
                {
                    continue 'outer;
                }

                // Fill in default data for optional fields.

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

                entries.push(Box::new(LogEntry {
                    version: version.unwrap(),
                    timestamp: timestamp.unwrap(),
                    hostname: hostname.unwrap(),
                    num_cores: num_cores.unwrap(),
                    user: user.unwrap(),
                    pid: pid.unwrap(),
                    job_id: job_id.unwrap(),
                    command: command.unwrap(),
                    cpu_pct: cpu_pct.unwrap(),
                    mem_gb: mem_gb.unwrap(),
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
            }
        }
    }
    Ok(())
}

/// A sensible "zero" LogEntry for use when we need it.  The user name and command are "_zero_" so
/// that we can recognize this weird LogEntry as intentional and not some mistake.

pub fn empty_logentry(t: Timestamp, hostname: &str) -> Box<LogEntry> {
    Box::new(LogEntry {
        version: "0.0.0".to_string(),
        timestamp: t,
        hostname: hostname.to_string(),
        num_cores: 0,
        user: "_zero_".to_string(),
        pid: 0,
        job_id: 0,
        command: "_zero_".to_string(),
        cpu_pct: 0.0,
        mem_gb: 0.0,
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

fn get_u32(s: &str) -> (Option<u32>, bool) {
    if let Ok(n) = u32::from_str(s) {
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

fn get_gpus_from_bitvector(s: &str) -> (Option<GpuSet>, bool) {
    match usize::from_str_radix(s, 2) {
        Ok(mut bit_mask) => {
            let mut gpus = None;
            if bit_mask != 0 {
                let mut set = HashSet::new();
                if bit_mask != !0usize {
                    let mut shift = 0;
                    while bit_mask != 0 {
                        if (bit_mask & 1) != 0 {
                            set.insert(shift);
                        }
                        shift += 1;
                        bit_mask >>= 1;
                    }
                }
                gpus = Some(set);
            }
            (Some(gpus), false)
        }
        Err(_) => (None, true),
    }
}

// The bool return value is "failed".

fn get_gpus_from_list(s: &str) -> (Option<GpuSet>, bool) {
    if s == "unknown" {
        (Some(unknown_gpuset()), false)
    } else if s == "none" {
        (Some(empty_gpuset()), false)
    } else {
        let mut set = empty_gpuset();
        let vs: std::result::Result<Vec<_>, _> = s.split(',').map(u32::from_str).collect();
        match vs {
            Err(_) => (None, true),
            Ok(vs) => {
                for v in vs {
                    adjoin_gpuset(&mut set, v);
                }
                (Some(set), false)
            }
        }
    }
}

#[test]
fn test_gpuset() {
    assert!(is_empty_gpuset(&empty_gpuset()));
    assert!(!is_empty_gpuset(&unknown_gpuset()));
    assert!(!is_empty_gpuset(&singleton_gpuset(Some(1))));
    let mut s = unknown_gpuset();
    adjoin_gpuset(&mut s, 1);
    assert!(is_unknown_gpuset(&s));
}

#[test]
fn test_get_gpus_from_list() {
    // Much more could be done here
    assert!(get_gpus_from_list("unknownx") == (None, true));
    assert!(get_gpus_from_list("unknown") == (Some(unknown_gpuset()), false));
    assert!(get_gpus_from_list("none") == (Some(empty_gpuset()), false));
    assert!(get_gpus_from_list("1") == (Some(singleton_gpuset(Some(1))), false));
    assert!(get_gpus_from_list("1,1,1") == (Some(singleton_gpuset(Some(1))), false));
    let mut s1 = singleton_gpuset(Some(1));
    adjoin_gpuset(&mut s1, 2);
    assert!(get_gpus_from_list("2,1") == (Some(s1), false));
    let mut s2 = unknown_gpuset();
    adjoin_gpuset(&mut s2, 1);
    assert!(s2 == unknown_gpuset());
}

// Other test cases are black-box, see ../../tests/sonarlog

