package node

import (
	"fmt"

	uslices "go-utils/slices"
	"sonalyze/data/common"
	"sonalyze/db"
	"sonalyze/db/repr"
)

type QueryFilter = common.QueryFilter

func Query(
	theLog db.SysinfoDataProvider,
	filter QueryFilter,
	verbose bool,
) ([]*repr.SysinfoNodeData, error) {
	f, err := filter.Instantiate()
	if err != nil {
		return nil, err
	}
	recordBlobs, _, err := theLog.ReadSysinfoNodeData(
		filter.FromDate,
		filter.ToDate,
		f.HostFilter(),
		verbose,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to read log records: %v", err)
	}
	return common.ApplyFilter(f, uslices.Catenate(recordBlobs)), nil
}
