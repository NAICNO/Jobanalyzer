package restapi

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"sonalyze/data/common"
	"sonalyze/data/config"
)

// List all nodes in a cluster with the latest hardware and OS information.  Note that the time
// window here must be for sysinfo data, not for sample data.

const nodesInfoName = "/cluster/{cluster}/nodes/info"

type NodesInfoResponse struct {
	// Map: node -> data
	Body map[string]NodesInfoResponse_Node
}

type NodesInfoResponse_Node struct {
	Time           string                   `json:"time" doc:"ISO timestamp"`
	Cluster        string                   `json:"cluster"`
	Node           string                   `json:"node"`
	OsName         string                   `json:"os_name"`
	OsRelease      string                   `json:"os_release"`
	Memory         uint64                   `json:"memory" doc:"KiB"`
	Sockets        uint64                   `json:"sockets"`
	CoresPerSocket uint64                   `json:"cores_per_socket"`
	ThreadsPerCore uint64                   `json:"threads_per_core"`
	Cards          []NodesInfoResponse_Card `json:"cards"`
}

type NodesInfoResponse_Card struct {
	UUID    string `json:"uuid"`
	Model   string `json:"model"`
	MemSize uint64 `json:"mem_size" doc:"KiB"`
}

func addNodesInfo(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "get-nodes-info",
			Method:      http.MethodGet,
			Path:        nodesInfoName,
			Summary:     "Retrieve available information about nodes in a cluster",
		},
		handleNodesInfo,
	)
}

func handleNodesInfo(
	ctx context.Context,
	input *struct {
		Cluster  string `path:"cluster" example:"my.cluster.name" doc:"Name of cluster"`
		Nodename string `query:"nodename" doc:"Compressed node name list"`
		TimeInS  uint64 `query:"time_in_s" doc:"Posix timestamp"`
	},
) (*NodesInfoResponse, error) {
	// Logic from data/config
	meta, hErr := getClusterContext(nodesInfoName, input.Cluster)
	if hErr != nil {
		return nil, hErr
	}
	from, to, hErr := timeWindowFromData(nodesInfoName, meta, input.TimeInS, input.TimeInS)
	if hErr != nil {
		return nil, hErr
	}
	// Create a large enough window for sysinfo data.
	//
	// This is far from perfect, and at a minimum this should have been a parameter to
	// timeWindowFromData, we could have 0 mean "default".
	from = from.Add(-24 * time.Hour)
	cdb, err := config.OpenConfigDataProvider(meta)
	if err != nil {
		return nil, huma.Error500InternalServerError(
			nodesInfoName+": Failed to open config data store", err)
	}
	var hostList []string
	if input.Nodename != "" {
		hostList = []string{input.Nodename}
	}
	rs, err := cdb.RawQuery(config.QueryFilter{
		QueryFilter: common.QueryFilter{
			HaveFrom: true,
			FromDate: from,
			HaveTo:   true,
			ToDate:   to,
			Host:     hostList,
		},
		Verbose: verbose,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError(
			nodesInfoName+": Failed to query config data", err)
	}
	resp := &NodesInfoResponse{Body: make(map[string]NodesInfoResponse_Node, len(rs))}
	for _, r := range rs {
		cards := make([]NodesInfoResponse_Card, len(r.Cards))
		for i, c := range r.Cards {
			cards[i].UUID = c.UUID
			cards[i].Model = c.Model
			cards[i].MemSize = c.Memory
		}
		resp.Body[r.Node.Node] = NodesInfoResponse_Node{
			Time:           r.Node.Time,
			Cluster:        r.Node.Cluster,
			Node:           r.Node.Node,
			OsName:         r.Node.OsName,
			OsRelease:      r.Node.OsRelease,
			Memory:         r.Node.Memory,
			Sockets:        r.Node.Sockets,
			CoresPerSocket: r.Node.CoresPerSocket,
			ThreadsPerCore: r.Node.ThreadsPerCore,
			Cards:          cards,
		}
	}
	return resp, nil
}
