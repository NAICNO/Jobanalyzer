// This is test code for perf testing.  It corresponds in some ways to `sonalyze parse`.
//
// Usage:
//   gsonarlog -data-dir dir [-from date] [-to date] [-cpuprofile filename]

package main

import (
	"flag"
	"fmt"
	"go-utils/sonarlog"
	gut "go-utils/time"
	"os"
	"path"
	"runtime/pprof"
	"time"
	"unsafe"
)

var (
	dataDir  string
	fromDate time.Time
	toDate   time.Time
	cpuProfile string
	verbose bool
)

func main() {
	flag.StringVar(&dataDir, "data-dir", "", "Root `directory` of data store")
	var fromDateStr, toDateStr string
	flag.StringVar(&fromDateStr, "from", "",
		"Earliest `date` to include, YYYY-MM-DD or Nd or Nw, default 1d")
	flag.StringVar(&toDateStr, "to", "",
		"Latest `date` to include, YYYY-MM-DD or Nd or Nw, default today")
	flag.StringVar(&cpuProfile, "cpuprofile", "", "write cpu profile to `filename`")
	flag.BoolVar(&verbose, "v", false, "debug info")
	flag.Parse()

	if dataDir == "" {
		panic("Required: -data-dir")
	}
	dataDir = path.Clean(dataDir)
	if fromDateStr != "" {
		var err error
		fromDate, err = gut.ParseRelativeDate(fromDateStr)
		if err != nil {
			panic(err)
		}
	} else {
		fromDate = time.Now().UTC().AddDate(0, 0, -1)
	}
	fromDate = gut.ThisDay(fromDate)

	if toDateStr != "" {
		var err error
		toDate, err = gut.ParseRelativeDate(toDateStr)
		if err != nil {
			panic(err)
		}
	} else {
		toDate = time.Now().UTC()
	}
	toDate = gut.NextDay(toDate)

	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if verbose {
		fmt.Printf("Size of Sample: %d\n", unsafe.Sizeof(sonarlog.Sample{}))
	}
	log, err := sonarlog.OpenDir(dataDir)
	if err != nil {
		panic(err)
	}
	readings, dropped, err := log.LogEntries(fromDate, toDate, nil, nil, verbose)
	if err != nil {
		panic(err)
	}
	if verbose {
		fmt.Printf("%d records, %d dropped\n", len(readings), dropped)
		sonarlog.UstrStats(false)
	}

	jobs(readings)
}

