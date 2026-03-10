package restapi

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	//	. "sonalyze/common"
	// "sonalyze/data/sample"
	"sonalyze/db"
	"sonalyze/db/special"
)

type NodesGpuTimeseriesResponse struct {
	// Mapping: node -> data
	Body map[string][]NodesGpuTimeseries_Card
}

type NodesGpuTimeseries_Card struct {
	UUID  string                         `json:"uuid"`
	Index uint64                         `json:"index"`
	Data  []NodesGpuTimeseries_CardPoint `json:"data"`
}

type NodesGpuTimeseries_CardPoint struct {
	Time    string  `json:"time"`
	CEUtil  float64 `json:"ce_util,omitempty" doc:"Percent 0..100 of total node compute capacity"`
	MemUtil float64 `json:"mem_util,omitempty" doc:Percent 0..100 of total node memory capacity"`
}

func addNodesGpuTimeseries(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "get-nodes-gpu-timeseries",
			Method:      http.MethodGet,
			Path:        "/cluster/{cluster}/nodes/gpu/timeseries",
			Summary:     "Compute per-node per-card aggregated gpu timeseries for timespan",
		},
		handleNodesGpuTimeseries,
	)
}

func handleNodesGpuTimeseries(
	ctx context.Context,
	input *struct {
		Cluster       string `path:"cluster" example:"my.cluster.name" doc:"Name of cluster"`
		StartTimeInS  uint64 `query:"start_time_in_s" doc:"Posix timestamp"`
		EndTimeInS    uint64 `query:"end_time_in_s" doc:"Posix timestamp"`
		ResolutionInS uint64 `query:"resolution_in_s" doc:"Default is 300"`
		Nodename      string `query:"nodename" doc:"Compressed node name list"`
	},
) (*NodesGpuTimeseriesResponse, error) {
	clusterName := input.Cluster
	cluster := special.LookupCluster(clusterName)
	if cluster == nil {
		return nil, fmt.Errorf("Failed to find cluster %s", clusterName)
	}
	meta := db.NewContextFromCluster(cluster)
	from, to, err := timeWindowFromData(meta, input.StartTimeInS, input.EndTimeInS)
	if err != nil {
		return nil, err
	}
	bucket := uint64(300)
	if input.ResolutionInS != 0 {
		bucket = input.ResolutionInS
	}

	var hostList []string
	if input.Nodename != "" {
		hostList = []string{input.Nodename}
	}
	nodeMap, err := getNodeMap(meta, from, to, hostList)
	if err != nil {
		return nil, err
	}

	// This is pretty fucked up.  For each card (= uuid), it can be on various nodes at various
	// times, it can be moved from one node to another and then back later.  So we need to construct
	// a series of gpu samples per (node, uuid) pair that may have missing segments.  Then for each
	// node we present all the uuids that were ever on the node in the requested time span.

	// FIXME: Implement - this is row 6
	_ = bucket
	_ = nodeMap
	return nil, fmt.Errorf("nodes/gpu/timeseries not implemented")
}
