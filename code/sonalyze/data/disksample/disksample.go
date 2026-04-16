package disksample

import (
	"fmt"

	uslices "go-utils/slices"
	"sonalyze/data/common"
	"sonalyze/db"
	"sonalyze/db/repr"
	"sonalyze/db/types"
)

type DiskSampleDataProvider struct {
	theLog db.DiskSampleDataProvider
}

func OpenDiskSampleDataProvider(meta types.Context) (*DiskSampleDataProvider, error) {
	theLog, err := db.OpenReadOnlyDB(meta, types.DiskSampleData)
	if err != nil {
		return nil, err
	}
	return &DiskSampleDataProvider{theLog}, nil
}

type QueryFilter = common.QueryFilter

func (nsp *DiskSampleDataProvider) Query(
	filter QueryFilter,
) ([]*repr.DiskSample, error) {
	f, err := filter.Instantiate()
	if err != nil {
		return nil, err
	}
	recordBlobs, _, err := nsp.theLog.ReadDiskSamples(
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
