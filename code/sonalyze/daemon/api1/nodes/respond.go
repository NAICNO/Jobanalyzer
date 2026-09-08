// Generated from nodes.go by generate-response.  DO NOT EDIT.

package nodes

import (
	"sonalyze/daemon/apiutil"
	"sonalyze/data/config"
)

const responseDefaults = "Hostname,CpuCores,MemGB,GpuCards,GpuMemGB,Description"

type Nodes_Node struct {
	Timestamp   string `json:"Timestamp,omitempty"`
	Hostname    string `json:"Hostname,omitempty"`
	Description string `json:"Description,omitempty"`
	CpuCores    int    `json:"CpuCores,omitempty"`
	NumaNodes   int    `json:"NumaNodes,omitempty"`
	MemGB       int    `json:"MemGB,omitempty"`
	GpuCards    int    `json:"GpuCards,omitempty"`
	GpuMemGB    int    `json:"GpuMemGB,omitempty"`
	GpuMemPct   bool   `json:"GpuMemPct,omitempty"`
	Distances   string `json:"Distances,omitempty"`
	TopoSVG     string `json:"TopoSVG,omitempty"`
	TopoText    string `json:"TopoText,omitempty"`
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
