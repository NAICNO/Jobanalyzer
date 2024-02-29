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
mod logclean;
mod logentry;
mod logfile;
mod logtree;
mod synthesize;

// Postprocess a vector of log data: compute the cpu_util_pct field, apply a record filter, clean up
// the GPU memory data, and bucket data for different sample streams properly.

pub use logclean::postprocess_log; // -> InputSampleStreams

// A datum representing a key in the map of sample streams: (hostname, stream-id, command).

pub use logclean::InputStreamKey;

pub use logclean::InputStreamSet;

pub use logentry::LogEntry;

// Return an empty Box<LogEntry> with the given time and host.  The user and command fields are
// "_zero_", so that we can recognize it; other fields are generally zero.

pub use logentry::empty_logentry;

pub use logentry::GpuStatus;

pub use logentry::merge_gpu_status;

// Parse a log file into a set of LogEntry structures, applying an application-defined filter to
// each record while reading.

pub use logfile::parse_logfile;

// Compute a set of plausible log file names within a directory tree, for a date range and a set of
// included host names.

pub use logtree::find_logfiles;

// Map from host name to (earliest, latest) time for host.

pub use logtree::Timebound;
pub use logtree::Timebounds;

// Read a set of logfiles into a vector and compute some simple metadata.

pub use logtree::read_logfiles;

// Given a set of sample streams, merge by host and job and return a vector of the merged streams.

pub use synthesize::merge_by_host_and_job; // -> MergedSampleStreams

// Given a set of sample streams, merge by job (across hosts) and return a vector of the merged
// streams.

pub use synthesize::merge_by_job; // -> MergedSampleStreams

// Given a set of sample streams, merge by host (across jobs) and return a vector of the merged
// streams.

pub use synthesize::merge_by_host; // -> MergedSampleStreams

// Given a set of already-merged streams, where each stream pertains to one host and all hosts are
// different, merge by timeslot to create cross-host cross-job data.

pub use synthesize::merge_across_hosts_by_time;

// Bucket samples in a single stream by various time quantities and compute averages.

pub use synthesize::fold_samples_daily;
pub use synthesize::fold_samples_half_daily;
pub use synthesize::fold_samples_half_hourly;
pub use synthesize::fold_samples_hourly;

// A datum representing a bag of merged streams, with no implied constraints on uniqueness of any
// type of key or any ordering.

pub use synthesize::MergedSampleStreams;

