// Misc utilities useful to both sonarlog and sonalyze.

mod configs;
mod csv;
mod dates;
mod hostglob;
mod pattern;

// A structure representing the configuration of one host.

pub use configs::ClusterConfig;
pub use configs::System;

// Read a set of host configurations from a file, and return a map from hostname to configuration.

pub use configs::read_cluster_config;

// Fast, non-allocating, flexible CSV parser.

pub use csv::CsvToken;
pub use csv::CsvTokenizer;
pub use csv::CSV_EQ_SENTINEL;

// Types and utilities for manipulating timestamps.

pub use dates::Timestamp;

// "A long long time ago".

pub use dates::epoch;

// The time right now.

pub use dates::now;

// A time that should not be in any sample record.

pub use dates::far_future;

// Parse a &str into a Timestamp.

pub use dates::parse_timestamp;

// Given year, month, day, hour, minute, second (all UTC), return a Timestamp.

pub use dates::timestamp_from_ymdhms;

// Given year, month, day (all UTC), return a Timestamp.

pub use dates::timestamp_from_ymd;

// Return the timestamp with various parts cleared out.

pub use dates::truncate_to_day;
pub use dates::truncate_to_half_day;
pub use dates::truncate_to_half_hour;
pub use dates::truncate_to_hour;

// Add various quantities to the timestamp

pub use dates::add_day;
pub use dates::add_half_day;
pub use dates::add_half_hour;
pub use dates::add_hour;

// ...

pub use dates::date_range;

// Structure representing a host name filter: basically a restricted automaton matching host names
// in useful ways.

pub use hostglob::HostGlobber;

// Formatter for sets of host names

pub use hostglob::compress_hostnames;
pub use hostglob::expand_pattern;
