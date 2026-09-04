package jobs

import (
	"context"

	"github.com/danielgtaylor/huma/v2"

	"sonalyze/cmd/jobs"
	"sonalyze/daemon/api1/common"
	dcommon "sonalyze/data/common"
)

//go:generate ../../../../generate-response/generate-response jobs.go

/*RESPONSE

package jobs

import (
	. "sonalyze/cmd/jobs"
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
		input *struct {
			common.StandardQueryFields
			// Extra record filter fields for sample data.  (The CLI also has ExcludeJob,
			// ExcludeCommand, ExcludeSystemUsers but they never were useful and sometimes were not
			// well-defined.  ExcludeSystemJobs + ExcludeUser covers most of it.)
			Job               string `query:"job" doc:"List of job IDs"`
			User              string `query:"user" doc:"List of user names to include (exact match)"`
			ExcludeUser       string `query:"exclude_user" doc:"List of user names to exclude (exact match)"`
			Command           string `query:"command" doc:"List of commands (exact match)"`
			ExcludeSystemJobs bool   `query:"exclude_system_jobs" doc:"Exclude processes with UID < 1000"`
			MergeAll          bool   `query:"merge_all" doc:"Specialized: Merge all sample streams"`
			MergeNone         bool   `query:"merge_none" doc:"Specialized: Merge no sample streams"`
			SacctFromSonar    bool   `query:"sacct_from_sonar" doc:"Specialized: synthesize sacct data from Sonar data"`
		},
	) (*JobsResponse, error) {
		meta, from, to, nodes, query, flds, hErr := input.Parameters(jobsCommandName, responseDefaults)
		if hErr != nil {
			return nil, hErr
		}
		users := common.StringList(input.User)
		excludeUsers := common.StringList(input.ExcludeUser)
		commands := common.StringList(input.Command)
		jobIds, err := common.UintList[uint32](input.Job)
		if err != nil {
			return nil, huma.Error400BadRequest("Non-numeric job ID: " + err.Error())
		}
		// This is different from the sonalyze command line in that there are fewer command-line
		// filters.  Fine-grained (jobs) filtering must be expressed in terms of the query filter
		// that applies to jobs.  This is OK: the large range of command line filters in the CLI was
		// a consequence of not having a query filter originally.
		records, err := jobs.Query(
			meta,
			jobs.QueryFilter{
				QueryFilter: dcommon.QueryFilter{
					HaveFrom: true,
					FromDate: from,
					HaveTo:   true,
					ToDate:   to,
					Host:     nodes,
				},
				MergeAll:          input.MergeAll,
				MergeNone:         input.MergeNone,
				SacctFromSonar:    input.SacctFromSonar,
				Jobs:              jobIds,
				Users:             users,
				ExcludeUsers:      excludeUsers,
				Commands:          commands,
				ExcludeSystemJobs: input.ExcludeSystemJobs,
			},
			query,
			flds.Names(),
		)
		if err != nil {
			return nil, huma.Error500InternalServerError(
				jobsCommandName+": Failed to query jobs data", err)
		}
		jobs := make([]Jobs_Job, 0, len(records))
		for _, r := range records {
			jobs = append(jobs, respond(flds, r))
		}
		return &JobsResponse{Body: jobs}, nil
	})
}
