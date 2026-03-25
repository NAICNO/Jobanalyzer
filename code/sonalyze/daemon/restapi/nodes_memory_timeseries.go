package restapi

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type NodesMemoryTimeseriesResponse struct {
	// Map: node -> data
	Body map[string][]NodesMemoryTimeseries_Point
}

type NodesMemoryTimeseries_Point struct {
	Time       string  `json:"time"`
	MemoryUtil float64 `json:"memory_util,omitempty" doc:"Percent 0..100 of total node memory capacity"`
}

func addNodesMemoryTimeseries(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "get-nodes-memory-timeseries",
			Method:      http.MethodGet,
			Path:        "/cluster/{cluster}/nodes/memory/timeseries",
			Summary:     "Compute per-node aggregated memory timeseries for timespan",
		},
		handleNodesMemoryTimeseries,
	)
}

func handleNodesMemoryTimeseries(
	ctx context.Context,
	input *struct {
		Cluster       string `path:"cluster" example:"my.cluster.name" doc:"Name of cluster"`
		StartTimeInS  uint64 `query:"start_time_in_s" doc:"Posix timestamp"`
		EndTimeInS    uint64 `query:"end_time_in_s" doc:"Posix timestamp"`
		ResolutionInS uint64 `query:"resolution_in_s" doc:"Default is 300"`
		Nodename      string `query:"nodename" doc:"Compressed node name list"`
	},
) (*NodesMemoryTimeseriesResponse, error) {
	prof, hErr := computeProfile(
		"/cluster/C/nodes/memory/timeseries",
		input.Cluster,
		input.StartTimeInS,
		input.EndTimeInS,
		input.ResolutionInS,
		input.Nodename,
	)
	if hErr != nil {
		return nil, hErr
	}
	resp := &NodesMemoryTimeseriesResponse{
		Body: make(map[string][]NodesMemoryTimeseries_Point),
	}
	for name, pdata := range prof {
		var profile []NodesMemoryTimeseries_Point
		for _, it := range pdata {
			if it.memory_util > 0 {
				profile = append(profile, NodesMemoryTimeseries_Point{
					Time:       formatTime(it.time),
					MemoryUtil: onePlace(it.memory_util),
				})
			}
		}
		resp.Body[name] = profile
	}

	return resp, nil
}
