package restapi

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"sonalyze/data/common"
	"sonalyze/data/config"
	"sonalyze/data/slurmpart"
	"sonalyze/db"
	"sonalyze/db/special"
)

type ClusterResponse struct {
	Body []*Cluster_Cluster
}

type Cluster_Cluster struct {
	Time       string   `json:"time" example:"2026-3-10T12:10:15Z" doc:"Time generated"`
	Cluster    string   `json:"cluster" example:"my.cluster.name" doc:"Canonical cluster name"`
	Slurm      int      `json:"slurm" example:"0" doc:"Slurm flag"`
	Partitions []string `json:"partitions" doc:"List of partition names"`
	Nodes      []string `json:"nodes" doc:"List of compressed node names"`
}

func addListClusters(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "get-cluster",
			Method:      http.MethodGet,
			Path:        "/cluster",
			Summary: `Retrieve all clusters.  For each cluster retrieve the most recent
information recorded no later than the time_in_s parameter - partition and node data are
time-dependent.`,
		},
		handleListClusters,
	)
}

func handleListClusters(
	ctx context.Context,
	input *struct {
		TimeInS uint64 `query:"time_in_s" doc:"Posix timestamp, default 'now'"`
	},
) (*ClusterResponse, error) {
	resp := &ClusterResponse{}
	t := time.Now().UTC().Format(time.RFC3339)
	for _, c := range special.AllClusters() {
		meta := db.NewContextFromCluster(c)
		from, to, err := timeWindowFromData(meta, input.TimeInS, input.TimeInS)
		if err != nil {
			return nil, err
		}
		filter := common.QueryFilter{HaveFrom: true, FromDate: from, HaveTo: true, ToDate: to}

		var partitions []string
		{
			// Logic from cmd/sparts
			spd, err := slurmpart.OpenSlurmPartitionDataProvider(meta)
			if err != nil {
				return nil, err
			}
			records, err := spd.Query(filter, verbose)
			if err != nil {
				return nil, fmt.Errorf("Failed to read log records: %v", err)
			}
			ps := make(map[string]bool)
			for _, r := range records {
				for _, p := range r.Partitions {
					ps[string(p.Name)] = true
				}
			}
			partitions = slices.Collect(maps.Keys(ps))
		}

		var nodes []string
		{
			// Logic from cmd/nodes
			cdp, err := config.OpenConfigDataProvider(meta)
			if err != nil {
				return nil, err
			}
			records, err := cdp.Query(config.QueryArgs{
				QueryFilter: config.QueryFilter{QueryFilter: filter, Verbose: verbose},
				Newest:      true,
			})
			if err != nil {
				return nil, err
			}
			for _, r := range records {
				nodes = append(nodes, r.Hostname)
			}
		}

		var slurm int
		if len(partitions) > 0 {
			slurm = 1
		}

		resp.Body = append(resp.Body, &Cluster_Cluster{
			Time:       t,
			Cluster:    c.Name,
			Slurm:      slurm,
			Partitions: partitions,
			Nodes:      nodes,
		})
	}
	return resp, nil
}
