package apiutil

import (
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"

	. "sonalyze/common"
	"sonalyze/data/common"
	"sonalyze/db"
	"sonalyze/db/special"
	"sonalyze/db/types"
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
)

const (
	apiName    = "sonalyze API"
	apiVersion = "2"
)

var router humago.Mux
var iface string

func CreateAPI(iface_ string) huma.API {
	iface = iface_
	router = http.NewServeMux()
	// This sucks - one version to rule them all even for /api/vx.  But in that case it's best to
	// report the highest version here.
	return humago.New(router, huma.DefaultConfig(apiName, apiVersion))
}

func RunAPI() {
	// FIXME - Not quite what we want, cf our http abstraction, but OK for now
	go http.ListenAndServe(iface, router)
}

func StopAPI() {
	// FIXME - implement StopAPI()
}

func GetClusterContext(opName, clusterName string) (types.Context, huma.StatusError) {
	cluster := special.LookupCluster(clusterName)
	if cluster == nil {
		return nil, huma.Error400BadRequest(opName + ": Failed to find cluster " + clusterName)
	}
	return db.NewContextFromCluster(cluster), nil
}

// Given a cluster, compute the from/to time based on the available data in the database for the cluster
// and any expressed from/to times.
func TimeWindowFromData(
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

func NewHostFilter(
	opName string,
	meta types.Context,
	multiPatterns string,
	from, to time.Time,
) (Multihost, huma.StatusError) {
	hostFilter, err := NewHostsFromMultiPatterns(multiPatterns)
	if err != nil {
		return Multihost{}, huma.Error400BadRequest(opName+": Bad host list", err)
	}
	host, err := common.ResolveWildcard(meta, hostFilter, from, to)
	if err != nil {
		return Multihost{}, huma.Error400BadRequest(opName+": Bad host list", err)
	}
	return host, nil
}
