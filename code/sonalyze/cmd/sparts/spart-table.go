// DO NOT EDIT.  Generated from sparts.go by generate-table

package sparts

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
var spartFormatters = map[string]Formatter[SpartData]{
	"Timestamp": {
		Fmt: func(d SpartData, ctx PrintMods) string {
			return FormatString((d.Timestamp), ctx)
		},
		Help: "(string) Full ISO timestamp of when the reading was taken",
	},
	"Partition": {
		Fmt: func(d SpartData, ctx PrintMods) string {
			return FormatString((d.Partition), ctx)
		},
		Help: "(string) Name of the partition",
	},
	"Nodes": {
		Fmt: func(d SpartData, ctx PrintMods) string {
			return FormatStrings((d.Nodes), ctx)
		},
		Help: "(string list) Node list",
	},
}

func init() {
	DefAlias(spartFormatters, "Timestamp", "timestamp")
	DefAlias(spartFormatters, "Partition", "part")
	DefAlias(spartFormatters, "Nodes", "nodes")
}

// MT: Constant after initialization; immutable
var spartPredicates = map[string]Predicate[SpartData]{
	"Timestamp": Predicate[SpartData]{
		Compare: func(d SpartData, v any) int {
			return cmp.Compare((d.Timestamp), v.(string))
		},
	},
	"Partition": Predicate[SpartData]{
		Compare: func(d SpartData, v any) int {
			return cmp.Compare((d.Partition), v.(string))
		},
	},
	"Nodes": Predicate[SpartData]{
		Convert: CvtString2Strings,
		SetCompare: func(d SpartData, v any, op int) bool {
			return SetCompareStrings((d.Nodes), v.([]string), op)
		},
	},
}

type SpartData struct {
	Timestamp string
	Partition string
	Nodes     []string
}

func (c *SpartCommand) Summary(out io.Writer) {
	fmt.Fprint(out, `  Print Slurm partition data
`)
}

const spartHelp = `
spart
  Slurm partitions are named and contain a set of nodes on the machine.  Not all nodes need be in a
  partition, some nodes may be in multiple partitions at the same time, and admins can move nodes
  among partitions.  Thus the partition information (also available from the "sinfo" command) is
  time-varying.  Output records are sorted by time and partition name, the default format is
  "fixed".
`

func (c *SpartCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(c.Fmt, spartHelp, spartFormatters, spartAliases, spartDefaultFields)
}

// MT: Constant after initialization; immutable
var spartAliases = map[string][]string{
	"default": []string{"part", "nodes"},
	"Default": []string{"Partition", "Nodes"},
	"all":     []string{"timestamp", "part", "nodes"},
	"All":     []string{"Timestamp", "Partition", "Nodes"},
}

const spartDefaultFields = "default"
