package jobs

import (
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
	"time"

	"go-utils/config"
	"go-utils/gpuset"
	"go-utils/sonalyze"

	. "sonalyze/cmd"
	. "sonalyze/common"
	"sonalyze/data/sample"
	"sonalyze/data/samplejob"
	"sonalyze/data/slurmjob"
	"sonalyze/db"
	. "sonalyze/table"
)

type needed struct {
	samplejob.NeededComputations
	sacct      bool
	jobAndMark bool
	sample     bool
}

// TODO: Super brittle!!!  If the print table changes, this must change.
// Would be good to have some kind of auto-generated interlock, or to generate this function,
// or to generate a predicate for slurm names?

func testName(nt *needed, name string) {
	switch name {
	case "cmd", "Cmd":
		nt.Cmd = true
	case "host", "hosts", "Hosts":
		nt.Hosts = true
	case "jobm", "JobAndMark":
		nt.jobAndMark = true
	case "Submit", "JobName", "State", "Account", "Layout", "Reservation",
		"Partition", "RequestedGpus", "DiskReadAvgGB", "DiskWriteAvgGB",
		"RequestedCpus", "RequestedMemGB", "RequestedNodes", "TimeLimit",
		"ExitCode":
		// Our names for the Slurm sacct data fields.  Mostly these are the same as in the sacct
		// data, but there's no shame in sticking to proper naming.
		nt.sacct = true
	default:
		nt.sample = true
	}
}

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

// A number of fields could come *either* from the sample job or the slurm job, depending on what
// data we have, and it's the result of that joining that we want to print when we print jobs.  If
// there is a SlurmJob but not a SampleJob then a SampleJob is synthesized from the slurmJob data.
// This keeps printing logic sane, and there are no samples exposed so that's fine.
type jobSummary struct {
	Now        DateTimeValue
	JobAndMark string
	selected   bool
	sampleJob  *samplejob.SampleJob
	slurmJob   *slurmjob.SlurmJob
	computed   [numF64Fields]float64
}

type JobsDataProvider interface {
	db.ProcessSampleDataProvider
	db.SacctDataProvider
}

func (jc *JobsCommand) Perform(
	out io.Writer,
	cfg *config.ClusterConfig,
	theDb JobsDataProvider,
	filter sample.QueryFilter,
	hosts *Hosts,
	recordFilter *sample.SampleFilter,
) error {
	var need needed
	for _, f := range jc.PrintFields {
		testName(&need, f.Name)
	}
	if jc.ParsedQuery != nil {
		names := make(map[string]bool)
		QueryNames(jc.ParsedQuery, names)
		for name := range names {
			testName(&need, name)
		}
	}

	var needConfig = NeedsConfig(jobsFormatters, jc.PrintFields)
	if needConfig && cfg == nil {
		return fmt.Errorf("Configuration file required for relative format arguments")
	}

	// We want to:
	//
	//  - compute an "OR" of slurm jobs and sample jobs so that if a job ID is in either data set
	//    then the job is in the result
	//  - synthesize SampleJob data for slurm jobs without a corresponding sample job
	//  - compute relative fields for all the jobs in the set, which depends either on the allocation
	//    for the job or on the node's configuration
	//  - filter the resulting summaries
	//
	// This needs to be done in a particular order to work at all.

	// Map from JobId to the summary
	var smap = make(map[uint32]*jobSummary)

	if need.sample {
		sampleJobs, err := jc.findSampleJobs(
			cfg,
			theDb,
			filter,
			hosts,
			recordFilter,
			need.NeededComputations,
		)
		if err != nil {
			return err
		}
		for _, j := range sampleJobs {
			smap[j.JobId] = &jobSummary{sampleJob: j}
		}
	}

	if need.sacct {
		slurmJobs, err := jc.findSlurmJobs(theDb, filter)
		if err != nil {
			return err
		}
		for _, j := range slurmJobs {
			if probe := smap[j.Id]; probe != nil {
				probe.slurmJob = j
			} else {
				smap[j.Id] = &jobSummary{
					sampleJob: jc.synthesizeSampleJob(j),
					slurmJob: j,
				}
			}
		}
	}

	for _, j := range smap {
		jc.computeComputedFields(cfg, j)
	}

	if sampleFilter := jc.buildSampleFilter(cfg != nil); sampleFilter != nil {
		maps.DeleteFunc(smap, func(k uint32, v *jobSummary) bool {
			return !sampleFilter.apply(v)
		})
	}

	// Also TODO: min-samples is a thing, but it got dropped on the floor somewhere.  There is now
	// SampleCount int the SampleJob record.  Be careful when applying the filter to synthesized
	// jobs above, or when synthesizing jobs, since the natural sample count for synthesized jobs is
	// zero.

	var summaries = slices.Collect(maps.Values(smap))

	var now = time.Now().UTC().Unix()
	for i := range summaries {
		summaries[i].Now = now
		if need.jobAndMark {
			mark := ""
			flags := summaries[i].sampleJob.ComputedFlags
			switch {
			case flags&(samplejob.KIsLiveAtStart|samplejob.KIsLiveAtEnd) == (samplejob.KIsLiveAtStart | samplejob.KIsLiveAtEnd):
				mark = "!"
			case flags&samplejob.KIsLiveAtStart != 0:
				mark = "<"
			case flags&samplejob.KIsLiveAtEnd != 0:
				mark = ">"
			}
			summaries[i].JobAndMark = fmt.Sprint(summaries[i].sampleJob.JobId, mark)
		}
	}

	return jc.printJobSummaries(out, summaries)
}

func (jc *JobsCommand) findSampleJobs(
	cfg *config.ClusterConfig,
	theDb JobsDataProvider,
	filter sample.QueryFilter,
	hosts *Hosts,
	recordFilter *sample.SampleFilter,
	need samplejob.NeededComputations,
) ([]*samplejob.SampleJob, error) {
	var merge samplejob.Merge
	switch {
	case jc.MergeAll:
		merge = samplejob.MergeAll
	case jc.MergeNone:
		merge = samplejob.MergeNone
	}

	isMergeable := func(k sample.InputStreamKey) bool {
		// TODO: Eventually we'll need to use the epoch here
		sys := cfg.LookupHost(k.Host.String())
		return sys != nil && sys.CrossNodeJobs
	}

	sampleJobs, err := samplejob.Query(
		theDb,
		isMergeable,
		filter.FromDate,
		filter.ToDate,
		hosts,
		recordFilter,
		need,
		merge,
		jc.Verbose,
	)
	if err != nil {
		return nil, err
	}
	if jc.Verbose {
		Log.Infof("Sample jobs after aggregation filtering: %d", len(sampleJobs))
	}

	return sampleJobs, nil
}

func (jc *JobsCommand) findSlurmJobs(
	theDb JobsDataProvider,
	filter sample.QueryFilter,
) (
	[]*slurmjob.SlurmJob,
	error,
) {
	slurmFilter := jc.buildSlurmFilter()
	if slurmFilter == nil {
		slurmFilter = &slurmjob.QueryFilter{}
	}

	slurmJobs, err := slurmjob.Query(
		theDb,
		filter.FromDate,
		filter.ToDate,
		*slurmFilter,
		jc.Verbose,
	)
	if err != nil {
		if jc.Verbose {
			Log.Warningf("Slurm data query failed: %v", err)
		}
		return nil, err
	}
	if jc.Verbose {
		Log.Infof("Slurm jobs after aggregation filtering: %d", len(slurmJobs))
	}

	return slurmJobs, nil
}

var (
	pending = StringToUstr("PENDING")
	running = StringToUstr("RUNNING")
)

// Synthesize a SampleJob from the SlurmJob to hold common data.
func (jc *JobsCommand) synthesizeSampleJob(j *slurmjob.SlurmJob) *samplejob.SampleJob {
	var gpus gpuset.GpuSet
	// TODO: compute gpus from ReqGPUS
	var hosts *Hostnames = NewHostnames()
	// TODO: compute hosts from NodeList
	var classification int
	if j.Main.State == pending || j.Main.State == running {
		classification |= sonalyze.LIVE_AT_END
	}
	var flags int
	if (classification & sonalyze.LIVE_AT_END) != 0 {
		flags |= samplejob.KIsLiveAtEnd
	} else {
		flags |= samplejob.KIsNotLiveAtEnd
	}
	if !gpus.IsEmpty() {
		flags |= samplejob.KUsesGpu
	} else {
		flags |= samplejob.KDoesNotUseGpu
	}
	var sampleJob = &samplejob.SampleJob{
		// `GpuFail` is not computable
		Gpus: gpus,
		// `Computed` is handled below
		// `IsZombie` is not applicable
		Cmd:      j.Main.JobName.String(),
		Hosts:    hosts,
		JobId:    j.Id,
		User:     j.Main.User,
		Duration: DurationValue(j.Main.ElapsedRaw),
		Start:    DateTimeValue(j.Main.Start),
		End:      DateTimeValue(j.Main.End),
		// `Job` is not applicable
		CpuTime: DurationValue(j.Main.SystemCPU + j.Main.UserCPU),
		// `GpuTime` is not applicable
		Classification: classification,
		ComputedFlags:  flags,
	}
	return sampleJob
}

func (jc *JobsCommand) computeComputedFields(cfg *config.ClusterConfig, js *jobSummary, ) {
	// TODO: Computed fields, at least these:
	js.computed[kDuration] = float64(js.sampleJob.Duration)
	// CpuPctAvg = (SystemCPU + UserCPU) / (End - Start)
	// RcpuPctAvg = ...

/*
   // This is completely ill-defined if hosts were merged because there's no such thing as
   // a merged config.  This is very old code, probably predating multi-node jobs.

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

	a.CpuPctAvg] = cpuPctAvg / n
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

*/

}

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

type sampleFilter struct {
	minFilters []filterVal
	maxFilters []filterVal
	flags      int
}

func (f *sampleFilter) apply(s *jobSummary) bool {
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
	return (f.flags & s.sampleJob.ComputedFlags) == f.flags
}

func (jc *JobsCommand) buildSampleFilter(allowRelative bool) *sampleFilter {
	minFilters := make([]filterVal, 0)
	maxFilters := make([]filterVal, 0)

	for _, v := range uintArgs {
		if v.aggregateIx != -1 && (allowRelative || !v.relative) {
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
		flags |= samplejob.KDoesNotUseGpu
	}
	if jc.SomeGpu {
		flags |= samplejob.KUsesGpu
	}
	if jc.Completed {
		flags |= samplejob.KIsNotLiveAtEnd
	}
	if jc.Running {
		flags |= samplejob.KIsLiveAtEnd
	}
	if jc.Zombie {
		flags |= samplejob.KIsZombie
	}
	if jc.Verbose && flags != 0 {
		Log.Infof("Flag-filtering (UTSL): %x", flags)
	}

	if len(minFilters) > 0 || len(maxFilters) > 0 || flags == 0 {
		return nil
	}

	return &sampleFilter{
		minFilters,
		maxFilters,
		flags,
	}
}

func (jc *JobsCommand) buildSlurmFilter() *slurmjob.QueryFilter {
	if len(jc.Partition)+len(jc.Reservation)+len(jc.Account)+len(jc.State)+len(jc.GpuType) == 0 {
		return nil
	}
	return &slurmjob.QueryFilter{
		Account:     jc.Account,
		Partition:   jc.Partition,
		Reservation: jc.Reservation,
		GpuType:     jc.GpuType,
		State:       jc.State,
	}
}
