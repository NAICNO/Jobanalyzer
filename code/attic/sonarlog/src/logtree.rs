/// Enumerate log files in a log tree; read sets of files.
use crate::{parse_logfile, LogEntry};

use anyhow::{bail, Result};
use core::cmp::{max, min};
use rustutils::{date_range, HostGlobber, Timestamp};
use std::boxed::Box;
use std::collections::HashMap;
use std::path;
use ustr::Ustr;

#[cfg(test)]
use rustutils::timestamp_from_ymd;

/// Create a set of plausible log file names within a directory tree, for a date range and a set of
/// included host files.  The returned names are sorted lexicographically.
///
/// `data_path` is the root of the tree.  `hostnames`, if not the empty set, is the set of hostnames
/// we will want data for.  `from` and `to` express the inclusive date range for the records we will
/// want to consider.
///
/// This returns an error if the `data_path` does not name a directory or if any directory that is
/// considered in the subtree, and which exists, cannot be read.
///
/// It does not return an error if the csv files cannot be read; that has to be handled later.
///
/// File names that are not representable as UTF8 are ignored.
///
/// The expected log format is this:
///
///    let file_name = format!("{}/{}/{}/{}/{}.csv", data_path, year, month, day, hostname);
///
/// where year is CE and month and day have leading zeroes if necessary, ie, these are split out
/// from a standard ISO timestamp.
///
/// We loop across dates in the tree below `data_path` and for each `hostname`.csv file, we check if
/// it names an included host name.
///
/// TODO: Cleaner would be for this to return Result<Vec<path::Path>>, and not do all this string
/// stuff.  Indeed we could require the caller to provide data_path as a Path.

pub fn find_logfiles(
    data_path: &str,
    hostnames: &HostGlobber,
    from: Timestamp,
    to: Timestamp,
) -> Result<Vec<String>> {
    if !path::Path::new(data_path).is_dir() {
        // Path redacted so as not to reveal secrets
        bail!("Not a viable data directory");
    }

    let mut filenames = vec![];
    for (year, month, day) in date_range(from, to) {
        let dir_name = format!("{}/{}/{:02}/{:02}", data_path, year, month, day);
        let p = std::path::Path::new(&dir_name);
        if p.is_dir() {
            let rd = p.read_dir()?;
            for entry in rd {
                if entry.is_err() {
                    // Bad directory entries are ignored, though these would probably be I/O errors.
                    // Note there is an assumption here that forward progress is guaranteed despite
                    // the error.  This is not properly documented but the example for the read_dir
                    // iterator in the rust docs assumes it as well.
                    continue;
                }
                let p = entry.unwrap().path();
                let pstr = p.to_str();
                if pstr.is_none() {
                    // Non-UTF8 paths are ignored.  The `data_path` is a string, hence UTF8, and we
                    // construct only UTF8 names, and host names are UTF8.  Hence non-UTF8 names
                    // will never match what we're looking for.
                    continue;
                }
                let ext = p.extension();
                if ext.is_none() || ext.unwrap() != "csv" {
                    // Non-csv files are ignored
                    continue;
                }
                if hostnames.is_empty() {
                    // If there's no hostname filter then keep the path
                    filenames.push(pstr.unwrap().to_string());
                    continue;
                }
                let h = p.file_stem();
                if h.is_none() {
                    // Log file names have to have a stem even if there is no host name filter.
                    // TODO: Kind of debatable actually.
                    continue;
                }
                let stem = h.unwrap().to_str().unwrap();
                // Filter the stem against the host names.
                if hostnames.match_hostname(stem) {
                    filenames.push(pstr.unwrap().to_string());
                    continue;
                }
            }
        }
    }
    filenames.sort();
    Ok(filenames)
}

/// Timebounds is a map from host name to earliest and latest time seen for the host.  For
/// synthesized (merged) hosts it maps the merged host name to the earliest and latest times seen
/// for all the merged streams for all those hosts.

pub type Timebounds = HashMap<Ustr, Timebound>;

#[derive(Clone)]
pub struct Timebound {
    pub earliest: Timestamp,
    pub latest: Timestamp,
}

/// Read all the files into an array and return some basic information about the data.
///
/// Returns error on I/O error and discards illegal records silently, but returns the number of
/// records discarded.

pub fn read_logfiles(logfiles: &[String]) -> Result<(Vec<Box<LogEntry>>, Timebounds, usize)> {
    let mut entries = Vec::<Box<LogEntry>>::new();
    let mut discarded: usize = 0;

    // Read all the files
    for file in logfiles {
        discarded += parse_logfile(file, &mut entries)?;
    }

    let mut bounds = Timebounds::new();
    for e in &entries {
        let h = e.hostname;
        if let Some(v) = bounds.get_mut(&h) {
            v.earliest = min(v.earliest, e.timestamp);
            v.latest = max(v.latest, e.timestamp);
        } else {
            bounds.insert(
                h,
                Timebound {
                    earliest: e.timestamp,
                    latest: e.timestamp,
                },
            );
        }
    }
    Ok((entries, bounds, discarded))
}

#[test]
fn test_find_logfiles1() {
    // Use the precise date bounds for the files in the directory to test that we get exactly the
    // expected files.  This will encounter non-csv files, which should not be listed.
    let hosts = HostGlobber::new(true);
    let xs = find_logfiles(
        "../../tests/sonarlog/whitebox-tree",
        &hosts,
        timestamp_from_ymd(2023, 5, 30),
        timestamp_from_ymd(2023, 6, 4),
    )
    .unwrap();
    assert!(xs.eq(&vec![
        "../../tests/sonarlog/whitebox-tree/2023/05/30/a.csv",
        "../../tests/sonarlog/whitebox-tree/2023/05/30/b.csv",
        "../../tests/sonarlog/whitebox-tree/2023/05/31/a.csv",
        "../../tests/sonarlog/whitebox-tree/2023/05/31/b.csv",
        "../../tests/sonarlog/whitebox-tree/2023/06/01/a.csv",
        "../../tests/sonarlog/whitebox-tree/2023/06/01/b.csv",
        "../../tests/sonarlog/whitebox-tree/2023/06/02/a.csv",
        "../../tests/sonarlog/whitebox-tree/2023/06/03/b.csv",
        "../../tests/sonarlog/whitebox-tree/2023/06/04/a.csv",
        "../../tests/sonarlog/whitebox-tree/2023/06/04/b.csv",
    ]));
}

#[test]
fn test_find_logfiles2() {
    // Use early date bounds for both limits to test that we get a subset.  Also this will run
    // into 2023/05/29 which is a file, not a directory.
    let hosts = HostGlobber::new(true);
    let xs = find_logfiles(
        "../../tests/sonarlog/whitebox-tree",
        &hosts,
        timestamp_from_ymd(2023, 5, 29),
        timestamp_from_ymd(2023, 6, 2),
    )
    .unwrap();
    assert!(xs.eq(&vec![
        "../../tests/sonarlog/whitebox-tree/2023/05/30/a.csv",
        "../../tests/sonarlog/whitebox-tree/2023/05/30/b.csv",
        "../../tests/sonarlog/whitebox-tree/2023/05/31/a.csv",
        "../../tests/sonarlog/whitebox-tree/2023/05/31/b.csv",
        "../../tests/sonarlog/whitebox-tree/2023/06/01/a.csv",
        "../../tests/sonarlog/whitebox-tree/2023/06/01/b.csv",
        "../../tests/sonarlog/whitebox-tree/2023/06/02/a.csv",
    ]));
}

#[test]
fn test_find_logfiles3() {
    // Filter by host name.
    let mut hosts = HostGlobber::new(true);
    hosts.insert("a").unwrap();
    let xs = find_logfiles(
        "../../tests/sonarlog/whitebox-tree",
        &hosts,
        timestamp_from_ymd(2023, 5, 20),
        timestamp_from_ymd(2023, 6, 2),
    )
    .unwrap();
    assert!(xs.eq(&vec![
        "../../tests/sonarlog/whitebox-tree/2023/05/28/a.csv",
        "../../tests/sonarlog/whitebox-tree/2023/05/30/a.csv",
        "../../tests/sonarlog/whitebox-tree/2023/05/31/a.csv",
        "../../tests/sonarlog/whitebox-tree/2023/06/01/a.csv",
        "../../tests/sonarlog/whitebox-tree/2023/06/02/a.csv",
    ]));
}

#[test]
fn test_find_logfiles4() {
    // Nonexistent data_path
    let hosts = HostGlobber::new(true);
    assert!(find_logfiles(
        "../sonar_test_data77",
        &hosts,
        timestamp_from_ymd(2023, 5, 30),
        timestamp_from_ymd(2023, 6, 4)
    )
    .is_err());
}

// Other test cases are black-box, see ../../tests/sonarlog.
