package restapi

import (
	"math"
	"net/http"
	"os"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"

	. "sonalyze/common"
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
	// By default the time window is the last hour before "now", or latest datum available in the
	// database.  This is compatible with slurm-monitor.
	defaultTimeWindow = 1 * time.Hour
	// The max time window is 2 weeks, or we risk overloading the server.
	maxTimeWindow = 24 * time.Hour * 14
)

var verbose = os.Getenv("SONALYZE_REST_VERBOSE") == "1"

// The iface is the local interface: name:port.  Use this only as `go RestAPI(iface)`.
func RestAPI(iface string) {
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
	maxTime, err := theLog.MaxTime(true, verbose)
	if err != nil {
		maxTime = time.Now()
	}
	minTime, err := theLog.MinTime(true, verbose)
	if err != nil {
		minTime = maxTime
	}
	if verbose {
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

func newHosts(opName, nodeName string) (*Hosts, huma.StatusError) {
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

// Retrieve node metadata for all the nodes on the cluster within the time window.
func getNodeMap(
	opName string,
	meta types.Context,
	from, to time.Time,
	hostList []string,
) (map[string]*repr.SysinfoNodeData, huma.StatusError) {
	ndp, err := node.OpenNodeDataProvider(meta)
	if err != nil {
		return nil, huma.Error500InternalServerError(opName+": Failed to open node store", err)
	}
	nodes, err := ndp.Query(
		common.QueryFilter{HaveFrom: true, FromDate: from, HaveTo: true, ToDate: to, Host: hostList},
		verbose,
	)
	if err != nil {
		return nil, huma.Error500InternalServerError(opName+": Failed to query node data", err)
	}
	nodeMap := make(map[string]*repr.SysinfoNodeData)
	for _, n := range nodes {
		if probe := nodeMap[n.Node]; probe != nil {
			if n.Time > probe.Time {
				nodeMap[n.Node] = n
			}
		} else {
			nodeMap[n.Node] = n
		}
	}
	return nodeMap, nil
}

func onePlace(f float64) float64 {
	return math.Round(f*10) / 10
}
