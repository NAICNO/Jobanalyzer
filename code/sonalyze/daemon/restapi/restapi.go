package restapi

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"

	"sonalyze/data/common"
	"sonalyze/data/node"
	"sonalyze/db/repr"
	"sonalyze/db/types"
)

const (
	apiName    = "slurm-monitor REST API"
	apiVersion = "2"
	timeWindow = 14 // days
)

var verbose = os.Getenv("SONALYZE_REST_VERBOSE") == "1"

// The iface is the local interface: name:port.  Use this only as `go RestAPI(iface)`.
func RestAPI(iface string) {
	router := http.NewServeMux()

	// TODO: Want to somehow document default timespan.  Can we attach that to the api somehow
	// without repeating it for every API?
	//
	// The default timespan should always be end-of-data minus the timeWindow; for some databases
	// this may then be rounded down to start-of-day for the starting time and end-of-day for the
	// ending time, so for a 24h window (say) we'd see data for two days.  (For live clusters
	// end-of-data will tend to be "now".)  But then of course the timespan is clamped to the
	// available data.  And when a specific time_in_s time is provided to an API we'll see data for
	// that day only.  And when specific start and end times are provided they should be rounded
	// down and up to start-of-day and end-of-day respectively, at least for some operations.
	//
	// TODO: Also want to document resolution_in_s.  If this is less than the sampling frequency
	// then there *will* be data points in time series with zero values, the way things are set up
	// now.  Ideally the resolution is the frequency or some multiple of the frequency, or things
	// may look weird / and/or need averaging on the presentation side.
	api := humago.New(router, huma.DefaultConfig(apiName, apiVersion))

	addErrorMessages(api)
	addListClusters(api)
	addNodesInfo(api)
	addNodesLastProbeTimestamp(api)
	addNodesCpuTimeseries(api)
	addNodesMemoryTimeseries(api)
	addNodesGpuTimeseries(api)
	addNodesDiskstatsTimeseries(api)
	addNodesProcessGpuUtil(api)
	addProcesses(api)
	addProcessesGpu(api)
	addProcessesTimeseries(api)

	http.ListenAndServe(iface, router)
}

// FIXME: Time window from meta
func timeWindowFromData(
	meta types.Context,
	startTimeInS, endTimeInS uint64,
) (from time.Time, to time.Time, err error) {
	to = time.Now()
	from = to.AddDate(0, 0, -timeWindow)
	if startTimeInS != 0 {
		from = time.Unix(int64(startTimeInS), 0)
	}
	if endTimeInS != 0 {
		to = time.Unix(int64(endTimeInS), 0)
	}
	return
}

func getNodeMap(
	meta types.Context,
	from, to time.Time,
	hostList []string,
) (map[string]*repr.SysinfoNodeData, error) {
	ndp, err := node.OpenNodeDataProvider(meta)
	if err != nil {
		return nil, err
	}
	nodes, err := ndp.Query(
		common.QueryFilter{HaveFrom: true, FromDate: from, HaveTo: true, ToDate: to, Host: hostList},
		verbose,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to read config log records: %v", err)
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
