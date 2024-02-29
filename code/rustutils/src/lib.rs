// Misc utilities useful to both sonarlog and sonalyze.

mod configs;
mod csv;
mod dates;
mod gpuset;
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

// A GpuSet is None, Some({}), or Some({a,b,...}), representing unknown, empty, or non-empty.

pub use gpuset::GpuSet;

// Create an empty GpuSet.

pub use gpuset::empty_gpuset;

// Test if a GpuSet is known to be the empty set (not unknown).

pub use gpuset::is_empty_gpuset;

// Create a GpuSet with unknown contents.

pub use gpuset::unknown_gpuset;

// Test if a GpuSet is known to be the unknown set.

pub use gpuset::is_unknown_gpuset;

// Create a GpuSet that is either None or Some({a}), depending on input.

pub use gpuset::singleton_gpuset;

// Union one GPU into a GpuSet (destructively).

pub use gpuset::adjoin_gpuset;

// Union one GpuSet into another (destructively).

pub use gpuset::union_gpuset;

// Convert to "unknown" or "none" or a list of numbers.

pub use gpuset::gpuset_to_string;

pub use gpuset::gpuset_from_bitvector;
pub use gpuset::gpuset_from_list;

// Structure representing a host name filter: basically a restricted automaton matching host names
// in useful ways.

pub use hostglob::HostGlobber;

// Formatter for sets of host names

pub use hostglob::compress_hostnames;
pub use hostglob::expand_pattern;
