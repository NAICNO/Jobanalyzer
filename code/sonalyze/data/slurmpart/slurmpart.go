package slurmpart

import (
	"fmt"

	uslices "go-utils/slices"
	"sonalyze/data/common"
	"sonalyze/db"
	"sonalyze/db/repr"
	"sonalyze/db/special"
)

type SlurmPartitionDataProvider struct {
	theLog db.CluzterDataProvider
}

func OpenSlurmPartitionDataProvider(meta special.ClusterMeta) (*SlurmPartitionDataProvider, error) {
	theLog, err := db.OpenReadOnlyDB(meta, special.SlurmPartitionData)
	if err != nil {
		return nil, err
	}
	return &SlurmPartitionDataProvider{theLog}, nil
}

type QueryFilter = common.QueryFilter

func (spd *SlurmPartitionDataProvider) Query(
	filter QueryFilter,
	verbose bool,
) ([]*repr.CluzterPartitions, error) {
	f, err := filter.Instantiate()
	if err != nil {
		return nil, err
	}
	recordBlobs, _, err := spd.theLog.ReadCluzterPartitionData(
		filter.FromDate,
		filter.ToDate,
		verbose,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to read log records: %v", err)
	}
	return common.ApplyFilter(f, uslices.Catenate(recordBlobs)), nil
}
