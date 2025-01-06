// DO NOT EDIT.  Generated from print.go by generate-table

package jobs

import (
	"cmp"
	"fmt"
	"io"
	. "sonalyze/common"
	. "sonalyze/table"
)

var (
	_ = cmp.Compare(0, 0)
	_ fmt.Formatter
	_ = io.SeekStart
	_ = UstrEmpty
)

// MT: Constant after initialization; immutable
var jobsFormatters = map[string]Formatter[*jobSummary]{
	"JobAndMark": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatString(d.JobAndMark, ctx)
		},
		Help: "(string) Job ID with mark indicating job running at start+end (!), start (<), or end (>) of time window",
	},
	"Job": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatUint32(d.JobId, ctx)
		},
		Help: "(uint32) Job ID",
	},
	"User": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatUstr(d.User, ctx)
		},
		Help: "(string) Name of user running the job",
	},
	"Duration": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatDurationValue(d.Duration, ctx)
		},
		Help: "(DurationValue) Time of last observation minus time of first",
	},
	"Start": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatDateTimeValue(d.Start, ctx)
		},
		Help: "(DateTimeValue) Time of first observation",
	},
	"End": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatDateTimeValue(d.End, ctx)
		},
		Help: "(DateTimeValue) Time of last observation",
	},
	"CpuAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil(d.computed[kCpuPctAvg], ctx)
		},
		Help: "(int) Average CPU utilization in percent (100% = 1 core)",
	},
	"CpuPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil(d.computed[kCpuPctPeak], ctx)
		},
		Help: "(int) Peak CPU utilization in percent (100% = 1 core)",
	},
	"RelativeCpuAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil(d.computed[kRcpuPctAvg], ctx)
		},
		Help: "(int) Average relative CPU utilization in percent (100% = all cores)",
	},
	"RelativeCpuPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil(d.computed[kRcpuPctPeak], ctx)
		},
		Help: "(int) Peak relative CPU utilization in percent (100% = all cores)",
	},
	"MemAvgGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil(d.computed[kCpuGBAvg], ctx)
		},
		Help: "(int) Average main virtual memory utilization in GB",
	},
	"MemPeakGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil(d.computed[kCpuGBPeak], ctx)
		},
		Help: "(int) Peak main virtual memory utilization in GB",
	},
	"RelativeMemAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil(d.computed[kRcpuGBAvg], ctx)
		},
		Help: "(int) Average relative main virtual memory utilization in percent (100% = system RAM)",
	},
	"RelativeMemPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil(d.computed[kRcpuGBPeak], ctx)
		},
		Help: "(int) Peak relative main virtual memory utilization in percent (100% = system RAM)",
	},
	"ResidentMemAvgGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil(d.computed[kRssAnonGBAvg], ctx)
		},
		Help: "(int) Average main resident memory utilization in GB",
	},
	"ResidentMemPeakGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil(d.computed[kRssAnonGBPeak], ctx)
		},
		Help: "(int) Peak main resident memory utilization in GB",
	},
	"RelativeResidentMemAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil(d.computed[kRrssAnonGBAvg], ctx)
		},
		Help: "(int) Average relative main resident memory utilization in percent (100% = all RAM)",
	},
	"RelativeResidentMemPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil(d.computed[kRrssAnonGBPeak], ctx)
		},
		Help: "(int) Peak relative main resident memory utilization in percent (100% = all RAM)",
	},
	"GpuAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil(d.computed[kGpuPctAvg], ctx)
		},
		Help: "(int) Average GPU utilization in percent (100% = 1 card)",
	},
	"GpuPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil(d.computed[kGpuPctPeak], ctx)
		},
		Help: "(int) Peak GPU utilization in percent (100% = 1 card)",
	},
	"RelativeGpuAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil(d.computed[kRgpuPctAvg], ctx)
		},
		Help: "(int) Average relative GPU utilization in percent (100% = all cards)",
	},
	"RelativeGpuPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil(d.computed[kRgpuPctPeak], ctx)
		},
		Help: "(int) Peak relative GPU utilization in percent (100% = all cards)",
	},
	"OccupiedRelativeGpuAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil(d.computed[kSgpuPctAvg], ctx)
		},
		Help: "(int) Average relative GPU utilization in percent (100% = all cards used by job)",
	},
	"OccupiedRelativeGpuPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil(d.computed[kSgpuPctPeak], ctx)
		},
		Help: "(int) Peak relative GPU utilization in percent (100% = all cards used by job)",
	},
	"GpuMemAvgGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil(d.computed[kGpuGBAvg], ctx)
		},
		Help: "(int) Average resident GPU memory utilization in GB",
	},
	"GpuMemPeakGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil(d.computed[kGpuGBPeak], ctx)
		},
		Help: "(int) Peak resident GPU memory utilization in GB",
	},
	"RelativeGpuMemAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil(d.computed[kRgpuGBAvg], ctx)
		},
		Help: "(int) Average relative GPU resident memory utilization in percent (100% = all GPU RAM)",
	},
	"RelativeGpuMemPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil(d.computed[kRgpuGBPeak], ctx)
		},
		Help: "(int) Peak relative GPU resident memory utilization in percent (100% = all GPU RAM)",
	},
	"OccupiedRelativeGpuMemAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil(d.computed[kSgpuGBAvg], ctx)
		},
		Help: "(int) Average relative GPU resident memory utilization in percent (100% = all GPU RAM on cards used by job)",
	},
	"OccupiedRelativeGpuMemPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil(d.computed[kSgpuGBPeak], ctx)
		},
		Help: "(int) Peak relative GPU resident memory utilization in percent (100% = all GPU RAM on cards used by job)",
	},
	"Gpus": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatGpuSet(d.Gpus, ctx)
		},
		Help: "(GpuSet) GPU device numbers used by the job, 'none' if none or 'unknown' in error states",
	},
	"GpuFail": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatInt(d.GpuFail, ctx)
		},
		Help: "(int) Flag indicating GPU status (0=Ok, 1=Failing)",
	},
	"Cmd": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatString(d.Cmd, ctx)
		},
		Help: "(string) The commands invoking the processes of the job",
	},
	"Host": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatString(d.Host, ctx)
		},
		Help: "(string) List of the host name(s) running the job (first elements of FQDNs, compressed)",
	},
	"Now": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatDateTimeValue(d.Now, ctx)
		},
		Help: "(DateTimeValue) The current time",
	},
	"Classification": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatInt(d.Classification, ctx)
		},
		Help: "(int) Bit vector of live-at-start (2) and live-at-end (1) flags",
	},
	"CpuTime": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatDurationValue(d.CpuTime, ctx)
		},
		Help: "(DurationValue) Total CPU time of the job across all cores",
	},
	"GpuTime": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatDurationValue(d.GpuTime, ctx)
		},
		Help: "(DurationValue) Total GPU time of the job across all cards",
	},
	"Submit": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if d.sacctInfo != nil {
				return FormatDateTimeValue(d.sacctInfo.Submit, ctx)
			}
			return "?"
		},
		Help: "(DateTimeValue) Submit time of job (Slurm)",
	},
	"JobName": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if d.sacctInfo != nil {
				return FormatUstr(d.sacctInfo.JobName, ctx)
			}
			return "?"
		},
		Help: "(string) Name of job (Slurm)",
	},
	"State": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if d.sacctInfo != nil {
				return FormatUstr(d.sacctInfo.State, ctx)
			}
			return "?"
		},
		Help: "(string) Completion state of job (Slurm)",
	},
	"Account": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if d.sacctInfo != nil {
				return FormatUstr(d.sacctInfo.Account, ctx)
			}
			return "?"
		},
		Help: "(string) Name of job's account (Slurm)",
	},
	"Layout": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if d.sacctInfo != nil {
				return FormatUstr(d.sacctInfo.Layout, ctx)
			}
			return "?"
		},
		Help: "(string) Layout spec of job (Slurm)",
	},
	"Reservation": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if d.sacctInfo != nil {
				return FormatUstr(d.sacctInfo.Reservation, ctx)
			}
			return "?"
		},
		Help: "(string) Name of job's reservation (Slurm)",
	},
	"Partition": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if d.sacctInfo != nil {
				return FormatUstr(d.sacctInfo.Partition, ctx)
			}
			return "?"
		},
		Help: "(string) Partition of job (Slurm)",
	},
	"RequestedGpus": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if d.sacctInfo != nil {
				return FormatUstr(d.sacctInfo.ReqGPUS, ctx)
			}
			return "?"
		},
		Help: "(string) Names of requested GPUs (Slurm AllocTRES)",
	},
	"DiskReadAvgGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if d.sacctInfo != nil {
				return FormatUint32(d.sacctInfo.AveDiskRead, ctx)
			}
			return "?"
		},
		Help: "(uint32) Average disk read activity in GB/s (Slurm AveDiskRead)",
	},
	"DiskWriteAvgGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if d.sacctInfo != nil {
				return FormatUint32(d.sacctInfo.AveDiskWrite, ctx)
			}
			return "?"
		},
		Help: "(uint32) Average disk write activity in GB/s (Slurm AveDiskWrite)",
	},
	"RequestedCpus": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if d.sacctInfo != nil {
				return FormatUint32(d.sacctInfo.ReqCPUS, ctx)
			}
			return "?"
		},
		Help: "(uint32) Number of requested CPUs (Slurm)",
	},
	"RequestedMemGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if d.sacctInfo != nil {
				return FormatUint32(d.sacctInfo.ReqMem, ctx)
			}
			return "?"
		},
		Help: "(uint32) Requested memory (Slurm)",
	},
	"RequestedNodes": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if d.sacctInfo != nil {
				return FormatUint32(d.sacctInfo.ReqNodes, ctx)
			}
			return "?"
		},
		Help: "(uint32) Number of requested nodes (Slurm)",
	},
	"TimeLimit": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if d.sacctInfo != nil {
				return FormatU32Duration(d.sacctInfo.TimelimitRaw, ctx)
			}
			return "?"
		},
		Help: "(U32Duration) Elapsed time limit (Slurm)",
	},
	"ExitCode": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if d.sacctInfo != nil {
				return FormatUint8(d.sacctInfo.ExitCode, ctx)
			}
			return "?"
		},
		Help: "(uint8) Exit code of job (Slurm)",
	},
}

func init() {
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

func (c *JobsCommand) Summary(out io.Writer) {
	fmt.Fprint(out, `Display jobs jobs aggregated from process samples.

A "job" is aggregated from sample streams from one or more processes on one
or more nodes of a cluster.  On some clusters, jobs have clearly defined job
numbers (provided by a batch system such as Slurm), while on other clusters,
the job numbers are inferred from the process tree.

As jobs are built from samples, the job data can be noisy and may sometimes
not represent true behavior.  This is especially true for short jobs.

Note also:

  - A job can be selected by job number, but a time window must be selected
    that contains the job or the job will not be found

  - By default, only the jobs for the current user's user name are selected,
    specify "-user -" to see all users
`)
}

const jobsHelp = `
jobs
  Aggregate process data into data about "jobs" and present them.  Output
  records are sorted in order of increasing start time of the job. The default
  format is 'fixed'.
`

func (c *JobsCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(c.Fmt, jobsHelp, jobsFormatters, jobsAliases, jobsDefaultFields)
}

// MT: Constant after initialization; immutable
var jobsAliases = map[string][]string{
	"all":                    []string{"jobm", "job", "user", "duration", "duration/sec", "start", "start/sec", "end", "end/sec", "cpu-avg", "cpu-peak", "rcpu-avg", "rcpu-peak", "mem-avg", "mem-peak", "rmem-avg", "rmem-peak", "res-avg", "res-peak", "rres-avg", "rres-peak", "gpu-avg", "gpu-peak", "rgpu-avg", "rgpu-peak", "sgpu-avg", "sgpu-peak", "gpumem-avg", "gpumem-peak", "rgpumem-avg", "rgpumem-peak", "sgpumem-avg", "sgpumem-peak", "gpus", "gpufail", "cmd", "host", "now", "now/sec", "classification", "cputime/sec", "cputime", "gputime/sec", "gputime"},
	"std":                    []string{"jobm", "user", "duration", "host"},
	"cpu":                    []string{"cpu-avg", "cpu-peak"},
	"rcpu":                   []string{"rcpu-avg", "rcpu-peak"},
	"mem":                    []string{"mem-avg", "mem-peak"},
	"rmem":                   []string{"rmem-avg", "rmem-peak"},
	"res":                    []string{"res-avg", "res-peak"},
	"rres":                   []string{"rres-avg", "rres-peak"},
	"gpu":                    []string{"gpu-avg", "gpu-peak"},
	"rgpu":                   []string{"rgpu-avg", "rgpu-peak"},
	"sgpu":                   []string{"sgpu-avg", "sgpu-peak"},
	"gpumem":                 []string{"gpumem-avg", "gpumem-peak"},
	"rgpumem":                []string{"rgpumem-avg", "rgpumem-peak"},
	"sgpumem":                []string{"sgpumem-avg", "sgpumem-peak"},
	"All":                    []string{"JobAndMark", "Job", "User", "Duration", "Duration/sec", "Start", "Start/sec", "End", "End/sec", "CpuAvgPct", "CpuPeakPct", "RelativeCpuAvgPct", "RelativeCpuPeakPct", "MemAvgGB", "MemPeakGB", "RelativeMemAvgPct", "RelativeMemPeakPct", "ResidentMemAvgGB", "ResidentMemPeakGB", "RelativeResidentMemAvgPct", "RelativeResidentMemPeakPct", "GpuAvgPct", "GpuPeakPct", "RelativeGpuAvgPct", "RelativeGpuPeakPct", "OccupiedRelativeGpuAvgPct", "OccupiedRelativeGpuPeakPct", "GpuMemAvgGB", "GpuMemPeakGB", "RelativeGpuMemAvgPct", "RelativeGpuMemPeakPct", "OccupiedRelativeGpuMemAvgPct", "OccupiedRelativeGpuMemPeakPct", "Gpus", "GpuFail", "Cmd", "Host", "Now", "Now/sec", "Classification", "CpuTime/sec", "CpuTime", "GpuTime/sec", "GpuTime"},
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
	"default":                []string{"std", "cpu", "mem", "gpu", "gpumem", "cmd"},
	"Default":                []string{"Std", "Cpu", "Mem", "Gpu", "GpuMem", "Cmd"},
}

const jobsDefaultFields = "default"
