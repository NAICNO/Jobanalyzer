// DO NOT EDIT.  Generated from clusters.go by generate-table

package clusters

import (
	"sonalyze/db"
	. "sonalyze/table"
)

// MT: Constant after initialization; immutable
var clusterFormatters = map[string]Formatter[*db.ClusterEntry]{
	"Name": {
		Fmt: func(d *db.ClusterEntry, ctx PrintMods) string {
			return FormatString(string(d.Name), ctx)
		},
		Help: "(string) Cluster name",
	},
	"Description": {
		Fmt: func(d *db.ClusterEntry, ctx PrintMods) string {
			return FormatString(string(d.Description), ctx)
		},
		Help: "(string) Human-consumable cluster summary",
	},
	"Aliases": {
		Fmt: func(d *db.ClusterEntry, ctx PrintMods) string {
			return FormatStrings([]string(d.Aliases), ctx)
		},
		Help: "(string list) Aliases of cluster",
	},
}

func init() {
	DefAlias(clusterFormatters, "Name", "cluster")
	DefAlias(clusterFormatters, "Description", "desc")
	DefAlias(clusterFormatters, "Aliases", "aliases")
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
	"default": []string{"all"},
	"Default": []string{"All"},
}

const clusterDefaultFields = "default"
