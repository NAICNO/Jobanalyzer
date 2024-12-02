package jobs

import (
	"cmp"
	"io"
	"slices"

	. "sonalyze/common"
	. "sonalyze/table"
)

//go:generate ../../../generate-table/generate-table -o jobs-table.go print.go

/*TABLE jobs

package jobs

import (
  "go-utils/gpuset"
    . "sonalyze/common"
	. "sonalyze/table"
)

type GpuSet = gpuset.GpuSet

%%

FIELDS *jobSummary

  JobAndMark         string        desc:"Job ID with mark indicating job running at start+end (!), start (<), or end (>) of time window" alias:"jobm"
  Job                int           desc:"Job ID" alias:"job" field:"JobId"
  User               Ustr          desc:"Name of user running the job" alias:"user"
  Duration           DurationValue desc:"Time of last observation minus time of first" alias:"duration"
  Start              DateTimeValue desc:"Time of first observation" alias:"start"
  End                DateTimeValue desc:"Time of last observation" alias:"end"
  CpuAvgPct          IntCeil       desc:"Average CPU utilization in percent (100% = 1 core)" field:"computed[kCpuPctAvg]" alias:"cpu-avg"
  CpuPeakPct         IntCeil       desc:"Peak CPU utilization in percent (100% = 1 core)" field:"computed[kCpuPctPeak]" alias:"cpu-peak"
  RelativeCpuAvgPct  IntCeil       desc:"Average relative CPU utilization in percent (100% = all cores)" field:"computed[kRcpuPctAvg]" alias:"rcpu-avg"
  RelativeCpuPeakPct IntCeil       desc:"Peak relative CPU utilization in percent (100% = all cores)" field:"computed[kRcpuPctPeak]" alias:"rcpu-peak"
  MemAvgGB           IntCeil       desc:"Average main virtual memory utilization in GB" field:"computed[kCpuGBAvg]" alias:"mem-avg"
  MemPeakGB          IntCeil       desc:"Peak main virtual memory utilization in GB" field:"computed[kCpuGBPeak]" alias:"mem-peak"
  RelativeMemAvgPct  IntCeil       desc:"Average relative main virtual memory utilization in percent (100% = system RAM)" \
                                   field:"computed[kRcpuGBAvg]" alias:"rmem-avg"
  RelativeMemPeakPct IntCeil       desc:"Peak relative main virtual memory utilization in percent (100% = system RAM)" \
                                   field:"computed[kRcpuGBPeak]" alias:"rmem-peak"
  ResidentMemAvgGB   IntCeil       desc:"Average main resident memory utilization in GB" field:"computed[kRssAnonGBAvg]" alias:"res-avg"
  ResidentMemPeakGB  IntCeil       desc:"Peak main resident memory utilization in GB" field:"computed[kRssAnonGBPeak]" alias:"res-peak"
  RelativeResidentMemAvgPct \
                     IntCeil       desc:"Average relative main resident memory utilization in percent (100% = all RAM)" \
                                   field:"computed[kRrssAnonGBAvg]" alias:"rres-avg"
  RelativeResidentMemPeakPct \
                     IntCeil       desc:"Peak relative main resident memory utilization in percent (100% = all RAM)" \
                                   field:"computed[kRrssAnonGBPeak]" alias:"rres-peak"
  GpuAvgPct          IntCeil       desc:"Average GPU utilization in percent (100% = 1 card)" field:"computed[kGpuPctAvg]" alias:"gpu-avg"
  GpuPeakPct         IntCeil       desc:"Peak GPU utilization in percent (100% = 1 card)" field:"computed[kGpuPctPeak]" alias:"gpu-peak"
  RelativeGpuAvgPct  IntCeil       desc:"Average relative GPU utilization in percent (100% = all cards)" field:"computed[kRgpuPctAvg]" alias:"rgpu-avg"
  RelativeGpuPeakPct IntCeil       desc:"Peak relative GPU utilization in percent (100% = all cards)" field:"computed[kRgpuPctPeak]" alias:"rgpu-peak"
  OccupiedRelativeGpuAvgPct \
                     IntCeil       desc:"Average relative GPU utilization in percent (100% = all cards used by job)" \
                                   field:"computed[kSgpuPctAvg]" alias:"sgpu-avg"
  OccupiedRelativeGpuPeakPct \
                     IntCeil       desc:"Peak relative GPU utilization in percent (100% = all cards used by job)" \
                                   field:"computed[kSgpuPctPeak]" alias:"sgpu-peak"
  GpuMemAvgGB        IntCeil       desc:"Average resident GPU memory utilization in GB" field:"computed[kGpuGBAvg]" alias:"gpumem-avg"
  GpuMemPeakGB       IntCeil       desc:"Peak resident GPU memory utilization in GB" field:"computed[kGpuGBPeak]" alias:"gpumem-peak"
  RelativeGpuMemAvgPct \
                     IntCeil       desc:"Average relative GPU resident memory utilization in percent (100% = all GPU RAM)" \
                                   field:"computed[kRgpuGBAvg]" alias:"rgpumem-avg"
  RelativeGpuMemPeakPct \
                     IntCeil       desc:"Peak relative GPU resident memory utilization in percent (100% = all GPU RAM)" \
                                   field:"computed[kRgpuGBPeak]" alias:"rgpumem-peak"
  OccupiedRelativeGpuMemAvgPct \
                     IntCeil       desc:"Average relative GPU resident memory utilization in percent (100% = all GPU RAM on cards used by job)" \
                                   field:"computed[kSgpuGBAvg]" alias:"sgpumem-avg"
  OccupiedRelativeGpuMemPeakPct \
                     IntCeil       desc:"Peak relative GPU resident memory utilization in percent (100% = all GPU RAM on cards used by job)" \
                                   field:"computed[kSgpuGBPeak]" alias:"sgpumem-peak"
  Gpus               GpuSet        desc:"GPU device numbers used by the job, 'none' if none or 'unknown' in error states" alias:"gpus"
  GpuFail            int           desc:"Flag indicating GPU status (0=Ok, 1=Failing)" alias:"gpufail"
  Cmd                string        desc:"The commands invoking the processes of the job" alias:"cmd"
  Host               string        desc:"List of the host name(s) running the job (first elements of FQDNs, compressed)" alias:"host"
  Now                DateTimeValue desc:"The current time" alias:"now"
  Classification     int           desc:"Bit vector of live-at-start (2) and live-at-end (1) flags" alias:"classification"
  CpuTime            DurationValue desc:"Total CPU time of the job across all cores" alias:"cputime"
  GpuTime            DurationValue desc:"Total GPU time of the job across all cards" alias:"gputime"

  # NOTE!  The slurm fields (via *sacctInfo) are checked for in perform.go.  We can add more slurm
  # fields here but if so they must also be added there.

  Submit             DateTimeValue desc:"Submit time of job (Slurm)" indirect:"sacctInfo"
  JobName            Ustr          desc:"Name of job (Slurm)" indirect:"sacctInfo"
  State              Ustr          desc:"Completion state of job (Slurm)" indirect:"sacctInfo"
  Account            Ustr          desc:"Name of job's account (Slurm)" indirect:"sacctInfo"
  Layout             Ustr          desc:"Layout spec of job (Slurm)" indirect:"sacctInfo"
  Reservation        Ustr          desc:"Name of job's reservation (Slurm)" indirect:"sacctInfo"
  Partition          Ustr          desc:"Partition of job (Slurm)" indirect:"sacctInfo"
  RequestedGpus      Ustr          desc:"Names of requested GPUs (Slurm AllocTRES)" indirect:"sacctInfo" field:"ReqGPUS"
  DiskReadAvgGB      int           desc:"Average disk read activity in GB/s (Slurm AveDiskRead)" indirect:"sacctInfo" field:"AveDiskRead"
  DiskWriteAvgGB     int           desc:"Average disk write activity in GB/s (Slurm AveDiskWrite)" indirect:"sacctInfo" field:"AveDiskWrite"
  RequestedCpus      int           desc:"Number of requested CPUs (Slurm)" indirect:"sacctInfo" field:"ReqCPUS"
  RequestedMemGB     int           desc:"Requested memory (Slurm)" indirect:"sacctInfo" field:"ReqMem"
  RequestedNodes     int           desc:"Number of requested nodes (Slurm)" indirect:"sacctInfo" field:"ReqNodes"
  TimeLimit          DurationValue desc:"Elapsed time limit (Slurm)" indirect:"sacctInfo" field:"TimelimitRaw"
  ExitCode           int           desc:"Exit code of job (Slurm)" indirect:"sacctInfo"

HELP JobsCommand

  Aggregate process data into data about "jobs" and present them.  Output
  records are sorted in order of increasing start time of the job. The default
  format is 'fixed'.

ALIASES

  all         jobm,job,user,duration,duration/sec,start,start/sec,end,end/sec,cpu-avg,cpu-peak,rcpu-avg,\
              rcpu-peak,mem-avg,mem-peak,rmem-avg,rmem-peak,res-avg,res-peak,rres-avg,rres-peak,gpu-avg,\
              gpu-peak,rgpu-avg,rgpu-peak,sgpu-avg,sgpu-peak,gpumem-avg,gpumem-peak,rgpumem-avg,rgpumem-peak,\
              sgpumem-avg,sgpumem-peak,gpus,gpufail,cmd,host,now,now/sec,classification,cputime/sec,cputime,\
              gputime/sec,gputime
  std         jobm,user,duration,host
  cpu         cpu-avg,cpu-peak
  rcpu        rcpu-avg,rcpu-peak
  mem         mem-avg,mem-peak
  rmem        rmem-avg,rmem-peak
  res         res-avg,res-peak
  rres        rres-avg,rres-peak
  gpu         gpu-avg,gpu-peak
  rgpu        rgpu-avg,rgpu-peak
  sgpu        sgpu-avg,sgpu-peak
  gpumem      gpumem-avg,gpumem-peak
  rgpumem     rgpumem-avg,rgpumem-peak
  sgpumem     sgpumem-avg,sgpumem-peak
  All         JobAndMark,Job,User,Duration,Duration/sec,Start,Start/sec,End,End/sec,CpuAvgPct,CpuPeakPct,\
              RelativeCpuAvgPct,RelativeCpuPeakPct,MemAvgGB,MemPeakGB,RelativeMemAvgPct,RelativeMemPeakPct,\
              ResidentMemAvgGB,ResidentMemPeakGB,RelativeResidentMemAvgPct,RelativeResidentMemPeakPct,\
              GpuAvgPct,GpuPeakPct,RelativeGpuAvgPct,RelativeGpuPeakPct,OccupiedRelativeGpuAvgPct,\
              OccupiedRelativeGpuPeakPct,GpuMemAvgGB,GpuMemPeakGB,RelativeGpuMemAvgPct,\
              RelativeGpuMemPeakPct,OccupiedRelativeGpuMemAvgPct,OccupiedRelativeGpuMemPeakPct,Gpus,GpuFail,\
              Cmd,Host,Now,Now/sec,Classification,CpuTime/sec,CpuTime,GpuTime/sec,GpuTime
  Std         JobAndMark,User,Duration,Host
  Cpu         CpuAvgPct,CpuPeakPct
  RelativeCpu RelativeCpuAvgPct,RelativeCpuPeakPct
  Mem         MemAvgGB,MemPeakGB
  RelativeMem RelativeMemAvgPct,RelativeMemPeakPct
  ResidentMem ResidentMemAvgGB,ResidentMemPeakGB
  RelativeResidentMem \
              RelativeResidentMemAvgPct,RelativeResidentMemPeakPct
  Gpu         GpuAvgPct,GpuPeakPct
  RelativeGpu RelativeGpuAvgPct,RelativeGpuPeakPct
  OccupiedRelativeGpu \
              OccupiedRelativeGpuAvgPct,OccupiedRelativeGpuPeakPct
  GpuMem      GpuMemAvgPct,GpuMemPeakPct
  RelativeGpuMem \
              RelativeGpuMemAvgPct,RelativeGpuMemPeakPct
  OccupiedRelativeGpuMem \
              OccupiedRelativeGpuMemAvgPct,OccupiedRelativeGpuMemPeakPct

  default     std,cpu,mem,gpu,gpumem,cmd
  Default     Std,Cpu,Mem,Gpu,GpuMem,Cmd

DEFAULTS default

ELBAT*/

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

	summaries = slices.DeleteFunc(summaries, func(s *jobSummary) bool { return !s.selected })
	FormatData(
		out,
		jc.PrintFields,
		jobsFormatters,
		jc.PrintOpts,
		summaries,
	)

	return nil
}
