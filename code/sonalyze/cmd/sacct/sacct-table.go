// DO NOT EDIT.  Generated from print.go by generate-table

package sacct

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
var sacctFormatters = map[string]Formatter[*SacctRegular]{
	"Start": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatIsoDateTimeOrUnknown((d.Start), ctx)
		},
		Help: "(IsoDateTimeValue) Start time of job, if any",
	},
	"End": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatIsoDateTimeOrUnknown((d.End), ctx)
		},
		Help: "(IsoDateTimeValue) End time of job",
	},
	"Submit": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatIsoDateTimeOrUnknown((d.Submit), ctx)
		},
		Help: "(IsoDateTimeValue) Submit time of job",
	},
	"RequestedCPU": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt((d.RequestedCPU), ctx)
		},
		Help: "(int) Requested CPU time (elapsed * cores * nodes)",
	},
	"UsedCPU": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt((d.UsedCPU), ctx)
		},
		Help: "(int) Used CPU time",
	},
	"RelativeCPU": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt((d.RelativeCPU), ctx)
		},
		Help:        "(int) Percent cpu utilization: UsedCPU/RequestedCPU*100",
		NeedsConfig: true,
	},
	"RelativeResidentMem": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt((d.RelativeResidentMem), ctx)
		},
		Help:        "(int) Percent memory utilization: MaxRSS/ReqMem*100",
		NeedsConfig: true,
	},
	"User": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstr((d.User), ctx)
		},
		Help: "(string) Job's user",
	},
	"JobName": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstrMax30((d.JobName), ctx)
		},
		Help: "(string) Job name",
	},
	"State": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstr((d.State), ctx)
		},
		Help: "(string) Job completion state",
	},
	"Account": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstr((d.Account), ctx)
		},
		Help: "(string) Job's account",
	},
	"Reservation": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstr((d.Reservation), ctx)
		},
		Help: "(string) Job's reservation, if any",
	},
	"Layout": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstr((d.Layout), ctx)
		},
		Help: "(string) Job's layout, if any",
	},
	"NodeList": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstr((d.NodeList), ctx)
		},
		Help: "(string) Job's node list",
	},
	"JobID": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt((d.JobID), ctx)
		},
		Help: "(int) Primary Job ID",
	},
	"MaxRSS": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt((d.MaxRSS), ctx)
		},
		Help: "(int) Max resident set size (RSS) across all steps (GB)",
	},
	"ReqMem": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt((d.ReqMem), ctx)
		},
		Help: "(int) Raw requested memory (GB)",
	},
	"ReqCPUS": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt((d.ReqCPUS), ctx)
		},
		Help: "(int) Raw requested CPU cores",
	},
	"ReqGPUS": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstr((d.ReqGPUS), ctx)
		},
		Help: "(string) Raw requested GPU cards",
	},
	"ReqNodes": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt((d.ReqNodes), ctx)
		},
		Help: "(int) Raw requested system nodes",
	},
	"Elapsed": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt((d.Elapsed), ctx)
		},
		Help: "(int) Time elapsed",
	},
	"Suspended": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt((d.Suspended), ctx)
		},
		Help: "(int) Time suspended",
	},
	"Timelimit": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt((d.Timelimit), ctx)
		},
		Help: "(int) Time limit in seconds",
	},
	"ExitCode": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt((d.ExitCode), ctx)
		},
		Help: "(int) Exit code",
	},
	"Wait": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt((d.Wait), ctx)
		},
		Help: "(int) Wait time of job (start - submit), in seconds",
	},
	"Partition": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstr((d.Partition), ctx)
		},
		Help: "(string) Requested partition",
	},
	"ArrayJobID": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt((d.ArrayJobID), ctx)
		},
		Help: "(int) ID of the overarching array job",
	},
	"ArrayIndex": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt((d.ArrayIndex), ctx)
		},
		Help: "(int) Index of this job within an array job",
	},
}

func init() {
	DefAlias(sacctFormatters, "RelativeCPU", "rcpu")
	DefAlias(sacctFormatters, "RelativeResidentMem", "rmem")
}

type SacctRegular struct {
	Start               IsoDateTimeOrUnknown
	End                 IsoDateTimeOrUnknown
	Submit              IsoDateTimeOrUnknown
	RequestedCPU        int
	UsedCPU             int
	RelativeCPU         int
	RelativeResidentMem int
	User                Ustr
	JobName             UstrMax30
	State               Ustr
	Account             Ustr
	Reservation         Ustr
	Layout              Ustr
	NodeList            Ustr
	JobID               int
	MaxRSS              int
	ReqMem              int
	ReqCPUS             int
	ReqGPUS             Ustr
	ReqNodes            int
	Elapsed             int
	Suspended           int
	Timelimit           int
	ExitCode            int
	Wait                int
	Partition           Ustr
	ArrayJobID          int
	ArrayIndex          int
}

func (c *SacctCommand) Summary(out io.Writer) {
	fmt.Fprint(out, `Experimental: Extract information from sacct data independent of sample data.

Data are extracted by sacct for completed jobs on a cluster and stored
in Jobanalyzer's database.  These data can be queried by "sonalyze
sacct".  The fields are generally the same as those of the sacct
output, and have the meaning defined by sacct.
`)
}

const sacctHelp = `
sacct
  Aggregate SLURM sacct data into data about jobs and present them.
`

func (c *SacctCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(c.Fmt, sacctHelp, sacctFormatters, sacctAliases, sacctDefaultFields)
}

// MT: Constant after initialization; immutable
var sacctAliases = map[string][]string{
	"default": []string{"JobID", "JobName", "User", "Account", "rcpu", "rmem"},
	"Default": []string{"JobID", "JobName", "User", "Account", "RelativeCPU", "RelativeResidentMem"},
}

const sacctDefaultFields = "default"
