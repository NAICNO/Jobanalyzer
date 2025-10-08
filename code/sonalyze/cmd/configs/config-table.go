// DO NOT EDIT.  Generated from configs.go by generate-table

package configs

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
var configFormatters = map[string]Formatter[*repr.NodeSummary]{
	"Timestamp": {
		Fmt: func(d *repr.NodeSummary, ctx PrintMods) string {
			return FormatString((d.Timestamp), ctx)
		},
		Help: "(string) Full ISO timestamp of when the reading was taken",
	},
	"Hostname": {
		Fmt: func(d *repr.NodeSummary, ctx PrintMods) string {
			return FormatString((d.Hostname), ctx)
		},
		Help: "(string) Name that host is known by on the cluster",
	},
	"Description": {
		Fmt: func(d *repr.NodeSummary, ctx PrintMods) string {
			return FormatString((d.Description), ctx)
		},
		Help: "(string) End-user description, not parseable",
	},
	"CpuCores": {
		Fmt: func(d *repr.NodeSummary, ctx PrintMods) string {
			return FormatInt((d.CpuCores), ctx)
		},
		Help: "(int) Total number of cores x threads",
	},
	"MemGB": {
		Fmt: func(d *repr.NodeSummary, ctx PrintMods) string {
			return FormatInt((d.MemGB), ctx)
		},
		Help: "(int) GB of installed main RAM",
	},
	"GpuCards": {
		Fmt: func(d *repr.NodeSummary, ctx PrintMods) string {
			return FormatInt((d.GpuCards), ctx)
		},
		Help: "(int) Number of installed cards",
	},
	"GpuMemGB": {
		Fmt: func(d *repr.NodeSummary, ctx PrintMods) string {
			return FormatInt((d.GpuMemGB), ctx)
		},
		Help: "(int) Total GPU memory across all cards",
	},
	"GpuMemPct": {
		Fmt: func(d *repr.NodeSummary, ctx PrintMods) string {
			return FormatBool((d.GpuMemPct), ctx)
		},
		Help: "(bool) True if GPUs report accurate memory usage in percent",
	},
}

func init() {
	DefAlias(configFormatters, "Timestamp", "timestamp")
	DefAlias(configFormatters, "Hostname", "host")
	DefAlias(configFormatters, "Description", "desc")
	DefAlias(configFormatters, "CpuCores", "cores")
	DefAlias(configFormatters, "MemGB", "mem")
	DefAlias(configFormatters, "GpuCards", "gpus")
	DefAlias(configFormatters, "GpuMemGB", "gpumem")
	DefAlias(configFormatters, "GpuMemPct", "gpumempct")
}

// MT: Constant after initialization; immutable
var configPredicates = map[string]Predicate[*repr.NodeSummary]{
	"Timestamp": Predicate[*repr.NodeSummary]{
		Compare: func(d *repr.NodeSummary, v any) int {
			return cmp.Compare((d.Timestamp), v.(string))
		},
	},
	"Hostname": Predicate[*repr.NodeSummary]{
		Compare: func(d *repr.NodeSummary, v any) int {
			return cmp.Compare((d.Hostname), v.(string))
		},
	},
	"Description": Predicate[*repr.NodeSummary]{
		Compare: func(d *repr.NodeSummary, v any) int {
			return cmp.Compare((d.Description), v.(string))
		},
	},
	"CpuCores": Predicate[*repr.NodeSummary]{
		Convert: CvtString2Int,
		Compare: func(d *repr.NodeSummary, v any) int {
			return cmp.Compare((d.CpuCores), v.(int))
		},
	},
	"MemGB": Predicate[*repr.NodeSummary]{
		Convert: CvtString2Int,
		Compare: func(d *repr.NodeSummary, v any) int {
			return cmp.Compare((d.MemGB), v.(int))
		},
	},
	"GpuCards": Predicate[*repr.NodeSummary]{
		Convert: CvtString2Int,
		Compare: func(d *repr.NodeSummary, v any) int {
			return cmp.Compare((d.GpuCards), v.(int))
		},
	},
	"GpuMemGB": Predicate[*repr.NodeSummary]{
		Convert: CvtString2Int,
		Compare: func(d *repr.NodeSummary, v any) int {
			return cmp.Compare((d.GpuMemGB), v.(int))
		},
	},
	"GpuMemPct": Predicate[*repr.NodeSummary]{
		Convert: CvtString2Bool,
		Compare: func(d *repr.NodeSummary, v any) int {
			return CompareBool((d.GpuMemPct), v.(bool))
		},
	},
}

func (c *ConfigCommand) Summary(out io.Writer) {
	fmt.Fprint(out, `Display information about nodes in a cluster configuration.

The node configuration is time-dependent and is computed from data reported
by the node and additional background information.

For overall cluster data, use "cluster".  Also see "node" for closely
related data.
`)
}

const configHelp = `
config
  Extract information about individual nodes on the cluster from config data and
  present them in primitive form.  Output records are sorted by node name.  The
  default format is 'fixed'.
`

func (c *ConfigCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(c.Fmt, configHelp, configFormatters, configAliases, configDefaultFields)
}

// MT: Constant after initialization; immutable
var configAliases = map[string][]string{
	"default": []string{"host", "cores", "mem", "gpus", "gpumem", "desc"},
	"Default": []string{"Hostname", "CpuCores", "MemGB", "GpuCards", "GpuMemGB", "Description"},
	"all":     []string{"timestamp", "host", "desc", "cores", "mem", "gpus", "gpumem", "gpumempct"},
	"All":     []string{"Timestamp", "Hostname", "Description", "CpuCores", "MemGB", "GpuCards", "GpuMemGB", "GpuMemPct"},
}

const configDefaultFields = "default"
