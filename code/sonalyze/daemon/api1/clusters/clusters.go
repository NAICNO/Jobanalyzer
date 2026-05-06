package clusters

import (
	"context"

	"github.com/danielgtaylor/huma/v2"

	. "sonalyze/daemon/api1/common"
	"sonalyze/daemon/apiutil"
	"sonalyze/db/special"
)

//go:generate ../../../../generate-response/generate-response clusters.go

/*RESPONSE

package clusters

import (
	"sonalyze/daemon/apiutil"
	"sonalyze/db/repr"
)

%%

TYPE     Cluster_Cluster
TABLE    ../../../cmd/clusters/clusters.go
DEFAULTS Name,Description

ESNOPSER*/

const clusterCommandName = "/cluster"

type ClusterResponse struct {
	Body []Cluster_Cluster
}

func AddCluster(api huma.API) {
	huma.Get(api, clusterCommandName, func(
		ctx context.Context,
		input *struct {
			apiutil.AuthHeader
			Fields string `query:"fields" doc:"List of JSON field names"`
		},
	) (*ClusterResponse, error) {
		if hErr := apiutil.CheckAuth(clusterCommandName, GetAuthenticator, input.Auth); hErr != nil {
			return nil, hErr
		}
		flds := apiutil.Fields(input.Fields, responseDefaults)
		resp := &ClusterResponse{
			Body: make([]Cluster_Cluster, 0),
		}
		for _, c := range special.AllClusters() {
			resp.Body = append(resp.Body, respond(flds, &c.Cluster))
		}
		return resp, nil
	})
}
