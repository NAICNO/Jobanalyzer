package api1

import (
	"time"

	"github.com/danielgtaylor/huma/v2"

	. "sonalyze/common"
	"sonalyze/daemon/apiutil"
	"sonalyze/db/types"
)

var verbose bool

func SetupAPI(api huma.API, verbose_ bool) {
	verbose = verbose_
	grp := huma.NewGroup(api, "/api/v1")
	addCard(grp)
}

type StandardQueryFields struct {
	Cluster    string `path:"cluster" example:"my.cluster.name" doc:"Name of cluster"`
	StartTimeS uint64 `query:"start_time_s" doc:"Posix timestamp"`
	EndTimeS   uint64 `query:"end_time_s" doc:"Posix timestamp"`
	StartDate  string `query:"start_date" doc:"Date yyyy-mm-dd, overrides start_time_s"`
	EndDate    string `query:"end_date" doc:"Date yyyy-mm-dd, overrides end_time_s"`
	Host       string `query:"host" doc:"List of compressed host names"`
	Fields     string `query:"fields" doc:"List of field names"`
	Query      string `query:"query" doc:"Query term"`
}

func (input *StandardQueryFields) Parameters(opName string) (
	meta types.Context, from time.Time, to time.Time, hosts *Hosts, hErr huma.StatusError,
) {
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
	hosts, hErr = apiutil.NewHostFilter(opName, input.Host)
	if hErr != nil {
		return
	}
	// TODO: Can take canonical field list as input (or type?) and vet them
	// TODO: Can parse/compile the query
	return
}
