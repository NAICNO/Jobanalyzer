// DO NOT EDIT.  Generated from print.go by generate-table

package load

import "go-utils/gpuset"

import (
	"cmp"
	"fmt"
	"io"
	. "sonalyze/common"
	. "sonalyze/table"
)

var (
	_ = cmp.Compare(0, 0)
	_ fmt.Formatter
	_ = io.SeekStart
	_ = UstrEmpty
)

// MT: Constant after initialization; immutable
var loadFormatters = map[string]Formatter[*ReportRecord]{
	"Now": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatDateTimeValue(d.Now, ctx)
		},
		Help: "(DateTimeValue) The current time",
	},
	"DateTime": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatDateTimeValue(d.DateTime, ctx)
		},
		Help: "(DateTimeValue) The starting date and time of the aggregation window",
	},
	"Date": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatDateValue(d.Date, ctx)
		},
		Help: "(DateValue) The starting date of the aggregation window",
	},
	"Time": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatTimeValue(d.Time, ctx)
		},
		Help: "(TimeValue) The startint time of the aggregation window",
	},
	"Cpu": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatInt(d.Cpu, ctx)
		},
		Help: "(int) Average CPU utilization in percent in the aggregation window (100% = 1 core)",
	},
	"RelativeCpu": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatInt(d.RelativeCpu, ctx)
		},
		Help:        "(int) Average relative CPU utilization in percent in the aggregation window (100% = all cores)",
		NeedsConfig: true,
	},
	"VirtualGB": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatInt(d.VirtualGB, ctx)
		},
		Help: "(int) Average virtual memory utilization in GiB in the aggregation window",
	},
	"RelativeVirtualMem": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatInt(d.RelativeVirtualMem, ctx)
		},
		Help:        "(int) Relative virtual memory utilization in GiB in the aggregation window (100% = system RAM)",
		NeedsConfig: true,
	},
	"ResidentGB": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatInt(d.ResidentGB, ctx)
		},
		Help: "(int) Average resident memory utilization in GiB in the aggregation window",
	},
	"RelativeResidentMem": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatInt(d.RelativeResidentMem, ctx)
		},
		Help:        "(int) Relative resident memory utilization in GiB in the aggregation window (100% = system RAM)",
		NeedsConfig: true,
	},
	"Gpu": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatInt(d.Gpu, ctx)
		},
		Help: "(int) Average GPU utilization in percent in the aggregation window (100% = 1 card)",
	},
	"RelativeGpu": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatInt(d.RelativeGpu, ctx)
		},
		Help:        "(int) Average relative GPU utilization in percent in the aggregation window (100% = all cards)",
		NeedsConfig: true,
	},
	"GpuGB": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatInt(d.GpuGB, ctx)
		},
		Help: "(int) Average gpu memory utilization in GiB in the aggregation window",
	},
	"RelativeGpuMem": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatInt(d.RelativeGpuMem, ctx)
		},
		Help:        "(int) Average relative gpu memory utilization in GiB in the aggregation window (100% = all GPU RAM)",
		NeedsConfig: true,
	},
	"Gpus": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatGpuSet(d.Gpus, ctx)
		},
		Help: "(GpuSet) GPU device numbers used by the job, 'none' if none or 'unknown' in error states",
	},
	"Hostname": {
		Fmt: func(d *ReportRecord, ctx PrintMods) string {
			return FormatUstr(d.Hostname, ctx)
		},
		Help: "(string) Combined host names of jobs active in the aggregation window",
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
	Gpus                gpuset.GpuSet
	Hostname            Ustr
}

func (c *LoadCommand) Summary(out io.Writer) {
	fmt.Fprint(out, `Compute aggregate load for hosts and groups of hosts.

The aggregation can be performed across various timeframes and will be
based on available sample data.

As not all processes' samples are stored (processes typically have to
be "significant"), the true load can be underreported somewhat, but
probably not by very much.
`)
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
