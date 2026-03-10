package restapi

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	. "sonalyze/common"
	"sonalyze/data/sample"
	"sonalyze/db"
	"sonalyze/db/special"
)

type NodesCpuTimeseriesResponse struct {
	// Map: node -> data
	Body map[string][]NodesCpuTimeseries_Point
}

type NodesCpuTimeseries_Point struct {
	Time    string  `json:"time"`
	CpuUtil float64 `json:"cpu_util,omitempty" doc:"Percent 0..100 of total node capacity"`
}

func addNodesCpuTimeseries(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "get-nodes-cpu-timeseries",
			Method:      http.MethodGet,
			Path:        "/cluster/{cluster}/nodes/cpu/timeseries",
			Summary:     `Compute per-node aggregated cpu timeseries for a timespan.`,
		},
		handleNodesCpuTimeseries,
	)
}

func handleNodesCpuTimeseries(
	ctx context.Context,
	input *struct {
		Cluster       string `path:"cluster" example:"my.cluster.name" doc:"Name of cluster"`
		StartTimeInS  uint64 `query:"start_time_in_s" doc:"Posix timestamp"`
		EndTimeInS    uint64 `query:"end_time_in_s" doc:"Posix timestamp"`
		ResolutionInS uint64 `query:"resolution_in_s" doc:"Default is 300"`
		Nodename      string `query:"nodename" doc:"Compressed node name list"`
	},
) (*NodesCpuTimeseriesResponse, error) {
	prof, err := computeProfile(input.Cluster, input.StartTimeInS, input.EndTimeInS, input.ResolutionInS, input.Nodename)
	if err != nil {
		return nil, err
	}
	resp := &NodesCpuTimeseriesResponse{
		Body: make(map[string][]NodesCpuTimeseries_Point),
	}
	for name, pdata := range prof {
		var profile []NodesCpuTimeseries_Point
		for _, it := range pdata {
			if it.cpu_util > 0 {
				profile = append(profile, NodesCpuTimeseries_Point{
					Time:    time.Unix(it.time, 0).UTC().Format(time.RFC3339),
					CpuUtil: it.cpu_util,
				})
			}
		}
		resp.Body[name] = profile
	}

	return resp, nil
}

// computeProfile() is shared between handleNodesCpuTimeseries and handleNodesMemoryTimeseries

type profStepData struct {
	time        int64
	cpu_util    float64 // percent utilized of sockets * cores * threads
	memory_util float64 // percent resident of physical memory
}

func computeProfile(
	clusterName string,
	startTimeInS, endTimeInS, resolutionInS uint64,
	nodename string,
) (map[string][]profStepData, error) {
	cluster := special.LookupCluster(clusterName)
	if cluster == nil {
		return nil, fmt.Errorf("Failed to find cluster %s", clusterName)
	}
	meta := db.NewContextFromCluster(cluster)
	from, to, err := timeWindowFromData(meta, startTimeInS, endTimeInS)
	if err != nil {
		return nil, err
	}
	bucket := uint64(300)
	if resolutionInS != 0 {
		bucket = resolutionInS
	}

	var hostList []string
	if nodename != "" {
		hostList = []string{nodename}
	}
	nodeMap, err := getNodeMap(meta, from, to, hostList)
	if err != nil {
		return nil, err
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

	reports := make(map[string][]profStepData)
	for _, stream := range sample.MergeByHost(streams) {
		samples := stream.Samples
		if len(samples) == 0 {
			continue
		}
		nodeName := samples[0].Hostname.String()
		node := nodeMap[nodeName]
		if node == nil {
			// No config data, can happen sometimes, no way to compute relative values.
			continue
		}
		cores := node.Sockets * node.CoresPerSocket * node.ThreadsPerCore
		var stepData []profStepData

		// Current time step.  There's an argument to be made that the first time value in the
		// series should be rounded down to something so that time steps are predictable and
		// comparable across hosts, and not dependent on when data came in.  Note any bucket value
		// is acceptable.  The most sane algorithm might be to take the value and round down to
		// minute (if < 60), 5-minute (if < 600), 15-minute, 30-minute, 60-minute, 2-hour, 4-hour,
		// 8-hour, 24-hour, etc.
		t := samples[0].Timestamp

		// Index in sample array
		i := 0
		for i < len(samples) {

			// Average quantities across the time bucket but empty buckets just get 0.
			var cpuAcc float64
			var memAcc uint64
			var n int
			for i < len(samples) && samples[i].Timestamp < t+int64(bucket) {
				// Note CpuUtilPct is percent of one core and hence can be > 100
				cpuAcc += float64(samples[i].CpuUtilPct)
				memAcc += samples[i].RssAnonKB
				n++
				i++
			}
			if n > 0 {
				// Divide by NumTasks to account for CpuUtilPct issue above
				cpuAcc /= float64(n) * float64(stream.NumTasks)
				memAcc /= uint64(n)
			}

			// Record a step always even if we should happen not to have consumed any samples in the
			// time step (empty bucket); printing logic can opt to not print zero values.
			var sd profStepData
			sd.time = t
			if cores > 0 {
				sd.cpu_util = math.Round(float64(cpuAcc)/float64(cores)*100.0*10.0) / 10.0
			}
			if node.Memory > 0 {
				sd.memory_util = math.Round(float64(memAcc)/float64(node.Memory)*100.0*10.0) / 10.0
			}
			stepData = append(stepData, sd)

			// Step time
			t += int64(bucket)
		}
		reports[nodeName] = stepData
	}

	return reports, nil
}
