// Print sysinfo for individual nodes.  The sysinfo is a simple join of node and gpu data, without
// any surrounding context.

package nodes

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"math"
	"slices"
	"time"

	umaps "go-utils/maps"

	. "sonalyze/cmd"
	"sonalyze/data/card"
	"sonalyze/data/node"
	"sonalyze/db"
	"sonalyze/db/repr"
	"sonalyze/db/special"
	. "sonalyze/table"
)

//go:generate ../../../generate-table/generate-table -o node-table.go nodes.go

/*TABLE node

package nodes

%%

FIELDS *NodeData

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

GENERATE NodeData

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
	theLog, err := db.OpenReadOnlyDB(
		meta,
		db.FileListNodeData|db.FileListCardData,
	)
	if err != nil {
		return err
	}

	records, err := Query(theLog, NodeQueryArgs{
		HaveFrom: nc.HaveFrom,
		FromDate: nc.FromDate,
		HaveTo:   nc.HaveTo,
		ToDate:   nc.ToDate,
		Host:     nc.HostArgs.Host,
		Verbose:  nc.Verbose,
		Newest:   nc.Newest,
		Query: func(records []*NodeData) ([]*NodeData, error) {
			return ApplyQuery(nc.ParsedQuery, nodeFormatters, nodePredicates, records)
		},
	})
	if err != nil {
		return err
	}

	// Sort by host name first and then by ascending time
	slices.SortFunc(records, func(a, b *NodeData) int {
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

type NodeQueryArgs struct {
	HaveFrom bool
	FromDate time.Time
	HaveTo   bool
	ToDate   time.Time
	Host     []string
	Verbose  bool
	Newest   bool
	Query    func(records []*NodeData) ([]*NodeData, error)
}

type joinedData struct {
	// the time and host are given by node
	Node  *repr.SysinfoNodeData
	Cards []*repr.SysinfoCardData
}

// Read and filter the raw sysinfo data.
//
// TODO: This is where we really want to lean on code in data/ and not use the repr directly, as
// we're computing a join and we'd like the database to do that for us if it is able to.
//
// What we want here is, for each point in time, join the cards on a host to the node data for
// the host, so that we can present at least some card information with the node.  Node names
// and time stamps are restricted strings so it's easy enough to use those as a key.  The logic
// that is here really belongs in data/sysinfo (or similar), but that does not exist yet.
//
// Join logic can be generalized by parameterizing by types and the key constructor, probably.
// Probably this would turn into a situation where we compute a list of data from both input
// sets.

func Query(theLog db.DataProvider, qa NodeQueryArgs) ([]*NodeData, error) {
	nodes, err := node.Query(
		theLog,
		node.QueryFilter{
			HaveFrom: qa.HaveFrom,
			FromDate: qa.FromDate,
			HaveTo:   qa.HaveTo,
			ToDate:   qa.ToDate,
			Host:     qa.Host,
		},
		qa.Verbose,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to read log records: %v", err)
	}
	cards, err := card.Query(
		theLog,
		card.QueryFilter{
			HaveFrom: qa.HaveFrom,
			FromDate: qa.FromDate,
			HaveTo:   qa.HaveTo,
			ToDate:   qa.ToDate,
			Host:     qa.Host,
		},
		qa.Verbose,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to read log records: %v", err)
	}
	joined := make(map[string]*joinedData)
	for _, r := range nodes {
		joined[r.Time+"|"+r.Node] = &joinedData{Node: r}
	}
	for _, r := range cards {
		if probe := joined[r.Time+"|"+r.Node]; probe != nil {
			probe.Cards = append(probe.Cards, r)
		}
	}
	rawRecords := umaps.Values(joined)
	records := make([]*NodeData, len(rawRecords))
	for i, r := range rawRecords {
		ht := ""
		if r.Node.ThreadsPerCore > 1 {
			ht = " (hyperthreaded)"
		}
		memGB := int(math.Round(float64(r.Node.Memory) / (1024 * 1024)))
		desc := fmt.Sprintf(
			"%dx%d%s %s, %d GiB", r.Node.Sockets, r.Node.CoresPerSocket, ht, r.Node.CpuModel, memGB)
		cores := r.Node.Sockets * r.Node.CoresPerSocket * r.Node.ThreadsPerCore
		numCards := len(r.Cards)
		cardTotalMemKB := uint64(0)
		for _, c := range r.Cards {
			cardTotalMemKB += c.Memory
		}
		cardTotalMemGB := int(math.Round(float64(cardTotalMemKB) / (1024 * 1024)))
		if numCards > 0 {
			desc += fmt.Sprintf(", %dx %s @ %dGiB", numCards, r.Cards[0].Model, (r.Cards[0].Memory)/(1024*1024))
		}
		distances := ""
		if r.Node.Distances != nil {
			distances = fmt.Sprintf("%v", r.Node.Distances)
		}
		records[i] = &NodeData{
			Timestamp:   r.Node.Time,
			Hostname:    r.Node.Node,
			Description: desc,
			CpuCores:    int(cores),
			MemGB:       memGB,
			GpuCards:    numCards,
			GpuMemGB:    cardTotalMemGB,
			Distances:   distances,
			TopoSVG:     r.Node.TopoSVG,
			TopoText:    r.Node.TopoText,
		}
	}
	if qa.Query != nil {
		records, err = qa.Query(records)
		if err != nil {
			return nil, err
		}
	}
	if qa.Newest {
		newr := make(map[string]*NodeData)
		for _, r := range records {
			if probe := newr[r.Hostname]; probe != nil {
				if r.Timestamp > probe.Timestamp {
					newr[r.Hostname] = r
				}
			} else {
				newr[r.Hostname] = r
			}
		}
		records = umaps.Values(newr)
	}
	return records, nil
}
