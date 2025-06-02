package common

// Earliest and latest time stamps found in a set of records.

type Timebound struct {
	Earliest int64
	Latest   int64
}

// Map from host name to bounds for the host name

type Timebounds map[Ustr]Timebound
