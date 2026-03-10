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

type ProcessesGpuResponse struct {
	// Map: node -> data
	Body map[string][]ProcessesGpu_Process
}

type ProcessesGpu_Process struct {
	Pid  uint64              `json:"pid" doc:"Process ID"`
	User string              `json:"user" doc:"User name"`
	Cmd  string              `json:"cmd" doc:"Command being run"`
	Gpus []ProcessesGpu_Card `json:"gpus" doc:"GPUs used by process"`
}

type ProcessesGpu_Card struct {
	Time      string  `json:"time" doc:"Sample timestamp"`
	GpuUUID   string  `json:"gpu_uuid"`
	GpuIndex  int     `json:"gpu_index"`
	GpuPct    float64 `json:"gpu_pct" doc:"GPU compute utilization % (100 = 1 full card)"`
	GpuMemPct float64 `json:"gpu_mem_pct" doc:"GPU memory utilization % (100 = 1 full card)"`
	GpuModel  string  `json:"gpu_model"`
}

func addProcessesGpu(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "get-processes-gpu",
			Method:      http.MethodGet,
			Path:        "/cluster/{cluster}/processes/gpu",
			Summary:     "Get utilization of processes at a time",
		},
		handleProcesses,
	)
}

// The meaning here is a little unclear but most likely we want the Pid to be unique (per node) and
// represent the sample value closest to the given time, as a separate API gets a time series.

func handleProcessesGpu(
	ctx context.Context,
	input *struct {
		Cluster  string `path:"cluster" example:"my.cluster.name" doc:"Name of cluster"`
		Nodename string `query:"nodename" doc:"Compressed node name list"`
		TimeInS  uint64 `query:"time_in_s" doc:"Posix timestamp"`
	},
) (*ProcessesGpuResponse, error) {
	// FIXME: Implement - this is row 11
	return nil, fmt.Errorf("/processes/gpu not implemented")
}
