// Annotation types and formatters for them.

package table

import (
	"fmt"
	"strconv"
	"time"

	. "sonalyze/common"
)

// Timestamp types.  These types hold an int64 unix-timestamp-since-epoch or second count and
// require a particular kind of formatting.

type DateTimeValue int64        // yyyy-mm-dd hh:mm
type DateTimeValueOrBlank int64 // yyyy-mm-dd hh:mm or 16 blanks
type IsoDateTimeOrUnknown int64 // yyyy-mm-ddThh:mmZhh:mm
type DateValue int64            // yyyy-mm-dd
type TimeValue int64            // hh:mm
type DurationValue int64        // _d_h_m for d(ays) h(ours) m(inutes), rounded to minute, round up on ties

// Other types

type IntOrEmpty int // the int value, but "" if zero
type UstrMax30 Ustr // the string value but only max 30 first chars in fixed mode

// Stringers for simple cases.  There could be more but in most cases the formatting takes a
// formatting context and a stringer could at most pick one of them.

func (val DateValue) String() string {
	return time.Unix(int64(val), 0).UTC().Format("2006-01-02")
}

func (val TimeValue) String() string {
	return time.Unix(int64(val), 0).UTC().Format("15:04")
}

func (val IntOrEmpty) String() string {
	if val == 0 {
		return ""
	}
	return strconv.FormatInt(int64(val), 10)
}

// Formatters that take a context value

func FormatDurationValue(secs int64, ctx PrintMods) string {
	if (ctx & PrintModSec) != 0 {
		return fmt.Sprint(secs)
	}
	return FormatDurationDHM(secs, ctx)
}

func FormatDateTimeValue(timestamp int64, ctx PrintMods) string {
	// Note, it is part of the API that PrintModSec takes precedence over PrintModIso (this
	// simplifies various other paths).
	if (ctx & PrintModSec) != 0 {
		return fmt.Sprint(timestamp)
	}
	if (ctx & PrintModIso) != 0 {
		return FormatIsoUtc(timestamp)
	}
	return FormatYyyyMmDdHhMmUtc(timestamp)
}

// The DurationValue had two different formats in older code: %2dd%2dh%2dm and %dd%2dh%2dm.  The
// embedded spaces made things line up properly in fixed-format outputs of jobs, and most scripts
// would likely use "duration/sec" etc instead, but the variability in the leading space is weird
// and probably an accident (the "duration" field of the jobs report is the oldest one in the code
// and did not have leading space).  Additionally, in older code there was a difference in rounding
// behavior in different printers (some would round, some would truncate).
//
// The fact that the output format does not correspond to the "duration" option format (which does
// not allow spaces) is also annoying.
//
// Here, we settle on the following interpretation and hope it won't break anything.  For fixed
// output, always use %2dd%2dh%2dm.  For other outputs, always use %dd%dh%dm.  Also, always round to
// the nearerest minute, rounding up on ties.

func FormatDurationDHM(durationSec int64, ctx PrintMods) string {
	if durationSec%60 >= 30 {
		durationSec += 30
	}
	minutes := (durationSec / 60) % 60
	hours := (durationSec / (60 * 60)) % 24
	days := durationSec / (60 * 60 * 24)
	if (ctx & PrintModFixed) != 0 {
		return fmt.Sprintf("%2dd%2dh%2dm", days, hours, minutes)
	}
	return fmt.Sprintf("%dd%dh%dm", days, hours, minutes)
}

func FormatYyyyMmDdHhMmUtc(t int64) string {
	return time.Unix(t, 0).UTC().Format("2006-01-02 15:04")
}

func FormatIsoUtc(t int64) string {
	return time.Unix(t, 0).UTC().Format(time.RFC3339)
}
