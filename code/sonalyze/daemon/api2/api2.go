package api2

import (
	"math"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"

	"go-utils/gpuset"
	. "sonalyze/common"
	"sonalyze/data/card"
	"sonalyze/data/common"
	"sonalyze/data/node"
	"sonalyze/data/sample"
	"sonalyze/db"
	"sonalyze/db/repr"
	"sonalyze/db/special"
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
func StartRestAPI(iface string) {
	go func() {
		router := http.NewServeMux()
		api := humago.New(router, huma.DefaultConfig(apiName, apiVersion))
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
		http.ListenAndServe(iface, router)
	}()
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

func getClusterContext(opName, clusterName string) (types.Context, huma.StatusError) {
	cluster := special.LookupCluster(clusterName)
	if cluster == nil {
		return nil, huma.Error400BadRequest(opName + ": Failed to find cluster " + clusterName)
	}
	return db.NewContextFromCluster(cluster), nil
}

// Given a cluster, compute the from/to time based on the available data in the database for the cluster
// and any expressed from/to times.
func timeWindowFromData(
	opName string,
	meta types.Context,
	startTimeInS, endTimeInS uint64,
) (from time.Time, to time.Time, hErr huma.StatusError) {
	// TODO: Want to somehow document default timespan.
	//
	// Can we attach that to the api somehow without repeating it for every API?

	theLog, err := db.OpenReadOnlyDB(meta, types.MetaData)
	if err != nil {
		hErr = huma.Error500InternalServerError(opName+": Can't open database", err)
		return
	}
	maxTime, err := theLog.MaxTime(true)
	if err != nil {
		maxTime = time.Now()
	}
	minTime, err := theLog.MinTime(true)
	if err != nil {
		minTime = maxTime
	}
	if Verbose {
		Log.Infof("Min/max time: %v %v", minTime, maxTime)
	}

	// Sensible defaults
	to = maxTime
	from = maxTime.Add(-defaultTimeWindow)

	// Overrides - start/end can be specified separately
	if startTimeInS != 0 {
		from = time.Unix(int64(startTimeInS), 0)
		if endTimeInS == 0 {
			to = from.Add(defaultTimeWindow)
		}
	}
	if endTimeInS != 0 {
		to = time.Unix(int64(endTimeInS), 0)
		if startTimeInS == 0 {
			from = to.Add(-defaultTimeWindow)
		}
	}

	// Validation
	if from.After(to) {
		hErr = huma.Error400BadRequest(opName+": Bad time value(s)", err)
		return
	}

	// Clamping to max window
	if to.Sub(from) > maxTimeWindow {
		from = to.Add(-maxTimeWindow)
	}

	// Clamping to max/min times
	if from.Before(minTime) {
		from = minTime
	}
	if to.After(maxTime) {
		to = maxTime
	}

	return
}

func openSampleDataProvider(opName string, meta types.Context) (*sample.SampleDataProvider, huma.StatusError) {
	sdp, err := sample.OpenSampleDataProvider(meta)
	if err != nil {
		return nil, huma.Error500InternalServerError(
			opName+": Failed to open sample store", err)
	}
	return sdp, nil
}

func newHostFilter(opName, nodeName string) (*Hosts, huma.StatusError) {
	var hostList []string
	if nodeName != "" {
		hostList = []string{nodeName}
	}
	hostFilter, err := NewHosts(true, hostList)
	if err != nil {
		return nil, huma.Error400BadRequest(opName+": Bad host list", err)
	}
	return hostFilter, nil
}

// Retrieve latest node metadata for the nodes within the time window.
func getSysinfoAt(
	opName string,
	meta types.Context,
	to time.Time,
	hostList []string,
) (map[string]*repr.SysinfoNodeData, huma.StatusError) {
	ndp, err := node.OpenNodeDataProvider(meta)
	if err != nil {
		return nil, huma.Error500InternalServerError(opName+": Failed to open node store", err)
	}
	from := to.Add(-sysinfoWindow)
	nodes, err := ndp.Query(
		common.QueryFilter{HaveFrom: true, FromDate: from, HaveTo: true, ToDate: to, Host: hostList},
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
	hostList []string,
) (map[string]*repr.SysinfoCardData, huma.StatusError) {
	cdp, err := card.OpenCardDataProvider(meta)
	if err != nil {
		return nil, huma.Error500InternalServerError(opName+": Failed to open card store", err)
	}
	from := to.Add(-sysinfoWindow)
	records, err :=
		cdp.Query(
			card.QueryFilter{HaveFrom: true, FromDate: from, HaveTo: true, ToDate: to, Host: hostList},
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
	hostList []string,
) (map[string][]*repr.SysinfoCardData, huma.StatusError) {
	cardInfo, hErr := getCardInfoByUUIDAt(opName, meta, to, hostList)
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
