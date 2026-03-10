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

type ProcessesResponse struct {
	// Map: node -> data
	Body map[string][]ProcessesResponse_Process
}

type ProcessesResponse_Process struct {
	Pid       uint64   `json:"pid" doc:"Process ID"`
	User      string   `json:"user" doc:"User name"`
	Cmd       string   `json:"cmd" doc:"Command being run"`
	CpuPct    float64  `json:"cpu_pct" doc:"CPU usage % (100 = 1 full core)"`
	MemKB     uint64   `json:"mem_kb" doc:"Resident memory in KB"`
	GpuPct    float64  `json:"gpu_pct" doc:"GPU compute utilization % (100 = 1 full card)"`
	GpuMemPct float64  `json:"gpu_mem_pct" doc:"GPU memory utilization % (100 = 1 full card)"`
	Gpus      []string `json:"gpus" doc:"GPU UUIDs used by process"`
	Time      string   `json:"time" doc:"Sample timestamp"`
}

func addProcesses(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "get-processes",
			Method:      http.MethodGet,
			Path:        "/cluster/{cluster}/processes",
			Summary:     "Get utilization of processes at a time",
		},
		handleProcesses,
	)
}

// The meaning here is a little unclear but most likely we want the Pid to be unique (per node) and
// represent the sample value closest to the given time, as a separate API gets a time series.

func handleProcesses(
	ctx context.Context,
	input *struct {
		Cluster  string `path:"cluster" example:"my.cluster.name" doc:"Name of cluster"`
		Nodename string `query:"nodename" doc:"Compressed node name list"`
		TimeInS  uint64 `query:"time_in_s" doc:"Posix timestamp"`
	},
) (*ProcessesResponse, error) {
	// FIXME: Implement - this is row 9
	return nil, fmt.Errorf("/processes not implemented")
}
