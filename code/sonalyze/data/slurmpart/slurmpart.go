package slurmpart

import (
	"fmt"

	uslices "go-utils/slices"
	"sonalyze/data/common"
	"sonalyze/db"
	"sonalyze/db/repr"
)

type QueryFilter = common.QueryFilter

func Query(
	theLog db.CluzterDataProvider,
	filter QueryFilter,
	verbose bool,
) ([]*repr.CluzterPartitions, error) {
	f, err := filter.Instantiate()
	if err != nil {
		return nil, err
	}
	recordBlobs, _, err := theLog.ReadCluzterPartitionData(
		filter.FromDate,
		filter.ToDate,
		verbose,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to read log records: %v", err)
	}
	return common.ApplyFilter(f, uslices.Catenate(recordBlobs)), nil
}
