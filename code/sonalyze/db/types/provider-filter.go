package types

import (
	"time"

	. "sonalyze/common"
)

// This filter is *advisory* - it can be used to optimize the query, but in general record filtering
// *must* be applied afterwards.
//
// FromDate and ToDate are always valid and must always be supplied - there can be no open-ended
// queries.  (In truth, these values are not used when the data store is a file list, but they
// really ought to be provided anyway.)  Of course, the zero value is valid and takes us back to
// 1970-01-01T00:00:00Z.
//
// The other fields *must* have non-zero values *only if* they are valid for the query in question.
// This is a client responsibility.  The database layer may try to apply filters that are
// inappropriate if the values are for nonexistent fields, and the query may then fail.
//
// In general, the pinch point for adding filters is in the calls to the db-level Read functions
// defined in db.DataProvider, so it's not very complicated to get the filtering right.

type DataProviderFilter struct {
	FromDate time.Time       // Earliest date
	ToDate   time.Time       // Latest date, inclusive
	Nodes    *Hosts          // Node names originating data
	Jobs     map[uint32]bool // Job IDs (exact match)
	Users    map[Ustr]bool   // User names (exact match)
}
