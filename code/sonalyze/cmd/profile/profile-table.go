// DO NOT EDIT.  Generated from print.go by generate-table

package profile

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
var profileFormatters = map[string]Formatter[*fixedLine]{
	"Timestamp": {
		Fmt: func(d *fixedLine, ctx PrintMods) string {
			return FormatDateTimeValueOrBlank((d.Timestamp), ctx)
		},
		Help: "(DateTimeValue) Time of the start of the profiling bucket",
	},
	"Hostname": {
		Fmt: func(d *fixedLine, ctx PrintMods) string {
			return FormatUstr((d.Hostname), ctx)
		},
		Help: "(string) Host on which process ran",
	},
	"CpuUtilPct": {
		Fmt: func(d *fixedLine, ctx PrintMods) string {
			return FormatInt((d.CpuUtilPct), ctx)
		},
		Help: "(int) CPU utilization in percent, 100% = 1 core (except for HTML)",
	},
	"VirtualMemGB": {
		Fmt: func(d *fixedLine, ctx PrintMods) string {
			return FormatInt((d.VirtualMemGB), ctx)
		},
		Help: "(int) Main virtual memory usage in GiB",
	},
	"ResidentMemGB": {
		Fmt: func(d *fixedLine, ctx PrintMods) string {
			return FormatInt((d.ResidentMemGB), ctx)
		},
		Help: "(int) Main resident memory usage in GiB",
	},
	"Gpu": {
		Fmt: func(d *fixedLine, ctx PrintMods) string {
			return FormatInt((d.Gpu), ctx)
		},
		Help: "(int) GPU utilization in percent, 100% = 1 card (except for HTML)",
	},
	"GpuMemGB": {
		Fmt: func(d *fixedLine, ctx PrintMods) string {
			return FormatInt((d.GpuMemGB), ctx)
		},
		Help: "(int) GPU resident memory usage in GiB (across all cards)",
	},
	"Command": {
		Fmt: func(d *fixedLine, ctx PrintMods) string {
			return FormatUstr((d.Command), ctx)
		},
		Help: "(string) Name of executable starting the process",
	},
	"NumProcs": {
		Fmt: func(d *fixedLine, ctx PrintMods) string {
			return FormatIntOrEmpty((d.NumProcs), ctx)
		},
		Help: "(int) Number of rolled-up processes, blank for zero",
	},
}

func init() {
	DefAlias(profileFormatters, "Timestamp", "time")
	DefAlias(profileFormatters, "Hostname", "host")
	DefAlias(profileFormatters, "CpuUtilPct", "cpu")
	DefAlias(profileFormatters, "VirtualMemGB", "mem")
	DefAlias(profileFormatters, "ResidentMemGB", "res")
	DefAlias(profileFormatters, "ResidentMemGB", "rss")
	DefAlias(profileFormatters, "Gpu", "gpu")
	DefAlias(profileFormatters, "GpuMemGB", "gpumem")
	DefAlias(profileFormatters, "Command", "cmd")
	DefAlias(profileFormatters, "NumProcs", "nproc")
}

// MT: Constant after initialization; immutable
var profilePredicates = map[string]Predicate[*fixedLine]{
	"Timestamp": Predicate[*fixedLine]{
		Convert: CvtString2DateTimeValue,
		Compare: func(d *fixedLine, v any) int {
			return cmp.Compare((d.Timestamp), v.(DateTimeValueOrBlank))
		},
	},
	"CpuUtilPct": Predicate[*fixedLine]{
		Convert: CvtString2Int,
		Compare: func(d *fixedLine, v any) int {
			return cmp.Compare((d.CpuUtilPct), v.(int))
		},
	},
	"VirtualMemGB": Predicate[*fixedLine]{
		Convert: CvtString2Int,
		Compare: func(d *fixedLine, v any) int {
			return cmp.Compare((d.VirtualMemGB), v.(int))
		},
	},
	"ResidentMemGB": Predicate[*fixedLine]{
		Convert: CvtString2Int,
		Compare: func(d *fixedLine, v any) int {
			return cmp.Compare((d.ResidentMemGB), v.(int))
		},
	},
	"Gpu": Predicate[*fixedLine]{
		Convert: CvtString2Int,
		Compare: func(d *fixedLine, v any) int {
			return cmp.Compare((d.Gpu), v.(int))
		},
	},
	"GpuMemGB": Predicate[*fixedLine]{
		Convert: CvtString2Int,
		Compare: func(d *fixedLine, v any) int {
			return cmp.Compare((d.GpuMemGB), v.(int))
		},
	},
	"Command": Predicate[*fixedLine]{
		Convert: CvtString2Ustr,
		Compare: func(d *fixedLine, v any) int {
			return cmp.Compare((d.Command), v.(Ustr))
		},
	},
	"NumProcs": Predicate[*fixedLine]{
		Convert: CvtString2Int,
		Compare: func(d *fixedLine, v any) int {
			return cmp.Compare((d.NumProcs), v.(IntOrEmpty))
		},
	},
}

type fixedLine struct {
	Timestamp     DateTimeValueOrBlank
	Hostname      Ustr
	CpuUtilPct    int
	VirtualMemGB  int
	ResidentMemGB int
	Gpu           int
	GpuMemGB      int
	Command       Ustr
	NumProcs      IntOrEmpty
}

func (c *ProfileCommand) Summary(out io.Writer) {
	fmt.Fprint(out, `Experimental: Print profile information for one aspect of a particular job.

This prints a table across time of utilization of various resources of the
processes in a job.  The job can be on multiple nodes.  For fixed formatting,
all resources are printed on one line per process per time step; similarly for
json all resources for a process at a time step are embedded in a single object.
For CSV, AWK and HTML output a single resource must be selected with -fmt, and
its utilization across processes per time step is printed; start with the CSV
output to understand this (eg -fmt csv,gpu will show the table for the gpu
resource in CSV form).  Commands, process IDs and host names are encoded in the
output header in an idiosyncratic, but useful, form.  Note that no header is
printed by default for AWK.  Be sure to file bugs for missing functionality.
`)
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
