// Enumerate log files in a log tree.

// For jobgraph, the expected log format is this:
//
//    let file_name = format!("{}/{}/{}/{}/{}.csv", data_path, year, month, day, hostname);
//
// where year is CE and month and day have leading zeroes if necessary, ie, these are split
// out from a standard ISO timestamp.
//
// We loop across dates in the tree below `data_path` and for each `hostname`.csv file, we check if
// it names an included host name.
//
// TODO: Cleaner would be for find_logfiles to return Result<Vec<path::Path>>, and not do all this
// string stuff.  Indeed we could require the caller to provide data_path as a Path.

use crate::dates;
use anyhow::{bail, Result};
use chrono::prelude::DateTime;
use chrono::Utc;
use std::collections::HashSet;
use std::path;
#[cfg(test)]
use chrono::NaiveDate;

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

pub fn find_logfiles(
    data_path: &str,
    hostnames: &HashSet<String>,
    from: DateTime<Utc>,
    to: DateTime<Utc>,
) -> Result<Vec<String>> {
    if !path::Path::new(data_path).is_dir() {
        bail!("Not a viable log directory: {}", data_path);
    }

    let mut filenames = vec![];
    for (year, month, day) in dates::date_range(from, to) {
        let dir_name = format!("{}/{}/{:02}/{:02}", data_path, year, month, day);
        let p = std::path::Path::new(&dir_name);
        if p.is_dir() {
            let rd = p.read_dir()?;
            for entry in rd {
                if let Err(_) = entry {
                    // Bad directory entries are ignored, though these would probably be I/O errors.
                    // Note there is an assumption here that forward progress is guaranteed despite
                    // the error.  This is not properly documented but the example for the read_dir
                    // iterator in the rust docs assumes it as well.
                    continue
                }
                let p = entry.unwrap().path();
                let pstr = p.to_str();
                if pstr.is_none() {
                    // Non-UTF8 paths are ignored.  The `data_path` is a string, hence UTF8, and we
                    // construct only UTF8 names, and host names are UTF8.  Hence non-UTF8 names
                    // will never match what we're looking for.
                    continue
                }
                let ext = p.extension();
                if ext.is_none() || ext.unwrap() != "csv" {
                    // Non-csv files are ignored
                    continue
                }
                if hostnames.is_empty() {
                    // If there's no hostname filter then keep the path
                    filenames.push(pstr.unwrap().to_string());
                    continue
                }
                let h = p.file_stem();
                if h.is_none() {
                    // Log file names have to have a stem even if there is no host name filter.
                    // TODO: Kind of debatable actually.
                    continue
                }
                let stem = h.unwrap().to_str().unwrap();
                // Filter the stem against the host names.
                //
                // TODO: The stem is usually some FQDN, but `hostnames` may contain just plain host
                // names, eg, `ml8` and not `ml8.hpc.uio.no`, because that's convenient for the
                // user.  So in this case, see if the stem's first element can be split off too.
                if hostnames.contains(stem) {
                    filenames.push(pstr.unwrap().to_string());
                    continue
                }
            }
        }
    }
    filenames.sort();
    Ok(filenames)
}

#[test]
fn test_find_logfiles1() {
    // Use the precise date bounds for the files in the directory to test that we get exactly the
    // expected files.
    let hosts : HashSet<String> = HashSet::new();
    let xs = find_logfiles("../sonar_test_data0",
                           &hosts,
                           DateTime::from_utc(NaiveDate::from_ymd_opt(2023, 5, 30).unwrap().and_hms_opt(0,0,0).unwrap(), Utc),
                           DateTime::from_utc(NaiveDate::from_ymd_opt(2023, 6, 4).unwrap().and_hms_opt(0,0,0).unwrap(), Utc)).unwrap();
    assert!(xs.eq(&vec![
        "../sonar_test_data0/2023/05/30/ml8.hpc.uio.no.csv",
        "../sonar_test_data0/2023/05/31/ml1.hpc.uio.no.csv",
        "../sonar_test_data0/2023/05/31/ml8.hpc.uio.no.csv",
        "../sonar_test_data0/2023/06/01/ml1.hpc.uio.no.csv",
        "../sonar_test_data0/2023/06/01/ml8.hpc.uio.no.csv",
        "../sonar_test_data0/2023/06/02/ml8.hpc.uio.no.csv",
        "../sonar_test_data0/2023/06/03/ml8.hpc.uio.no.csv",
        "../sonar_test_data0/2023/06/04/ml8.hpc.uio.no.csv"]));
}

#[test]
fn test_find_logfiles2() {
    // Use early date bounds for both limits to test that we get a subset.
    let hosts : HashSet<String> = HashSet::new();
    let xs = find_logfiles("../sonar_test_data0",
                           &hosts,
                           DateTime::from_utc(NaiveDate::from_ymd_opt(2023, 5, 20).unwrap().and_hms_opt(0,0,0).unwrap(), Utc),
                           DateTime::from_utc(NaiveDate::from_ymd_opt(2023, 6, 2).unwrap().and_hms_opt(0,0,0).unwrap(), Utc)).unwrap();
    assert!(xs.eq(&vec![
        "../sonar_test_data0/2023/05/30/ml8.hpc.uio.no.csv",
        "../sonar_test_data0/2023/05/31/ml1.hpc.uio.no.csv",
        "../sonar_test_data0/2023/05/31/ml8.hpc.uio.no.csv",
        "../sonar_test_data0/2023/06/01/ml1.hpc.uio.no.csv",
        "../sonar_test_data0/2023/06/01/ml8.hpc.uio.no.csv",
        "../sonar_test_data0/2023/06/02/ml8.hpc.uio.no.csv"]));
}

#[test]
fn test_find_logfiles3() {
    // Filter by host name.
    let mut hosts : HashSet<String> = HashSet::new();
    hosts.insert("ml1.hpc.uio.no".to_string());
    let xs = find_logfiles("../sonar_test_data0",
                           &hosts,
                           DateTime::from_utc(NaiveDate::from_ymd_opt(2023, 5, 20).unwrap().and_hms_opt(0,0,0).unwrap(), Utc),
                           DateTime::from_utc(NaiveDate::from_ymd_opt(2023, 6, 2).unwrap().and_hms_opt(0,0,0).unwrap(), Utc)).unwrap();
    assert!(xs.eq(&vec![
        "../sonar_test_data0/2023/05/31/ml1.hpc.uio.no.csv",
        "../sonar_test_data0/2023/06/01/ml1.hpc.uio.no.csv"]));
}

