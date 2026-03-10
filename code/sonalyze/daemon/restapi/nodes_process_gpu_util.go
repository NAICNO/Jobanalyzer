// Operation nodes/process/gpu/util computes per-node GPU usage at a point in time.  It lumps
// together all pids on the node that use GPU at some time and sums their GPU usage.
//
// This is not a time series, but the most recent point in such an implied sequence.  The parameter
// reference_time_in_s gives us the point we are interested in, the parameter window_in_s presumably
// the how far we can stray from that to look for data *or* what time span to average over, but this
// is not well defined at the moment.
//
// The default point in time is probably the time of the most recent datum, and the window might be
// the usual 300s.
package restapi

import (
	"context"
	"fmt"
	_ "math"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	//	. "sonalyze/common"
	// "sonalyze/data/sample"
	_ "sonalyze/db"
	_ "sonalyze/db/special"
)

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
			Path:        "/cluster/{cluster}/nodes/process/gpu/util",
			Summary:     "Per-node GPU usage at a point in time",
		},
		handleNodesProcessGpuUtil,
	)
}

func handleNodesProcessGpuUtil(
	ctx context.Context,
	input *struct {
		Cluster          string `path:"cluster" example:"my.cluster.name" doc:"Name of cluster"`
		Nodename         string `query:"nodename" doc:"Compressed node name list"`
		ReferenceTimeInS uint64 `query:"reference_time_in_s"`
		WindowInS        uint64 `query:"window_in_s"`
	},
) (*NodesProcessGpuUtilResponse, error) {
	// FIXME: Implement - this is row 8
	return nil, fmt.Errorf("nodes/process/gpu/util not implemented")
}
