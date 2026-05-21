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
var jobsFormatters = map[string]Formatter[*JobSummary]{
	"JobAndMark": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatString((d.JobAndMark), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.JobAndMark
		},
		Help: "(string) Job ID with mark indicating job running at start+end (!), start (<), or end (>) of time window",
	},
	"Job": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatUint32((d.JobId), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.JobId
		},
		Help: "(uint32) Job ID",
	},
	"User": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatUstr((d.User), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.User
		},
		Help: "(string) Name of user running the job",
	},
	"Duration": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatDurationValue((d.Duration), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Duration
		},
		Help: "(DurationValue) Time of last observation minus time of first",
	},
	"Start": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatDateTimeValue((d.Start), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Start
		},
		Help: "(DateTimeValue) Time of first observation",
	},
	"End": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatDateTimeValue((d.End), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.End
		},
		Help: "(DateTimeValue) Time of last observation",
	},
	"CpuAvgPct": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KCpuPctAvg]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KCpuPctAvg]
		},
		Help: "(int) Average CPU utilization in percent (100% = 1 core)",
	},
	"CpuPeakPct": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KCpuPctPeak]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KCpuPctPeak]
		},
		Help: "(int) Peak CPU utilization in percent (100% = 1 core)",
	},
	"RelativeCpuAvgPct": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KRcpuPctAvg]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KRcpuPctAvg]
		},
		Help:        "(int) Average relative CPU utilization in percent (100% = all cores)",
		NeedsConfig: true,
	},
	"RelativeCpuPeakPct": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KRcpuPctPeak]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KRcpuPctPeak]
		},
		Help:        "(int) Peak relative CPU utilization in percent (100% = all cores)",
		NeedsConfig: true,
	},
	"MemAvgGB": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KCpuGBAvg]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KCpuGBAvg]
		},
		Help: "(int) Average main virtual memory utilization in GB",
	},
	"MemPeakGB": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KCpuGBPeak]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KCpuGBPeak]
		},
		Help: "(int) Peak main virtual memory utilization in GB",
	},
	"RelativeMemAvgPct": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KRcpuGBAvg]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KRcpuGBAvg]
		},
		Help:        "(int) Average relative main virtual memory utilization in percent (100% = system RAM)",
		NeedsConfig: true,
	},
	"RelativeMemPeakPct": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KRcpuGBPeak]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KRcpuGBPeak]
		},
		Help:        "(int) Peak relative main virtual memory utilization in percent (100% = system RAM)",
		NeedsConfig: true,
	},
	"ResidentMemAvgGB": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KRssAnonGBAvg]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KRssAnonGBAvg]
		},
		Help: "(int) Average main resident memory utilization in GB",
	},
	"ResidentMemPeakGB": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KRssAnonGBPeak]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KRssAnonGBPeak]
		},
		Help: "(int) Peak main resident memory utilization in GB",
	},
	"RelativeResidentMemAvgPct": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KRrssAnonGBAvg]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KRrssAnonGBAvg]
		},
		Help:        "(int) Average relative main resident memory utilization in percent (100% = all RAM)",
		NeedsConfig: true,
	},
	"RelativeResidentMemPeakPct": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KRrssAnonGBPeak]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KRrssAnonGBPeak]
		},
		Help:        "(int) Peak relative main resident memory utilization in percent (100% = all RAM)",
		NeedsConfig: true,
	},
	"GpuAvgPct": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KGpuPctAvg]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KGpuPctAvg]
		},
		Help: "(int) Average GPU utilization in percent (100% = 1 card)",
	},
	"GpuPeakPct": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KGpuPctPeak]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KGpuPctPeak]
		},
		Help: "(int) Peak GPU utilization in percent (100% = 1 card)",
	},
	"RelativeGpuAvgPct": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KRgpuPctAvg]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KRgpuPctAvg]
		},
		Help:        "(int) Average relative GPU utilization in percent (100% = all cards)",
		NeedsConfig: true,
	},
	"RelativeGpuPeakPct": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KRgpuPctPeak]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KRgpuPctPeak]
		},
		Help:        "(int) Peak relative GPU utilization in percent (100% = all cards)",
		NeedsConfig: true,
	},
	"OccupiedRelativeGpuAvgPct": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KSgpuPctAvg]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KSgpuPctAvg]
		},
		Help:        "(int) Average relative GPU utilization in percent (100% = all cards used by job)",
		NeedsConfig: true,
	},
	"OccupiedRelativeGpuPeakPct": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KSgpuPctPeak]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KSgpuPctPeak]
		},
		Help:        "(int) Peak relative GPU utilization in percent (100% = all cards used by job)",
		NeedsConfig: true,
	},
	"GpuMemAvgGB": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KGpuGBAvg]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KGpuGBAvg]
		},
		Help: "(int) Average resident GPU memory utilization in GB",
	},
	"GpuMemPeakGB": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KGpuGBPeak]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KGpuGBPeak]
		},
		Help: "(int) Peak resident GPU memory utilization in GB",
	},
	"RelativeGpuMemAvgPct": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KRgpuGBAvg]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KRgpuGBAvg]
		},
		Help:        "(int) Average relative GPU resident memory utilization in percent (100% = all GPU RAM)",
		NeedsConfig: true,
	},
	"RelativeGpuMemPeakPct": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KRgpuGBPeak]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KRgpuGBPeak]
		},
		Help:        "(int) Peak relative GPU resident memory utilization in percent (100% = all GPU RAM)",
		NeedsConfig: true,
	},
	"OccupiedRelativeGpuMemAvgPct": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KSgpuGBAvg]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KSgpuGBAvg]
		},
		Help:        "(int) Average relative GPU resident memory utilization in percent (100% = all GPU RAM on cards used by job)",
		NeedsConfig: true,
	},
	"OccupiedRelativeGpuMemPeakPct": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KSgpuGBPeak]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KSgpuGBPeak]
		},
		Help:        "(int) Peak relative GPU resident memory utilization in percent (100% = all GPU RAM on cards used by job)",
		NeedsConfig: true,
	},
	"ThreadAvg": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KThreadAvg]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KThreadAvg]
		},
		Help: "(int) Average number of active threads summed across all processes",
	},
	"ThreadPeak": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatF64Ceil((d.Computed[KThreadPeak]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Computed[KThreadPeak]
		},
		Help: "(int) Peak number of active threads summed across all processes",
	},
	"Gpus": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatGpuSet((d.Gpus), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Gpus
		},
		Help: "(GpuSet) GPU device numbers used by the job, 'none' if none or 'unknown' in error states",
	},
	"GpuFail": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatInt((d.GpuFail), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.GpuFail
		},
		Help: "(int) Flag indicating GPU status (0=Ok, 1=Failing)",
	},
	"Cmd": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatString((d.Cmd), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Cmd
		},
		Help: "(string) The commands invoking the processes of the job",
	},
	"Hosts": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatHostnames((d.Hosts), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Hosts
		},
		Help: "(Hostnames) List of the host name(s) running the job",
	},
	"Now": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatDateTimeValue((d.Now), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Now
		},
		Help: "(DateTimeValue) The current time",
	},
	"Classification": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatInt((d.Classification), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.Classification
		},
		Help: "(int) Bit vector of live-at-start (2) and live-at-end (1) flags",
	},
	"CpuTime": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatDurationValue((d.CpuTime), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.CpuTime
		},
		Help: "(DurationValue) Total CPU time of the job across all cores",
	},
	"GpuTime": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatDurationValue((d.GpuTime), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.GpuTime
		},
		Help: "(DurationValue) Total GPU time of the job across all cards",
	},
	"ReadGB": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatUint64((d.U64[UReadGBTotal]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.U64[UReadGBTotal]
		},
		Help: "(uint64) Total read traffic",
	},
	"WrittenGB": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatUint64((d.U64[UWrittenGBTotal]), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.U64[UWrittenGBTotal]
		},
		Help: "(uint64) Total read traffic",
	},
	"SomeGpu": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatBool((d.ComputedFlags&KUsesGpu != 0), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.ComputedFlags&KUsesGpu != 0
		},
		Help: "(bool) True iff process was seen to use some GPU",
	},
	"NoGpu": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatBool((d.ComputedFlags&KDoesNotUseGpu != 0), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.ComputedFlags&KDoesNotUseGpu != 0
		},
		Help: "(bool) True iff process was seen to use no GPU",
	},
	"Running": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatBool((d.ComputedFlags&KIsLiveAtEnd != 0), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.ComputedFlags&KIsLiveAtEnd != 0
		},
		Help: "(bool) True iff process appears to still be running at end of time window",
	},
	"Completed": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatBool((d.ComputedFlags&KIsNotLiveAtEnd != 0), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.ComputedFlags&KIsNotLiveAtEnd != 0
		},
		Help: "(bool) True iff process appears not to be running at end of time window",
	},
	"Zombie": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatBool((d.ComputedFlags&KIsZombie != 0), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.ComputedFlags&KIsZombie != 0
		},
		Help: "(bool) True iff the process looks like a zombie",
	},
	"Primordial": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatBool((d.ComputedFlags&KIsLiveAtStart != 0), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.ComputedFlags&KIsLiveAtStart != 0
		},
		Help: "(bool) True iff the process appears to have been alive at the start of the time window",
	},
	"BornLater": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			return FormatBool((d.ComputedFlags&KIsNotLiveAtStart != 0), ctx)
		},
		Xtract: func(d *JobSummary) any {
			return d.ComputedFlags&KIsNotLiveAtStart != 0
		},
		Help: "(bool) True iff the process appears not to have been alive at the start of the time window",
	},
	"Account": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUstr((d.SacctInfo.Account), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.Account
			}
			return "?"
		},
		Help: "(string) Name of job's account (Slurm)",
	},
	"ArrayJobID": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUint32((d.SacctInfo.ArrayJobID), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.ArrayJobID
			}
			return "?"
		},
		Help: "(uint32) The overarching ID of an array job, or 0 (Slurm)",
	},
	"ArrayStep": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUstr((d.SacctInfo.ArrayStep), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.ArrayStep
			}
			return "?"
		},
		Help: "(string) The name of the step, or empty string (Slurm)",
	},
	"ArrayTaskID": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUint32((d.SacctInfo.ArrayTaskID), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.ArrayTaskID
			}
			return "?"
		},
		Help: "(uint32) The index of the array element (Slurm)",
	},
	"AveCPU": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUint64((d.SacctInfo.AveCPU), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.AveCPU
			}
			return "?"
		},
		Help: "(uint64) Average (system + user) CPU time of all tasks in job (sec) (Slurm)",
	},
	"AveDiskRead": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUint64((d.SacctInfo.AveDiskRead), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.AveDiskRead
			}
			return "?"
		},
		Help: "(uint64) Average number of KB read by all tasks in job (Slurm)",
	},
	"AveDiskWrite": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUint64((d.SacctInfo.AveDiskWrite), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.AveDiskWrite
			}
			return "?"
		},
		Help: "(uint64) Average number of KB written by all tasks in job (Slurm)",
	},
	"AveRSS": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUint64((d.SacctInfo.AveRSS), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.AveRSS
			}
			return "?"
		},
		Help: "(uint64) Average resident set size of all tasks in job (KB) (Slurm)",
	},
	"AveVMSize": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUint64((d.SacctInfo.AveVMSize), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.AveVMSize
			}
			return "?"
		},
		Help: "(uint64) Average Virtual Memory size of all tasks in job (KB) (Slurm)",
	},
	"ElapsedRaw": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUint32((d.SacctInfo.ElapsedRaw), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.ElapsedRaw
			}
			return "?"
		},
		Help: "(uint32) The job's elapsed time (sec) (Slurm)",
	},
	"ExitCode": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUint8((d.SacctInfo.ExitCode), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.ExitCode
			}
			return "?"
		},
		Help: "(uint8) Exit code of job (Slurm)",
	},
	"HetJobID": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUint32((d.SacctInfo.HetJobID), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.HetJobID
			}
			return "?"
		},
		Help: "(uint32) The overarching ID of a heterogenous job, or 0 (Slurm).",
	},
	"HetJobOffset": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUint32((d.SacctInfo.HetJobOffset), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.HetJobOffset
			}
			return "?"
		},
		Help: "(uint32) The het job element's index (Slurm)",
	},
	"HetStep": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUstr((d.SacctInfo.HetStep), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.HetStep
			}
			return "?"
		},
		Help: "(string) The name of the step, or empty string (Slurm)",
	},
	"JobName": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUstr((d.SacctInfo.JobName), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.JobName
			}
			return "?"
		},
		Help: "(string) Name of the job (Slurm)",
	},
	"JobStep": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUstr((d.SacctInfo.JobStep), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.JobStep
			}
			return "?"
		},
		Help: "(string) Name of step if any (Slurm)",
	},
	"Layout": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUstr((d.SacctInfo.Layout), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.Layout
			}
			return "?"
		},
		Help: "(string) Layout spec of job (Slurm)",
	},
	"MaxRSS": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUint64((d.SacctInfo.MaxRSS), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.MaxRSS
			}
			return "?"
		},
		Help: "(uint64) Maximum resident set size of all tasks in job (KB) (Slurm)",
	},
	"MaxVMSize": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUint64((d.SacctInfo.MaxVMSize), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.MaxVMSize
			}
			return "?"
		},
		Help: "(uint64) Maximum Virtual Memory size of all tasks in job (KB) (Slurm)",
	},
	"MinCPU": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUint64((d.SacctInfo.MinCPU), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.MinCPU
			}
			return "?"
		},
		Help: "(uint64) Minimum (system + user) CPU time of all tasks in job (KB) (Slurm)",
	},
	"NodeList": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUstr((d.SacctInfo.NodeList), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.NodeList
			}
			return "?"
		},
		Help: "(string) The nodes allocated to the job or step (Slurm)",
	},
	"Partition": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUstr((d.SacctInfo.Partition), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.Partition
			}
			return "?"
		},
		Help: "(string) Partition of job (Slurm)",
	},
	"Priority": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUint64((d.SacctInfo.Priority), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.Priority
			}
			return "?"
		},
		Help: "(uint64) Job priority (Slurm)",
	},
	"ReqCPUS": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUint32((d.SacctInfo.ReqCPUS), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.ReqCPUS
			}
			return "?"
		},
		Help: "(uint32) Number of requested CPUs (Slurm)",
	},
	"ReqGPUS": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUstr((d.SacctInfo.ReqGPUS), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.ReqGPUS
			}
			return "?"
		},
		Help: "(string) Names of requested GPUs (Slurm AllocTRES)",
	},
	"ReqMem": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUint64((d.SacctInfo.ReqMem), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.ReqMem
			}
			return "?"
		},
		Help: "(uint64) Requested memory in KB (Slurm)",
	},
	"ReqNodes": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUint32((d.SacctInfo.ReqNodes), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.ReqNodes
			}
			return "?"
		},
		Help: "(uint32) Number of requested nodes (Slurm)",
	},
	"Reservation": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUstr((d.SacctInfo.Reservation), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.Reservation
			}
			return "?"
		},
		Help: "(string) Name of job's reservation (Slurm)",
	},
	"State": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUstr((d.SacctInfo.State), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.State
			}
			return "?"
		},
		Help: "(string) Completion state of job (Slurm)",
	},
	"Submit": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatDateTimeValue((d.SacctInfo.Submit), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.Submit
			}
			return "?"
		},
		Help: "(DateTimeValue) Submit time of job (Slurm)",
	},
	"Suspended": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUint32((d.SacctInfo.Suspended), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.Suspended
			}
			return "?"
		},
		Help: "(uint32) Number of seconds the job was suspended (Slurm)",
	},
	"SystemCPU": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUint64((d.SacctInfo.SystemCPU), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.SystemCPU
			}
			return "?"
		},
		Help: "(uint64) The amount of system CPU time used by the job or job step (sec) (Slurm)",
	},
	"Time": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatDateTimeValue((d.SacctInfo.Time), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.Time
			}
			return "?"
		},
		Help: "(DateTimeValue) Time stamp of reading (Slurm)",
	},
	"TimelimitRaw": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatU32Duration((d.SacctInfo.TimelimitRaw), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.TimelimitRaw
			}
			return "?"
		},
		Help: "(U32Duration) Elapsed time limit (Slurm)",
	},
	"UserCPU": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUint64((d.SacctInfo.UserCPU), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.UserCPU
			}
			return "?"
		},
		Help: "(uint64) The amount of user CPU time used by the job or job step (sec) (Slurm)",
	},
	"Version": {
		Fmt: func(d *JobSummary, ctx PrintMods) string {
			if (d.SacctInfo) != nil {
				return FormatUstr((d.SacctInfo.Version), ctx)
			}
			return "?"
		},
		Xtract: func(d *JobSummary) any {
			if (d.SacctInfo) != nil {
				return d.SacctInfo.Version
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
var jobsPredicates = map[string]Predicate[*JobSummary]{
	"JobAndMark": Predicate[*JobSummary]{
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.JobAndMark), v.(string))
		},
	},
	"Job": Predicate[*JobSummary]{
		Convert: CvtString2Uint32,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.JobId), v.(uint32))
		},
	},
	"User": Predicate[*JobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.User), v.(Ustr))
		},
	},
	"Duration": Predicate[*JobSummary]{
		Convert: CvtString2DurationValue,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Duration), v.(DurationValue))
		},
	},
	"Start": Predicate[*JobSummary]{
		Convert: CvtString2DateTimeValue,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Start), v.(DateTimeValue))
		},
	},
	"End": Predicate[*JobSummary]{
		Convert: CvtString2DateTimeValue,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.End), v.(DateTimeValue))
		},
	},
	"CpuAvgPct": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KCpuPctAvg]), v.(F64Ceil))
		},
	},
	"CpuPeakPct": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KCpuPctPeak]), v.(F64Ceil))
		},
	},
	"RelativeCpuAvgPct": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KRcpuPctAvg]), v.(F64Ceil))
		},
	},
	"RelativeCpuPeakPct": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KRcpuPctPeak]), v.(F64Ceil))
		},
	},
	"MemAvgGB": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KCpuGBAvg]), v.(F64Ceil))
		},
	},
	"MemPeakGB": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KCpuGBPeak]), v.(F64Ceil))
		},
	},
	"RelativeMemAvgPct": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KRcpuGBAvg]), v.(F64Ceil))
		},
	},
	"RelativeMemPeakPct": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KRcpuGBPeak]), v.(F64Ceil))
		},
	},
	"ResidentMemAvgGB": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KRssAnonGBAvg]), v.(F64Ceil))
		},
	},
	"ResidentMemPeakGB": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KRssAnonGBPeak]), v.(F64Ceil))
		},
	},
	"RelativeResidentMemAvgPct": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KRrssAnonGBAvg]), v.(F64Ceil))
		},
	},
	"RelativeResidentMemPeakPct": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KRrssAnonGBPeak]), v.(F64Ceil))
		},
	},
	"GpuAvgPct": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KGpuPctAvg]), v.(F64Ceil))
		},
	},
	"GpuPeakPct": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KGpuPctPeak]), v.(F64Ceil))
		},
	},
	"RelativeGpuAvgPct": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KRgpuPctAvg]), v.(F64Ceil))
		},
	},
	"RelativeGpuPeakPct": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KRgpuPctPeak]), v.(F64Ceil))
		},
	},
	"OccupiedRelativeGpuAvgPct": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KSgpuPctAvg]), v.(F64Ceil))
		},
	},
	"OccupiedRelativeGpuPeakPct": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KSgpuPctPeak]), v.(F64Ceil))
		},
	},
	"GpuMemAvgGB": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KGpuGBAvg]), v.(F64Ceil))
		},
	},
	"GpuMemPeakGB": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KGpuGBPeak]), v.(F64Ceil))
		},
	},
	"RelativeGpuMemAvgPct": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KRgpuGBAvg]), v.(F64Ceil))
		},
	},
	"RelativeGpuMemPeakPct": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KRgpuGBPeak]), v.(F64Ceil))
		},
	},
	"OccupiedRelativeGpuMemAvgPct": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KSgpuGBAvg]), v.(F64Ceil))
		},
	},
	"OccupiedRelativeGpuMemPeakPct": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KSgpuGBPeak]), v.(F64Ceil))
		},
	},
	"ThreadAvg": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KThreadAvg]), v.(F64Ceil))
		},
	},
	"ThreadPeak": Predicate[*JobSummary]{
		Convert: CvtString2Float64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Computed[KThreadPeak]), v.(F64Ceil))
		},
	},
	"Gpus": Predicate[*JobSummary]{
		Convert: CvtString2GpuSet,
		SetCompare: func(d *JobSummary, v any, op int) bool {
			return SetCompareGpuSets((d.Gpus), v.(gpuset.GpuSet), op)
		},
	},
	"GpuFail": Predicate[*JobSummary]{
		Convert: CvtString2Int,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.GpuFail), v.(int))
		},
	},
	"Cmd": Predicate[*JobSummary]{
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Cmd), v.(string))
		},
	},
	"Hosts": Predicate[*JobSummary]{
		Convert: CvtString2Hostnames,
		SetCompare: func(d *JobSummary, v any, op int) bool {
			return SetCompareHostnames((d.Hosts), v.(*Hostnames), op)
		},
	},
	"Now": Predicate[*JobSummary]{
		Convert: CvtString2DateTimeValue,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Now), v.(DateTimeValue))
		},
	},
	"Classification": Predicate[*JobSummary]{
		Convert: CvtString2Int,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.Classification), v.(int))
		},
	},
	"CpuTime": Predicate[*JobSummary]{
		Convert: CvtString2DurationValue,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.CpuTime), v.(DurationValue))
		},
	},
	"GpuTime": Predicate[*JobSummary]{
		Convert: CvtString2DurationValue,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.GpuTime), v.(DurationValue))
		},
	},
	"ReadGB": Predicate[*JobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.U64[UReadGBTotal]), v.(uint64))
		},
	},
	"WrittenGB": Predicate[*JobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *JobSummary, v any) int {
			return cmp.Compare((d.U64[UWrittenGBTotal]), v.(uint64))
		},
	},
	"SomeGpu": Predicate[*JobSummary]{
		Convert: CvtString2Bool,
		Compare: func(d *JobSummary, v any) int {
			return CompareBool((d.ComputedFlags&KUsesGpu != 0), v.(bool))
		},
	},
	"NoGpu": Predicate[*JobSummary]{
		Convert: CvtString2Bool,
		Compare: func(d *JobSummary, v any) int {
			return CompareBool((d.ComputedFlags&KDoesNotUseGpu != 0), v.(bool))
		},
	},
	"Running": Predicate[*JobSummary]{
		Convert: CvtString2Bool,
		Compare: func(d *JobSummary, v any) int {
			return CompareBool((d.ComputedFlags&KIsLiveAtEnd != 0), v.(bool))
		},
	},
	"Completed": Predicate[*JobSummary]{
		Convert: CvtString2Bool,
		Compare: func(d *JobSummary, v any) int {
			return CompareBool((d.ComputedFlags&KIsNotLiveAtEnd != 0), v.(bool))
		},
	},
	"Zombie": Predicate[*JobSummary]{
		Convert: CvtString2Bool,
		Compare: func(d *JobSummary, v any) int {
			return CompareBool((d.ComputedFlags&KIsZombie != 0), v.(bool))
		},
	},
	"Primordial": Predicate[*JobSummary]{
		Convert: CvtString2Bool,
		Compare: func(d *JobSummary, v any) int {
			return CompareBool((d.ComputedFlags&KIsLiveAtStart != 0), v.(bool))
		},
	},
	"BornLater": Predicate[*JobSummary]{
		Convert: CvtString2Bool,
		Compare: func(d *JobSummary, v any) int {
			return CompareBool((d.ComputedFlags&KIsNotLiveAtStart != 0), v.(bool))
		},
	},
	"Account": Predicate[*JobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.Account), v.(Ustr))
			}
			return -1
		},
	},
	"ArrayJobID": Predicate[*JobSummary]{
		Convert: CvtString2Uint32,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.ArrayJobID), v.(uint32))
			}
			return -1
		},
	},
	"ArrayStep": Predicate[*JobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.ArrayStep), v.(Ustr))
			}
			return -1
		},
	},
	"ArrayTaskID": Predicate[*JobSummary]{
		Convert: CvtString2Uint32,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.ArrayTaskID), v.(uint32))
			}
			return -1
		},
	},
	"AveCPU": Predicate[*JobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.AveCPU), v.(uint64))
			}
			return -1
		},
	},
	"AveDiskRead": Predicate[*JobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.AveDiskRead), v.(uint64))
			}
			return -1
		},
	},
	"AveDiskWrite": Predicate[*JobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.AveDiskWrite), v.(uint64))
			}
			return -1
		},
	},
	"AveRSS": Predicate[*JobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.AveRSS), v.(uint64))
			}
			return -1
		},
	},
	"AveVMSize": Predicate[*JobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.AveVMSize), v.(uint64))
			}
			return -1
		},
	},
	"ElapsedRaw": Predicate[*JobSummary]{
		Convert: CvtString2Uint32,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.ElapsedRaw), v.(uint32))
			}
			return -1
		},
	},
	"ExitCode": Predicate[*JobSummary]{
		Convert: CvtString2Uint8,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.ExitCode), v.(uint8))
			}
			return -1
		},
	},
	"HetJobID": Predicate[*JobSummary]{
		Convert: CvtString2Uint32,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.HetJobID), v.(uint32))
			}
			return -1
		},
	},
	"HetJobOffset": Predicate[*JobSummary]{
		Convert: CvtString2Uint32,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.HetJobOffset), v.(uint32))
			}
			return -1
		},
	},
	"HetStep": Predicate[*JobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.HetStep), v.(Ustr))
			}
			return -1
		},
	},
	"JobName": Predicate[*JobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.JobName), v.(Ustr))
			}
			return -1
		},
	},
	"JobStep": Predicate[*JobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.JobStep), v.(Ustr))
			}
			return -1
		},
	},
	"Layout": Predicate[*JobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.Layout), v.(Ustr))
			}
			return -1
		},
	},
	"MaxRSS": Predicate[*JobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.MaxRSS), v.(uint64))
			}
			return -1
		},
	},
	"MaxVMSize": Predicate[*JobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.MaxVMSize), v.(uint64))
			}
			return -1
		},
	},
	"MinCPU": Predicate[*JobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.MinCPU), v.(uint64))
			}
			return -1
		},
	},
	"NodeList": Predicate[*JobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.NodeList), v.(Ustr))
			}
			return -1
		},
	},
	"Partition": Predicate[*JobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.Partition), v.(Ustr))
			}
			return -1
		},
	},
	"Priority": Predicate[*JobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.Priority), v.(uint64))
			}
			return -1
		},
	},
	"ReqCPUS": Predicate[*JobSummary]{
		Convert: CvtString2Uint32,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.ReqCPUS), v.(uint32))
			}
			return -1
		},
	},
	"ReqGPUS": Predicate[*JobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.ReqGPUS), v.(Ustr))
			}
			return -1
		},
	},
	"ReqMem": Predicate[*JobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.ReqMem), v.(uint64))
			}
			return -1
		},
	},
	"ReqNodes": Predicate[*JobSummary]{
		Convert: CvtString2Uint32,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.ReqNodes), v.(uint32))
			}
			return -1
		},
	},
	"Reservation": Predicate[*JobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.Reservation), v.(Ustr))
			}
			return -1
		},
	},
	"State": Predicate[*JobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.State), v.(Ustr))
			}
			return -1
		},
	},
	"Submit": Predicate[*JobSummary]{
		Convert: CvtString2DateTimeValue,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.Submit), v.(DateTimeValue))
			}
			return -1
		},
	},
	"Suspended": Predicate[*JobSummary]{
		Convert: CvtString2Uint32,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.Suspended), v.(uint32))
			}
			return -1
		},
	},
	"SystemCPU": Predicate[*JobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.SystemCPU), v.(uint64))
			}
			return -1
		},
	},
	"Time": Predicate[*JobSummary]{
		Convert: CvtString2DateTimeValue,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.Time), v.(DateTimeValue))
			}
			return -1
		},
	},
	"TimelimitRaw": Predicate[*JobSummary]{
		Convert: CvtString2U32Duration,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.TimelimitRaw), v.(U32Duration))
			}
			return -1
		},
	},
	"UserCPU": Predicate[*JobSummary]{
		Convert: CvtString2Uint64,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.UserCPU), v.(uint64))
			}
			return -1
		},
	},
	"Version": Predicate[*JobSummary]{
		Convert: CvtString2Ustr,
		Compare: func(d *JobSummary, v any) int {
			if (d.SacctInfo) != nil {
				return cmp.Compare((d.SacctInfo.Version), v.(Ustr))
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
