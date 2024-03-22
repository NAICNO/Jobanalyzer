// Mini-version of `sonalyze jobs`

package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"go-utils/hostglob"
	"go-utils/minmax"
	"go-utils/sonalyze"
	"go-utils/sonarlog"
)

const (
	minSamples = 2
)

type jobSummary struct {
	aggregate jobAggregate
	job       sonarlog.SampleStream
}

// The float fields of this are not rounded.
type jobAggregate struct {
	first          int64 // seconds since epoch
	last           int64 // seconds since epoch
	duration       int64 // total seconds
	minutes        int64 // m part of d:h:m
	hours          int64 // h part of d:h:m
	days           int64 // d part of d:h:m
	usesGpu        bool
	gpuFail        uint8
	cpuPctAvg      float64
	cpuPctPeak     float64
	cpuKibAvg      uint64
	cpuKibPeak     uint64
	rssAnonKibAvg  uint64
	rssAnonKibPeak uint64
	gpuPctAvg      float64
	gpuPctPeak     float64
	gpuKibAvg      uint64
	gpuKibPeak     uint64
	classification int

	// For printing
	selected bool
}

type jobCtx struct {
	now         int64
	fixedFormat bool
}

type SortableSummaries []*jobSummary

func (ss SortableSummaries) Len() int      { return len(ss) }
func (ss SortableSummaries) Swap(i, j int) { ss[i], ss[j] = ss[j], ss[i] }
func (ss SortableSummaries) Less(i, j int) bool {
	if ss[i].aggregate.first == ss[j].aggregate.first {
		return ss[i].job[0].Job < ss[j].job[0].Job
	}
	return ss[i].aggregate.first < ss[j].aggregate.first
}

func jobs(readings sonarlog.SampleStream) error {
	// Time bounds are computed from the full set of samples before filtering.
	bounds := sonarlog.ComputeTimeBounds(readings)

	// TODO: if -from or -to or there are no bounds then we do further processing here.

	streams := sonarlog.PostprocessLog(
		readings,
		func(s *sonarlog.Sample) bool {
			return true // s.Job == 25153
		},
		nil,
	)
	jobs := sonarlog.MergeByHostAndJob(streams)
	summaries := make([]*jobSummary, 0)
	for _, job := range jobs {
		if len(*job) >= minSamples {
			summaries = append(summaries, &jobSummary{
				aggregate: aggregateJob(*job, bounds),
				job:       *job,
			})
		}
	}

	// hosts := make(map[sonarlog.Ustr]bool)
	// for _, s := range streams {
	// 	hosts[(*s)[0].Host] = true
	// }

	// Print
	// We have selected all jobs... and no relative fields... so, it's easy

	sort.Sort(SortableSummaries(summaries))

	jobFields, _, err := ParseFields(jobDefaultFields, jobFormatters, jobAliases)
	if err != nil {
		return err
	}
	FormatData(
		jobFields,
		jobFormatters,
		&FormatOptions{
			Fixed:  true,
			Header: true,
		},
		summaries,
		jobCtx(jobCtx{now: time.Now().UTC().Unix(), fixedFormat: true}),
	)
	return nil
}

// Given a list of log entries for a job, sorted ascending by timestamp and with no duplicated
// timestamps, and the earliest and latest timestamps from all records read, return a JobAggregate
// for the job.

func aggregateJob(job sonarlog.SampleStream, bounds map[sonarlog.Ustr]sonarlog.Timebound) jobAggregate {
	first := job[0].Timestamp
	last := job[len(job)-1].Timestamp
	host := job[0].Host
	duration := last - first
	minutes := int64(math.Round(float64(duration) / 60))
	// Accumulators
	var (
		usesGpu                                                                     bool
		gpuFail                                                                     uint8
		cpuPctAvg, gpuPctAvg, cpuPctPeak, gpuPctPeak                                float64
		cpuKibAvg, cpuKibPeak, rssAnonKibAvg, rssAnonKibPeak, gpuKibAvg, gpuKibPeak uint64
	)
	for _, s := range job {
		usesGpu = usesGpu || !s.Gpus.IsEmpty()
		gpuFail = sonarlog.MergeGpuFail(gpuFail, s.GpuFail)
		cpuPctAvg += float64(s.CpuUtilPct)
		cpuPctPeak = math.Max(cpuPctPeak, float64(s.CpuUtilPct))
		gpuPctAvg += float64(s.GpuPct)
		gpuPctPeak = math.Max(gpuPctPeak, float64(s.GpuPct))
		cpuKibAvg += s.CpuKib
		cpuKibPeak = minmax.MaxUint64(cpuKibPeak, s.CpuKib)
		rssAnonKibAvg += s.RssAnonKib
		rssAnonKibPeak = minmax.MaxUint64(rssAnonKibPeak, s.RssAnonKib)
		gpuKibAvg += s.GpuKib
		gpuKibPeak = minmax.MaxUint64(gpuKibPeak, s.GpuKib)
	}
	classification := 0
	bound, haveBound := bounds[host]
	if !haveBound {
		panic("Expected to find bound")
	}
	if first == bound.Earliest {
		classification |= sonalyze.LIVE_AT_START
	}
	if last == bound.Latest {
		classification |= sonalyze.LIVE_AT_END
	}
	//    fmt.Println(classification, time.Unix(first, 0), time.Unix(bound.Earliest, 0), time.Unix(last, 0), time.Unix(bound.Latest, 0));
	n := uint64(len(job))
	fn := float64(len(job))
	return jobAggregate{
		first:          first,
		last:           last,
		duration:       duration,
		minutes:        minutes % 60,
		hours:          (minutes / 60) % 24,
		days:           minutes / (60 * 24),
		usesGpu:        usesGpu,
		gpuFail:        gpuFail,
		cpuPctAvg:      cpuPctAvg / fn,
		cpuPctPeak:     cpuPctPeak,
		cpuKibAvg:      cpuKibAvg / n,
		cpuKibPeak:     cpuKibPeak,
		rssAnonKibAvg:  rssAnonKibAvg / n,
		rssAnonKibPeak: rssAnonKibPeak,
		gpuPctAvg:      gpuPctAvg / fn,
		gpuPctPeak:     gpuPctPeak,
		gpuKibAvg:      gpuKibAvg / n,
		gpuKibPeak:     gpuKibPeak,
		classification: classification,
		selected:       true,
	}
}

var jobDefaultFields = "std,cpu,mem,gpu,gpumem,cmd"

var jobAliases = map[string][]string{
	"std":    []string{"jobm", "user", "duration", "host"},
	"cpu":    []string{"cpu-avg", "cpu-peak"},
	"mem":    []string{"mem-avg", "mem-peak"},
	"gpu":    []string{"gpu-avg", "gpu-peak"},
	"gpumem": []string{"gpumem-avg", "gpumem-peak"},
}

const (
	KibToGibFactor = 1024 * 1024
)

var jobFormatters = map[string]func(d *jobSummary, ctx jobCtx) string{

	"jobm": func(d *jobSummary, _ jobCtx) string {
		mark := ""
		c := d.aggregate.classification
		switch {
		case c & (sonalyze.LIVE_AT_START|sonalyze.LIVE_AT_END) == (sonalyze.LIVE_AT_START|sonalyze.LIVE_AT_END):
			mark = "!"
		case c & sonalyze.LIVE_AT_START != 0:
			mark = "<"
		case c & sonalyze.LIVE_AT_END != 0:
			mark = ">"
		}
		return fmt.Sprint(d.job[0].Job, mark)
	},

	"user": func(d *jobSummary, _ jobCtx) string {
		return d.job[0].User.String()
	},

	"duration": func(d *jobSummary, _ jobCtx) string {
		return fmt.Sprintf("%dd%02dh%02dm", d.aggregate.days, d.aggregate.hours, d.aggregate.minutes)
	},

	"host": func(d *jobSummary, c jobCtx) string {
		hosts := make(map[string]bool)
		names := make([]string, 0, len(hosts))
		for _, s := range d.job {
			var name string
			if c.fixedFormat {
				name, _, _ = strings.Cut(s.Host.String(), ".")
			} else {
				name = s.Host.String()
			}
			if _, found := hosts[name]; !found {
				hosts[name] = true
				names = append(names, name)
			}
		}
		return strings.Join(hostglob.CompressHostnames(names), ", ")
	},

	"cpu-avg": func(d *jobSummary, _ jobCtx) string {
		return fmt.Sprintf("%g", math.Min(1000000, math.Ceil(d.aggregate.cpuPctAvg)))
	},

	"cpu-peak": func(d *jobSummary, _ jobCtx) string {
		return fmt.Sprintf("%g", math.Min(1000000, math.Ceil(d.aggregate.cpuPctPeak)))
	},

	"mem-avg": func(d *jobSummary, _ jobCtx) string {
		return fmt.Sprintf("%g", math.Ceil(float64(d.aggregate.cpuKibAvg)/KibToGibFactor))
	},

	"mem-peak": func(d *jobSummary, _ jobCtx) string {
		return fmt.Sprintf("%g", math.Ceil(float64(d.aggregate.cpuKibPeak)/KibToGibFactor))
	},

	"gpu-avg": func(d *jobSummary, _ jobCtx) string {
		return fmt.Sprintf("%g", math.Ceil(d.aggregate.gpuPctAvg))
	},

	"gpu-peak": func(d *jobSummary, _ jobCtx) string {
		return fmt.Sprintf("%g", math.Ceil(d.aggregate.gpuPctPeak))
	},

	"gpumem-avg": func(d *jobSummary, _ jobCtx) string {
		return fmt.Sprintf("%g", math.Ceil(float64(d.aggregate.gpuKibAvg)/KibToGibFactor))
	},

	"gpumem-peak": func(d *jobSummary, _ jobCtx) string {
		return fmt.Sprintf("%g", math.Ceil(float64(d.aggregate.gpuKibPeak)/KibToGibFactor))
	},

	"cmd": func(d *jobSummary, _ jobCtx) string {
		names := make(map[sonarlog.Ustr]bool)
		name := ""
		for _, sample := range d.job {
			if _, found := names[sample.Cmd]; found {
				continue
			}
			if name != "" {
				name += ", "
			}
			name += sample.Cmd.String()
			names[sample.Cmd] = true
		}
		return name
	},
}
