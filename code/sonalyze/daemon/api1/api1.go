// The v1 API follows the old v0 API but:
//
//  - there's better adherence to REST API design principles in the API names, see below
//  - GET requests return JSON objects instead of strings
//  - the returned JSON objects carry non-string values for non-string fields
//  - there are new/different insertion points for the "new data format", represented as JSON
//  - the result of a POST is a JSON object with some data about the data that were received
//
// Regarding the API naming, the operations are all sensible plural nouns:
//
//   /clusters        cluster data (from metadata)
//   /cards/{cluster} card data (from sysinfo data)
//   /jobs/{cluster}  job data (from samples, slurm and sysinfo data)
//
// Apart from the cluster name, all request parameters are HTTP query parameters.  For initial
// record selection, there are the "coarse" query parameters:
//
//   start_time_in_s   Start of query window, seconds since Posix epoch, default now-1h
//   end_time_in_s     End of query window, seconds since Posix epoch, default now
//   start_date        Start of query window, date in local time zone, no default
//   end_date          End of query window, date in local time zone, no default
//   node              Comma-separated list of SLURM-style compressed node names and ranges
//
// The start/end date, if present, are translated to start/end times, and then checked/clamped.
//
// The entities produced by the initial query and subsequent processing can be further filtered by
// including a query term (there are currently no ad-hoc / per-operation record filters as in the v0
// API) selecting certain values of the JSON response fields:
//
//   query             Expression in the query language, see the manual or ../../table/queryexpr.y
//
// Finally, it is possible to ask for specific output fields (instead of all fields), to keep the
// data volume down:
//
//   fields            Comma-separated list of JSON field names
//
// For example (formatted for readability and ignoring proper HTTP escaping):
//
//   /api/v0/jobs/fox.educloud.no?
//     start_date=2026-05-17&
//     end_date=2026-05-19&
//     node=gpu-[1,2,8,9],c1-[12-15]&
//     query=User=larstha and SomeGpu=true&
//     fields=Job,CpuAvgPct,GpuAvgPct,Cmd
//
// The list of JSON field names in each response (for query and fields) can be obtained by reading
// the source in each `respond.go` file in each subdirectory here, or by examining the REST API spec
// that is obtained by asking for openapi.json or openapi.yaml on the root interface.

package api1

import (
	"go-utils/auth"

	"github.com/danielgtaylor/huma/v2"

	"sonalyze/daemon/api1/cards"
	"sonalyze/daemon/api1/clusters"
	"sonalyze/daemon/api1/common"
	"sonalyze/daemon/api1/insert"
	"sonalyze/daemon/api1/jobs"
)

func SetupAPI(
	api huma.API,
	insertAPI bool,
	getAuthenticator_ *auth.Authenticator,
	postAuthenticator_ *auth.Authenticator,
) {
	common.GetAuthenticator = getAuthenticator_
	common.PostAuthenticator = postAuthenticator_
	grp := huma.NewGroup(api, "/api/v1")

	cards.AddCard(grp)
	clusters.AddCluster(grp)
	jobs.AddJobs(grp)

	if insertAPI {
		insert.AddInsertSysinfoData(grp)
		insert.AddInsertSampleData(grp)
		insert.AddInsertJobData(grp)
		insert.AddInsertClusterData(grp)
	}
}
