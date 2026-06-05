package common

import (
	"math"
	"slices"
	"time"

	. "sonalyze/common"
	"sonalyze/db/repr"
)

// Common semantics:
//
// - FromDate and ToDate are always valid, defaulted if necessary, as appropriate for each command
// - HaveFrom and HaveTo are informational and are set iff the dates were provided explicitly
// - Hosts may be empty, and if so it always means "all hosts"
//
// Also see comment block below.

type QueryFilter struct {
	HaveFrom bool // FromDate was user input, not default; see below
	FromDate time.Time
	HaveTo   bool // ToDate was user input, not default; see below
	ToDate   time.Time
	Host     Multihost
}

func (filter *QueryFilter) Instantiate() (*CompiledFilter, error) {
	// This is an idiom.  The from/to dates may be defaulted and are in any case used to find
	// records in a directory database by ingestion time.  If the from/to dates are given
	// explicitly, then those dates are also used to filter data within that window (since data
	// ingested in the window may be timestamped outside the window).  If the from/to dates are not
	// given explicitly, then we do not filter by date *except* we discard records whose timestamp
	// is invalid.  Also see application/local.go, which has a private copy of the same (ancient)
	// code.  Also see common invariants above.

	var scanFrom int64 = 0
	if filter.HaveFrom {
		scanFrom = filter.FromDate.Unix()
	}
	var scanTo int64 = math.MaxInt64
	if filter.HaveTo {
		scanTo = filter.ToDate.Unix()
	}
	return &CompiledFilter{
		filter.Host,
		scanFrom,
		scanTo,
	}, nil
}

type CompiledFilter struct {
	hostFilter       Multihost
	scanFrom, scanTo int64
}

func (c *CompiledFilter) HostFilter() Multihost {
	return c.hostFilter
}

func ApplyFilter[T repr.Filterable](filter *CompiledFilter, records []T) []T {
	return slices.DeleteFunc(records, func(s T) bool {
		timeVal, nodeStr := s.TimeAndNode()
		if !filter.hostFilter.Match(nodeStr) {
			return true
		}
		var parsed time.Time
		switch v := timeVal.(type) {
		case string:
			var err error
			parsed, err = time.Parse(time.RFC3339, v)
			if err != nil {
				return true
			}
		case time.Time:
			parsed = v
		default:
			panic("Internal error: value from TimeAndNode is not string or time.Time")
		}
		t := parsed.Unix()
		if filter.scanFrom <= t && t <= filter.scanTo {
			return false
		}
		return true
	})
}
