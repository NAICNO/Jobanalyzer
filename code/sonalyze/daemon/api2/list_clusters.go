package api2

import (
	"context"
	"maps"
	"net/http"
	"slices"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"sonalyze/data/common"
	"sonalyze/data/node"
	"sonalyze/data/slurmpart"
	"sonalyze/db"
	"sonalyze/db/special"
)

const listClustersName = "/cluster"

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
			Path:        listClustersName,
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
	for _, c := range special.AllClusters() {
		meta := db.NewContextFromCluster(c)
		_, to, hErr := timeWindowFromData(listClustersName, meta, 0, input.TimeInS)
		from := to.Add(-24 * time.Hour)
		if hErr != nil {
			return nil, hErr
		}
		t := to.UTC().Format(time.RFC3339)
		filter := common.QueryFilter{HaveFrom: true, FromDate: from, HaveTo: true, ToDate: to}

		var partitions []string
		{
			// Logic from cmd/sparts
			spd, err := slurmpart.OpenSlurmPartitionDataProvider(meta)
			if err != nil {
				return nil, huma.Error500InternalServerError(
					listClustersName+": Unable to open slurm data store", err)
			}
			records, err := spd.Query(filter, verbose)
			if err != nil {
				return nil, huma.Error500InternalServerError(
					listClustersName+": Failed to query slurm data", err)
			}
			ps := make(map[string]bool)
			for _, r := range records {
				for _, p := range r.Partitions {
					ps[string(p.Name)] = true
				}
			}
			if len(ps) == 0 {
				partitions = make([]string, 0)
			} else {
				partitions = slices.Collect(maps.Keys(ps))
				slices.Sort(partitions)
			}
		}

		var nodes []string
		{
			ndp, err := node.OpenNodeDataProvider(meta)
			if err != nil {
				return nil, huma.Error500InternalServerError(
					listClustersName+": Failed to open config data store", err)
			}
			records, err := ndp.Query(filter, verbose)
			if err != nil {
				return nil, huma.Error500InternalServerError(
					listClustersName+": Failed to query config data", err)
			}
			ns := make(map[string]bool)
			for _, r := range records {
				ns[r.Node] = true
			}
			if len(ns) == 0 {
				nodes = make([]string, 0)
			} else {
				nodes = slices.Collect(maps.Keys(ns))
				slices.Sort(nodes)
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
