package jobs

import (
	"fmt"
	"io"
	"math"
	"slices"
	"strings"
	"time"

	"go-utils/gpuset"
	"go-utils/sonalyze"

	. "sonalyze/cmd"
	. "sonalyze/common"
	"sonalyze/data/common"
	"sonalyze/data/config"
	"sonalyze/data/sample"
	"sonalyze/data/slurmjob"
	"sonalyze/db/repr"
	"sonalyze/db/types"
	. "sonalyze/table"
)

// Computed float64 fields in JobAggregate.f64.  These are for the job as a whole, being
// computed from the single stream that is the synthesized / merged job.
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
	KThreadAvg             // Average number of active threads (summed across all processes)
	KThreadPeak            // Peak number of active threads (ditto)
	numF64Fields
)

// Both GB and KB as we use both sometimes.
const (
	UReadGBTotal = iota
	UReadKBTotal
	UWrittenGBTotal
	UWrittenKBTotal
	UResidentKBPeak
	UVirtualKBPeak
	UCpuTimeSecTotal
	UDurationSec
	numU64Fields
)

// Computed flag bits in JobAggregate.computedFlags
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

const kb2gb = 1.0 / (1024 * 1024)

// Package for results from aggregation and summation.
type JobSummary struct {
	JobAggregate
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
	job            sample.MergedJob
	ComputedFlags  int
	selected       bool // Initially true, used to deselect the record before printing
	SacctInfo      *repr.SacctInfo
}

// Aggregate figures for a job.  For some cross-job data like user and host, go to the sample stream
// in the JobSummary that owns this aggregate.
//
// The float fields of this are *not* rounded in any way.
//
// GPU memory: If a system config is present and conf.GpuMemPct is true then KGpuGB* are derived
// from the recorded percentage figure, otherwise KRgpuGB* are derived from the recorded absolute
// figures.  If a system config is not present then all fields will represent the recorded values
// (KRgpuKB * the recorded percentages).
type JobAggregate struct {
	GpuFail     int
	Gpus        gpuset.GpuSet
	Computed    [numF64Fields]float64
	U64         [numU64Fields]uint64
	IsZombie    bool
	InContainer bool
	Cmd         string
	Hosts       *Hostnames
}

type QueryFilter = common.QueryFilter

func Query(meta types.Context, f QueryFilter, parsedQuery PNode) ([]*JobSummary, error) {
	rf := &sample.SampleFilter{To: math.MaxInt64}
	return query(meta, f, parsedQuery, rf)
}

func query(
	meta types.Context,
	filter QueryFilter,
	parsedQuery PNode,
 	recordFilter *sample.SampleFilter,
) ([]*JobSummary, error) {
	sdp, err := sample.OpenSampleDataProvider(meta)
	if err != nil {
		return nil, err
	}
	// TODO: Should just accept the filter!!
	streams, bounds, read, dropped, err :=
		sdp.Query(
			filter.FromDate,
			filter.ToDate,
			hosts,
			true,
		)
	if err != nil {
		return nil, fmt.Errorf("Failed to read log records: %v", err)
	}
	if Verbose {
		Log.Infof("%d records read + %d dropped\n", read, dropped)
		UstrStats(out, false)
	}
	if Verbose {
		Log.Infof("Streams constructed by postprocessing: %d", len(streams))
		numSamples := 0
		for _, stream := range streams {
			numSamples += len(*stream)
		}
		Log.Infof("Samples retained after filtering: %d", numSamples)
	}

	cfg := config.MaybeOpenConfigDataProvider(meta)

	if NeedsConfig(jobsFormatters, jc.PrintFields) {
		var err error
		streams, err = EnsureConfigForInputStreams(cfg, streams, "relative format arguments")
		if err != nil {
			return nil, err
		}
	}

	summaries := jc.summarizeAndFilterJobs(meta, cfg, streams, bounds)
	if Verbose {
		Log.Infof("Jobs after aggregation filtering: %d", len(summaries))
	}

	return ApplyQuery(jc.ParsedQuery, jobsFormatters, jobsPredicates, summaries)
}

func (jc *JobsCommand) Perform(
	out io.Writer,
	meta types.Context,
	filter sample.QueryFilter,	// Why not our own QueryFilter?
	hosts *Hosts,				// Why not in the QueryFilter??!?!
	recordFilter *sample.SampleFilter,
) error {
	summaries, err = query(meta, filter.QueryFilter, jc.ParsedQuery, recordFilter)
	if err != nil {
		return err
	}
	return jc.printJobSummaries(out, summaries)
}

// A sample stream is a quadruple (host, command, job-related-id, record-list).  A stream is only
// ever about one job.  There may be multiple streams per job, they will all have the same
// job-related-id which is unique but not necessarily equal to any field in any of the records.
//
// This function collects the data per job and returns a vector of (summary, records) pairs where
// the summary describes the job in aggregate and the records is a synthesized stream of sample
// records for the job, based on all the input streams for the job.  The manner of the synthesis
// depends on arguments to the program: with --merge-all we merge across all hosts; with
// --merge-none we do not merge; otherwise the config file can specify the hosts to merge across;
// otherwise if there is no config we do not merge.

func (jc *JobsCommand) summarizeAndFilterJobs(
	meta types.Context,
	cfg *config.ConfigDataProvider,
	streams sample.InputStreamSet,
	bounds Timebounds,
) []*JobSummary {
	var jobs sample.MergedJobs
	if jc.MergeAll {
		jobs, bounds = sample.MergeByJob(streams, bounds)
	} else if !jc.MergeNone {
		jobs, bounds = mergeAcrossSomeNodes(streams, bounds)
	} else {
		jobs = sample.MergeByHostAndJob(streams)
	}
	if Verbose {
		Log.Infof("Jobs constructed by merging: %d", len(jobs))
	}
	summaryFilter, slurmFilter := jc.buildFilters()
	fb := flagBag{
		needSacctInfo: !jc.SacctFromSonar && slurmFilter != nil,
		needCmd:       jc.SacctFromSonar,
		needHosts:     jc.SacctFromSonar,
		needZombie:    jc.Zombie,
	}
	summaries, discarded, fb :=
		jc.summarizeJobsFromSonarData(cfg, bounds, jobs, summaryFilter, fb)
	if Verbose {
		Log.Infof("Jobs discarded by aggregation filtering: %d", discarded)
	}
	var slurmDiscarded int
	if jc.SacctFromSonar {
		slurmDiscarded = synthesizeSacctDataFromSonarData(summaries, slurmFilter)
	}
	if fb.needSacctInfo {
		slurmDiscarded = jc.joinSacctData(meta, summaries, slurmFilter)
	}
	if (jc.SacctFromSonar || fb.needSacctInfo) && Verbose {
		Log.Infof("Jobs discarded by aggregation filtering: %d", slurmDiscarded)
	}
	return summaries
}

func (jc *JobsCommand) summarizeJobsFromSonarData(
	cfg *config.ConfigDataProvider,
	bounds Timebounds,
	jobs sample.MergedJobs,
	summaryFilter *aggregationFilter,
	fb flagBag,
) ([]*JobSummary, int, flagBag) {
	var now = time.Now().UTC().Unix()
	summaries := make([]*JobSummary, 0)
	minSamples := jc.lookupUint("min-samples")
	if Verbose && minSamples > 1 {
		Log.Infof("Excluding jobs with fewer than %d samples", minSamples)
	}
	if !jc.SacctFromSonar {
		for _, f := range jc.PrintFields {
			fb.setFromFieldName(f.Name)
		}
	}
	if jc.ParsedQuery != nil {
		names := make(map[string]bool)
		QueryNames(jc.ParsedQuery, names)
		for name := range names {
			fb.setFromFieldName(name)
		}
	}

	discarded := 0
	for _, job := range jobs {
		if uint(len(job.Samples)) >= minSamples {
			summary := summarizeSingleJobFromSonarData(cfg, bounds, job, now, fb)
			if summaryFilter == nil || summaryFilter.apply(summary) {
				summaries = append(summaries, summary)
			} else {
				discarded++
			}
		} else {
			discarded++
		}
	}
	return summaries, discarded, fb
}

// Aggregate and summarize but do not attach any sacct data.
func summarizeSingleJobFromSonarData(
	cfg *config.ConfigDataProvider,
	bounds Timebounds,
	job sample.MergedJob,
	now int64,
	fb flagBag,
) *JobSummary {
	samples := job.Samples
	host := samples[0].Hostname
	jobId := samples[0].Job
	user := samples[0].User
	first := samples[0].Timestamp
	last := samples[len(samples)-1].Timestamp
	duration := last - first
	aggregate := aggregateSingleJobFromSonarData(cfg, host, samples, fb)
	aggregate.U64[UDurationSec] = uint64(duration)
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
	jobAndMark := ""
	if fb.needJobAndMark {
		mark := ""
		switch {
		case flags&(KIsLiveAtStart|KIsLiveAtEnd) == (KIsLiveAtStart | KIsLiveAtEnd):
			mark = "!"
		case flags&KIsLiveAtStart != 0:
			mark = "<"
		case flags&KIsLiveAtEnd != 0:
			mark = ">"
		}
		jobAndMark = fmt.Sprint(jobId, mark)
	}
	classification := 0
	if (flags & KIsLiveAtStart) != 0 {
		classification |= sonalyze.LIVE_AT_START
	}
	if (flags & KIsLiveAtEnd) != 0 {
		classification |= sonalyze.LIVE_AT_END
	}
	return &JobSummary{
		JobAggregate:   aggregate,
		JobId:          jobId,
		JobAndMark:     jobAndMark,
		User:           user,
		CpuTime:        DurationValue(aggregate.U64[UCpuTimeSecTotal]),
		GpuTime:        DurationValue(math.Round(aggregate.Computed[KGpuPctAvg] * float64(duration) / 100)),
		Duration:       DurationValue(duration),
		Now:            DateTimeValue(now),
		Start:          DateTimeValue(first),
		End:            DateTimeValue(last),
		selected:       true,
		Classification: classification,
		job:            job,
		ComputedFlags:  flags,
		// SacctInfo is attached later, if it is needed
	}
}

// Given a list of log entries for a job - a single stream of samples where each sample is the merge
// across all processes for the job at some time point - sorted ascending by timestamp and with no
// duplicated timestamps, return a JobAggregate for the job, with values that are computed from all
// log entries.
func aggregateSingleJobFromSonarData(
	cfg *config.ConfigDataProvider,
	host Ustr,
	job []sample.Sample,
	fb flagBag,
) JobAggregate {
	gpus := gpuset.EmptyGpuSet()
	var (
		gpuFail                          uint8
		cpuPctAvg, cpuPctPeak            float64
		rCpuPctAvg, rCpuPctPeak          float64
		cpuGBAvg, cpuGBPeak              float64
		rCpuGBAvg, rCpuGBPeak            float64
		gpuPctAvg, gpuPctPeak            float64
		rGpuPctAvg, rGpuPctPeak          float64
		sGpuPctAvg, sGpuPctPeak          float64
		rssAnonGBAvg, rssAnonGBPeak      float64
		rRssAnonGBAvg, rRssAnonGBPeak    float64
		gpuGBAvg, gpuGBPeak              float64
		rGpuGBAvg, rGpuGBPeak            float64
		sGpuGBAvg, sGpuGBPeak            float64
		cpuTime                          uint64
		threadAvg, threadPeak            uint32
		isZombie, inContainer            bool
		dataReadGB, dataWrittenGB        uint64
		dataReadKB, dataWrittenKB        uint64
		vmSizeKBPeak, residentSizeKBPeak uint64
	)

	for _, s := range job {
		gpus = gpuset.UnionGpuSets(gpus, s.Gpus)
		gpuFail = sample.MergeGpuFail(gpuFail, s.GpuFail)
		cpuPctAvg += float64(s.CpuUtilPct)
		cpuPctPeak = max(cpuPctPeak, float64(s.CpuUtilPct))
		gpuPctAvg += float64(s.GpuPct)
		gpuPctPeak = max(gpuPctPeak, float64(s.GpuPct))
		cpuTime += s.CpuTimeSec
		cpuGBAvg += float64(s.CpuKB) * kb2gb
		cpuGBPeak = max(cpuGBPeak, float64(s.CpuKB)*kb2gb)
		rssAnonGBAvg += float64(s.RssAnonKB) * kb2gb
		rssAnonGBPeak = max(rssAnonGBPeak, float64(s.RssAnonKB)*kb2gb)
		gpuGBAvg += float64(s.GpuKB) * kb2gb
		gpuGBPeak = max(gpuGBPeak, float64(s.GpuKB)*kb2gb)
		threadAvg += s.NumThreads
		threadPeak = max(threadPeak, s.NumThreads)
		inContainer = inContainer || s.InContainer
		// ignore CpuSampledUtilPct for now
		dataReadKB += s.DataReadKB
		dataReadGB += s.DataReadKB / (1024 * 1024)
		dataWrittenKB += s.DataWrittenKB
		dataWrittenKB += s.DataWrittenKB / (1024 * 1024)
		residentSizeKBPeak = max(residentSizeKBPeak, s.RssAnonKB)
		vmSizeKBPeak = max(vmSizeKBPeak, s.CpuKB)

		if fb.needZombie && !isZombie {
			cmd := s.Cmd.String()
			isZombie = strings.Contains(cmd, "<defunct>") || strings.HasPrefix(cmd, "_zombie_")
		}
	}
	usesGpu := !gpus.IsEmpty()

	if sys := cfg.LookupHostByTime(host, job[0].Timestamp); sys != nil {
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

	cmd := ""
	if fb.needCmd {
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
	if fb.needHosts {
		hosts = NewHostnames()
		for _, s := range job {
			hosts.Add(s.Hostname.String())
		}
	}
	n := float64(len(job))
	a := JobAggregate{
		Gpus:        gpus,
		GpuFail:     int(gpuFail),
		Cmd:         cmd,
		Hosts:       hosts,
		IsZombie:    isZombie,
		InContainer: inContainer,
	}
	a.Computed[KCpuPctAvg] = cpuPctAvg / n
	a.Computed[KCpuPctPeak] = cpuPctPeak
	a.Computed[KRcpuPctAvg] = rCpuPctAvg / n
	a.Computed[KRcpuPctPeak] = rCpuPctPeak

	a.Computed[KCpuGBAvg] = cpuGBAvg / n
	a.Computed[KCpuGBPeak] = cpuGBPeak
	a.Computed[KRcpuGBAvg] = rCpuGBAvg / n
	a.Computed[KRcpuGBPeak] = rCpuGBPeak

	a.Computed[KRssAnonGBAvg] = rssAnonGBAvg / n
	a.Computed[KRssAnonGBPeak] = rssAnonGBPeak
	a.Computed[KRrssAnonGBAvg] = rRssAnonGBAvg / n
	a.Computed[KRrssAnonGBPeak] = rRssAnonGBPeak

	a.Computed[KGpuPctAvg] = gpuPctAvg / n
	a.Computed[KGpuPctPeak] = gpuPctPeak
	a.Computed[KRgpuPctAvg] = rGpuPctAvg / n
	a.Computed[KRgpuPctPeak] = rGpuPctPeak
	a.Computed[KSgpuPctAvg] = sGpuPctAvg / n
	a.Computed[KSgpuPctPeak] = sGpuPctPeak

	a.Computed[KGpuGBAvg] = gpuGBAvg / n
	a.Computed[KGpuGBPeak] = gpuGBPeak
	a.Computed[KRgpuGBAvg] = rGpuGBAvg / n
	a.Computed[KRgpuGBPeak] = rGpuGBPeak
	a.Computed[KSgpuGBAvg] = sGpuGBAvg / n
	a.Computed[KSgpuGBPeak] = sGpuGBPeak

	a.Computed[KThreadAvg] = float64(threadAvg) / n
	a.Computed[KThreadPeak] = float64(threadPeak)

	a.U64[UReadGBTotal] = dataReadGB
	a.U64[UReadKBTotal] = dataReadKB
	a.U64[UWrittenGBTotal] = dataWrittenGB
	a.U64[UWrittenKBTotal] = dataWrittenKB
	a.U64[UCpuTimeSecTotal] = cpuTime
	a.U64[UResidentKBPeak] = residentSizeKBPeak
	a.U64[UVirtualKBPeak] = vmSizeKBPeak

	return a
}

// The synthesis is imperfect, and would be so even if the Slurm documentation were better.
func synthesizeSacctDataFromSonarData(
	summaries []*JobSummary,
	slurmFilter *slurmjob.QueryFilter,
) int {
	// TODO: This needs to apply the slurmFilter if it is defined.
	var discarded int
	for _, s := range summaries {
		var state Ustr
		if (s.ComputedFlags & KIsNotLiveAtEnd) != 0 {
			state = StringToUstr("COMPLETED")
		} else {
			state = StringToUstr("RUNNING")
		}

		var requestedGpus string
		if s.Gpus.IsUnknown() {
			requestedGpus = "*=1"
		} else if s.Gpus.Size() > 0 {
			requestedGpus = fmt.Sprintf("*=%d", s.Gpus.Size())
		}

		var (
			minCpu        = uint64(math.MaxUint64)
			maxRSS, maxVM uint64
			aveRSS, aveVM uint64
		)
		for _, t := range s.job.Tasks {
			minCpu = min(minCpu, t[len(t)-1].CpuTimeSec)
			var sumVM, sumRSS uint64
			for _, sample := range t {
				sumRSS += sample.RssAnonKB
				sumVM += sample.CpuKB
				maxVM = max(maxVM, sample.CpuKB)
				maxRSS = max(maxRSS, sample.RssAnonKB)
			}
			aveVM += sumVM / uint64(len(t))
			aveRSS += sumRSS / uint64(len(t))
		}
		if minCpu == math.MaxUint64 {
			minCpu = 0
		}
		aveVM /= uint64(len(s.job.Tasks))
		aveRSS /= uint64(len(s.job.Tasks))
		s.SacctInfo = &repr.SacctInfo{
			Account:      s.User,
			AveCPU:       s.U64[UCpuTimeSecTotal] / uint64(s.job.NumTasks),
			AveDiskRead:  s.U64[UReadKBTotal] / uint64(s.job.NumTasks),
			AveDiskWrite: s.U64[UWrittenKBTotal] / uint64(s.job.NumTasks),
			AveRSS:       aveRSS,
			AveVMSize:    aveVM,
			ElapsedRaw:   uint32(s.U64[UDurationSec]),
			End:          s.End,
			JobID:        s.JobId,
			JobName:      StringToUstr(s.User.String() + ": " + s.Cmd),
			MaxRSS:       maxRSS,
			MaxVMSize:    maxVM,
			MinCPU:       minCpu,
			NodeList:     StringToUstr(FormatHostnames(s.Hosts, PrintModFixed)),
			ReqCPUS:      uint32(s.Computed[KThreadPeak]), // Requested = peak threads observed, not great
			ReqGPUS:      StringToUstr(requestedGpus),
			ReqRes:       StringToUstr(requestedGpus),
			ReqMem:       maxVM, // Requested = max of any task at any time
			Start:        s.Start,
			State:        state,
			Submit:       s.Start,
			Time:         s.job.Samples[len(s.job.Samples)-1].Timestamp, // Last synthesized time
			UserCPU:      uint64(s.CpuTime),
			User:         s.User,
			Version:      s.job.Samples[0].Version, // For synthesized data, always 0.0.0

			// We don't have these:
			// AllocRes
			// ArrayJobID
			// ArrayStep
			// ArrayTaskID
			// ExitCode
			// HetJobID
			// HetJobOffset
			// HetStep
			// JobStep
			// Layout
			// Partition
			// Priority
			// ReqNodes
			// Reservation
			// Suspended
			// SystemCPU
			// TimeLimitRaw
		}
	}
	return discarded
}

func (jc *JobsCommand) joinSacctData(
	meta types.Context,
	summaries []*JobSummary,
	slurmFilter *slurmjob.QueryFilter,
) int {
	// TODO: If we have slurm data then those data may have precise measurements for some of the
	// fields here and we might use them instead.  If so, do so here and not in printing, to
	// avoid messiness vis-a-vis filtering.

	var discarded int
	if sdp, err := slurmjob.OpenSlurmjobDataProvider(meta); err == nil {
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
			aJobs, bJobs []*slurmjob.SlurmJob
			bMap         map[uint32]*slurmjob.SlurmJob
		)
		aJobs, err = sdp.Query(
			slurmjob.QueryFilter{
				QueryFilter: common.QueryFilter{
					HaveFrom: jc.HaveFrom,
					FromDate: jc.FromDate,
					HaveTo:   jc.HaveTo,
					ToDate:   jc.ToDate,
				},
				Job: jobIds,
			},
		)
		if err != nil {
			if Verbose {
				Log.Warningf("Slurm data query failed: %v", err)
			}
			// Oh well
			return discarded
		}

		if slurmFilter != nil {
			var err error
			bJobs, err = slurmjob.FilterJobs(
				aJobs,
				*slurmFilter,
			)
			if err != nil {
				if Verbose {
					Log.Warningf("Slurm data filter failed (bizarrely): %v", err)
				}
				bJobs = aJobs
				// Ignore it, fall through to attach job info
			} else {
				bMap = make(map[uint32]*slurmjob.SlurmJob)
				for _, j := range bJobs {
					bMap[j.Id] = j
				}
				cullSet := make(map[uint32]bool)
				for _, a := range aJobs {
					if bMap[a.Id] == nil {
						cullSet[a.Id] = true
					}
				}
				summaries = slices.DeleteFunc(summaries, func(s *JobSummary) bool {
					return cullSet[s.JobId]
				})
			}
		} else {
			bJobs = aJobs
		}

		if bMap == nil {
			bMap = make(map[uint32]*slurmjob.SlurmJob)
			for _, j := range bJobs {
				bMap[j.Id] = j
			}
		}

		for _, summary := range summaries {
			if probe, found := bMap[summary.JobId]; found {
				summary.SacctInfo = probe.Main // Hm
			}
		}
	} else {
		if Verbose {
			Log.Warningf("Needed slurm data but can't read those from transient cluster")
		}
	}
	return discarded
}

// Merge mergeable streams as if by --merge-all; the remaining streams are merged as if by
// --merge-none, and the two sets of merged jobs are combined into one set.

func mergeAcrossSomeNodes(
	streams sample.InputStreamSet,
	bounds Timebounds,
) (sample.MergedJobs, Timebounds) {
	mergeable := make(sample.InputStreamSet)
	mBounds := make(Timebounds)
	solo := make(sample.InputStreamSet)
	sBounds := make(Timebounds)
	for k, v := range streams {
		bound := bounds[k.Host]
		if (*v)[0].Epoch == 0 {
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

// Container for computations we would prefer not to do but will need to do if certain names are
// used for printing or in queries.

type flagBag struct {
	needCmd        bool
	needHosts      bool
	needJobAndMark bool
	needSacctInfo  bool
	needZombie     bool
}

func (nt *flagBag) setFromFieldName(name string) {
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

// Aggregation filters.
//
// Filtering is mostly wasted work.  Very frequently, all the filters will pass because the coarse
// filtering (job number, user, command, host) has been applied already and most of the filters
// applied to the aggregate are not very interesting to many users and will not be used to reject
// many records.
//
// There are several ways to represent a filter.  The simplest is just as a table of values
// representing min and max values.

type ffilterVal struct {
	limit float64
	ix    int
}

type ufilterVal struct {
	limit uint64
	ix    int
}

type aggregationFilter struct {
	fminFilters []ffilterVal
	uminFilters []ufilterVal
	fmaxFilters []ffilterVal
	flags       int
}

func (f *aggregationFilter) apply(s *JobSummary) bool {
	for _, v := range f.fminFilters {
		if s.Computed[v.ix] < v.limit {
			return false
		}
	}
	for _, v := range f.uminFilters {
		if s.U64[v.ix] < v.limit {
			return false
		}
	}
	for _, v := range f.fmaxFilters {
		if s.Computed[v.ix] > v.limit {
			return false
		}
	}
	return (f.flags & s.ComputedFlags) == f.flags
}

func (jc *JobsCommand) buildFilters() (*aggregationFilter, *slurmjob.QueryFilter) {
	fminFilters := make([]ffilterVal, 0)
	uminFilters := make([]ufilterVal, 0)
	fmaxFilters := make([]ffilterVal, 0)

	for _, v := range uintArgs {
		// There's a general assumption here that we will always have node config data and so we can
		// always handle "relative" fields - that require quantities to be computed relative to the
		// node config - properly.  This is not strictly true: config data for a node can be missing
		// when we're running on a file list, especially.  Buyer beware.
		if v.aggregateIx != -1 {
			val := jc.lookupUint(v.name)
			if strings.HasPrefix(v.name, "min-") && val != 0 {
				if Verbose {
					Log.Infof("Excluding jobs: Min-filtering %s for %d", v.name, val)
				}
				fminFilters = append(fminFilters, ffilterVal{float64(val), v.aggregateIx})
			}
			if strings.HasPrefix(v.name, "max-") && val != v.initial {
				if Verbose {
					Log.Infof("Excluding jobs: Max-filtering %s for %d", v.name, val)
				}
				fmaxFilters = append(fmaxFilters, ffilterVal{float64(val), v.aggregateIx})
			}
		}
	}
	if jc.MinRuntimeSec > 0 {
		// This is *running time*, not CPU time
		if Verbose {
			Log.Infof("Excluding jobs: Min-filtering by elapsed time < %ds", jc.MinRuntimeSec)
		}
		uminFilters = append(uminFilters, ufilterVal{uint64(jc.MinRuntimeSec), UDurationSec})
	}

	// For the flags, set all the conditions we care about.  They must all be set in the summary's
	// computed flags.
	flags := 0
	if jc.NoGpu {
		flags |= KDoesNotUseGpu
	}
	if jc.SomeGpu {
		flags |= KUsesGpu
	}
	if jc.Completed {
		flags |= KIsNotLiveAtEnd
	}
	if jc.Running {
		flags |= KIsLiveAtEnd
	}
	if jc.Zombie {
		flags |= KIsZombie
	}
	if Verbose && flags != 0 {
		Log.Infof("Flag-filtering (UTSL): %x", flags)
	}

	var summaryFilter *aggregationFilter
	var slurmFilter *slurmjob.QueryFilter

	if len(fminFilters) > 0 || len(fmaxFilters) > 0 || len(uminFilters) > 0 || flags != 0 {
		summaryFilter = &aggregationFilter{
			fminFilters,
			uminFilters,
			fmaxFilters,
			flags,
		}
	}

	if len(jc.Partition)+len(jc.Reservation)+len(jc.Account)+len(jc.State)+len(jc.GpuType) > 0 {
		slurmFilter = &slurmjob.QueryFilter{
			Account:     jc.Account,
			Partition:   jc.Partition,
			Reservation: jc.Reservation,
			GpuType:     jc.GpuType,
			State:       jc.State,
		}
	}

	return summaryFilter, slurmFilter
}
