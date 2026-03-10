package restapi

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type ErrorMessagesResponse struct {
	// Map: node -> data
	Body map[string]ErrorMessages_Message
}

type ErrorMessages_Message struct {
	Cluster string `json:"cluster"`
	Node    string `json:"node"`
	Time    string `json:"time"`
	Details string `json:"details"`
}

func addErrorMessages(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "get-error-messages",
			Method:      http.MethodGet,
			Path:        "/cluster/{cluster}/error-messages",
			Summary: `Retrieve Sonar errors for requested nodes or all nodes, logged
before the time_in_s parameter.`,
		},
		handleErrorMessages,
	)
}

func handleErrorMessages(
	ctx context.Context,
	input *struct {
		Cluster  string `path:"cluster" example:"my.cluster.name" doc:"Name of cluster"`
		Nodename string `query:"nodename" doc:"Compressed node name list"`
		TimeInS  uint64 `query:"time_in_s" doc:"Posix timestamp, default 'now'"`
	},
) (*ErrorMessagesResponse, error) {
	// TODO: Implement this, if anyone cares - we don't have APIs for it in sonalyze, the JSON
	// parsers just discard error messages at the moment.  To do better the parsers would have to
	// return an additional error message set.
	resp := &ErrorMessagesResponse{
		Body: make(map[string]ErrorMessages_Message),
	}
	return resp, nil
}
