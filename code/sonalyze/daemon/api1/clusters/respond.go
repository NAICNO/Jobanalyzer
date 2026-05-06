// Generated from clusters.go by generate-response.  DO NOT EDIT.

package clusters

import (
	"sonalyze/daemon/apiutil"
	"sonalyze/db/repr"
)

const responseDefaults = "Name,Description"

type Cluster_Cluster struct {
	Name        string   `json:"Name,omitempty"`
	Description string   `json:"Description,omitempty"`
	Aliases     []string `json:"Aliases,omitempty"`
}

func respond(flds *apiutil.FieldMap, r *repr.Cluster) Cluster_Cluster {
	var x Cluster_Cluster
	if flds.Has("Name") {
		x.Name = r.Name
	}
	if flds.Has("Description") {
		x.Description = r.Description
	}
	if flds.Has("Aliases") {
		x.Aliases = r.Aliases
	}
	return x
}
