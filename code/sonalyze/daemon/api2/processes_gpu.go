package api2

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	. "sonalyze/common"
	"sonalyze/data/gpusample"
	"sonalyze/data/sample"
)

const processesGpuName = "/cluster/{cluster}/processes/gpu"

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
			Path:        processesGpuName,
			Summary:     "Get utilization of GPUs for a process at a time",
		},
		handleProcessesGpu,
	)
}

// The meaning here is a little unclear but most likely we want the Pid to be unique (per node) and
// represent the sample value closest to the given time, as a separate API gets a time series.
//
// This is very close to /processes, it just computes more GPU data and slightly less process data.

func handleProcessesGpu(
	ctx context.Context,
	input *struct {
		Cluster  string `path:"cluster" example:"my.cluster.name" doc:"Name of cluster"`
		Nodename string `query:"nodename" doc:"Compressed node name list"`
		TimeInS  uint64 `query:"time_in_s" doc:"Posix timestamp"`
	},
) (*ProcessesGpuResponse, error) {
	meta, hErr := getClusterContext(processesGpuName, input.Cluster)
	if hErr != nil {
		return nil, hErr
	}
	from, to, hErr := timeWindowFromData(processesGpuName, meta, 0, input.TimeInS)
	if hErr != nil {
		return nil, hErr
	}
	hostFilter, hErr := newHostFilter(processesGpuName, input.Nodename)
	if hErr != nil {
		return nil, hErr
	}
	sdp, hErr := openSampleDataProvider(processesGpuName, meta)
	if hErr != nil {
		return nil, hErr
	}
	sampleStreams, _, _, _, err :=
		sdp.Query(
			from,
			to,
			hostFilter,
			&sample.SampleFilter{From: from.Unix(), To: to.Unix()},
			false, // bounds
			verbose,
		)
	if err != nil {
		return nil, huma.Error500InternalServerError(
			processesGpuName+": Failed to query sample data", err)
	}
	cardsByNode, hErr := getCardInfoByNodeAt(nodesInfoName, meta, to, hostFilter.Patterns())
	if hErr != nil {
		return nil, hErr
	}
	gsd, err := gpusample.OpenGpuSampleDataProvider(meta)
	if err != nil {
		return nil, huma.Error500InternalServerError(
			processesGpuName+": failed to open gpu sample store", err)
	}
	gpuStreams, _, _, _, err := gsd.Query(from, to, hostFilter, verbose)
	if err != nil {
		return nil, huma.Error500InternalServerError(
			processesGpuName+": Failed to query gpu sample data", err)
	}

	resp := &ProcessesGpuResponse{
		Body: make(map[string][]ProcessesGpu_Process),
	}
	for _, s := range sampleStreams {
		samples := *s
		item := samples[len(samples)-1]
		node := item.Hostname.String()
		var proc ProcessesGpu_Process
		proc.Pid = item.Pid
		proc.User = item.User.String()
		proc.Cmd = item.Cmd.String()
		proc.Gpus = make([]ProcessesGpu_Card, 0)
		gpus := gpuSetToGpus(item.Gpus, cardsByNode[node])
		if gpuStreamsForNode := gpuStreams[StringToUstr(node)]; gpuStreamsForNode != nil {
			for _, gpu := range gpus {
				for _, stream := range gpuStreamsForNode.Data {
					last := stream.Decoded[len(stream.Decoded)-1]
					if string(last.UUID) == gpu.UUID {
						var g ProcessesGpu_Card
						g.Time = formatTime(stream.Time)
						g.GpuUUID = string(last.UUID)
						g.GpuIndex = int(last.Index)
						g.GpuPct = float64(last.CEUtil)
						g.GpuMemPct = float64(last.MemoryUtil)
						g.GpuModel = gpu.Model
						proc.Gpus = append(proc.Gpus, g)
						break
					}
				}
			}
		}
		resp.Body[node] = append(resp.Body[node], proc)
	}
	return resp, nil
}
