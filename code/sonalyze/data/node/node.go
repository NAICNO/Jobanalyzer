// Query raw node static configuration data without any GPU cards: CPU, RAM, topology, etc.  Also
// see "card".
package node

import (
	"fmt"
	"time"

	uslices "go-utils/slices"
	. "sonalyze/common"
	"sonalyze/data/common"
	"sonalyze/db"
	"sonalyze/db/repr"
	"sonalyze/db/types"
)

type NodeDataProvider struct {
	theLog db.SysinfoDataProvider
}

func OpenNodeDataProvider(meta types.Context) (*NodeDataProvider, error) {
	theLog, err := db.OpenReadOnlyDB(meta, types.NodeData)
	if err != nil {
		return nil, err
	}
	return &NodeDataProvider{theLog}, nil
}

type QueryFilter = common.QueryFilter

func (ndp *NodeDataProvider) Query(
	filter QueryFilter,
	verbose bool,
) ([]*repr.SysinfoNodeData, error) {
	f, err := filter.Instantiate()
	if err != nil {
		return nil, err
	}
	recordBlobs, _, err := ndp.theLog.ReadSysinfoNodeData(
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

func (ndp *NodeDataProvider) QueryRaw(
	fromDate, toDate time.Time,
	hosts *Hosts,
	verbose bool,
) (recordBlobs [][]*repr.SysinfoNodeData, dropped int, err error) {
	return ndp.theLog.ReadSysinfoNodeData(fromDate, toDate, hosts, verbose)
}
