// A materialized table of jobs constructed from one or more sample streams.  Input streams are
// partitioned into mergeable sets (based on flags and mergeability) and then each such set is
// aggregated into a coherent "job".
//
// TODO: Presently we construct it afresh every time, but once a job has completed the values will
// not change (for a given record filter), and the results could be stored in the database, esp if
// the sample stream is not needed by the client, as it is usually not.  However, the exact set of
// input streams and the input flags both play a role in how merging is performed and will affect
// how we can cache anything.
package samplejob

import (
	"fmt"
	"math"
	"strings"
	"time"

	"go-utils/gpuset"
	"go-utils/sonalyze"
	. "sonalyze/common"
	"sonalyze/data/sample"
	"sonalyze/db"
	. "sonalyze/table"
)

// Computed flag bits in SampleJob.ComputedFlags
const (
	KUsesGpu          = (1 << iota) // True if there's reason to believe a gpu was used by job
	KDoesNotUseGpu                  // Opposite
	KGpuFail                        // GPU failed
	KIsLiveAtStart                  // Job had record at earliest timestamp of input set for host
	KIsNotLiveAtStart               // Opposite
	KIsLiveAtEnd                    // Job had record at latest timestamp of input set for host
	KIsNotLiveAtEnd                 // Opposite
	KIsZombie                       // Command contains <defunct> or user starts with _zombie_
)

// Aggregate figures for a job synthesized from a set of sample streams.
type SampleJob struct {
	// Nonzero if any gpu in the job had a failure
	GpuFail int

	// Gpus involved in the job
	Gpus gpuset.GpuSet

	// True if any process in the job is believed to be a zombie
	IsZombie bool

	// Combined command line of processes in the job
	Cmd string

	// Combined host set of the job
	Hosts *Hostnames

	// Job's primary ID
	JobId uint32

	// Job's epoch
	Epoch uint64

	// User running the job
	User Ustr

	// Real time taken by the job
	Duration DurationValue

	// Earliest time seen for the job
	Start DateTimeValue

	// Latest time ditto
	End DateTimeValue

	// The list of sample records, may be shared with the database layer, may be nil.
	Job sample.SampleStream

	// The number of samples (even if Job is nil)
	SampleCount uint64

	// Bitwise 'or' of sonalyze.LIVE_AT_START and sonalyze.LIVE_AT_END
	Classification int

	// Bitwise 'or' of flag values above
	ComputedFlags int

	// Sum of CPU utilization figures across time steps, 1 core = 100.0
	CpuPctSum float64

	// Sum of main memory in use across time steps
	CpuKBSum uint64

	// Sum of resident main memory across time steps
	RssAnonKBSum uint64

	// Sum of GPU utilization figures across time steps, 1 card = 100.0
	GpuPctSum float64

	// Sum of GPU memory use across time steps
	GpuKBSum uint64

	// Aggregated CPU time for the job, based on utilization data
	CpuTime DurationValue

	// Aggregated GPU compute time for the job, based on utilization data
	GpuTime DurationValue

	// Maximum CPU utilization across time steps
	CpuPctMax float64

	// Maximum main memory in use across time steps
	CpuKBMax uint64

	// Maximum resident memory use across time steps
	RssAnonKBMax uint64

	// Maximum GPU utilization across time steps
	GpuPctMax float64

	// Maximum GPU memory use across time steps
	GpuKBMax uint64
}

type Merge int

const (
	MergeDefault = 0
	MergeAll     = 1
	MergeNone    = 2
)

// Query the materialized jobs with a filter.
//
// TODO: This interface is just too weird.  But it works.
func Query(
	theLog db.ProcessSampleDataProvider,
	isMergeable func(sample.InputStreamKey) bool,
	fromDate time.Time,
	toDate time.Time,
	hosts *Hosts,
	recordFilter *sample.SampleFilter,
	needRecords bool,
	merge Merge,
	verbose bool,
) (
	[]*SampleJob,
	error,
) {
	streams, bounds, read, dropped, err :=
		sample.ReadSampleStreamsAndMaybeBounds(
			theLog,
			fromDate,
			toDate,
			hosts,
			recordFilter,
			true,
			verbose,
		)
	if err != nil {
		return nil, fmt.Errorf("Failed to read log records: %v", err)
	}
	if verbose {
		Log.Infof("%d records read + %d dropped\n", read, dropped)
	}
	if verbose {
		Log.Infof("Streams constructed by postprocessing: %d", len(streams))
		numSamples := 0
		for _, stream := range streams {
			numSamples += len(*stream)
		}
		Log.Infof("Samples retained after filtering: %d", numSamples)
	}

	summaries := aggregateAndFilterJobs(isMergeable, streams, bounds, needRecords, merge, verbose)
	if verbose {
		Log.Infof("Jobs after aggregation filtering: %d", len(summaries))
	}
	return summaries, nil
}

func aggregateAndFilterJobs(
	isMergeable func(sample.InputStreamKey) bool,
	streams sample.InputStreamSet,
	bounds Timebounds,
	needRecords bool,
	merge Merge,
	verbose bool,
) []*SampleJob {
	var anyMergeableNodes bool
	if merge != MergeNone && isMergeable != nil {
		for k := range streams {
			anyMergeableNodes = isMergeable(k)
			if anyMergeableNodes {
				break
			}
		}
	}

	var jobs sample.SampleStreams
	if merge == MergeAll {
		jobs, bounds = sample.MergeByJob(streams, bounds)
	} else if anyMergeableNodes {
		jobs, bounds = mergeAcrossSomeNodes(isMergeable, streams, bounds)
	} else {
		jobs = sample.MergeByHostAndJob(streams)
	}
	if verbose {
		Log.Infof("Jobs constructed by merging: %d", len(jobs))
	}

	summaries := make([]*SampleJob, 0)
	for _, job := range jobs {
		host := (*job)[0].Hostname
		jobId := (*job)[0].Job
		epoch := (*job)[0].Epoch
		user := (*job)[0].User
		first := (*job)[0].Timestamp
		last := (*job)[len(*job)-1].Timestamp
		duration := last - first
		aggregate := computeAggregate(host, *job)
		usesGpu := !aggregate.Gpus.IsEmpty()
		flags := 0
		if usesGpu {
			flags |= KUsesGpu
		} else {
			flags |= KDoesNotUseGpu
		}
		if aggregate.GpuFail != 0 {
			flags |= KGpuFail
		}
		bound, haveBound := bounds[host]
		if !haveBound {
			panic("Expected to find bound")
		}
		if first == bound.Earliest {
			flags |= KIsLiveAtStart
		} else {
			flags |= KIsNotLiveAtStart
		}
		if last == bound.Latest {
			flags |= KIsLiveAtEnd
		} else {
			flags |= KIsNotLiveAtEnd
		}
		if aggregate.IsZombie {
			flags |= KIsZombie
		}
		classification := 0
		if (flags & KIsLiveAtStart) != 0 {
			classification |= sonalyze.LIVE_AT_START
		}
		if (flags & KIsLiveAtEnd) != 0 {
			classification |= sonalyze.LIVE_AT_END
		}
		aggregate.JobId = jobId
		aggregate.Epoch = epoch
		aggregate.User = user
		aggregate.CpuTime = DurationValue(math.Round(aggregate.CpuPctSum * float64(duration) / 100))
		aggregate.GpuTime = DurationValue(math.Round(aggregate.GpuPctSum * float64(duration) / 100))
		aggregate.Duration = DurationValue(duration)
		aggregate.Start = DateTimeValue(first)
		aggregate.End = DateTimeValue(last)
		aggregate.Classification = classification
		if needRecords {
			aggregate.Job = *job
		}
		aggregate.ComputedFlags = flags
		summaries = append(summaries, aggregate)
	}
	return summaries
}

// Given a list of samples, sorted ascending by timestamp and with no duplicated timestamps, return
// a partial SampleJob, with values that are aggregated from all the samples.
func computeAggregate(
	host Ustr,
	job sample.SampleStream,
) *SampleJob {
	gpus := gpuset.EmptyGpuSet()
	var (
		gpuFail                    uint8
		cpuPctSum, cpuPctMax       float64
		cpuKBSum, cpuKBMax         uint64
		gpuPctSum, gpuPctMax       float64
		rssAnonKBSum, rssAnonKBMax uint64
		gpuKBSum, gpuKBMax         uint64
		isZombie                   bool
	)
	for _, s := range job {
		gpus = gpuset.UnionGpuSets(gpus, s.Gpus)
		gpuFail = sample.MergeGpuFail(gpuFail, s.GpuFail)
		cpuPctSum += float64(s.CpuUtilPct)
		cpuPctMax = max(cpuPctMax, float64(s.CpuUtilPct))
		gpuPctSum += float64(s.GpuPct)
		gpuPctMax = max(gpuPctMax, float64(s.GpuPct))
		cpuKBSum += s.CpuKB
		cpuKBMax = max(cpuKBMax, s.CpuKB)
		rssAnonKBSum += s.RssAnonKB
		rssAnonKBMax = max(rssAnonKBMax, s.RssAnonKB)
		gpuKBSum += s.GpuKB
		gpuKBMax = max(gpuKBMax, s.GpuKB)

		if !isZombie {
			cmd := s.Cmd.String()
			isZombie = strings.Contains(cmd, "<defunct>") || strings.HasPrefix(cmd, "_zombie_")
		}
	}

	cmd := ""
	names := make(map[Ustr]bool)
	for _, sample := range job {
		if _, found := names[sample.Cmd]; found {
			continue
		}
		if cmd != "" {
			cmd += ", "
		}
		cmd += sample.Cmd.String()
		names[sample.Cmd] = true
	}

	var hosts = NewHostnames()
	for _, s := range job {
		hosts.Add(s.Hostname.String())
	}
	a := SampleJob{
		Gpus:         gpus,
		GpuFail:      int(gpuFail),
		Cmd:          cmd,
		Hosts:        hosts,
		IsZombie:     isZombie,
		CpuPctSum:    cpuPctSum,
		CpuPctMax:    cpuPctMax,
		CpuKBSum:     cpuKBSum,
		CpuKBMax:     cpuKBMax,
		RssAnonKBSum: rssAnonKBSum,
		RssAnonKBMax: rssAnonKBMax,
		GpuPctSum:    gpuPctSum,
		GpuPctMax:    gpuPctMax,
		GpuKBSum:     gpuKBSum,
		GpuKBMax:     gpuKBMax,
		SampleCount:  uint64(len(job)),
	}

	return &a
}

// Merge mergeable streams as if by --merge-all; the remaining streams are merged as if by
// --merge-none, and the two sets of merged streams are combined into one set.
func mergeAcrossSomeNodes(
	isMergeable func(sample.InputStreamKey) bool,
	streams sample.InputStreamSet,
	bounds Timebounds,
) (sample.SampleStreams, Timebounds) {
	mergeable := make(sample.InputStreamSet)
	mBounds := make(Timebounds)
	solo := make(sample.InputStreamSet)
	sBounds := make(Timebounds)
	for k, s := range streams {
		bound := bounds[k.Host]
		if isMergeable(k) {
			mBounds[k.Host] = bound
			mergeable[k] = s
		} else {
			sBounds[k.Host] = bound
			solo[k] = s
		}
	}
	mergedJobs, mergedBounds := sample.MergeByJob(mergeable, mBounds)
	otherJobs := sample.MergeByHostAndJob(solo)
	mergedJobs = append(mergedJobs, otherJobs...)
	for k, v := range sBounds {
		mergedBounds[k] = v
	}
	return mergedJobs, mergedBounds
}
