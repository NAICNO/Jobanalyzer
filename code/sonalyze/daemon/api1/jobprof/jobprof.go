package jobprof

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"sonalyze/cmd/profile"
	"sonalyze/daemon/api1/common"
	dcommon "sonalyze/data/common"
	"sonalyze/data/sample"
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
// source.  This could change.
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
	CpuPct    int    `json:"CpuPct,omitempty" doc:"CPU utilization in percent, 100% = 1 core"`
	VirtMemGB int    `json:"VirtMemGB,omitempty" doc:"Main virtual memory usage in GiB"`
	ResMemGB  int    `json:"ResMemGB,omitempty" doc:"Main resident memory usage in GiB"`
	GpuPct    int    `json:"GpuPct,omitempty" doc:"GPU utilization in percent, 100% = 1 card"`
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
		// TODO: Not obvious that "query" from the std fields is sensible here.
		common.StandardQueryFields
		Job    uint `path:"jobid" example:"12345" doc:"Job ID"`
		Bucket uint `query:"bucket" example:"5" doc:"Number of adjacent samples to average"`
	},
) (*JobProfileResponse, error) {
	meta, from, to, nodes, _, flds, hErr := input.Parameters(jobprofCommandName, responseDefaults)
	if hErr != nil {
		return nil, hErr
	}
	qFilter := dcommon.QueryFilter{
		HaveFrom: true,
		FromDate: from,
		HaveTo:   true,
		ToDate:   to,
	}
	rFilter := sample.SampleFilter{
		IncludeHosts: nodes,
		IncludeJobs:  map[uint32]bool{uint32(input.Job): true},
		From:         from.UTC().Unix(),
		To:           to.UTC().Unix(),
	}
	maxMem := 0.0
	pd, err := profile.ComputeProfileData(meta, qFilter, nodes, &rFilter, uint32(input.Job), maxMem, input.Bucket)
	if err != nil {
		return nil, huma.Error400BadRequest(jobprofCommandName + ": " + err.Error())
	}
	jd := profile.ComputeJSONFromSamples(pd.M, pd.Processes, pd.Pif, false)
	timeline := make([]JobProfileTimestep, len(jd))
	for _, jt := range jd {
		points := make([]JobProfilePoint, len(jt.Points))
		for _, p := range jt.Points {
			var pp JobProfilePoint
			if flds.Has("Pid") {
				pp.Pid = uint64(p.Pid)
			}
			if flds.Has("Command") {
				pp.Command = p.Command
			}
			if flds.Has("Node") {
				pp.Node = p.Host
			}
			if flds.Has("CpuPct") {
				pp.CpuPct = p.CpuUtilPct
			}
			if flds.Has("VirtMemGB") {
				pp.VirtMemGB = int(p.CpuGB)
			}
			if flds.Has("ResMemGB") {
				pp.ResMemGB = int(p.RssAnonGB)
			}
			if flds.Has("GpuPct") {
				pp.GpuPct = p.GpuPct
			}
			if flds.Has("GpuMemGB") {
				pp.GpuMemGB = int(p.GpuMemGB)
			}
			if flds.Has("NumProcs") {
				pp.NumProcs = p.Nproc
			}
			points = append(points, pp)
		}
		timeline = append(timeline, JobProfileTimestep{
			Time: jt.Time,
			Data: points,
		})
	}
	return &JobProfileResponse{timeline}, nil
}
