package command

import (
	"testing"
)

func TestDurationToSeconds(t *testing.T) {
	expect(t, "-d", "1", true, 0)
	expect(t, "-d", "1m", false, 60)
	expect(t, "-d", "1h", false, 60*60)
	expect(t, "-d", "1d", false, 60*60*24)
	expect(t, "-d", "1w", false, 60*60*24*7)
	expect(t, "-d", "1w1d1h1m", false, 60*60*24*7 + 60*60*24 + 60*60 + 60)
	expect(t, "-d", "1w1d1m1h", true, 0)
	expect(t, "-d", "", false, 0)
}

func expect(t *testing.T, option, s string, e bool, ans int64) {
	n, err := DurationToSeconds(option, s)
	if !e && err != nil {
		t.Fatal(err)
	}
	if e && err == nil {
		t.Fatal("Expected error, got none")
	}
	if !e && n != ans {
		t.Fatalf("Expected %d, got %d", ans, n)
	}
}

