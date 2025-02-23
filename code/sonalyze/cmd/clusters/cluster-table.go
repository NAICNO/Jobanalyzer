// DO NOT EDIT.  Generated from clusters.go by generate-table

package clusters

import "sonalyze/db"

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
var clusterFormatters = map[string]Formatter[*db.ClusterEntry]{
	"Name": {
		Fmt: func(d *db.ClusterEntry, ctx PrintMods) string {
			return FormatString((d.Name), ctx)
		},
		Help: "(string) Cluster name",
	},
	"Description": {
		Fmt: func(d *db.ClusterEntry, ctx PrintMods) string {
			return FormatString((d.Description), ctx)
		},
		Help: "(string) Human-consumable cluster summary",
	},
	"Aliases": {
		Fmt: func(d *db.ClusterEntry, ctx PrintMods) string {
			return FormatStrings((d.Aliases), ctx)
		},
		Help: "(string list) Aliases of cluster",
	},
}

func init() {
	DefAlias(clusterFormatters, "Name", "cluster")
	DefAlias(clusterFormatters, "Description", "desc")
	DefAlias(clusterFormatters, "Aliases", "aliases")
}

// MT: Constant after initialization; immutable
var clusterPredicates = map[string]Predicate[*db.ClusterEntry]{
	"Name": Predicate[*db.ClusterEntry]{
		Compare: func(d *db.ClusterEntry, v any) int {
			return cmp.Compare((d.Name), v.(string))
		},
	},
	"Description": Predicate[*db.ClusterEntry]{
		Compare: func(d *db.ClusterEntry, v any) int {
			return cmp.Compare((d.Description), v.(string))
		},
	},
	"Aliases": Predicate[*db.ClusterEntry]{
		Convert: CvtString2Strings,
		SetCompare: func(d *db.ClusterEntry, v any, op int) bool {
			return SetCompareStrings((d.Aliases), v.([]string), op)
		},
	},
}

func (c *ClusterCommand) Summary(out io.Writer) {
	fmt.Fprint(out, `Display information about the clusters and overall cluster configuration.

As this operates on the store and not on cluster data in the store, there is
no -cluster argument for remote runs.

For per-node data, use "config" and/or "node".
`)
}

const clusterHelp = `
cluster
  Extract information about individual clusters in the data store.
  Output records are sorted by cluster name.  The default format is 'fixed'.
`

func (c *ClusterCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(c.Fmt, clusterHelp, clusterFormatters, clusterAliases, clusterDefaultFields)
}

// MT: Constant after initialization; immutable
var clusterAliases = map[string][]string{
	"all":     []string{"cluster", "desc", "aliases"},
	"All":     []string{"Name", "Description", "Aliases"},
	"default": []string{"cluster", "aliases", "desc"},
	"Default": []string{"Name", "Aliases", "Description"},
}

const clusterDefaultFields = "default"
