// DO NOT EDIT.  Generated from print.go by generate-table

package sacct

import (
	. "sonalyze/common"
	. "sonalyze/table"
)

// MT: Constant after initialization; immutable
var sacctFormatters = map[string]Formatter[*SacctRegular]{
	"Start": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatIsoDateTimeOrUnknown(IsoDateTimeOrUnknown(d.Start), ctx)
		},
		Help: "Start time of job, if any",
	},
	"End": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatIsoDateTimeOrUnknown(IsoDateTimeOrUnknown(d.End), ctx)
		},
		Help: "End time of job",
	},
	"Submit": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatIsoDateTimeOrUnknown(IsoDateTimeOrUnknown(d.Submit), ctx)
		},
		Help: "Submit time of job",
	},
	"RequestedCPU": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.RequestedCPU), ctx)
		},
		Help: "Requested CPU time (elapsed * cores * nodes)",
	},
	"UsedCPU": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.UsedCPU), ctx)
		},
		Help: "Used CPU time",
	},
	"RelativeCPU": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.RelativeCPU), ctx)
		},
		Help:        "Percent cpu utilization: UsedCPU/RequestedCPU*100",
		NeedsConfig: true,
	},
	"RelativeResidentMem": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.RelativeResidentMem), ctx)
		},
		Help:        "Percent memory utilization: MaxRSS/ReqMem*100",
		NeedsConfig: true,
	},
	"User": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstr(Ustr(d.User), ctx)
		},
		Help: "Job's user",
	},
	"JobName": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstrMax30(UstrMax30(d.JobName), ctx)
		},
		Help: "Job name",
	},
	"State": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstr(Ustr(d.State), ctx)
		},
		Help: "Job completion state",
	},
	"Account": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstr(Ustr(d.Account), ctx)
		},
		Help: "Job's account",
	},
	"Reservation": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstr(Ustr(d.Reservation), ctx)
		},
		Help: "Job's reservation, if any",
	},
	"Layout": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstr(Ustr(d.Layout), ctx)
		},
		Help: "Job's layout, if any",
	},
	"NodeList": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstr(Ustr(d.NodeList), ctx)
		},
		Help: "Job's node list",
	},
	"JobID": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.JobID), ctx)
		},
		Help: "Primary Job ID",
	},
	"MaxRSS": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.MaxRSS), ctx)
		},
		Help: "Max resident set size (RSS) across all steps (GB)",
	},
	"ReqMem": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.ReqMem), ctx)
		},
		Help: "Raw requested memory (GB)",
	},
	"ReqCPUS": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.ReqCPUS), ctx)
		},
		Help: "Raw requested CPU cores",
	},
	"ReqGPUS": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstr(Ustr(d.ReqGPUS), ctx)
		},
		Help: "Raw requested GPU cards",
	},
	"ReqNodes": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.ReqNodes), ctx)
		},
		Help: "Raw requested system nodes",
	},
	"Elapsed": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.Elapsed), ctx)
		},
		Help: "Time elapsed",
	},
	"Suspended": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.Suspended), ctx)
		},
		Help: "Time suspended",
	},
	"Timelimit": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.Timelimit), ctx)
		},
		Help: "Time limit in seconds",
	},
	"ExitCode": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.ExitCode), ctx)
		},
		Help: "Exit code",
	},
	"Wait": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.Wait), ctx)
		},
		Help: "Wait time of job (start - submit), in seconds",
	},
	"Partition": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatUstr(Ustr(d.Partition), ctx)
		},
		Help: "Requested partition",
	},
	"ArrayJobID": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.ArrayJobID), ctx)
		},
		Help: "ID of the overarching array job",
	},
	"ArrayIndex": {
		Fmt: func(d *SacctRegular, ctx PrintMods) string {
			return FormatInt(int(d.ArrayIndex), ctx)
		},
		Help: "Index of this job within an array job",
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
