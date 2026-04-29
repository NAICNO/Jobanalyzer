package api2

import (
	"cmp"
	"context"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"

	. "sonalyze/common"
	"sonalyze/daemon/apiutil"
	"sonalyze/data/common"
	"sonalyze/data/disksample"
	"sonalyze/db/repr"
)

const nodesDiskstatsTimeseriesName = "/cluster/{cluster}/nodes/diskstats/timeseries"

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
			Path:        nodesDiskstatsTimeseriesName,
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
	meta, hErr := apiutil.GetClusterContext(nodesDiskstatsTimeseriesName, input.Cluster)
	if hErr != nil {
		return nil, hErr
	}
	from, to, hErr := apiutil.TimeWindowFromData(
		nodesDiskstatsTimeseriesName, meta, input.StartTimeInS, input.EndTimeInS)
	if hErr != nil {
		return nil, hErr
	}
	bucket := int64(300)
	if input.ResolutionInS != 0 {
		bucket = int64(input.ResolutionInS)
	}

	dsdp, err := disksample.OpenDiskSampleDataProvider(meta)
	if err != nil {
		return nil, huma.Error500InternalServerError(
			nodesDiskstatsTimeseriesName+": Failed to open disk sample store", err)
	}
	var hostList []string
	if input.Nodename != "" {
		hostList = []string{input.Nodename}
	}
	samples, err := dsdp.Query(
		common.QueryFilter{HaveFrom: true, FromDate: from, HaveTo: true, ToDate: to, Host: hostList},
	)
	if err != nil {
		return nil, huma.Error500InternalServerError(
			nodesDiskstatsTimeseriesName+": Failed to query disk sample store", err)
	}

	// Partition the data by node and disk name
	type key struct {
		node Ustr
		name Ustr
	}
	disks := make(map[key][]*repr.DiskSample)
	for _, s := range samples {
		disks[key{s.Hostname, s.Name}] = append(disks[key{s.Hostname, s.Name}], s)
	}

	// Aggregate per (node,disk) pair
	resp := &NodesDiskstatsTimeseriesResponse{
		Body: make(map[string][]NodesDiskstatsTimeseries_Disk),
	}
	for k, samples := range disks {
		slices.SortFunc(samples, func(a, b *repr.DiskSample) int {
			return cmp.Compare(a.Timestamp, b.Timestamp)
		})

		var (
			t        = canonicalizeInitialTimestep(samples[0].Timestamp, bucket)
			i        = 0
			stepData []NodeDiskstatsTimeseries_DiskDetails
		)
		for i < len(samples) {

			// Average all quantitites
			var (
				acc NodeDiskstatsTimeseries_DiskDetails
				n   uint64
			)
			for i < len(samples) && samples[i].Timestamp < t+bucket {
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
			acc.DeltaTime = uint64(bucket)
			stepData = append(stepData, acc)

			t += bucket
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
