// FIXME: It is arguably a bug that the timeline is flattened here and not nested as it is for the
// job profile (similar things should look similar).  This needs to be fixed in the "gpus" query
// code, which actually has the nested timeline but flattens it in the query result, probably to
// simplify subsequent fixed-format printing.

package cardprof

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"sonalyze/cmd/gpus"
	"sonalyze/daemon/api1/common"
	dcommon "sonalyze/data/common"
)

//go:generate ../../../../generate-response/generate-response cardprof.go

/*RESPONSE

package cardprof

import (
	. "sonalyze/cmd/gpus"
	"sonalyze/daemon/apiutil"
	. "sonalyze/table"
)

%%

TYPE     CardProfile_Timestep
TABLE    ../../../cmd/gpus/gpus.go
DEFAULTS Timestamp,Hostname,Memory,Power

ESNOPSER*/

const cardProfileCommandName = "/card-profile/{cluster}"

type CardProfileResponse struct {
	Body []CardProfile_Timestep
}

func AddCardProfile(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "card-profile-command",
			Method:      http.MethodGet,
			Path:        cardProfileCommandName,
			Summary:     "Retrieve card profile",
		},
		handleCardProfile,
	)
}

func handleCardProfile(
	ctx context.Context,
	input *common.MinimalQueryFields,
) (*CardProfileResponse, error) {
	meta, from, to, nodes, _, flds, hErr := input.Parameters(cardProfileCommandName, responseDefaults)
	if hErr != nil {
		return nil, hErr
	}

	records, err := gpus.Query(
		meta,
		gpus.QueryFilter{
			dcommon.QueryFilter{
				HaveFrom: true,
				FromDate: from,
				HaveTo:   true,
				ToDate:   to,
				Host:     nodes,
			},
			-1,
		},
		nil,
	)
	if err != nil {
		return nil, huma.Error500InternalServerError(
			cardProfileCommandName+": Failed to query card profile data", err)
	}

	timesteps := make([]CardProfile_Timestep, 0, len(records))
	for _, r := range records {
		timesteps = append(timesteps, respond(flds, r))
	}
	return &CardProfileResponse{Body: timesteps}, nil
}
