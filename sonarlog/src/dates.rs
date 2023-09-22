/// Date and time utilities for sonarlog.
///
/// Not all of these are obvious exports from sonarlog but they are useful and there's no real win
/// (yet) from breaking them out as a separate library.
///
/// TODO: As noted in parse_timestamp() and now() below, timestamps may carry subsecond data.  They
/// may need to be truncated for proper comparison results, or perhaps the subsecond data should be
/// cleared on timestamp creation.

use anyhow::{bail, Result};
use chrono::{DateTime, Datelike, Duration, NaiveDate, Timelike, Utc};

/// Timestamps are always Utc.

pub type Timestamp = DateTime<Utc>;

/// Construct timestamp from its date and time components.

pub fn timestamp_from_ymdhms(y: i32, m: u32, d: u32, h: u32, min: u32, s: u32) -> Timestamp {
    DateTime::from_utc(
        NaiveDate::from_ymd_opt(y, m, d)
            .unwrap()
            .and_hms_opt(h, min, s)
            .unwrap(),
        Utc,
    )
}

/// Construct timestamp from its date components.

pub fn timestamp_from_ymd(y: i32, m: u32, d: u32) -> Timestamp {
    timestamp_from_ymdhms(y, m, d, 0, 0, 0)
}

/// Zero out the minute, second, and subsecond components.

pub fn truncate_to_hour(t: Timestamp) -> Timestamp {
    timestamp_from_ymdhms(t.year(), t.month(), t.day(), t.hour(), 0, 0)
}

/// Add one day to the timestamp.

pub fn add_day(t: Timestamp) -> Timestamp {
    t + Duration::days(1)
}

/// Zero out the hour, minute, second, and subsecond components.

pub fn truncate_to_day(t: Timestamp) -> Timestamp {
    timestamp_from_ymd(t.year(), t.month(), t.day())
}

/// Add one hour to the timestamp.

pub fn add_hour(t: Timestamp) -> Timestamp {
    t + Duration::hours(1)
}

/// epoch: "a long long time ago", before any of our timestamps

pub fn epoch() -> Timestamp {
    // TODO: should do better, but this is currently good enough for all our uses.
    timestamp_from_ymd(2000, 1, 1)
}

/// now: the current time.
///
/// Note the returned timestamp may contain non-zero subsecond data.

pub fn now() -> Timestamp {
    Utc::now()
}

/// Parse the date, which may contain a non-zero TZO, into a UTC timestamp.
///
/// Note the returned timestamp may contain non-zero subsecond data, if the input had subsecond
/// data.

pub fn parse_timestamp(ts: &str) -> Result<Timestamp> {
    match DateTime::parse_from_rfc3339(ts) {
        Err(_) => {
            bail!("Bad time stamp {ts}")
        }
        Ok(v) => {
            // v is DateTime<FixedOffset>, convert to Utc
            Ok(v.into())
        }
    }
}

/// Return a vector of (year, month, day) with times inclusive between the days of t1 and t2;
/// sub-day information in t1 and t2 is ignored.

pub fn date_range(t1: Timestamp, t2: Timestamp) -> Vec<(i32, u32, u32)> {
    let d1 = truncate_to_day(t1);
    let d2 = truncate_to_day(t2);
    let mut date_range = Vec::new();
    let mut current_date = d1;
    while current_date <= d2 {
        date_range.push((
            current_date.year(),
            current_date.month(),
            current_date.day(),
        ));
        current_date += Duration::days(1);
    }
    date_range
}

#[test]
fn test_date_range() {
    let from = timestamp_from_ymdhms(2023, 05, 30, 5, 20, 33);
    let to = timestamp_from_ymd(2023, 06, 04);
    assert!(date_range(from, to).eq(&vec![
        (2023, 5, 30),
        (2023, 5, 31),
        (2023, 6, 1),
        (2023, 6, 2),
        (2023, 6, 3),
        (2023, 6, 4)
    ]));
}

#[test]
fn test_parse() {
    // Test that parsing works and that local time is converted to Utc
    let x = parse_timestamp("2023-07-01T01:20:30+02:00").unwrap();
    assert!(x.year() == 2023);
    assert!(x.month() == 6);
    assert!(x.day() == 30);
    assert!(x.hour() == 23);
    assert!(x.minute() == 20);
    assert!(x.second() == 30);

    // Test that we fail to parse some bogus formats
    assert!(parse_timestamp("2023-07-01T01:20.30+02:00").is_err());
}
