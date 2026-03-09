// DO NOT EDIT.  Generated from print.go by generate-table

package jobs

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
		Xtract: func(d *jobSummary) any {
			return d.JobAndMark
		},
		Help: "(string) Job ID with mark indicating job running at start+end (!), start (<), or end (>) of time window",
	},
	"Job": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatUint32((d.JobId), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.JobId
		},
		Help: "(uint32) Job ID",
	},
	"User": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatUstr((d.User), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.User
		},
		Help: "(string) Name of user running the job",
	},
	"Duration": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatDurationValue((d.Duration), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.Duration
		},
		Help: "(DurationValue) Time of last observation minus time of first",
	},
	"Start": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatDateTimeValue((d.Start), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.Start
		},
		Help: "(DateTimeValue) Time of first observation",
	},
	"End": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatDateTimeValue((d.End), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.End
		},
		Help: "(DateTimeValue) Time of last observation",
	},
	"CpuAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kCpuPctAvg]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kCpuPctAvg]
		},
		Help: "(int) Average CPU utilization in percent (100% = 1 core)",
	},
	"CpuPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kCpuPctPeak]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kCpuPctPeak]
		},
		Help: "(int) Peak CPU utilization in percent (100% = 1 core)",
	},
	"RelativeCpuAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kRcpuPctAvg]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kRcpuPctAvg]
		},
		Help:        "(int) Average relative CPU utilization in percent (100% = all cores)",
		NeedsConfig: true,
	},
	"RelativeCpuPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kRcpuPctPeak]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kRcpuPctPeak]
		},
		Help:        "(int) Peak relative CPU utilization in percent (100% = all cores)",
		NeedsConfig: true,
	},
	"MemAvgGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kCpuGBAvg]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kCpuGBAvg]
		},
		Help: "(int) Average main virtual memory utilization in GB",
	},
	"MemPeakGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kCpuGBPeak]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kCpuGBPeak]
		},
		Help: "(int) Peak main virtual memory utilization in GB",
	},
	"RelativeMemAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kRcpuGBAvg]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kRcpuGBAvg]
		},
		Help:        "(int) Average relative main virtual memory utilization in percent (100% = system RAM)",
		NeedsConfig: true,
	},
	"RelativeMemPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kRcpuGBPeak]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kRcpuGBPeak]
		},
		Help:        "(int) Peak relative main virtual memory utilization in percent (100% = system RAM)",
		NeedsConfig: true,
	},
	"ResidentMemAvgGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kRssAnonGBAvg]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kRssAnonGBAvg]
		},
		Help: "(int) Average main resident memory utilization in GB",
	},
	"ResidentMemPeakGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kRssAnonGBPeak]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kRssAnonGBPeak]
		},
		Help: "(int) Peak main resident memory utilization in GB",
	},
	"RelativeResidentMemAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kRrssAnonGBAvg]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kRrssAnonGBAvg]
		},
		Help:        "(int) Average relative main resident memory utilization in percent (100% = all RAM)",
		NeedsConfig: true,
	},
	"RelativeResidentMemPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kRrssAnonGBPeak]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kRrssAnonGBPeak]
		},
		Help:        "(int) Peak relative main resident memory utilization in percent (100% = all RAM)",
		NeedsConfig: true,
	},
	"GpuAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kGpuPctAvg]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kGpuPctAvg]
		},
		Help: "(int) Average GPU utilization in percent (100% = 1 card)",
	},
	"GpuPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kGpuPctPeak]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kGpuPctPeak]
		},
		Help: "(int) Peak GPU utilization in percent (100% = 1 card)",
	},
	"RelativeGpuAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kRgpuPctAvg]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kRgpuPctAvg]
		},
		Help:        "(int) Average relative GPU utilization in percent (100% = all cards)",
		NeedsConfig: true,
	},
	"RelativeGpuPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kRgpuPctPeak]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kRgpuPctPeak]
		},
		Help:        "(int) Peak relative GPU utilization in percent (100% = all cards)",
		NeedsConfig: true,
	},
	"OccupiedRelativeGpuAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kSgpuPctAvg]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kSgpuPctAvg]
		},
		Help:        "(int) Average relative GPU utilization in percent (100% = all cards used by job)",
		NeedsConfig: true,
	},
	"OccupiedRelativeGpuPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kSgpuPctPeak]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kSgpuPctPeak]
		},
		Help:        "(int) Peak relative GPU utilization in percent (100% = all cards used by job)",
		NeedsConfig: true,
	},
	"GpuMemAvgGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kGpuGBAvg]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kGpuGBAvg]
		},
		Help: "(int) Average resident GPU memory utilization in GB",
	},
	"GpuMemPeakGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kGpuGBPeak]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kGpuGBPeak]
		},
		Help: "(int) Peak resident GPU memory utilization in GB",
	},
	"RelativeGpuMemAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kRgpuGBAvg]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kRgpuGBAvg]
		},
		Help:        "(int) Average relative GPU resident memory utilization in percent (100% = all GPU RAM)",
		NeedsConfig: true,
	},
	"RelativeGpuMemPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kRgpuGBPeak]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kRgpuGBPeak]
		},
		Help:        "(int) Peak relative GPU resident memory utilization in percent (100% = all GPU RAM)",
		NeedsConfig: true,
	},
	"OccupiedRelativeGpuMemAvgPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kSgpuGBAvg]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kSgpuGBAvg]
		},
		Help:        "(int) Average relative GPU resident memory utilization in percent (100% = all GPU RAM on cards used by job)",
		NeedsConfig: true,
	},
	"OccupiedRelativeGpuMemPeakPct": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kSgpuGBPeak]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kSgpuGBPeak]
		},
		Help:        "(int) Peak relative GPU resident memory utilization in percent (100% = all GPU RAM on cards used by job)",
		NeedsConfig: true,
	},
	"ThreadAvg": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kThreadAvg]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kThreadAvg]
		},
		Help: "(int) Average number of active threads summed across all processes",
	},
	"ThreadPeak": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.computed[kThreadPeak]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computed[kThreadPeak]
		},
		Help: "(int) Peak number of active threads summed across all processes",
	},
	"Gpus": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatGpuSet((d.Gpus), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.Gpus
		},
		Help: "(GpuSet) GPU device numbers used by the job, 'none' if none or 'unknown' in error states",
	},
	"GpuFail": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatInt((d.GpuFail), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.GpuFail
		},
		Help: "(int) Flag indicating GPU status (0=Ok, 1=Failing)",
	},
	"Cmd": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatString((d.Cmd), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.Cmd
		},
		Help: "(string) The commands invoking the processes of the job",
	},
	"Hosts": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatHostnames((d.Hosts), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.Hosts
		},
		Help: "(Hostnames) List of the host name(s) running the job",
	},
	"Now": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatDateTimeValue((d.Now), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.Now
		},
		Help: "(DateTimeValue) The current time",
	},
	"Classification": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatInt((d.Classification), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.Classification
		},
		Help: "(int) Bit vector of live-at-start (2) and live-at-end (1) flags",
	},
	"CpuTime": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatDurationValue((d.CpuTime), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.CpuTime
		},
		Help: "(DurationValue) Total CPU time of the job across all cores",
	},
	"GpuTime": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatDurationValue((d.GpuTime), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.GpuTime
		},
		Help: "(DurationValue) Total GPU time of the job across all cards",
	},
	"ReadGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatUint64((d.u64[uReadGBTotal]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.u64[uReadGBTotal]
		},
		Help: "(uint64) Total read traffic",
	},
	"WrittenGB": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatUint64((d.u64[uWrittenGBTotal]), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.u64[uWrittenGBTotal]
		},
		Help: "(uint64) Total read traffic",
	},
	"SomeGpu": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatBool((d.computedFlags&kUsesGpu != 0), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computedFlags&kUsesGpu != 0
		},
		Help: "(bool) True iff process was seen to use some GPU",
	},
	"NoGpu": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatBool((d.computedFlags&kDoesNotUseGpu != 0), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computedFlags&kDoesNotUseGpu != 0
		},
		Help: "(bool) True iff process was seen to use no GPU",
	},
	"Running": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatBool((d.computedFlags&kIsLiveAtEnd != 0), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computedFlags&kIsLiveAtEnd != 0
		},
		Help: "(bool) True iff process appears to still be running at end of time window",
	},
	"Completed": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatBool((d.computedFlags&kIsNotLiveAtEnd != 0), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computedFlags&kIsNotLiveAtEnd != 0
		},
		Help: "(bool) True iff process appears not to be running at end of time window",
	},
	"Zombie": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatBool((d.computedFlags&kIsZombie != 0), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computedFlags&kIsZombie != 0
		},
		Help: "(bool) True iff the process looks like a zombie",
	},
	"Primordial": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatBool((d.computedFlags&kIsLiveAtStart != 0), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computedFlags&kIsLiveAtStart != 0
		},
		Help: "(bool) True iff the process appears to have been alive at the start of the time window",
	},
	"BornLater": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			return FormatBool((d.computedFlags&kIsNotLiveAtStart != 0), ctx)
		},
		Xtract: func(d *jobSummary) any {
			return d.computedFlags&kIsNotLiveAtStart != 0
		},
		Help: "(bool) True iff the process appears not to have been alive at the start of the time window",
	},
	"Account": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUstr((d.sacctInfo.Account), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.Account
			}
			return "?"
		},
		Help: "(string) Name of job's account (Slurm)",
	},
	"ArrayJobID": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUint32((d.sacctInfo.ArrayJobID), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.ArrayJobID
			}
			return "?"
		},
		Help: "(uint32) The overarching ID of an array job, or 0 (Slurm)",
	},
	"ArrayStep": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUstr((d.sacctInfo.ArrayStep), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.ArrayStep
			}
			return "?"
		},
		Help: "(string) The name of the step, or empty string (Slurm)",
	},
	"ArrayTaskID": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUint32((d.sacctInfo.ArrayTaskID), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.ArrayTaskID
			}
			return "?"
		},
		Help: "(uint32) The index of the array element (Slurm)",
	},
	"AveCPU": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUint64((d.sacctInfo.AveCPU), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.AveCPU
			}
			return "?"
		},
		Help: "(uint64) Average (system + user) CPU time of all tasks in job (sec) (Slurm)",
	},
	"AveDiskRead": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUint64((d.sacctInfo.AveDiskRead), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.AveDiskRead
			}
			return "?"
		},
		Help: "(uint64) Average number of KB read by all tasks in job (Slurm)",
	},
	"AveDiskWrite": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUint64((d.sacctInfo.AveDiskWrite), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.AveDiskWrite
			}
			return "?"
		},
		Help: "(uint64) Average number of KB written by all tasks in job (Slurm)",
	},
	"AveRSS": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUint64((d.sacctInfo.AveRSS), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.AveRSS
			}
			return "?"
		},
		Help: "(uint64) Average resident set size of all tasks in job (KB) (Slurm)",
	},
	"AveVMSize": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUint64((d.sacctInfo.AveVMSize), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.AveVMSize
			}
			return "?"
		},
		Help: "(uint64) Average Virtual Memory size of all tasks in job (KB) (Slurm)",
	},
	"ElapsedRaw": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUint32((d.sacctInfo.ElapsedRaw), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.ElapsedRaw
			}
			return "?"
		},
		Help: "(uint32) The job's elapsed time (sec) (Slurm)",
	},
	"ExitCode": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUint8((d.sacctInfo.ExitCode), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.ExitCode
			}
			return "?"
		},
		Help: "(uint8) Exit code of job (Slurm)",
	},
	"ExitSignal": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUint8((d.sacctInfo.ExitSignal), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.ExitSignal
			}
			return "?"
		},
		Help: "(uint8) Exit signal of job (Slurm)",
	},
	"HetJobID": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUint32((d.sacctInfo.HetJobID), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.HetJobID
			}
			return "?"
		},
		Help: "(uint32) The overarching ID of a heterogenous job, or 0 (Slurm).",
	},
	"HetJobOffset": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUint32((d.sacctInfo.HetJobOffset), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.HetJobOffset
			}
			return "?"
		},
		Help: "(uint32) The het job element's index (Slurm)",
	},
	"HetStep": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUstr((d.sacctInfo.HetStep), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.HetStep
			}
			return "?"
		},
		Help: "(string) The name of the step, or empty string (Slurm)",
	},
	"JobName": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUstr((d.sacctInfo.JobName), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.JobName
			}
			return "?"
		},
		Help: "(string) Name of the job (Slurm)",
	},
	"JobStep": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUstr((d.sacctInfo.JobStep), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.JobStep
			}
			return "?"
		},
		Help: "(string) Name of step if any (Slurm)",
	},
	"Layout": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUstr((d.sacctInfo.Layout), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.Layout
			}
			return "?"
		},
		Help: "(string) Layout spec of job (Slurm)",
	},
	"MaxRSS": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUint64((d.sacctInfo.MaxRSS), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.MaxRSS
			}
			return "?"
		},
		Help: "(uint64) Maximum resident set size of all tasks in job (KB) (Slurm)",
	},
	"MaxVMSize": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUint64((d.sacctInfo.MaxVMSize), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.MaxVMSize
			}
			return "?"
		},
		Help: "(uint64) Maximum Virtual Memory size of all tasks in job (KB) (Slurm)",
	},
	"MinCPU": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUint64((d.sacctInfo.MinCPU), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.MinCPU
			}
			return "?"
		},
		Help: "(uint64) Minimum (system + user) CPU time of all tasks in job (KB) (Slurm)",
	},
	"NodeList": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUstr((d.sacctInfo.NodeList), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.NodeList
			}
			return "?"
		},
		Help: "(string) The nodes allocated to the job or step (Slurm)",
	},
	"Partition": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUstr((d.sacctInfo.Partition), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.Partition
			}
			return "?"
		},
		Help: "(string) Partition of job (Slurm)",
	},
	"ReqCPUS": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUint32((d.sacctInfo.ReqCPUS), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.ReqCPUS
			}
			return "?"
		},
		Help: "(uint32) Number of requested CPUs (Slurm)",
	},
	"ReqGPUS": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUstr((d.sacctInfo.ReqGPUS), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.ReqGPUS
			}
			return "?"
		},
		Help: "(string) Names of requested GPUs (Slurm AllocTRES)",
	},
	"ReqMem": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUint64((d.sacctInfo.ReqMem), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.ReqMem
			}
			return "?"
		},
		Help: "(uint64) Requested memory in KB (Slurm)",
	},
	"ReqNodes": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUint32((d.sacctInfo.ReqNodes), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.ReqNodes
			}
			return "?"
		},
		Help: "(uint32) Number of requested nodes (Slurm)",
	},
	"Reservation": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUstr((d.sacctInfo.Reservation), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.Reservation
			}
			return "?"
		},
		Help: "(string) Name of job's reservation (Slurm)",
	},
	"State": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUstr((d.sacctInfo.State), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.State
			}
			return "?"
		},
		Help: "(string) Completion state of job (Slurm)",
	},
	"Submit": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatDateTimeValue((d.sacctInfo.Submit), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.Submit
			}
			return "?"
		},
		Help: "(DateTimeValue) Submit time of job (Slurm)",
	},
	"Suspended": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUint32((d.sacctInfo.Suspended), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.Suspended
			}
			return "?"
		},
		Help: "(uint32) Number of seconds the job was suspended (Slurm)",
	},
	"SystemCPU": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUint64((d.sacctInfo.SystemCPU), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.SystemCPU
			}
			return "?"
		},
		Help: "(uint64) The amount of system CPU time used by the job or job step (sec) (Slurm)",
	},
	"Time": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatDateTimeValue((d.sacctInfo.Time), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.Time
			}
			return "?"
		},
		Help: "(DateTimeValue) Time stamp of reading (Slurm)",
	},
	"TimelimitRaw": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatU32Duration((d.sacctInfo.TimelimitRaw), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.TimelimitRaw
			}
			return "?"
		},
		Help: "(U32Duration) Elapsed time limit (Slurm)",
	},
	"UserCPU": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUint64((d.sacctInfo.UserCPU), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.UserCPU
			}
			return "?"
		},
		Help: "(uint64) The amount of user CPU time used by the job or job step (sec) (Slurm)",
	},
	"Version": {
		Fmt: func(d *jobSummary, ctx PrintMods) string {
			if (d.sacctInfo) != nil {
				return FormatUstr((d.sacctInfo.Version), ctx)
			}
			return "?"
		},
		Xtract: func(d *jobSummary) any {
			if (d.sacctInfo) != nil {
				return d.sacctInfo.Version
			}
			return "?"
		},
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
	DefAlias(jobsFormatters, "ThreadAvg", "thread-avg")
	DefAlias(jobsFormatters, "ThreadPeak", "thread-peak")
	DefAlias(jobsFormatters, "Gpus", "gpus")
	DefAlias(jobsFormatters, "GpuFail", "gpufail")
	DefAlias(jobsFormatters, "Cmd", "cmd")
	DefAlias(jobsFormatters, "Hosts", "host")
	DefAlias(jobsFormatters, "Hosts", "hosts")
	DefAlias(jobsFormatters, "Now", "now")
	DefAlias(jobsFormatters, "Classification", "classification")
	DefAlias(jobsFormatters, "CpuTime", "cputime")
	DefAlias(jobsFormatters, "GpuTime", "gputime")
	DefAlias(jobsFormatters, "ReadGB", "read")
	DefAlias(jobsFormatters, "WrittenGB", "written")
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
			return cmp.Compare((d.JobId), v.(uint32))
		},
	},
	"User": Predicate[*jobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.User), v.(Ustr))
		},
	},
	"Duration": Predicate[*jobSummary]{
		Convert: CvtString2DurationValue,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.Duration), v.(DurationValue))
		},
	},
	"Start": Predicate[*jobSummary]{
		Convert: CvtString2DateTimeValue,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.Start), v.(DateTimeValue))
		},
	},
	"End": Predicate[*jobSummary]{
		Convert: CvtString2DateTimeValue,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.End), v.(DateTimeValue))
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
	"ThreadAvg": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kThreadAvg]), v.(F64Ceil))
		},
	},
	"ThreadPeak": Predicate[*jobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.computed[kThreadPeak]), v.(F64Ceil))
		},
	},
	"Gpus": Predicate[*jobSummary]{
		Convert: CvtString2GpuSet,
		SetCompare: func(d *jobSummary, v any, op int) bool {
			return SetCompareGpuSets((d.Gpus), v.(gpuset.GpuSet), op)
		},
	},
	"GpuFail": Predicate[*jobSummary]{
		Convert: CvtString2Int,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.GpuFail), v.(int))
		},
	},
	"Cmd": Predicate[*jobSummary]{
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.Cmd), v.(string))
		},
	},
	"Hosts": Predicate[*jobSummary]{
		Convert: CvtString2Hostnames,
		SetCompare: func(d *jobSummary, v any, op int) bool {
			return SetCompareHostnames((d.Hosts), v.(*Hostnames), op)
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
			return cmp.Compare((d.Classification), v.(int))
		},
	},
	"CpuTime": Predicate[*jobSummary]{
		Convert: CvtString2DurationValue,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.CpuTime), v.(DurationValue))
		},
	},
	"GpuTime": Predicate[*jobSummary]{
		Convert: CvtString2DurationValue,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.GpuTime), v.(DurationValue))
		},
	},
	"ReadGB": Predicate[*jobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.u64[uReadGBTotal]), v.(uint64))
		},
	},
	"WrittenGB": Predicate[*jobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *jobSummary, v any) int {
			return cmp.Compare((d.u64[uWrittenGBTotal]), v.(uint64))
		},
	},
	"SomeGpu": Predicate[*jobSummary]{
		Convert: CvtString2Bool,
		Compare: func(d *jobSummary, v any) int {
			return CompareBool((d.computedFlags&kUsesGpu != 0), v.(bool))
		},
	},
	"NoGpu": Predicate[*jobSummary]{
		Convert: CvtString2Bool,
		Compare: func(d *jobSummary, v any) int {
			return CompareBool((d.computedFlags&kDoesNotUseGpu != 0), v.(bool))
		},
	},
	"Running": Predicate[*jobSummary]{
		Convert: CvtString2Bool,
		Compare: func(d *jobSummary, v any) int {
			return CompareBool((d.computedFlags&kIsLiveAtEnd != 0), v.(bool))
		},
	},
	"Completed": Predicate[*jobSummary]{
		Convert: CvtString2Bool,
		Compare: func(d *jobSummary, v any) int {
			return CompareBool((d.computedFlags&kIsNotLiveAtEnd != 0), v.(bool))
		},
	},
	"Zombie": Predicate[*jobSummary]{
		Convert: CvtString2Bool,
		Compare: func(d *jobSummary, v any) int {
			return CompareBool((d.computedFlags&kIsZombie != 0), v.(bool))
		},
	},
	"Primordial": Predicate[*jobSummary]{
		Convert: CvtString2Bool,
		Compare: func(d *jobSummary, v any) int {
			return CompareBool((d.computedFlags&kIsLiveAtStart != 0), v.(bool))
		},
	},
	"BornLater": Predicate[*jobSummary]{
		Convert: CvtString2Bool,
		Compare: func(d *jobSummary, v any) int {
			return CompareBool((d.computedFlags&kIsNotLiveAtStart != 0), v.(bool))
		},
	},
	"Account": Predicate[*jobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.Account), v.(Ustr))
			}
			return -1
		},
	},
	"ArrayJobID": Predicate[*jobSummary]{
		Convert: CvtString2Uint32,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.ArrayJobID), v.(uint32))
			}
			return -1
		},
	},
	"ArrayStep": Predicate[*jobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.ArrayStep), v.(Ustr))
			}
			return -1
		},
	},
	"ArrayTaskID": Predicate[*jobSummary]{
		Convert: CvtString2Uint32,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.ArrayTaskID), v.(uint32))
			}
			return -1
		},
	},
	"AveCPU": Predicate[*jobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.AveCPU), v.(uint64))
			}
			return -1
		},
	},
	"AveDiskRead": Predicate[*jobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.AveDiskRead), v.(uint64))
			}
			return -1
		},
	},
	"AveDiskWrite": Predicate[*jobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.AveDiskWrite), v.(uint64))
			}
			return -1
		},
	},
	"AveRSS": Predicate[*jobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.AveRSS), v.(uint64))
			}
			return -1
		},
	},
	"AveVMSize": Predicate[*jobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.AveVMSize), v.(uint64))
			}
			return -1
		},
	},
	"ElapsedRaw": Predicate[*jobSummary]{
		Convert: CvtString2Uint32,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.ElapsedRaw), v.(uint32))
			}
			return -1
		},
	},
	"ExitCode": Predicate[*jobSummary]{
		Convert: CvtString2Uint8,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.ExitCode), v.(uint8))
			}
			return -1
		},
	},
	"ExitSignal": Predicate[*jobSummary]{
		Convert: CvtString2Uint8,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.ExitSignal), v.(uint8))
			}
			return -1
		},
	},
	"HetJobID": Predicate[*jobSummary]{
		Convert: CvtString2Uint32,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.HetJobID), v.(uint32))
			}
			return -1
		},
	},
	"HetJobOffset": Predicate[*jobSummary]{
		Convert: CvtString2Uint32,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.HetJobOffset), v.(uint32))
			}
			return -1
		},
	},
	"HetStep": Predicate[*jobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.HetStep), v.(Ustr))
			}
			return -1
		},
	},
	"JobName": Predicate[*jobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.JobName), v.(Ustr))
			}
			return -1
		},
	},
	"JobStep": Predicate[*jobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.JobStep), v.(Ustr))
			}
			return -1
		},
	},
	"Layout": Predicate[*jobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.Layout), v.(Ustr))
			}
			return -1
		},
	},
	"MaxRSS": Predicate[*jobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.MaxRSS), v.(uint64))
			}
			return -1
		},
	},
	"MaxVMSize": Predicate[*jobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.MaxVMSize), v.(uint64))
			}
			return -1
		},
	},
	"MinCPU": Predicate[*jobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.MinCPU), v.(uint64))
			}
			return -1
		},
	},
	"NodeList": Predicate[*jobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.NodeList), v.(Ustr))
			}
			return -1
		},
	},
	"Partition": Predicate[*jobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.Partition), v.(Ustr))
			}
			return -1
		},
	},
	"ReqCPUS": Predicate[*jobSummary]{
		Convert: CvtString2Uint32,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.ReqCPUS), v.(uint32))
			}
			return -1
		},
	},
	"ReqGPUS": Predicate[*jobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.ReqGPUS), v.(Ustr))
			}
			return -1
		},
	},
	"ReqMem": Predicate[*jobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.ReqMem), v.(uint64))
			}
			return -1
		},
	},
	"ReqNodes": Predicate[*jobSummary]{
		Convert: CvtString2Uint32,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.ReqNodes), v.(uint32))
			}
			return -1
		},
	},
	"Reservation": Predicate[*jobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.Reservation), v.(Ustr))
			}
			return -1
		},
	},
	"State": Predicate[*jobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.State), v.(Ustr))
			}
			return -1
		},
	},
	"Submit": Predicate[*jobSummary]{
		Convert: CvtString2DateTimeValue,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.Submit), v.(DateTimeValue))
			}
			return -1
		},
	},
	"Suspended": Predicate[*jobSummary]{
		Convert: CvtString2Uint32,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.Suspended), v.(uint32))
			}
			return -1
		},
	},
	"SystemCPU": Predicate[*jobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.SystemCPU), v.(uint64))
			}
			return -1
		},
	},
	"Time": Predicate[*jobSummary]{
		Convert: CvtString2DateTimeValue,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.Time), v.(DateTimeValue))
			}
			return -1
		},
	},
	"TimelimitRaw": Predicate[*jobSummary]{
		Convert: CvtString2U32Duration,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.TimelimitRaw), v.(U32Duration))
			}
			return -1
		},
	},
	"UserCPU": Predicate[*jobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.UserCPU), v.(uint64))
			}
			return -1
		},
	},
	"Version": Predicate[*jobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *jobSummary, v any) int {
			if (d.sacctInfo) != nil {
				return cmp.Compare((d.sacctInfo.Version), v.(Ustr))
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
	"all":                    []string{"jobm", "job", "user", "duration", "duration/sec", "start", "start/sec", "end", "end/sec", "cpu-avg", "cpu-peak", "rcpu-avg", "rcpu-peak", "mem-avg", "mem-peak", "rmem-avg", "rmem-peak", "res-avg", "res-peak", "rres-avg", "rres-peak", "gpu-avg", "gpu-peak", "rgpu-avg", "rgpu-peak", "sgpu-avg", "sgpu-peak", "gpumem-avg", "gpumem-peak", "rgpumem-avg", "rgpumem-peak", "sgpumem-avg", "sgpumem-peak", "thread-avg", "thread-peak", "gpus", "gpufail", "cmd", "host", "now", "now/sec", "classification", "cputime/sec", "cputime", "gputime/sec", "gputime"},
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
	"threads":                []string{"thread-avg", "thread-peak"},
	"All":                    []string{"JobAndMark", "Job", "User", "Duration", "Duration/sec", "Start", "Start/sec", "End", "End/sec", "CpuAvgPct", "CpuPeakPct", "RelativeCpuAvgPct", "RelativeCpuPeakPct", "MemAvgGB", "MemPeakGB", "RelativeMemAvgPct", "RelativeMemPeakPct", "ResidentMemAvgGB", "ResidentMemPeakGB", "RelativeResidentMemAvgPct", "RelativeResidentMemPeakPct", "GpuAvgPct", "GpuPeakPct", "RelativeGpuAvgPct", "RelativeGpuPeakPct", "OccupiedRelativeGpuAvgPct", "OccupiedRelativeGpuPeakPct", "GpuMemAvgGB", "GpuMemPeakGB", "RelativeGpuMemAvgPct", "RelativeGpuMemPeakPct", "OccupiedRelativeGpuMemAvgPct", "OccupiedRelativeGpuMemPeakPct", "ThreadAvg", "ThreadPeak", "Gpus", "GpuFail", "Cmd", "Hosts", "Now", "Now/sec", "Classification", "CpuTime/sec", "CpuTime", "GpuTime/sec", "GpuTime", "SomeGpu", "NoGpu", "Running", "Completed", "Zombie", "Primordial", "BornLater"},
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
	"Threads":                []string{"ThreadAvg", "ThreadPeak"},
	"default":                []string{"std", "cpu", "mem", "gpu", "gpumem", "cmd"},
	"Default":                []string{"Std", "Cpu", "Mem", "Gpu", "GpuMem", "Cmd"},
}

const jobsDefaultFields = "default"
