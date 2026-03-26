package api1

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"sonalyze/data/card"
	"sonalyze/db/repr"
)

const cardCommandName = "/cluster/{cluster}/card"

type CardResponse struct {
	// List of card data.  (Time,UUID) pairs are unique.
	Body []Card_Card
}

type Card_Card struct {
	*repr.SysinfoCardData
}

func addCard(api huma.API) {
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

func handleCard(ctx context.Context, input *StandardQueryFields) (*CardResponse, error) {
	meta, from, to, hosts, hErr := input.Parameters(cardCommandName)
	if hErr != nil {
		return nil, hErr
	}

	// TODO: authentication, how do we do this?  Do we?

	// FIXME: Do not duplicate this, but factor

	// Begin logic from cmd/cards/cards.go

	cdp, err := card.OpenCardDataProvider(meta)
	if err != nil {
		return nil, huma.Error500InternalServerError(
			cardCommandName+": Failed to open card store", err)
	}
	records, err :=
		cdp.Query(
			card.QueryFilter{HaveFrom: true, FromDate: from, HaveTo: true, ToDate: to, Host: hosts.Patterns()},
			verbose,
		)
	if err != nil {
		return nil, huma.Error500InternalServerError(
			cardCommandName+": Failed to query card data", err)
	}
	// FIXME: Apply compiled query, if any

	// End logic from cmd/cards/cards.go

	// Now construct the result set from the retrieved records by copying only the fields that are
	// requested.  There is a default set of fields.
	//
	// FIXME: Implement that

	var cards []Card_Card
	for _, r := range records {
		cards = append(cards, Card_Card{SysinfoCardData: r})
	}
	return &CardResponse{
		Body: cards,
	}, nil
}
