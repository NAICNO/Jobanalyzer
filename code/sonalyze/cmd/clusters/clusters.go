package clusters

import (
	"cmp"
	_ "embed"
	"errors"
	"io"
	"reflect"
	"slices"
	"strings"

	umaps "go-utils/maps"
	"go-utils/options"
	uslices "go-utils/slices"

	. "sonalyze/cmd"
	. "sonalyze/common"
	"sonalyze/db"
	. "sonalyze/table"
)

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
			clustersDefaultFields,
			clustersFormatters,
			clustersAliases,
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

	sortable := umaps.Values(clusters)
	slices.SortFunc(sortable, func(a, b *db.ClusterEntry) int {
		return cmp.Compare(a.Name, b.Name)
	})
	printable := uslices.Map(sortable, func(x *db.ClusterEntry) any { return x })

	FormatData(
		stdout,
		cc.PrintFields,
		clustersFormatters,
		cc.PrintOpts,
		printable,
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
cluster
  Extract information about individual clusters in the data store.
  Output records are sorted by cluster name.  The default format is 'fixed'.
`

const v0ClustersDefaultFields = "cluster,aliases,desc"
const v1ClustersDefaultFields = "Name,Aliases,Description"
const clustersDefaultFields = v0ClustersDefaultFields

// MT: Constant after initialization; immutable
var clustersAliases = map[string][]string{
	"default":   strings.Split(clustersDefaultFields, ","),
	"v0default": strings.Split(v0ClustersDefaultFields, ","),
	"v1default": strings.Split(v1ClustersDefaultFields, ","),
}

type SFS = SimpleFormatSpec

// MT: Constant after initialization; immutable
var clustersFormatters = DefineTableFromMap(
	reflect.TypeFor[db.ClusterEntry](),
	map[string]any{
		"Name":        SFS{"Cluster name", "cluster"},
		"Description": SFS{"Human-consumable cluster summary", "desc"},
		"Aliases":     SFS{"Aliases of cluster", "aliases"},
	},
)
