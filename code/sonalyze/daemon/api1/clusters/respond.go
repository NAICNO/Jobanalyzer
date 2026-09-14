// Generated from clusters.go by generate-response.  DO NOT EDIT.

package clusters

import (
	"sonalyze/daemon/apiutil"
	"sonalyze/db/repr"
)

const responseDefaults = "Name,Description"

type Cluster struct {
	Name        string   `json:"Name,omitempty" doc:"Cluster name"`
	Description string   `json:"Description,omitempty" doc:"Human-consumable cluster summary"`
	Aliases     []string `json:"Aliases,omitempty" doc:"Aliases of cluster"`
}

func respond(flds *apiutil.FieldMap, r *repr.Cluster) Cluster {
	var x Cluster
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
