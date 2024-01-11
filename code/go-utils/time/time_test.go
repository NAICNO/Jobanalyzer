package time

import (
	"testing"
	"time"
)

func TestParseRelativeDate(t *testing.T) {
	tm, err := ParseRelativeDate("2023-09-12")
	if err != nil || tm.Year() != 2023 || tm.Month() != 9 || tm.Day() != 12 {
		t.Fatalf("Failed parsing day")
	}

	n3 := time.Now().UTC().AddDate(0, 0, -3)
	tm, err = ParseRelativeDate("3d")
	if err != nil || tm.Year() != n3.Year() || tm.Month() != n3.Month() || tm.Day() != n3.Day() {
		t.Fatalf("Failed parsing days-ago")
	}

	n14 := time.Now().UTC().AddDate(0, 0, -14)
	tm, err = ParseRelativeDate("2w")
	if err != nil || tm.Year() != n14.Year() || tm.Month() != n14.Month() || tm.Day() != n14.Day() {
		t.Fatalf("Failed parsing weeks-ago")
	}
}
