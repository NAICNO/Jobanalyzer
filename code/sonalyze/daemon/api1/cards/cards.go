package cards

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"sonalyze/cmd/cards"
	"sonalyze/daemon/api1/common"
	"sonalyze/daemon/apiutil"
	"sonalyze/data/card"
)

//go:generate ../../../../generate-response/generate-response cards.go

/*RESPONSE

package cards

import (
	"sonalyze/daemon/apiutil"
	"sonalyze/db/repr"
)

%%

TYPE     Card_Card
TABLE    ../../../cmd/cards/cards.go
DEFAULTS Time,Node,Manufacturer,Model,Memory

ESNOPSER*/

const cardCommandName = "/cards/{cluster}"

type CardResponse struct {
	// List of card data.  (Time,UUID) pairs are unique.
	Body []Card_Card
}

func AddCard(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "card-command",
			Method:      http.MethodGet,
			Path:        cardCommandName,
			Summary:     "Retrieve card information",
		},
		handleCard,
	)
}

func handleCard(
	ctx context.Context,
	input *common.StandardQueryFields,
) (*CardResponse, error) {
	meta, from, to, nodes, query, hErr := input.Parameters(cardCommandName)
	if hErr != nil {
		return nil, hErr
	}

	records, err := cards.Query(
		meta,
		card.QueryFilter{
			HaveFrom: true,
			FromDate: from,
			HaveTo:   true,
			ToDate:   to,
			Host:     nodes.Patterns(),
		},
		query,
	)
	if err != nil {
		return nil, huma.Error500InternalServerError(
			cardCommandName+": Failed to query card data", err)
	}

	flds := apiutil.Fields(input.Fields, responseDefaults)
	cards := make([]Card_Card, 0, len(records))
	for _, r := range records {
		cards = append(cards, respond(flds, r))
	}
	return &CardResponse{Body: cards}, nil
}
