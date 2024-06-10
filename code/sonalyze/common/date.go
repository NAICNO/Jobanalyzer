package common

import (
	"errors"
	"regexp"
	"strconv"
	"time"
)

// Parse a relative date string and return the time in the UTC time zone.  The base time is folded
// to UTC if it is not already utc.
//
// The following date parser is explicitly compatible with the Rust version.  Specifically,
//
//  - relative time of the form Nd or Nw ignore the endOfDay flag, see below
//  - endOfDay takes us to the last second of the last hour of the day, even though this is slightly
//    wrong and requires the use of <= to compare times rather than the more obvious <.
//
// The date range returned is thus an inclusive range.
//
// The date format is one of:
//
//  YYYY-MM-DD
//  Nd (days ago)
//  Nw (weeks ago)
//
// NOTE: we're opting in to the Go semantics here: the nonexistent yyyy-09-31 is silently
// reinterpreted as yyyy-10-01.  This differs from the Rust semantics - Chrono signals an error.
//
// NOTE: endOfDay is not applied to Nd and Nw because the base time is "now" and there will be no
// records with a timestamp later than "now"; this structure comes from the Rust code.  I'm not sure
// that it wouldn't be better to be uniform.

func ParseRelativeDateUtc(now time.Time, s string, endOfDay bool) (time.Time, error) {
	now = now.UTC()
	if probe := daysRe.FindSubmatch([]byte(s)); probe != nil {
		days, _ := strconv.ParseUint(string(probe[1]), 10, 32)
		return now.AddDate(0, 0, -int(days)), nil
	}

	if probe := weeksRe.FindSubmatch([]byte(s)); probe != nil {
		weeks, _ := strconv.ParseUint(string(probe[1]), 10, 32)
		return now.AddDate(0, 0, -int(weeks)*7), nil
	}

	if probe := dateRe.FindSubmatch([]byte(s)); probe != nil {
		yyyy, _ := strconv.ParseUint(string(probe[1]), 10, 32)
		mm, _ := strconv.ParseUint(string(probe[2]), 10, 32)
		dd, _ := strconv.ParseUint(string(probe[3]), 10, 32)
		var h, m, s int
		if endOfDay {
			h, m, s = 23, 59, 59
		}
		return time.Date(int(yyyy), time.Month(mm), int(dd), h, m, s, 0, time.UTC), nil
	}

	return now, errors.New("Bad time specification")
}

// t should be UTC, the result is always UTC
func ThisDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// t should be UTC, the result is always UTC
func RoundupDay(t time.Time) time.Time {
	// Add less than one full day so as to make RoundupDay idempotent.
	return ThisDay(t.Add(24*time.Hour - 1*time.Second))
}

var dateRe = regexp.MustCompile(`^(\d\d\d\d)-(\d\d)-(\d\d)$`)
var daysRe = regexp.MustCompile(`^(\d+)d$`)
var weeksRe = regexp.MustCompile(`^(\d+)w$`)

func TruncateToHalfHour(t int64) int64 {
	u := time.Unix(t, 0).UTC()
	m := 0
	if u.Minute() >= 30 {
		m = 30
	}
	return time.Date(u.Year(), u.Month(), u.Day(), u.Hour(), m, 0, 0, time.UTC).Unix()
}

func TruncateToHour(t int64) int64 {
	u := time.Unix(t, 0).UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), u.Hour(), 0, 0, 0, time.UTC).Unix()
}

func TruncateToHalfDay(t int64) int64 {
	u := time.Unix(t, 0).UTC()
	h := 0
	if u.Hour() >= 12 {
		h = 12
	}
	return time.Date(u.Year(), u.Month(), u.Day(), h, 0, 0, 0, time.UTC).Unix()
}

func TruncateToDay(t int64) int64 {
	u := time.Unix(t, 0).UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC).Unix()
}

func TruncateToWeek(t int64) int64 {
	// Week starts on monday, while t.Weekday starts on Sunday
	u := time.Unix(t, 0).UTC()
	u = time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
	u = u.AddDate(0, 0, -daysFromMonday(u.Weekday()))
	return u.Unix()
}

func daysFromMonday(m time.Weekday) int {
	t := int(m) - int(time.Monday)
	if t < 0 {
		t += 7
	}
	return t
}

func AddHalfHour(t int64) int64 {
	return t + 30*60
}

func AddHour(t int64) int64 {
	return t + 60*60
}

func AddHalfDay(t int64) int64 {
	return t + 12*60*60
}

func AddDay(t int64) int64 {
	return t + 24*60*60
}

func AddWeek(t int64) int64 {
	return t + 7*24*60*60
}

func FormatYyyyMmDdHhMmUtc(t int64) string {
	return time.Unix(t, 0).UTC().Format("2006-01-02 15:04")
}
