// DO NOT EDIT.  Generated from print.go by generate-table

package jobs

import "sonalyze/data/samplejob"

import (
	"cmp"
	"fmt"
	"go-utils/gpuset"
	"io"
	. "sonalyze/common"
	. "sonalyze/table"
)

var (
	_ = cmp.Compare(0, 0)
	_ fmt.Formatter
	_ = io.SeekStart
	_ = UstrEmpty
	_ gpuset.GpuSet
)

// MT: Constant after initialization; immutable
var jobsFormatters = map[string]Formatter[*jobSummary]{
	"JobAndMark": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatString((d.JobAndMark), ctx)
		},
		Help: "(string) Job ID with mark indicating job running at start+end (!), start (<), or end (>) of time window",
	},
	"Job": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sampleJob) != nil {
				return FormatUint32((d.sampleJob.JobId), ctx)
			}
			return "?"
		},
		Help: "(uint32) Job ID",
	},
	"User": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sampleJob) != nil {
				return FormatUstr((d.sampleJob.User), ctx)
			}
			return "?"
		},
		Help: "(string) Name of user running the job",
	},
	"Duration": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sampleJob) != nil {
				return FormatDurationValue((d.sampleJob.Duration), ctx)
			}
			return "?"
		},
		Help: "(DurationValue) Time of last observation minus time of first",
	},
	"Start": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sampleJob) != nil {
				return FormatDateTimeValue((d.sampleJob.Start), ctx)
			}
			return "?"
		},
		Help: "(DateTimeValue) Time of first observation",
	},
	"End": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sampleJob) != nil {
				return FormatDateTimeValue((d.sampleJob.End), ctx)
			}
			return "?"
		},
		Help: "(DateTimeValue) Time of last observation",
	},
	"CpuAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kCpuPctAvg]), ctx)
		},
		Help: "(int) Average CPU utilization in percent (100% = 1 core)",
	},
	"CpuPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kCpuPctPeak]), ctx)
		},
		Help: "(int) Peak CPU utilization in percent (100% = 1 core)",
	},
	"RelativeCpuAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kRcpuPctAvg]), ctx)
		},
		Help:        "(int) Average relative CPU utilization in percent (100% = all cores)",
		NeedsConfig: true,
	},
	"RelativeCpuPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kRcpuPctPeak]), ctx)
		},
		Help:        "(int) Peak relative CPU utilization in percent (100% = all cores)",
		NeedsConfig: true,
	},
	"MemAvgGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kCpuGBAvg]), ctx)
		},
		Help: "(int) Average main virtual memory utilization in GB",
	},
	"MemPeakGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kCpuGBPeak]), ctx)
		},
		Help: "(int) Peak main virtual memory utilization in GB",
	},
	"RelativeMemAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kRcpuGBAvg]), ctx)
		},
		Help:        "(int) Average relative main virtual memory utilization in percent (100% = system RAM)",
		NeedsConfig: true,
	},
	"RelativeMemPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kRcpuGBPeak]), ctx)
		},
		Help:        "(int) Peak relative main virtual memory utilization in percent (100% = system RAM)",
		NeedsConfig: true,
	},
	"ResidentMemAvgGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kRssAnonGBAvg]), ctx)
		},
		Help: "(int) Average main resident memory utilization in GB",
	},
	"ResidentMemPeakGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kRssAnonGBPeak]), ctx)
		},
		Help: "(int) Peak main resident memory utilization in GB",
	},
	"RelativeResidentMemAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kRrssAnonGBAvg]), ctx)
		},
		Help:        "(int) Average relative main resident memory utilization in percent (100% = all RAM)",
		NeedsConfig: true,
	},
	"RelativeResidentMemPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kRrssAnonGBPeak]), ctx)
		},
		Help:        "(int) Peak relative main resident memory utilization in percent (100% = all RAM)",
		NeedsConfig: true,
	},
	"GpuAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kGpuPctAvg]), ctx)
		},
		Help: "(int) Average GPU utilization in percent (100% = 1 card)",
	},
	"GpuPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kGpuPctPeak]), ctx)
		},
		Help: "(int) Peak GPU utilization in percent (100% = 1 card)",
	},
	"RelativeGpuAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kRgpuPctAvg]), ctx)
		},
		Help:        "(int) Average relative GPU utilization in percent (100% = all cards)",
		NeedsConfig: true,
	},
	"RelativeGpuPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kRgpuPctPeak]), ctx)
		},
		Help:        "(int) Peak relative GPU utilization in percent (100% = all cards)",
		NeedsConfig: true,
	},
	"OccupiedRelativeGpuAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kSgpuPctAvg]), ctx)
		},
		Help:        "(int) Average relative GPU utilization in percent (100% = all cards used by job)",
		NeedsConfig: true,
	},
	"OccupiedRelativeGpuPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kSgpuPctPeak]), ctx)
		},
		Help:        "(int) Peak relative GPU utilization in percent (100% = all cards used by job)",
		NeedsConfig: true,
	},
	"GpuMemAvgGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kGpuGBAvg]), ctx)
		},
		Help: "(int) Average resident GPU memory utilization in GB",
	},
	"GpuMemPeakGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kGpuGBPeak]), ctx)
		},
		Help: "(int) Peak resident GPU memory utilization in GB",
	},
	"RelativeGpuMemAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kRgpuGBAvg]), ctx)
		},
		Help:        "(int) Average relative GPU resident memory utilization in percent (100% = all GPU RAM)",
		NeedsConfig: true,
	},
	"RelativeGpuMemPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kRgpuGBPeak]), ctx)
		},
		Help:        "(int) Peak relative GPU resident memory utilization in percent (100% = all GPU RAM)",
		NeedsConfig: true,
	},
	"OccupiedRelativeGpuMemAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kSgpuGBAvg]), ctx)
		},
		Help:        "(int) Average relative GPU resident memory utilization in percent (100% = all GPU RAM on cards used by job)",
		NeedsConfig: true,
	},
	"OccupiedRelativeGpuMemPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kSgpuGBPeak]), ctx)
		},
		Help:        "(int) Peak relative GPU resident memory utilization in percent (100% = all GPU RAM on cards used by job)",
		NeedsConfig: true,
	},
	"Gpus": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sampleJob) != nil {
				return FormatGpuSet((d.sampleJob.Gpus), ctx)
			}
			return "?"
		},
		Help: "(GpuSet) GPU device numbers used by the job, 'none' if none or 'unknown' in error states",
	},
	"GpuFail": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sampleJob) != nil {
				return FormatInt((d.sampleJob.GpuFail), ctx)
			}
			return "?"
		},
		Help: "(int) Flag indicating GPU status (0=Ok, 1=Failing)",
	},
	"Cmd": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sampleJob) != nil {
				return FormatString((d.sampleJob.Cmd), ctx)
			}
			return "?"
		},
		Help: "(string) The commands invoking the processes of the job",
	},
	"Hosts": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sampleJob) != nil {
				return FormatHostnames((d.sampleJob.Hosts), ctx)
			}
			return "?"
		},
		Help: "(Hostnames) List of the host name(s) running the job",
	},
	"Now": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatDateTimeValue((d.Now), ctx)
		},
		Help: "(DateTimeValue) The current time",
	},
	"Classification": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sampleJob) != nil {
				return FormatInt((d.sampleJob.Classification), ctx)
			}
			return "?"
		},
		Help: "(int) Bit vector of live-at-start (2) and live-at-end (1) flags",
	},
	"CpuTime": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sampleJob) != nil {
				return FormatDurationValue((d.sampleJob.CpuTime), ctx)
			}
			return "?"
		},
		Help: "(DurationValue) Total CPU time of the job across all cores",
	},
	"GpuTime": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sampleJob) != nil {
				return FormatDurationValue((d.sampleJob.GpuTime), ctx)
			}
			return "?"
		},
		Help: "(DurationValue) Total GPU time of the job across all cards",
	},
	"SomeGpu": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sampleJob) != nil {
				return FormatBool((d.sampleJob.ComputedFlags&samplejob.KUsesGpu != 0), ctx)
			}
			return "?"
		},
		Help: "(bool) True iff process was seen to use some GPU",
	},
	"NoGpu": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sampleJob) != nil {
				return FormatBool((d.sampleJob.ComputedFlags&samplejob.KDoesNotUseGpu != 0), ctx)
			}
			return "?"
		},
		Help: "(bool) True iff process was seen to use no GPU",
	},
	"Running": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sampleJob) != nil {
				return FormatBool((d.sampleJob.ComputedFlags&samplejob.KIsLiveAtEnd != 0), ctx)
			}
			return "?"
		},
		Help: "(bool) True iff process appears to still be running at end of time window",
	},
	"Completed": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sampleJob) != nil {
				return FormatBool((d.sampleJob.ComputedFlags&samplejob.KIsNotLiveAtEnd != 0), ctx)
			}
			return "?"
		},
		Help: "(bool) True iff process appears not to be running at end of time window",
	},
	"Zombie": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sampleJob) != nil {
				return FormatBool((d.sampleJob.ComputedFlags&samplejob.KIsZombie != 0), ctx)
			}
			return "?"
		},
		Help: "(bool) True iff the process looks like a zombie",
	},
	"Primordial": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sampleJob) != nil {
				return FormatBool((d.sampleJob.ComputedFlags&samplejob.KIsLiveAtStart != 0), ctx)
			}
			return "?"
		},
		Help: "(bool) True iff the process appears to have been alive at the start of the time window",
	},
	"BornLater": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sampleJob) != nil {
				return FormatBool((d.sampleJob.ComputedFlags&samplejob.KIsNotLiveAtStart != 0), ctx)
			}
			return "?"
		},
		Help: "(bool) True iff the process appears not to have been alive at the start of the time window",
	},
	"Submit": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.slurmJob.Main) != nil {
				return FormatDateTimeValue((d.slurmJob.Main.Submit), ctx)
			}
			return "?"
		},
		Help: "(DateTimeValue) Submit time of job (Slurm)",
	},
	"JobName": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.slurmJob.Main) != nil {
				return FormatUstr((d.slurmJob.Main.JobName), ctx)
			}
			return "?"
		},
		Help: "(string) Name of job (Slurm)",
	},
	"State": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.slurmJob.Main) != nil {
				return FormatUstr((d.slurmJob.Main.State), ctx)
			}
			return "?"
		},
		Help: "(string) Completion state of job (Slurm)",
	},
	"Account": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.slurmJob.Main) != nil {
				return FormatUstr((d.slurmJob.Main.Account), ctx)
			}
			return "?"
		},
		Help: "(string) Name of job's account (Slurm)",
	},
	"Layout": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.slurmJob.Main) != nil {
				return FormatUstr((d.slurmJob.Main.Layout), ctx)
			}
			return "?"
		},
		Help: "(string) Layout spec of job (Slurm)",
	},
	"Reservation": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.slurmJob.Main) != nil {
				return FormatUstr((d.slurmJob.Main.Reservation), ctx)
			}
			return "?"
		},
		Help: "(string) Name of job's reservation (Slurm)",
	},
	"Partition": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.slurmJob.Main) != nil {
				return FormatUstr((d.slurmJob.Main.Partition), ctx)
			}
			return "?"
		},
		Help: "(string) Partition of job (Slurm)",
	},
	"RequestedGpus": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.slurmJob.Main) != nil {
				return FormatUstr((d.slurmJob.Main.ReqGPUS), ctx)
			}
			return "?"
		},
		Help: "(string) Names of requested GPUs (Slurm AllocTRES)",
	},
	"DiskReadAvgGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.slurmJob.Main) != nil {
				return FormatUint32((d.slurmJob.Main.AveDiskRead), ctx)
			}
			return "?"
		},
		Help: "(uint32) Average disk read activity in GB/s (Slurm AveDiskRead)",
	},
	"DiskWriteAvgGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.slurmJob.Main) != nil {
				return FormatUint32((d.slurmJob.Main.AveDiskWrite), ctx)
			}
			return "?"
		},
		Help: "(uint32) Average disk write activity in GB/s (Slurm AveDiskWrite)",
	},
	"RequestedCpus": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.slurmJob.Main) != nil {
				return FormatUint32((d.slurmJob.Main.ReqCPUS), ctx)
			}
			return "?"
		},
		Help: "(uint32) Number of requested CPUs (Slurm)",
	},
	"RequestedMemGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.slurmJob.Main) != nil {
				return FormatUint32((d.slurmJob.Main.ReqMem), ctx)
			}
			return "?"
		},
		Help: "(uint32) Requested memory (Slurm)",
	},
	"RequestedNodes": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.slurmJob.Main) != nil {
				return FormatUint32((d.slurmJob.Main.ReqNodes), ctx)
			}
			return "?"
		},
		Help: "(uint32) Number of requested nodes (Slurm)",
	},
	"TimeLimit": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.slurmJob.Main) != nil {
				return FormatU32Duration((d.slurmJob.Main.TimelimitRaw), ctx)
			}
			return "?"
		},
		Help: "(U32Duration) Elapsed time limit (Slurm)",
	},
	"ExitCode": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.slurmJob.Main) != nil {
				return FormatUint8((d.slurmJob.Main.ExitCode), ctx)
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
	DefAlias(jobsFormatters, "Hosts", "host")
	DefAlias(jobsFormatters, "Hosts", "hosts")
	DefAlias(jobsFormatters, "Now", "now")
	DefAlias(jobsFormatters, "Classification", "classification")
	DefAlias(jobsFormatters, "CpuTime", "cputime")
	DefAlias(jobsFormatters, "GpuTime", "gputime")
}

// MT: Constant after initialization; immutable
var jobsPredicates = map[string]Predicate[*jobSummary]{
	"JobAndMark": Predicate[*jobSummary]{
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.JobAndMark), v.(string))
		},
	},
	"Job": Predicate[*jobSummary]{
		Convert: CvtString2Uint32,
		Compare: func(d *jobSummary, v any) int {
			if (d.sampleJob) != nil {
				return cmp.Compare((d.sampleJob.JobId), v.(uint32))
			}
			return -1
		},
	},
	"User": Predicate[*jobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *jobSummary, v any) int {
			if (d.sampleJob) != nil {
				return cmp.Compare((d.sampleJob.User), v.(Ustr))
			}
			return -1
		},
	},
	"Duration": Predicate[*jobSummary]{
		Convert: CvtString2DurationValue,
		Compare: func(d *jobSummary, v any) int {
			if (d.sampleJob) != nil {
				return cmp.Compare((d.sampleJob.Duration), v.(DurationValue))
			}
			return -1
		},
	},
	"Start": Predicate[*jobSummary]{
		Convert: CvtString2DateTimeValue,
		Compare: func(d *jobSummary, v any) int {
			if (d.sampleJob) != nil {
				return cmp.Compare((d.sampleJob.Start), v.(DateTimeValue))
			}
			return -1
		},
	},
	"End": Predicate[*jobSummary]{
		Convert: CvtString2DateTimeValue,
		Compare: func(d *jobSummary, v any) int {
			if (d.sampleJob) != nil {
				return cmp.Compare((d.sampleJob.End), v.(DateTimeValue))
			}
			return -1
		},
	},
	"CpuAvgPct": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kCpuPctAvg]), v.(F64Ceil))
		},
	},
	"CpuPeakPct": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kCpuPctPeak]), v.(F64Ceil))
		},
	},
	"RelativeCpuAvgPct": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kRcpuPctAvg]), v.(F64Ceil))
		},
	},
	"RelativeCpuPeakPct": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kRcpuPctPeak]), v.(F64Ceil))
		},
	},
	"MemAvgGB": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kCpuGBAvg]), v.(F64Ceil))
		},
	},
	"MemPeakGB": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kCpuGBPeak]), v.(F64Ceil))
		},
	},
	"RelativeMemAvgPct": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kRcpuGBAvg]), v.(F64Ceil))
		},
	},
	"RelativeMemPeakPct": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kRcpuGBPeak]), v.(F64Ceil))
		},
	},
	"ResidentMemAvgGB": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kRssAnonGBAvg]), v.(F64Ceil))
		},
	},
	"ResidentMemPeakGB": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kRssAnonGBPeak]), v.(F64Ceil))
		},
	},
	"RelativeResidentMemAvgPct": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kRrssAnonGBAvg]), v.(F64Ceil))
		},
	},
	"RelativeResidentMemPeakPct": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kRrssAnonGBPeak]), v.(F64Ceil))
		},
	},
	"GpuAvgPct": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kGpuPctAvg]), v.(F64Ceil))
		},
	},
	"GpuPeakPct": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kGpuPctPeak]), v.(F64Ceil))
		},
	},
	"RelativeGpuAvgPct": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kRgpuPctAvg]), v.(F64Ceil))
		},
	},
	"RelativeGpuPeakPct": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kRgpuPctPeak]), v.(F64Ceil))
		},
	},
	"OccupiedRelativeGpuAvgPct": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kSgpuPctAvg]), v.(F64Ceil))
		},
	},
	"OccupiedRelativeGpuPeakPct": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kSgpuPctPeak]), v.(F64Ceil))
		},
	},
	"GpuMemAvgGB": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kGpuGBAvg]), v.(F64Ceil))
		},
	},
	"GpuMemPeakGB": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kGpuGBPeak]), v.(F64Ceil))
		},
	},
	"RelativeGpuMemAvgPct": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kRgpuGBAvg]), v.(F64Ceil))
		},
	},
	"RelativeGpuMemPeakPct": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kRgpuGBPeak]), v.(F64Ceil))
		},
	},
	"OccupiedRelativeGpuMemAvgPct": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kSgpuGBAvg]), v.(F64Ceil))
		},
	},
	"OccupiedRelativeGpuMemPeakPct": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kSgpuGBPeak]), v.(F64Ceil))
		},
	},
	"Gpus": Predicate[*jobSummary]{
		Convert: CvtString2GpuSet,
		SetCompare: func(d *jobSummary, v any, op int) bool {
			if (d.sampleJob) != nil {
				return SetCompareGpuSets((d.sampleJob.Gpus), v.(gpuset.GpuSet), op)
			}
			return false
		},
	},
	"GpuFail": Predicate[*jobSummary]{
		Convert: CvtString2Int,
		Compare: func(d *jobSummary, v any) int {
			if (d.sampleJob) != nil {
				return cmp.Compare((d.sampleJob.GpuFail), v.(int))
			}
			return -1
		},
	},
	"Cmd": Predicate[*jobSummary]{
		Compare: func(d *jobSummary, v any) int {
			if (d.sampleJob) != nil {
				return cmp.Compare((d.sampleJob.Cmd), v.(string))
			}
			return -1
		},
	},
	"Hosts": Predicate[*jobSummary]{
		Convert: CvtString2Hostnames,
		SetCompare: func(d *jobSummary, v any, op int) bool {
			if (d.sampleJob) != nil {
				return SetCompareHostnames((d.sampleJob.Hosts), v.(*Hostnames), op)
			}
			return false
		},
	},
	"Now": Predicate[*jobSummary]{
		Convert: CvtString2DateTimeValue,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.Now), v.(DateTimeValue))
		},
	},
	"Classification": Predicate[*jobSummary]{
		Convert: CvtString2Int,
		Compare: func(d *jobSummary, v any) int {
			if (d.sampleJob) != nil {
				return cmp.Compare((d.sampleJob.Classification), v.(int))
			}
			return -1
		},
	},
	"CpuTime": Predicate[*jobSummary]{
		Convert: CvtString2DurationValue,
		Compare: func(d *jobSummary, v any) int {
			if (d.sampleJob) != nil {
				return cmp.Compare((d.sampleJob.CpuTime), v.(DurationValue))
			}
			return -1
		},
	},
	"GpuTime": Predicate[*jobSummary]{
		Convert: CvtString2DurationValue,
		Compare: func(d *jobSummary, v any) int {
			if (d.sampleJob) != nil {
				return cmp.Compare((d.sampleJob.GpuTime), v.(DurationValue))
			}
			return -1
		},
	},
	"SomeGpu": Predicate[*jobSummary]{
		Convert: CvtString2Bool,
		Compare: func(d *jobSummary, v any) int {
			if (d.sampleJob) != nil {
				return CompareBool((d.sampleJob.ComputedFlags&samplejob.KUsesGpu != 0), v.(bool))
			}
			return -1
		},
	},
	"NoGpu": Predicate[*jobSummary]{
		Convert: CvtString2Bool,
		Compare: func(d *jobSummary, v any) int {
			if (d.sampleJob) != nil {
				return CompareBool((d.sampleJob.ComputedFlags&samplejob.KDoesNotUseGpu != 0), v.(bool))
			}
			return -1
		},
	},
	"Running": Predicate[*jobSummary]{
		Convert: CvtString2Bool,
		Compare: func(d *jobSummary, v any) int {
			if (d.sampleJob) != nil {
				return CompareBool((d.sampleJob.ComputedFlags&samplejob.KIsLiveAtEnd != 0), v.(bool))
			}
			return -1
		},
	},
	"Completed": Predicate[*jobSummary]{
		Convert: CvtString2Bool,
		Compare: func(d *jobSummary, v any) int {
			if (d.sampleJob) != nil {
				return CompareBool((d.sampleJob.ComputedFlags&samplejob.KIsNotLiveAtEnd != 0), v.(bool))
			}
			return -1
		},
	},
	"Zombie": Predicate[*jobSummary]{
		Convert: CvtString2Bool,
		Compare: func(d *jobSummary, v any) int {
			if (d.sampleJob) != nil {
				return CompareBool((d.sampleJob.ComputedFlags&samplejob.KIsZombie != 0), v.(bool))
			}
			return -1
		},
	},
	"Primordial": Predicate[*jobSummary]{
		Convert: CvtString2Bool,
		Compare: func(d *jobSummary, v any) int {
			if (d.sampleJob) != nil {
				return CompareBool((d.sampleJob.ComputedFlags&samplejob.KIsLiveAtStart != 0), v.(bool))
			}
			return -1
		},
	},
	"BornLater": Predicate[*jobSummary]{
		Convert: CvtString2Bool,
		Compare: func(d *jobSummary, v any) int {
			if (d.sampleJob) != nil {
				return CompareBool((d.sampleJob.ComputedFlags&samplejob.KIsNotLiveAtStart != 0), v.(bool))
			}
			return -1
		},
	},
	"Submit": Predicate[*jobSummary]{
		Convert: CvtString2DateTimeValue,
		Compare: func(d *jobSummary, v any) int {
			if (d.slurmJob.Main) != nil {
				return cmp.Compare((d.slurmJob.Main.Submit), v.(DateTimeValue))
			}
			return -1
		},
	},
	"JobName": Predicate[*jobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *jobSummary, v any) int {
			if (d.slurmJob.Main) != nil {
				return cmp.Compare((d.slurmJob.Main.JobName), v.(Ustr))
			}
			return -1
		},
	},
	"State": Predicate[*jobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *jobSummary, v any) int {
			if (d.slurmJob.Main) != nil {
				return cmp.Compare((d.slurmJob.Main.State), v.(Ustr))
			}
			return -1
		},
	},
	"Account": Predicate[*jobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *jobSummary, v any) int {
			if (d.slurmJob.Main) != nil {
				return cmp.Compare((d.slurmJob.Main.Account), v.(Ustr))
			}
			return -1
		},
	},
	"Layout": Predicate[*jobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *jobSummary, v any) int {
			if (d.slurmJob.Main) != nil {
				return cmp.Compare((d.slurmJob.Main.Layout), v.(Ustr))
			}
			return -1
		},
	},
	"Reservation": Predicate[*jobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *jobSummary, v any) int {
			if (d.slurmJob.Main) != nil {
				return cmp.Compare((d.slurmJob.Main.Reservation), v.(Ustr))
			}
			return -1
		},
	},
	"Partition": Predicate[*jobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *jobSummary, v any) int {
			if (d.slurmJob.Main) != nil {
				return cmp.Compare((d.slurmJob.Main.Partition), v.(Ustr))
			}
			return -1
		},
	},
	"RequestedGpus": Predicate[*jobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *jobSummary, v any) int {
			if (d.slurmJob.Main) != nil {
				return cmp.Compare((d.slurmJob.Main.ReqGPUS), v.(Ustr))
			}
			return -1
		},
	},
	"DiskReadAvgGB": Predicate[*jobSummary]{
		Convert: CvtString2Uint32,
		Compare: func(d *jobSummary, v any) int {
			if (d.slurmJob.Main) != nil {
				return cmp.Compare((d.slurmJob.Main.AveDiskRead), v.(uint32))
			}
			return -1
		},
	},
	"DiskWriteAvgGB": Predicate[*jobSummary]{
		Convert: CvtString2Uint32,
		Compare: func(d *jobSummary, v any) int {
			if (d.slurmJob.Main) != nil {
				return cmp.Compare((d.slurmJob.Main.AveDiskWrite), v.(uint32))
			}
			return -1
		},
	},
	"RequestedCpus": Predicate[*jobSummary]{
		Convert: CvtString2Uint32,
		Compare: func(d *jobSummary, v any) int {
			if (d.slurmJob.Main) != nil {
				return cmp.Compare((d.slurmJob.Main.ReqCPUS), v.(uint32))
			}
			return -1
		},
	},
	"RequestedMemGB": Predicate[*jobSummary]{
		Convert: CvtString2Uint32,
		Compare: func(d *jobSummary, v any) int {
			if (d.slurmJob.Main) != nil {
				return cmp.Compare((d.slurmJob.Main.ReqMem), v.(uint32))
			}
			return -1
		},
	},
	"RequestedNodes": Predicate[*jobSummary]{
		Convert: CvtString2Uint32,
		Compare: func(d *jobSummary, v any) int {
			if (d.slurmJob.Main) != nil {
				return cmp.Compare((d.slurmJob.Main.ReqNodes), v.(uint32))
			}
			return -1
		},
	},
	"TimeLimit": Predicate[*jobSummary]{
		Convert: CvtString2U32Duration,
		Compare: func(d *jobSummary, v any) int {
			if (d.slurmJob.Main) != nil {
				return cmp.Compare((d.slurmJob.Main.TimelimitRaw), v.(U32Duration))
			}
			return -1
		},
	},
	"ExitCode": Predicate[*jobSummary]{
		Convert: CvtString2Uint8,
		Compare: func(d *jobSummary, v any) int {
			if (d.slurmJob.Main) != nil {
				return cmp.Compare((d.slurmJob.Main.ExitCode), v.(uint8))
			}
			return -1
		},
	},
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
	"All":                    []string{"JobAndMark", "Job", "User", "Duration", "Duration/sec", "Start", "Start/sec", "End", "End/sec", "CpuAvgPct", "CpuPeakPct", "RelativeCpuAvgPct", "RelativeCpuPeakPct", "MemAvgGB", "MemPeakGB", "RelativeMemAvgPct", "RelativeMemPeakPct", "ResidentMemAvgGB", "ResidentMemPeakGB", "RelativeResidentMemAvgPct", "RelativeResidentMemPeakPct", "GpuAvgPct", "GpuPeakPct", "RelativeGpuAvgPct", "RelativeGpuPeakPct", "OccupiedRelativeGpuAvgPct", "OccupiedRelativeGpuPeakPct", "GpuMemAvgGB", "GpuMemPeakGB", "RelativeGpuMemAvgPct", "RelativeGpuMemPeakPct", "OccupiedRelativeGpuMemAvgPct", "OccupiedRelativeGpuMemPeakPct", "Gpus", "GpuFail", "Cmd", "Hosts", "Now", "Now/sec", "Classification", "CpuTime/sec", "CpuTime", "GpuTime/sec", "GpuTime", "SomeGpu", "NoGpu", "Running", "Completed", "Zombie", "Primordial", "BornLater"},
	"Std":                    []string{"JobAndMark", "User", "Duration", "Hosts"},
	"Cpu":                    []string{"CpuAvgPct", "CpuPeakPct"},
	"RelativeCpu":            []string{"RelativeCpuAvgPct", "RelativeCpuPeakPct"},
	"Mem":                    []string{"MemAvgGB", "MemPeakGB"},
	"RelativeMem":            []string{"RelativeMemAvgPct", "RelativeMemPeakPct"},
	"ResidentMem":            []string{"ResidentMemAvgGB", "ResidentMemPeakGB"},
	"RelativeResidentMem":    []string{"RelativeResidentMemAvgPct", "RelativeResidentMemPeakPct"},
	"Gpu":                    []string{"GpuAvgPct", "GpuPeakPct"},
	"RelativeGpu":            []string{"RelativeGpuAvgPct", "RelativeGpuPeakPct"},
	"OccupiedRelativeGpu":    []string{"OccupiedRelativeGpuAvgPct", "OccupiedRelativeGpuPeakPct"},
	"GpuMem":                 []string{"GpuMemAvgGB", "GpuMemPeakGB"},
	"RelativeGpuMem":         []string{"RelativeGpuMemAvgPct", "RelativeGpuMemPeakPct"},
	"OccupiedRelativeGpuMem": []string{"OccupiedRelativeGpuMemAvgPct", "OccupiedRelativeGpuMemPeakPct"},
	"default":                []string{"std", "cpu", "mem", "gpu", "gpumem", "cmd"},
	"Default":                []string{"Std", "Cpu", "Mem", "Gpu", "GpuMem", "Cmd"},
}

const jobsDefaultFields = "default"
