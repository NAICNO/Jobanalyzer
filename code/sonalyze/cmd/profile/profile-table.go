// DO NOT EDIT.  Generated from print.go by generate-table

package profile

import (
	. "sonalyze/common"
	. "sonalyze/table"
)

// MT: Constant after initialization; immutable
var profileFormatters = map[string]Formatter[*fixedLine]{
	"Timestamp": {
		Fmt: func(d *fixedLine, ctx PrintMods) string {
			return FormatDateTimeValueOrBlank(DateTimeValueOrBlank(d.Timestamp), ctx)
		},
		Help: "(DateTimeValue) Time of the start of the profiling bucket",
	},
	"CpuUtilPct": {
		Fmt: func(d *fixedLine, ctx PrintMods) string {
			return FormatInt(int(d.CpuUtilPct), ctx)
		},
		Help: "(int) CPU utilization in percent, 100% = 1 core (except for HTML)",
	},
	"VirtualMemGB": {
		Fmt: func(d *fixedLine, ctx PrintMods) string {
			return FormatInt(int(d.VirtualMemGB), ctx)
		},
		Help: "(int) Main virtual memory usage in GiB",
	},
	"ResidentMemGB": {
		Fmt: func(d *fixedLine, ctx PrintMods) string {
			return FormatInt(int(d.ResidentMemGB), ctx)
		},
		Help: "(int) Main resident memory usage in GiB",
	},
	"Gpu": {
		Fmt: func(d *fixedLine, ctx PrintMods) string {
			return FormatInt(int(d.Gpu), ctx)
		},
		Help: "(int) GPU utilization in percent, 100% = 1 card (except for HTML)",
	},
	"GpuMemGB": {
		Fmt: func(d *fixedLine, ctx PrintMods) string {
			return FormatInt(int(d.GpuMemGB), ctx)
		},
		Help: "(int) GPU resident memory usage in GiB (across all cards)",
	},
	"Command": {
		Fmt: func(d *fixedLine, ctx PrintMods) string {
			return FormatUstr(Ustr(d.Command), ctx)
		},
		Help: "(string) Name of executable starting the process",
	},
	"NumProcs": {
		Fmt: func(d *fixedLine, ctx PrintMods) string {
			return FormatIntOrEmpty(IntOrEmpty(d.NumProcs), ctx)
		},
		Help: "(int) Number of rolled-up processes, blank for zero",
	},
}

func init() {
	DefAlias(profileFormatters, "Timestamp", "time")
	DefAlias(profileFormatters, "CpuUtilPct", "cpu")
	DefAlias(profileFormatters, "VirtualMemGB", "mem")
	DefAlias(profileFormatters, "ResidentMemGB", "res")
	DefAlias(profileFormatters, "ResidentMemGB", "rss")
	DefAlias(profileFormatters, "Gpu", "gpu")
	DefAlias(profileFormatters, "GpuMemGB", "gpumem")
	DefAlias(profileFormatters, "Command", "cmd")
	DefAlias(profileFormatters, "NumProcs", "nproc")
}

type fixedLine struct {
	Timestamp     DateTimeValueOrBlank
	CpuUtilPct    int
	VirtualMemGB  int
	ResidentMemGB int
	Gpu           int
	GpuMemGB      int
	Command       Ustr
	NumProcs      IntOrEmpty
}

const profileHelp = `
profile
  Compute aggregate job behavior across processes by time step, for some job
  attributes.  Default output format is 'fixed'.
`

func (c *ProfileCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(c.Fmt, profileHelp, profileFormatters, profileAliases, profileDefaultFields)
}

// MT: Constant after initialization; immutable
var profileAliases = map[string][]string{
	"default": []string{"time", "cpu", "mem", "gpu", "gpumem", "cmd"},
	"Default": []string{"Timestamp", "CpuUtilPct", "VirtualMemGB", "Gpu", "GpuMemGB", "Command"},
}

const profileDefaultFields = "default"
