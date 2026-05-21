// Generated from jobs.go by generate-response.  DO NOT EDIT.

package jobs

import (
	"go-utils/gpuset"
	. "sonalyze/cmd/jobs"
	. "sonalyze/common"
	"sonalyze/daemon/apiutil"
	. "sonalyze/table"
)

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
		x.CpuAvgPct = r.Computed[KCpuPctAvg]
	}
	if flds.Has("CpuPeakPct") {
		x.CpuPeakPct = r.Computed[KCpuPctPeak]
	}
	if flds.Has("RelativeCpuAvgPct") {
		x.RelativeCpuAvgPct = r.Computed[KRcpuPctAvg]
	}
	if flds.Has("RelativeCpuPeakPct") {
		x.RelativeCpuPeakPct = r.Computed[KRcpuPctPeak]
	}
	if flds.Has("MemAvgGB") {
		x.MemAvgGB = r.Computed[KCpuGBAvg]
	}
	if flds.Has("MemPeakGB") {
		x.MemPeakGB = r.Computed[KCpuGBPeak]
	}
	if flds.Has("RelativeMemAvgPct") {
		x.RelativeMemAvgPct = r.Computed[KRcpuGBAvg]
	}
	if flds.Has("RelativeMemPeakPct") {
		x.RelativeMemPeakPct = r.Computed[KRcpuGBPeak]
	}
	if flds.Has("ResidentMemAvgGB") {
		x.ResidentMemAvgGB = r.Computed[KRssAnonGBAvg]
	}
	if flds.Has("ResidentMemPeakGB") {
		x.ResidentMemPeakGB = r.Computed[KRssAnonGBPeak]
	}
	if flds.Has("RelativeResidentMemAvgPct") {
		x.RelativeResidentMemAvgPct = r.Computed[KRrssAnonGBAvg]
	}
	if flds.Has("RelativeResidentMemPeakPct") {
		x.RelativeResidentMemPeakPct = r.Computed[KRrssAnonGBPeak]
	}
	if flds.Has("GpuAvgPct") {
		x.GpuAvgPct = r.Computed[KGpuPctAvg]
	}
	if flds.Has("GpuPeakPct") {
		x.GpuPeakPct = r.Computed[KGpuPctPeak]
	}
	if flds.Has("RelativeGpuAvgPct") {
		x.RelativeGpuAvgPct = r.Computed[KRgpuPctAvg]
	}
	if flds.Has("RelativeGpuPeakPct") {
		x.RelativeGpuPeakPct = r.Computed[KRgpuPctPeak]
	}
	if flds.Has("OccupiedRelativeGpuAvgPct") {
		x.OccupiedRelativeGpuAvgPct = r.Computed[KSgpuPctAvg]
	}
	if flds.Has("OccupiedRelativeGpuPeakPct") {
		x.OccupiedRelativeGpuPeakPct = r.Computed[KSgpuPctPeak]
	}
	if flds.Has("GpuMemAvgGB") {
		x.GpuMemAvgGB = r.Computed[KGpuGBAvg]
	}
	if flds.Has("GpuMemPeakGB") {
		x.GpuMemPeakGB = r.Computed[KGpuGBPeak]
	}
	if flds.Has("RelativeGpuMemAvgPct") {
		x.RelativeGpuMemAvgPct = r.Computed[KRgpuGBAvg]
	}
	if flds.Has("RelativeGpuMemPeakPct") {
		x.RelativeGpuMemPeakPct = r.Computed[KRgpuGBPeak]
	}
	if flds.Has("OccupiedRelativeGpuMemAvgPct") {
		x.OccupiedRelativeGpuMemAvgPct = r.Computed[KSgpuGBAvg]
	}
	if flds.Has("OccupiedRelativeGpuMemPeakPct") {
		x.OccupiedRelativeGpuMemPeakPct = r.Computed[KSgpuGBPeak]
	}
	if flds.Has("ThreadAvg") {
		x.ThreadAvg = r.Computed[KThreadAvg]
	}
	if flds.Has("ThreadPeak") {
		x.ThreadPeak = r.Computed[KThreadPeak]
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
		x.ReadGB = r.U64[UReadGBTotal]
	}
	if flds.Has("WrittenGB") {
		x.WrittenGB = r.U64[UWrittenGBTotal]
	}
	if flds.Has("SomeGpu") {
		x.SomeGpu = r.ComputedFlags&KUsesGpu != 0
	}
	if flds.Has("NoGpu") {
		x.NoGpu = r.ComputedFlags&KDoesNotUseGpu != 0
	}
	if flds.Has("Running") {
		x.Running = r.ComputedFlags&KIsLiveAtEnd != 0
	}
	if flds.Has("Completed") {
		x.Completed = r.ComputedFlags&KIsNotLiveAtEnd != 0
	}
	if flds.Has("Zombie") {
		x.Zombie = r.ComputedFlags&KIsZombie != 0
	}
	if flds.Has("Primordial") {
		x.Primordial = r.ComputedFlags&KIsLiveAtStart != 0
	}
	if flds.Has("BornLater") {
		x.BornLater = r.ComputedFlags&KIsNotLiveAtStart != 0
	}
	if flds.Has("Account") {
		if (r.SacctInfo) != nil {
			x.Account = r.SacctInfo.Account
		}
	}
	if flds.Has("ArrayJobID") {
		if (r.SacctInfo) != nil {
			x.ArrayJobID = r.SacctInfo.ArrayJobID
		}
	}
	if flds.Has("ArrayStep") {
		if (r.SacctInfo) != nil {
			x.ArrayStep = r.SacctInfo.ArrayStep
		}
	}
	if flds.Has("ArrayTaskID") {
		if (r.SacctInfo) != nil {
			x.ArrayTaskID = r.SacctInfo.ArrayTaskID
		}
	}
	if flds.Has("AveCPU") {
		if (r.SacctInfo) != nil {
			x.AveCPU = r.SacctInfo.AveCPU
		}
	}
	if flds.Has("AveDiskRead") {
		if (r.SacctInfo) != nil {
			x.AveDiskRead = r.SacctInfo.AveDiskRead
		}
	}
	if flds.Has("AveDiskWrite") {
		if (r.SacctInfo) != nil {
			x.AveDiskWrite = r.SacctInfo.AveDiskWrite
		}
	}
	if flds.Has("AveRSS") {
		if (r.SacctInfo) != nil {
			x.AveRSS = r.SacctInfo.AveRSS
		}
	}
	if flds.Has("AveVMSize") {
		if (r.SacctInfo) != nil {
			x.AveVMSize = r.SacctInfo.AveVMSize
		}
	}
	if flds.Has("ElapsedRaw") {
		if (r.SacctInfo) != nil {
			x.ElapsedRaw = r.SacctInfo.ElapsedRaw
		}
	}
	if flds.Has("ExitCode") {
		if (r.SacctInfo) != nil {
			x.ExitCode = r.SacctInfo.ExitCode
		}
	}
	if flds.Has("HetJobID") {
		if (r.SacctInfo) != nil {
			x.HetJobID = r.SacctInfo.HetJobID
		}
	}
	if flds.Has("HetJobOffset") {
		if (r.SacctInfo) != nil {
			x.HetJobOffset = r.SacctInfo.HetJobOffset
		}
	}
	if flds.Has("HetStep") {
		if (r.SacctInfo) != nil {
			x.HetStep = r.SacctInfo.HetStep
		}
	}
	if flds.Has("JobName") {
		if (r.SacctInfo) != nil {
			x.JobName = r.SacctInfo.JobName
		}
	}
	if flds.Has("JobStep") {
		if (r.SacctInfo) != nil {
			x.JobStep = r.SacctInfo.JobStep
		}
	}
	if flds.Has("Layout") {
		if (r.SacctInfo) != nil {
			x.Layout = r.SacctInfo.Layout
		}
	}
	if flds.Has("MaxRSS") {
		if (r.SacctInfo) != nil {
			x.MaxRSS = r.SacctInfo.MaxRSS
		}
	}
	if flds.Has("MaxVMSize") {
		if (r.SacctInfo) != nil {
			x.MaxVMSize = r.SacctInfo.MaxVMSize
		}
	}
	if flds.Has("MinCPU") {
		if (r.SacctInfo) != nil {
			x.MinCPU = r.SacctInfo.MinCPU
		}
	}
	if flds.Has("NodeList") {
		if (r.SacctInfo) != nil {
			x.NodeList = r.SacctInfo.NodeList
		}
	}
	if flds.Has("Partition") {
		if (r.SacctInfo) != nil {
			x.Partition = r.SacctInfo.Partition
		}
	}
	if flds.Has("Priority") {
		if (r.SacctInfo) != nil {
			x.Priority = r.SacctInfo.Priority
		}
	}
	if flds.Has("ReqCPUS") {
		if (r.SacctInfo) != nil {
			x.ReqCPUS = r.SacctInfo.ReqCPUS
		}
	}
	if flds.Has("ReqGPUS") {
		if (r.SacctInfo) != nil {
			x.ReqGPUS = r.SacctInfo.ReqGPUS
		}
	}
	if flds.Has("ReqMem") {
		if (r.SacctInfo) != nil {
			x.ReqMem = r.SacctInfo.ReqMem
		}
	}
	if flds.Has("ReqNodes") {
		if (r.SacctInfo) != nil {
			x.ReqNodes = r.SacctInfo.ReqNodes
		}
	}
	if flds.Has("Reservation") {
		if (r.SacctInfo) != nil {
			x.Reservation = r.SacctInfo.Reservation
		}
	}
	if flds.Has("State") {
		if (r.SacctInfo) != nil {
			x.State = r.SacctInfo.State
		}
	}
	if flds.Has("Submit") {
		if (r.SacctInfo) != nil {
			x.Submit = r.SacctInfo.Submit
		}
	}
	if flds.Has("Suspended") {
		if (r.SacctInfo) != nil {
			x.Suspended = r.SacctInfo.Suspended
		}
	}
	if flds.Has("SystemCPU") {
		if (r.SacctInfo) != nil {
			x.SystemCPU = r.SacctInfo.SystemCPU
		}
	}
	if flds.Has("Time") {
		if (r.SacctInfo) != nil {
			x.Time = r.SacctInfo.Time
		}
	}
	if flds.Has("TimelimitRaw") {
		if (r.SacctInfo) != nil {
			x.TimelimitRaw = r.SacctInfo.TimelimitRaw
		}
	}
	if flds.Has("UserCPU") {
		if (r.SacctInfo) != nil {
			x.UserCPU = r.SacctInfo.UserCPU
		}
	}
	if flds.Has("Version") {
		if (r.SacctInfo) != nil {
			x.Version = r.SacctInfo.Version
		}
	}
	return x
}
