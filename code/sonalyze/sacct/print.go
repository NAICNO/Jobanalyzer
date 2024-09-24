package sacct

import (
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	. "sonalyze/command"
)

func (sc *SacctCommand) printRegularJobs(stdout io.Writer, regular []*sacctSummary) {
	FormatData(stdout, sc.PrintFields, sacctFormatters, sc.PrintOpts, regular, sc.PrintOpts.Fixed)
}

func (sc *SacctCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(sc.Fmt, sacctHelp, sacctFormatters, sacctAliases, sacctDefaultFields)
}

const sacctHelp = `
parse
  Aggregate SLURM sacct data into data about jobs and present them.
`

const sacctDefaultFields = "JobID,JobName,User,Account,rcpu,rmem"

// MT: Constant after initialization; immutable
var sacctAliases = map[string][]string{
	"default": strings.Split(sacctDefaultFields, ","),
}

type sacctCtx = bool // fixed format or not

// MT: Constant after initialization; immutable
var sacctFormatters = map[string]Formatter[*sacctSummary, sacctCtx]{
	"Start": {
		func(d *sacctSummary, _ sacctCtx) string {
			if d.main.Start == 0 {
				return "Unknown"
			}
			return time.Unix(d.main.Start, 0).UTC().Format(time.RFC3339)
		},
		"Start time of job, if any",
	},
	"End": {
		func(d *sacctSummary, _ sacctCtx) string {
			return time.Unix(d.main.End, 0).UTC().Format(time.RFC3339)
		},
		"End time of job",
	},
	"Submit": {
		func(d *sacctSummary, _ sacctCtx) string {
			return time.Unix(d.main.Submit, 0).UTC().Format(time.RFC3339)
		},
		"Submit time of job",
	},
	"RequestedCPU": {
		func(d *sacctSummary, _ sacctCtx) string {
			return fmt.Sprint(d.requestedCpu)
		},
		"Requested CPU time (elapsed * cores * nodes)",
	},
	"UsedCPU": {
		func(d *sacctSummary, _ sacctCtx) string {
			return fmt.Sprint(d.usedCpu)
		},
		"Used CPU time",
	},
	"rcpu": {
		func(d *sacctSummary, _ sacctCtx) string {
			if d.requestedCpu == 0 {
				return "0"
			}
			return fmt.Sprint(math.Round(100 * float64(d.usedCpu) / float64(d.requestedCpu)))
		},
		"Percent cpu utilization: UsedCPU/RequestedCPU*100",
	},
	"rmem": {
		func(d *sacctSummary, _ sacctCtx) string {
			if d.main.ReqMem == 0 {
				return "0"
			}
			return fmt.Sprint(math.Round(100 * float64(d.maxrss) / float64(d.main.ReqMem)))
		},
		"Percent memory utilization: MaxRSS/ReqMem*100",
	},
	"User": {
		func(d *sacctSummary, _ sacctCtx) string {
			return d.main.User.String()
		},
		"Job's user",
	},
	"JobName": {
		func(d *sacctSummary, ctx sacctCtx) string {
			s := d.main.JobName.String()
			if ctx && len(s) > 30 {
				s = s[:30]
			}
			return s
		},
		"Job name",
	},
	"State": {
		func(d *sacctSummary, _ sacctCtx) string {
			return d.main.State.String()
		},
		"Job completion state",
	},
	"Account": {
		func(d *sacctSummary, _ sacctCtx) string {
			return d.main.Account.String()
		},
		"Job's account",
	},
	"Reservation": {
		func(d *sacctSummary, _ sacctCtx) string {
			return d.main.Reservation.String()
		},
		"Job's reservation, if any",
	},
	"Layout": {
		func(d *sacctSummary, _ sacctCtx) string {
			return d.main.Layout.String()
		},
		"Job's layout, if any",
	},
	"NodeList": {
		func(d *sacctSummary, _ sacctCtx) string {
			return d.main.NodeList.String()
		},
		"Job's node list",
	},
	"JobID": {
		func(d *sacctSummary, _ sacctCtx) string {
			return fmt.Sprint(d.main.JobID)
		},
		"Primary Job ID",
	},
	"MaxRSS": {
		func(d *sacctSummary, _ sacctCtx) string {
			return fmt.Sprint(d.maxrss)
		},
		"Max resident set size (RSS) across all steps",
	},
	"ReqMem": {
		func(d *sacctSummary, _ sacctCtx) string {
			return fmt.Sprint(d.main.ReqMem)
		},
		"Raw requested memory (GB)",
	},
	"ReqCPUS": {
		func(d *sacctSummary, _ sacctCtx) string {
			return fmt.Sprint(d.main.ReqCPUS)
		},
		"Raw requested CPU cores",
	},
	"ReqNodes": {
		func(d *sacctSummary, _ sacctCtx) string {
			return fmt.Sprint(d.main.ReqNodes)
		},
		"Raw requested system nodes)",
	},
	"Elapsed": {
		func(d *sacctSummary, _ sacctCtx) string {
			return fmt.Sprint(d.main.ElapsedRaw)
		},
		"Time elapsed",
	},
	"Suspended": {
		func(d *sacctSummary, _ sacctCtx) string {
			return fmt.Sprint(d.main.Suspended)
		},
		"Time suspended",
	},
	"Timelimit": {
		func(d *sacctSummary, _ sacctCtx) string {
			return fmt.Sprint(d.main.TimelimitRaw)
		},
		"Time limit in seconds",
	},
	"ExitCode": {
		func(d *sacctSummary, _ sacctCtx) string {
			return fmt.Sprint(d.main.ExitCode)
		},
		"Exit code",
	},
	"Wait": {
		func(d *sacctSummary, _ sacctCtx) string {
			if d.main.Start == 0 {
				return "0"
			}
			return fmt.Sprint(d.main.Start - d.main.Submit)
		},
		"Wait time of job (start - submit), in seconds",
	},
	"Partition": {
		func(d *sacctSummary, _ sacctCtx) string {
			return d.main.Partition.String()
		},
		"Requested partition",
	},
	"ReqGPUS": {
		func(d *sacctSummary, _ sacctCtx) string {
			return d.main.ReqGPUS.String()
		},
		"Raw requested GPU cards",
	},
	"ArrayJobID": {
		func(d *sacctSummary, _ sacctCtx) string {
			return fmt.Sprint(d.main.ArrayJobID)
		},
		"ID of the overarching array job",
	},
	"ArrayIndex": {
		func(d *sacctSummary, _ sacctCtx) string {
			return fmt.Sprint(d.main.ArrayIndex)
		},
		"Index of this job within an array job",
	},
}
