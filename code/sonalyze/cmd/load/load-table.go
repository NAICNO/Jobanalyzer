// DO NOT EDIT.  Generated from print.go by generate-table

package load

import (
	"go-utils/gpuset"
	. "sonalyze/common"
	. "sonalyze/table"
)

type GpuSet = gpuset.GpuSet

// MT: Constant after initialization; immutable
var loadFormatters = map[string]Formatter[*ReportRecord]{
	"Now": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatDateTimeValue(DateTimeValue(d.Now), ctx)
		},
		Help: "The current time (yyyy-mm-dd hh:mm)",
	},
	"DateTime": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatDateTimeValue(DateTimeValue(d.DateTime), ctx)
		},
		Help: "The starting date and time of the aggregation window (yyyy-mm-dd hh:mm)",
	},
	"Date": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatDateValue(DateValue(d.Date), ctx)
		},
		Help: "The starting date of the aggregation window (yyyy-mm-dd)",
	},
	"Time": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatTimeValue(TimeValue(d.Time), ctx)
		},
		Help: "The startint time of the aggregation window (hh:mm)",
	},
	"Cpu": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatInt(int(d.Cpu), ctx)
		},
		Help: "Average CPU utilization in percent in the aggregation window (100% = 1 core)",
	},
	"RelativeCpu": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatInt(int(d.RelativeCpu), ctx)
		},
		Help:        "Average relative CPU utilization in percent in the aggregation window (100% = all cores)",
		NeedsConfig: true,
	},
	"VirtualGB": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatInt(int(d.VirtualGB), ctx)
		},
		Help: "Average virtual memory utilization in GiB in the aggregation window",
	},
	"RelativeVirtualMem": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatInt(int(d.RelativeVirtualMem), ctx)
		},
		Help:        "Relative virtual memory utilization in GiB in the aggregation window (100% = system RAM)",
		NeedsConfig: true,
	},
	"ResidentGB": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatInt(int(d.ResidentGB), ctx)
		},
		Help: "Average resident memory utilization in GiB in the aggregation window",
	},
	"RelativeResidentMem": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatInt(int(d.RelativeResidentMem), ctx)
		},
		Help:        "Relative resident memory utilization in GiB in the aggregation window (100% = system RAM)",
		NeedsConfig: true,
	},
	"Gpu": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatInt(int(d.Gpu), ctx)
		},
		Help: "Average GPU utilization in percent in the aggregation window (100% = 1 card)",
	},
	"RelativeGpu": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatInt(int(d.RelativeGpu), ctx)
		},
		Help:        "Average relative GPU utilization in percent in the aggregation window (100% = all cards)",
		NeedsConfig: true,
	},
	"GpuGB": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatInt(int(d.GpuGB), ctx)
		},
		Help: "Average gpu memory utilization in GiB in the aggregation window",
	},
	"RelativeGpuMem": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatInt(int(d.RelativeGpuMem), ctx)
		},
		Help:        "Average relative gpu memory utilization in GiB in the aggregation window (100% = all GPU RAM)",
		NeedsConfig: true,
	},
	"Gpus": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatGpuSet(GpuSet(d.Gpus), ctx)
		},
		Help: "GPU device numbers used by the job, 'none' if none or 'unknown' in error states",
	},
	"Hostname": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatUstr(Ustr(d.Hostname), ctx)
		},
		Help: "Combined host names of jobs active in the aggregation window",
	},
}

func init() {
	DefAlias(loadFormatters, "Now", "now")
	DefAlias(loadFormatters, "DateTime", "datetime")
	DefAlias(loadFormatters, "Date", "date")
	DefAlias(loadFormatters, "Time", "time")
	DefAlias(loadFormatters, "Cpu", "cpu")
	DefAlias(loadFormatters, "RelativeCpu", "rcpu")
	DefAlias(loadFormatters, "VirtualGB", "mem")
	DefAlias(loadFormatters, "RelativeVirtualMem", "rmem")
	DefAlias(loadFormatters, "ResidentGB", "res")
	DefAlias(loadFormatters, "RelativeResidentMem", "rres")
	DefAlias(loadFormatters, "Gpu", "gpu")
	DefAlias(loadFormatters, "RelativeGpu", "rgpu")
	DefAlias(loadFormatters, "GpuGB", "gpumem")
	DefAlias(loadFormatters, "RelativeGpuMem", "rgpumem")
	DefAlias(loadFormatters, "Gpus", "gpus")
	DefAlias(loadFormatters, "Hostname", "host")
}

type ReportRecord struct {
	Now                 DateTimeValue
	DateTime            DateTimeValue
	Date                DateValue
	Time                TimeValue
	Cpu                 int
	RelativeCpu         int
	VirtualGB           int
	RelativeVirtualMem  int
	ResidentGB          int
	RelativeResidentMem int
	Gpu                 int
	RelativeGpu         int
	GpuGB               int
	RelativeGpuMem      int
	Gpus                GpuSet
	Hostname            Ustr
}

const loadHelp = `
load
  Aggregate process data across users and commands on a host and bucket into
  time slots, producing a view of system load.  Default output format is 'fixed'.
`

func (c *LoadCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(c.Fmt, loadHelp, loadFormatters, loadAliases, loadDefaultFields)
}

// MT: Constant after initialization; immutable
var loadAliases = map[string][]string{
	"default": []string{"date", "time", "cpu", "mem", "gpu", "gpumem", "gpumask"},
	"Default": []string{"Date", "Time", "Cpu", "ResidentGB", "Gpu", "GpuGB", "Gpus"},
}

const loadDefaultFields = "default"
