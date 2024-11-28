package jobs

import (
	"cmp"
	"io"
	"math"
	"slices"
	"strings"

	uslices "go-utils/slices"

	. "sonalyze/common"
	. "sonalyze/table"
)

// TODO: Several commands have the constraint that certain fields require a config, and aliases
// continue to be a headache too.  It's dumb to have multiple tables.  Possibly better to add the
// constraint to the field definition and let shared logic take care of it.
//
// (In practice there is always a config for remote commands so this is a hazard only for local
// uses.)

var relativeFields = map[string]bool{
	"RelativeCpuAvgPct":             true,
	"rcpu-avg":                      true,
	"RelativeCpuPeakPct":            true,
	"rcpu-peak":                     true,
	"RelativeMemAvgPct":             true,
	"rmem-avg":                      true,
	"RelativeMemPeakPct":            true,
	"rmem-peak":                     true,
	"RelativeResidentMemAvgPct":     true,
	"rres-avg":                      true,
	"RelativeResidentMemPeakPct":    true,
	"rres-peak":                     true,
	"RelativeGpuAvgPct":             true,
	"rgpu-avg":                      true,
	"RelativeGpuPeakPct":            true,
	"rgpu-peak":                     true,
	"RelativeGpuMemAvgPct":          true,
	"rgpumem-avg":                   true,
	"RelativeGpuMemPeakPct":         true,
	"rgpumem-peak":                  true,
	"OccupiedRelativeGpuAvgPct":     true,
	"sgpu-avg":                      true,
	"OccupiedRelativeGpuPeakPct":    true,
	"sgpu-peak":                     true,
	"OccupiedRelativeGpuMemAvgPct":  true,
	"sgpumem-avg":                   true,
	"OccupiedRelativeGpuMemPeakPct": true,
	"sgpumem-peak":                  true,
}

func (jc *JobsCommand) printRequiresConfig() bool {
	for _, f := range jc.PrintFields {
		if relativeFields[f.Name] {
			return true
		}
	}
	return false
}

func (jc *JobsCommand) printJobSummaries(out io.Writer, summaries []*jobSummary) error {
	// Sort ascending by lowest beginning timestamp, and if those are equal, by job number.
	slices.SortStableFunc(summaries, func(a, b *jobSummary) int {
		c := cmp.Compare(a.Start, b.Start)
		if c == 0 {
			c = cmp.Compare(a.JobId, b.JobId)
		}
		return c
	})

	// Select a number of jobs per user, if applicable.  This means working from the bottom up
	// in the vector and marking the numJobs first per user.
	numRemoved := 0
	if jc.NumJobs > 0 {
		if jc.Verbose {
			Log.Infof("Selecting only %d top jobs per user", jc.NumJobs)
		}
		counts := make(map[Ustr]uint)
		for i := len(summaries) - 1; i >= 0; i-- {
			u := summaries[i].job[0].User
			c := counts[u] + 1
			counts[u] = c
			if c > jc.NumJobs {
				if summaries[i].selected {
					numRemoved++
					summaries[i].selected = false
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
		if summaries[src].selected {
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
		uslices.Map(summaries, func(x *jobSummary) any { return x }),
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

const v0JobsDefaultFields = "std,cpu,mem,gpu,gpumem,cmd"
const v1JobsDefaultFields = "Std,Cpu,Mem,Gpu,GpuMem,Cmd"
const jobsDefaultFields = v0JobsDefaultFields

// Instead of struggling with how to represent formatters for the indexed accesses, just define
// the formatters directly, and handle aliases below.

// MT: Constant after initialization; immutable
var jobsFormatters = map[string]Formatter{
	"JobAndMark": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatString(d.(*jobSummary).JobAndMark, ctx)
		},
		Help: "Job ID with mark indicating job running at start+end (!), start (<), or end (>) of time window",
	},
	"Job": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(d.(*jobSummary).JobId), ctx)
		},
		Help: "Job ID",
	},
	"User": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatString(d.(*jobSummary).User.String(), ctx)
		},
		Help: "Name of user running the job",
	},
	"Duration": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatDurationValue(int64(d.(*jobSummary).Duration), ctx)
		},
		Help: "Duration of job: time of last observation minus time of first",
	},
	"Start": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatDateTimeValue(int64(d.(*jobSummary).Start), ctx)
		},
		Help: "Time of first observation",
	},
	"End": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatDateTimeValue(int64(d.(*jobSummary).End), ctx)
		},
		Help: "Time of last observation",
	},
	"CpuAvgPct": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(math.Ceil(d.(*jobSummary).computed[kCpuPctAvg])), ctx)
		},
		Help: "Average CPU utilization in percent (100% = 1 core)",
	},
	"CpuPeakPct": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(math.Ceil(d.(*jobSummary).computed[kCpuPctPeak])), ctx)
		},
		Help: "Peak CPU utilization in percent (100% = 1 core)",
	},
	"RelativeCpuAvgPct": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(math.Ceil(d.(*jobSummary).computed[kRcpuPctAvg])), ctx)
		},
		Help: "Average relative CPU utilization in percent (100% = all cores)",
	},
	"RelativeCpuPeakPct": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(math.Ceil(d.(*jobSummary).computed[kRcpuPctPeak])), ctx)
		},
		Help: "Peak relative CPU utilization in percent (100% = all cores)",
	},
	"MemAvgGB": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(math.Ceil(d.(*jobSummary).computed[kCpuGBAvg])), ctx)
		},
		Help: "Average main virtual memory utilization in GB",
	},
	"MemPeakGB": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(math.Ceil(d.(*jobSummary).computed[kCpuGBPeak])), ctx)
		},
		Help: "Peak main virtual memory utilization in GB",
	},
	"RelativeMemAvgPct": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(math.Ceil(d.(*jobSummary).computed[kRcpuGBAvg])), ctx)
		},
		Help: "Average relative main virtual memory utilization in percent (100% = system RAM)",
	},
	"RelativeMemPeakPct": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(uint64(math.Ceil(d.(*jobSummary).computed[kRcpuGBPeak])), ctx)
		},
		Help: "Peak relative main virtual memory utilization in percent (100% = system RAM)",
	},
	"ResidentMemAvgGB": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(math.Ceil(d.(*jobSummary).computed[kRssAnonGBAvg])), ctx)
		},
		Help: "Average main resident memory utilization in GB",
	},
	"ResidentMemPeakGB": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(uint64(math.Ceil(d.(*jobSummary).computed[kRssAnonGBPeak])), ctx)
		},
		Help: "Peak main resident memory utilization in GB",
	},
	"RelativeResidentMemAvgPct": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(math.Ceil(d.(*jobSummary).computed[kRrssAnonGBAvg])), ctx)
		},
		Help: "Average relative main resident memory utilization in percent (100% = all RAM)",
	},
	"RelativeResidentMemPeakPct": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(math.Ceil(d.(*jobSummary).computed[kRrssAnonGBPeak])), ctx)
		},
		Help: "Peak relative main resident memory utilization in percent (100% = all RAM)",
	},
	"GpuAvgPct": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(math.Ceil(d.(*jobSummary).computed[kGpuPctAvg])), ctx)
		},
		Help: "Average GPU utilization in percent (100% = 1 card)",
	},
	"GpuPeakPct": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(math.Ceil(d.(*jobSummary).computed[kGpuPctPeak])), ctx)
		},
		Help: "Peak GPU utilization in percent (100% = 1 card)",
	},
	"RelativeGpuAvgPct": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(math.Ceil(d.(*jobSummary).computed[kRgpuPctAvg])), ctx)
		},
		Help: "Average relative GPU utilization in percent (100% = all cards)",
	},
	"RelativeGpuPeakPct": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(math.Ceil(d.(*jobSummary).computed[kRgpuPctPeak])), ctx)
		},
		Help: "Peak relative GPU utilization in percent (100% = all cards)",
	},
	"OccupiedRelativeGpuAvgPct": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(math.Ceil(d.(*jobSummary).computed[kSgpuPctAvg])), ctx)
		},
		Help: "Average relative GPU utilization in percent (100% = all cards used by job)",
	},
	"OccupiedRelativeGpuPeakPct": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(math.Ceil(d.(*jobSummary).computed[kSgpuPctPeak])), ctx)
		},
		Help: "Peak relative GPU utilization in percent (100% = all cards used by job)",
	},
	"GpuMemAvgGB": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(math.Ceil(d.(*jobSummary).computed[kGpuGBAvg])), ctx)
		},
		Help: "Average resident GPU memory utilization in GB",
	},
	"GpuMemPeakGB": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(math.Ceil(d.(*jobSummary).computed[kGpuGBPeak])), ctx)
		},
		Help: "Peak resident GPU memory utilization in GB",
	},
	"RelativeGpuMemAvgPct": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(math.Ceil(d.(*jobSummary).computed[kRgpuGBAvg])), ctx)
		},
		Help: "Average relative GPU resident memory utilization in percent (100% = all GPU RAM)",
	},
	"RelativeGpuMemPeakPct": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(math.Ceil(d.(*jobSummary).computed[kRgpuGBPeak])), ctx)
		},
		Help: "Peak relative GPU resident memory utilization in percent (100% = all GPU RAM)",
	},
	"OccupiedRelativeGpuMemAvgPct": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(math.Ceil(d.(*jobSummary).computed[kSgpuGBAvg])), ctx)
		},
		Help: "Average relative GPU resident memory utilization in percent (100% = all GPU RAM on cards used by job)",
	},
	"OccupiedRelativeGpuMemPeakPct": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(math.Ceil(d.(*jobSummary).computed[kSgpuGBPeak])), ctx)
		},
		Help: "Peak relative GPU resident memory utilization in percent (100% = all GPU RAM on cards used by job)",
	},
	"Gpus": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatGpuSet(d.(*jobSummary).Gpus, ctx)
		},
		Help: "GPU device numbers used by the job, 'none' if none or 'unknown' in error states",
	},
	"GpuFail": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(d.(*jobSummary).GpuFail), ctx)
		},
		Help: "Flag indicating GPU status (0=Ok, 1=Failing)",
	},
	"Cmd": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatString(d.(*jobSummary).Cmd, ctx)
		},
		Help: "The commands invoking the processes of the job",
	},
	"Host": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatString(d.(*jobSummary).Host, ctx)
		},
		Help: "List of the host name(s) running the job (first elements of FQDNs, compressed)",
	},
	"Now": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatDateTimeValue(int64(d.(*jobSummary).Now), ctx)
		},
		Help: "The current time",
	},
	"Classification": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatInt64(int64(d.(*jobSummary).Classification), ctx)
		},
		Help: "Bit vector of live-at-start (2) and live-at-end (1) flags",
	},
	"CpuTime": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatDurationValue(int64(d.(*jobSummary).CpuTime), ctx)
		},
		Help: "Total CPU time of the job across all cores",
	},
	"GpuTime": {
		Fmt: func(d any, ctx PrintMods) string {
			return FormatDurationValue(int64(d.(*jobSummary).GpuTime), ctx)
		},
		Help: "Total GPU time of the job across all cards",
	},
}

func init() {
	// Define aliases for the traditional names
	DefAlias(jobsFormatters, "JobAndMark", "jobm")
	DefAlias(jobsFormatters, "Job", "job")
	DefAlias(jobsFormatters, "User", "user")
	DefAlias(jobsFormatters, "Duration", "duration")
	DefAlias(jobsFormatters, "Start", "start")
	DefAlias(jobsFormatters, "End", "end")
	DefAlias(jobsFormatters, "CpuAvgPct", "cpu-avg")
	DefAlias(jobsFormatters, "CpuPeakPct", "cpu-peak")
	DefAlias(jobsFormatters, "RelativeCpuAvgPct", "rcpu-avg")
	DefAlias(jobsFormatters, "RelativeCpuPeakPct", "rcpu-peak")
	DefAlias(jobsFormatters, "MemAvgGB", "mem-avg")
	DefAlias(jobsFormatters, "MemPeakGB", "mem-peak")
	DefAlias(jobsFormatters, "RelativeMemAvgPct", "rmem-avg")
	DefAlias(jobsFormatters, "RelativeMemPeakPct", "rmem-peak")
	DefAlias(jobsFormatters, "ResidentMemAvgGB", "res-avg")
	DefAlias(jobsFormatters, "ResidentMemPeakGB", "res-peak")
	DefAlias(jobsFormatters, "RelativeResidentMemAvgPct", "rres-avg")
	DefAlias(jobsFormatters, "RelativeResidentMemPeakPct", "rres-peak")
	DefAlias(jobsFormatters, "GpuAvgPct", "gpu-avg")
	DefAlias(jobsFormatters, "GpuPeakPct", "gpu-peak")
	DefAlias(jobsFormatters, "RelativeGpuAvgPct", "rgpu-avg")
	DefAlias(jobsFormatters, "RelativeGpuPeakPct", "rgpu-peak")
	DefAlias(jobsFormatters, "OccupiedRelativeGpuAvgPct", "sgpu-avg")
	DefAlias(jobsFormatters, "OccupiedRelativeGpuPeakPct", "sgpu-peak")
	DefAlias(jobsFormatters, "GpuMemAvgGB", "gpumem-avg")
	DefAlias(jobsFormatters, "GpuMemPeakGB", "gpumem-peak")
	DefAlias(jobsFormatters, "RelativeGpuMemAvgPct", "rgpumem-avg")
	DefAlias(jobsFormatters, "RelativeGpuMemPeakPct", "rgpumem-peak")
	DefAlias(jobsFormatters, "OccupiedRelativeGpuMemAvgPct", "sgpumem-avg")
	DefAlias(jobsFormatters, "OccupiedRelativeGpuMemPeakPct", "sgpumem-peak")
	DefAlias(jobsFormatters, "Gpus", "gpus")
	DefAlias(jobsFormatters, "GpuFail", "gpufail")
	DefAlias(jobsFormatters, "Cmd", "cmd")
	DefAlias(jobsFormatters, "Host", "host")
	DefAlias(jobsFormatters, "Now", "now")
	DefAlias(jobsFormatters, "Classification", "classification")
	DefAlias(jobsFormatters, "CpuTime", "cputime")
	DefAlias(jobsFormatters, "GpuTime", "gputime")
}

// MT: Constant after initialization; immutable
var jobsAliases = map[string][]string{
	// Traditional names
	"default": strings.Split(v0JobsDefaultFields, ","),
	"all": []string{
		"jobm", "job", "user", "duration", "duration/sec", "start", "start/sec", "end", "end/sec",
		"cpu-avg", "cpu-peak", "rcpu-avg", "rcpu-peak", "mem-avg", "mem-peak", "rmem-avg",
		"rmem-peak", "res-avg", "res-peak", "rres-avg", "rres-peak", "gpu-avg", "gpu-peak",
		"rgpu-avg", "rgpu-peak", "sgpu-avg", "sgpu-peak", "gpumem-avg", "gpumem-peak",
		"rgpumem-avg", "rgpumem-peak", "sgpumem-avg", "sgpumem-peak", "gpus", "gpufail",
		"cmd", "host", "now", "now/sec", "classification", "cputime/sec", "cputime",
		"gputime/sec", "gputime",
	},
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

	// New names
	"Default": strings.Split(v1JobsDefaultFields, ","),
	"All": []string{
		"JobAndMark", "Job", "User", "Duration", "Duration/sec", "Start", "Start/sec", "End",
		"End/sec", "CpuAvgPct", "CpuPeakPct", "RelativeCpuAvgPct", "RelativeCpuPeakPct", "MemAvgGB",
		"MemPeakGB", "RelativeMemAvgPct", "RelativeMemPeakPct", "ResidentMemAvgGB",
		"ResidentMemPeakGB", "RelativeResidentMemAvgPct", "RelativeResidentMemPeakPct",
		"GpuAvgPct", "GpuPeakPct", "RelativeGpuAvgPct", "RelativeGpuPeakPct",
		"OccupiedRelativeGpuAvgPct", "OccupiedRelativeGpuPeakPct", "GpuMemAvgGB",
		"GpuMemPeakGB", "RelativeGpuMemAvgPct", "RelativeGpuMemPeakPct",
		"OccupiedRelativeGpuMemAvgPct", "OccupiedRelativeGpuMemPeakPct", "Gpus", "GpuFail",
		"Cmd", "Host", "Now", "Now/sec", "Classification", "CpuTime/sec", "CpuTime",
		"GpuTime/sec", "GpuTime",
	},
	"Std":                    []string{"JobAndMark", "User", "Duration", "Host"},
	"Cpu":                    []string{"CpuAvgPct", "CpuPeakPct"},
	"RelativeCpu":            []string{"RelativeCpuAvgPct", "RelativeCpuPeakPct"},
	"Mem":                    []string{"MemAvgGB", "MemPeakGB"},
	"RelativeMem":            []string{"RelativeMemAvgPct", "RelativeMemPeakPct"},
	"ResidentMem":            []string{"ResidentMemAvgGB", "ResidentMemPeakGB"},
	"RelativeResidentMem":    []string{"RelativeResidentMemAvgPct", "RelativeResidentMemPeakPct"},
	"Gpu":                    []string{"GpuAvgPct", "GpuPeakPct"},
	"RelativeGpu":            []string{"RelativeGpuAvgPct", "RelativeGpuPeakPct"},
	"OccupiedRelativeGpu":    []string{"OccupiedRelativeGpuAvgPct", "OccupiedRelativeGpuPeakPct"},
	"GpuMem":                 []string{"GpuMemAvgPct", "GpuMemPeakPct"},
	"RelativeGpuMem":         []string{"RelativeGpuMemAvgPct", "RelativeGpuMemPeakPct"},
	"OccupiedRelativeGpuMem": []string{"OccupiedRelativeGpuMemAvgPct", "OccupiedRelativeGpuMemPeakPct"},
}
