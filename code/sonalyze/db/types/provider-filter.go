package types

import (
	"time"

	. "sonalyze/common"
)

// The DataProviderFilter is used to optimize initial data lookup.  In the Sonalyze storage model,
// initial data lookup return a slice of slices of data, where each inner slice normally comes from
// a single host on a single date.  For this, we need only From/To dates and Node.  However, as a
// concession to the Timescaledb storage engine, other fields (jobs, users) can be added for the
// initial filtering to reduce scan and memory pressure.
//
// The filter is only ever *advisory* - it can be used to optimize the query, but full record
// filtering *must* be applied afterwards.  The storage engine can choose to ignore any or all of
// the fields of the filter.  For example, the file list storage engine ignores all; the directory
// tree engine ignores users and jobs.
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
// Usually, the pinch point for adding filter values is in the calls to the db-level Read functions
// defined in db.DataProvider, so it's not very complicated to get the filtering right.

type DataProviderFilter struct {
	FromDate time.Time       // Earliest date
	ToDate   time.Time       // Latest date, inclusive
	Node     *Hosts          // Names matching a single node in the data
	Jobs     map[uint32]bool // Job IDs (exact match)
	Users    map[Ustr]bool   // User names (exact match)
	// Nodes, when we match a set against another set
	// Account
	// Partition
	// State
	// Completed and Running flags could be translated to State?
}
