use rustutils::{empty_gpuset, GpuSet, Timestamp};
use std::boxed::Box;
use ustr::Ustr;

/// The LogEntry structure holds slightly processed data from a log record: Percentages have been
/// normalized to the range [0.0,1.0] (except that the CPU and GPU percentages are sums across
/// multiple cores/cards and the sums may exceed 1.0), and memory sizes have been normalized to GB.
///
/// Any discrepancies between the documentation in this structure and the documentation for Sonar
/// (in its top-level README.md) should be considered a bug.
///
/// Space is at a premium here because we allocate very many of these structures.  Hence:
///
/// - fields are f32 instead of f64, we use u16/u32 instead of usize, and we use Ustr for
///   all the strings to avoid allocating very common strings many times over.
/// - fields are ordered to improve packing

#[derive(Debug, Clone)] // TODO: Clone needed by a corner case in sonalyze/load
pub struct LogEntry {
    /// Format "major.minor.bugfix"
    pub major: u16,
    pub minor: u16,
    pub bugfix: u16,

    /// The time is common to all records created by the same sonar invocation.  It has no subsecond
    /// precision.
    pub timestamp: Timestamp,

    /// Fully qualified domain name.
    pub hostname: Ustr,

    /// Number of cores on the node.  This may be zero if there's no information.
    pub num_cores: u16,

    /// Total memory installed on the node.  This may be zero if there's no information.
    pub memtotal_gb: f32,

    /// Unix user name, or `_zombie_<PID>`
    pub user: Ustr,

    /// For a unique process, this is its pid.
    ///
    /// Semi-computed field.  For a rolled-up job record with multiple processes, this is initially
    /// zero, but logclean converts it to job_id + logclean::JOB_ID_TAG.
    pub pid: u32,

    /// The job_id.  This has some complicated constraints, see the Sonar docs.
    pub job_id: u32,

    /// The command contains at least the executable name.  It may contain spaces and other special
    /// characters.  This can be `_unknown_` for zombie jobs and `_noinfo_` for non-zombie jobs when
    /// the command can't be found.
    ///
    /// Semi-computed field.  For merged records, this is either a comma-joined sorted list of the
    /// command names of the original records, or the string "_merged_" when the record represents
    /// the merging of multiple jobs.
    pub command: Ustr,

    /// This is a running average of the CPU usage of the job, over the lifetime of the job, summed
    /// across all the processes of the job.  IT IS NOT A SAMPLE.  100.0=1 core's worth (100%).
    /// Generally, `cpu_util_pct` (below) will be more useful.
    pub cpu_pct: f32,

    /// Main memory used by the job on the node (the memory is shared by all cores on the node) at
    /// the time of sampling.  This is virtual memory, data+stack.  It is f64 because the extra
    /// precision is needed in some cases when we convert back to KiB.
    pub mem_gb: f64,

    /// Resident memory used by the job on the node at the time of sampling.  This should be real
    /// memory, owned exclusively by the process.  RssAnon is not a perfect measure of that, but a
    /// compromise; see comments in Sonar.
    pub rssanon_gb: f32,

    /// The set of GPUs used by the job on the node, None for "none", Some({}) for "unknown",
    /// otherwise Some({m,n,...}).
    pub gpus: GpuSet,

    /// Percent of the sum of the capacity of all GPUs in `gpus`.  100.0 means 1 card's worth of
    /// compute (100%).  This value may be larger than 100.0 as it's the sum across cards.
    ///
    /// For NVIDIA, this is utilization since the last sample.  (nvidia-smi pmon -c 1 -s mu).
    /// For AMD, this is instantaneous utilization (rocm-smi or rocm-smi -u)
    pub gpu_pct: f32,

    /// GPU memory used by the job on the node at the time of sampling, as a percentage of all the
    /// memory on all the cards in `gpus`.  100.0 means 1 card's worth of memory (100%).  This value
    /// may be larger than 100.0 as it's the sum across cards.
    ///
    /// Semi-computed field.  This is not always reliable in its raw form (see Sonar documentation).
    /// The logclean module will tidy this up if presented with an appropriate system configuration.
    pub gpumem_pct: f32,

    /// GPU memory used by the job on the node at the time of sampling, naturally across all GPUs in
    /// `gpus`.
    ///
    /// Semi-computed field.  This is not always reliable in its raw form (see Sonar documentation).
    /// The logclean module will tidy this up if presented with an appropriate system configuration.
    /// It is f64 because the extra precision is needed in some cases when we convert back to KiB.
    pub gpumem_gb: f64,

    /// Status of GPUs, as seen by sonar at the time.
    pub gpu_status: GpuStatus,

    /// Accumulated CPU time for the process since the start, including time for any of its children
    /// that have terminated.  This is f64 because it's easy to overflow the mantissa of an f32 with
    /// this quantity.
    pub cputime_sec: f64,

    /// Number of *other* processes (with the same host and command name) that were rolled up into
    /// this process record.
    pub rolledup: u32,

    /// Computed field.  CPU utilization in percent (100% = one full core) in the time interval
    /// since the previous record for this job.  This is computed by logclean from consecutive
    /// `cputime_sec` fields for records that represent the same job, when the information is
    /// available: it requires the `pid` and `cputime_sec` fields to be meaningful.  For the first
    /// record (where there is no previous record to diff against), the `cpu_pct` value is used
    /// here.
    pub cpu_util_pct: f32,
}

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

/// GPU Status value

#[derive(Debug, Copy, Clone, PartialEq, Eq)]
pub enum GpuStatus {
    Ok = 0,
    UnknownFailure = 1,
}

pub fn merge_gpu_status(lhs: GpuStatus, rhs: GpuStatus) -> GpuStatus {
    match (lhs, rhs) {
        (v, w) if v == w => v,
        (v, GpuStatus::Ok) => v,
        (GpuStatus::Ok, v) => v,
        (_, _) => GpuStatus::UnknownFailure,
    }
}
