// Query static configuration data, ie, the join of node and card data with the option of some
// filtering.
package config

import (
	"fmt"
	"math"
	"time"

	umaps "go-utils/maps"

	"sonalyze/data/card"
	"sonalyze/data/node"
	"sonalyze/db"
	"sonalyze/db/repr"
)

type NodeConfig struct {
	repr.NodeSummary
	Distances string
	TopoSVG   string
	TopoText  string
}

// From time to time it is appealing to add flags here to not generate various parts of the
// structure, for the sake of performance.  Don't do that, it messes up caching.
type QueryArgs struct {
	HaveFrom bool
	FromDate time.Time
	HaveTo   bool
	ToDate   time.Time
	Host     []string
	Verbose  bool
	Newest   bool
	Query    func(records []*NodeConfig) ([]*NodeConfig, error)
}

func Query(theLog db.DataProvider, qa QueryArgs) ([]*NodeConfig, error) {
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
	type joinedData struct {
		// the time and host are given by node
		node  *repr.SysinfoNodeData
		cards []*repr.SysinfoCardData
	}
	joined := make(map[string]*joinedData)
	for _, r := range nodes {
		joined[r.Time+"|"+r.Node] = &joinedData{node: r}
	}
	for _, r := range cards {
		if probe := joined[r.Time+"|"+r.Node]; probe != nil {
			probe.cards = append(probe.cards, r)
		}
	}
	rawRecords := umaps.Values(joined)
	records := make([]*NodeConfig, len(rawRecords))
	for i, r := range rawRecords {
		ht := ""
		if r.node.ThreadsPerCore > 1 {
			ht = " (hyperthreaded)"
		}
		memGB := int(math.Round(float64(r.node.Memory) / (1024 * 1024)))
		desc := fmt.Sprintf(
			"%dx%d%s %s, %d GiB", r.node.Sockets, r.node.CoresPerSocket, ht, r.node.CpuModel, memGB)
		cores := r.node.Sockets * r.node.CoresPerSocket * r.node.ThreadsPerCore
		numCards := len(r.cards)
		cardTotalMemKB := uint64(0)
		for _, c := range r.cards {
			cardTotalMemKB += c.Memory
		}
		cardTotalMemGB := int(math.Round(float64(cardTotalMemKB) / (1024 * 1024)))
		if numCards > 0 {
			desc += fmt.Sprintf(", %dx %s @ %dGiB", numCards, r.cards[0].Model, (r.cards[0].Memory)/(1024*1024))
		}
		distances := ""
		if r.node.Distances != nil {
			distances = fmt.Sprintf("%v", r.node.Distances)
		}
		records[i] = &NodeConfig{
			NodeSummary: repr.NodeSummary{
				Timestamp:   r.node.Time,
				Hostname:    r.node.Node,
				Description: desc,
				CpuCores:    int(cores),
				MemGB:       memGB,
				GpuCards:    numCards,
				GpuMemGB:    cardTotalMemGB,
				// CrossNodeJobs is not being set here because it is ill-defined and will
				// be removed from sonalyze.
				//
				// `Metadata` is unused by sonalyze.
			},
			Distances: distances,
			TopoSVG:   r.node.TopoSVG,
			TopoText:  r.node.TopoText,
		}
	}
	if qa.Query != nil {
		records, err = qa.Query(records)
		if err != nil {
			return nil, err
		}
	}
	if qa.Newest {
		newr := make(map[string]*NodeConfig)
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
