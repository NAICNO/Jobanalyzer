package samplejob

import (
	. "sonalyze/common"
	"sonalyze/data/sample"
)

// Computed float64 fields in SampleJob.Computed
const (
	KCpuPctAvg      = iota // Average CPU utilization, 1 core == 100%
	KCpuPctPeak            // Peak CPU utilization ditto
	KRcpuPctAvg            // Average CPU utilization, all cores == 100%
	KRcpuPctPeak           // Peak CPU utilization ditto
	KCpuGBAvg              // Average main memory utilization, GiB
	KCpuGBPeak             // Peak memory utilization ditto
	KRcpuGBAvg             // Average main memory utilization, all memory = 100%
	KRcpuGBPeak            // Peak memory utilization ditto
	KRssAnonGBAvg          // Average resident main memory utilization, GiB
	KRssAnonGBPeak         // Peak memory utilization ditto
	KRrssAnonGBAvg         // Average resident main memory utilization, all memory = 100%
	KRrssAnonGBPeak        // Peak memory utilization ditto
	KGpuPctAvg             // Average GPU utilization, 1 card == 100%
	KGpuPctPeak            // Peak GPU utilization ditto
	KRgpuPctAvg            // Average GPU utilization, all cards == 100%
	KRgpuPctPeak           // Peak GPU utilization ditto
	KGpuGBAvg              // Average GPU memory utilization, GiB
	KGpuGBPeak             // Peak memory utilization ditto
	KRgpuGBAvg             // Average GPU memory utilization, all cards == 100%
	KRgpuGBPeak            // Peak GPU memory utilization ditto
	KSgpuPctAvg            // Average GPU utilization, all cards used by job == 100%
	KSgpuPctPeak           // Peak GPU utilization, all cards used by job == 100%
	KSgpuGBAvg             // Average GPU memory utilization, all cards used by job == 100%
	KSgpuGBPeak            // Peak GPU memory utilization ditto
	KDuration              // Duration of job in seconds (wall clock, not CPU)
	NumF64Fields
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

// A "SampleJob" is a job constructed entirely from sample records coming from one or more sample
// streams belonging to the same job.
//
// TODO: Here we compute it from raw data afresh every time, but once a job has completed the values
// do not change (for a given record filter), and the results could be stored in the database, esp
// if the sample stream is not needed.
//
// Aggregate figures for a job:  For some cross-job data like user and host, go to the sample stream.
//
// The computed float fields of this are *not* rounded in any way.
//
// GPU memory fields: If a system config is present and conf.GpuMemPct is true then kGpuGB* are
// derived from the recorded percentage figure, otherwise kRgpuGB* are derived from the recorded
// absolute figures.  If a system config is not present then all fields will represent the recorded
// values (kRgpuKB * the recorded percentages).
type SampleJob struct {
	// Nonzero if any gpu had a failure
	GpuFail        int
	// Gpus involved in the job
	Gpus           gpuset.GpuSet
	// Aggregated fields, see enumeration earlier
	Computed       [NumF64Fields]float64
	// True if the job is believed to be a zombie
	IsZombie       bool
	// Combined command line
	Cmd            string
	// Combined host set
	Hosts          *Hostnames
	// Job's primary ID
	JobId          uint32
	// User running the job
	User           Ustr
	// Real time taken by the job
	Duration       DurationValue
	// Earliest time seen for the job
	Start          DateTimeValue
	// Latest time ditto
	End            DateTimeValue
	// The list of sample records, may be shared with the database layer
	Job            sample.SampleStream
	//
	CpuTime        DurationValue
	//
	GpuTime        DurationValue
	// Bit vector of flags
	Classification int
	//
	ComputedFlags  int
}

type QueryFilter = sample.QueryFilter

type NeededComputations struct {
	NeedCmd        bool
	NeedHosts      bool
	NeedJobAndMark bool
	NeedSacctInfo  bool
}

func Query(
	theLog db.SampleDataProvider,
	cfg *config.ClusterConfig,
	filter QueryFilter,
	needed NeededComputations,
	verbose bool,
) ([]*SampleJob, error) {
) {
	// Basically, read the records as in local.go, then apply aggregation as in jobs/perform.go
	panic("NYI")
}

func (jc *JobsCommand) aggregateAndFilterJobs(
	cfg *config.ClusterConfig,
	theDb db.SampleDataProvider,
	streams sample.InputStreamSet,
	bounds Timebounds,
) []*jobSummary {
	var now = time.Now().UTC().Unix()
	var anyMergeableNodes bool
	if !jc.MergeNone && cfg != nil {
		anyMergeableNodes = cfg.HasCrossNodeJobs()
	}

	var jobs sample.SampleStreams
	if jc.MergeAll {
		jobs, bounds = sample.MergeByJob(streams, bounds)
	} else if anyMergeableNodes {
		jobs, bounds = mergeAcrossSomeNodes(cfg, streams, bounds)
	} else {
		jobs = sample.MergeByHostAndJob(streams)
	}
	if jc.Verbose {
		Log.Infof("Jobs constructed by merging: %d", len(jobs))
	}

	summaries := make([]*jobSummary, 0)
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
				// fixme
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
				Classification: classification,
				Job:            *job,
				ComputedFlags:  flags,
			}
		}
	}
}

// Given a list of log entries for a job, sorted ascending by timestamp and with no duplicated
// timestamps, return a JobAggregate for the job, with values that are computed from all log
// entries.

func computeAggregate(
	cfg *config.ClusterConfig,
	host Ustr,
	job sample.SampleStream,
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
		gpuFail = sample.MergeGpuFail(gpuFail, s.GpuFail)
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
			if usesGpu && !gpus.IsUnknown() {
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

// Look to the config to find nodes that have CrossNodeJobs set, and merge their streams as if by
// --merge-all; the remaining streams are merged as if by --merge-none, and the two sets of merged
// jobs are combined into one set.

func mergeAcrossSomeNodes(
	cfg *config.ClusterConfig,
	streams sample.InputStreamSet,
	bounds Timebounds,
) (sample.SampleStreams, Timebounds) {
	mergeable := make(sample.InputStreamSet)
	mBounds := make(Timebounds)
	solo := make(sample.InputStreamSet)
	sBounds := make(Timebounds)
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
	mergedJobs, mergedBounds := sample.MergeByJob(mergeable, mBounds)
	otherJobs := sample.MergeByHostAndJob(solo)
	mergedJobs = append(mergedJobs, otherJobs...)
	for k, v := range sBounds {
		mergedBounds[k] = v
	}
	return mergedJobs, mergedBounds
}

