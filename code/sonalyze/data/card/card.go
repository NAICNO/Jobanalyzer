// Query raw GPU card static configuration data. Also see "node".
package card

import (
	"fmt"

	uslices "go-utils/slices"
	"sonalyze/data/common"
	"sonalyze/db"
	"sonalyze/db/repr"
	"sonalyze/db/special"
)

type CardDataProvider struct {
	theLog db.SysinfoDataProvider
}

func OpenCardDataProvider(meta special.ClusterMeta) (*CardDataProvider, error) {
	theLog, err := db.OpenReadOnlyDB(meta, special.CardData)
	if err != nil {
		return nil, err
	}
	return &CardDataProvider{theLog}, nil
}

type QueryFilter = common.QueryFilter

func (cdp *CardDataProvider) Query(
	filter QueryFilter,
	verbose bool,
) ([]*repr.SysinfoCardData, error) {
	f, err := filter.Instantiate()
	if err != nil {
		return nil, err
	}
	recordBlobs, _, err := cdp.theLog.ReadSysinfoCardData(
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
