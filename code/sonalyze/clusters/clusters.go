// The output of this verb is a list of clusters with any available metadata.
//
// * For `-remote` there must not be a `-cluster` option
// * When running locally there must be `-jobanalyzer-dir` as for `daemon`
// * `-data-dir` and `-config-file` are illegal, as are all filtering options

package clusters

import (
	"cmp"
	"errors"
	"flag"
	"io"
	"slices"
	"strings"

	umaps "go-utils/maps"
	"go-utils/options"

	. "sonalyze/command"
	. "sonalyze/common"
	"sonalyze/db"
)

type ClusterCommand struct {
	DevArgs
	RemotingArgsNoCluster
	VerboseArgs
	FormatArgs
	jobanalyzerDir string
}

func (cc *ClusterCommand) Summary() []string {
	return []string{
		"Extract information about clusters in the data store",
	}
}

func (cc *ClusterCommand) Add(fs *flag.FlagSet) {
	cc.DevArgs.Add(fs)
	cc.RemotingArgsNoCluster.Add(fs)
	cc.VerboseArgs.Add(fs)
	cc.FormatArgs.Add(fs)
	fs.StringVar(&cc.jobanalyzerDir, "jobanalyzer-dir", "", "Jobanalyzer root `directory`")
}

func (cc *ClusterCommand) ReifyForRemote(x *Reifier) error {
	return errors.Join(
		cc.DevArgs.ReifyForRemote(x),
		cc.FormatArgs.ReifyForRemote(x),
	)
}

func (cc *ClusterCommand) Validate() error {
	var e1, e2 error

	if cc.jobanalyzerDir == "" {
		ApplyDefault(&cc.Remote, "data-source", "remote")
		ApplyDefault(&cc.AuthFile, "data-source", "auth-file")
	}

	e1 = errors.Join(
		cc.DevArgs.Validate(),
		cc.VerboseArgs.Validate(),
		cc.RemotingArgsNoCluster.Validate(),
		ValidateFormatArgs(
			&cc.FormatArgs,
			clustersDefaultFields,
			clustersFormatters,
			clustersAliases,
			DefaultFixed,
		),
	)

	if cc.RemotingFlags().Remoting {
		if cc.jobanalyzerDir != "" {
			e2 = errors.New("-jobanalyzer-dir not valid for remote execution")
		}
	} else {
		cc.jobanalyzerDir, e2 = options.RequireDirectory(cc.jobanalyzerDir, "-jobanalyzer-dir")
	}

	return errors.Join(e1, e2)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Analysis

func (cc *ClusterCommand) Clusters(_ io.Reader, stdout, stderr io.Writer) error {
	clusters, _, err := db.ReadClusterData(cc.jobanalyzerDir)
	if err != nil {
		return err
	}

	printable := umaps.Values(clusters)
	slices.SortFunc(printable, func(a, b *db.ClusterEntry) int {
		return cmp.Compare(a.Name, b.Name)
	})

	FormatData(
		stdout,
		cc.FormatArgs.PrintFields,
		clustersFormatters,
		cc.FormatArgs.PrintOpts,
		printable,
		false,
	)

	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Printing

func (cc *ClusterCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(
		cc.Fmt, clustersHelp, clustersFormatters, clustersAliases, clustersDefaultFields)
}

const clustersHelp = `
nodes
  Extract information about individual clusters in the data store.
  Output records are sorted by cluster name.  The default format is 'fixed'.
`

type clustersCtx = bool
type clustersSummary = db.ClusterEntry

const clustersDefaultFields = "cluster,aliases,desc"

// MT: Constant after initialization; immutable
var clustersAliases = map[string][]string{
	"default":     strings.Split(clustersDefaultFields, ","),
	"description": []string{"desc"},
}

// MT: Constant after initialization; immutable
var clustersFormatters = map[string]Formatter[*clustersSummary, clustersCtx]{
	"cluster": {
		func(i *clustersSummary, _ clustersCtx) string {
			return i.Name
		},
		"Cluster name",
	},
	"desc": {
		func(i *clustersSummary, _ clustersCtx) string {
			return i.Description
		},
		"Human-consumable cluster summary",
	},
	"aliases": {
		func(i *clustersSummary, _ clustersCtx) string {
			// Print aliases in deterministic order
			aliases := slices.Clone(i.Aliases)
			slices.Sort(aliases)
			return strings.Join(aliases, ",")
		},
		"Aliases of cluster",
	},
}
