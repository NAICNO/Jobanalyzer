// DO NOT EDIT.  Generated from nodes.go by generate-table

package nodes

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
var nodeFormatters = map[string]Formatter[*NodeData]{
	"Timestamp": {
		Fmt: func(d *NodeData, ctx PrintMods) string {
			return FormatString((d.Timestamp), ctx)
		},
		Help: "(string) Full ISO timestamp of when the reading was taken",
	},
	"Hostname": {
		Fmt: func(d *NodeData, ctx PrintMods) string {
			return FormatString((d.Hostname), ctx)
		},
		Help: "(string) Name that host is known by on the cluster",
	},
	"Description": {
		Fmt: func(d *NodeData, ctx PrintMods) string {
			return FormatString((d.Description), ctx)
		},
		Help: "(string) End-user description, not parseable",
	},
	"CpuCores": {
		Fmt: func(d *NodeData, ctx PrintMods) string {
			return FormatInt((d.CpuCores), ctx)
		},
		Help: "(int) Total number of cores x threads",
	},
	"MemGB": {
		Fmt: func(d *NodeData, ctx PrintMods) string {
			return FormatInt((d.MemGB), ctx)
		},
		Help: "(int) GB of installed main RAM",
	},
	"GpuCards": {
		Fmt: func(d *NodeData, ctx PrintMods) string {
			return FormatInt((d.GpuCards), ctx)
		},
		Help: "(int) Number of installed cards",
	},
	"GpuMemGB": {
		Fmt: func(d *NodeData, ctx PrintMods) string {
			return FormatInt((d.GpuMemGB), ctx)
		},
		Help: "(int) Total GPU memory across all cards",
	},
	"GpuMemPct": {
		Fmt: func(d *NodeData, ctx PrintMods) string {
			return FormatBool((d.GpuMemPct), ctx)
		},
		Help: "(bool) True if GPUs report accurate memory usage in percent",
	},
	"Distances": {
		Fmt: func(d *NodeData, ctx PrintMods) string {
			return FormatString((d.Distances), ctx)
		},
		Help: "(string) NUMA distance matrix",
	},
	"TopoSVG": {
		Fmt: func(d *NodeData, ctx PrintMods) string {
			return FormatString((d.TopoSVG), ctx)
		},
		Help: "(string) SVG encoding of node topology",
	},
	"TopoText": {
		Fmt: func(d *NodeData, ctx PrintMods) string {
			return FormatString((d.TopoText), ctx)
		},
		Help: "(string) Text encoding of node topology",
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
	DefAlias(nodeFormatters, "Distances", "distances")
	DefAlias(nodeFormatters, "TopoSVG", "toposvg")
	DefAlias(nodeFormatters, "TopoText", "topotext")
}

// MT: Constant after initialization; immutable
var nodePredicates = map[string]Predicate[*NodeData]{
	"Timestamp": Predicate[*NodeData]{
		Compare: func(d *NodeData, v any) int {
			return cmp.Compare((d.Timestamp), v.(string))
		},
	},
	"Hostname": Predicate[*NodeData]{
		Compare: func(d *NodeData, v any) int {
			return cmp.Compare((d.Hostname), v.(string))
		},
	},
	"Description": Predicate[*NodeData]{
		Compare: func(d *NodeData, v any) int {
			return cmp.Compare((d.Description), v.(string))
		},
	},
	"CpuCores": Predicate[*NodeData]{
		Convert: CvtString2Int,
		Compare: func(d *NodeData, v any) int {
			return cmp.Compare((d.CpuCores), v.(int))
		},
	},
	"MemGB": Predicate[*NodeData]{
		Convert: CvtString2Int,
		Compare: func(d *NodeData, v any) int {
			return cmp.Compare((d.MemGB), v.(int))
		},
	},
	"GpuCards": Predicate[*NodeData]{
		Convert: CvtString2Int,
		Compare: func(d *NodeData, v any) int {
			return cmp.Compare((d.GpuCards), v.(int))
		},
	},
	"GpuMemGB": Predicate[*NodeData]{
		Convert: CvtString2Int,
		Compare: func(d *NodeData, v any) int {
			return cmp.Compare((d.GpuMemGB), v.(int))
		},
	},
	"GpuMemPct": Predicate[*NodeData]{
		Convert: CvtString2Bool,
		Compare: func(d *NodeData, v any) int {
			return CompareBool((d.GpuMemPct), v.(bool))
		},
	},
	"Distances": Predicate[*NodeData]{
		Compare: func(d *NodeData, v any) int {
			return cmp.Compare((d.Distances), v.(string))
		},
	},
	"TopoSVG": Predicate[*NodeData]{
		Compare: func(d *NodeData, v any) int {
			return cmp.Compare((d.TopoSVG), v.(string))
		},
	},
	"TopoText": Predicate[*NodeData]{
		Compare: func(d *NodeData, v any) int {
			return cmp.Compare((d.TopoText), v.(string))
		},
	},
}

type NodeData struct {
	Timestamp   string
	Hostname    string
	Description string
	CpuCores    int
	MemGB       int
	GpuCards    int
	GpuMemGB    int
	GpuMemPct   bool
	Distances   string
	TopoSVG     string
	TopoText    string
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

  Note that 'all' and 'All' do not include the topology data (toposvg, topotext), which
  are usually large if present.  The most practical extraction method would be with e.g.
  "-from ... -host ... -newest -fmt noheader,csv,topotext" for whatever single node is desired.
`

func (c *NodeCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(c.Fmt, nodeHelp, nodeFormatters, nodeAliases, nodeDefaultFields)
}

// MT: Constant after initialization; immutable
var nodeAliases = map[string][]string{
	"default": []string{"host", "cores", "mem", "gpus", "gpumem", "desc"},
	"Default": []string{"Hostname", "CpuCores", "MemGB", "GpuCards", "GpuMemGB", "Description"},
	"all":     []string{"timestamp", "host", "desc", "cores", "mem", "gpus", "gpumem", "gpumempct", "distances"},
	"All":     []string{"Timestamp", "Hostname", "Description", "CpuCores", "MemGB", "GpuCards", "GpuMemGB", "GpuMemPct", "Distances"},
}

const nodeDefaultFields = "default"
