package jobs

import (
	"fmt"
	"io"
	"math"
	"sort"
	"strings"
	"time"

	"go-utils/gpuset"
	"go-utils/hostglob"
	"go-utils/maps"
	"go-utils/sonalyze"

	. "sonalyze/command"
	. "sonalyze/common"
)

func (jc *JobsCommand) printRequiresConfig() bool {
	for _, f := range jc.PrintFields {
		switch f {
		case "rcpu-avg", "rcpu-peak", "rmem-avg", "rmem-peak", "rgpu-avg", "rgpu-peak",
			"rgpumem-avg", "rgpumem-peak", "rres-avg", "rres-peak":
			return true
		}
	}
	return false
}

type sortableSummaries []*jobSummary

func (ss sortableSummaries) Len() int {
	return len(ss)
}

func (ss sortableSummaries) Swap(i, j int) {
	ss[i], ss[j] = ss[j], ss[i]
}

func (ss sortableSummaries) Less(i, j int) bool {
	if ss[i].aggregate.first == ss[j].aggregate.first {
		return ss[i].job[0].S.Job < ss[j].job[0].S.Job
	}
	return ss[i].aggregate.first < ss[j].aggregate.first
}

func (jc *JobsCommand) printJobSummaries(out io.Writer, summaries []*jobSummary) error {
	// Sort ascending by lowest beginning timestamp, and if those are equal, by job number.
	sort.Stable(sortableSummaries(summaries))

	// Select a number of jobs per user, if applicable.  This means working from the bottom up
	// in the vector and marking the numJobs first per user.
	numRemoved := 0
	if jc.NumJobs > 0 {
		if jc.Verbose {
			Log.Infof("Selecting only %d top jobs per user", jc.NumJobs)
		}
		counts := make(map[Ustr]uint)
		for i := len(summaries) - 1; i >= 0; i-- {
			u := summaries[i].job[0].S.User
			c := counts[u] + 1
			counts[u] = c
			if c > jc.NumJobs {
				if summaries[i].aggregate.selected {
					numRemoved++
					summaries[i].aggregate.selected = false
				}
			}
		}
	}

	if jc.Verbose {
		Log.Infof("Number of jobs after output filtering: %d", len(summaries)-numRemoved)
	}

	// Pick the summaries that have been selected
	dst := 0
	for src := 0; src < len(summaries); src++ {
		if summaries[src].aggregate.selected {
			summaries[dst] = summaries[src]
			dst++
		}
	}
	summaries = summaries[:dst]

	FormatData(
		out,
		jc.PrintFields,
		jobsFormatters,
		jc.PrintOpts,
		summaries,
		jobCtx(jobCtx{
			now:         time.Now().UTC().Unix(),
			fixedFormat: jc.PrintOpts.Fixed,
		}),
	)
	return nil
}

func (jc *JobsCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(jc.Fmt, jobsHelp, jobsFormatters, jobsAliases, jobsDefaultFields)
}

const jobsHelp = `
jobs
  Aggregate process data into data about "jobs" and present them.  Output
  records are sorted in order of increasing start time of the job. The default
  format is 'fixed'.
`

type jobCtx struct {
	now         int64
	fixedFormat bool
}

const jobsDefaultFields = "std,cpu,mem,gpu,gpumem,cmd"

// MT: Constant after initialization; immutable
var jobsAliases = map[string][]string{
	"std":     []string{"jobm", "user", "duration", "host"},
	"cpu":     []string{"cpu-avg", "cpu-peak"},
	"rcpu":    []string{"rcpu-avg", "rcpu-peak"},
	"mem":     []string{"mem-avg", "mem-peak"},
	"rmem":    []string{"rmem-avg", "rmem-peak"},
	"res":     []string{"res-avg", "res-peak"},
	"rres":    []string{"rres-avg", "rres-peak"},
	"gpu":     []string{"gpu-avg", "gpu-peak"},
	"rgpu":    []string{"rgpu-avg", "rgpu-peak"},
	"sgpu":    []string{"sgpu-avg", "sgpu-peak"},
	"gpumem":  []string{"gpumem-avg", "gpumem-peak"},
	"rgpumem": []string{"rgpumem-avg", "rgpumem-peak"},
	"sgpumem": []string{"sgpumem-avg", "sgpumem-peak"},
}

const (
	KibToGibFactor = 1024 * 1024
)

// MT: Constant after initialization; immutable
var jobsFormatters = map[string]Formatter[*jobSummary, jobCtx]{
	"jobm": {
		func(d *jobSummary, _ jobCtx) string {
			mark := ""
			c := d.aggregate.computedFlags
			switch {
			case c&(kIsLiveAtStart|kIsLiveAtEnd) == (kIsLiveAtStart | kIsLiveAtEnd):
				mark = "!"
			case c&kIsLiveAtStart != 0:
				mark = "<"
			case c&kIsLiveAtEnd != 0:
				mark = ">"
			}
			return fmt.Sprint(d.job[0].S.Job, mark)
		},
		"Job ID with mark indicating job running at start+end (!), start (<), or end (>) of time window",
	},
	"job": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(d.job[0].S.Job)
		},
		"Job ID",
	},
	"user": {
		func(d *jobSummary, _ jobCtx) string {
			return d.job[0].S.User.String()
		},
		"Name of user running the job",
	},
	"duration": {
		func(d *jobSummary, _ jobCtx) string {
			mins := int64(math.Round(d.aggregate.computed[kDuration] / 60))
			minutes := mins % 60
			hours := (mins / 60) % 24
			days := mins / (60 * 24)
			return fmt.Sprintf("%dd%2dh%2dm", days, hours, minutes)
		},
		"Duration in minutes of job: time of last observation minus time of first (DdHhMm)",
	},
	"duration/sec": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(int64(d.aggregate.computed[kDuration]))
		},
		"Duration in seconds of job: time of last observation minus time of first",
	},
	"start": {
		func(d *jobSummary, _ jobCtx) string {
			return FormatYyyyMmDdHhMmUtc(d.aggregate.first)
		},
		"Time of first observation (yyyy-dd-mm hh:mm)",
	},
	"start/sec": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(d.aggregate.first)
		},
		"Time of first observation (Unix timestamp)",
	},
	"end": {
		func(d *jobSummary, _ jobCtx) string {
			return FormatYyyyMmDdHhMmUtc(d.aggregate.last)
		},
		"Time of last observation (yyyy-dd-mm hh:mm)",
	},
	"end/sec": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(d.aggregate.last)
		},
		"Time of last observation (Unix timestamp)",
	},
	"cpu-avg": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(uint64(math.Ceil(d.aggregate.computed[kCpuPctAvg])))
		},
		"Average CPU utilization in percent (100% = 1 core)",
	},
	"cpu-peak": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(uint64(math.Ceil(d.aggregate.computed[kCpuPctPeak])))
		},
		"Peak CPU utilization in percent (100% = 1 core)",
	},
	"rcpu-avg": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(uint64(math.Ceil(d.aggregate.computed[kRcpuPctAvg])))
		},
		"Average relative CPU utilization in percent (100% = all cores)",
	},
	"rcpu-peak": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(uint64(math.Ceil(d.aggregate.computed[kRcpuPctPeak])))
		},
		"Peak relative CPU utilization in percent (100% = all cores)",
	},
	"mem-avg": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(uint64(math.Ceil(d.aggregate.computed[kCpuGibAvg])))
		},
		"Average main virtual memory utilization in GiB",
	},
	"mem-peak": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(uint64(math.Ceil(d.aggregate.computed[kCpuGibPeak])))
		},
		"Peak main virtual memory utilization in GiB",
	},
	"rmem-avg": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(uint64(math.Ceil(d.aggregate.computed[kRcpuGibAvg])))
		},
		"Average relative main virtual memory utilization in percent (100% = system RAM)",
	},
	"rmem-peak": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(uint64(math.Ceil(d.aggregate.computed[kRcpuGibPeak])))
		},
		"Peak relative main virtual memory utilization in percent (100% = system RAM)",
	},
	"res-avg": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(uint64(math.Ceil(d.aggregate.computed[kRssAnonGibAvg])))
		},
		"Average main resident memory utilization in GiB",
	},
	"res-peak": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(uint64(math.Ceil(d.aggregate.computed[kRssAnonGibPeak])))
		},
		"Peak main resident memory utilization in GiB",
	},
	"rres-avg": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(uint64(math.Ceil(d.aggregate.computed[kRrssAnonGibAvg])))
		},
		"Average relative main resident memory utilization in percent (100% = all RAM)",
	},
	"rres-peak": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(uint64(math.Ceil(d.aggregate.computed[kRrssAnonGibPeak])))
		},
		"Peak relative main resident memory utilization in percent (100% = all RAM)",
	},
	"gpu-avg": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(uint64(math.Ceil(d.aggregate.computed[kGpuPctAvg])))
		},
		"Average GPU utilization in percent (100% = 1 card)",
	},
	"gpu-peak": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(uint64(math.Ceil(d.aggregate.computed[kGpuPctPeak])))
		},
		"Peak GPU utilization in percent (100% = 1 card)",
	},
	"rgpu-avg": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(uint64(math.Ceil(d.aggregate.computed[kRgpuPctAvg])))
		},
		"Average relative GPU utilization in percent (100% = all cards)",
	},
	"rgpu-peak": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(uint64(math.Ceil(d.aggregate.computed[kRgpuPctPeak])))
		},
		"Peak relative GPU utilization in percent (100% = all cards)",
	},
	"sgpu-avg": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(uint64(math.Ceil(d.aggregate.computed[kSgpuPctAvg])))
		},
		"Average relative GPU utilization in percent (100% = all cards used by job)",
	},
	"sgpu-peak": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(uint64(math.Ceil(d.aggregate.computed[kSgpuPctPeak])))
		},
		"Peak relative GPU utilization in percent (100% = all cards used by job)",
	},
	"gpumem-avg": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(uint64(math.Ceil(d.aggregate.computed[kGpuGibAvg])))
		},
		"Average resident GPU memory utilization in GiB",
	},
	"gpumem-peak": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(uint64(math.Ceil(d.aggregate.computed[kGpuGibPeak])))
		},
		"Peak resident GPU memory utilization in GiB",
	},
	"rgpumem-avg": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(uint64(math.Ceil(d.aggregate.computed[kRgpuGibAvg])))
		},
		"Average relative GPU resident memory utilization in percent (100% = all GPU RAM)",
	},
	"rgpumem-peak": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(uint64(math.Ceil(d.aggregate.computed[kRgpuGibPeak])))
		},
		"Peak relative GPU resident memory utilization in percent (100% = all GPU RAM)",
	},
	"sgpumem-avg": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(uint64(math.Ceil(d.aggregate.computed[kSgpuGibAvg])))
		},
		"Average relative GPU resident memory utilization in percent (100% = all GPU RAM on cards used by job)",
	},
	"sgpumem-peak": {
		func(d *jobSummary, _ jobCtx) string {
			return fmt.Sprint(uint64(math.Ceil(d.aggregate.computed[kSgpuGibPeak])))
		},
		"Peak relative GPU resident memory utilization in percent (100% = all GPU RAM on cards used by job)",
	},
	"gpus": {
		func(d *jobSummary, _ jobCtx) string {
			gpus := gpuset.EmptyGpuSet()
			for _, j := range d.job {
				gpus = gpuset.UnionGpuSets(gpus, j.S.Gpus)
			}
			return gpus.String()
		},
		"GPU device numbers used by the job, 'none' if none or 'unknown' in error states",
	},
	"gpufail": {
		func(d *jobSummary, _ jobCtx) string {
			if (d.aggregate.computedFlags & kGpuFail) != 0 {
				return "1"
			}
			return "0"
		},
		"Flag indicating GPU status (0=Ok, 1=Failing)",
	},
	"cmd": {
		func(d *jobSummary, _ jobCtx) string {
			names := make(map[Ustr]bool)
			name := ""
			for _, sample := range d.job {
				if _, found := names[sample.S.Cmd]; found {
					continue
				}
				if name != "" {
					name += ", "
				}
				name += sample.S.Cmd.String()
				names[sample.S.Cmd] = true
			}
			return name
		},
		"The commands invoking the processes of the job",
	},
	"host": {
		func(d *jobSummary, c jobCtx) string {
			hosts := make(map[string]bool)
			for _, s := range d.job {
				var name string
				if c.fixedFormat {
					name, _, _ = strings.Cut(s.S.Host.String(), ".")
				} else {
					name = s.S.Host.String()
				}
				hosts[name] = true
			}
			return strings.Join(hostglob.CompressHostnames(maps.Keys(hosts)), ", ")
		},
		"List of the host name(s) running the job (first elements of FQDNs, compressed)",
	},
	"now": {
		func(_ *jobSummary, c jobCtx) string {
			return FormatYyyyMmDdHhMmUtc(c.now)
		},
		"The current time (yyyy-mm-dd hh:mm)",
	},
	"now/sec": {
		func(_ *jobSummary, c jobCtx) string {
			return fmt.Sprint(c.now)
		},
		"The current time (Unix timestamp)",
	},
	"classification": {
		func(d *jobSummary, _ jobCtx) string {
			n := 0
			if (d.aggregate.computedFlags & kIsLiveAtStart) != 0 {
				n |= sonalyze.LIVE_AT_START
			}
			if (d.aggregate.computedFlags & kIsLiveAtEnd) != 0 {
				n |= sonalyze.LIVE_AT_END
			}
			return fmt.Sprint(n)
		},
		"Bit vector of live-at-start (2) and live-at-end (1) flags",
	},
	"cputime/sec": {
		func(d *jobSummary, _ jobCtx) string {
			// The unit for average cpu utilization is core-seconds per second, we multiply this by
			// duration (whose units is second) to get total core-seconds for the job.  Finally scale by
			// 100 because the cpu_avg numbers are expressed in integer percentage point.
			//
			// Note, this `duration` may be incompatible from the Rust version, as we have second resolution.
			duration := d.aggregate.computed[kDuration]
			cputimeSec := int64(math.Round(d.aggregate.computed[kCpuPctAvg] * duration / 100))
			return fmt.Sprint(cputimeSec)
		},
		"Total CPU time of the job across all cores (seconds)",
	},
	"cputime": {
		func(d *jobSummary, _ jobCtx) string {
			// As above, but formatted and rounded differently.
			duration := d.aggregate.computed[kDuration]
			cputimeSec := int64(math.Round(d.aggregate.computed[kCpuPctAvg] * duration / 100))
			if cputimeSec%60 >= 30 {
				cputimeSec += 30
			}
			minutes := (cputimeSec / 60) % 60
			hours := (cputimeSec / (60 * 60)) % 24
			days := cputimeSec / (60 * 60 * 24)
			return fmt.Sprintf("%2dd%2dh%2dm", days, hours, minutes)
		},
		"Total CPU time of the job across all cores (DdHhMm)",
	},
	"gputime/sec": {
		func(d *jobSummary, _ jobCtx) string {
			// The unit for average gpu utilization is card-seconds per second, we multiply this by
			// duration (whose units is second) to get total card-seconds for the job.  Finally scale by
			// 100 because the gpu_avg numbers are expressed in integer percentage point.
			//
			// Note, this `duration` may be incompatible from the Rust version, as we have second resolution.
			duration := d.aggregate.computed[kDuration]
			gputimeSec := int64(math.Round(d.aggregate.computed[kGpuPctAvg] * duration / 100))
			return fmt.Sprint(gputimeSec)
		},
		"Total GPU time of the job across all cards (seconds)",
	},
	"gputime": {
		func(d *jobSummary, _ jobCtx) string {
			// As above, but formatted and rounded differently.
			duration := d.aggregate.computed[kDuration]
			gputimeSec := int64(math.Round(d.aggregate.computed[kGpuPctAvg] * duration / 100))
			if gputimeSec%60 >= 30 {
				gputimeSec += 30
			}
			minutes := (gputimeSec / 60) % 60
			hours := (gputimeSec / (60 * 60)) % 24
			days := gputimeSec / (60 * 60 * 24)
			return fmt.Sprintf("%2dd%2dh%2dm", days, hours, minutes)
		},
		"Total GPU time of the job across all cards (DdHhMm)",
	},
}
