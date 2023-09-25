package util

import (
	"os"
	"path"
	"testing"
	"time"
)

func TestOptionsDataPath(t *testing.T) {
	opt := NewStandardOptions("hi")
	err := opt.Parse([]string{"--data-path", "ho/hum"})
	if err != nil {
		t.Fatalf("Failed data path #1: %v", err)
	}
	wd, _ := os.Getwd()
	if opt.DataPath != path.Join(wd, "ho/hum") {
		t.Fatalf("Failed data path #2")
	}

	opt = NewStandardOptions("hi")
	err = opt.Parse([]string{"--data-path", "/ho/hum"})
	if err != nil {
		t.Fatalf("Failed data path #1")
	}
	if opt.DataPath != "/ho/hum" {
		t.Fatalf("Failed data path #3")
	}
}

func TestOptionsDateRange(t *testing.T) {
	opt := NewStandardOptions("hi")
	err := opt.Parse([]string{"--data-path", "irrelevant", "--from", "3d", "--to", "2d"})
	if err != nil {
		t.Fatalf("Failed date range #1: %v", err)
	}
	a := time.Now().UTC().AddDate(0, 0, -3)
	b := time.Now().UTC().AddDate(0, 0, -1)
	if opt.From.Year() != a.Year() || opt.From.Month() != a.Month() || opt.From.Day() != a.Day() {
		t.Fatalf("Bad `from` date: %v", opt.From)
	}
	if opt.To.Year() != b.Year() || opt.To.Month() != b.Month() || opt.To.Day() != b.Day() {
		t.Fatalf("Bad `to` date: got %v, wanted %v", opt.To, b)
	}
}

func TestMatchWhen(t *testing.T) {
	tm, err := matchWhen("2023-09-12")
	if err != nil || tm.Year() != 2023 || tm.Month() != 9 || tm.Day() != 12 {
		t.Fatalf("Failed parsing day")
	}

	n3 := time.Now().UTC().AddDate(0, 0, -3)
	tm, err = matchWhen("3d")
	if err != nil || tm.Year() != n3.Year() || tm.Month() != n3.Month() || tm.Day() != n3.Day() {
		t.Fatalf("Failed parsing days-ago")
	}

	n14 := time.Now().UTC().AddDate(0, 0, -14)
	tm, err = matchWhen("2w")
	if err != nil || tm.Year() != n14.Year() || tm.Month() != n14.Month() || tm.Day() != n14.Day() {
		t.Fatalf("Failed parsing weeks-ago")
	}
}
