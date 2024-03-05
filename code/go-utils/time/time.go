package time

import (
	"errors"
	"regexp"
	"strconv"
	"time"
)

func MinTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func MaxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

// The format of `from` and `to` is one of:
//  YYYY-MM-DD
//  Nd (days ago)
//  Nw (weeks ago)

var dateRe = regexp.MustCompile(`^(\d\d\d\d)-(\d\d)-(\d\d)$`)
var daysRe = regexp.MustCompile(`^(\d+)d$`)
var weeksRe = regexp.MustCompile(`^(\d+)w$`)

func ParseRelativeDate(s string) (time.Time, error) {
	probe := dateRe.FindSubmatch([]byte(s))
	if probe != nil {
		yyyy, _ := strconv.ParseUint(string(probe[1]), 10, 32)
		mm, _ := strconv.ParseUint(string(probe[2]), 10, 32)
		dd, _ := strconv.ParseUint(string(probe[3]), 10, 32)
		return time.Date(int(yyyy), time.Month(mm), int(dd), 0, 0, 0, 0, time.UTC), nil
	}
	probe = daysRe.FindSubmatch([]byte(s))
	if probe != nil {
		days, _ := strconv.ParseUint(string(probe[1]), 10, 32)
		return ThisDay(time.Now().UTC().AddDate(0, 0, -int(days))), nil
	}
	probe = weeksRe.FindSubmatch([]byte(s))
	if probe != nil {
		weeks, _ := strconv.ParseUint(string(probe[1]), 10, 32)
		return ThisDay(time.Now().UTC().AddDate(0, 0, -int(weeks)*7)), nil
	}
	return time.Now(), errors.New("Bad time specification")
}

func ThisDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func NextDay(t time.Time) time.Time {
	return ThisDay(t.AddDate(0, 0, 1))
}
