package jobprof

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"sonalyze/daemon/api1/common"
)

//go:generate ../../../../generate-response/generate-response jobprof.go

/*RESPONSE

package jobprof

import (
	"sonalyze/daemon/apiutil"
	"sonalyze/db/repr"
)

%%

TYPE     Jobprof_Process
TABLE    jobprof.go
DEFAULTS Time,Node,Command,Pid,CpuPct,ResidentMemGB

ESNOPSER*/

// There's no suitable table in the cmd/profile code itself.
//
// Really, instead of ProfileStep what we're looking for is probably the jsonJob private
// struct in the printing code...
//
// Although that code generates a time line with an array holding one datum per process
// at the time point.  That may be good enough?

/*TABLE profile

package jobprof

%%

FIELDS *profile.ProfileStep

 Time      IsoDateTimeValue desc:"Time of the start of the profiling bucket"
 Node      Ustr             desc:"Host on which process ran"
 Command   Ustr             desc:"Name of executable starting the process"
 CpuPct    int              desc:"CPU utilization in percent, 100% = 1 core (except for HTML)"
 VirtMemGB int              desc:"Main virtual memory usage in GiB"
 ResMemGB  int              desc:"Main resident memory usage in GiB"
 GpuPct    int              desc:"GPU utilization in percent, 100% = 1 card (except for HTML)"
 GpuMemGB  int              desc:"GPU resident memory usage in GiB (across all cards)"
 NumProcs  int              desc:"Number of rolled-up processes"

ELBAT*/

const jobprofCommandName = "/job-profile/{cluster}/{jobid}"

type JobProfileResponse struct {
	Body []Jobprof_Process
}

func AddJobProfile(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "jobprof-command",
			Method:      http.MethodGet,
			Path:        jobprofCommandName,
			Summary:     "Retrieve job profile information",
		},
		handleJobProfile,
	)
}

func handleJobProfile(
	ctx context.Context,
	input *struct {
		// TODO: Not obvious that "node" and "query" from the std fields are sensible here?
		common.StandardQueryFields
		Job int `path:"jobid" example:"12345" doc:"Job ID"`
	},
) (*JobProfileResponse, error) {
	panic("NYI")
}
