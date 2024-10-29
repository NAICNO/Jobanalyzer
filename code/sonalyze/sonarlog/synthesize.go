// Logic for merging sample streams.

package sonarlog

import (
	"math"
	"strings"

	"go-utils/gpuset"
	"go-utils/hostglob"
	"go-utils/maps"
	"go-utils/slices"
	. "sonalyze/common"
	"sonalyze/db"
)

const (
	farFuture int64 = math.MaxInt64
)

// Merge streams that have the same host and job ID into synthesized data.
//
// Each output stream is sorted ascending by timestamp.  No two records have exactly the same time.
// All records within a stream have the same host, command, user, and job ID.
//
// The command name for synthesized data collects all the commands that went into the synthesized
// stream.

func MergeByHostAndJob(streams InputStreamSet) SampleStreams {
	type HostAndJob struct {
		host Ustr
		job  uint32
	}

	type CommandsAndSampleStreams struct {
		commands map[Ustr]bool
		streams  SampleStreams
	}

	// The value is a set of command names and a vector of the individual streams.
	collections := make(map[HostAndJob]*CommandsAndSampleStreams)

	// The value is a map (by host) of the individual streams with job ID zero, these can't be
	// merged and must just be passed on.
	zero := make(map[Ustr]*SampleStreams)

	// Partition into jobs with ID and jobs without, and group the former by host and job and
	// collect information about each bag.
	for key, stream := range streams {
		id := (*stream)[0].Job
		if id == 0 {
			if bag, found := zero[key.Host]; found {
				*bag = append(*bag, stream)
			} else {
				zero[key.Host] = &SampleStreams{stream}
			}
		} else {
			k := HostAndJob{key.Host, id}
			if box, found := collections[k]; found {
				box.commands[key.Cmd] = true
				box.streams = append(box.streams, stream)
			} else {
				collections[k] = &CommandsAndSampleStreams{
					commands: map[Ustr]bool{key.Cmd: true},
					streams:  SampleStreams{stream},
				}
			}
		}
	}

	merged := make(SampleStreams, 0)
	for key, cmdsAndStreams := range collections {
		if zeroes, found := zero[key.host]; found {
			delete(zero, key.host)
			merged = append(merged, *zeroes...)
		}
		commands := maps.Keys(cmdsAndStreams.commands)
		UstrSortAscending(commands)
		username := mergedUserName(cmdsAndStreams.streams)
		merged = append(merged, mergeStreams(
			key.host,
			UstrJoin(commands, StringToUstr(",")),
			username,
			key.job,
			cmdsAndStreams.streams,
		))
	}

	return merged
}

// Any user from any record is fine because we really should never be merging records from different
// users.  However, in some testing scenarios we sometimes do merge records from different users and
// to make that predictable we sort and join the names.
func mergedUserName(streams SampleStreams) Ustr {
	nameset := make(map[Ustr]bool)
	for _, s := range streams {
		nameset[(*s)[0].User] = true
	}
	names := maps.Keys(nameset)
	UstrSortAscending(names)
	return UstrJoin(names, StringToUstr(","))
}

// Merge streams that have the same job ID (across hosts) into synthesized data.
//
// Each output stream is sorted ascending by timestamp.  No two records have exactly the same time.
// All records within an output stream have the same host name, job ID, command name, and user.
//
// The command name for synthesized data collects all the commands that went into the synthesized
// stream.  The host name for synthesized data collects all the hosts that went into the
// synthesized stream.
//
// This must also merge the metadata from the different hosts: the time bounds.  For a merged
// stream, the "earliest" time is the min across the earliest times for the different host streams
// that go into the merged stream, and the "latest" time is the max across the latest times ditto.

func MergeByJob(streams InputStreamSet, bounds Timebounds) (SampleStreams, Timebounds) {
	type jobDataTy struct {
		commands map[Ustr]bool
		hosts    map[Ustr]bool
		streams  SampleStreams
	}

	// The key is the job ID
	collections := make(map[uint32]*jobDataTy)

	// The value is a vector of the individual streams with job ID zero, these can't be merged and
	// must just be passed on.
	zero := make(SampleStreams, 0)

	// Distribute the streams into `zero` or some box in `collections`
	for k, v := range streams {
		id := (*v)[0].Job
		if id == 0 {
			zero = append(zero, v)
		} else if jobData, ok := collections[id]; ok {
			jobData.commands[k.Cmd] = true
			jobData.hosts[k.Host] = true
			jobData.streams = append(jobData.streams, v)
		} else {
			collections[id] = &jobDataTy{
				commands: map[Ustr]bool{k.Cmd: true},
				hosts:    map[Ustr]bool{k.Host: true},
				streams:  SampleStreams{v},
			}
		}
	}

	// The key is the host name
	newBounds := make(map[Ustr]Timebound)

	// Initialize the set of new bounds with the zero jobs
	for _, z := range zero {
		hostname := (*z)[0].Host
		if _, found := newBounds[hostname]; !found {
			probe, found := bounds[hostname]
			if !found {
				panic("Host should be in bounds")
			}
			newBounds[hostname] = probe
		}
	}

	// Initialize the set of result streams with the zero jobs
	var newStreams SampleStreams = zero

	// Iterate across the non-zero jobs and update newBounds with merged bounds and newStreams with
	// merged streams.
	for jobId, jobData := range collections {
		names := maps.MapKeys(jobData.hosts, Ustr.String)
		hostname := StringToUstr(strings.Join(hostglob.CompressHostnames(names), ","))
		if _, found := newBounds[hostname]; !found {
			if len(jobData.hosts) == 0 {
				panic("Host list should not be empty")
			}
			earliest := farFuture
			latest := int64(0)
			for host := range jobData.hosts {
				probe := bounds[host]
				earliest = min(earliest, probe.Earliest)
				latest = max(latest, probe.Latest)
			}
			newBounds[hostname] = Timebound{earliest, latest}
		}
		commands := maps.Keys(jobData.commands)
		UstrSortAscending(commands)
		user := mergedUserName(jobData.streams)
		newStreams = append(newStreams, mergeStreams(
			hostname,
			UstrJoin(commands, StringToUstr(",")),
			user,
			jobId,
			jobData.streams,
		))
	}

	return newStreams, newBounds
}

// Merge streams that have the same host (across jobs) into synthesized data.
//
// Each output stream is sorted ascending by timestamp.  No two records have exactly the same time.
// All records within an output stream have the same host name, job ID, command name, and user.
//
// The command name and user name for synthesized data are "_merged_".  It would be possible to do
// something more interesting, such as aggregating them.
//
// The job ID for synthesized data is 0, which is not ideal but probably OK so long as the consumer
// knows it.

func MergeByHost(streams InputStreamSet) SampleStreams {
	// The key is the host name.
	collections := make(map[Ustr]SampleStreams)

	for k, s := range streams {
		// This lumps jobs with job ID 0 in with the others.
		if vs, ok := collections[k.Host]; ok {
			collections[k.Host] = append(vs, s)
		} else {
			collections[k.Host] = SampleStreams{s}
		}
	}

	vs := make(SampleStreams, 0)
	cmdname := StringToUstr("_merged_")
	username := cmdname
	jobId := uint32(0)
	for hostname, streams := range collections {
		vs = append(vs, mergeStreams(hostname, cmdname, username, jobId, streams))
	}

	return vs
}

// TODO: DOCUMENTME

func MergeAcrossHostsByTime(streams SampleStreams) SampleStreams {
	if len(streams) == 0 {
		return streams
	}
	names := slices.Map(streams, func(s *SampleStream) string {
		return (*s)[0].Host.String()
	})
	hostname := StringToUstr(strings.Join(hostglob.CompressHostnames(names), ","))
	tmp := mergeStreams(
		hostname,
		StringToUstr("_merged_"),
		StringToUstr("_merged_"),
		0,
		streams,
	)
	return []*SampleStream{tmp}
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

func mergeStreams(
	hostname Ustr,
	command Ustr,
	username Ustr,
	jobId uint32,
	streams SampleStreams,
) *SampleStream {
	// Generated records
	records := make(SampleStream, 0)

	v000 := StringToUstr("0.0.0")

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
	indices := make([]int, len(streams))

	// Streams that have moved into the past have their indices[i] value set to STREAM_ENDED, this
	// enables some fast filtering, described later.
	const StreamEnded = math.MaxInt

	// selected holds the records selected by the second inner loop, we allocate it once.
	selected := make([]Sample, 0, len(streams))

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
	// TODO: IMPROVEME: The inner loop counts could be reduced significantly if we could partition
	// the streams array precisely into streams that are expired, current, and in the future.
	// However, the current tests are very quick, and any scheme to introduce that partitioning must
	// be very, very cheap, and benefits may not show until the number of inputs is exceptionally
	// large (perhaps 90 or more days of data instead of the 30 days of data I've been testing
	// with).  In addition, attempts at implementing this partitioning have so far resulted in major
	// slowdowns, possibly because the resulting code confuses bounds checking optimizations in the
	// Rust compiler.  This needs to be investigated further.

	// The first stream that is known not to be expired.
	live := 0

	sentinelTime := farFuture
	for {
		// You'd think that it'd be better to have this loop down below where values are set to
		// STREAM_ENDED, but empirically it's better to have it here.  The difference is fairly
		// pronounced.
		for live < len(streams) && indices[live] == StreamEnded {
			live++
		}

		// Loop across streams to find smallest head.
		minTime := sentinelTime
		for i := live; i < len(streams); i++ {
			if indices[i] >= len(*streams[i]) {
				continue
			}
			// stream[i] has a value, select this stream if we have no stream or if the value is
			// smaller than the one at the head of the smallest stream.
			if minTime > (*streams[i])[indices[i]].Timestamp {
				minTime = (*streams[i])[indices[i]].Timestamp
			}
		}

		// Exit if no values in any stream
		if minTime == sentinelTime {
			break
		}

		limTime := minTime + 10
		nearPast := minTime - 30
		deepPast := minTime - 60

		// Now select values from all streams (either a value in the time window or the most recent
		// value before the time window) and advance the stream pointers for the ones in the window.
		//
		// The cases marked "highly likely" get most of the hits in long runs, then the case marked
		// "fairly likely" gets one hit per record, usually, and then the case that retires a stream
		// gets one hit per stream.

		for i := live; i < len(streams); i++ {
			s := *streams[i]
			ix := indices[i]
			lim := len(s)

			// lim > 0 because no stream is empty

			if ix < lim {
				// Live or future stream.

				// ix < lim

				if s[ix].Timestamp >= limTime {
					// Highly likely - the stream starts in the future.
					continue
				}

				// ix < lim
				// s[ix].timestamp < lim_time

				if s[ix].Timestamp == minTime {
					// Fairly likely in normal input - sample time is equal to the min_time.  This
					// would be subsumed by the following test using >= for > but the equality test
					// is faster.
					selected = append(selected, s[ix])
					indices[i]++
					continue
				}

				// ix < lim
				// s[ix].timestamp < lim_time
				// s[ix].timestamp != min_time

				if s[ix].Timestamp > minTime {
					// Unlikely in normal input - sample time is in in the time window but not equal
					// to min_time.
					selected = append(selected, s[ix])
					indices[i]++
					continue
				}

				// ix < lim
				// s[ix].timestamp < min_time

				if ix > 0 && s[ix-1].Timestamp >= nearPast {
					// Unlikely in normal input - Previous exists and is not last and is in the near
					// past (redundant test for t < lim_time removed).  The condition is tricky.
					// ix>0 guarantees that there is a past record at ix - 1, while ix<lim says that
					// there is also a future record at ix.
					//
					// This is hard to make reliable.  The guard on the time is necessary to avoid
					// picking up records from a lot of dead processes.  Intra-host it is OK.
					// Cross-host it depends on sonar runs being more or less synchronized.
					selected = append(selected, s[ix-1])
					continue
				}

				// ix < lim
				// s[ix].timestamp < min_time
				// s[ix-1].timestamp < near_past

				// This is duplicated within the ix==lim nest below, in a different form.
				if ix > 0 && s[ix-1].Timestamp >= deepPast {
					// Previous exists (and is last) and is not in the deep past, pick it up
					selected = append(selected, s[ix-1])
					continue
				}

				// ix < lim
				// s[ix].timestamp < min_time
				// s[ix-1].timestamp < deep_past

				// This is an old record and we can ignore it.
				continue
			} else if ix == StreamEnded {
				// Highly likely - stream already marked as exhausted.
				continue
			} else {
				// About-to-be exhausted stream.

				// ix == lim
				// ix > 0 because lim > 0

				if s[ix-1].Timestamp < deepPast {
					// Previous is in the deep past and no current - stream is done.
					indices[i] = StreamEnded
					continue
				}

				// ix == lim
				// ix > 0
				// s[ix-1].timestamp >= deep_past

				// This case is a duplicate from the ix<lim nest above, in a different form.
				if s[ix-1].Timestamp < minTime {
					selected = append(selected, s[ix-1])
					continue
				}

				// ix == lim
				// ix > 0
				// s[ix-1].timestamp >= min_time

				// This is a contradiction probably and it seems we should not come this far.  Don't
				// worry about it.
				continue
			}
		}

		records = append(records, sumRecords(
			v000,
			minTime,
			hostname,
			username,
			jobId,
			command,
			selected,
		))
		selected = selected[0:0]
	}

	return &records
}

func MergeGpuFail(a, b uint8) uint8 {
	if a > 0 || b > 0 {
		return 1
	}
	return 0
}

func sumRecords(
	version Ustr,
	timestamp int64,
	hostname Ustr,
	username Ustr,
	jobId uint32,
	command Ustr,
	selected []Sample,
) Sample {
	var cpuPct, gpuPct, gpuMemPct, cpuUtilPct float32
	var cpuKB, rssAnonKB, gpuKB, cpuTimeSec uint64
	var rolledup uint32
	var gpuFail uint8
	var gpus = gpuset.EmptyGpuSet()
	for _, s := range selected {
		cpuPct += s.CpuPct
		gpuPct += s.GpuPct
		gpuMemPct += s.GpuMemPct
		cpuUtilPct += s.CpuUtilPct
		cpuKB += s.CpuKB
		rssAnonKB += s.RssAnonKB
		gpuKB += s.GpuKB
		cpuTimeSec += s.CpuTimeSec
		rolledup += s.Rolledup
		gpuFail = MergeGpuFail(gpuFail, s.GpuFail)
		gpus = gpuset.UnionGpuSets(gpus, s.Gpus)
	}
	// The invariant is that rolledup is the number of *other* processes rolled up into this one.
	// So we add one for each in the list + the others rolled into each of those, and subtract one
	// at the end to maintain the invariant.
	rolledup -= uint32(len(selected) + 1)

	// Synthesize the record.
	return Sample{
		Sample: &db.Sample{
			Version:    version,
			Timestamp:  timestamp,
			Host:       hostname,
			User:       username,
			Job:        jobId,
			Cmd:        command,
			CpuPct:     cpuPct,
			CpuKB:      cpuKB,
			RssAnonKB:  rssAnonKB,
			Gpus:       gpus,
			GpuPct:     gpuPct,
			GpuMemPct:  gpuMemPct,
			GpuKB:      gpuKB,
			GpuFail:    gpuFail,
			CpuTimeSec: cpuTimeSec,
			Rolledup:   rolledup,
		},
		CpuUtilPct: cpuUtilPct,
	}
}

func FoldSamplesHalfHourly(samples SampleStream) SampleStream {
	return foldSamples(samples, TruncateToHalfHour)
}

func FoldSamplesHourly(samples SampleStream) SampleStream {
	return foldSamples(samples, TruncateToHour)
}

func FoldSamplesHalfDaily(samples SampleStream) SampleStream {
	return foldSamples(samples, TruncateToHalfDay)
}

func FoldSamplesDaily(samples SampleStream) SampleStream {
	return foldSamples(samples, TruncateToDay)
}

func FoldSamplesWeekly(samples SampleStream) SampleStream {
	return foldSamples(samples, TruncateToWeek)
}

func foldSamples(samples SampleStream, truncTime func(int64) int64) SampleStream {
	result := make(SampleStream, 0)
	i := 0
	v000 := StringToUstr("0.0.0")
	merged := StringToUstr("_merged_")
	for i < len(samples) {
		first := i
		s0 := samples[i]
		t0 := truncTime(s0.Timestamp)
		i++
		for i < len(samples) && truncTime(samples[i].Timestamp) == t0 {
			i++
		}
		lim := i
		r := sumRecords(
			v000,
			t0,
			s0.Host,
			merged,
			0,
			merged,
			samples[first:lim],
		)
		nf32 := float32(lim - first)
		nu64 := uint64(lim - first)
		r.CpuPct /= nf32
		r.CpuKB /= nu64
		r.RssAnonKB /= nu64
		r.GpuPct /= nf32
		r.GpuMemPct /= nf32
		r.GpuKB /= nu64
		r.CpuTimeSec /= nu64
		r.CpuUtilPct /= nf32
		result = append(result, r)
	}

	return result
}
