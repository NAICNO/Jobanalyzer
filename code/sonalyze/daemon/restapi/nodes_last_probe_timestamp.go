// List all nodes in a cluster with last precise timestamp for sample data.  The question is whether
// we should list nodes for which we have no data.  For an HPC cluster that's maybe meaningful, but not
// for the primary use case for this API.
//
// TODO: The meaning of the time_in_s parameter for last-probe-timestamp is unclear.
//
// Plausibly it means "last records not after this time" but for all I know it could also be a lower
// bound?
package restapi

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"sonalyze/data/sample"
)

const nodesLastProbeTimestampName = "/cluster/{cluster}/nodes/last-probe-timestamp"

type NodesLastProbeTimestampResponse struct {
	// Map: node -> date
	Body map[string]string
}

func addNodesLastProbeTimestamp(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "get-nodes-last-probe-timestamp",
			Method:      http.MethodGet,
			Path:        nodesLastProbeTimestampName,
			Summary:     `Retrieve the last known timestamps of records added for nodes in the cluster.`,
		},
		handleNodesLastProbeTimestamp,
	)
}

func handleNodesLastProbeTimestamp(
	ctx context.Context,
	input *struct {
		Cluster string `path:"cluster" example:"my.cluster.name" doc:"Name of cluster"`
		TimeInS uint64 `query:"time_in_s" doc:"Posix timestamp"`
	},
) (*NodesLastProbeTimestampResponse, error) {
	// Logic from cmd/metadata
	meta, hErr := getClusterContext(nodesLastProbeTimestampName, input.Cluster)
	if hErr != nil {
		return nil, hErr
	}
	from, to, hErr := timeWindowFromData(
		nodesLastProbeTimestampName, meta, input.TimeInS, input.TimeInS)
	if hErr != nil {
		return nil, hErr
	}
	sdp, hErr := openSampleDataProvider(nodesLastProbeTimestampName, meta)
	if hErr != nil {
		return nil, hErr
	}
	_, bounds, _, _, err :=
		sdp.Query(
			from,
			to,
			nil, // hosts
			&sample.SampleFilter{From: from.Unix(), To: to.Unix()},
			true, // bounds
			verbose,
		)
	if err != nil {
		return nil, huma.Error500InternalServerError(
			nodesLastProbeTimestampName+": Failed to query sample data", err)
	}
	rsp := &NodesLastProbeTimestampResponse{Body: make(map[string]string)}
	for k, v := range bounds {
		rsp.Body[k.String()] = time.Unix(v.Latest, 0).UTC().Format(time.RFC3339)
	}
	return rsp, nil
}
