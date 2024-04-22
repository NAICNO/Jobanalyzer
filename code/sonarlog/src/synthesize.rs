/// Helpers for merging sample streams.
use crate::{merge_gpu_status, GpuStatus, InputStreamSet, LogEntry, Timebound, Timebounds};

use rustutils::{
    compress_hostnames, empty_gpuset, epoch, far_future, now, union_gpuset, Timestamp,
};
use std::boxed::Box;
use std::cmp::{max, min};
use std::collections::{HashMap, HashSet};
use std::iter::Iterator;
use ustr::Ustr;

/// A bag of merged streams.  The constraints on the individual streams in terms of uniqueness and
/// so on depends on how they were merged and are not implied by the type.

pub type MergedSampleStreams = Vec<Vec<Box<LogEntry>>>;

/// Merge streams that have the same host and job ID into synthesized data.
///
/// Each output stream is sorted ascending by timestamp.  No two records have exactly the same time.
/// All records within a stream have the same host, command, user, and job ID.
///
/// The command name for synthesized data collects all the commands that went into the synthesized
/// stream.

pub fn merge_by_host_and_job(mut streams: InputStreamSet) -> MergedSampleStreams {
    // The value is a set of command names and a vector of the individual streams.
    let mut collections: HashMap<(Ustr, u32), (HashSet<Ustr>, Vec<Vec<Box<LogEntry>>>)> =
        HashMap::new();

    // The value is a map (by host) of the individual streams with job ID zero, these can't be
    // merged and must just be passed on.
    let mut zero: HashMap<Ustr, Vec<Vec<Box<LogEntry>>>> = HashMap::new();

    streams.drain().for_each(|((host, _, cmd), v)| {
        let id = v[0].job_id;
        if id == 0 {
            if let Some(vs) = zero.get_mut(&host) {
                vs.push(v);
            } else {
                zero.insert(host, vec![v]);
            }
        } else {
            let key = (host, id);
            if let Some((cmds, vs)) = collections.get_mut(&key) {
                cmds.insert(cmd);
                vs.push(v);
            } else {
                let mut cmds = HashSet::new();
                cmds.insert(cmd);
                collections.insert(key, (cmds, vec![v]));
            }
        }
    });

    let mut vs: MergedSampleStreams = vec![];
    for ((hostname, job_id), (mut cmds, streams)) in collections.drain() {
        if let Some(zeroes) = zero.remove(&hostname) {
            vs.extend(zeroes);
        }
        let mut commands = cmds.drain().collect::<Vec<Ustr>>();
        commands.sort();
        // Any user from any record is fine.  There should be an invariant that no stream is empty,
        // so this should always be safe.
        let user = streams[0][0].user;
        vs.push(merge_streams(
            hostname,
            ustr_join(commands, ","),
            user,
            job_id,
            streams,
        ));
    }

    vs
}

/// Merge streams that have the same job ID (across hosts) into synthesized data.
///
/// Each output stream is sorted ascending by timestamp.  No two records have exactly the same time.
/// All records within an output stream have the same host name, job ID, command name, and user.
///
/// The command name for synthesized data collects all the commands that went into the synthesized
/// stream.  The host name for synthesized data collects all the hosts that went into the
/// synthesized stream.
///
/// This must also merge the metadata from the different hosts: the time bounds.  For a merged
/// stream, the "earliest" time is the min across the earliest times for the different host streams
/// that go into the merged stream, and the "latest" time is the max across the latest times ditto.

pub fn merge_by_job(
    mut streams: InputStreamSet,
    bounds: &Timebounds,
) -> (MergedSampleStreams, Timebounds) {
    // The value is a set of command names, a set of host names, and a vector of the individual
    // streams.
    let mut collections: HashMap<u32, (HashSet<Ustr>, HashSet<Ustr>, Vec<Vec<Box<LogEntry>>>)> =
        HashMap::new();

    // The value is a vector of the individual streams with job ID zero, these can't be merged and
    // must just be passed on.
    let mut zero: Vec<Vec<Box<LogEntry>>> = vec![];

    streams.drain().for_each(|((host, _, cmd), v)| {
        let id = v[0].job_id;
        if id == 0 {
            zero.push(v);
        } else if let Some((cmds, hosts, vs)) = collections.get_mut(&id) {
            cmds.insert(cmd);
            hosts.insert(host);
            vs.push(v);
        } else {
            let mut cmds = HashSet::new();
            cmds.insert(cmd);
            let mut hosts = HashSet::new();
            hosts.insert(host);
            collections.insert(id, (cmds, hosts, vec![v]));
        }
    });

    let mut new_bounds = HashMap::new();
    for z in zero.iter() {
        let hn = z[0].hostname;
        if !new_bounds.contains_key(&hn) {
            let probe = bounds.get(&hn).expect("Host should be in bounds");
            new_bounds.insert(hn, probe.clone());
        }
    }
    let mut vs: MergedSampleStreams = zero;
    for (job_id, (mut cmds, hosts, streams)) in collections.drain() {
        let hostname = Ustr::from(
            &compress_hostnames(&hosts.iter().copied().collect::<Vec<Ustr>>()).join(","),
        );
        if !new_bounds.contains_key(&hostname) {
            assert!(!hosts.is_empty());
            let (earliest, latest) = hosts.iter().fold((now(), epoch()), |(acc_e, acc_l), hn| {
                let probe = bounds.get(hn).expect("Host should be in bounds");
                (min(acc_e, probe.earliest), max(acc_l, probe.latest))
            });
            new_bounds.insert(hostname, Timebound { earliest, latest });
        }
        let mut commands = cmds.drain().collect::<Vec<Ustr>>();
        commands.sort();
        // Any user from any record is fine.  There should be an invariant that no stream is empty,
        // so this should always be safe.
        let user = streams[0][0].user;
        vs.push(merge_streams(
            Ustr::from(&hostname),
            ustr_join(commands, ","),
            user,
            job_id,
            streams,
        ));
    }

    (vs, new_bounds)
}

/// Merge streams that have the same host (across jobs) into synthesized data.
///
/// Each output stream is sorted ascending by timestamp.  No two records have exactly the same time.
/// All records within an output stream have the same host name, job ID, command name, and user.
///
/// The command name and user name for synthesized data are "_merged_".  It would be possible to do
/// something more interesting, such as aggregating them.
///
/// The job ID for synthesized data is 0, which is not ideal but probably OK so long as the consumer
/// knows it.

pub fn merge_by_host(mut streams: InputStreamSet) -> MergedSampleStreams {
    // The key is the host name.
    let mut collections: HashMap<Ustr, Vec<Vec<Box<LogEntry>>>> = HashMap::new();

    streams.drain().for_each(|((host, _, _), v)| {
        // This lumps jobs with job ID 0 in with the others.
        if let Some(vs) = collections.get_mut(&host) {
            vs.push(v);
        } else {
            collections.insert(host, vec![v]);
        }
    });

    let mut vs: MergedSampleStreams = vec![];
    for (hostname, streams) in collections.drain() {
        let cmdname = Ustr::from("_merged_");
        let username = Ustr::from("_merged_");
        let job_id = 0;
        vs.push(merge_streams(hostname, cmdname, username, job_id, streams));
    }

    vs
}

pub fn merge_across_hosts_by_time(streams: MergedSampleStreams) -> MergedSampleStreams {
    if streams.is_empty() {
        return vec![];
    }
    let hostname = Ustr::from(
        &compress_hostnames(&streams.iter().map(|s| s[0].hostname).collect::<Vec<Ustr>>())
            .join(","),
    );
    vec![merge_streams(
        hostname,
        Ustr::from("_merged_"),
        Ustr::from("_merged_"),
        0,
        streams,
    )]
}

// What does it mean to sample a job that runs on multiple hosts, or to sample a host that runs
// multiple jobs concurrently?
//
// Consider peak CPU utilization.  The single-host interpretation of this is the highest valued
// sample for CPU utilization across the run (sample stream).  For cross-host jobs we want the
// highest valued sum-of-samples (for samples taken at the same time) for CPU utilization across the
// run.  However, in general samples will not have been taken on different hosts at the same time so
// this is not completely trivial.
//
// Consider all sample streams from all hosts in the job in parallel, here "+" denotes a sample and
// "-" denotes time just passing, we have three cores C1 C2 C3, and each character is one time tick:
//
//   t= 01234567890123456789
//   C1 --+---+---
//   C2 -+----+---
//   C3 ---+----+-
//
// At t=1, we get a reading for C2.  This value is now in effect until t=6 when we have a new
// sample for C2.  For C1, we have readings at t=2 and t=6.  We wish to "reconstruct" a CPU
// utilization sample across C1, C2, and C3.  An obvious way to do it is to create samples at t=1,
// t=2, t=3, t=6, t=8.  The values that we create for the sample at eg t=3 are the values in effect
// for C1 and C2 from earlier and the new value for C3 at t=3.  The total CPU utilization at that
// time is the sum of the three values, and that goes into computing the peak.
//
// Thus a cross-host sample stream is a vector of these synthesized samples. The synthesized
// LogEntries that we create will have aggregate host sets (effectively just an aggregate host name
// that is the same value in every record) and gpu sets (just a union).
//
// Algorithm:
//
//  given vector V of sample streams for a set of hosts and a common job ID:
//  given vector A of "current observed values for all streams", initially "0"
//  while some streams in V are not empty
//     get lowest time  (*) (**) across nonempty streams of V
//     update A with values from the those streams
//     advance those streams
//     push out a new sample record with current values
//
// (*) There may be multiple record with the lowest time, and we should do all of them at the same
//     time, to reduce the volume of output.
//
// (**) In practice, sonar will be run by cron and cron is pretty good about running jobs when
//      they're supposed to run.  Therefore there will be a fair amount of correlation across hosts
//      about when these samples are gathered, ie, records will cluster around points in time.  We
//      should capture these clusters by considering all records that are within a W-second window
//      after the earliest next record to have the same time.  In practice W will be small (on the
//      order of 5, I'm guessing).  The time for the synthesized record could be the time of the
//      earliest record, or a midpoint or other statistical quantity of the times that go into the
//      record.
//
// Our normal aggregation logic can be run on the synthesized sample stream.
//
// merge_streams() takes a set of streams for an individual job (along with names for the host, the
// command, the user, and the job) and returns a single, merged stream for the job, where the
// synthesized records for a single job all have the following artifacts.  Let R be the records that
// went into synthesizing a single record according to the algorithm above and S be all the input
// records for the job.  Then:
//
//   - version is "0.0.0".
//   - hostname, command,  user, and job_id are as given to the function
//   - timestamp is synthesized from the timestamps of R
//   - num_cores is 0
//   - memtotal_gb is 0.0
//   - pid is 0
//   - cpu_pct is the sum across the cpu_pct of R
//   - mem_gb is the sum across the mem_gb of R
//   - rssanon_gb is the sum across the rssanon_gb of R
//   - gpus is the union of the gpus across R
//   - gpu_pct is the sum across the gpu_pct of R
//   - gpumem_pct is the sum across the gpumem_pct of R
//   - gpumem_gb is the sum across the gpumem_gb of R
//   - cputime_sec is the sum across the cputime_sec of R
//   - rolledup is the number of records in the list
//   - cpu_util_pct is the sum across the cpu_util_pct of R (roughly the best we can do)
//
// Invariants of the input that are used:
//
// - streams are never empty
// - streams are sorted by ascending timestamp
// - in no stream are there two adjacent records with the same timestamp
//
// Invariants not used:
//
// - records may be obtained from the same host and the streams may therefore be synchronized

fn merge_streams(
    hostname: Ustr,
    command: Ustr,
    username: Ustr,
    job_id: u32,
    streams: Vec<Vec<Box<LogEntry>>>,
) -> Vec<Box<LogEntry>> {
    // Generated records
    let mut records = vec![];

    // Some further observations about the input:
    //
    // Sonar uses the same timestamp for all the jobs seen during the same invocation (this is by
    // design) and even with multi-node jobs the runs of cron will tend to be highly correlated.
    // With records only having 1s resolution for the timestamp, even streams from different nodes
    // will tend to have the same timestamp.  Thus many streams may be the "earliest" stream at each
    // time step, indeed, during normal operation it will be common that all the active streams have
    // the same timestamp in their oldest unconsumed sample.
    //
    // At each time the number of active streams will be O(1) - basically proportional to the number
    // of nodes in the cluster, which is constant.  But the number of streams in a stream set will
    // tend to be O(t) - proportional to the number of time steps covered by the set.
    //
    // As we move forward through time, we start in a situation where most streams are not started
    // yet, then streams become live as we reach their starting point, and then become inactive
    // again as we move past their end point and even the point where they are considered residually
    // live.

    // indices[i] has the index of the next element of stream[i]
    let mut indices = [0].repeat(streams.len());

    // Streams that have moved into the past have their indices[i] value set to STREAM_ENDED, this
    // enables some fast filtering, described later.
    const STREAM_ENDED: usize = 0xFFFFFFFF;

    // selected holds the records selected by the second inner loop, we allocate it once.
    let mut selected: Vec<&Box<LogEntry>> = vec![];
    selected.reserve(streams.len());

    // The following loop nest is O(t^2) and very performance-sensitive.  The number of streams can
    // be very large when running analyses over longer time ranges (month or longer).  The common
    // case is that the outer loop makes one iteration per time step and each inner loop loops over
    // all the streams.  The number of streams will tend to grow with the length of the time window
    // (because new jobs are started and there is at least one stream per job), hence the total
    // amount of work will tend to grow quadratically with time.
    //
    // Conditions have been ordered carefully and some have been added to reduce the number of tests
    // and ensure quick exits.  Computations have been hoisted or sunk to take them off hot paths.
    // Conditions have been combined or avoided by introducing sentinel values (sentinel_time and
    // STREAM_ENDED).
    //
    // There are additional tweaks here:
    //
    // First, the hot inner loops have tests that quickly skip expired streams (note the tests are
    // expressed differently), and in addition, the variable `live` keeps track of the first stream
    // that is definitely not expired, and loops start from this value.  The reason we have both
    // `live` and the fast test is that there may be expired streams following non-expired streams
    // in the array of streams.
    //
    // Second, the second loop also very quickly skips unstarted streams.
    //
    // TODO: The inner loop counts could be reduced significantly if we could partition the streams
    // array precisely into streams that are expired, current, and in the future.  However, the
    // current tests are very quick, and any scheme to introduce that partitioning must be very,
    // very cheap, and benefits may not show until the number of inputs is exceptionally large
    // (perhaps 90 or more days of data instead of the 30 days of data I've been testing with).  In
    // addition, attempts at implementing this partitioning have so far resulted in major slowdowns,
    // possibly because the resulting code confuses bounds checking optimizations in the Rust
    // compiler.  This needs to be investigated further.

    // The first stream that is known not to be expired.
    let mut live = 0;

    let sentinel_time = far_future();
    loop {
        // You'd think that it'd be better to have this loop down below where values are set to
        // STREAM_ENDED, but empirically it's better to have it here.  The difference is fairly
        // pronounced.
        while live < streams.len() && indices[live] == STREAM_ENDED {
            live += 1;
        }

        // Loop across streams to find smallest head.
        let mut min_time = sentinel_time;
        for i in live..streams.len() {
            if indices[i] >= streams[i].len() {
                continue;
            }
            // stream[i] has a value, select this stream if we have no stream or if the value is
            // smaller than the one at the head of the smallest stream.
            if min_time > streams[i][indices[i]].timestamp {
                min_time = streams[i][indices[i]].timestamp;
            }
        }

        // Exit if no values in any stream
        if min_time == sentinel_time {
            break;
        }

        let lim_time = min_time + chrono::Duration::seconds(10);
        let near_past = min_time - chrono::Duration::seconds(30);
        let deep_past = min_time - chrono::Duration::seconds(60);

        // Now select values from all streams (either a value in the time window or the most recent
        // value before the time window) and advance the stream pointers for the ones in the window.
        //
        // The cases marked "highly likely" get most of the hits in long runs, then the case marked
        // "fairly likely" gets one hit per record, usually, and then the case that retires a stream
        // gets one hit per stream.

        for i in live..streams.len() {
            let s = &streams[i];
            let ix = indices[i];
            let lim = s.len();

            // lim > 0 because no stream is empty

            if ix < lim {
                // Live or future stream.

                // ix < lim

                if s[ix].timestamp >= lim_time {
                    // Highly likely - the stream starts in the future.
                    continue;
                }

                // ix < lim
                // s[ix].timestamp < lim_time

                if s[ix].timestamp == min_time {
                    // Fairly likely in normal input - sample time is equal to the min_time.  This
                    // would be subsumed by the following test using >= for > but the equality test
                    // is faster.
                    selected.push(&s[ix]);
                    indices[i] += 1;
                    continue;
                }

                // ix < lim
                // s[ix].timestamp < lim_time
                // s[ix].timestamp != min_time

                if s[ix].timestamp > min_time {
                    // Unlikely in normal input - sample time is in in the time window but not equal
                    // to min_time.
                    selected.push(&s[ix]);
                    indices[i] += 1;
                    continue;
                }

                // ix < lim
                // s[ix].timestamp < min_time

                if ix > 0 && s[ix - 1].timestamp >= near_past {
                    // Unlikely in normal input - Previous exists and is not last and is in the near
                    // past (redundant test for t < lim_time removed).  The condition is tricky.
                    // ix>0 guarantees that there is a past record at ix - 1, while ix<lim says that
                    // there is also a future record at ix.
                    //
                    // This is hard to make reliable.  The guard on the time is necessary to avoid
                    // picking up records from a lot of dead processes.  Intra-host it is OK.
                    // Cross-host it depends on sonar runs being more or less synchronized.
                    selected.push(&s[ix - 1]);
                    continue;
                }

                // ix < lim
                // s[ix].timestamp < min_time
                // s[ix-1].timestamp < near_past

                // This is duplicated within the ix==lim nest below, in a different form.
                if ix > 0 && s[ix - 1].timestamp >= deep_past {
                    // Previous exists (and is last) and is not in the deep past, pick it up
                    selected.push(&s[ix - 1]);
                    continue;
                }

                // ix < lim
                // s[ix].timestamp < min_time
                // s[ix-1].timestamp < deep_past

                // This is an old record and we can ignore it.
                continue;
            } else if ix == STREAM_ENDED {
                // Highly likely - stream already marked as exhausted.
                continue;
            } else {
                // About-to-be exhausted stream.

                // ix == lim
                // ix > 0 because lim > 0

                if s[ix - 1].timestamp < deep_past {
                    // Previous is in the deep past and no current - stream is done.
                    indices[i] = STREAM_ENDED;
                    continue;
                }

                // ix == lim
                // ix > 0
                // s[ix-1].timestamp >= deep_past

                // This case is a duplicate from the ix<lim nest above, in a different form.
                if s[ix - 1].timestamp < min_time {
                    // Previous exists (and is last) and is not in the deep past, pick it up
                    selected.push(&s[ix - 1]);
                    continue;
                }

                // ix == lim
                // ix > 0
                // s[ix-1].timestamp >= min_time

                // This is a contradiction probably and it seems we should not come this far.  Don't
                // worry about it.
                continue;
            }
        }

        records.push(sum_records(
            (0u16, 0u16, 0u16),
            min_time,
            hostname,
            username,
            job_id,
            command,
            &selected,
        ));
        selected.clear();
    }

    records
}

fn sum_records(
    version: (u16, u16, u16),
    timestamp: Timestamp,
    hostname: Ustr,
    user: Ustr,
    job_id: u32,
    command: Ustr,
    selected: &[&Box<LogEntry>],
) -> Box<LogEntry> {
    let cpu_pct = selected.iter().fold(0.0, |acc, x| acc + x.cpu_pct);
    let mem_gb = selected.iter().fold(0.0, |acc, x| acc + x.mem_gb);
    let rssanon_gb = selected.iter().fold(0.0, |acc, x| acc + x.rssanon_gb);
    let gpu_pct = selected.iter().fold(0.0, |acc, x| acc + x.gpu_pct);
    let gpumem_pct = selected.iter().fold(0.0, |acc, x| acc + x.gpumem_pct);
    let gpumem_gb = selected.iter().fold(0.0, |acc, x| acc + x.gpumem_gb);
    let cputime_sec = selected.iter().fold(0.0, |acc, x| acc + x.cputime_sec);
    let cpu_util_pct = selected.iter().fold(0.0, |acc, x| acc + x.cpu_util_pct);
    // The invariant here is that rolledup is the number of *other* processes rolled up into
    // this one.  So we add one for each in the list + the others rolled into each of those,
    // and subtract one at the end to maintain the invariant.
    let rolledup = selected.iter().fold(0, |acc, x| acc + x.rolledup + 1) - 1;
    let gpu_status = selected
        .iter()
        .fold(GpuStatus::Ok, |acc, x| merge_gpu_status(acc, x.gpu_status));
    let mut gpus = empty_gpuset();
    for s in selected {
        union_gpuset(&mut gpus, &s.gpus);
    }

    // Synthesize the record.
    Box::new(LogEntry {
        major: version.0,
        minor: version.1,
        bugfix: version.2,
        timestamp,
        hostname,
        num_cores: 0,
        memtotal_gb: 0.0,
        user,
        pid: 0,
        job_id,
        command,
        cpu_pct,
        mem_gb,
        rssanon_gb,
        gpus,
        gpu_pct,
        gpumem_pct,
        gpumem_gb,
        gpu_status,
        cputime_sec,
        rolledup,
        cpu_util_pct,
    })
}

pub fn fold_samples_hourly(samples: Vec<Box<LogEntry>>) -> Vec<Box<LogEntry>> {
    fold_samples(samples, rustutils::truncate_to_hour)
}

pub fn fold_samples_half_hourly(samples: Vec<Box<LogEntry>>) -> Vec<Box<LogEntry>> {
    fold_samples(samples, rustutils::truncate_to_half_hour)
}

pub fn fold_samples_daily(samples: Vec<Box<LogEntry>>) -> Vec<Box<LogEntry>> {
    fold_samples(samples, rustutils::truncate_to_day)
}

pub fn fold_samples_half_daily(samples: Vec<Box<LogEntry>>) -> Vec<Box<LogEntry>> {
    fold_samples(samples, rustutils::truncate_to_half_day)
}

pub fn fold_samples_weekly(samples: Vec<Box<LogEntry>>) -> Vec<Box<LogEntry>> {
    fold_samples(samples, rustutils::truncate_to_week)
}

fn fold_samples(
    samples: Vec<Box<LogEntry>>,
    get_time: fn(Timestamp) -> Timestamp,
) -> Vec<Box<LogEntry>> {
    let mut result = vec![];
    let mut i = 0;
    while i < samples.len() {
        let s0 = &samples[i];
        let t0 = get_time(s0.timestamp);
        i += 1;
        let mut bucket = vec![s0];
        while i < samples.len() && get_time(samples[i].timestamp) == t0 {
            bucket.push(&samples[i]);
            i += 1;
        }
        let mut r = sum_records(
            (0u16, 0u16, 0u16),
            t0,
            s0.hostname,
            Ustr::from("_merged_"),
            0,
            Ustr::from("_merged_"),
            &bucket,
        );
        let n32 = bucket.len() as f32;
        let n64 = bucket.len() as f64;
        r.cpu_pct /= n32;
        r.mem_gb /= n64;
        r.rssanon_gb /= n32;
        r.gpu_pct /= n32;
        r.gpumem_pct /= n32;
        r.gpumem_gb /= n64;
        r.cpu_util_pct /= n32;
        r.cputime_sec /= n64;
        result.push(r);
    }

    result
}

fn ustr_join(ss: Vec<Ustr>, joiner: &str) -> Ustr {
    if ss.is_empty() {
        return Ustr::from("");
    }
    let mut s = ss[0].to_string();
    for t in ss[1..].iter() {
        s += joiner;
        s += t.as_str();
    }
    Ustr::from(&s)
}
