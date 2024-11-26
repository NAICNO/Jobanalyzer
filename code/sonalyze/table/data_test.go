package table

import (
	"fmt"
	"testing"
)

const (
	now = 1732518173          // 2024-11-25T07:02:53Z
	dur = 3600*33 + 60*6 + 38 // 1d 9h 7m, rounded up
)

func TestDataFormatting(t *testing.T) {
	if s := DateValue(now).String(); s != "2024-11-25" {
		t.Fatalf("DateValue %s", s)
	}
	if s := TimeValue(now).String(); s != "07:02" {
		t.Fatalf("TimeValue %s", s)
	}

	if s := IntOrEmpty(7).String(); s != "7" {
		t.Fatalf("IntOrEmpty %s", s)
	}
	if s := IntOrEmpty(0).String(); s != "" {
		t.Fatalf("IntOrEmpty %s", s)
	}

	if s := FormatDurationValue(dur, 0); s != "1d9h7m" {
		t.Fatalf("Duration %s", s)
	}
	if s := FormatDurationValue(dur, PrintModFixed); s != " 1d 9h 7m" {
		t.Fatalf("Duration %s", s)
	}
	if s := FormatDurationValue(dur, PrintModSec); s != fmt.Sprint(dur) {
		t.Fatalf("Duration %s", s)
	}

	if s := FormatDateTimeValue(now, 0); s != "2024-11-25 07:02" {
		t.Fatalf("DateTimeValue %s", s)
	}
	if s := FormatDateTimeValue(now, PrintModSec|PrintModIso); s != fmt.Sprint(now) {
		t.Fatalf("DateTimeValue %s", s)
	}
	if s := FormatDateTimeValue(now, PrintModIso); s != "2024-11-25T07:02:53Z" {
		t.Fatalf("DateTimeValue %s", s)
	}

	// For the other types, the formatters are all embedded in the reflection code.
}
