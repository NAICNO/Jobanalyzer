package clusters

import (
	"cmp"
	_ "embed"
	"errors"
	"io"
	"slices"

	umaps "go-utils/maps"
	"go-utils/options"

	. "sonalyze/cmd"
	. "sonalyze/common"
	"sonalyze/db"
	. "sonalyze/table"
)

//go:generate ../../../generate-table/generate-table -o cluster-table.go clusters.go

/*TABLE cluster

package clusters

import (
	"sonalyze/db"
	. "sonalyze/table"
)

%%

FIELDS *db.ClusterEntry

 Name        string   desc:"Cluster name" alias:"cluster"
 Description string   desc:"Human-consumable cluster summary" alias:"desc"
 Aliases     []string desc:"Aliases of cluster" alias:"aliases"

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
	RemotingArgsNoCluster
	VerboseArgs
	FormatArgs
	JobanalyzerDir string
}

//go:embed summary.txt
var summary string

func (cc *ClusterCommand) Summary() string {
	return summary
}

func (cc *ClusterCommand) Add(fs *CLI) {
	cc.DevArgs.Add(fs)
	cc.RemotingArgsNoCluster.Add(fs)
	cc.VerboseArgs.Add(fs)
	cc.FormatArgs.Add(fs)
	fs.Group("local-data-source")
	fs.StringVar(&cc.JobanalyzerDir, "jobanalyzer-dir", "", "Jobanalyzer root `directory`")
}

func (cc *ClusterCommand) ReifyForRemote(x *ArgReifier) error {
	return errors.Join(
		cc.DevArgs.ReifyForRemote(x),
		cc.FormatArgs.ReifyForRemote(x),
	)
}

func (cc *ClusterCommand) Validate() error {
	var e1, e2 error

	if cc.JobanalyzerDir == "" {
		ApplyDefault(&cc.Remote, "data-source", "remote")
		ApplyDefault(&cc.AuthFile, "data-source", "auth-file")
	}

	e1 = errors.Join(
		cc.DevArgs.Validate(),
		cc.VerboseArgs.Validate(),
		cc.RemotingArgsNoCluster.Validate(),
		ValidateFormatArgs(
			&cc.FormatArgs,
			clusterDefaultFields,
			clusterFormatters,
			clusterAliases,
			DefaultFixed,
		),
	)

	if cc.RemotingFlags().Remoting {
		if cc.JobanalyzerDir != "" {
			e2 = errors.New("-jobanalyzer-dir not valid for remote execution")
		}
	} else {
		cc.JobanalyzerDir, e2 = options.RequireDirectory(cc.JobanalyzerDir, "-jobanalyzer-dir")
	}

	return errors.Join(e1, e2)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Analysis

func (cc *ClusterCommand) Perform(_ io.Reader, stdout, stderr io.Writer) error {
	clusters, _, err := db.ReadClusterData(cc.JobanalyzerDir)
	if err != nil {
		return err
	}

	printable := umaps.Values(clusters)
	slices.SortFunc(printable, func(a, b *db.ClusterEntry) int {
		return cmp.Compare(a.Name, b.Name)
	})

	FormatData(
		stdout,
		cc.PrintFields,
		clusterFormatters,
		cc.PrintOpts,
		printable,
	)

	return nil
}
