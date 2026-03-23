package types

import (
	"time"

	. "sonalyze/common"
)

// This filter is *advisory* - it can be used to optimize the query, but in general record filtering
// will have to be applied afterwards.
type DataProviderFilter struct {
	FromDate time.Time
	ToDate   time.Time
	Nodes    *Hosts
}
