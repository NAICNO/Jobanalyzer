// The v1 API follows the old v0 API but GET requests returns proper JSON objects instead of
// strings, and the JSON objects can have non-string values.
//
// The v1 API also has new insertion points for the new data, represented as JSON.  The result of a
// POST is a JSON object with some data about the data that were received.

package api1

import (
	"go-utils/auth"

	"github.com/danielgtaylor/huma/v2"

	"sonalyze/daemon/api1/cards"
	"sonalyze/daemon/api1/clusters"
	"sonalyze/daemon/api1/common"
	"sonalyze/daemon/api1/insert"
)

func SetupAPI(
	api huma.API,
	insertAPI bool,
	getAuthenticator_ *auth.Authenticator,
	postAuthenticator_ *auth.Authenticator,
) {
	common.GetAuthenticator = getAuthenticator_
	common.PostAuthenticator = postAuthenticator_
	grp := huma.NewGroup(api, "/api/v1")

	cards.AddCard(grp)
	clusters.AddCluster(grp)

	if insertAPI {
		insert.AddInsertSysinfoData(grp)
		insert.AddInsertSampleData(grp)
		insert.AddInsertJobData(grp)
		insert.AddInsertClusterData(grp)
	}
}
