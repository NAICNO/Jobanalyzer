package api2

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	. "sonalyze/common"
	"sonalyze/daemon/apiutil"
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
	meta, hErr := apiutil.GetClusterContext(nodesInfoName, input.Cluster)
	if hErr != nil {
		return nil, hErr
	}
	from, to, hErr := apiutil.TimeWindowFromData(nodesInfoName, meta, 0, input.TimeInS)
	if hErr != nil {
		return nil, hErr
	}
	var host Hosts
	if input.Nodename != "" {
		host, hErr = apiutil.NewHostFilter(nodesInfoName, meta, input.Nodename, from, to)
		if hErr != nil {
			return nil, hErr
		}
	}
	sysinfo, hErr := getSysinfoAt(nodesInfoName, meta, to, host)
	if hErr != nil {
		return nil, hErr
	}
	cardInfo, hErr := getCardInfoByUUIDAt(nodesInfoName, meta, to, host)
	if hErr != nil {
		return nil, hErr
	}
	resp := &NodesInfoResponse{Body: make(map[string]NodesInfoResponse_Node, len(sysinfo))}
	for _, system := range sysinfo {
		cards := make([]NodesInfoResponse_Card, len(system.Cards))
		for i, c := range system.Cards {
			cards[i].UUID = c
			if ci := cardInfo[c]; ci != nil {
				cards[i].Model = ci.Model
				cards[i].MemSize = ci.Memory
			}
		}
		resp.Body[system.Node] = NodesInfoResponse_Node{
			Time:           system.Time,
			Cluster:        system.Cluster,
			Node:           system.Node,
			OsName:         system.OsName,
			OsRelease:      system.OsRelease,
			Memory:         system.Memory,
			Sockets:        system.Sockets,
			CoresPerSocket: system.CoresPerSocket,
			ThreadsPerCore: system.ThreadsPerCore,
			Cards:          cards,
		}
	}
	return resp, nil
}
