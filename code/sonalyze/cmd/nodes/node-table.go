// DO NOT EDIT.  Generated from nodes.go by generate-table

package nodes

import (
	"go-utils/config"
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
var nodeFormatters = map[string]Formatter[*config.NodeConfigRecord]{
	"Timestamp": {
		Fmt: func(d *config.NodeConfigRecord, ctx PrintMods) string {
			return FormatString(string(d.Timestamp), ctx)
		},
		Help: "(string) Full ISO timestamp of when the reading was taken",
	},
	"Hostname": {
		Fmt: func(d *config.NodeConfigRecord, ctx PrintMods) string {
			return FormatString(string(d.Hostname), ctx)
		},
		Help: "(string) Name that host is known by on the cluster",
	},
	"Description": {
		Fmt: func(d *config.NodeConfigRecord, ctx PrintMods) string {
			return FormatString(string(d.Description), ctx)
		},
		Help: "(string) End-user description, not parseable",
	},
	"CpuCores": {
		Fmt: func(d *config.NodeConfigRecord, ctx PrintMods) string {
			return FormatInt(int(d.CpuCores), ctx)
		},
		Help: "(int) Total number of cores x threads",
	},
	"MemGB": {
		Fmt: func(d *config.NodeConfigRecord, ctx PrintMods) string {
			return FormatInt(int(d.MemGB), ctx)
		},
		Help: "(int) GB of installed main RAM",
	},
	"GpuCards": {
		Fmt: func(d *config.NodeConfigRecord, ctx PrintMods) string {
			return FormatInt(int(d.GpuCards), ctx)
		},
		Help: "(int) Number of installed cards",
	},
	"GpuMemGB": {
		Fmt: func(d *config.NodeConfigRecord, ctx PrintMods) string {
			return FormatInt(int(d.GpuMemGB), ctx)
		},
		Help: "(int) Total GPU memory across all cards",
	},
	"GpuMemPct": {
		Fmt: func(d *config.NodeConfigRecord, ctx PrintMods) string {
			return FormatBool(bool(d.GpuMemPct), ctx)
		},
		Help: "(bool) True if GPUs report accurate memory usage in percent",
	},
}

func init() {
	DefAlias(nodeFormatters, "Timestamp", "timestamp")
	DefAlias(nodeFormatters, "Hostname", "host")
	DefAlias(nodeFormatters, "Description", "desc")
	DefAlias(nodeFormatters, "CpuCores", "cores")
	DefAlias(nodeFormatters, "MemGB", "mem")
	DefAlias(nodeFormatters, "GpuCards", "gpus")
	DefAlias(nodeFormatters, "GpuMemGB", "gpumem")
	DefAlias(nodeFormatters, "GpuMemPct", "gpumempct")
}

func (c *NodeCommand) Summary(out io.Writer) {
	fmt.Fprint(out, `Display self-reported information about nodes in a cluster.

For overall cluster data, use "cluster".  Also see "config" for
closely related data.

The node configuration is time-dependent and is reported by the node
periodically, it will usually only change if the node is upgraded or
components are inserted/removed.
`)
}

const nodeHelp = `
node
  Extract information about individual nodes on the cluster from sysinfo and present
  them in primitive form.  Output records are sorted by node name.  The default
  format is 'fixed'.
`

func (c *NodeCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(c.Fmt, nodeHelp, nodeFormatters, nodeAliases, nodeDefaultFields)
}

// MT: Constant after initialization; immutable
var nodeAliases = map[string][]string{
	"default": []string{"host", "cores", "mem", "gpus", "gpumem", "desc"},
	"Default": []string{"Hostname", "CpuCores", "MemGB", "GpuCards", "GpuMemGB", "Description"},
	"all":     []string{"timestamp", "host", "desc", "cores", "mem", "gpus", "gpumem", "gpumempct"},
	"All":     []string{"Timestamp", "Hostname", "Description", "CpuCores", "MemGB", "GpuCards", "GpuMemGB", "GpuMemPct"},
}

const nodeDefaultFields = "default"
