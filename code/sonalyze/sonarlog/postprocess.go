// Postprocess and clean up log data after ingestion.

package sonarlog

import (
	"sort"
	"strconv"
	"strings"

	"go-utils/config"
	"go-utils/minmax"
)

// The InputStreamKey is (hostname, stream-id, cmd), where the stream-id is defined below; it is
// meaningful only for non-merged streams.
//
// An InputStreamSet maps a InputStreamKey to a SampleStream pertinent to that key.  It is named as
// it is because the InputStreamKey is meaningful only for non-merged streams.

type InputStreamKey struct {
	Host     Ustr
	StreamId uint32
	Cmd      Ustr
}

// The streams are heap-allocated so that we can update them without also updating the map.
//
// After postprocessing, there are some important invariants on the records that make up an input
// stream in addition to them having the same key:
//
// - the vector is sorted ascending by timestamp
// - no two adjacent timestamps are the same

type InputStreamSet map[InputStreamKey]*SampleStream

// Apply postprocessing to the records in the array:
//
// - reconstruct individual sample streams
// - compute the cpu_util_pct field from cputime_sec and timestamp for consecutive records in
//   streams
// - if `configs` is not None and there is the necessary information for a given host, clean up the
//   gpumem_pct and gpumem_gb fields so that they are internally consistent
// - after all that, remove records for which the filter function returns false
//
// Returns the individual streams as a map from (hostname, id, cmd) to a vector of Samples, where
// each vector is sorted in ascending order of time.  In each vector, there may be no adjacent
// records with the same timestamp.
//
// The id is necessary to distinguish the different event streams for a single job.  Consider a run
// of records from the same host.  There may be multiple records per job in that run, and they may
// or may not also have the same cmd, and they may or may not have been rolled up.  There are two
// cases:
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

const JobIdTag = 10000000

func PostprocessLog(
	entries SampleStream,
	filter func(*Sample) bool,
	configs *config.ClusterConfig,
) InputStreamSet {
	streams := make(InputStreamSet)

	// Reconstruct the individual sample streams.  Records for job id 0 are always not rolled up and
	// we'll use the pid, which is unique.  But consumers of the data must be sure to treat job id 0
	// specially.
	for _, e := range entries {
		syntheticPid := e.Pid
		if e.Rolledup > 0 {
			syntheticPid = JobIdTag + e.Job
		}
		key := InputStreamKey{e.Host, syntheticPid, e.Cmd}
		if stream, found := streams[key]; found {
			*stream = append(*stream, e)
		} else {
			streams[key] = &SampleStream{e}
		}
	}

	// Sort the streams by ascending timestamp.
	for _, stream := range streams {
		sort.Stable(TimeSortableSampleStream(*stream))
	}

	// Remove duplicate timestamps.  These may appear due to system effects, notably, sonar log
	// generation may be delayed due to disk waits, which may be long because network disks may
	// go away.  It should not matter which of the duplicate records we remove here, they should
	// be identical.
	for _, stream := range streams {
		es := *stream
		good := 0
		candidate := good + 1
		// Invariant: good < len(es)
		// Invariant: es[good] is a record we will keep
		// Invariant: candidate > good points to unexamined record or past end
		for candidate < len(es) {
			// Skip until end or we find a record that has different time
			for candidate < len(es) && es[good].Timestamp == es[candidate].Timestamp {
				candidate++
			}
			// Keep the new candidate, if there is one
			if candidate < len(es) {
				good++
				es[good] = es[candidate]
				candidate++
			}
		}
		*stream = es[:good+1]
	}

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
			for i := 1; i < len(es); i++ {
				dt := float64(es[i].Timestamp - es[i-1].Timestamp)
				dc := float64(int64(es[i].CpuTimeSec) - int64(es[i-1].CpuTimeSec))
				// It can happen that dc < 0, see https://github.com/NAICNO/Jobanalyzer/issues/63.
				// We filter these below.
				es[i].CpuUtilPct = float32((dc / dt) * 100)
			}
		}
	}

	// For each stream, clean up the gpumem_pct and gpumem_gb fields based on system information, if
	// available.
	if configs != nil {
		for _, stream := range streams {
			if conf := configs.LookupHost((*stream)[0].Host.String()); conf != nil {
				cardsizeKib := float64(conf.GpuMemGB) * 1024 * 1024 / float64(conf.GpuCards)
				for _, entry := range *stream {
					if conf.GpuMemPct {
						entry.GpuKib = uint64(float64(entry.GpuMemPct) * cardsizeKib)
					} else {
						entry.GpuMemPct = float32(float64(entry.GpuKib) / cardsizeKib)
					}
				}
			}
		}
	}

	// Remove elements that don't pass the filter and pack the array.  This preserves ordering.
	for _, stream := range streams {
		dst := 0
		es := *stream
		for src := range es {
			// See comments above re the test for cpu_util_pct
			if (filter == nil || filter(es[src])) && es[src].CpuUtilPct >= 0 {
				es[dst] = es[src]
				dst++
			}
		}
		*stream = es[:dst]
	}

	// Some streams may now be empty; remove them.
	dead := make([]InputStreamKey, 0)
	for key, stream := range streams {
		if len(*stream) == 0 {
			dead = append(dead, key)
		}
	}
	for _, key := range dead {
		delete(streams, key)
	}

	return streams
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
		return
	}
	bugfix = tmp
	return
}

// Bounds are computed from a (normally unsorted) sample stream, because we compute bounds on the
// entire input set before filtering or other postprocessing.  This is closest to what's expected.
func ComputeTimeBounds(samples SampleStream) Timebounds {
	bounds := make(Timebounds)
	for _, s := range samples {
		host := s.Host
		if bound, found := bounds[host]; found {
			bounds[host] = Timebound{
				Earliest: minmax.MinInt64(bound.Earliest, s.Timestamp),
				Latest:   minmax.MaxInt64(bound.Latest, s.Timestamp),
			}
		} else {
			bounds[host] = Timebound{
				Earliest: s.Timestamp,
				Latest:   s.Timestamp,
			}
		}
	}
	return bounds
}

// Utility for applying the filter if we're not using PostprocessLog.
func ApplyFilter(
	readings SampleStream,
	recordFilter func(*Sample) bool,
) SampleStream {
	if recordFilter != nil {
		dest := 0
		for src := 0; src < len(readings); src++ {
			if recordFilter(readings[src]) {
				readings[dest] = readings[src]
				dest++
			}
		}
		readings = readings[:dest]
	}
	return readings
}
