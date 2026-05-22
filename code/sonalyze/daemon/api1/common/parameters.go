package common

import (
	"time"

	"go-utils/auth"

	"github.com/danielgtaylor/huma/v2"

	. "sonalyze/common"
	"sonalyze/daemon/apiutil"
	"sonalyze/db/types"
	. "sonalyze/table"
)

var (
	GetAuthenticator  *auth.Authenticator
	PostAuthenticator *auth.Authenticator
)

// Queries

type StandardQueryFields struct {
	Cluster    string `path:"cluster" example:"my.cluster.name" doc:"Name of cluster"`
	StartTimeS uint64 `query:"start_time_s" doc:"Posix timestamp"`
	EndTimeS   uint64 `query:"end_time_s" doc:"Posix timestamp"`
	StartDate  string `query:"start_date" doc:"Date yyyy-mm-dd, overrides start_time_s"`
	EndDate    string `query:"end_date" doc:"Date yyyy-mm-dd, overrides end_time_s"`
	Node       string `query:"node" doc:"List of compressed node names"`
	Fields     string `query:"fields" doc:"List of JSON field names"`
	Query      string `query:"query" doc:"Query term"`
	apiutil.AuthHeader
}

func (input *StandardQueryFields) Parameters(opName, defaultFields string) (
	meta types.Context,
	from time.Time, to time.Time,
	nodes *Hosts,
	query PNode,
	fields *apiutil.FieldMap,
	hErr huma.StatusError,
) {
	if hErr = apiutil.CheckAuth(opName, GetAuthenticator, input.Auth); hErr != nil {
		return
	}
	meta, hErr = apiutil.GetClusterContext(opName, input.Cluster)
	if hErr != nil {
		return
	}
	if input.StartDate != "" {
		probe, err := time.Parse(time.DateOnly, input.StartDate)
		if err == nil {
			input.StartTimeS = uint64(probe.Unix())
		}
	}
	if input.EndDate != "" {
		// TODO: should be careful here: may need to interpret this as end-of-day or
		// start-of-next-day for some queries to work.  This ties into the HaveFrom/HaveTo logic.
		probe, err := time.Parse(time.DateOnly, input.EndDate)
		if err == nil {
			input.EndTimeS = uint64(probe.Unix())
		}
	}
	from, to, hErr = apiutil.TimeWindowFromData(opName, meta, input.StartTimeS, input.EndTimeS)
	if hErr != nil {
		return
	}
	nodes, hErr = apiutil.NewHostFilter(opName, input.Node)
	if hErr != nil {
		return
	}
	fields = apiutil.Fields(input.Fields, defaultFields)
	// TODO: Can in principle take canonical field list as input and vet the field list against
	// that, though the benefits are probably only slight.
	//
	// TODO: Must parse/compile the query.
	return
}
