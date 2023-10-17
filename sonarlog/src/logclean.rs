/// Postprocess and clean up log data after ingestion.

use crate::{LogEntry, System};

use chrono::Duration;
use std::boxed::Box;
use std::collections::HashMap;

#[cfg(test)]
use crate::read_logfiles;

/// The InputStreamKey is (hostname, stream-id, cmd), where the stream-id is defined below; it is
/// meaningful only for non-merged streams.

pub type InputStreamKey = (String, u32, String);

/// A InputStreamSet maps a InputStreamKey to a list of records pertinent to that key.  It is named
/// as it is because the InputStreamKey is meaningful only for non-merged streams.
///
/// There are some important invariants on the records that make up a stream in addition to them
/// having the same key:
///
/// - the vector is sorted ascending by timestamp
/// - no two adjacent timestamps are the same

pub type InputStreamSet = HashMap<InputStreamKey, Vec<Box<LogEntry>>>;

/// Apply postprocessing to the records in the array:
///
/// - reconstruct individual sample streams
/// - compute the cpu_util_pct field from cputime_sec and timestamp for consecutive records in streams
/// - if `configs` is not None and there is the necessary information for a given host, clean up the
///   gpumem_pct and gpumem_gb fields so that they are internally consistent
/// - after all that, remove records for which the filter function returns false
///
/// Returns the individual streams as a map from (hostname, id, cmd) to a vector of LogEntries,
/// where each vector is sorted in ascending order of time.  In each vector, there may be no
/// adjacent records with the same timestamp.
///
/// The id is necessary to distinguish the different event streams for a single job.  Consider a run
/// of records from the same host.  There may be multiple records per job in that run, and they may
/// or may not also have the same cmd, and they may or may not have been rolled up.  There are two
/// cases:
///
/// - If the job is not rolled-up then we know that for a given pid there is only ever one record at
///   a given time.
///
/// - If the job is rolled-up then we know that for a given (job_id, cmd) pair there is only one
///   record, but job_id by itself is not enough to distinguish records, and there is no obvious
///   distinguishing pid value, as the set of rolled-up processes may change from invocation to
///   invocation of sonar.  We also know a rolled-up record has rolledup > 0.
///
/// Therefore, let the pid for a rolled-up record r be JOB_ID_TAG + r.job_id.  Then (pid, cmd) is
/// enough to distinguish a record always, though it is more complicated than necessary for
/// non-rolled-up jobs.
///
/// TODO: JOB_ID_TAG is currently 1e8 because Linux pids are smaller than 1e8, so this guarantees
/// that there is not a clash with a pid, but it's possible job IDs can be larger than PIDS.

pub const JOB_ID_TAG : u32 = 10000000;

pub fn postprocess_log<F>(
    mut entries: Vec<Box<LogEntry>>,
    filter: F,
    configs: &Option<HashMap<String, System>>,
) -> InputStreamSet
where
    F: Fn(&LogEntry) -> bool
{
    let mut streams : InputStreamSet = HashMap::new();

    // Reconstruct the individual sample streams.  Records for job id 0 are always not rolled up and
    // we'll use the pid, which is unique.  But consumers of the data must be sure to treat job id 0
    // specially.
    entries
        .drain(0..)
        .for_each(|e| {
            let synthetic_pid = if e.rolledup > 0 {
                JOB_ID_TAG + e.job_id
            } else {
                e.pid
            };
            let key = (e.hostname.clone(), synthetic_pid, e.command.clone());
            if let Some(stream) = streams.get_mut(&key) {
                stream.push(e);
            } else {
                streams.insert(key, vec![e]);
            }
        });

    // Sort the streams by ascending timestamp.
    streams
        .iter_mut()
        .for_each(|(_, stream)| stream.sort_by(|a, b| a.timestamp.cmp(&b.timestamp)));

    // Remove duplicate timestamps.  These may appear due to system effects, notably, sonar log
    // generation may be delayed due to disk waits, which may be long because network disks may
    // go away.  It should not matter which of the duplicate records we remove here, they should
    // be identical.
    streams
        .iter_mut()
	.for_each(|(_, stream)| {
            let mut good = 0;
	    let mut candidate = good+1;
            while candidate < stream.len() {
	        while candidate < stream.len() && stream[good].timestamp == stream[candidate].timestamp {
		    candidate += 1;
		}
                if candidate < stream.len() {
		    good += 1;
		    stream.swap(good, candidate);
		    candidate += 1;
	        }
	    }
            stream.truncate(good + 1);
	});

    // For each stream, compute the cpu_util_pct field as the difference in cputime_sec between
    // adjacent records divided by the time difference between them.  The first record instead gets
    // a copy of the cpu_pct field.
    streams
        .iter_mut()
        .for_each(|(_, stream)| {
            // By construction, every stream is non-empty.
            stream[0].cpu_util_pct = stream[0].cpu_pct;
            for i in 1..stream.len() {
                let dt = ((stream[i].timestamp - stream[i-1].timestamp) as Duration).num_seconds() as f64;
                let dc = stream[i].cputime_sec - stream[i-1].cputime_sec;
		// It can happen that dc < 0, see https://github.com/NAICNO/Jobanalyzer/issues/63.
		// We filter these below.
                stream[i].cpu_util_pct = (dc / dt) * 100.0;
            }
        });

    // For each stream, clean up the gpumem_pct and gpumem_gb fields based on system information, if
    // available.
    if let Some(confs) = configs {
        streams
            .iter_mut()
            .for_each(|(_, stream)| {
                if let Some(conf) = confs.get(&stream[0].hostname) {
                    let cardsize = (conf.gpumem_gb as f64) / (conf.gpu_cards as f64);
                    for entry in stream {
                        if conf.gpumem_pct {
                            entry.gpumem_gb = entry.gpumem_pct * cardsize;
                        } else {
                            entry.gpumem_pct = entry.gpumem_gb / cardsize;
                        }
                    }
                }
            });
    }

    // Remove elements that don't pass the filter and pack the array.  This preserves ordering.
    streams
        .iter_mut()
        .for_each(|(_, stream)| {
            let mut dst = 0;
            for src in 0..stream.len() {
	        // See comments above re the test for cpu_util_pct
                if filter(&stream[src]) && stream[src].cpu_util_pct >= 0.0 {
                    stream.swap(dst, src);
                    dst += 1;
                }
            }
            stream.truncate(dst);
        });

    // Some streams may now be empty; remove them.
    let dead = streams
        .iter()
        .filter_map(|(k, v)| {
            if v.len() == 0 {
                Some(k.clone())
            } else {
                None
            }
        }).
        collect::<Vec<InputStreamKey>>();

    for d in dead {
        streams.remove(&d);
    }

    streams
}

#[test]
fn test_postprocess_log_cpu_util_pct() {
    // This file has field names, cputime_sec, pid, and rolledup
    // There are two hosts.
    let (entries, _) = read_logfiles(&vec!["../tests/sonarlog/whitebox-logclean.csv".to_string()]).unwrap();
    assert!(entries.len() == 7);

    let any = |e:&LogEntry| e.user != "root";
    let streams = postprocess_log(entries, any, &None);

    // Filtering removed one entry and grouped the rest into four streams.
    assert!(streams.len() == 4);

    let s1 = streams.get(&("ml4.hpc.uio.no".to_string(), JOB_ID_TAG + 4093, "zabbix_agentd".to_string()));
    assert!(s1.is_some());
    assert!(s1.unwrap().len() == 1);

    let s2 = streams.get(&("ml4.hpc.uio.no".to_string(), 1090, "python".to_string()));
    assert!(s2.is_some());
    let v2 = s2.unwrap();
    assert!(v2.len() == 3);
    assert!(v2[0].timestamp < v2[1].timestamp);
    assert!(v2[1].timestamp < v2[2].timestamp);
    // For this pid (1090) there are three records for ml4, pairwise 300 seconds apart (and
    // disordered in the input), and the cputime_sec field for the second record is 300 seconds
    // higher, giving us 100% utilization for that time window, and for the third record 150 seconds
    // higher, giving us 50% utilization for that window.
    assert!(v2[0].cpu_util_pct == 1473.7); // The cpu_pct value
    assert!(v2[1].cpu_util_pct == 100.0);
    assert!(v2[2].cpu_util_pct == 50.0);

    // This has the same pid *but* a different host, so the utilization for the first record should
    // once again be set to the cpu_pct value.
    let s3 = streams.get(&("ml5.hpc.uio.no".to_string(), 1090, "python".to_string()));
    assert!(s3.is_some());
    assert!(s3.unwrap().len() == 1);
    assert!(s3.unwrap()[0].cpu_util_pct == 128.0);

    let s4 = streams.get(&("ml4.hpc.uio.no".to_string(), 1089, "python".to_string()));
    assert!(s4.is_some());
    assert!(s4.unwrap().len() == 1);
}
