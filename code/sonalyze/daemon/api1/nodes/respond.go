// Generated from nodes.go by generate-response.  DO NOT EDIT.

package nodes

import (
	"sonalyze/daemon/apiutil"
	"sonalyze/data/config"
)

const responseDefaults = "Hostname,CpuCores,MemGB,GpuCards,GpuMemGB,Description"

type Nodes_Node struct {
	Timestamp   string `json:"Timestamp,omitempty" doc:"Full ISO timestamp of when the reading was taken"`
	Hostname    string `json:"Hostname,omitempty" doc:"Name that host is known by on the cluster"`
	Description string `json:"Description,omitempty" doc:"End-user description, not parseable"`
	CpuCores    int    `json:"CpuCores,omitempty" doc:"Total number of cores x threads"`
	NumaNodes   int    `json:"NumaNodes,omitempty" doc:"NUMA nodes"`
	MemGB       int    `json:"MemGB,omitempty" doc:"GB of installed main RAM"`
	GpuCards    int    `json:"GpuCards,omitempty" doc:"Number of installed cards"`
	GpuMemGB    int    `json:"GpuMemGB,omitempty" doc:"Total GPU memory across all cards"`
	GpuMemPct   bool   `json:"GpuMemPct,omitempty" doc:"True if GPUs report accurate memory usage in percent"`
	Distances   string `json:"Distances,omitempty" doc:"NUMA distance matrix"`
	TopoSVG     string `json:"TopoSVG,omitempty" doc:"SVG encoding of node topology"`
	TopoText    string `json:"TopoText,omitempty" doc:"Text encoding of node topology"`
}

func respond(flds *apiutil.FieldMap, r *config.NodeConfig) Nodes_Node {
	var x Nodes_Node
	if flds.Has("Timestamp") {
		x.Timestamp = r.Timestamp
	}
	if flds.Has("Hostname") {
		x.Hostname = r.Hostname
	}
	if flds.Has("Description") {
		x.Description = r.Description
	}
	if flds.Has("CpuCores") {
		x.CpuCores = r.CpuCores
	}
	if flds.Has("NumaNodes") {
		x.NumaNodes = r.NumaNodes
	}
	if flds.Has("MemGB") {
		x.MemGB = r.MemGB
	}
	if flds.Has("GpuCards") {
		x.GpuCards = r.GpuCards
	}
	if flds.Has("GpuMemGB") {
		x.GpuMemGB = r.GpuMemGB
	}
	if flds.Has("GpuMemPct") {
		x.GpuMemPct = r.GpuMemPct
	}
	if flds.Has("Distances") {
		x.Distances = r.Distances
	}
	if flds.Has("TopoSVG") {
		x.TopoSVG = r.TopoSVG
	}
	if flds.Has("TopoText") {
		x.TopoText = r.TopoText
	}
	return x
}
