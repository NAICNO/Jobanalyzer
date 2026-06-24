package api2

import (
	"math"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"go-utils/gpuset"
	. "sonalyze/common"
	"sonalyze/data/card"
	"sonalyze/data/common"
	"sonalyze/data/node"
	"sonalyze/data/sample"
	"sonalyze/db/repr"
	"sonalyze/db/types"
)

const (
	apiName    = "slurm-monitor REST API"
	apiVersion = "2"
)

// Time windows for searching for data corresponding to something else.  These ought to be
// parameters, probably, not hardcoded.
const (
	// By default the time window is the last hour before "now", or latest datum available in the
	// database.  This is compatible with slurm-monitor.
	defaultTimeWindow = 1 * time.Hour

	// The max time window for searching is 2 weeks, or we risk overloading the server.  This too is
	// compatible with slurm-monitor.
	maxTimeWindow = 24 * time.Hour * 14

	// We assume that sysinfo is collected more often that this.  Slurm-monitor uses 12h, which is
	// too little.
	sysinfoWindow = 24 * time.Hour
)

// The iface is the local interface: name:port.
func SetupAPI(api huma.API) {
	grp := huma.NewGroup(api, "/api/v2")
	addErrorMessages(grp)
	addListClusters(grp)
	addNodesInfo(grp)
	addNodesLastProbeTimestamp(grp)
	addNodesCpuTimeseries(grp)
	addNodesMemoryTimeseries(grp)
	addNodesGpuTimeseries(grp)
	addNodesDiskstatsTimeseries(grp)
	addNodesProcessGpuUtil(grp)
	addProcesses(grp)
	addProcessesGpu(grp)
	addProcessesTimeseries(grp)
}

// This is called from the daemon's main thread when interrupted by signals.
func StopRestAPI() {
	// TODO: Implement StopRestAPI().
	//
	// It probably means changing how we do listening in the function above.
}

// Clean up the time for the first time stamp in a time series.  Normally this is the time of the
// first record, adjusted somehow.
func canonicalizeInitialTimestep(t int64, resolution int64) int64 {
	// Just return the starting time unadjusted, for now.
	//
	// TODO: Round down first time value of time series?
	//
	// There's an argument to be made that the first time value in the series should be rounded down
	// to something so that time steps are predictable and comparable across hosts, and not
	// dependent on when data came in.  Note any bucket value is acceptable.  The most sane
	// algorithm might be to take the value and round down to minute (if < 60), 5-minute (if < 600),
	// 15-minute, 30-minute, 60-minute, 2-hour, 4-hour, 8-hour, 24-hour, etc.
	//
	// TODO: Also want to document resolution_in_s.
	//
	// If this is less than the sampling frequency then there *will* be data points in time series
	// with zero values, the way things are set up now.  Ideally the resolution is the frequency or
	// some multiple of the frequency, or things may look weird / and/or need averaging on the
	// presentation side.
	return t
}

func openSampleDataProvider(opName string, meta types.Context) (*sample.SampleDataProvider, huma.StatusError) {
	sdp, err := sample.OpenSampleDataProvider(meta)
	if err != nil {
		return nil, huma.Error500InternalServerError(
			opName+": Failed to open sample store", err)
	}
	return sdp, nil
}

// Retrieve latest node metadata for the nodes within the time window.
func getSysinfoAt(
	opName string,
	meta types.Context,
	to time.Time,
	host Multihost,
) (map[string]*repr.SysinfoNodeData, huma.StatusError) {
	ndp, err := node.OpenNodeDataProvider(meta)
	if err != nil {
		return nil, huma.Error500InternalServerError(opName+": Failed to open node store", err)
	}
	from := to.Add(-sysinfoWindow)
	nodes, err := ndp.Query(
		common.QueryFilter{HaveFrom: true, FromDate: from, HaveTo: true, ToDate: to, Host: host},
	)
	if err != nil {
		return nil, huma.Error500InternalServerError(opName+": Failed to query node data", err)
	}
	nodeMap := make(map[string]*repr.SysinfoNodeData)
	for _, n := range nodes {
		if probe := nodeMap[n.Node]; probe != nil {
			if n.Time < probe.Time {
				continue
			}
		}
		nodeMap[n.Node] = n
	}
	return nodeMap, nil
}

func getCardInfoByUUIDAt(
	opName string,
	meta types.Context,
	to time.Time,
	host Multihost,
) (map[string]*repr.SysinfoCardData, huma.StatusError) {
	cdp, err := card.OpenCardDataProvider(meta)
	if err != nil {
		return nil, huma.Error500InternalServerError(opName+": Failed to open card store", err)
	}
	from := to.Add(-sysinfoWindow)
	records, err :=
		cdp.Query(
			card.QueryFilter{HaveFrom: true, FromDate: from, HaveTo: true, ToDate: to, Host: host},
		)
	if err != nil {
		return nil, huma.Error500InternalServerError(opName+": Failed to query card data", err)
	}
	cardMap := make(map[string]*repr.SysinfoCardData)
	for _, c := range records {
		if probe := cardMap[c.UUID]; probe != nil {
			if c.Time < probe.Time {
				continue
			}
		}
		cardMap[c.UUID] = c
	}
	return cardMap, nil
}

// Cards are unsorted in each node's slice.
func getCardInfoByNodeAt(
	opName string,
	meta types.Context,
	to time.Time,
	host Multihost,
) (map[string][]*repr.SysinfoCardData, huma.StatusError) {
	cardInfo, hErr := getCardInfoByUUIDAt(opName, meta, to, host)
	if hErr != nil {
		return nil, hErr
	}
	cardsByNode := make(map[string][]*repr.SysinfoCardData)
	for _, c := range cardInfo {
		cardsByNode[c.Node] = append(cardsByNode[c.Node], c)
	}
	return cardsByNode, nil
}

// Translate an index set to a index-sorted card set.  The assumption is that the `cards` are all
// from the same node as the index set, at the same time.  Normally the `cards` are all the cards on
// the node, unsorted, and the `gpus` represent cards used by a process.
func gpuSetToGpus(gpus gpuset.GpuSet, cards []*repr.SysinfoCardData) []*repr.SysinfoCardData {
	var result []*repr.SysinfoCardData
	if !gpus.IsUnknown() && cards != nil {
		for _, ix := range gpus.AsSlice() {
			for _, c := range cards {
				if c.Index == uint64(ix) {
					result = append(result, c)
					break
				}
			}
		}
	}
	return result
}

func formatTime(t int64) string {
	return time.Unix(t, 0).UTC().Format(time.RFC3339)
}

func onePlace(f float64) float64 {
	return math.Round(f*10) / 10
}
