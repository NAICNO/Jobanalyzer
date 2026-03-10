package restapi

import (
	"cmp"
	"context"
	"fmt"
	_ "math"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"

	. "sonalyze/common"
	"sonalyze/data/common"
	"sonalyze/data/disksample"
	"sonalyze/db"
	"sonalyze/db/repr"
	"sonalyze/db/special"
)

type NodesDiskstatsTimeseriesResponse struct {
	Body map[string][]NodesDiskstatsTimeseries_Disk
}

type NodesDiskstatsTimeseries_Disk struct {
	Major uint64                                `json:"major"`
	Minor uint64                                `json:"minor"`
	Name  string                                `json:"name"`
	Data  []NodeDiskstatsTimeseries_DiskDetails `json:"data"`
}

type NodeDiskstatsTimeseries_DiskDetails struct {
	DeltaTime               uint64 `json:"delta_time_in_s"`
	DiscardsCompleted       uint64 `json:"discards_completed"`
	DiscardsMerged          uint64 `json:"discards_merged"`
	FlushRequestsCompleted  uint64 `json:"flush_requests_completed"`
	IOsCurrentlyInProgress  uint64 `json:"ios_currently_in_progress"`
	MsSpentDiscarding       uint64 `json:"ms_spent_discarding"`
	MsSpentDoingIos         uint64 `json:"ms_spent_doing_ios"`
	MsSpentFlushing         uint64 `json:"ms_spent_flushing"`
	MsSpentReading          uint64 `json:"ms_spent_reading"`
	MsSpentWriting          uint64 `json:"ms_spent_writing"`
	ReadsCompleted          uint64 `json:"reads_completed"`
	ReadsMerged             uint64 `json:"reads_merged"`
	SectorsDiscarded        uint64 `json:"sectors_discarded"`
	SectorsRead             uint64 `json:"sectors_read"`
	SectorsWritten          uint64 `json:"sectors_written"`
	Time                    uint64 `json:"time"`
	WeightedMsSpentDoingIOs uint64 `json:"weighted_ms_spent_doing_ios"`
	WritesCompleted         uint64 `json:"writes_completed"`
	WritesMerged            uint64 `json:"writes_merged"`
}

func addNodesDiskstatsTimeseries(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "get-nodes-diskstats-timeseries",
			Method:      http.MethodGet,
			Path:        "/cluster/{cluster}/nodes/diskstats/timeseries",
			Summary:     "Compute timeseries data of disk samples for nodes in a given cluster",
		},
		handleNodesDiskstatsTimeseries,
	)
}

func handleNodesDiskstatsTimeseries(
	ctx context.Context,
	input *struct {
		Cluster       string `path:"cluster" example:"my.cluster.name" doc:"Name of cluster"`
		StartTimeInS  uint64 `query:"start_time_in_s" doc:"Posix timestamp"`
		EndTimeInS    uint64 `query:"end_time_in_s" doc:"Posix timestamp"`
		ResolutionInS uint64 `query:"resolution_in_s" doc:"Default is 300"`
		Nodename      string `query:"nodename" doc:"Compressed node name list"`
	},
) (*NodesDiskstatsTimeseriesResponse, error) {
	clusterName := input.Cluster
	cluster := special.LookupCluster(clusterName)
	if cluster == nil {
		return nil, fmt.Errorf("Failed to find cluster %s", clusterName)
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

	dsdp, err := disksample.OpenDiskSampleDataProvider(meta)
	if err != nil {
		return nil, err
	}
	var hostList []string
	if input.Nodename != "" {
		hostList = []string{input.Nodename}
	}
	samples, err := dsdp.Query(
		common.QueryFilter{HaveFrom: true, FromDate: from, HaveTo: true, ToDate: to, Host: hostList},
		verbose,
	)
	if err != nil {
		return nil, err
	}

	// Partition by node and disk name
	type key struct {
		node Ustr
		name Ustr
	}
	disks := make(map[key][]*repr.DiskSample)
	for _, s := range samples {
		disks[key{s.Hostname, s.Name}] = append(disks[key{s.Hostname, s.Name}], s)
	}

	resp := &NodesDiskstatsTimeseriesResponse{
		Body: make(map[string][]NodesDiskstatsTimeseries_Disk),
	}
	for k, samples := range disks {
		slices.SortFunc(samples, func(a, b *repr.DiskSample) int {
			return cmp.Compare(a.Timestamp, b.Timestamp)
		})

		// See comment in nodes_cpu_timeseries.go about maybe rounding the initial time value for
		// bucketing.
		var (
			t        = samples[0].Timestamp
			i        = 0
			stepData []NodeDiskstatsTimeseries_DiskDetails
		)
		for i < len(samples) {

			// Average all quantitites
			var (
				acc NodeDiskstatsTimeseries_DiskDetails
				n   uint64
			)
			for i < len(samples) && samples[i].Timestamp < t+int64(bucket) {
				s := samples[i]
				acc.DiscardsCompleted += s.DiscardsCompleted
				acc.DiscardsMerged += s.DiscardsMerged
				acc.FlushRequestsCompleted += s.FlushesCompleted
				acc.IOsCurrentlyInProgress += s.IOsInProgress
				acc.MsSpentDiscarding += s.MsDiscarding
				acc.MsSpentDoingIos += s.MsDoingIO
				acc.MsSpentFlushing += s.MsFlushing
				acc.MsSpentReading += s.MsReading
				acc.MsSpentWriting += s.MsWriting
				acc.ReadsCompleted += s.ReadsCompleted
				acc.ReadsMerged += s.ReadsMerged
				acc.SectorsDiscarded += s.SectorsDiscarded
				acc.SectorsRead += s.SectorsRead
				acc.SectorsWritten += s.SectorsWritten
				acc.WeightedMsSpentDoingIOs += s.WeightedMsDoingIO
				acc.WritesCompleted += s.WritesCompleted
				acc.WritesMerged += s.WritesMerged
				n++
				i++
			}
			if n > 1 {
				acc.DiscardsCompleted /= n
				acc.DiscardsMerged /= n
				acc.FlushRequestsCompleted /= n
				acc.IOsCurrentlyInProgress /= n
				acc.MsSpentDiscarding /= n
				acc.MsSpentDoingIos /= n
				acc.MsSpentFlushing /= n
				acc.MsSpentReading /= n
				acc.MsSpentWriting /= n
				acc.ReadsCompleted /= n
				acc.ReadsMerged /= n
				acc.SectorsDiscarded /= n
				acc.SectorsRead /= n
				acc.SectorsWritten /= n
				acc.WeightedMsSpentDoingIOs /= n
				acc.WritesCompleted /= n
				acc.WritesMerged /= n
			}

			// Record
			acc.Time = uint64(t)
			acc.DeltaTime = bucket
			stepData = append(stepData, acc)

			t += int64(bucket)
		}
		node := k.node.String()
		name := k.name.String()
		resp.Body[node] = append(resp.Body[node], NodesDiskstatsTimeseries_Disk{
			Name:  name,
			Major: samples[0].Major,
			Minor: samples[0].Minor,
			Data:  stepData,
		})
	}
	return resp, nil
}
