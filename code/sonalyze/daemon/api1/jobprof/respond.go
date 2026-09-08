// Generated from jobprof.go by generate-response.  DO NOT EDIT.

package jobprof

import (
	"sonalyze/daemon/apiutil"
	"sonalyze/db/repr"
)

const responseDefaults = "Time,Node,Command,Pid,CpuPct,ResidentMemGB"

type Jobprof_Process struct {
	Time          IsoDateTimeValue `json:"Time,omitempty"`
	Node          string           `json:"Node,omitempty"`
	Command       string           `json:"Command,omitempty"`
	CpuUtilPct    int              `json:"CpuUtilPct,omitempty"`
	VirtualMemGB  int              `json:"VirtualMemGB,omitempty"`
	ResidentMemGB int              `json:"ResidentMemGB,omitempty"`
	Gpu           int              `json:"Gpu,omitempty"`
	GpuMemGB      int              `json:"GpuMemGB,omitempty"`
	NumProcs      IntOrEmpty       `json:"NumProcs,omitempty"`
}

func respond(flds *apiutil.FieldMap, r *ProfileStep) Jobprof_Process {
	var x Jobprof_Process
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
