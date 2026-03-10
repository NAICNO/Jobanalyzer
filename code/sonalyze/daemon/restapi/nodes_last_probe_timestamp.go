package restapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"sonalyze/data/sample"
	"sonalyze/db"
	"sonalyze/db/special"
)

// List all nodes in a cluster with last timestamp for sample data.  The question is whether we
// should list nodes for which we have no data.  For an HPC cluster that's meaningful, but not for
// the primary use case for this API.

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
			Path:        "/cluster/{cluster}/nodes/last-probe-timestamp",
			Summary:     "Retrieve the last known timestamps of records added for nodes in the cluster",
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
	cluster := special.LookupCluster(input.Cluster)
	if cluster == nil {
		return nil, fmt.Errorf("Failed to find cluster %s", input.Cluster)
	}
	meta := db.NewContextFromCluster(cluster)
	sdp, err := sample.OpenSampleDataProvider(meta)
	if err != nil {
		return nil, err
	}
	from, to, err := timeWindowFromData(meta, input.TimeInS, input.TimeInS)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("Failed to read log records: %v", err)
	}
	rsp := &NodesLastProbeTimestampResponse{Body: make(map[string]string)}
	for k, v := range bounds {
		rsp.Body[k.String()] = time.Unix(v.Latest, 0).UTC().Format(time.RFC3339)
	}
	return rsp, nil
}
