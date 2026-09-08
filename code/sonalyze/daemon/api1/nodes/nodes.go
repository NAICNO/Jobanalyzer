package nodes

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"sonalyze/cmd/nodes"
	"sonalyze/daemon/api1/common"
	dcommon "sonalyze/data/common"
)

//go:generate ../../../../generate-response/generate-response nodes.go

/*RESPONSE

package nodes

import (
	"sonalyze/daemon/apiutil"
    "sonalyze/data/config"
)

%%

TYPE     Nodes_Node
TABLE    ../../../cmd/nodes/nodes.go
DEFAULTS Hostname,CpuCores,MemGB,GpuCards,GpuMemGB,Description

ESNOPSER*/

const nodesCommandName = "/nodes/{cluster}"

type NodesResponse struct {
	// List of node data.  (Time,UUID) pairs are unique.
	Body []Nodes_Node
}

func AddNodes(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "nodes-command",
			Method:      http.MethodGet,
			Path:        nodesCommandName,
			Summary:     "Retrieve node information",
		},
		handleNodes,
	)
}

func handleNodes(
	ctx context.Context,
	input *common.StandardQueryFields,
) (*NodesResponse, error) {
	meta, from, to, hosts, query, flds, hErr := input.Parameters(nodesCommandName, responseDefaults)
	if hErr != nil {
		return nil, hErr
	}
	records, err := nodes.Query(
		meta,
		nodes.QueryFilter{
			QueryFilter: dcommon.QueryFilter{
				HaveFrom: true,
				FromDate: from,
				HaveTo:   true,
				ToDate:   to,
				Host:     hosts,
			},
			Newest: false,
		},
		query,
	)
	if err != nil {
		return nil, huma.Error500InternalServerError(
			nodesCommandName+": Failed to query node data", err)
	}

	ns := make([]Nodes_Node, 0, len(records))
	for _, r := range records {
		ns = append(ns, respond(flds, r))
	}
	return &NodesResponse{Body: ns}, nil
}
