package restapi

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"sonalyze/data/gpusample"
	"sonalyze/db/repr"
)

const nodesGpuTimeseriesName = "/cluster/{cluster}/nodes/gpu/timeseries"

type NodesGpuTimeseriesResponse struct {
	// Mapping: node -> data
	Body map[string][]NodesGpuTimeseries_Card
}

type NodesGpuTimeseries_Card struct {
	UUID  string                         `json:"uuid" doc:"Card identity"`
	Index uint64                         `json:"index" doc:"Card index for this timeseries"`
	Data  []NodesGpuTimeseries_CardPoint `json:"data"`
}

type NodesGpuTimeseries_CardPoint struct {
	Time    string  `json:"time"`
	CEUtil  float64 `json:"ce_util,omitempty" doc:"Percent 0..100 of total node compute capacity"`
	MemUtil float64 `json:"mem_util,omitempty" doc:"Percent 0..100 of total node memory capacity"`
}

func addNodesGpuTimeseries(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "get-nodes-gpu-timeseries",
			Method:      http.MethodGet,
			Path:        nodesGpuTimeseriesName,
			Summary:     "Compute per-node per-card aggregated gpu timeseries for timespan",
		},
		handleNodesGpuTimeseries,
	)
}

func handleNodesGpuTimeseries(
	ctx context.Context,
	input *struct {
		Cluster       string `path:"cluster" example:"my.cluster.name" doc:"Name of cluster"`
		StartTimeInS  uint64 `query:"start_time_in_s" doc:"Posix timestamp"`
		EndTimeInS    uint64 `query:"end_time_in_s" doc:"Posix timestamp"`
		ResolutionInS uint64 `query:"resolution_in_s" doc:"Default is 300"`
		Nodename      string `query:"nodename" doc:"Compressed node name list"`
	},
) (*NodesGpuTimeseriesResponse, error) {
	meta, hErr := getClusterContext(nodesGpuTimeseriesName, input.Cluster)
	if hErr != nil {
		return nil, hErr
	}
	from, to, hErr := timeWindowFromData(
		nodesGpuTimeseriesName, meta, input.StartTimeInS, input.EndTimeInS)
	if hErr != nil {
		return nil, hErr
	}
	bucket := int64(300)
	if input.ResolutionInS != 0 {
		bucket = int64(input.ResolutionInS)
	}

	// This is a profile broken down first by node, then by card-on-node.  Each physical card device
	// (= uuid) can be on various nodes at various times, and can have different indices on a node
	// at different times.  In the response data, Index is a single value.  Hence what we're
	// constructing is time series for (uuid, index) per node.  Given node-centric data, we need to
	// construct a series of gpu samples per (node, uuid, index) triplet that may have missing
	// segments.  Then for each node we present all the (uuids, index) pairs that were ever on the
	// node in the requested time span.
	//
	// Logic should more or less follow cmd/gpus (the "gpu" command).

	gsd, err := gpusample.OpenGpuSampleDataProvider(meta)
	if err != nil {
		return nil, huma.Error500InternalServerError(
			nodesGpuTimeseriesName+": failed to open gpu sample store", err)
	}
	hostGlobber, hErr := newHosts(nodesGpuTimeseriesName, input.Nodename)
	if err != nil {
		return nil, hErr

	}

	// streams : GpuSamplesByHostSet
	// GpuSamplesByHostSet = map[Ustr]*GpuSamplesByHost (map key is the HostName of the value)
	// GpuSamplesByHost = struct {HostName, Data: []GpuSamples}
	// GpuSamples = struct { Time, Decoded: []repr.PerGpuSample } (one sample per GPU at this time, ascending by time)
	// PerGpuSample = struct {Attr, newfmt.SampleGpu }
	// SampleGpu = struct {Index, UUID, CEUtil, MemoryUtil, ...} (very many fields)
	streams, _, _, _, err := gsd.Query(from, to, hostGlobber, verbose)
	if err != nil {
		return nil, huma.Error500InternalServerError(
			nodesGpuTimeseriesName+": Failed to query gpu sample data", err)
	}

	resp := &NodesGpuTimeseriesResponse{
		Body: make(map[string][]NodesGpuTimeseries_Card),
	}
	for node, samplesByHost := range streams {
		nodeName := node.String()

		// Transform stream-of-cards-across-time to streams-of-card-across-time, on the one node.
		type key struct {
			uuid  string
			index uint64
		}
		type sampleBox struct {
			time int64
			*repr.PerGpuSample
		}
		perCard := make(map[key][]sampleBox)
		for _, samples := range samplesByHost.Data {
			t := samples.Time
			for _, sample := range samples.Decoded {
				k := key{uuid: string(sample.UUID), index: sample.Index}
				perCard[k] = append(perCard[k], sampleBox{time: t, PerGpuSample: &sample})
			}
		}

		// Build time series for all cards (= unique (uuid,index) pairs) on this node.
		for k, v := range perCard {
			var card NodesGpuTimeseries_Card
			card.UUID = k.uuid
			card.Index = k.index
			t := canonicalizeInitialTimestep(v[0].time, bucket)
			i := 0
			for i < len(v) {
				var point NodesGpuTimeseries_CardPoint
				var n int
				for i < len(v) && v[i].time < t+bucket {
					point.CEUtil += float64(v[i].CEUtil)
					point.MemUtil += float64(v[i].MemoryUtil)
					n++
					i++
				}
				if n > 1 {
					point.CEUtil /= float64(n)
					point.MemUtil /= float64(n)
				}
				point.Time = time.Unix(t, 0).Format(time.RFC3339)
				point.CEUtil = onePlace(point.CEUtil)
				point.MemUtil = onePlace(point.MemUtil)
				card.Data = append(card.Data, point)
				t += bucket
			}
			resp.Body[nodeName] = append(resp.Body[nodeName], card)
		}
	}

	return resp, nil
}
