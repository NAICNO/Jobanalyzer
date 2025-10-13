package clusters

import (
	"cmp"
	"errors"
	"io"
	"slices"

	. "sonalyze/cmd"
	"sonalyze/db/special"
	. "sonalyze/table"
)

//go:generate ../../../generate-table/generate-table -o cluster-table.go clusters.go

/*TABLE cluster

package clusters

import "sonalyze/db/special"

%%

FIELDS *special.ClusterEntry

 Name        string   desc:"Cluster name" alias:"cluster"
 Description string   desc:"Human-consumable cluster summary" alias:"desc"
 Aliases     []string desc:"Aliases of cluster" alias:"aliases"

SUMMARY ClusterCommand

Display information about the clusters and overall cluster configuration.

As this operates on the store and not on cluster data in the store, there is
no -cluster argument for remote runs.

For per-node data, use "config" and/or "node".

HELP ClusterCommand

  Extract information about individual clusters in the data store.
  Output records are sorted by cluster name.  The default format is 'fixed'.

ALIASES

  all      cluster,desc,aliases
  All      Name,Description,Aliases
  default  cluster,aliases,desc
  Default  Name,Aliases,Description

DEFAULTS default

ELBAT*/

type ClusterCommand struct {
	DevArgs
	DatabaseArgs
	QueryArgs
	VerboseArgs
	FormatArgs
}

var _ = PrimitiveCommand((*ClusterCommand)(nil))

func (cc *ClusterCommand) Add(fs *CLI) {
	cc.DevArgs.Add(fs)
	cc.DatabaseArgs.Add(fs, DBArgOptions{OmitCluster: true})
	cc.QueryArgs.Add(fs)
	cc.VerboseArgs.Add(fs)
	cc.FormatArgs.Add(fs)
}

func (cc *ClusterCommand) ReifyForRemote(x *ArgReifier) error {
	return errors.Join(
		cc.DevArgs.ReifyForRemote(x),
		cc.DatabaseArgs.ReifyForRemote(x),
		cc.QueryArgs.ReifyForRemote(x),
		cc.FormatArgs.ReifyForRemote(x),
	)
}

func (cc *ClusterCommand) Validate() error {
	return errors.Join(
		cc.DevArgs.Validate(),
		cc.VerboseArgs.Validate(),
		cc.DatabaseArgs.Validate(),
		cc.QueryArgs.Validate(),
		ValidateFormatArgs(
			&cc.FormatArgs,
			clusterDefaultFields,
			clusterFormatters,
			clusterAliases,
			DefaultFixed,
		),
	)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Analysis

func (cc *ClusterCommand) Perform(_ io.Reader, stdout, stderr io.Writer) error {
	printable := special.AllClusters()
	slices.SortFunc(printable, func(a, b *special.ClusterEntry) int {
		return cmp.Compare(a.Name, b.Name)
	})

	var err error
	printable, err = ApplyQuery(
		cc.ParsedQuery, clusterFormatters, clusterPredicates, printable)
	if err != nil {
		return err
	}

	FormatData(
		stdout,
		cc.PrintFields,
		clusterFormatters,
		cc.PrintOpts,
		printable,
	)

	return nil
}
