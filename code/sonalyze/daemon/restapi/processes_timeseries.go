package restapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	. "sonalyze/common"
	"sonalyze/data/sample"
	"sonalyze/db"
	"sonalyze/db/special"
)

type ProcessesTimeseriesResponse struct {
	// Map: node -> data
	Body map[string][]ProcessesTimeseries_Process
}

type ProcessesTimeseries_Process struct {
	Pid  uint64                      `json:"pid" doc:"Process ID"`
	User string                      `json:"user" doc:"User name"`
	Cmd  string                      `json:"cmd" doc:"Command being run"`
	Data []ProcessesTimeseries_Point `json:"data"`
}

type ProcessesTimeseries_Point struct {
	Time      string  `json:"time"`
	CpuPct    float64 `json:"cpu_pct"`
	MemKB     uint64  `json:"mem_kb"`
	GpuPct    float64 `json:"gpu_pct" doc:"GPU compute utilization % (100 = 1 full card)"`
	GpuMemPct float64 `json:"gpu_mem_pct" doc:"GPU memory utilization % (100 = 1 full card)"`
}

func addProcessesTimeseries(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "get-processes-timeseries",
			Method:      http.MethodGet,
			Path:        "/cluster/{cluster}/processes/timeseries",
			Summary:     "...",
		},
		handleProcesses,
	)
}

// The meaning here is a little unclear but most likely we want the Pid to be unique (per node) and
// represent the sample value closest to the given time, as a separate API gets a time series.

func handleProcessesTimeseries(
	ctx context.Context,
	input *struct {
		Cluster       string `path:"cluster" example:"my.cluster.name" doc:"Name of cluster"`
		StartTimeInS  uint64 `query:"start_time_in_s" doc:"Posix timestamp"`
		EndTimeInS    uint64 `query:"end_time_in_s" doc:"Posix timestamp"`
		ResolutionInS uint64 `query:"resolution_in_s" doc:"Default is 300"`
		Nodename      string `query:"nodename" doc:"Compressed node name list"`
	},
) (*ProcessesTimeseriesResponse, error) {
	cluster := special.LookupCluster(input.Cluster)
	if cluster == nil {
		return nil, fmt.Errorf("Failed to find cluster %s", input.Cluster)
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
	sdp, err := sample.OpenSampleDataProvider(meta)
	if err != nil {
		return nil, fmt.Errorf("Failed to access sample database: %v", err)
	}
	hostFilter, err := NewHosts(true, hostList)
	if err != nil {
		return nil, err
	}
	streams, _, _, _, err :=
		sdp.Query(
			from,
			to,
			hostFilter,
			&sample.SampleFilter{From: from.Unix(), To: to.Unix()},
			false, // bounds
			verbose,
		)
	if err != nil {
		return nil, fmt.Errorf("Failed to read sample log records: %v", err)
	}

	// TODO: Implement bucketing
	_ = bucket

	resp := &ProcessesTimeseriesResponse{
		Body: make(map[string][]ProcessesTimeseries_Process),
	}
	for _, s := range streams {
		stream := *s
		var proc ProcessesTimeseries_Process
		node := stream[0].Hostname.String()
		proc.Pid = stream[0].Pid
		proc.User = stream[0].User.String()
		proc.Cmd = stream[0].Cmd.String()
		var data []ProcessesTimeseries_Point
		for _, sample := range stream {
			data = append(data, ProcessesTimeseries_Point{
				Time:      time.Unix(sample.Timestamp, 0).UTC().Format(time.RFC3339),
				CpuPct:    float64(sample.CpuPct),
				MemKB:     sample.CpuKB,
				GpuPct:    float64(sample.GpuPct),
				GpuMemPct: float64(sample.GpuMemPct),
			})
		}
		proc.Data = data
		resp.Body[node] = append(resp.Body[node], proc)
	}
	return resp, nil
}
