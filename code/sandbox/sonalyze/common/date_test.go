package common

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	// Relative dates and weeks do not clear out hour/minute/second
	now := time.Now()
	x, err := ParseRelativeDate(now, "3d", false)
	if err != nil {
		t.Fatal(err)
	}
	if x.Location() != time.UTC {
		t.Fatal("Should return UTC time")
	}
	if !now.UTC().Equal(x.AddDate(0, 0, 3)) {
		t.Fatal("parse")
	}

	x, err = ParseRelativeDate(now, "2w", false)
	if err != nil {
		t.Fatal(err)
	}
	if !now.UTC().Equal(x.AddDate(0, 0, 14)) {
		t.Fatal("parse")
	}

	var q time.Time
	x, err = ParseRelativeDate(q, "2023-10-12", false)
	if err != nil {
		t.Fatal(err)
	}
	if x.Location() != time.UTC {
		t.Fatal("Should return UTC time")
	}
	if x.Hour() != 0 || x.Minute() != 0 || x.Second() != 0 {
		t.Fatal("hour/minute/second")
	}
	if x.Year() != 2023 || x.Month() != 10 || x.Day() != 12 {
		t.Fatal("year/month/day")
	}

	x, err = ParseRelativeDate(q, "2023-10-12", true)
	if err != nil {
		t.Fatal(err)
	}
	if x.Location() != time.UTC {
		t.Fatal("Should return UTC time")
	}
	if x.Hour() != 23 || x.Minute() != 59 || x.Second() != 59 {
		t.Fatal("hour/minute/second")
	}
	if x.Year() != 2023 || x.Month() != 10 || x.Day() != 12 {
		t.Fatal("year/month/day")
	}
}

// These are feeble but at least test that something works
func TestAdd(t *testing.T) {
	var q time.Time
	x, _ := ParseRelativeDate(q, "2023-10-12", false)
	z := time.Unix(AddHalfHour(x.Unix()), 0).UTC()
	if z.Hour() != 0 || z.Minute() != 30 {
		t.Fatal("Half-hour")
	}
	z = time.Unix(AddHour(x.Unix()), 0).UTC()
	if z.Hour() != 1 || z.Minute() != 0 {
		t.Fatal("Hour")
	}
	z = time.Unix(AddHalfDay(x.Unix()), 0).UTC()
	if z.Hour() != 12 || z.Minute() != 0 {
		t.Fatal("Half-day")
	}
	z = time.Unix(AddDay(x.Unix()), 0).UTC()
	if z.Hour() != 0 || z.Minute() != 0 || z.Day() != 13 {
		t.Fatal("Day")
	}
	z = time.Unix(AddWeek(x.Unix()), 0).UTC()
	if z.Hour() != 0 || z.Minute() != 0 || z.Day() != 19 {
		t.Fatal("Day")
	}
}

func TestTrunc(t *testing.T) {
	z := time.Unix(TruncateToHalfHour(time.Now().Unix()), 0).UTC()
	if z.Second() != 0 || z.Nanosecond() != 0 {
		t.Fatal("Fractions")
	}
	if z.Minute() != 0 && z.Minute() != 30 {
		t.Fatal("Half-hour")
	}

	z = time.Unix(TruncateToHour(time.Now().Unix()), 0).UTC()
	if z.Second() != 0 || z.Nanosecond() != 0 || z.Minute() != 0 {
		t.Fatal("Fractions")
	}

	z = time.Unix(TruncateToHalfDay(time.Now().Unix()), 0).UTC()
	if z.Second() != 0 || z.Nanosecond() != 0 || z.Minute() != 0 {
		t.Fatal("Fractions")
	}
	if z.Hour() != 0 && z.Hour() != 12 {
		t.Fatal("Half-day")
	}

	z = time.Unix(TruncateToDay(time.Now().Unix()), 0).UTC()
	if z.Second() != 0 || z.Nanosecond() != 0 || z.Minute() != 0 || z.Hour() != 0 {
		t.Fatal("Fractions")
	}

	if daysFromMonday(time.Monday) != 0 {
		t.Fatal("Monday")
	}

	if daysFromMonday(time.Sunday) != 6 {
		t.Fatalf("Sunday %d", daysFromMonday(time.Sunday))
	}

	assertTruncatesToMonday(t, "2023-09-03", "2023-08-28")
	assertTruncatesToMonday(t, "2023-09-04", "2023-09-04")
	assertTruncatesToMonday(t, "2023-09-05", "2023-09-04")
	assertTruncatesToMonday(t, "2023-09-06", "2023-09-04")
	assertTruncatesToMonday(t, "2023-09-07", "2023-09-04")
	assertTruncatesToMonday(t, "2023-09-08", "2023-09-04")
	assertTruncatesToMonday(t, "2023-09-09", "2023-09-04")
	assertTruncatesToMonday(t, "2023-09-10", "2023-09-04")
	assertTruncatesToMonday(t, "2023-09-11", "2023-09-11")
}

func assertTruncatesToMonday(t *testing.T, input, desired string) {
	tm, err := time.Parse(time.DateOnly, input)
	if err != nil {
		t.Fatal(err)
	}
	z := time.Unix(TruncateToWeek(tm.Unix()), 0).UTC()
	if z.Second() != 0 || z.Nanosecond() != 0 || z.Minute() != 0 || z.Hour() != 0 {
		t.Fatal("Fractions")
	}
	if z.Weekday() != time.Monday {
		t.Fatal("Weekday")
	}
	formatted := z.Format(time.DateOnly)
	if formatted != desired {
		t.Fatalf("Not the same! Desired=%s formatted=%s", desired, formatted)
	}
}
