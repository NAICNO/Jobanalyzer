package jobs

import (
	"fmt"
	"io"
	"math"
	"slices"
	"strings"
	"time"

	"go-utils/config"
	"go-utils/gpuset"
	"go-utils/sonalyze"

	. "sonalyze/cmd"
	. "sonalyze/common"
	"sonalyze/db"
	"sonalyze/db/repr"
	"sonalyze/slurmlog"
	"sonalyze/sonarlog"
	. "sonalyze/table"
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
	jobAggregate
	JobId          uint32
	User           Ustr
	JobAndMark     string
	Now            DateTimeValue
	Duration       DurationValue
	Start          DateTimeValue // Earliest time seen for the job, seconds since epoch
	End            DateTimeValue // Latest time ditto
	CpuTime        DurationValue
	GpuTime        DurationValue
	Classification int // Bit vector of flags
	job            sonarlog.SampleStream
	computedFlags  int
	selected       bool // Initially true, used to deselect the record before printing
	sacctInfo      *repr.SacctInfo
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
	GpuFail  int
	Gpus     gpuset.GpuSet
	computed [numF64Fields]float64
	IsZombie bool
	Cmd      string
	Hosts    *Hostnames
}

func (jc *JobsCommand) NeedsBounds() bool {
	return true
}

func (jc *JobsCommand) Perform(
	out io.Writer,
	cfg *config.ClusterConfig,
	theDb db.SampleCluster,
	streams sonarlog.InputStreamSet,
	bounds sonarlog.Timebounds,
	hostGlobber *Hosts,
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

	if NeedsConfig(jobsFormatters, jc.PrintFields) {
		var err error
		streams, err = EnsureConfigForInputStreams(cfg, streams, "relative format arguments")
		if err != nil {
			return err
		}
	}

	summaries := jc.aggregateAndFilterJobs(cfg, theDb, streams, bounds)
	if jc.Verbose {
		Log.Infof("Jobs after aggregation filtering: %d", len(summaries))
	}

	return jc.printJobSummaries(out, summaries)
}

// Container for computations we would prefer not to do but will need to do if certain names are
// used for printing or in queries.

type nameTester struct {
	needCmd        bool
	needHosts      bool
	needJobAndMark bool
	needSacctInfo  bool
}

func (nt *nameTester) testName(name string) {
	switch name {
	case "cmd", "Cmd":
		nt.needCmd = true
	case "host", "hosts", "Hosts":
		nt.needHosts = true
	case "jobm", "JobAndMark":
		nt.needJobAndMark = true
	case "Submit", "JobName", "State", "Account", "Layout", "Reservation",
		"Partition", "RequestedGpus", "DiskReadAvgGB", "DiskWriteAvgGB",
		"RequestedCpus", "RequestedMemGB", "RequestedNodes", "TimeLimit",
		"ExitCode":
		// Our names for the Slurm sacct data fields.  Mostly these are the same as in the sacct
		// data, but there's no shame in sticking to proper naming.
		nt.needSacctInfo = true
	}
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
	theDb db.SampleCluster,
	streams sonarlog.InputStreamSet,
	bounds sonarlog.Timebounds,
) []*jobSummary {
	var now = time.Now().UTC().Unix()
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

	summaryFilter, slurmFilter := jc.buildFilters(cfg)

	summaries := make([]*jobSummary, 0)
	minSamples := jc.lookupUint("min-samples")
	if jc.Verbose && minSamples > 1 {
		Log.Infof("Excluding jobs with fewer than %d samples", minSamples)
	}
	nt := nameTester{
		needSacctInfo: slurmFilter != nil,
	}
	for _, f := range jc.PrintFields {
		nt.testName(f.Name)
	}
	if jc.ParsedQuery != nil {
		names := make(map[string]bool)
		QueryNames(jc.ParsedQuery, names)
		for name := range names {
			nt.testName(name)
		}
	}
	discarded := 0
	for _, job := range jobs {
		if uint(len(*job)) >= minSamples {
			host := (*job)[0].Hostname
			jobId := (*job)[0].Job
			user := (*job)[0].User
			first := (*job)[0].Timestamp
			last := (*job)[len(*job)-1].Timestamp
			duration := last - first
			aggregate := jc.aggregateJob(cfg, host, *job, nt.needCmd, nt.needHosts, jc.Zombie)
			aggregate.computed[kDuration] = float64(duration)
			usesGpu := !aggregate.Gpus.IsEmpty()
			flags := 0
			if usesGpu {
				flags |= kUsesGpu
			} else {
				flags |= kDoesNotUseGpu
			}
			if aggregate.GpuFail != 0 {
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
			if aggregate.IsZombie {
				flags |= kIsZombie
			}
			jobAndMark := ""
			if nt.needJobAndMark {
				mark := ""
				switch {
				case flags&(kIsLiveAtStart|kIsLiveAtEnd) == (kIsLiveAtStart | kIsLiveAtEnd):
					mark = "!"
				case flags&kIsLiveAtStart != 0:
					mark = "<"
				case flags&kIsLiveAtEnd != 0:
					mark = ">"
				}
				jobAndMark = fmt.Sprint(jobId, mark)
			}
			classification := 0
			if (flags & kIsLiveAtStart) != 0 {
				classification |= sonalyze.LIVE_AT_START
			}
			if (flags & kIsLiveAtEnd) != 0 {
				classification |= sonalyze.LIVE_AT_END
			}
			summary := &jobSummary{
				jobAggregate:   aggregate,
				JobId:          jobId,
				JobAndMark:     jobAndMark,
				User:           user,
				CpuTime:        DurationValue(math.Round(aggregate.computed[kCpuPctAvg] * float64(duration) / 100)),
				GpuTime:        DurationValue(math.Round(aggregate.computed[kGpuPctAvg] * float64(kDuration) / 100)),
				Duration:       DurationValue(duration),
				Now:            DateTimeValue(now),
				Start:          DateTimeValue(first),
				End:            DateTimeValue(last),
				selected:       true,
				Classification: classification,
				job:            *job,
				computedFlags:  flags,
			}
			if summaryFilter == nil || summaryFilter.apply(summary) {
				summaries = append(summaries, summary)
			}
		} else {
			discarded++
		}
	}
	if jc.Verbose {
		Log.Infof("Jobs discarded by aggregation filtering: %d", discarded)
	}

	if nt.needSacctInfo {
		// TODO: If we have slurm data then those data may have precise measurements for some of the
		// fields here and we might use them instead.  If so, do so here and not in printing, to
		// avoid messiness vis-a-vis filtering.

		if slurmDb, ok := theDb.(*db.PersistentCluster); ok {

			var err error

			// Two things happen here:
			//
			// - attach slurm info to summaries we have
			// - reduce the set of summaries we have by filtering on slurm information for those
			//   summaries that do have slurm information
			//
			// Importantly, the first step cannot incorporate the second step, because it is valid
			// for a job in the first set to not have a slurm aspect.
			//
			// So:
			//
			// - compute a set A of SlurmJobs from the job IDs alone
			// - then another smaller set B of SlurmJobs from A with the other filters
			// - then A \ B is the set of jobs to remove from the list of summaries
			// - and B is the set of jobs contributing info for the remaining jobs

			jobIds := make([]uint32, 0)
			for _, summary := range summaries {
				if summary.JobId != 0 {
					jobIds = append(jobIds, summary.JobId)
				}
			}

			var (
				aJobs, bJobs []*slurmlog.SlurmJob
				bMap         map[uint32]*slurmlog.SlurmJob
			)
			aJobs, err = slurmlog.Query(
				slurmDb,
				jc.FromDate,
				jc.ToDate,
				slurmlog.QueryFilter{
					Job: jobIds,
				},
				jc.Verbose,
			)
			if err != nil {
				if jc.Verbose {
					Log.Warningf("Slurm data query failed: %v", err)
				}
				// Oh well
				return summaries
			}

			if slurmFilter != nil {
				var err error
				bJobs, err = slurmlog.FilterJobs(
					aJobs,
					*slurmFilter,
					jc.Verbose,
				)
				if err != nil {
					if jc.Verbose {
						Log.Warningf("Slurm data filter failed (bizarrely): %v", err)
					}
					bJobs = aJobs
					// Ignore it, fall through to attach job info
				} else {
					bMap = make(map[uint32]*slurmlog.SlurmJob)
					for _, j := range bJobs {
						bMap[j.Id] = j
					}
					cullSet := make(map[uint32]bool)
					for _, a := range aJobs {
						if bMap[a.Id] == nil {
							cullSet[a.Id] = true
						}
					}
					summaries = slices.DeleteFunc(summaries, func(s *jobSummary) bool {
						return cullSet[s.JobId]
					})
				}
			} else {
				bJobs = aJobs
			}

			if bMap == nil {
				bMap = make(map[uint32]*slurmlog.SlurmJob)
				for _, j := range bJobs {
					bMap[j.Id] = j
				}
			}

			for _, summary := range summaries {
				if probe, found := bMap[summary.JobId]; found {
					summary.sacctInfo = probe.Main // Hm
				}
			}
		} else {
			if jc.Verbose {
				Log.Warningf("Needed slurm data but can't read those from transient cluster")
			}
		}
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
// timestamps, return a JobAggregate for the job, with values that are computed from all log
// entries.

func (jc *JobsCommand) aggregateJob(
	cfg *config.ClusterConfig,
	host Ustr,
	job sonarlog.SampleStream,
	needCmd, needHosts, needZombie bool,
) jobAggregate {
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

	cmd := ""
	if needCmd {
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
	}

	var hosts *Hostnames
	if needHosts {
		hosts = NewHostnames()
		for _, s := range job {
			hosts.Add(s.Hostname.String())
		}
	}
	n := float64(len(job))
	a := jobAggregate{
		Gpus:     gpus,
		GpuFail:  int(gpuFail),
		Cmd:      cmd,
		Hosts:    hosts,
		IsZombie: isZombie,
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
		if s.computed[v.ix] < v.limit {
			return false
		}
	}
	for _, v := range f.maxFilters {
		if s.computed[v.ix] > v.limit {
			return false
		}
	}
	return (f.flags & s.computedFlags) == f.flags
}

func (jc *JobsCommand) buildFilters(
	cfg *config.ClusterConfig,
) (*aggregationFilter, *slurmlog.QueryFilter) {
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

	var summaryFilter *aggregationFilter
	var slurmFilter *slurmlog.QueryFilter

	if len(minFilters) > 0 || len(maxFilters) > 0 || flags != 0 {
		summaryFilter = &aggregationFilter{
			minFilters,
			maxFilters,
			flags,
		}
	}

	if len(jc.Partition)+len(jc.Reservation)+len(jc.Account)+len(jc.State)+len(jc.GpuType) > 0 {
		slurmFilter = &slurmlog.QueryFilter{
			Account:     jc.Account,
			Partition:   jc.Partition,
			Reservation: jc.Reservation,
			GpuType:     jc.GpuType,
			State:       jc.State,
		}
	}

	return summaryFilter, slurmFilter
}
