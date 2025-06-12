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

import "sonalyze/data/samplejob"

%%

FIELDS *jobSummary

  JobAndMark         string        desc:"Job ID with mark indicating job running at start+end (!), start (<), or end (>) of time window" alias:"jobm"
  Job                uint32        desc:"Job ID" alias:"job" field:"JobId" indirect:"sampleJob"
  User               Ustr          desc:"Name of user running the job" alias:"user" indirect:"sampleJob"
  Duration           DurationValue desc:"Time of last observation minus time of first" alias:"duration" indirect:"sampleJob"
  Start              DateTimeValue desc:"Time of first observation" alias:"start" indirect:"sampleJob"
  End                DateTimeValue desc:"Time of last observation" alias:"end" indirect:"sampleJob"
  CpuAvgPct          F64Ceil       desc:"Average CPU utilization in percent (100% = 1 core)" field:"computed[kCpuPctAvg]" alias:"cpu-avg"
  CpuPeakPct         F64Ceil       desc:"Peak CPU utilization in percent (100% = 1 core)" field:"computed[kCpuPctPeak]" alias:"cpu-peak"
  RelativeCpuAvgPct  F64Ceil       desc:"Average relative CPU utilization in percent (100% = all cores)" field:"computed[kRcpuPctAvg]" alias:"rcpu-avg"
  RelativeCpuPeakPct F64Ceil       desc:"Peak relative CPU utilization in percent (100% = all cores)" field:"computed[kRcpuPctPeak]" alias:"rcpu-peak"
  MemAvgGB           F64Ceil       desc:"Average main virtual memory utilization in GB" field:"computed[kCpuGBAvg]" alias:"mem-avg"
  MemPeakGB          F64Ceil       desc:"Peak main virtual memory utilization in GB" field:"computed[kCpuGBPeak]" alias:"mem-peak"
  RelativeMemAvgPct  F64Ceil       desc:"Average relative main virtual memory utilization in percent (100% = system RAM)" \
                                   field:"computed[kRcpuGBAvg]" alias:"rmem-avg"
  RelativeMemPeakPct F64Ceil       desc:"Peak relative main virtual memory utilization in percent (100% = system RAM)" \
                                   field:"computed[kRcpuGBPeak]" alias:"rmem-peak"
  ResidentMemAvgGB   F64Ceil       desc:"Average main resident memory utilization in GB" field:"computed[kRssAnonGBAvg]" alias:"res-avg"
  ResidentMemPeakGB  F64Ceil       desc:"Peak main resident memory utilization in GB" field:"computed[kRssAnonGBPeak]" alias:"res-peak"
  RelativeResidentMemAvgPct \
                     F64Ceil       desc:"Average relative main resident memory utilization in percent (100% = all RAM)" \
                                   field:"computed[kRrssAnonGBAvg]" alias:"rres-avg"
  RelativeResidentMemPeakPct \
                     F64Ceil       desc:"Peak relative main resident memory utilization in percent (100% = all RAM)" \
                                   field:"computed[kRrssAnonGBPeak]" alias:"rres-peak"
  GpuAvgPct          F64Ceil       desc:"Average GPU utilization in percent (100% = 1 card)" field:"computed[kGpuPctAvg]" alias:"gpu-avg"
  GpuPeakPct         F64Ceil       desc:"Peak GPU utilization in percent (100% = 1 card)" field:"computed[kGpuPctPeak]" alias:"gpu-peak"
  RelativeGpuAvgPct  F64Ceil       desc:"Average relative GPU utilization in percent (100% = all cards)" field:"computed[kRgpuPctAvg]" alias:"rgpu-avg"
  RelativeGpuPeakPct F64Ceil       desc:"Peak relative GPU utilization in percent (100% = all cards)" field:"computed[kRgpuPctPeak]" alias:"rgpu-peak"
  OccupiedRelativeGpuAvgPct \
                     F64Ceil       desc:"Average relative GPU utilization in percent (100% = all cards used by job)" \
                                   field:"computed[kSgpuPctAvg]" alias:"sgpu-avg"
  OccupiedRelativeGpuPeakPct \
                     F64Ceil       desc:"Peak relative GPU utilization in percent (100% = all cards used by job)" \
                                   field:"computed[kSgpuPctPeak]" alias:"sgpu-peak"
  GpuMemAvgGB        F64Ceil       desc:"Average resident GPU memory utilization in GB" field:"computed[kGpuGBAvg]" alias:"gpumem-avg"
  GpuMemPeakGB       F64Ceil       desc:"Peak resident GPU memory utilization in GB" field:"computed[kGpuGBPeak]" alias:"gpumem-peak"
  RelativeGpuMemAvgPct \
                     F64Ceil       desc:"Average relative GPU resident memory utilization in percent (100% = all GPU RAM)" \
                                   field:"computed[kRgpuGBAvg]" alias:"rgpumem-avg"
  RelativeGpuMemPeakPct \
                     F64Ceil       desc:"Peak relative GPU resident memory utilization in percent (100% = all GPU RAM)" \
                                   field:"computed[kRgpuGBPeak]" alias:"rgpumem-peak"
  OccupiedRelativeGpuMemAvgPct \
                     F64Ceil       desc:"Average relative GPU resident memory utilization in percent (100% = all GPU RAM on cards used by job)" \
                                   field:"computed[kSgpuGBAvg]" alias:"sgpumem-avg"
  OccupiedRelativeGpuMemPeakPct \
                     F64Ceil       desc:"Peak relative GPU resident memory utilization in percent (100% = all GPU RAM on cards used by job)" \
                                   field:"computed[kSgpuGBPeak]" alias:"sgpumem-peak"
  Gpus               gpuset.GpuSet desc:"GPU device numbers used by the job, 'none' if none or 'unknown' in error states" alias:"gpus" indirect:"sampleJob"
  GpuFail            int           desc:"Flag indicating GPU status (0=Ok, 1=Failing)" alias:"gpufail" indirect:"sampleJob"
  Cmd                string        desc:"The commands invoking the processes of the job" alias:"cmd" indirect:"sampleJob"
  Hosts              *Hostnames    desc:"List of the host name(s) running the job" alias:"host,hosts" indirect:"sampleJob"
  Now                DateTimeValue desc:"The current time" alias:"now"
  Classification     int           desc:"Bit vector of live-at-start (2) and live-at-end (1) flags" alias:"classification" indirect:"sampleJob"
  CpuTime            DurationValue desc:"Total CPU time of the job across all cores" alias:"cputime" indirect:"sampleJob"
  GpuTime            DurationValue desc:"Total GPU time of the job across all cards" alias:"gputime" indirect:"sampleJob"

  # The expressions extracting bit flags happen to work for well-understood reasons, but this is
  # brittle and works in Go only because the operator precedence is right (in C it would not work).
  # See TODO in generate-table/README.md.

  SomeGpu            bool          desc:"True iff process was seen to use some GPU" \
                                   field:"ComputedFlags & samplejob.KUsesGpu != 0" indirect:"sampleJob"
  NoGpu              bool          desc:"True iff process was seen to use no GPU" \
                                   field:"ComputedFlags & samplejob.KDoesNotUseGpu != 0" indirect:"sampleJob"
  Running            bool          desc:"True iff process appears to still be running at end of time window" \
                                   field:"ComputedFlags & samplejob.KIsLiveAtEnd != 0" indirect:"sampleJob"
  Completed          bool          desc:"True iff process appears not to be running at end of time window" \
                                   field:"ComputedFlags & samplejob.KIsNotLiveAtEnd != 0" indirect:"sampleJob"
  Zombie             bool          desc:"True iff the process looks like a zombie" \
                                   field:"ComputedFlags & samplejob.KIsZombie != 0" indirect:"sampleJob"
  Primordial         bool          desc:"True iff the process appears to have been alive at the start of the time window" \
                                   field:"ComputedFlags & samplejob.KIsLiveAtStart != 0" indirect:"sampleJob"
  BornLater          bool          desc:"True iff the process appears not to have been alive at the start of the time window" \
                                   field:"ComputedFlags & samplejob.KIsNotLiveAtStart != 0" indirect:"sampleJob"

  # NOTE!  The slurm fields (via *slurmJob) are checked for in perform.go.  We can add more slurm
  # fields here but if so they must also be added there.

  Submit             DateTimeValue desc:"Submit time of job (Slurm)" indirect:"slurmJob.Main"
  JobName            Ustr          desc:"Name of job (Slurm)" indirect:"slurmJob.Main"
  State              Ustr          desc:"Completion state of job (Slurm)" indirect:"slurmJob.Main"
  Account            Ustr          desc:"Name of job's account (Slurm)" indirect:"slurmJob.Main"
  Layout             Ustr          desc:"Layout spec of job (Slurm)" indirect:"slurmJob.Main"
  Reservation        Ustr          desc:"Name of job's reservation (Slurm)" indirect:"slurmJob.Main"
  Partition          Ustr          desc:"Partition of job (Slurm)" indirect:"slurmJob.Main"
  RequestedGpus      Ustr          desc:"Names of requested GPUs (Slurm AllocTRES)" indirect:"slurmJob.Main" field:"ReqGPUS"
  DiskReadAvgGB      uint32        desc:"Average disk read activity in GB/s (Slurm AveDiskRead)" indirect:"slurmJob.Main" field:"AveDiskRead"
  DiskWriteAvgGB     uint32        desc:"Average disk write activity in GB/s (Slurm AveDiskWrite)" indirect:"slurmJob.Main" field:"AveDiskWrite"
  RequestedCpus      uint32        desc:"Number of requested CPUs (Slurm)" indirect:"slurmJob.Main" field:"ReqCPUS"
  RequestedMemGB     uint32        desc:"Requested memory (Slurm)" indirect:"slurmJob.Main" field:"ReqMem"
  RequestedNodes     uint32        desc:"Number of requested nodes (Slurm)" indirect:"slurmJob.Main" field:"ReqNodes"
  TimeLimit          U32Duration   desc:"Elapsed time limit (Slurm)" indirect:"slurmJob.Main" field:"TimelimitRaw"
  ExitCode           uint8         desc:"Exit code of job (Slurm)" indirect:"slurmJob.Main"

SUMMARY JobsCommand

Display jobs jobs aggregated from process samples.

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
              Cmd,Hosts,Now,Now/sec,Classification,CpuTime/sec,CpuTime,GpuTime/sec,GpuTime,\
              SomeGpu,NoGpu,Running,Completed,Zombie,Primordial,BornLater
  Std         JobAndMark,User,Duration,Hosts
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
  GpuMem      GpuMemAvgGB,GpuMemPeakGB
  RelativeGpuMem \
              RelativeGpuMemAvgPct,RelativeGpuMemPeakPct
  OccupiedRelativeGpuMem \
              OccupiedRelativeGpuMemAvgPct,OccupiedRelativeGpuMemPeakPct

  default     std,cpu,mem,gpu,gpumem,cmd
  Default     Std,Cpu,Mem,Gpu,GpuMem,Cmd

DEFAULTS default

ELBAT*/

func (jc *JobsCommand) printJobSummaries(out io.Writer, summaries []*jobSummary) error {
	summaries, err := ApplyQuery(jc.ParsedQuery, jobsFormatters, jobsPredicates, summaries)
	if err != nil {
		return err
	}

	// Sort ascending by lowest beginning timestamp, and if those are equal, by job number.
	slices.SortStableFunc(summaries, func(a, b *jobSummary) int {
		c := cmp.Compare(a.sampleJob.Start, b.sampleJob.Start)
		if c == 0 {
			c = cmp.Compare(a.sampleJob.JobId, b.sampleJob.JobId)
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
			u := summaries[i].sampleJob.User
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
