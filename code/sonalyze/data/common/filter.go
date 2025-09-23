package common

import (
	"math"
	"slices"
	"time"

	"go-utils/hostglob"
	. "sonalyze/common"
	"sonalyze/db/repr"
)

type QueryFilter struct {
	HaveFrom bool // FromDate was user input, not default; see below
	FromDate time.Time
	HaveTo   bool // ToDate was user input, not default; see below
	ToDate   time.Time
	Host     []string
}

func (filter *QueryFilter) Instantiate() (*CompiledFilter, error) {
	var hostFilter *Hosts
	if len(filter.Host) > 0 {
		var err error
		hostFilter, err = NewHosts(true, filter.Host)
		if err != nil {
			return nil, err
		}
	}

	// This is an idiom.  The from/to dates may be defaulted and are in any case used to find
	// records in a directory database by ingestion time.  If the from/to dates are given
	// explicitly, then those dates are also used to filter data within that window (since data
	// ingested in the window may be timestamped outside the window).  If the from/to dates are not
	// given explicitly, then we do not filter by date *except* we discard records whose timestamp
	// is invalid.  Also see application/local.go, which has a private copy of the same (ancient)
	// code.

	var scanFrom int64 = 0
	if filter.HaveFrom {
		scanFrom = filter.FromDate.Unix()
	}
	var scanTo int64 = math.MaxInt64
	if filter.HaveTo {
		scanTo = filter.ToDate.Unix()
	}
	var globber *hostglob.HostGlobber
	if hostFilter != nil {
		globber = hostFilter.HostnameGlobber()
	}
	return &CompiledFilter{
		hostFilter,
		scanFrom,
		scanTo,
		globber,
	}, nil
}

type CompiledFilter struct {
	hostFilter       *Hosts
	scanFrom, scanTo int64
	globber          *hostglob.HostGlobber
}

func (c *CompiledFilter) HostFilter() *Hosts {
	return c.hostFilter
}

func ApplyFilter[T repr.Filterable](filter *CompiledFilter, records []T) []T {
	return slices.DeleteFunc(records, func(s T) bool {
		timeVal, nodeStr := s.TimeAndNode()
		if filter.globber != nil && !filter.globber.IsEmpty() && !filter.globber.Match(nodeStr) {
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
