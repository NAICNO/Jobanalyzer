// Postprocess and clean up log data after ingestion.

package sample

import (
	"cmp"
	"maps"
	"slices"
	"strconv"
	"strings"

	"go-utils/config"

	. "sonalyze/common"
	"sonalyze/db/repr"
)

// About stream IDs
//
// The stream id is necessary to distinguish the different event streams within a single job.
// Consider a run of records from the same host.  There may be multiple records per job in that run,
// and they may or may not also have the same cmd, and they may or may not have been rolled up.
// There are two cases:
//
// - If the job is not rolled-up then we know that for a given pid there is only ever one record at
//   a given time.
//
// - If the job is rolled-up then we know that for a given (job_id, cmd) pair there is only one
//   record, but job_id by itself is not enough to distinguish records, and there is no obvious
//   distinguishing pid value, as the set of rolled-up processes may change from invocation to
//   invocation of sonar.  We also know a rolled-up record has rolledup > 0.
//
// Therefore, let the pid for a rolled-up record r be JOB_ID_TAG + r.job_id.  Then (pid, cmd) is
// enough to distinguish a record always, though it is more complicated than necessary for
// non-rolled-up jobs.
//
// TODO: INVESTIGATE: JobIdTag is currently 1e8 because Linux pids are smaller than 1e8, so this
// guarantees that there is not a clash with a pid, but it's possible job IDs can be larger than
// PIDS.
//
// TODO: INVESTIGATE: Admins swear that Slurm Job IDs are never reused (and that chaos ensues when
// they are) and are currently above 1e6 on some systems.
//
// TODO: The invariant stated above is not true, because rollups only happen for jobs that have
// the same ppid and no children.  Thus we could easily construct a process hierarchy where there
// are several groups of rolled-up processes with the same job ID and command.

const JobIdTag = 10000000

// Records for job id 0 are always not rolled up and we'll use the pid for those, which is unique.
func streamId(e *repr.Sample) uint32 {
	syntheticPid := e.Pid
	if e.Rolledup > 0 {
		syntheticPid = JobIdTag + e.Job
	}
	return syntheticPid
}

// The SampleRectifier is applied to samples when they are read from a file, before caching, and can
// also (eventually) be applied to samples that are read from in-memory records to be appended to a
// file.  All samples are from the same host and the same date (UTC), but otherwise there are few
// guarantees.
//
// For now, clean up the gpumem_pct and gpumem_gb fields based on system information.
func standardSampleRectifier(xs []*repr.Sample, cfg *config.ClusterConfig) []*repr.Sample {
	if cfg == nil || len(xs) == 0 {
		return xs
	}

	conf := cfg.LookupHost(xs[0].Hostname.String())
	if conf == nil {
		return xs
	}

	cardsizeKB := float64(conf.GpuMemGB) * 1024 * 1024 / float64(conf.GpuCards)
	for _, x := range xs {
		if conf.GpuMemPct {
			x.GpuKB = uint64(float64(x.GpuMemPct) / 100 * cardsizeKB)
		} else {
			x.GpuMemPct = float32(float64(x.GpuKB) / cardsizeKB * 100)
		}
	}

	return xs
}

// Given a set of records, reconstruct individual sample streams, sort them, drop duplicates, and
// perform intra-record fixups based on config or other data.
//
// Returns the individual non-empty streams as a map from (hostname, id, cmd) to a vector of
// Samples, where each vector is sorted in ascending order of time.  In each vector, there may be no
// adjacent records with the same timestamp.

func createInputStreams(
	entryBlobs [][]*repr.Sample,
	recordFilter *SampleFilter,
	wantBounds bool,
) (InputStreamSet, Timebounds) {
	streams := make(InputStreamSet)
	bounds := make(Timebounds)

	filterFunc := InstantiateSampleFilter(recordFilter)

	// Reconstruct the individual sample streams.  Each job with job id 0 gets its own stream, these
	// must not be merged downstream (they get different stream IDs but the job IDs remain 0).
	//
	// Also compute time bounds.  Bounds are computed from the entire input set before filtering or
	// other postprocessing - this is closest to what's expected.  But that makes the bound
	// computation a significant expense.
	for _, entries := range entryBlobs {
		for _, e := range entries {
			if wantBounds {
				if bound, found := bounds[e.Hostname]; found {
					bounds[e.Hostname] = Timebound{
						Earliest: min(bound.Earliest, e.Timestamp),
						Latest:   max(bound.Latest, e.Timestamp),
					}
				} else {
					bounds[e.Hostname] = Timebound{
						Earliest: e.Timestamp,
						Latest:   e.Timestamp,
					}
				}
			}

			if !filterFunc(e) {
				continue
			}

			key := InputStreamKey{e.Hostname, streamId(e), e.Cmd}
			if stream, found := streams[key]; found {
				*stream = append(*stream, Sample{Sample: e})
			} else {
				streams[key] = &SampleStream{Sample{Sample: e}}
			}
		}
	}

	// Sort the streams by ascending timestamp.  This sort is stable so that records coming from the
	// same data file are kept in order, in corner cases (esp around roundtripping with `parse`)
	// it's surprising if they are not in order.
	for _, stream := range streams {
		slices.SortStableFunc(*stream, func(a, b Sample) int {
			return cmp.Compare(a.Timestamp, b.Timestamp)
		})
	}

	// Remove duplicate timestamps.  These may appear due to system effects, notably, sonar log
	// generation may be delayed due to disk waits, which may be long because network disks may
	// go away.  It should not matter which of the duplicate records we remove here, they should
	// be identical.  CompactFunc keeps the first.
	for _, stream := range streams {
		*stream = slices.CompactFunc(
			*stream,
			func(a, b Sample) bool {
				return a.Timestamp == b.Timestamp
			},
		)
	}

	// Some streams may now be empty; remove them.
	removeEmptyStreams(streams)

	return streams, bounds
}

// Apply some postprocessing to the records in the array:
//
// - compute the cpu_util_pct field from cputime_sec and timestamp for consecutive records in
//   streams
// - remove records for which the cpu utilization is less than zero
//
// This updates the individual streams and will also remove empty streams from the set.

func computePerSampleFields(streams InputStreamSet) {
	// For each stream, compute the cpu_util_pct field of each record.
	//
	// For v0.7.0 and later, compute this as the difference in cputime_sec between adjacent records
	// divided by the time difference between them.  The first record gets a copy of the cpu_pct
	// field.
	//
	// For v0.6.0 and earlier, we don't have cputime_sec.  The best we can do with the data we have
	// is copy the cpu_pct field into cpu_util_pct.
	for _, stream := range streams {
		// By construction, every stream is non-empty.
		es := *stream
		es[0].CpuUtilPct = es[0].CpuPct
		major, minor, _ := parseVersion(es[0].Version)
		if major == 0 && minor <= 6 {
			for i := 1; i < len(es); i++ {
				es[i].CpuUtilPct = es[i].CpuPct
			}
		} else {
			dst := 1
			src := 1
			for ; src < len(es); src++ {
				dt := float64(es[src].Timestamp - es[src-1].Timestamp)
				dc := float64(int64(es[src].CpuTimeSec) - int64(es[src-1].CpuTimeSec))
				// It can happen that dc < 0, see https://github.com/NAICNO/Jobanalyzer/issues/63.
				es[src].CpuUtilPct = float32((dc / dt) * 100)
				if es[src].CpuUtilPct >= 0 {
					es[dst] = es[src]
					dst++
				}
			}

			// Clear the tail, if there is any, just to not hang onto garbage.
			clear(es[dst:])
			*stream = es[:dst]
		}
	}

	// Some streams may now be empty; remove them.  Technically this will only be the case for v>0.6
	// streams, but it doesn't matter.
	removeEmptyStreams(streams)
}

func parseVersion(v Ustr) (major, minor, bugfix int) {
	smajor, s1, found := strings.Cut(v.String(), ".")
	if !found {
		return
	}
	sminor, sbugfix, found := strings.Cut(s1, ".")
	if !found {
		return
	}
	tmp, err := strconv.Atoi(smajor)
	if err != nil {
		return
	}
	major = tmp
	tmp, err = strconv.Atoi(sminor)
	if err != nil {
		return
	}
	minor = tmp
	tmp, err = strconv.Atoi(sbugfix)
	if err != nil {
		if pre, found := strings.CutSuffix(sbugfix, "-devel"); found {
			tmp, err = strconv.Atoi(pre)
			tmp += 100
		}
		if err != nil {
			return
		}
	}
	bugfix = tmp
	return
}

func removeEmptyStreams(streams InputStreamSet) {
	maps.DeleteFunc(streams, func(key InputStreamKey, stream *SampleStream) bool {
		return len(*stream) == 0
	})
}
