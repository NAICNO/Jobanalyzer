// Mini-version of `sonalyze jobs`

package main

import (
	"fmt"
	"math"

	"go-utils/sonarlog"
)

const (
	minSamples = 2
	LiveAtStart = 1
	LiveAtEnd = 2
)

type jobSummary struct {
	aggregate jobAggregate
	job sonarlog.SampleStream
}

// The float fields of this are not rounded.

type jobAggregate struct {
	first int64					// seconds since epoch
	last int64					// seconds since epoch
	duration int64				// total seconds
	minutes int64				// m part of d:h:m
	hours int64					// h part of d:h:m
	days int64					// d part of d:h:m
	usesGpu bool
	gpuFail uint8
	cpuPctAvg float64
	cpuPctPeak float64
	cpuKibAvg uint64
	cpuKibPeak uint64
	rssAnonKibAvg uint64
	rssAnonKibPeak uint64
	gpuPctAvg float64
	gpuPctPeak float64
	gpuKibAvg uint64
	gpuKibPeak uint64
	classification int

	// For printing
	selected bool
}

// "now"

type jobCtx int64

func jobs(readings sonarlog.SampleStream) {
	streams := sonarlog.PostprocessLog(readings, func(s *sonarlog.Sample) bool { return true }, nil)
	jobs := sonarlog.MergeByHostAndJob(streams)

	summaries := make([]*jobSummary, 0)
	for _, job := range jobs {
		if len(*job) >= minSamples {
			summaries = append(summaries, &jobSummary{
				aggregate: aggregateJob(*job),
				job: *job,
			})
		}
	}

	hosts := make(map[sonarlog.Ustr]bool)
	for _, s := range streams {
		hosts[(*s)[0].Host] = true
	}

	// Print

	// sort summaries by timestamp (primary) and job id (secondary)
	// TODO

	// We have selected all jobs... and no relative fields... so, it's easy

	FormatData(
		jobDefaultFields,
		jobFormatters,
		&FormatOptions{
			Fixed: true,
			Header: true,
		},
		summaries,
		jobCtx(0),
	)
}

// Given a list of log entries for a job, sorted ascending by timestamp and with no duplicated
// timestamps, and the earliest and latest timestamps from all records read, return a JobAggregate
// for the job.

func aggregateJob(job sonarlog.SampleStream) jobAggregate {
	first := job[0].Timestamp
	last := job[len(job)-1].Timestamp
	//host := job[0].Host
	duration := last - first
	minutes := int64(math.Round(float64(duration) / 60))
	// Accumulators
	var (
		usesGpu bool
		gpuFail uint8
		cpuPctAvg, gpuPctAvg, cpuPctPeak, gpuPctPeak float64
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
		cpuKibPeak = maxu64(cpuKibPeak, s.CpuKib)
		rssAnonKibAvg += s.RssAnonKib
		rssAnonKibPeak = maxu64(rssAnonKibPeak, s.RssAnonKib)
		gpuKibAvg += s.GpuKib
		gpuKibPeak = maxu64(gpuKibPeak, s.GpuKib)
	}
	classification := 0
	n := uint64(len(job))
	fn := float64(len(job))
	return jobAggregate{
        first: first,
		last: last,
        duration: duration,
        minutes: minutes % 60,
        hours: (minutes / 60) % 24,
        days: minutes / (60 * 24),
        usesGpu: usesGpu,
        gpuFail: gpuFail,
        cpuPctAvg: cpuPctAvg / fn,
        cpuPctPeak: cpuPctPeak,
		cpuKibAvg: cpuKibAvg / n,
		cpuKibPeak: cpuKibPeak,
		rssAnonKibAvg: rssAnonKibAvg / n,
		rssAnonKibPeak: rssAnonKibPeak,
        gpuPctAvg: gpuPctAvg / fn,
        gpuPctPeak: gpuPctPeak,
		gpuKibAvg: gpuKibAvg / n,
		gpuKibPeak: gpuKibPeak,
        classification: classification,
        selected: true,
    }
}

func maxu64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

var jobDefaultFields = []string{
	"jobm",
	"user",
	/*
	"duration",
	"host",
	"cpu-avg",
	"mem-avg",
	"gpu-avg",
	"gpumem-avg",
	"cmd",
	*/
}

var jobFormatters = map[string]func(d *jobSummary, ctx jobCtx) string {

	"jobm": func(d *jobSummary, _ jobCtx) string {
		mark := ""
		c := d.aggregate.classification
		switch {
		case c & (LiveAtStart|LiveAtEnd) == (LiveAtStart|LiveAtEnd):
			mark = "!"
		case c & LiveAtStart != 0:
			mark = "<"
		case c & LiveAtEnd != 0:
			mark = ">"
		}
		return fmt.Sprint(d.job[0].Job, mark)
	},

	"user": func(d *jobSummary, _ jobCtx) string {
		return d.job[0].User.String()
	},

}
