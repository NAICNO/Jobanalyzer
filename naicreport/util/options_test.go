package util

import (
	"errors"
	"flag"
	"os"
	"path"
	"testing"
	"time"
)

func parse(
	fileOpt *DataFilesOptions,
	filterOpt *DateFilterOptions,
	flags *flag.FlagSet,
	args []string,
) error {
	err := flags.Parse(args)
	if err != nil {
		return err
	}
	err = RectifyDataFilesOptions(fileOpt, flags)
	err2 := RectifyDateFilterOptions(filterOpt, flags)
	return errors.Join(err, err2)
}

func newSonarLogOptions(progname string) (*DataFilesOptions, *DateFilterOptions, *flag.FlagSet) {
	flags := flag.NewFlagSet(progname, flag.ExitOnError)
	fileOpts := AddDataFilesOptions(flags, "data-zappa", "Data directory")
	filterOpts := AddDateFilterOptions(flags)
	return fileOpts, filterOpts, flags
}

func TestOptionsDataPath(t *testing.T) {
	fileOpt, filterOpt, flags := newSonarLogOptions("hi")
	err := parse(fileOpt, filterOpt, flags, []string{"--data-zappa", "ho/hum"})
	if err != nil {
		t.Fatalf("Failed data path #1: %v", err)
	}
	wd, _ := os.Getwd()
	if fileOpt.Path != path.Join(wd, "ho/hum") {
		t.Fatalf("Failed data path #2")
	}

	fileOpt, filterOpt, flags = newSonarLogOptions("hi")
	err = parse(fileOpt, filterOpt, flags, []string{"--data-zappa", "/ho/hum"})
	if err != nil {
		t.Fatalf("Failed data path #1")
	}
	if fileOpt.Path != "/ho/hum" {
		t.Fatalf("Failed data path #3")
	}
}

func TestOptionsDateRange(t *testing.T) {
	fileOpt, filterOpt, flags := newSonarLogOptions("hi")
	err := parse(fileOpt, filterOpt, flags, []string{"--data-zappa", "irrelevant", "--from", "3d", "--to", "2d"})
	if err != nil {
		t.Fatalf("Failed date range #1: %v", err)
	}
	a := time.Now().UTC().AddDate(0, 0, -3)
	b := time.Now().UTC().AddDate(0, 0, -1)
	if filterOpt.From.Year() != a.Year() || filterOpt.From.Month() != a.Month() || filterOpt.From.Day() != a.Day() {
		t.Fatalf("Bad `from` date: %v", filterOpt.From)
	}
	if filterOpt.To.Year() != b.Year() || filterOpt.To.Month() != b.Month() || filterOpt.To.Day() != b.Day() {
		t.Fatalf("Bad `to` date: got %v, wanted %v", filterOpt.To, b)
	}
}
