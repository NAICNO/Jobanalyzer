// Operation nodes/process/gpu/util computes per-node GPU usage at a point in time.  It lumps
// together all pids on the node that use GPU at some time and sums their GPU usage.
//
// This is not a time series, but a point in such an implied sequence.  The parameter
// reference_time_in_s gives us the point we are interested in (default now), the parameter
// window_in_s the time span over which we average (default 300s).
//
// This is very similar to nodes/cpu/timeseries and nodes/memory/timeseries.
package api2

import (
	"context"
	"net/http"
	"slices"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"sonalyze/data/sample"
)

const nodesProcessGpuUtilName = "/cluster/{cluster}/nodes/process/gpu/util"

type NodesProcessGpuUtilResponse struct {
	// Map: node -> data
	Body map[string]NodesProcessGpuUtil_Point
}

type NodesProcessGpuUtil_Point struct {
	GpuMemory     uint64   `json:"gpu_memory" doc:"GPU Memory being utilized in KiB"`
	GpuMemoryUtil float64  `json:"gpu_memory_util" doc:"GPU Memory utilization in percentage"`
	GpuUtil       float64  `json:"gpu_util" doc:"GPU Compute utilization in percentage"`
	Pids          []uint64 `json:"pids" doc:"Process ids related to an accumulated sample"`
	Time          string   `json:"time" doc:"Timezone Aware timestamp"`
}

func addNodesProcessGpuUtil(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "get-nodes-process-gpu-util",
			Method:      http.MethodGet,
			Path:        nodesProcessGpuUtilName,
			Summary:     "Per-node GPU usage at a point in time, averaged over a time window",
		},
		handleNodesProcessGpuUtil,
	)
}

func handleNodesProcessGpuUtil(
	ctx context.Context,
	input *struct {
		Cluster          string `path:"cluster" example:"my.cluster.name" doc:"Name of cluster"`
		Nodename         string `query:"nodename" doc:"Compressed node name list"`
		ReferenceTimeInS uint64 `query:"reference_time_in_s" doc:"Center point of averaging interval"`
		WindowInS        uint64 `query:"window_in_s" doc:"Width of averaging interval, default 300"`
	},
) (*NodesProcessGpuUtilResponse, error) {
	meta, hErr := getClusterContext(nodesProcessGpuUtilName, input.Cluster)
	if hErr != nil {
		return nil, hErr
	}
	from, to, hErr := timeWindowFromData(
		nodesProcessGpuUtilName, meta, input.ReferenceTimeInS, input.ReferenceTimeInS)
	if hErr != nil {
		return nil, hErr
	}
	t := from
	w := time.Duration(int64(input.WindowInS/2)) * time.Second
	from = from.Add(-w)
	to = to.Add(w)

	hostFilter, hErr := newHostFilter(nodesProcessGpuUtilName, input.Nodename)
	if hErr != nil {
		return nil, hErr
	}
	sdp, hErr := openSampleDataProvider(nodesProcessGpuUtilName, meta)
	if hErr != nil {
		return nil, hErr
	}
	streams, _, _, _, err :=
		sdp.Query(
			from,
			to,
			hostFilter,
			&sample.SampleFilter{From: from.Unix(), To: to.Unix()},
			false, // bounds
		)
	if err != nil {
		return nil, huma.Error500InternalServerError(
			nodesProcessGpuUtilName+": Failed to query sample data", err)
	}

	// There one synthesized sample stream per host.  The samples will all have different
	// timestamps, and each stream will be sorted ascending by timestamp.

	resp := &NodesProcessGpuUtilResponse{
		Body: make(map[string]NodesProcessGpuUtil_Point),
	}
	for _, stream := range sample.MergeByHost(streams) {
		samples := stream.Samples
		if len(samples) == 0 {
			continue
		}
		nodeName := samples[0].Hostname.String()

		var acc NodesProcessGpuUtil_Point
		var n uint64
		for _, s := range samples {
			if s.GpuPct > 0 || s.GpuMemPct > 0 || s.GpuKB > 0 {
				acc.GpuUtil += float64(s.GpuPct)
				acc.GpuMemoryUtil += float64(s.GpuMemPct)
				acc.GpuMemory += s.GpuKB
				n++
			}
		}

		if n == 0 {
			continue
		}

		// Not null, the JSON distinguishes
		pids := make([]uint64, 0)
		for _, t := range stream.Tasks {
			if t[0].Pid != 0 {
				pids = append(pids, t[0].Pid)
			}
		}
		slices.Sort(pids)

		acc.Time = t.UTC().Format(time.RFC3339)
		if n > 1 {
			acc.GpuUtil /= float64(n)
			acc.GpuMemoryUtil /= float64(n)
			acc.GpuMemory /= n
		}
		acc.GpuUtil = onePlace(acc.GpuUtil)
		acc.GpuMemoryUtil = onePlace(acc.GpuMemoryUtil)
		acc.Pids = pids
		resp.Body[nodeName] = acc
	}

	return resp, nil
}
