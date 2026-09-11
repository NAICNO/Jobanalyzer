package jobprof

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	_ "sonalyze/cmd/profile"
	"sonalyze/daemon/api1/common"
	_ "sonalyze/daemon/apiutil"
)

// The response is a timeline: an array of objects where each object is a point in time, sorted
// ascending, carrying an inner array of per-process information at that point in time.  The fields
// of each profile point represent the various quantities (cpu, gpu, memory, etc) and are omitted
// unless requested.  Processes in the job not running at that time are omitted entirely.
//
// As this is a 2d sparse grid of data points, it could have been represented differently, notably
// with Pid as the primary index (list of rows rather than list of columns).
//
// I elected not to generate the response structure from any table since no table existed for the
// source.  This may change.
//
// The most important redundancy here is the mapping from (Node,Pid) to Command name.  The Command
// will almost never change for that pair (it can change if the job runs long enough for the pid to
// be recycled on the system, and the pid is assigned to two different commands in the same job, at
// different times), and some command names are very long: this inflates the data size, possibly
// significantly.  This could be fixed by including a map at the high level of the body and then
// using the map key for the command name in each data point, ie, by interning command names.  But
// if data are compressed in transit then that'll just happen by itself anyway, so is it worth it?

type JobProfileTimestep struct {
	Time string            `json:"Time,omitempty" doc:"The time at this time step (ISO)"`
	Data []JobProfilePoint `json:"Data,omitempty" doc:"Per-process data at this time"`
}

type JobProfilePoint struct {
	Pid       uint64 `json:"Pid,omitempty" doc:"Process ID for process"`
	Command   string `json:"Command,omitempty" doc:"Command name for process"`
	Node      string `json:"Node,omitempty" doc:"Name of node for process"`
	CpuPct    int    `json:"CpuPct,omitempty" doc:"CPU utilization in percent, 100% = 1 core (except for HTML)"`
	VirtMemGB int    `json:"VirtMemGB,omitempty" doc:"Main virtual memory usage in GiB"`
	ResMemGB  int    `json:"ResMemGB,omitempty" doc:"Main resident memory usage in GiB"`
	GpuPct    int    `json:"GpuPct,omitempty" doc:"GPU utilization in percent, 100% = 1 card (except for HTML)"`
	GpuMemGB  int    `json:"GpuMemGB,omitempty" doc:"GPU resident memory usage in GiB (across all cards)"`
	NumProcs  int    `json:"NumProcs,omitempty" doc:"Number of rolled-up processes"`
}

const responseDefaults = "Node,Command,Pid,CpuPct,ResMemGB"

const jobprofCommandName = "/job-profile/{cluster}/{jobid}"

type JobProfileResponse struct {
	Body []JobProfileTimestep
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

// This is wrong but hints at the solution

/*
func respond(flds *apiutil.FieldMap, r *profile.JsonPoint) Jobprof_Point {
	var x Jobprof_Point
	if flds.Has("Time") {
		x.Time = r.Time
	}
	if flds.Has("Node") {
		x.Node = JSONFromUstr(r.Node)
	}
	if flds.Has("Command") {
		x.Command = JSONFromUstr(r.Command)
	}
	if flds.Has("CpuUtilPct") {
		x.CpuUtilPct = r.CpuUtilPct
	}
	if flds.Has("VirtualMemGB") {
		x.VirtualMemGB = r.VirtualMemGB
	}
	if flds.Has("ResidentMemGB") {
		x.ResidentMemGB = r.ResidentMemGB
	}
	if flds.Has("Gpu") {
		x.Gpu = r.Gpu
	}
	if flds.Has("GpuMemGB") {
		x.GpuMemGB = r.GpuMemGB
	}
	if flds.Has("NumProcs") {
		x.NumProcs = r.NumProcs
	}
	return x
}
*/
