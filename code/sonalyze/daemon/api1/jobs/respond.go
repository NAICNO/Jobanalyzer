// Generated from jobs.go by generate-response.  DO NOT EDIT.

package jobs

import (
	"go-utils/gpuset"
	"sonalyze/cmd/jobs"
	. "sonalyze/common"
	"sonalyze/daemon/apiutil"
	. "sonalyze/table"
)

type JobSummary = jobs.JobSummary

const responseDefaults = "Job,User,Duration,Hosts,CpuTime,ResidentMemAvgGB,GpuTime,GpuMemAvgGB,Cmd"

type Jobs_Job struct {
	JobAndMark                    string        `json:"JobAndMark,omitempty"`
	Job                           uint32        `json:"Job,omitempty"`
	User                          Ustr          `json:"User,omitempty"`
	Duration                      DurationValue `json:"Duration,omitempty"`
	Start                         DateTimeValue `json:"Start,omitempty"`
	End                           DateTimeValue `json:"End,omitempty"`
	CpuAvgPct                     F64Ceil       `json:"CpuAvgPct,omitempty"`
	CpuPeakPct                    F64Ceil       `json:"CpuPeakPct,omitempty"`
	RelativeCpuAvgPct             F64Ceil       `json:"RelativeCpuAvgPct,omitempty"`
	RelativeCpuPeakPct            F64Ceil       `json:"RelativeCpuPeakPct,omitempty"`
	MemAvgGB                      F64Ceil       `json:"MemAvgGB,omitempty"`
	MemPeakGB                     F64Ceil       `json:"MemPeakGB,omitempty"`
	RelativeMemAvgPct             F64Ceil       `json:"RelativeMemAvgPct,omitempty"`
	RelativeMemPeakPct            F64Ceil       `json:"RelativeMemPeakPct,omitempty"`
	ResidentMemAvgGB              F64Ceil       `json:"ResidentMemAvgGB,omitempty"`
	ResidentMemPeakGB             F64Ceil       `json:"ResidentMemPeakGB,omitempty"`
	RelativeResidentMemAvgPct     F64Ceil       `json:"RelativeResidentMemAvgPct,omitempty"`
	RelativeResidentMemPeakPct    F64Ceil       `json:"RelativeResidentMemPeakPct,omitempty"`
	GpuAvgPct                     F64Ceil       `json:"GpuAvgPct,omitempty"`
	GpuPeakPct                    F64Ceil       `json:"GpuPeakPct,omitempty"`
	RelativeGpuAvgPct             F64Ceil       `json:"RelativeGpuAvgPct,omitempty"`
	RelativeGpuPeakPct            F64Ceil       `json:"RelativeGpuPeakPct,omitempty"`
	OccupiedRelativeGpuAvgPct     F64Ceil       `json:"OccupiedRelativeGpuAvgPct,omitempty"`
	OccupiedRelativeGpuPeakPct    F64Ceil       `json:"OccupiedRelativeGpuPeakPct,omitempty"`
	GpuMemAvgGB                   F64Ceil       `json:"GpuMemAvgGB,omitempty"`
	GpuMemPeakGB                  F64Ceil       `json:"GpuMemPeakGB,omitempty"`
	RelativeGpuMemAvgPct          F64Ceil       `json:"RelativeGpuMemAvgPct,omitempty"`
	RelativeGpuMemPeakPct         F64Ceil       `json:"RelativeGpuMemPeakPct,omitempty"`
	OccupiedRelativeGpuMemAvgPct  F64Ceil       `json:"OccupiedRelativeGpuMemAvgPct,omitempty"`
	OccupiedRelativeGpuMemPeakPct F64Ceil       `json:"OccupiedRelativeGpuMemPeakPct,omitempty"`
	ThreadAvg                     F64Ceil       `json:"ThreadAvg,omitempty"`
	ThreadPeak                    F64Ceil       `json:"ThreadPeak,omitempty"`
	Gpus                          gpuset.GpuSet `json:"Gpus,omitempty"`
	GpuFail                       int           `json:"GpuFail,omitempty"`
	Cmd                           string        `json:"Cmd,omitempty"`
	Hosts                         *Hostnames    `json:"Hosts,omitempty"`
	Now                           DateTimeValue `json:"Now,omitempty"`
	Classification                int           `json:"Classification,omitempty"`
	CpuTime                       DurationValue `json:"CpuTime,omitempty"`
	GpuTime                       DurationValue `json:"GpuTime,omitempty"`
	ReadGB                        uint64        `json:"ReadGB,omitempty"`
	WrittenGB                     uint64        `json:"WrittenGB,omitempty"`
	SomeGpu                       bool          `json:"SomeGpu,omitempty"`
	NoGpu                         bool          `json:"NoGpu,omitempty"`
	Running                       bool          `json:"Running,omitempty"`
	Completed                     bool          `json:"Completed,omitempty"`
	Zombie                        bool          `json:"Zombie,omitempty"`
	Primordial                    bool          `json:"Primordial,omitempty"`
	BornLater                     bool          `json:"BornLater,omitempty"`
	Account                       Ustr          `json:"Account,omitempty"`
	ArrayJobID                    uint32        `json:"ArrayJobID,omitempty"`
	ArrayStep                     Ustr          `json:"ArrayStep,omitempty"`
	ArrayTaskID                   uint32        `json:"ArrayTaskID,omitempty"`
	AveCPU                        uint64        `json:"AveCPU,omitempty"`
	AveDiskRead                   uint64        `json:"AveDiskRead,omitempty"`
	AveDiskWrite                  uint64        `json:"AveDiskWrite,omitempty"`
	AveRSS                        uint64        `json:"AveRSS,omitempty"`
	AveVMSize                     uint64        `json:"AveVMSize,omitempty"`
	ElapsedRaw                    uint32        `json:"ElapsedRaw,omitempty"`
	ExitCode                      uint8         `json:"ExitCode,omitempty"`
	HetJobID                      uint32        `json:"HetJobID,omitempty"`
	HetJobOffset                  uint32        `json:"HetJobOffset,omitempty"`
	HetStep                       Ustr          `json:"HetStep,omitempty"`
	JobName                       Ustr          `json:"JobName,omitempty"`
	JobStep                       Ustr          `json:"JobStep,omitempty"`
	Layout                        Ustr          `json:"Layout,omitempty"`
	MaxRSS                        uint64        `json:"MaxRSS,omitempty"`
	MaxVMSize                     uint64        `json:"MaxVMSize,omitempty"`
	MinCPU                        uint64        `json:"MinCPU,omitempty"`
	NodeList                      Ustr          `json:"NodeList,omitempty"`
	Partition                     Ustr          `json:"Partition,omitempty"`
	Priority                      uint64        `json:"Priority,omitempty"`
	ReqCPUS                       uint32        `json:"ReqCPUS,omitempty"`
	ReqGPUS                       Ustr          `json:"ReqGPUS,omitempty"`
	ReqMem                        uint64        `json:"ReqMem,omitempty"`
	ReqNodes                      uint32        `json:"ReqNodes,omitempty"`
	Reservation                   Ustr          `json:"Reservation,omitempty"`
	State                         Ustr          `json:"State,omitempty"`
	Submit                        DateTimeValue `json:"Submit,omitempty"`
	Suspended                     uint32        `json:"Suspended,omitempty"`
	SystemCPU                     uint64        `json:"SystemCPU,omitempty"`
	Time                          DateTimeValue `json:"Time,omitempty"`
	TimelimitRaw                  U32Duration   `json:"TimelimitRaw,omitempty"`
	UserCPU                       uint64        `json:"UserCPU,omitempty"`
	Version                       Ustr          `json:"Version,omitempty"`
}

func respond(flds *apiutil.FieldMap, r *JobSummary) Jobs_Job {
	var x Jobs_Job
	if flds.Has("JobAndMark") {
		x.JobAndMark = r.JobAndMark
	}
	if flds.Has("Job") {
		x.Job = r.JobId
	}
	if flds.Has("User") {
		x.User = r.User
	}
	if flds.Has("Duration") {
		x.Duration = r.Duration
	}
	if flds.Has("Start") {
		x.Start = r.Start
	}
	if flds.Has("End") {
		x.End = r.End
	}
	if flds.Has("CpuAvgPct") {
		x.CpuAvgPct = r.computed[kCpuPctAvg]
	}
	if flds.Has("CpuPeakPct") {
		x.CpuPeakPct = r.computed[kCpuPctPeak]
	}
	if flds.Has("RelativeCpuAvgPct") {
		x.RelativeCpuAvgPct = r.computed[kRcpuPctAvg]
	}
	if flds.Has("RelativeCpuPeakPct") {
		x.RelativeCpuPeakPct = r.computed[kRcpuPctPeak]
	}
	if flds.Has("MemAvgGB") {
		x.MemAvgGB = r.computed[kCpuGBAvg]
	}
	if flds.Has("MemPeakGB") {
		x.MemPeakGB = r.computed[kCpuGBPeak]
	}
	if flds.Has("RelativeMemAvgPct") {
		x.RelativeMemAvgPct = r.computed[kRcpuGBAvg]
	}
	if flds.Has("RelativeMemPeakPct") {
		x.RelativeMemPeakPct = r.computed[kRcpuGBPeak]
	}
	if flds.Has("ResidentMemAvgGB") {
		x.ResidentMemAvgGB = r.computed[kRssAnonGBAvg]
	}
	if flds.Has("ResidentMemPeakGB") {
		x.ResidentMemPeakGB = r.computed[kRssAnonGBPeak]
	}
	if flds.Has("RelativeResidentMemAvgPct") {
		x.RelativeResidentMemAvgPct = r.computed[kRrssAnonGBAvg]
	}
	if flds.Has("RelativeResidentMemPeakPct") {
		x.RelativeResidentMemPeakPct = r.computed[kRrssAnonGBPeak]
	}
	if flds.Has("GpuAvgPct") {
		x.GpuAvgPct = r.computed[kGpuPctAvg]
	}
	if flds.Has("GpuPeakPct") {
		x.GpuPeakPct = r.computed[kGpuPctPeak]
	}
	if flds.Has("RelativeGpuAvgPct") {
		x.RelativeGpuAvgPct = r.computed[kRgpuPctAvg]
	}
	if flds.Has("RelativeGpuPeakPct") {
		x.RelativeGpuPeakPct = r.computed[kRgpuPctPeak]
	}
	if flds.Has("OccupiedRelativeGpuAvgPct") {
		x.OccupiedRelativeGpuAvgPct = r.computed[kSgpuPctAvg]
	}
	if flds.Has("OccupiedRelativeGpuPeakPct") {
		x.OccupiedRelativeGpuPeakPct = r.computed[kSgpuPctPeak]
	}
	if flds.Has("GpuMemAvgGB") {
		x.GpuMemAvgGB = r.computed[kGpuGBAvg]
	}
	if flds.Has("GpuMemPeakGB") {
		x.GpuMemPeakGB = r.computed[kGpuGBPeak]
	}
	if flds.Has("RelativeGpuMemAvgPct") {
		x.RelativeGpuMemAvgPct = r.computed[kRgpuGBAvg]
	}
	if flds.Has("RelativeGpuMemPeakPct") {
		x.RelativeGpuMemPeakPct = r.computed[kRgpuGBPeak]
	}
	if flds.Has("OccupiedRelativeGpuMemAvgPct") {
		x.OccupiedRelativeGpuMemAvgPct = r.computed[kSgpuGBAvg]
	}
	if flds.Has("OccupiedRelativeGpuMemPeakPct") {
		x.OccupiedRelativeGpuMemPeakPct = r.computed[kSgpuGBPeak]
	}
	if flds.Has("ThreadAvg") {
		x.ThreadAvg = r.computed[kThreadAvg]
	}
	if flds.Has("ThreadPeak") {
		x.ThreadPeak = r.computed[kThreadPeak]
	}
	if flds.Has("Gpus") {
		x.Gpus = r.Gpus
	}
	if flds.Has("GpuFail") {
		x.GpuFail = r.GpuFail
	}
	if flds.Has("Cmd") {
		x.Cmd = r.Cmd
	}
	if flds.Has("Hosts") {
		x.Hosts = r.Hosts
	}
	if flds.Has("Now") {
		x.Now = r.Now
	}
	if flds.Has("Classification") {
		x.Classification = r.Classification
	}
	if flds.Has("CpuTime") {
		x.CpuTime = r.CpuTime
	}
	if flds.Has("GpuTime") {
		x.GpuTime = r.GpuTime
	}
	if flds.Has("ReadGB") {
		x.ReadGB = r.u64[uReadGBTotal]
	}
	if flds.Has("WrittenGB") {
		x.WrittenGB = r.u64[uWrittenGBTotal]
	}
	if flds.Has("SomeGpu") {
		x.SomeGpu = r.computedFlags&kUsesGpu != 0
	}
	if flds.Has("NoGpu") {
		x.NoGpu = r.computedFlags&kDoesNotUseGpu != 0
	}
	if flds.Has("Running") {
		x.Running = r.computedFlags&kIsLiveAtEnd != 0
	}
	if flds.Has("Completed") {
		x.Completed = r.computedFlags&kIsNotLiveAtEnd != 0
	}
	if flds.Has("Zombie") {
		x.Zombie = r.computedFlags&kIsZombie != 0
	}
	if flds.Has("Primordial") {
		x.Primordial = r.computedFlags&kIsLiveAtStart != 0
	}
	if flds.Has("BornLater") {
		x.BornLater = r.computedFlags&kIsNotLiveAtStart != 0
	}
	if flds.Has("Account") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.Account
		}
	}
	if flds.Has("ArrayJobID") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.ArrayJobID
		}
	}
	if flds.Has("ArrayStep") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.ArrayStep
		}
	}
	if flds.Has("ArrayTaskID") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.ArrayTaskID
		}
	}
	if flds.Has("AveCPU") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.AveCPU
		}
	}
	if flds.Has("AveDiskRead") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.AveDiskRead
		}
	}
	if flds.Has("AveDiskWrite") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.AveDiskWrite
		}
	}
	if flds.Has("AveRSS") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.AveRSS
		}
	}
	if flds.Has("AveVMSize") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.AveVMSize
		}
	}
	if flds.Has("ElapsedRaw") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.ElapsedRaw
		}
	}
	if flds.Has("ExitCode") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.ExitCode
		}
	}
	if flds.Has("HetJobID") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.HetJobID
		}
	}
	if flds.Has("HetJobOffset") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.HetJobOffset
		}
	}
	if flds.Has("HetStep") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.HetStep
		}
	}
	if flds.Has("JobName") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.JobName
		}
	}
	if flds.Has("JobStep") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.JobStep
		}
	}
	if flds.Has("Layout") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.Layout
		}
	}
	if flds.Has("MaxRSS") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.MaxRSS
		}
	}
	if flds.Has("MaxVMSize") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.MaxVMSize
		}
	}
	if flds.Has("MinCPU") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.MinCPU
		}
	}
	if flds.Has("NodeList") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.NodeList
		}
	}
	if flds.Has("Partition") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.Partition
		}
	}
	if flds.Has("Priority") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.Priority
		}
	}
	if flds.Has("ReqCPUS") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.ReqCPUS
		}
	}
	if flds.Has("ReqGPUS") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.ReqGPUS
		}
	}
	if flds.Has("ReqMem") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.ReqMem
		}
	}
	if flds.Has("ReqNodes") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.ReqNodes
		}
	}
	if flds.Has("Reservation") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.Reservation
		}
	}
	if flds.Has("State") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.State
		}
	}
	if flds.Has("Submit") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.Submit
		}
	}
	if flds.Has("Suspended") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.Suspended
		}
	}
	if flds.Has("SystemCPU") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.SystemCPU
		}
	}
	if flds.Has("Time") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.Time
		}
	}
	if flds.Has("TimelimitRaw") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.TimelimitRaw
		}
	}
	if flds.Has("UserCPU") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.UserCPU
		}
	}
	if flds.Has("Version") {
		if (r.sacctInfo) != nil {
			return r.sacctInfo.Version
		}
	}
	return x
}
