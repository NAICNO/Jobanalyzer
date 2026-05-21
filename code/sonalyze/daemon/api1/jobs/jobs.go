package jobs

import (
	"context"

	"github.com/danielgtaylor/huma/v2"

	"sonalyze/daemon/apiutil"
	_ "sonalyze/db/special"
)

// TODO: And here we might also want a "job" that always takes an ID, to just look at one,
// instead of passing the job ID as a query parameter?

//go:generate ../../../../generate-response/generate-response jobs.go

/*RESPONSE

package jobs

import (
	"go-utils/gpuset"
	. "sonalyze/cmd/jobs"
	. "sonalyze/common"
	"sonalyze/daemon/apiutil"
	. "sonalyze/table"
)

%%

TYPE     Jobs_Job
TABLE    ../../../cmd/jobs/print.go
DEFAULTS Job,User,Duration,Hosts,CpuTime,ResidentMemAvgGB,GpuTime,GpuMemAvgGB,Cmd

ESNOPSER*/

const jobsCommandName = "/cluster/{cluster}/jobs"

type JobsResponse struct {
	Body []Jobs_Job
}

func AddJobs(api huma.API) {
	huma.Get(api, jobsCommandName, func(
		ctx context.Context,
		input *struct {
			apiutil.AuthHeader
			Fields string `query:"fields" doc:"List of JSON field names"`
		},
	) (*JobsResponse, error) {
		panic("NYI")
		// Query, which requires a bit of refactoring in the jobs code
		//
		// And then formatting
		//
		// TODO: Lots of internal types in the generated code, which really
		// must be mapped (somewhere) to something else.  And that will
		// require casts in the copy code.
		//
		// Notably Ustr fields require actual computation.  The others may or may
		// not be transparent to the JSON code.  Hostnames may not be.
		// Gpu set may not be.
	})
}
