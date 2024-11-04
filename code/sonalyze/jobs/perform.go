package jobs

import (
	"io"
	"math"
	"strings"

	"go-utils/config"
	"go-utils/gpuset"
	"go-utils/hostglob"
	. "sonalyze/command"
	. "sonalyze/common"
	"sonalyze/db"
	"sonalyze/sonarlog"
)

// Computed float64 fields in jobAggregate.computed
const (
	kCpuPctAvg      = iota // Average CPU utilization, 1 core == 100%
	kCpuPctPeak            // Peak CPU utilization ditto
	kRcpuPctAvg            // Average CPU utilization, all cores == 100%
	kRcpuPctPeak           // Peak CPU utilization ditto
	kCpuGBAvg              // Average main memory utilization, GiB
	kCpuGBPeak             // Peak memory utilization ditto
	kRcpuGBAvg             // Average main memory utilization, all memory = 100%
	kRcpuGBPeak            // Peak memory utilization ditto
	kRssAnonGBAvg          // Average resident main memory utilization, GiB
	kRssAnonGBPeak         // Peak memory utilization ditto
	kRrssAnonGBAvg         // Average resident main memory utilization, all memory = 100%
	kRrssAnonGBPeak        // Peak memory utilization ditto
	kGpuPctAvg             // Average GPU utilization, 1 card == 100%
	kGpuPctPeak            // Peak GPU utilization ditto
	kRgpuPctAvg            // Average GPU utilization, all cards == 100%
	kRgpuPctPeak           // Peak GPU utilization ditto
	kGpuGBAvg              // Average GPU memory utilization, GiB
	kGpuGBPeak             // Peak memory utilization ditto
	kRgpuGBAvg             // Average GPU memory utilization, all cards == 100%
	kRgpuGBPeak            // Peak GPU memory utilization ditto
	kSgpuPctAvg            // Average GPU utilization, all cards used by job == 100%
	kSgpuPctPeak           // Peak GPU utilization, all cards used by job == 100%
	kSgpuGBAvg             // Average GPU memory utilization, all cards used by job == 100%
	kSgpuGBPeak            // Peak GPU memory utilization ditto
	kDuration              // Duration of job in seconds (wall clock, not CPU)
	numF64Fields
)

// Computed flag bits in jobAggregate.computedFlags
const (
	kUsesGpu          = (1 << iota) // True if there's reason to believe a gpu was used by job
	kDoesNotUseGpu                  // Opposite
	kGpuFail                        // GPU failed
	kIsLiveAtStart                  // Job had record at earliest timestamp of input set for host
	kIsNotLiveAtStart               // Opposite
	kIsLiveAtEnd                    // Job had record at latest timestamp of input set for host
	kIsNotLiveAtEnd                 // Opposite
	kIsZombie                       // Command contains <defunct> or user starts with _zombie_
)

// Package for results from aggregation.
type jobSummary struct {
	aggregate jobAggregate
	job       sonarlog.SampleStream
}

// Aggregate figures for a job.  For some cross-job data like user and host, go to the sample stream
// in the jobSummary that owns this aggregate.
//
// The float fields of this are *not* rounded in any way.
//
// GPU memory: If a system config is present and conf.GpuMemPct is true then kGpuGB* are derived
// from the recorded percentage figure, otherwise kRgpuGB* are derived from the recorded absolute
// figures.  If a system config is not present then all fields will represent the recorded values
// (kRgpuKB * the recorded percentages).
type jobAggregate struct {
	first         int64 // Earliest time seen for the job, seconds since epoch
	last          int64 // Latest time ditto
	selected      bool  // Initially true, used to deselect the record before printing
	computedFlags int
	computed      [numF64Fields]float64
}

func (jc *JobsCommand) NeedsBounds() bool {
	return true
}

func (jc *JobsCommand) Perform(
	out io.Writer,
	cfg *config.ClusterConfig,
	_ db.SampleCluster,
	streams sonarlog.InputStreamSet,
	bounds sonarlog.Timebounds,
	hostGlobber *hostglob.HostGlobber,
	_ *db.SampleFilter,
) error {
	if jc.Verbose {
		Log.Infof("Streams constructed by postprocessing: %d", len(streams))
		numSamples := 0
		for _, stream := range streams {
			numSamples += len(*stream)
		}
		Log.Infof("Samples retained after filtering: %d", numSamples)
	}

	if jc.printRequiresConfig() {
		var err error
		streams, err = EnsureConfigForInputStreams(cfg, streams, "relative format arguments")
		if err != nil {
			return err
		}
	}

	summaries := jc.aggregateAndFilterJobs(cfg, streams, bounds)
	if jc.Verbose {
		Log.Infof("Jobs after aggregation filtering: %d", len(summaries))
	}

	return jc.printJobSummaries(out, summaries)
}

// A sample stream is a quadruple (host, command, job-related-id, record-list).  A stream is only
// ever about one job.  There may be multiple streams per job, they will all have the same
// job-related-id which is unique but not necessarily equal to any field in any of the records.
//
// This function collects the data per job and returns a vector of (aggregate, records) pairs where
// the aggregate describes the job in aggregate and the records is a synthesized stream of sample
// records for the job, based on all the input streams for the job.  The manner of the synthesis
// depends on arguments to the program: with --merge-all we merge across all hosts; with
// --merge-none we do not merge; otherwise the config file can specify the hosts to merge across;
// otherwise if there is no config we do not merge.

func (jc *JobsCommand) aggregateAndFilterJobs(
	cfg *config.ClusterConfig,
	streams sonarlog.InputStreamSet,
	bounds sonarlog.Timebounds,
) []*jobSummary {
	var anyMergeableNodes bool
	if !jc.MergeNone && cfg != nil {
		anyMergeableNodes = cfg.HasCrossNodeJobs()
	}

	var jobs sonarlog.SampleStreams
	if jc.MergeAll {
		jobs, bounds = sonarlog.MergeByJob(streams, bounds)
	} else if anyMergeableNodes {
		jobs, bounds = mergeAcrossSomeNodes(cfg, streams, bounds)
	} else {
		jobs = sonarlog.MergeByHostAndJob(streams)
	}
	if jc.Verbose {
		Log.Infof("Jobs constructed by merging: %d", len(jobs))
	}

	filter := jc.buildAggregationFilter(cfg)

	summaries := make([]*jobSummary, 0)
	minSamples := jc.lookupUint("min-samples")
	if jc.Verbose && minSamples > 1 {
		Log.Infof("Excluding jobs with fewer than %d samples", minSamples)
	}
	discarded := 0
	for _, job := range jobs {
		if uint(len(*job)) >= minSamples {
			aggregate := &jobSummary{
				aggregate: jc.aggregateJob(cfg, *job, bounds),
				job:       *job,
			}
			if filter == nil || filter.apply(aggregate) {
				summaries = append(summaries, aggregate)
			}
		} else {
			discarded++
		}
	}
	if jc.Verbose {
		Log.Infof("Jobs discarded by aggregation filtering: %d", discarded)
	}

	return summaries
}

// Look to the config to find nodes that have CrossNodeJobs set, and merge their streams as if by
// --merge-all; the remaining streams are merged as if by --merge-none, and the two sets of merged
// jobs are combined into one set.

func mergeAcrossSomeNodes(
	cfg *config.ClusterConfig,
	streams sonarlog.InputStreamSet,
	bounds sonarlog.Timebounds,
) (sonarlog.SampleStreams, sonarlog.Timebounds) {
	mergeable := make(sonarlog.InputStreamSet)
	mBounds := make(sonarlog.Timebounds)
	solo := make(sonarlog.InputStreamSet)
	sBounds := make(sonarlog.Timebounds)
	for k, v := range streams {
		bound := bounds[k.Host]
		if sys := cfg.LookupHost(k.Host.String()); sys != nil && sys.CrossNodeJobs {
			mBounds[k.Host] = bound
			mergeable[k] = v
		} else {
			sBounds[k.Host] = bound
			solo[k] = v
		}
	}
	mergedJobs, mergedBounds := sonarlog.MergeByJob(mergeable, mBounds)
	otherJobs := sonarlog.MergeByHostAndJob(solo)
	mergedJobs = append(mergedJobs, otherJobs...)
	for k, v := range sBounds {
		mergedBounds[k] = v
	}
	return mergedJobs, mergedBounds
}

// Given a list of log entries for a job, sorted ascending by timestamp and with no duplicated
// timestamps, and the earliest and latest timestamps from all records read, return a JobAggregate
// for the job.

func (jc *JobsCommand) aggregateJob(
	cfg *config.ClusterConfig,
	job sonarlog.SampleStream,
	bounds sonarlog.Timebounds,
) jobAggregate {
	first := job[0].Timestamp
	last := job[len(job)-1].Timestamp
	host := job[0].Host
	duration := last - first
	needZombie := jc.Zombie
	gpus := gpuset.EmptyGpuSet()
	var (
		gpuFail                       uint8
		cpuPctAvg, cpuPctPeak         float64
		rCpuPctAvg, rCpuPctPeak       float64
		cpuGBAvg, cpuGBPeak           float64
		rCpuGBAvg, rCpuGBPeak         float64
		gpuPctAvg, gpuPctPeak         float64
		rGpuPctAvg, rGpuPctPeak       float64
		sGpuPctAvg, sGpuPctPeak       float64
		rssAnonGBAvg, rssAnonGBPeak   float64
		rRssAnonGBAvg, rRssAnonGBPeak float64
		gpuGBAvg, gpuGBPeak           float64
		rGpuGBAvg, rGpuGBPeak         float64
		sGpuGBAvg, sGpuGBPeak         float64
		flags                         int
		isZombie                      bool
	)
	const kb2gb = 1.0 / (1024 * 1024)

	for _, s := range job {
		gpus = gpuset.UnionGpuSets(gpus, s.Gpus)
		gpuFail = sonarlog.MergeGpuFail(gpuFail, s.GpuFail)
		cpuPctAvg += float64(s.CpuUtilPct)
		cpuPctPeak = math.Max(cpuPctPeak, float64(s.CpuUtilPct))
		gpuPctAvg += float64(s.GpuPct)
		gpuPctPeak = math.Max(gpuPctPeak, float64(s.GpuPct))
		cpuGBAvg += float64(s.CpuKB) * kb2gb
		cpuGBPeak = math.Max(cpuGBPeak, float64(s.CpuKB)*kb2gb)
		rssAnonGBAvg += float64(s.RssAnonKB) * kb2gb
		rssAnonGBPeak = math.Max(rssAnonGBPeak, float64(s.RssAnonKB)*kb2gb)
		gpuGBAvg += float64(s.GpuKB) * kb2gb
		gpuGBPeak = math.Max(gpuGBPeak, float64(s.GpuKB)*kb2gb)

		if needZombie && !isZombie {
			cmd := s.Cmd.String()
			isZombie = strings.Contains(cmd, "<defunct>") || strings.HasPrefix(cmd, "_zombie_")
		}
	}
	usesGpu := !gpus.IsEmpty()

	if cfg != nil {
		if sys := cfg.LookupHost(host.String()); sys != nil {
			// Quantities can be zero in surprising ways, so always guard divisions
			if cores := float64(sys.CpuCores); cores > 0 {
				rCpuPctAvg = cpuPctAvg / cores
				rCpuPctPeak = cpuPctPeak / cores
			}
			if memory := float64(sys.MemGB); memory > 0 {
				rCpuGBAvg = (cpuGBAvg * 100) / memory
				rCpuGBPeak = (cpuGBPeak * 100) / memory
				rRssAnonGBAvg = (rssAnonGBAvg * 100) / memory
				rRssAnonGBPeak = (rssAnonGBPeak * 100) / memory
			}
			if gpuCards := float64(sys.GpuCards); gpuCards > 0 {
				rGpuPctAvg = gpuPctAvg / gpuCards
				rGpuPctPeak = gpuPctPeak / gpuCards
			}
			if gpuMemory := float64(sys.GpuMemGB); gpuMemory > 0 {
				// As we have a config, logclean will have computed proper GPU memory values for the
				// job, so we need not look to sys.GpuMemPct here.
				rGpuGBAvg = (gpuGBAvg * 100) / gpuMemory
				rGpuGBPeak = (gpuGBPeak * 100) / gpuMemory
			}
			if usesGpu {
				nCards := float64(gpus.Size())
				sGpuPctAvg = gpuPctAvg / nCards
				sGpuPctPeak = gpuPctPeak / nCards
				if gpuCards := float64(sys.GpuCards); gpuCards > 0 {
					if gpuMemory := float64(sys.GpuMemGB); gpuMemory > 0 {
						jobGpuGB := nCards * (gpuMemory / gpuCards)
						sGpuGBAvg = (gpuGBAvg * 100) / jobGpuGB
						sGpuGBPeak = (gpuGBPeak * 100) / jobGpuGB
					}
				}
			}
		}
	}

	if usesGpu {
		flags |= kUsesGpu
	} else {
		flags |= kDoesNotUseGpu
	}
	if gpuFail != 0 {
		flags |= kGpuFail
	}
	bound, haveBound := bounds[host]
	if !haveBound {
		panic("Expected to find bound")
	}
	if first == bound.Earliest {
		flags |= kIsLiveAtStart
	} else {
		flags |= kIsNotLiveAtStart
	}
	if last == bound.Latest {
		flags |= kIsLiveAtEnd
	} else {
		flags |= kIsNotLiveAtEnd
	}
	if isZombie {
		flags |= kIsZombie
	}
	n := float64(len(job))
	a := jobAggregate{
		first:         first,
		last:          last,
		selected:      true,
		computedFlags: flags,
	}
	a.computed[kCpuPctAvg] = cpuPctAvg / n
	a.computed[kCpuPctPeak] = cpuPctPeak
	a.computed[kRcpuPctAvg] = rCpuPctAvg / n
	a.computed[kRcpuPctPeak] = rCpuPctPeak

	a.computed[kCpuGBAvg] = cpuGBAvg / n
	a.computed[kCpuGBPeak] = cpuGBPeak
	a.computed[kRcpuGBAvg] = rCpuGBAvg / n
	a.computed[kRcpuGBPeak] = rCpuGBPeak

	a.computed[kRssAnonGBAvg] = rssAnonGBAvg / n
	a.computed[kRssAnonGBPeak] = rssAnonGBPeak
	a.computed[kRrssAnonGBAvg] = rRssAnonGBAvg / n
	a.computed[kRrssAnonGBPeak] = rRssAnonGBPeak

	a.computed[kGpuPctAvg] = gpuPctAvg / n
	a.computed[kGpuPctPeak] = gpuPctPeak
	a.computed[kRgpuPctAvg] = rGpuPctAvg / n
	a.computed[kRgpuPctPeak] = rGpuPctPeak
	a.computed[kSgpuPctAvg] = sGpuPctAvg / n
	a.computed[kSgpuPctPeak] = sGpuPctPeak

	a.computed[kGpuGBAvg] = gpuGBAvg / n
	a.computed[kGpuGBPeak] = gpuGBPeak
	a.computed[kRgpuGBAvg] = rGpuGBAvg / n
	a.computed[kRgpuGBPeak] = rGpuGBPeak
	a.computed[kSgpuGBAvg] = sGpuGBAvg / n
	a.computed[kSgpuGBPeak] = sGpuGBPeak

	a.computed[kDuration] = float64(duration)

	return a
}

// Aggregation filters.
//
// Filtering is mostly wasted work.  Very frequently, all the filters will pass because the coarse
// filtering (job number, user, command, host) has been applied already and most of the filters
// applied to the aggregate are not very interesting to many users and will not be used to reject
// many records.
//
// There are several ways to represent a filter.  The simplest is just as a table of values
// representing min and max values.

type filterVal struct {
	limit float64
	ix    int
}

type aggregationFilter struct {
	minFilters []filterVal
	maxFilters []filterVal
	flags      int
}

func (f *aggregationFilter) apply(s *jobSummary) bool {
	for _, v := range f.minFilters {
		if s.aggregate.computed[v.ix] < v.limit {
			return false
		}
	}
	for _, v := range f.maxFilters {
		if s.aggregate.computed[v.ix] > v.limit {
			return false
		}
	}
	return (f.flags & s.aggregate.computedFlags) == f.flags
}

func (jc *JobsCommand) buildAggregationFilter(
	cfg *config.ClusterConfig,
) *aggregationFilter {
	minFilters := make([]filterVal, 0)
	maxFilters := make([]filterVal, 0)

	for _, v := range uintArgs {
		if v.aggregateIx != -1 && (cfg != nil || !v.relative) {
			val := jc.lookupUint(v.name)
			if strings.HasPrefix(v.name, "min-") && val != 0 {
				if jc.Verbose {
					Log.Infof("Excluding jobs: Min-filtering %s for %d", v.name, val)
				}
				minFilters = append(minFilters, filterVal{float64(val), v.aggregateIx})
			}
			if strings.HasPrefix(v.name, "max-") && val != v.initial {
				if jc.Verbose {
					Log.Infof("Excluding jobs: Max-filtering %s for %d", v.name, val)
				}
				maxFilters = append(maxFilters, filterVal{float64(val), v.aggregateIx})
			}
		}
	}
	if jc.MinRuntimeSec > 0 {
		// This is *running time*, not CPU time
		if jc.Verbose {
			Log.Infof("Excluding jobs: Min-filtering by elapsed time < %ds", jc.MinRuntimeSec)
		}
		minFilters = append(minFilters, filterVal{float64(jc.MinRuntimeSec), kDuration})
	}

	// For the flags, set all the conditions we care about.  They must all be set in the summary's
	// computed flags.
	flags := 0
	if jc.NoGpu {
		flags |= kDoesNotUseGpu
	}
	if jc.SomeGpu {
		flags |= kUsesGpu
	}
	if jc.Completed {
		flags |= kIsNotLiveAtEnd
	}
	if jc.Running {
		flags |= kIsLiveAtEnd
	}
	if jc.Zombie {
		flags |= kIsZombie
	}
	if jc.Verbose && flags != 0 {
		Log.Infof("Flag-filtering (UTSL): %x", flags)
	}

	if len(minFilters) == 0 && len(maxFilters) == 0 && flags == 0 {
		return nil
	}
	return &aggregationFilter{
		minFilters,
		maxFilters,
		flags,
	}
}
