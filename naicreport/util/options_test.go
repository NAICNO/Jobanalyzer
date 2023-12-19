package util

import (
	"flag"
	"os"
	"path"
	"testing"
	"time"
)

func parse(s *SonarLogOptions, flags *flag.FlagSet, args []string) error {
	err := flags.Parse(args)
	if err != nil {
		return err
	}
	return RectifySonarLogOptions(s, flags)
}

func newSonarLogOptions(progname string) (*SonarLogOptions, *flag.FlagSet) {
	flags := flag.NewFlagSet(progname, flag.ExitOnError)
	logOpts := AddSonarLogOptions(flags)
	return logOpts, flags
}

func TestOptionsDataPath(t *testing.T) {
	opt, flags := newSonarLogOptions("hi")
	err := parse(opt, flags, []string{"--data-path", "ho/hum"})
	if err != nil {
		t.Fatalf("Failed data path #1: %v", err)
	}
	wd, _ := os.Getwd()
	if opt.DataPath != path.Join(wd, "ho/hum") {
		t.Fatalf("Failed data path #2")
	}

	opt, flags = newSonarLogOptions("hi")
	err = parse(opt, flags, []string{"--data-path", "/ho/hum"})
	if err != nil {
		t.Fatalf("Failed data path #1")
	}
	if opt.DataPath != "/ho/hum" {
		t.Fatalf("Failed data path #3")
	}
}

func TestOptionsDateRange(t *testing.T) {
	opt, flags := newSonarLogOptions("hi")
	err := parse(opt, flags, []string{"--data-path", "irrelevant", "--from", "3d", "--to", "2d"})
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
