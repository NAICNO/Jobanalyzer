// Generated from jobs.go by generate-response.  DO NOT EDIT.

package jobs

import (
	. "sonalyze/cmd/jobs"
	"sonalyze/daemon/apiutil"
	. "sonalyze/table"
)

const responseDefaults = "Job,User,Duration,Hosts,CpuTime,ResidentMemAvgGB,GpuTime,GpuMemAvgGB,Cmd"

type Jobs_Job struct {
	JobAndMark                    string        `json:"JobAndMark,omitempty" doc:"Job ID with mark indicating job running at start+end (!), start (<), or end (>) of time window"`
	Job                           uint32        `json:"Job,omitempty" doc:"Job ID"`
	User                          string        `json:"User,omitempty" doc:"Name of user running the job"`
	Duration                      DurationValue `json:"Duration,omitempty" doc:"Time of last observation minus time of first"`
	Start                         string        `json:"Start,omitempty" doc:"Time of first observation"`
	End                           string        `json:"End,omitempty" doc:"Time of last observation"`
	CpuAvgPct                     F64Ceil       `json:"CpuAvgPct,omitempty" doc:"Average CPU utilization in percent (100% = 1 core)"`
	CpuPeakPct                    F64Ceil       `json:"CpuPeakPct,omitempty" doc:"Peak CPU utilization in percent (100% = 1 core)"`
	RelativeCpuAvgPct             F64Ceil       `json:"RelativeCpuAvgPct,omitempty" doc:"Average relative CPU utilization in percent (100% = all cores)"`
	RelativeCpuPeakPct            F64Ceil       `json:"RelativeCpuPeakPct,omitempty" doc:"Peak relative CPU utilization in percent (100% = all cores)"`
	MemAvgGB                      F64Ceil       `json:"MemAvgGB,omitempty" doc:"Average main virtual memory utilization in GB"`
	MemPeakGB                     F64Ceil       `json:"MemPeakGB,omitempty" doc:"Peak main virtual memory utilization in GB"`
	RelativeMemAvgPct             F64Ceil       `json:"RelativeMemAvgPct,omitempty" doc:"Average relative main virtual memory utilization in percent (100% = system RAM)"`
	RelativeMemPeakPct            F64Ceil       `json:"RelativeMemPeakPct,omitempty" doc:"Peak relative main virtual memory utilization in percent (100% = system RAM)"`
	ResidentMemAvgGB              F64Ceil       `json:"ResidentMemAvgGB,omitempty" doc:"Average main resident memory utilization in GB"`
	ResidentMemPeakGB             F64Ceil       `json:"ResidentMemPeakGB,omitempty" doc:"Peak main resident memory utilization in GB"`
	RelativeResidentMemAvgPct     F64Ceil       `json:"RelativeResidentMemAvgPct,omitempty" doc:"Average relative main resident memory utilization in percent (100% = all RAM)"`
	RelativeResidentMemPeakPct    F64Ceil       `json:"RelativeResidentMemPeakPct,omitempty" doc:"Peak relative main resident memory utilization in percent (100% = all RAM)"`
	GpuAvgPct                     F64Ceil       `json:"GpuAvgPct,omitempty" doc:"Average GPU utilization in percent (100% = 1 card)"`
	GpuPeakPct                    F64Ceil       `json:"GpuPeakPct,omitempty" doc:"Peak GPU utilization in percent (100% = 1 card)"`
	RelativeGpuAvgPct             F64Ceil       `json:"RelativeGpuAvgPct,omitempty" doc:"Average relative GPU utilization in percent (100% = all cards)"`
	RelativeGpuPeakPct            F64Ceil       `json:"RelativeGpuPeakPct,omitempty" doc:"Peak relative GPU utilization in percent (100% = all cards)"`
	OccupiedRelativeGpuAvgPct     F64Ceil       `json:"OccupiedRelativeGpuAvgPct,omitempty" doc:"Average relative GPU utilization in percent (100% = all cards used by job)"`
	OccupiedRelativeGpuPeakPct    F64Ceil       `json:"OccupiedRelativeGpuPeakPct,omitempty" doc:"Peak relative GPU utilization in percent (100% = all cards used by job)"`
	GpuMemAvgGB                   F64Ceil       `json:"GpuMemAvgGB,omitempty" doc:"Average resident GPU memory utilization in GB"`
	GpuMemPeakGB                  F64Ceil       `json:"GpuMemPeakGB,omitempty" doc:"Peak resident GPU memory utilization in GB"`
	RelativeGpuMemAvgPct          F64Ceil       `json:"RelativeGpuMemAvgPct,omitempty" doc:"Average relative GPU resident memory utilization in percent (100% = all GPU RAM)"`
	RelativeGpuMemPeakPct         F64Ceil       `json:"RelativeGpuMemPeakPct,omitempty" doc:"Peak relative GPU resident memory utilization in percent (100% = all GPU RAM)"`
	OccupiedRelativeGpuMemAvgPct  F64Ceil       `json:"OccupiedRelativeGpuMemAvgPct,omitempty" doc:"Average relative GPU resident memory utilization in percent (100% = all GPU RAM on cards used by job)"`
	OccupiedRelativeGpuMemPeakPct F64Ceil       `json:"OccupiedRelativeGpuMemPeakPct,omitempty" doc:"Peak relative GPU resident memory utilization in percent (100% = all GPU RAM on cards used by job)"`
	ThreadAvg                     F64Ceil       `json:"ThreadAvg,omitempty" doc:"Average number of active threads summed across all processes"`
	ThreadPeak                    F64Ceil       `json:"ThreadPeak,omitempty" doc:"Peak number of active threads summed across all processes"`
	Gpus                          []int         `json:"Gpus,omitempty" doc:"GPU device numbers used by the job, 'none' if none or 'unknown' in error states"`
	GpuFail                       int           `json:"GpuFail,omitempty" doc:"Flag indicating GPU status (0=Ok, 1=Failing)"`
	Cmd                           string        `json:"Cmd,omitempty" doc:"The commands invoking the processes of the job"`
	Hosts                         []string      `json:"Hosts,omitempty" doc:"List of the host name(s) running the job"`
	Now                           string        `json:"Now,omitempty" doc:"The current time"`
	Classification                int           `json:"Classification,omitempty" doc:"Bit vector of live-at-start (2) and live-at-end (1) flags"`
	CpuTime                       DurationValue `json:"CpuTime,omitempty" doc:"Total CPU time of the job across all cores"`
	GpuTime                       DurationValue `json:"GpuTime,omitempty" doc:"Total GPU time of the job across all cards"`
	ReadGB                        uint64        `json:"ReadGB,omitempty" doc:"Total read traffic"`
	WrittenGB                     uint64        `json:"WrittenGB,omitempty" doc:"Total read traffic"`
	SomeGpu                       bool          `json:"SomeGpu,omitempty" doc:"True iff process was seen to use some GPU"`
	NoGpu                         bool          `json:"NoGpu,omitempty" doc:"True iff process was seen to use no GPU"`
	Running                       bool          `json:"Running,omitempty" doc:"True iff process appears to still be running at end of time window"`
	Completed                     bool          `json:"Completed,omitempty" doc:"True iff process appears not to be running at end of time window"`
	Zombie                        bool          `json:"Zombie,omitempty" doc:"True iff the process looks like a zombie"`
	Primordial                    bool          `json:"Primordial,omitempty" doc:"True iff the process appears to have been alive at the start of the time window"`
	BornLater                     bool          `json:"BornLater,omitempty" doc:"True iff the process appears not to have been alive at the start of the time window"`
	Account                       string        `json:"Account,omitempty" doc:"Name of job's account (Slurm)"`
	ArrayJobID                    uint32        `json:"ArrayJobID,omitempty" doc:"The overarching ID of an array job, or 0 (Slurm)"`
	ArrayStep                     string        `json:"ArrayStep,omitempty" doc:"The name of the step, or empty string (Slurm)"`
	ArrayTaskID                   uint32        `json:"ArrayTaskID,omitempty" doc:"The index of the array element (Slurm)"`
	AveCPU                        uint64        `json:"AveCPU,omitempty" doc:"Average (system + user) CPU time of all tasks in job (sec) (Slurm)"`
	AveDiskRead                   uint64        `json:"AveDiskRead,omitempty" doc:"Average number of KB read by all tasks in job (Slurm)"`
	AveDiskWrite                  uint64        `json:"AveDiskWrite,omitempty" doc:"Average number of KB written by all tasks in job (Slurm)"`
	AveRSS                        uint64        `json:"AveRSS,omitempty" doc:"Average resident set size of all tasks in job (KB) (Slurm)"`
	AveVMSize                     uint64        `json:"AveVMSize,omitempty" doc:"Average Virtual Memory size of all tasks in job (KB) (Slurm)"`
	ElapsedRaw                    uint32        `json:"ElapsedRaw,omitempty" doc:"The job's elapsed time (sec) (Slurm)"`
	ExitCode                      uint8         `json:"ExitCode,omitempty" doc:"Exit code of job (Slurm)"`
	HetJobID                      uint32        `json:"HetJobID,omitempty" doc:"The overarching ID of a heterogenous job, or 0 (Slurm)."`
	HetJobOffset                  uint32        `json:"HetJobOffset,omitempty" doc:"The het job element's index (Slurm)"`
	HetStep                       string        `json:"HetStep,omitempty" doc:"The name of the step, or empty string (Slurm)"`
	JobName                       string        `json:"JobName,omitempty" doc:"Name of the job (Slurm)"`
	JobStep                       string        `json:"JobStep,omitempty" doc:"Name of step if any (Slurm)"`
	Layout                        string        `json:"Layout,omitempty" doc:"Layout spec of job (Slurm)"`
	MaxRSS                        uint64        `json:"MaxRSS,omitempty" doc:"Maximum resident set size of all tasks in job (KB) (Slurm)"`
	MaxVMSize                     uint64        `json:"MaxVMSize,omitempty" doc:"Maximum Virtual Memory size of all tasks in job (KB) (Slurm)"`
	MinCPU                        uint64        `json:"MinCPU,omitempty" doc:"Minimum (system + user) CPU time of all tasks in job (KB) (Slurm)"`
	NodeList                      string        `json:"NodeList,omitempty" doc:"The nodes allocated to the job or step (Slurm)"`
	Partition                     string        `json:"Partition,omitempty" doc:"Partition of job (Slurm)"`
	Priority                      uint64        `json:"Priority,omitempty" doc:"Job priority (Slurm)"`
	ReqCPUS                       uint32        `json:"ReqCPUS,omitempty" doc:"Number of requested CPUs (Slurm)"`
	ReqGPUS                       string        `json:"ReqGPUS,omitempty" doc:"Names of requested GPUs (Slurm AllocTRES)"`
	ReqMem                        uint64        `json:"ReqMem,omitempty" doc:"Requested memory in KB (Slurm)"`
	ReqNodes                      uint32        `json:"ReqNodes,omitempty" doc:"Number of requested nodes (Slurm)"`
	Reservation                   string        `json:"Reservation,omitempty" doc:"Name of job's reservation (Slurm)"`
	State                         string        `json:"State,omitempty" doc:"Completion state of job (Slurm)"`
	Submit                        string        `json:"Submit,omitempty" doc:"Submit time of job (Slurm)"`
	Suspended                     uint32        `json:"Suspended,omitempty" doc:"Number of seconds the job was suspended (Slurm)"`
	SystemCPU                     uint64        `json:"SystemCPU,omitempty" doc:"The amount of system CPU time used by the job or job step (sec) (Slurm)"`
	Time                          string        `json:"Time,omitempty" doc:"Time stamp of reading (Slurm)"`
	TimelimitRaw                  U32Duration   `json:"TimelimitRaw,omitempty" doc:"Elapsed time limit (Slurm)"`
	UserCPU                       uint64        `json:"UserCPU,omitempty" doc:"The amount of user CPU time used by the job or job step (sec) (Slurm)"`
	Version                       string        `json:"Version,omitempty" doc:""`
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
		x.User = JSONFromUstr(r.User)
	}
	if flds.Has("Duration") {
		x.Duration = r.Duration
	}
	if flds.Has("Start") {
		x.Start = JSONFromDateTimeValue(r.Start)
	}
	if flds.Has("End") {
		x.End = JSONFromDateTimeValue(r.End)
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
		x.Gpus = JSONFromGpuSet(r.Gpus)
	}
	if flds.Has("GpuFail") {
		x.GpuFail = r.GpuFail
	}
	if flds.Has("Cmd") {
		x.Cmd = r.Cmd
	}
	if flds.Has("Hosts") {
		x.Hosts = JSONFromHostnames(r.Hosts)
	}
	if flds.Has("Now") {
		x.Now = JSONFromDateTimeValue(r.Now)
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
			x.Account = JSONFromUstr(r.SacctInfo.Account)
		}
	}
	if flds.Has("ArrayJobID") {
		if (r.SacctInfo) != nil {
			x.ArrayJobID = r.SacctInfo.ArrayJobID
		}
	}
	if flds.Has("ArrayStep") {
		if (r.SacctInfo) != nil {
			x.ArrayStep = JSONFromUstr(r.SacctInfo.ArrayStep)
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
			x.HetStep = JSONFromUstr(r.SacctInfo.HetStep)
		}
	}
	if flds.Has("JobName") {
		if (r.SacctInfo) != nil {
			x.JobName = JSONFromUstr(r.SacctInfo.JobName)
		}
	}
	if flds.Has("JobStep") {
		if (r.SacctInfo) != nil {
			x.JobStep = JSONFromUstr(r.SacctInfo.JobStep)
		}
	}
	if flds.Has("Layout") {
		if (r.SacctInfo) != nil {
			x.Layout = JSONFromUstr(r.SacctInfo.Layout)
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
			x.NodeList = JSONFromUstr(r.SacctInfo.NodeList)
		}
	}
	if flds.Has("Partition") {
		if (r.SacctInfo) != nil {
			x.Partition = JSONFromUstr(r.SacctInfo.Partition)
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
			x.ReqGPUS = JSONFromUstr(r.SacctInfo.ReqGPUS)
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
			x.Reservation = JSONFromUstr(r.SacctInfo.Reservation)
		}
	}
	if flds.Has("State") {
		if (r.SacctInfo) != nil {
			x.State = JSONFromUstr(r.SacctInfo.State)
		}
	}
	if flds.Has("Submit") {
		if (r.SacctInfo) != nil {
			x.Submit = JSONFromDateTimeValue(r.SacctInfo.Submit)
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
			x.Time = JSONFromDateTimeValue(r.SacctInfo.Time)
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
			x.Version = JSONFromUstr(r.SacctInfo.Version)
		}
	}
	return x
}
