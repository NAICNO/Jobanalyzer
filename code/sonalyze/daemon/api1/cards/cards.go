package cards

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"sonalyze/cmd/cards"
	"sonalyze/daemon/api1/common"
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

TYPE     Cards_Card
TABLE    ../../../cmd/cards/cards.go
DEFAULTS Time,Node,Manufacturer,Model,Memory

ESNOPSER*/

const cardsCommandName = "/cards/{cluster}"

type CardsResponse struct {
	// List of card data.  (Time,UUID) pairs are unique.
	Body []Cards_Card
}

func AddCards(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "cards-command",
			Method:      http.MethodGet,
			Path:        cardsCommandName,
			Summary:     "Retrieve card information",
		},
		handleCards,
	)
}

func handleCards(
	ctx context.Context,
	input *common.StandardQueryFields,
) (*CardsResponse, error) {
	meta, from, to, nodes, query, flds, hErr := input.Parameters(cardsCommandName, responseDefaults)
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
			Host:     nodes,
		},
		query,
	)
	if err != nil {
		return nil, huma.Error500InternalServerError(
			cardsCommandName+": Failed to query card data", err)
	}

	cards := make([]Cards_Card, 0, len(records))
	for _, r := range records {
		cards = append(cards, respond(flds, r))
	}
	return &CardsResponse{Body: cards}, nil
}
