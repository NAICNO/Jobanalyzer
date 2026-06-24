package api2

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"sonalyze/daemon/apiutil"
	"sonalyze/data/sample"
)

const nodesCpuTimeseriesName = "/cluster/{cluster}/nodes/cpu/timeseries"

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
			Path:        nodesCpuTimeseriesName,
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
		Nodenames     string `query:"nodename" doc:"Compressed node name list"`
	},
) (*NodesCpuTimeseriesResponse, error) {
	prof, hErr := computeProfile(
		nodesCpuTimeseriesName,
		input.Cluster,
		input.StartTimeInS,
		input.EndTimeInS,
		input.ResolutionInS,
		input.Nodenames,
	)
	if hErr != nil {
		return nil, hErr
	}
	resp := &NodesCpuTimeseriesResponse{
		Body: make(map[string][]NodesCpuTimeseries_Point),
	}
	for name, pdata := range prof {
		var profile []NodesCpuTimeseries_Point
		for _, it := range pdata {
			if it.cpu_util > 0 {
				profile = append(profile, NodesCpuTimeseries_Point{
					Time:    formatTime(it.time),
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
	opName, clusterName string,
	startTimeInS, endTimeInS, resolutionInS uint64,
	nodenames string,
) (map[string][]profStepData, huma.StatusError) {
	meta, hErr := apiutil.GetClusterContext(opName, clusterName)
	if hErr != nil {
		return nil, hErr
	}
	from, to, hErr := apiutil.TimeWindowFromData(opName, meta, startTimeInS, endTimeInS)
	if hErr != nil {
		return nil, hErr
	}
	bucket := int64(300)
	if resolutionInS != 0 {
		bucket = int64(resolutionInS)
	}

	hostFilter, hErr := apiutil.NewHostFilter(opName, meta, nodenames, from, to)
	if hErr != nil {
		return nil, hErr
	}
	sysinfo, hErr := getSysinfoAt(opName, meta, to, hostFilter)
	if hErr != nil {
		return nil, hErr
	}
	sdp, hErr := openSampleDataProvider(opName, meta)
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
			opName+": Failed to query sample data", err)
	}

	reports := make(map[string][]profStepData)
	for _, stream := range sample.MergeByHost(streams) {
		samples := stream.Samples
		if len(samples) == 0 {
			continue
		}
		nodeName := samples[0].Hostname.String()
		node := sysinfo[nodeName]
		if node == nil {
			// No config data, can happen sometimes, no way to compute relative values.
			continue
		}
		cores := node.Sockets * node.CoresPerSocket * node.ThreadsPerCore
		var stepData []profStepData

		t := canonicalizeInitialTimestep(samples[0].Timestamp, bucket)

		// Index in sample array
		i := 0
		for i < len(samples) {

			// Average quantities across the time bucket but empty buckets just get 0.
			var cpuAcc float64
			var memAcc uint64
			var n int
			for i < len(samples) && samples[i].Timestamp < t+bucket {
				// Note CpuUtilPct is percent of one core and hence can be > 100
				cpuAcc += float64(samples[i].CpuUtilPct)
				memAcc += samples[i].RssAnonKB
				n++
				i++
			}
			if n > 1 {
				cpuAcc /= float64(n)
				memAcc /= uint64(n)
			}
			// Divide by NumTasks to account for CpuUtilPct issue above
			cpuAcc /= float64(stream.NumTasks)

			// Record a step always even if we should happen not to have consumed any samples in the
			// time step (empty bucket); printing logic can opt to not print zero values.
			var sd profStepData
			sd.time = t
			if cores > 0 {
				sd.cpu_util = float64(cpuAcc) / float64(cores) * 100.0
			}
			sd.cpu_util = onePlace(sd.cpu_util)
			if node.Memory > 0 {
				sd.memory_util = float64(memAcc) / float64(node.Memory) * 100.0
			}
			sd.memory_util = onePlace(sd.memory_util)
			stepData = append(stepData, sd)

			// Step time
			t += bucket
		}
		reports[nodeName] = stepData
	}

	return reports, nil
}
