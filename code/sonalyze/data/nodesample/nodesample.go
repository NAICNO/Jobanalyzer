package nodesample

import (
	"fmt"

	uslices "go-utils/slices"
	"sonalyze/data/common"
	"sonalyze/db"
	"sonalyze/db/repr"
	"sonalyze/db/types"
)

type NodeSampleDataProvider struct {
	theLog db.NodeSampleDataProvider
}

func OpenNodeSampleDataProvider(meta types.Context) (*NodeSampleDataProvider, error) {
	theLog, err := db.OpenReadOnlyDB(meta, types.NodeSampleData)
	if err != nil {
		return nil, err
	}
	return &NodeSampleDataProvider{theLog}, nil
}

type QueryFilter = common.QueryFilter

func (nsp *NodeSampleDataProvider) Query(
	filter QueryFilter,
) ([]*repr.NodeSample, error) {
	f, err := filter.Instantiate()
	if err != nil {
		return nil, err
	}
	recordBlobs, _, err := nsp.theLog.ReadNodeSamples(
		types.DataProviderFilter{
			FromDate: filter.FromDate,
			ToDate:   filter.ToDate,
			Node:     f.HostFilter(),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to read log records: %v", err)
	}
	return common.ApplyFilter(f, uslices.Catenate(recordBlobs)), nil
}
