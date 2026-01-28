package slurmnode

import (
	"fmt"

	uslices "go-utils/slices"
	"sonalyze/data/common"
	"sonalyze/db"
	"sonalyze/db/repr"
	"sonalyze/db/types"
)

type SlurmNodeDataProvider struct {
	theLog db.CluzterDataProvider
}

func OpenSlurmNodeDataProvider(meta types.Context) (*SlurmNodeDataProvider, error) {
	theLog, err := db.OpenReadOnlyDB(meta, types.SlurmNodeData)
	if err != nil {
		return nil, err
	}
	return &SlurmNodeDataProvider{theLog}, nil
}

// TODO: Various fields here, TBD
//
// Probably at least:
//  - select subset of hosts
//  - select subset of states
//  - select latest record for each host
//
// Note some of these may be applied by the query operator later, to the formatted records.
// Not sure yet how that plays out.

type QueryFilter = common.QueryFilter

func (sdp *SlurmNodeDataProvider) Query(
	filter QueryFilter,
	verbose bool,
) ([]*repr.CluzterNodes, error) {
	f, err := filter.Instantiate()
	if err != nil {
		return nil, err
	}
	recordBlobs, _, err := sdp.theLog.ReadCluzterNodeData(
		filter.FromDate,
		filter.ToDate,
		verbose,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to read log records: %v", err)
	}
	return common.ApplyFilter(f, uslices.Catenate(recordBlobs)), nil
}
