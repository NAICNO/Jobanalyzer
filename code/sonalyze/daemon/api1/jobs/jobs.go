package jobs

import (
	"context"

	"github.com/danielgtaylor/huma/v2"

	"sonalyze/cmd/jobs"
	"sonalyze/daemon/api1/common"
	"sonalyze/daemon/apiutil"
	_ "sonalyze/db/special"
)

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

const jobsCommandName = "/jobs/{cluster}"

type JobsResponse struct {
	Body []Jobs_Job
}

func AddJobs(api huma.API) {
	huma.Get(api, jobsCommandName, func(
		ctx context.Context,
		input *common.StandardQueryFields,
	) (*JobsResponse, error) {
		meta, from, to, nodes, query, hErr := input.Parameters(jobsCommandName)
		if hErr != nil {
			return nil, hErr
		}
		// This is different from the sonalyze command line in that there are no command-line record
		// filters other than the date range and hosts.  All other filtering must be expressed in
		// terms of the query filter.
		//
		// TODO: This is much too simplistic, there are many more parameters that control how jobs
		// are created.
		//
		// TODO: We *probably* want to improve that: query by job ID and user ID will be very common
		// and will greatly reduce the load on the back-end if we can filter those quickly at the
		// record level.
		records, err := jobs.Query(
			meta,
			jobs.QueryFilter{
				HaveFrom: true,
				FromDate: from,
				HaveTo:   true,
				ToDate:   to,
				Host:     nodes.Patterns(),
			},
			query,
		)
		if err != nil {
			return nil, huma.Error500InternalServerError(
				jobsCommandName+": Failed to query jobs data", err)
		}
		flds := apiutil.Fields(input.Fields, responseDefaults)
		jobs := make([]Jobs_Job, 0, len(records))
		for _, r := range records {
			jobs = append(jobs, respond(flds, r))
		}
		return &JobsResponse{Body: jobs}, nil
	})
}
