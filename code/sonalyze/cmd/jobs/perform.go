package jobs

import (
	"fmt"
	"io"
	"maps"
	"slices"
	"strconv"
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
	kSampleCount           // Number of samples
	numF64Fields
)

// A number of fields could come *either* from the sample job or the slurm job, depending on what
// data we have, and it's the result of that joining that we want to print when we print jobs.  If
// there is a SlurmJob but not a SampleJob then a SampleJob is synthesized from the SlurmJob data.
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

// As we have a config, logclean will have computed proper GPU memory values for the job, so we need
// not look to sys.GpuMemPct here.
type hostResources struct {
	cpuCores int
	memGB    int
	gpuCards int
	gpuMemGB int
}

func (jc *JobsCommand) Perform(
	out io.Writer,
	cfg *config.ClusterConfig,
	theDb JobsDataProvider,
	filter sample.QueryFilter,
	hosts *Hosts,
	recordFilter *sample.SampleFilter,
) error {
	var needConfig = NeedsConfig(jobsFormatters, jc.PrintFields)
	if needConfig && cfg == nil {
		return fmt.Errorf("Configuration file required for relative format arguments")
	}

	isMergeable := func(k sample.InputStreamKey) bool {
		// TODO: Eventually we'll need to use the epoch here
		var sys *config.NodeConfigRecord
		if cfg != nil {
			sys = cfg.LookupHost(k.Host.String())
		}
		return sys != nil && sys.CrossNodeJobs
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

	sampleJobs, err := jc.findSampleJobs(
		isMergeable,
		theDb,
		filter,
		hosts,
		recordFilter,
	)
	if err != nil {
		return err
	}
	for _, j := range sampleJobs {
		smap[j.JobId] = &jobSummary{sampleJob: j, selected: true}
	}

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
				slurmJob:  j,
				selected:  true,
			}
		}
	}

	for _, j := range smap {
		jc.computeComputedFields(j, cfg)
	}

	if sampleFilter := jc.buildSampleFilter(cfg != nil); sampleFilter != nil {
		maps.DeleteFunc(smap, func(k uint32, v *jobSummary) bool {
			return !sampleFilter.apply(v)
		})
	}

	var summaries = slices.Collect(maps.Values(smap))

	var now = time.Now().UTC().Unix()
	for i := range summaries {
		summaries[i].Now = now
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

	return jc.printJobSummaries(out, summaries)
}

func (jc *JobsCommand) findSampleJobs(
	isMergeable func(sample.InputStreamKey) bool,
	theDb JobsDataProvider,
	filter sample.QueryFilter,
	hosts *Hosts,
	recordFilter *sample.SampleFilter,
) ([]*samplejob.SampleJob, error) {
	var merge samplejob.Merge
	switch {
	case jc.MergeAll:
		merge = samplejob.MergeAll
	case jc.MergeNone:
		merge = samplejob.MergeNone
	}

	sampleJobs, err := samplejob.Query(
		theDb,
		isMergeable,
		filter.FromDate,
		filter.ToDate,
		hosts,
		recordFilter,
		false,
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
	// Compute gpus from ReqGPUS.  This is not so easy b/c it does not necessarily have indices,
	// just device counts.  We fake it.  We use only the first element in the list because the
	// meaning of the list is that it is ordered from highest to lowest precedence.  Not obvious how
	// to incorporate the other steps here.  But since we take the value from the request and not
	// from any kind of usage data, it's probably OK to just use Main.
	if j.Main.ReqGPUS != UstrEmpty {
		var v uint32
		a, _, _ := strings.Cut(j.Main.ReqGPUS.String(), ",")
		_, x, _ := strings.Cut(a, "=")
		n, err := strconv.ParseUint(x, 10, 64)
		if err != nil {
			n = 1
		}
		for i := uint64(0); i < n; i++ {
			gpus, _ = gpuset.Adjoin(gpus, v)
			v++
		}
	}
	var hosts *Hostnames = NewHostnames()
	err := hosts.AddCompressed(j.Main.NodeList.String())
	if err != nil {
		Log.Warningf("Bad node list from slurm data: %s", j.Main.NodeList.String())
	}
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
	// Set SampleCount to 1000 b/c we want the synthesized job to pass any plausible min-samples
	// filter.  Note this value will be used as a divisor for various averages in
	// computeComputedFields.
	sampleCount := 1000
	adjustedSampleCount := float64(sampleCount) / float64(len(j.Steps)+1)

	cpuPctSum := float64(j.Main.AveCPU)
	cpuPctMax := float64(j.Main.AveCPU)
	cpuKBSum := uint64(j.Main.AveVMSize)
	rssAnonKBSum := uint64(j.Main.AveRSS)
	cpuKBMax := uint64(j.Main.MaxVMSize)
	rssAnonKBMax := uint64(j.Main.MaxRSS)
	for _, s := range j.Steps {
		cpuPctSum += float64(s.AveCPU)
		cpuPctMax = max(cpuPctMax, float64(s.AveCPU))
		cpuKBSum += uint64(s.AveVMSize)
		rssAnonKBSum += uint64(s.AveRSS)
		cpuKBMax = max(cpuKBMax, uint64(s.MaxVMSize))
		rssAnonKBMax = max(rssAnonKBMax, uint64(s.MaxRSS))
	}

	// Scale GB -> KB
	cpuKBSum *= 1024 * 1024
	rssAnonKBSum *= 1024 * 1024
	cpuKBMax *= 1024 * 1024
	rssAnonKBMax *= 1024 * 1024

	// Adjust according to sampleCount so that later averaging will work out
	cpuPctSum *= adjustedSampleCount
	cpuKBSum = uint64(float64(cpuKBSum) * adjustedSampleCount)
	rssAnonKBSum = uint64(float64(rssAnonKBSum) * adjustedSampleCount)

	var sampleJob = &samplejob.SampleJob{
		// `GpuFail` is not computable
		Gpus: gpus,
		// `IsZombie` is not applicable
		Cmd:      j.Main.JobName.String(),
		Hosts:    hosts,
		JobId:    j.Id,
		User:     j.Main.User,
		Duration: DurationValue(j.Main.ElapsedRaw),
		Start:    DateTimeValue(j.Main.Start),
		End:      DateTimeValue(j.Main.End),
		// `Job` is not applicable.
		SampleCount:    uint64(sampleCount),
		Classification: classification,
		ComputedFlags:  flags,
		CpuPctSum:      cpuPctSum,
		CpuKBSum:       cpuKBSum,
		RssAnonKBSum:   rssAnonKBSum,
		// `GpuPctSum` is not applicable
		// `GpuKBSum` is not applicable
		CpuTime: DurationValue(j.Main.SystemCPU + j.Main.UserCPU),
		// `GpuTime` is not applicable
		CpuPctMax:    cpuPctMax,
		CpuKBMax:     cpuKBMax,
		RssAnonKBMax: rssAnonKBMax,
		// `GpuPctMax` is not applicable
		// `GpuKBMax` is not applicable
	}
	return sampleJob
}

// The computed fields are always relative to js.sampleJob but the available resources may be
// computed from the slurmJob as well (notably the allocation).

func (jc *JobsCommand) computeComputedFields(js *jobSummary, cfg *config.ClusterConfig) {
	j := js.sampleJob
	js.computed[kDuration] = float64(j.Duration)
	js.computed[kSampleCount] = float64(j.SampleCount)

	// What things mean:
	//
	// "Average cpu utilization" is the sum across all samples of cpu utilization (which can be
	// >100% for individual samples in merged jobs, jobs with multiple threads, etc), divided by the
	// number of samples.
	//
	// "Average relative cpu utilization" further divides that by the number of cores allocated (ie,
	// "relative" is always "relative to allocated resources" or "a fraction of allocated resources").
	//
	// If there's a peak here it would be the "peak cpu utilization" across the time series, ie, the
	// peak observed value of the job's cpu utilization.  Again this is a sum across streams and can
	// easily be >100%.
	//
	// The "relative peak" divides the peak by the number of cores allocated.

	var rCpuPctAvg, rCpuPctPeak float64
	var rCpuGBAvg, rCpuGBPeak float64
	var rRssAnonGBAvg, rRssAnonGBPeak float64
	var rGpuPctAvg, rGpuPctPeak float64
	var rGpuGBAvg, rGpuGBPeak float64
	var sGpuPctAvg, sGpuPctPeak float64
	var sGpuGBAvg, sGpuGBPeak float64
	// Division by number of samples happens below, as necessary
	if cfg != nil {
		if sys := jc.allocatedResources(js, cfg); sys != nil {
			// Quantities can be zero in surprising ways, so always guard divisions
			if cores := float64(sys.cpuCores); cores > 0 {
				rCpuPctAvg = j.CpuPctSum / cores
				rCpuPctPeak = j.CpuPctMax / cores
			}
			if memory := float64(sys.memGB); memory > 0 {
				rCpuGBAvg = float64(j.CpuKBSum*100) / memory / (1024 * 1024)
				rCpuGBPeak = float64(j.CpuKBMax*100) / memory / (1024 * 1024)
				rRssAnonGBAvg = float64(j.RssAnonKBSum*100) / memory / (1024 * 1024)
				rRssAnonGBPeak = float64(j.RssAnonKBMax*100) / memory / (1024 * 1024)
			}
			if gpuCards := float64(sys.gpuCards); gpuCards > 0 {
				rGpuPctAvg = j.GpuPctSum / gpuCards
				rGpuPctPeak = j.GpuPctMax / gpuCards
			}
			if gpuMemory := float64(sys.gpuMemGB); gpuMemory > 0 {
				rGpuGBAvg = float64(j.GpuKBSum*100) / gpuMemory / (1024 * 1024)
				rGpuGBPeak = float64(j.GpuKBMax*100) / gpuMemory / (1024 * 1024)
			}
			if !js.sampleJob.Gpus.IsUnknown() && !js.sampleJob.Gpus.IsEmpty() {
				nCards := float64(js.sampleJob.Gpus.Size())
				sGpuPctAvg = j.GpuPctSum / nCards
				sGpuPctPeak = j.GpuPctMax / nCards
				if gpuCards := float64(sys.gpuCards); gpuCards > 0 {
					if gpuMemory := float64(sys.gpuMemGB); gpuMemory > 0 {
						jobGpuGB := nCards * (gpuMemory / gpuCards)
						sGpuGBAvg = float64(j.GpuKBSum*100) / jobGpuGB / (1024 * 1024)
						sGpuGBPeak = float64(j.GpuKBMax*100) / jobGpuGB / (1024 * 1024)
					}
				}
			}
		}
	}

	n := float64(j.SampleCount)
	js.computed[kCpuPctAvg] = j.CpuPctSum / n
	js.computed[kCpuPctPeak] = j.CpuPctMax
	js.computed[kRcpuPctAvg] = rCpuPctAvg / n
	js.computed[kRcpuPctPeak] = rCpuPctPeak

	js.computed[kCpuGBAvg] = float64(j.CpuKBSum) / n / (1024 * 1024)
	js.computed[kCpuGBPeak] = float64(j.CpuKBMax) / (1024 * 1024)
	js.computed[kRcpuGBAvg] = rCpuGBAvg / n
	js.computed[kRcpuGBPeak] = rCpuGBPeak

	js.computed[kRssAnonGBAvg] = float64(j.RssAnonKBSum) / n / (1024 * 1024)
	js.computed[kRssAnonGBPeak] = float64(j.RssAnonKBMax) / (1024 * 1024)
	js.computed[kRrssAnonGBAvg] = rRssAnonGBAvg / n
	js.computed[kRrssAnonGBPeak] = rRssAnonGBPeak

	js.computed[kGpuPctAvg] = j.GpuPctSum / n
	js.computed[kGpuPctPeak] = j.GpuPctMax
	js.computed[kRgpuPctAvg] = rGpuPctAvg / n
	js.computed[kRgpuPctPeak] = rGpuPctPeak
	js.computed[kSgpuPctAvg] = sGpuPctAvg / n
	js.computed[kSgpuPctPeak] = sGpuPctPeak

	js.computed[kGpuGBAvg] = float64(j.GpuKBSum) / n / (1024 * 1024)
	js.computed[kGpuGBPeak] = float64(j.GpuKBMax) / (1024 * 1024)
	js.computed[kRgpuGBAvg] = rGpuGBAvg / n
	js.computed[kRgpuGBPeak] = rGpuGBPeak
	js.computed[kSgpuGBAvg] = sGpuGBAvg / n
	js.computed[kSgpuGBPeak] = sGpuGBPeak
}

// If there is a single host then we just get the config's data.  For multiple hosts we sum those
// data.  We don't cache anything now b/c the underlying config code has a hashmap already.
//
// TODO: In the future, for jobs with a slurm aspect, we'll instead return the allocated resources,
// when available, falling back on config values when necessary.  In the future after that, we're
// not going to have a config in the static sense, but a sysinfo blob that was valid at the time of
// the sample.  That can be lazily computed and for slurm jobs it may not need to be computed at
// all.

func (jc *JobsCommand) allocatedResources(js *jobSummary, cfg *config.ClusterConfig) *hostResources {
	sum := new(hostResources)
	for name := range js.sampleJob.Hosts.FullNames {
		if sys := cfg.LookupHost(name); sys != nil {
			sum.cpuCores += sys.CpuCores
			sum.memGB += sys.MemGB
			sum.gpuCards += sys.GpuCards
			sum.gpuMemGB += sys.GpuMemGB
		}
	}
	return sum
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

	if len(minFilters) == 0 && len(maxFilters) == 0 && flags == 0 {
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
