package api2

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"sonalyze/data/sample"
)

const processesTimeseriesName = "/cluster/{cluster}/processes/timeseries"

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
			Path:        processesTimeseriesName,
			Summary:     "...",
		},
		handleProcessesTimeseries,
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
	meta, hErr := getClusterContext(processesTimeseriesName, input.Cluster)
	if hErr != nil {
		return nil, hErr
	}
	from, to, hErr := timeWindowFromData(processesTimeseriesName, meta, input.StartTimeInS, input.EndTimeInS)
	if hErr != nil {
		return nil, hErr
	}
	bucket := int64(300)
	if input.ResolutionInS != 0 {
		bucket = int64(input.ResolutionInS)
	}
	hostFilter, hErr := newHostFilter(processesTimeseriesName, input.Nodename)
	if hErr != nil {
		return nil, hErr
	}
	sdp, hErr := openSampleDataProvider(processesTimeseriesName, meta)
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
			verbose,
		)
	if err != nil {
		return nil, huma.Error500InternalServerError(
			processesTimeseriesName+": Failed to query sample data", err)
	}

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
		t := canonicalizeInitialTimestep(stream[0].Timestamp, bucket)
		i := 0
		for i < len(stream) {
			var acc ProcessesTimeseries_Point
			var n uint64
			acc.Time = formatTime(t)
			for i < len(stream) && stream[i].Timestamp < t+bucket {
				sample := stream[i]
				acc.CpuPct += float64(sample.CpuPct)
				acc.MemKB += sample.CpuKB
				acc.GpuPct += float64(sample.GpuPct)
				acc.GpuMemPct += float64(sample.GpuMemPct)
				n++
				i++
			}
			if n > 1 {
				acc.CpuPct /= float64(n)
				acc.MemKB /= n
				acc.GpuPct /= float64(n)
				acc.GpuMemPct /= float64(n)
			}
			acc.CpuPct = onePlace(acc.CpuPct)
			acc.GpuPct = onePlace(acc.GpuPct)
			acc.GpuMemPct = onePlace(acc.GpuMemPct)
			data = append(data, acc)
			t += bucket
		}
		proc.Data = data
		resp.Body[node] = append(resp.Body[node], proc)
	}
	return resp, nil
}
