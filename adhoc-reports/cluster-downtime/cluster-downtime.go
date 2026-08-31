// Cluster-downtime takes CSV output from `sonalyze uptime` and postprocesses it into a time-based histogram.
//
// Run as:
//
//   go run cluster-downtime.go [options] input-file
//
// where the input-file is a csv with five fields: junk, hostname, junk, start, end (ie, the default
// output from `sonalyze uptime`).
//
// Timestamps are rounded to 10 minutes by default.
//
// Run with -h to see the options.
//
// A typical sonalyze command line to generate the input-file might be this:
//
//   sonalyze uptime \
//       -data-dir ../data/betzy.sigma2.no \
//       -f 2025-01-01 -t 2025-09-01 \
//       -host 'b[4101-4396]' \
//       -interval 60 \
//       -only-down \
//       -fmt csv,default \
//   | grep '^host'
//
// Obviously the code in this script is easily adapted to other output fields, just hack the constants below.
package main

import (
	"cmp"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"time"
)

//// ===== CSV Format Customization ======

// CSV field offsets
const (
	HostOffs = 1
	StartTimeOffs = 3
	EndTimeOffs = 4
)

// CSV time format
const TimeFmt = "2006-01-02 15:04"

//// ===== Internal ======

type event struct {
	t int64
	d int
}

const (
	Down = 1
	Up   = -1
)

var (
	histo = flag.Bool("histo", false, "Print histogram scaled by 10")
	round = flag.Int("round", 10, "Round times to `minutes`")
	scale = flag.Int("scale", 10, "Scale histogram by `factor`")
)

func main() {
	flag.Usage = func () {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "cluster-downtime [options] inputfile\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	rest := flag.Args()
	if len(rest) != 1 {
		flag.Usage()
		os.Exit(2)
	}
	inf, err := os.Open(rest[0])
	if err != nil {
		log.Fatal(err)
	}
	r := csv.NewReader(inf)
	var events []event
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		// We round every time to the nearest 10 minute
		start, err := time.Parse(TimeFmt, record[StartTimeOffs])
		if err != nil {
			log.Fatal(err)
		}
		start = start.Round(10*time.Minute)
		end, err := time.Parse(TimeFmt, record[EndTimeOffs])
		if err != nil {
			log.Fatal(err)
		}
		end = end.Round(time.Duration(*round) * time.Minute)
		events = append(events, event{start.Unix(), Down}, event{end.Unix(), Up})
	}
	slices.SortFunc(events, func(a, b event) int {
		return cmp.Compare(a.t, b.t)
	})
	// We want the report to be an ascending timeline with <date> <down-count> where the first
	// <down-count> will be nonzero and the <down-count> drops to zero when none are down, and then
	// next rises again when a host is up.
	i := 0
	downCount := 0
	for i < len(events) {
		var ev = events[i]
		downCount += ev.d
		i++
		for i < len(events) && events[i].t == ev.t && events[i].d == ev.d {
			downCount += ev.d
			i++
		}
		if *histo {
			fmt.Println(time.Unix(ev.t, 0).Format(TimeFmt), "     ", stars((downCount+(*scale-1))/(*scale)), downCount)
		} else {
			fmt.Println(time.Unix(ev.t, 0).Format(TimeFmt), "     ", downCount)
		}
	}
}

func stars(n int) string {
	s := ""
	for n > 0 {
		s += "*"
		n--
	}
	return s
}
