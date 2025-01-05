package sacct

import (
	"io"
	"math"

	. "sonalyze/table"
)

//go:generate ../../../generate-table/generate-table -o sacct-table.go print.go

/*TABLE sacct

package sacct

%%

FIELDS *SacctRegular

 Start               IsoDateTimeOrUnknown desc:"Start time of job, if any"
 End                 IsoDateTimeOrUnknown desc:"End time of job"
 Submit              IsoDateTimeOrUnknown desc:"Submit time of job"
 RequestedCPU        int                  desc:"Requested CPU time (elapsed * cores * nodes)"
 UsedCPU             int                  desc:"Used CPU time"
 RelativeCPU         int                  alias:"rcpu" desc:"Percent cpu utilization: UsedCPU/RequestedCPU*100"
 RelativeResidentMem int                  alias:"rmem" desc:"Percent memory utilization: MaxRSS/ReqMem*100"
 User                Ustr                 desc:"Job's user"
 JobName             UstrMax30            desc:"Job name"
 State               Ustr                 desc:"Job completion state"
 Account             Ustr                 desc:"Job's account"
 Reservation         Ustr                 desc:"Job's reservation, if any"
 Layout              Ustr                 desc:"Job's layout, if any"
 NodeList            Ustr                 desc:"Job's node list"
 JobID               int                  desc:"Primary Job ID"
 MaxRSS              int                  desc:"Max resident set size (RSS) across all steps (GB)"
 ReqMem              int                  desc:"Raw requested memory (GB)"
 ReqCPUS             int                  desc:"Raw requested CPU cores"
 ReqGPUS             Ustr                 desc:"Raw requested GPU cards"
 ReqNodes            int                  desc:"Raw requested system nodes"
 Elapsed             int                  desc:"Time elapsed"
 Suspended           int                  desc:"Time suspended"
 Timelimit           int                  desc:"Time limit in seconds"
 ExitCode            int                  desc:"Exit code"
 Wait                int                  desc:"Wait time of job (start - submit), in seconds"
 Partition           Ustr                 desc:"Requested partition"
 ArrayJobID          int                  desc:"ID of the overarching array job"
 ArrayIndex          int                  desc:"Index of this job within an array job"

GENERATE SacctRegular

SUMMARY SacctCommand

Experimental: Extract information from sacct data independent of sample data.

Data are extracted by sacct for completed jobs on a cluster and stored
in Jobanalyzer's database.  These data can be queried by "sonalyze
sacct".  The fields are generally the same as those of the sacct
output, and have the meaning defined by sacct.

HELP SacctCommand

  Aggregate SLURM sacct data into data about jobs and present them.

ALIASES

  default JobID,JobName,User,Account,rcpu,rmem
  Default JobID,JobName,User,Account,RelativeCPU,RelativeResidentMem

DEFAULTS default

ELBAT*/

func (sc *SacctCommand) printRegularJobs(stdout io.Writer, regular []*sacctSummary) error {
	// TODO: By and by it may be possible to lift this extra loop into the loop already being run in
	// perform.go to compute the `regular` values, and not allocate extra values here.
	toPrint := make([]*SacctRegular, len(regular))
	for i, r := range regular {
		var relativeCpu, relativeResidentMem int
		var waitTime int64
		if r.requestedCpu > 0 {
			relativeCpu = int(math.Round(100 * float64(r.usedCpu) / float64(r.requestedCpu)))
		}
		if r.Main.ReqMem > 0 {
			relativeResidentMem = int(math.Round(100 * float64(r.maxrss) / float64(r.Main.ReqMem)))
		}
		if r.Main.Start > 0 {
			waitTime = r.Main.Start - r.Main.Submit
		}
		toPrint[i] = &SacctRegular{
			Start:               IsoDateTimeOrUnknown(r.Main.Start),
			End:                 IsoDateTimeOrUnknown(r.Main.End),
			Submit:              IsoDateTimeOrUnknown(r.Main.Submit),
			RequestedCPU:        int(r.requestedCpu),
			UsedCPU:             int(r.usedCpu),
			RelativeCPU:         relativeCpu,
			RelativeResidentMem: relativeResidentMem,
			User:                r.Main.User,
			JobName:             UstrMax30(r.Main.JobName),
			State:               r.Main.State,
			Account:             r.Main.Account,
			Reservation:         r.Main.Reservation,
			Layout:              r.Main.Layout,
			NodeList:            r.Main.NodeList,
			JobID:               int(r.Main.JobID),
			MaxRSS:              int(r.maxrss),
			ReqMem:              int(r.Main.ReqMem),
			ReqCPUS:             int(r.Main.ReqCPUS),
			ReqNodes:            int(r.Main.ReqNodes),
			Elapsed:             int(r.Main.ElapsedRaw),
			Suspended:           int(r.Main.Suspended),
			Timelimit:           int(r.Main.TimelimitRaw),
			ExitCode:            int(r.Main.ExitCode),
			Wait:                int(waitTime),
			Partition:           r.Main.Partition,
			ReqGPUS:             r.Main.ReqGPUS,
			ArrayJobID:          int(r.Main.ArrayJobID),
			ArrayIndex:          int(r.Main.ArrayIndex),
		}
	}
	FormatData(
		stdout,
		sc.PrintFields,
		sacctFormatters,
		sc.PrintOpts,
		toPrint,
	)
	return nil
}
