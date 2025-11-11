// Print sysinfo for individual nodes.  The sysinfo is a simple join of node and gpu data, without
// any surrounding context.

package nodes

import (
	"cmp"
	"errors"
	"io"
	"slices"

	. "sonalyze/cmd"
	"sonalyze/data/common"
	"sonalyze/data/config"
	"sonalyze/db/special"
	. "sonalyze/table"
)

//go:generate ../../../generate-table/generate-table -o node-table.go nodes.go

/*TABLE node

package nodes

import "sonalyze/data/config"

%%

FIELDS *config.NodeConfig

 Timestamp   string desc:"Full ISO timestamp of when the reading was taken" alias:"timestamp"
 Hostname    string desc:"Name that host is known by on the cluster" alias:"host"
 Description string desc:"End-user description, not parseable" alias:"desc"
 CpuCores    int    desc:"Total number of cores x threads" alias:"cores"
 MemGB       int    desc:"GB of installed main RAM" alias:"mem"
 GpuCards    int    desc:"Number of installed cards" alias:"gpus"
 GpuMemGB    int    desc:"Total GPU memory across all cards" alias:"gpumem"
 GpuMemPct   bool   desc:"True if GPUs report accurate memory usage in percent" alias:"gpumempct"
 Distances   string desc:"NUMA distance matrix" alias:"distances"
 TopoSVG     string desc:"SVG encoding of node topology" alias:"toposvg"
 TopoText    string desc:"Text encoding of node topology" alias:"topotext"

SUMMARY NodeCommand

Display self-reported information about nodes in a cluster.

For overall cluster data, use "cluster".  Also see "config" for
closely related data.

The node configuration is time-dependent and is reported by the node
periodically, it will usually only change if the node is upgraded or
components are inserted/removed.

HELP NodeCommand

  Extract information about individual nodes on the cluster from sysinfo and present
  them in primitive form.  Output records are sorted by node name.  The default
  format is 'fixed'.

  Note that 'all' and 'All' do not include the topology data (toposvg, topotext), which
  are usually large if present.  The most practical extraction method would be with e.g.
  "-from ... -host ... -newest -fmt noheader,fixed,topotext" for whatever single node
  is desired.

ALIASES

  default  host,cores,mem,gpus,gpumem,desc
  Default  Hostname,CpuCores,MemGB,GpuCards,GpuMemGB,Description
  all      timestamp,host,desc,cores,mem,gpus,gpumem,gpumempct,distances
  All      Timestamp,Hostname,Description,CpuCores,MemGB,GpuCards,GpuMemGB,GpuMemPct,Distances

DEFAULTS default

ELBAT*/

type NodeCommand struct {
	HostAnalysisArgs
	FormatArgs
	Newest bool
}

var _ = SimpleCommand((*NodeCommand)(nil))

func (nc *NodeCommand) Add(fs *CLI) {
	nc.HostAnalysisArgs.Add(fs)
	nc.FormatArgs.Add(fs)

	fs.Group("printing")
	fs.BoolVar(&nc.Newest, "newest", false, "Print newest record per host only")
}

func (nc *NodeCommand) ReifyForRemote(x *ArgReifier) error {
	// As per normal, do not forward VerboseArgs.
	x.Bool("newest", nc.Newest)
	return errors.Join(
		nc.HostAnalysisArgs.ReifyForRemote(x),
		nc.FormatArgs.ReifyForRemote(x),
	)
}

func (nc *NodeCommand) Validate() error {
	return errors.Join(
		nc.HostAnalysisArgs.Validate(),
		ValidateFormatArgs(
			&nc.FormatArgs, nodeDefaultFields, nodeFormatters, nodeAliases, DefaultFixed),
	)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Processing

func (nc *NodeCommand) Perform(meta special.ClusterMeta, _ io.Reader, stdout, stderr io.Writer) error {
	cdp, err := config.OpenConfigDataProvider(meta)
	if err != nil {
		return err
	}
	records, err := cdp.Query(config.QueryArgs{
		QueryFilter: common.QueryFilter{
			HaveFrom: nc.HaveFrom,
			FromDate: nc.FromDate,
			HaveTo:   nc.HaveTo,
			ToDate:   nc.ToDate,
			Host:     nc.Host,
		},
		Verbose: nc.Verbose,
		Newest:  nc.Newest,
		Query: func(records []*config.NodeConfig) ([]*config.NodeConfig, error) {
			return ApplyQuery(nc.ParsedQuery, nodeFormatters, nodePredicates, records)
		},
	})
	if err != nil {
		return err
	}

	// Sort by host name first and then by ascending time
	slices.SortFunc(records, func(a, b *config.NodeConfig) int {
		if h := cmp.Compare(a.Hostname, b.Hostname); h != 0 {
			return h
		}
		return cmp.Compare(a.Timestamp, b.Timestamp)
	})

	FormatData(
		stdout,
		nc.PrintFields,
		nodeFormatters,
		nc.PrintOpts,
		records,
	)

	return nil
}
