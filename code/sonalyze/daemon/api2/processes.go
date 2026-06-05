package api2

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"sonalyze/daemon/apiutil"
	"sonalyze/data/sample"
)

const processesName = "/cluster/{cluster}/processes"

type ProcessesResponse struct {
	// Map: node -> data
	Body map[string][]Processes_Process
}

type Processes_Process struct {
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
			Path:        processesName,
			Summary:     "Get utilization of processes at a time",
		},
		handleProcesses,
	)
}

// The meaning here is a little unclear but most likely we want the Pid to be unique (per node) and
// represent the sample value closest to the given time, as a separate API gets a time series.  Most
// of the time the time will be "now"?  The doc does not list any query possibilities.  That being
// so, we treat "TimeInS" as a "to" timestamp and do a 1-hour query against sample data on the
// node(s), and then get a set of streams.  For each stream, we take the latest sample, and extract
// the raw data from that.

func handleProcesses(
	ctx context.Context,
	input *struct {
		Cluster  string `path:"cluster" example:"my.cluster.name" doc:"Name of cluster"`
		Nodename string `query:"nodename" doc:"Compressed node name list"`
		TimeInS  uint64 `query:"time_in_s" doc:"Posix timestamp"`
	},
) (*ProcessesResponse, error) {
	meta, hErr := apiutil.GetClusterContext(processesName, input.Cluster)
	if hErr != nil {
		return nil, hErr
	}
	from, to, hErr := apiutil.TimeWindowFromData(processesName, meta, 0, input.TimeInS)
	if hErr != nil {
		return nil, hErr
	}
	hostFilter, hErr := apiutil.NewHostFilter(processesName, meta, input.Nodename, from, to)
	if hErr != nil {
		return nil, hErr
	}
	sdp, hErr := openSampleDataProvider(processesName, meta)
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
			processesName+": Failed to query sample data", err)
	}
	cardsByNode, hErr := getCardInfoByNodeAt(nodesInfoName, meta, to, hostFilter)
	if hErr != nil {
		return nil, hErr
	}
	resp := &ProcessesResponse{
		Body: make(map[string][]Processes_Process),
	}
	for _, s := range streams {
		stream := *s
		item := stream[len(stream)-1]
		node := item.Hostname.String()
		var proc Processes_Process
		proc.Time = formatTime(item.Timestamp)
		proc.Pid = item.Pid
		proc.User = item.User.String()
		proc.Cmd = item.Cmd.String()
		proc.CpuPct = onePlace(float64(item.CpuSampledUtilPct))
		proc.MemKB = item.CpuKB
		proc.GpuPct = onePlace(float64(item.GpuPct))
		proc.GpuMemPct = onePlace(float64(item.GpuMemPct))
		proc.Gpus = make([]string, 0)
		for _, c := range gpuSetToGpus(item.Gpus, cardsByNode[node]) {
			proc.Gpus = append(proc.Gpus, c.UUID)
		}
		resp.Body[node] = append(resp.Body[node], proc)
	}

	return resp, nil
}
