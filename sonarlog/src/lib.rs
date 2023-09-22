/// A Sonar log is a structured log: Individual *log records* are structured such that data fields
/// can be found in them and extracted from them, and the various fields have specific and
/// documented meanings.  Log records are found in *log files*, which are in turn present in *log
/// trees* in the file system.
/// 
/// Though a log tree is usually coded in such a way that the location and name of a file indicates
/// the origin (host) and time ranges of the records within it, this is an assumption that is only
/// used by this library when filtering the files to examine in a log tree.  Once a log file is
/// ingested, it is processed without the knowledge of its origin.  In a raw log file, there may
/// thus be records for multiple processes per job and for multiple hosts, and the file need not be
/// sorted in any particular way.
///
/// Log data represent a set of *sample streams* from a set of running systems.  Each stream
/// represents samples from a single *job artifact* - a single process or a set of processes of the
/// same job id and command name that were rolled up by Sonar.  The stream is uniquely identified by
/// the triple (hostname, id, command), where `id` is either the process ID for non-rolled-up
/// processes or the job ID + logclean::JOB_ID_TAG for rolled-up processes (see logclean.rs for a
/// lot more detail).  There may be multiple job artifacts, and hence multiple sample streams, per
/// job - both on a single host and across hosts.
///
/// There is an important invariant on the raw log records:
///
/// - no two records in a single stream will have the same timestamp
///
/// This library has as its fundamental task to reconstruct the set of sample streams from the raw
/// logs and provide utilities to manipulate that set.  This task breaks down into a number of
/// subtasks:
///
/// - Find log files within the log tree, applying filters by date and host name.
///
/// - Parse the log records within the log files, handling both the older record format (no fields
///   names) and the newer record format (field names) transparently.  Support for older field names
///   is now opt-in under the feature "untagged-sonar-data".
///
/// - Clean up and filter and bucket the log data by stream.
///
/// - Merge and fold sample streams, to create complete views of jobs or systems

mod configs;
mod dates;
mod hosts;
mod logclean;
mod logfile;
mod logtree;
mod pattern;
mod synthesize;

// Types and utilities for manipulating timestamps.

pub use dates::Timestamp;

// "A long long time ago".

pub use dates::epoch;

// The time right now.

pub use dates::now;

// Parse a &str into a Timestamp.

pub use dates::parse_timestamp;

// Given year, month, day, hour, minute, second (all UTC), return a Timestamp.

pub use dates::timestamp_from_ymdhms;

// Given year, month, day (all UTC), return a Timestamp.

pub use dates::timestamp_from_ymd;

// Return the timestamp with minutes and seconds cleared out.

pub use dates::truncate_to_hour;

// Return the timestamp with hours, minutes, and seconds cleared out.

pub use dates::truncate_to_day;

// Add one hour to the timestamp

pub use dates::add_hour;

// Add one day to the timestamp

pub use dates::add_day;

// Compute a set of plausible log file names within a directory tree, for a date range and a set of
// included host names.

pub use logtree::find_logfiles;

// Read a set of logfiles into a vector and compute some simple metadata.

pub use logtree::read_logfiles;

// Parse a log file into a set of LogEntry structures, applying an application-defined filter to
// each record while reading.

pub use logfile::parse_logfile;

// A GpuSet is None, Some({}), or Some({a,b,...}), representing unknown, empty, or non-empty.

pub use logfile::GpuSet;

// Create an empty GpuSet.

pub use logfile::empty_gpuset;

// Create a GpuSet with unknown contents.

pub use logfile::unknown_gpuset;

// Create a GpuSet that is either None or Some({a}), depending on input.

pub use logfile::singleton_gpuset;

// Union one GPU into a GpuSet (destructively).

pub use logfile::adjoin_gpuset;

// Union one GpuSet into another (destructively).

pub use logfile::union_gpuset;

// Return an empty Box<LogEntry> with the given time and host.  The user and command fields are "_zero_",
// so that we can recognize it; other fields are generally zero.

pub use logfile::empty_logentry;

// Postprocess a vector of log data: compute the cpu_util_pct field, apply a record filter, clean up
// the GPU memory data, and bucket data for different sample streams properly.

pub use logclean::postprocess_log; // -> InputSampleStreams

// Given a set of sample streams, merge by host and job and return a vector of the merged streams.

pub use synthesize::merge_by_host_and_job; // -> MergedSampleStreams

// Given a set of sample streams, merge by job (across hosts) and return a vector of the merged
// streams.

pub use synthesize::merge_by_job; // -> MergedSampleStreams

// Given a set of sample streams, merge by host (across jobs) and return a vector of the merged
// streams.

pub use synthesize::merge_by_host; // -> MergedSampleStreams

// Bucket samples in a single stream by hour and compute averages.

pub use synthesize::fold_samples_hourly;

// Bucket samples in a single stream by day and compute averages.

pub use synthesize::fold_samples_daily;

// A datum representing a bag of merged streams, with no implied constraints on uniqueness of any
// type of key or any ordering.

pub use synthesize::MergedSampleStreams;

// A datum representing a key in the map of sample streams: (hostname, stream-id, command).

pub use logclean::InputStreamKey;

pub use logclean::InputStreamSet;

/// The LogEntry structure holds slightly processed data from a log record: Percentages have been
/// normalized to the range [0.0,1.0] (except that the CPU and GPU percentages are sums across
/// multiple cores/cards and the sums may exceed 1.0), and memory sizes have been normalized to GB.
///
/// Any discrepancies between the documentation in this structure and the documentation for Sonar
/// (in its top-level README.md) should be considered a bug.

#[derive(Debug, Clone)]         // TODO: Clone needed by a corner case in sonalyze/load
pub struct LogEntry {
    /// Format "major.minor.bugfix"
    pub version: String,

    /// The time is common to all records created by the same sonar invocation.  It has no subsecond
    /// precision.
    pub timestamp: Timestamp,

    /// Fully qualified domain name.
    pub hostname: String,

    /// Number of cores on the node.  This may be zero if there's no information.
    pub num_cores: u32,

    /// Unix user name, or `_zombie_<PID>`
    pub user: String,

    /// For a unique process, this is its pid.  For a rolled-up job record with multiple processes,
    /// this is initially zero, but logclean converts it to job_id + 10000000.
    pub pid: u32,

    /// The job_id.  This has some complicated constraints, see the Sonar docs.
    pub job_id: u32,

    /// The command contains at least the executable name.  It may contain spaces and other special
    /// characters.  This can be `_unknown_` for zombie jobs and `_noinfo_` for non-zombie jobs when
    /// the command can't be found.
    pub command: String,

    /// This is a running average of the CPU usage of the job, over the lifetime of the job, summed
    /// across all the processes of the job.  IT IS NOT A SAMPLE.  100.0=1 core's worth (100%).
    /// Generally, `cpu_util_pct` (below) will be more useful.
    pub cpu_pct: f64,

    /// Main memory used by the job on the node (the memory is shared by all cores on the node) at
    /// the time of sampling.
    pub mem_gb: f64,

    /// The set of GPUs used by the job on the node, None for "none", Some({}) for "unknown",
    /// otherwise Some({m,n,...}).
    pub gpus: GpuSet,

    /// Percent of the sum of the capacity of all GPUs in `gpus`.  100.0 means 1 card's worth of
    /// compute (100%).  This value may be larger than 100.0 as it's the sum across cards.
    ///
    /// For NVIDIA, this is utilization since the last sample.  (nvidia-smi pmon -c 1 -s mu).
    /// For AMD, this is instantaneous utilization (rocm-smi or rocm-smi -u)
    pub gpu_pct: f64,

    /// GPU memory used by the job on the node at the time of sampling, as a percentage of all the
    /// memory on all the cards in `gpus`.  100.0 means 1 card's worth of memory (100%).  This value
    /// may be larger than 100.0 as it's the sum across cards.
    ///
    /// Note this is not always reliable in its raw form (see Sonar documentation).  The logclean
    /// module will tidy this up if presented with an appropriate system configuration.
    pub gpumem_pct: f64,

    /// GPU memory used by the job on the node at the time of sampling, naturally across all GPUs in
    /// `gpus`.
    ///
    /// Note this is not always reliable in its raw form (see Sonar documentation).  The logclean
    /// module will tidy this up if presented with an appropriate system configuration.
    pub gpumem_gb: f64,

    /// Accumulated CPU time for the process since the start, including time for any of its children
    /// that have terminated.
    pub cputime_sec: f64,

    /// Number of *other* processes (with the same host and command name) that were rolled up into
    /// this process record.
    pub rolledup: u32,

    // Computed fields.  Also see above about pid, gpumem_pct, and gpumem_gb.

    /// CPU utilization in percent (100% = one full core) in the time interval since the previous
    /// record for this job.  This is computed by logclean from consecutive `cputime_sec` fields for
    /// records that represent the same job, when the information is available: it requires the
    /// `pid` and `cputime_sec` fields to be meaningful.  For the first record (where there is no
    /// previous record to diff against), the `cpu_pct` value is used here.
    pub cpu_util_pct: f64,
}

// Structure representing a host name filter: basically a restricted automaton matching host names
// in useful ways.

pub use hosts::HostFilter;

// Formatter for sets of host names

pub use hosts::combine_hosts;

// A structure representing the configuration of one host.

pub use configs::System;

// Read a set of host configurations from a file, and return a map from hostname to configuration.

pub use configs::read_from_json;
