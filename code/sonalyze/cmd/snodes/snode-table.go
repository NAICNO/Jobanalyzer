// DO NOT EDIT.  Generated from snodes.go by generate-table

package snodes

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
var snodeFormatters = map[string]Formatter[SnodeData]{
	"Timestamp": {
		Fmt: func(d SnodeData, ctx PrintMods) string {
			return FormatString((d.Timestamp), ctx)
		},
		Help: "(string) Full ISO timestamp of when the reading was taken",
	},
	"Nodes": {
		Fmt: func(d SnodeData, ctx PrintMods) string {
			return FormatStrings((d.Nodes), ctx)
		},
		Help: "(string list) Node list",
	},
	"States": {
		Fmt: func(d SnodeData, ctx PrintMods) string {
			return FormatStrings((d.States), ctx)
		},
		Help: "(string list) State list",
	},
}

func init() {
	DefAlias(snodeFormatters, "Timestamp", "timestamp")
	DefAlias(snodeFormatters, "Nodes", "nodes")
	DefAlias(snodeFormatters, "States", "states")
}

// MT: Constant after initialization; immutable
var snodePredicates = map[string]Predicate[SnodeData]{
	"Timestamp": Predicate[SnodeData]{
		Compare: func(d SnodeData, v any) int {
			return cmp.Compare((d.Timestamp), v.(string))
		},
	},
	"Nodes": Predicate[SnodeData]{
		Convert: CvtString2Strings,
		SetCompare: func(d SnodeData, v any, op int) bool {
			return SetCompareStrings((d.Nodes), v.([]string), op)
		},
	},
	"States": Predicate[SnodeData]{
		Convert: CvtString2Strings,
		SetCompare: func(d SnodeData, v any, op int) bool {
			return SetCompareStrings((d.States), v.([]string), op)
		},
	},
}

type SnodeData struct {
	Timestamp string
	Nodes     []string
	States    []string
}

func (c *SnodeCommand) Summary(out io.Writer) {
	fmt.Fprint(out, `  Print Slurm partition data
`)
}

const snodeHelp = `
snode
  TODO
`

func (c *SnodeCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(c.Fmt, snodeHelp, snodeFormatters, snodeAliases, snodeDefaultFields)
}

// MT: Constant after initialization; immutable
var snodeAliases = map[string][]string{
	"default": []string{"host", "nodes", "states"},
	"Default": []string{"Hostname", "Nodes", "States"},
	"all":     []string{"timestamp", "host", "nodes", "states"},
	"All":     []string{"Timestamp", "Hostname", "Nodes", "States"},
}

const snodeDefaultFields = "default"
