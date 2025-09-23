// DO NOT EDIT.  Generated from nodeprof.go by generate-table

package nodeprof

import "sonalyze/db/repr"

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
var nodeprofFormatters = map[string]Formatter[*repr.NodeSample]{
	"Timestamp": {
		Fmt: func(d *repr.NodeSample, ctx PrintMods) string {
			return FormatDateTimeValue((d.Timestamp), ctx)
		},
		Help: "(DateTimeValue) Full ISO timestamp of when the reading was taken",
	},
	"Hostname": {
		Fmt: func(d *repr.NodeSample, ctx PrintMods) string {
			return FormatUstr((d.Hostname), ctx)
		},
		Help: "(string) Name that host is known by on the cluster",
	},
	"UsedMemory": {
		Fmt: func(d *repr.NodeSample, ctx PrintMods) string {
			return FormatUint64((d.UsedMemory), ctx)
		},
		Help: "(uint64) Amount of memory in use",
	},
	"Load1": {
		Fmt: func(d *repr.NodeSample, ctx PrintMods) string {
			return FormatFloat64((d.Load1), ctx)
		},
		Help: "(float64) 1-minute load average",
	},
	"Load5": {
		Fmt: func(d *repr.NodeSample, ctx PrintMods) string {
			return FormatFloat64((d.Load5), ctx)
		},
		Help: "(float64) 5-minute load average",
	},
	"Load15": {
		Fmt: func(d *repr.NodeSample, ctx PrintMods) string {
			return FormatFloat64((d.Load15), ctx)
		},
		Help: "(float64) 15-minute load average",
	},
	"RunnableEntities": {
		Fmt: func(d *repr.NodeSample, ctx PrintMods) string {
			return FormatUint64((d.RunnableEntities), ctx)
		},
		Help: "(uint64) Number of runnable entities on system (threads)",
	},
	"ExistingEntities": {
		Fmt: func(d *repr.NodeSample, ctx PrintMods) string {
			return FormatUint64((d.ExistingEntities), ctx)
		},
		Help: "(uint64) Number of entities on system",
	},
}

func init() {
	DefAlias(nodeprofFormatters, "Timestamp", "timestamp")
	DefAlias(nodeprofFormatters, "Hostname", "host")
	DefAlias(nodeprofFormatters, "UsedMemory", "usedmem")
	DefAlias(nodeprofFormatters, "Load1", "load1")
	DefAlias(nodeprofFormatters, "Load5", "load5")
	DefAlias(nodeprofFormatters, "Load15", "load15")
	DefAlias(nodeprofFormatters, "RunnableEntities", "runnable")
	DefAlias(nodeprofFormatters, "ExistingEntities", "entitites")
}

// MT: Constant after initialization; immutable
var nodeprofPredicates = map[string]Predicate[*repr.NodeSample]{
	"Timestamp": Predicate[*repr.NodeSample]{
		Convert: CvtString2DateTimeValue,
		Compare: func(d *repr.NodeSample, v any) int {
			return cmp.Compare((d.Timestamp), v.(DateTimeValue))
		},
	},
	"Hostname": Predicate[*repr.NodeSample]{
		Convert: CvtString2Ustr,
		Compare: func(d *repr.NodeSample, v any) int {
			return cmp.Compare((d.Hostname), v.(Ustr))
		},
	},
	"UsedMemory": Predicate[*repr.NodeSample]{
		Convert: CvtString2Uint64,
		Compare: func(d *repr.NodeSample, v any) int {
			return cmp.Compare((d.UsedMemory), v.(uint64))
		},
	},
	"Load1": Predicate[*repr.NodeSample]{
		Convert: CvtString2Float64,
		Compare: func(d *repr.NodeSample, v any) int {
			return cmp.Compare((d.Load1), v.(float64))
		},
	},
	"Load5": Predicate[*repr.NodeSample]{
		Convert: CvtString2Float64,
		Compare: func(d *repr.NodeSample, v any) int {
			return cmp.Compare((d.Load5), v.(float64))
		},
	},
	"Load15": Predicate[*repr.NodeSample]{
		Convert: CvtString2Float64,
		Compare: func(d *repr.NodeSample, v any) int {
			return cmp.Compare((d.Load15), v.(float64))
		},
	},
	"RunnableEntities": Predicate[*repr.NodeSample]{
		Convert: CvtString2Uint64,
		Compare: func(d *repr.NodeSample, v any) int {
			return cmp.Compare((d.RunnableEntities), v.(uint64))
		},
	},
	"ExistingEntities": Predicate[*repr.NodeSample]{
		Convert: CvtString2Uint64,
		Compare: func(d *repr.NodeSample, v any) int {
			return cmp.Compare((d.ExistingEntities), v.(uint64))
		},
	},
}

func (c *NodeProfCommand) Summary(out io.Writer) {
	fmt.Fprint(out, `  Display node sample data - memory usage, load averages, and run queue length.
`)
}

const nodeprofHelp = `
nodeprof
  Extract node profiling data from sample data and present it in primitive form.  Output
  records are sorted by node name and time.  The default format is 'fixed'.

  RunnableEntities is the length of the run queue on the node.  If the number of runnable
  entities is much higher than the available number of cores for the job then the job may
  be starved for CPU.  Note however that the run queue is per-node while the amount of CPU
  available may be per-job; the data will require careful interpretation.
`

func (c *NodeProfCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(c.Fmt, nodeprofHelp, nodeprofFormatters, nodeprofAliases, nodeprofDefaultFields)
}

// MT: Constant after initialization; immutable
var nodeprofAliases = map[string][]string{
	"Default": []string{"Timestamp", "Hostname", "UsedMemory", "Load1", "Load5", "RunnableEntities"},
	"All":     []string{"Timestamp", "Hostname", "UsedMemory", "Load1", "Load5", "Load15", "RunnableEntities", "ExistingEntities"},
}

const nodeprofDefaultFields = "Default"
