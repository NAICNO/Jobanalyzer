package jobs

import (
	"cmp"
	"fmt"
	"io"
	"math"
	_ "reflect"
	"slices"

	uslices "go-utils/slices"

	. "sonalyze/command"
	. "sonalyze/common"
)

func (jc *JobsCommand) printRequiresConfig() bool {
	for _, f := range jc.PrintFields {
		switch f.Name {
		case "rcpu-avg", "rcpu-peak", "rmem-avg", "rmem-peak", "rgpu-avg", "rgpu-peak",
			"rgpumem-avg", "rgpumem-peak", "rres-avg", "rres-peak":
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

const jobsDefaultFields = "std,cpu,mem,gpu,gpumem,cmd"

// MT: Constant after initialization; immutable
var jobsAliases = map[string][]string{
	"default": []string{"jobm", "user", "duration", "host", "cpu", "mem", "gpu", "gpumem", "cmd"},
	"all":     []string{"jobm", "job", "user", "duration", "duration/sec", "start", "start/sec", "end", "end/sec", "cpu-avg", "cpu-peak", "rcpu-avg", "rcpu-peak", "mem-avg", "mem-peak", "rmem-avg", "rmem-peak", "res-avg", "res-peak", "rres-avg", "rres-peak", "gpu-avg", "gpu-peak", "rgpu-avg", "rgpu-peak", "sgpu-avg", "sgpu-peak", "gpumem-avg", "gpumem-peak", "rgpumem-avg", "rgpumem-peak", "sgpumem-avg", "sgpumem-peak", "gpus", "gpufail", "cmd", "host", "now", "now/sec", "classification", "cputime/sec", "cputime", "gputime/sec", "gputime"},
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

// TODO:
//  - indexed field access: key: XFA{desc, realname, indexval, attr} is exactly like synthesized ZFA
//    realname is an array or slice, indexval is an int index, key is the display name, there should be
//    only one realname[index] entry per key
//  - does the field need to be "Computed" and not "computed"?

type SFS = SimpleFormatSpec
type XFA = SynthesizedIndexedFormatSpecWithAttr

/*
var newJobsFormatters = DefineTableFromMap(
	reflect.TypeOf((*jobSummary)(nil)).Elem(),
	map[string]any{
		"JobAndMark":         SFS{"Job ID with mark indicating job running at start+end (!), start (<), or end (>) of time window", "jobm"},
		"JobId":              SFS{"Job ID", "job"},
		"User":               SFS{"Name of user running the job", "user"},
		"Duration":           SFS{"Duration in minutes of job: time of last observation minus time of first", "duration"},
		"Start":              SFS{"Time of first observation", "start"},
		"End":                SFS{"Time of last observation", "end"},
		"CpuAvgPct":          XFA{"Average CPU utilization in percent (100% = 1 core)", "computed", kCpuPctAvg, "cpu-avg", FmtCeil},
		"CpuPeakPct":         XFA{"Peak CPU utilization in percent (100% = 1 core)", "computed", kCpuPctPeak, "cpu-peak", FmtCeil},
		"RelativeCpuAvgPct":  XFA{"Average relative CPU utilization in percent (100% = all cores)", "computed", kRcpuPctAvg, "rcpu-avg", FmtCeil},
		"RelativeCpuPeakPct": XFA{"Peak relative CPU utilization in percent (100% = all cores)", "computed", kRcpuPctPeak, "rcpu-peak", FmtCeil},
		"MemAvgGB":           XFA{"Average main virtual memory utilization in GB", "computed", kCpuGBAvg, "mem-avg", FmtCeil},
		"MemPeakGB":          XFA{"Peak main virtual memory utilization in GB", "computed", kCpuGBPeak, "mem-peak", FmtCeil},
		//"RelativeMemAvgPct": XFA{},
		//"RelativeMemPeakPct": XFA{},
		// ... FIXME ...
		"Gpus":           SFS{"GPU device numbers used by the job, 'none' if none or 'unknown' in error states", "gpus"},
		"GpuFail":        SFS{"Flag indicating GPU status (0=Ok, not 0=failing)", "gpufail"},
		"Cmd":            SFS{"The commands invoking the processes of the job", "cmd"},
		"Host":           SFS{"List of the host name(s) running the job (first elements of FQDNs, compressed)", "host"},
		"Now":            SFS{"The current time", "now"},
		"Classification": SFS{"Bit vector of live-at-start (2) and live-at-end (1) flags", "classification"},
		"CpuTime":        SFS{"Total CPU time of the job across all cores", "cputime"},
		"GpuTime":        SFS{"Total GPU time of the job across all cores", "gputime"},
	},
)
*/

// MT: Constant after initialization; immutable
var jobsFormatters = map[string]Formatter{
	"jobm": {
		func(d any, _ PrintMods) string {
			return d.(*jobSummary).JobAndMark
		},
		"Job ID with mark indicating job running at start+end (!), start (<), or end (>) of time window",
	},
	"job": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(d.(*jobSummary).JobId)
		},
		"Job ID",
	},
	"user": {
		func(d any, _ PrintMods) string {
			return d.(*jobSummary).User.String()
		},
		"Name of user running the job",
	},
	"duration": {
		func(d any, ctx PrintMods) string {
			return FormatDurationValue(int64(d.(*jobSummary).Duration), ctx)
		},
		"Duration of job: time of last observation minus time of first",
	},
	"start": {
		func(d any, ctx PrintMods) string {
			return FormatDateTimeValue(int64(d.(*jobSummary).Start), ctx)
		},
		"Time of first observation",
	},
	"end": {
		func(d any, ctx PrintMods) string {
			return FormatDateTimeValue(int64(d.(*jobSummary).End), ctx)
		},
		"Time of last observation",
	},
	"cpu-avg": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(uint64(math.Ceil(d.(*jobSummary).computed[kCpuPctAvg])))
		},
		"Average CPU utilization in percent (100% = 1 core)",
	},
	"cpu-peak": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(uint64(math.Ceil(d.(*jobSummary).computed[kCpuPctPeak])))
		},
		"Peak CPU utilization in percent (100% = 1 core)",
	},
	"rcpu-avg": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(uint64(math.Ceil(d.(*jobSummary).computed[kRcpuPctAvg])))
		},
		"Average relative CPU utilization in percent (100% = all cores)",
	},
	"rcpu-peak": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(uint64(math.Ceil(d.(*jobSummary).computed[kRcpuPctPeak])))
		},
		"Peak relative CPU utilization in percent (100% = all cores)",
	},
	"mem-avg": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(uint64(math.Ceil(d.(*jobSummary).computed[kCpuGBAvg])))
		},
		"Average main virtual memory utilization in GiB",
	},
	"mem-peak": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(uint64(math.Ceil(d.(*jobSummary).computed[kCpuGBPeak])))
		},
		"Peak main virtual memory utilization in GiB",
	},
	"rmem-avg": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(uint64(math.Ceil(d.(*jobSummary).computed[kRcpuGBAvg])))
		},
		"Average relative main virtual memory utilization in percent (100% = system RAM)",
	},
	"rmem-peak": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(uint64(math.Ceil(d.(*jobSummary).computed[kRcpuGBPeak])))
		},
		"Peak relative main virtual memory utilization in percent (100% = system RAM)",
	},
	"res-avg": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(uint64(math.Ceil(d.(*jobSummary).computed[kRssAnonGBAvg])))
		},
		"Average main resident memory utilization in GiB",
	},
	"res-peak": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(uint64(math.Ceil(d.(*jobSummary).computed[kRssAnonGBPeak])))
		},
		"Peak main resident memory utilization in GiB",
	},
	"rres-avg": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(uint64(math.Ceil(d.(*jobSummary).computed[kRrssAnonGBAvg])))
		},
		"Average relative main resident memory utilization in percent (100% = all RAM)",
	},
	"rres-peak": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(uint64(math.Ceil(d.(*jobSummary).computed[kRrssAnonGBPeak])))
		},
		"Peak relative main resident memory utilization in percent (100% = all RAM)",
	},
	"gpu-avg": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(uint64(math.Ceil(d.(*jobSummary).computed[kGpuPctAvg])))
		},
		"Average GPU utilization in percent (100% = 1 card)",
	},
	"gpu-peak": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(uint64(math.Ceil(d.(*jobSummary).computed[kGpuPctPeak])))
		},
		"Peak GPU utilization in percent (100% = 1 card)",
	},
	"rgpu-avg": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(uint64(math.Ceil(d.(*jobSummary).computed[kRgpuPctAvg])))
		},
		"Average relative GPU utilization in percent (100% = all cards)",
	},
	"rgpu-peak": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(uint64(math.Ceil(d.(*jobSummary).computed[kRgpuPctPeak])))
		},
		"Peak relative GPU utilization in percent (100% = all cards)",
	},
	"sgpu-avg": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(uint64(math.Ceil(d.(*jobSummary).computed[kSgpuPctAvg])))
		},
		"Average relative GPU utilization in percent (100% = all cards used by job)",
	},
	"sgpu-peak": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(uint64(math.Ceil(d.(*jobSummary).computed[kSgpuPctPeak])))
		},
		"Peak relative GPU utilization in percent (100% = all cards used by job)",
	},
	"gpumem-avg": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(uint64(math.Ceil(d.(*jobSummary).computed[kGpuGBAvg])))
		},
		"Average resident GPU memory utilization in GiB",
	},
	"gpumem-peak": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(uint64(math.Ceil(d.(*jobSummary).computed[kGpuGBPeak])))
		},
		"Peak resident GPU memory utilization in GiB",
	},
	"rgpumem-avg": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(uint64(math.Ceil(d.(*jobSummary).computed[kRgpuGBAvg])))
		},
		"Average relative GPU resident memory utilization in percent (100% = all GPU RAM)",
	},
	"rgpumem-peak": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(uint64(math.Ceil(d.(*jobSummary).computed[kRgpuGBPeak])))
		},
		"Peak relative GPU resident memory utilization in percent (100% = all GPU RAM)",
	},
	"sgpumem-avg": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(uint64(math.Ceil(d.(*jobSummary).computed[kSgpuGBAvg])))
		},
		"Average relative GPU resident memory utilization in percent (100% = all GPU RAM on cards used by job)",
	},
	"sgpumem-peak": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(uint64(math.Ceil(d.(*jobSummary).computed[kSgpuGBPeak])))
		},
		"Peak relative GPU resident memory utilization in percent (100% = all GPU RAM on cards used by job)",
	},
	"gpus": {
		func(d any, _ PrintMods) string {
			return d.(*jobSummary).Gpus.String()
		},
		"GPU device numbers used by the job, 'none' if none or 'unknown' in error states",
	},
	"gpufail": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(d.(*jobSummary).GpuFail)
		},
		"Flag indicating GPU status (0=Ok, 1=Failing)",
	},
	"cmd": {
		func(d any, _ PrintMods) string {
			return d.(*jobSummary).Cmd
		},
		"The commands invoking the processes of the job",
	},
	"host": {
		func(d any, _ PrintMods) string {
			return d.(*jobSummary).Host
		},
		"List of the host name(s) running the job (first elements of FQDNs, compressed)",
	},
	"now": {
		func(d any, ctx PrintMods) string {
			return FormatDateTimeValue(int64(d.(*jobSummary).Now), ctx)
		},
		"The current time",
	},
	"classification": {
		func(d any, _ PrintMods) string {
			return fmt.Sprint(d.(*jobSummary).Classification)
		},
		"Bit vector of live-at-start (2) and live-at-end (1) flags",
	},
	"cputime": {
		func(d any, ctx PrintMods) string {
			return FormatDurationValue(int64(d.(*jobSummary).CpuTime), ctx)
		},
		"Total CPU time of the job across all cores",
	},
	"gputime": {
		func(d any, ctx PrintMods) string {
			return FormatDurationValue(int64(d.(*jobSummary).GpuTime), ctx)
		},
		"Total GPU time of the job across all cards",
	},
}
