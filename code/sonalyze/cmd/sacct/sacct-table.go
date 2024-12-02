// DO NOT EDIT.  Generated from print.go by generate-table

package sacct

import (
	. "sonalyze/common"
	. "sonalyze/table"
)

import (
	"fmt"
	"io"
)

var (
	_ fmt.Formatter
	_ = io.SeekStart
)

// MT: Constant after initialization; immutable
var sacctFormatters = map[string]Formatter[*SacctRegular]{
	"Start": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatIsoDateTimeOrUnknown(IsoDateTimeOrUnknown(d.Start), ctx)
		},
		Help: "(IsoDateTimeValue) Start time of job, if any",
	},
	"End": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatIsoDateTimeOrUnknown(IsoDateTimeOrUnknown(d.End), ctx)
		},
		Help: "(IsoDateTimeValue) End time of job",
	},
	"Submit": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatIsoDateTimeOrUnknown(IsoDateTimeOrUnknown(d.Submit), ctx)
		},
		Help: "(IsoDateTimeValue) Submit time of job",
	},
	"RequestedCPU": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.RequestedCPU), ctx)
		},
		Help: "(int) Requested CPU time (elapsed * cores * nodes)",
	},
	"UsedCPU": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.UsedCPU), ctx)
		},
		Help: "(int) Used CPU time",
	},
	"RelativeCPU": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.RelativeCPU), ctx)
		},
		Help:        "(int) Percent cpu utilization: UsedCPU/RequestedCPU*100",
		NeedsConfig: true,
	},
	"RelativeResidentMem": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.RelativeResidentMem), ctx)
		},
		Help:        "(int) Percent memory utilization: MaxRSS/ReqMem*100",
		NeedsConfig: true,
	},
	"User": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstr(Ustr(d.User), ctx)
		},
		Help: "(string) Job's user",
	},
	"JobName": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstrMax30(UstrMax30(d.JobName), ctx)
		},
		Help: "(string) Job name",
	},
	"State": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstr(Ustr(d.State), ctx)
		},
		Help: "(string) Job completion state",
	},
	"Account": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstr(Ustr(d.Account), ctx)
		},
		Help: "(string) Job's account",
	},
	"Reservation": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstr(Ustr(d.Reservation), ctx)
		},
		Help: "(string) Job's reservation, if any",
	},
	"Layout": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstr(Ustr(d.Layout), ctx)
		},
		Help: "(string) Job's layout, if any",
	},
	"NodeList": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstr(Ustr(d.NodeList), ctx)
		},
		Help: "(string) Job's node list",
	},
	"JobID": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.JobID), ctx)
		},
		Help: "(int) Primary Job ID",
	},
	"MaxRSS": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.MaxRSS), ctx)
		},
		Help: "(int) Max resident set size (RSS) across all steps (GB)",
	},
	"ReqMem": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.ReqMem), ctx)
		},
		Help: "(int) Raw requested memory (GB)",
	},
	"ReqCPUS": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.ReqCPUS), ctx)
		},
		Help: "(int) Raw requested CPU cores",
	},
	"ReqGPUS": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstr(Ustr(d.ReqGPUS), ctx)
		},
		Help: "(string) Raw requested GPU cards",
	},
	"ReqNodes": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.ReqNodes), ctx)
		},
		Help: "(int) Raw requested system nodes",
	},
	"Elapsed": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.Elapsed), ctx)
		},
		Help: "(int) Time elapsed",
	},
	"Suspended": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.Suspended), ctx)
		},
		Help: "(int) Time suspended",
	},
	"Timelimit": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.Timelimit), ctx)
		},
		Help: "(int) Time limit in seconds",
	},
	"ExitCode": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.ExitCode), ctx)
		},
		Help: "(int) Exit code",
	},
	"Wait": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.Wait), ctx)
		},
		Help: "(int) Wait time of job (start - submit), in seconds",
	},
	"Partition": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstr(Ustr(d.Partition), ctx)
		},
		Help: "(string) Requested partition",
	},
	"ArrayJobID": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.ArrayJobID), ctx)
		},
		Help: "(int) ID of the overarching array job",
	},
	"ArrayIndex": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.ArrayIndex), ctx)
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
