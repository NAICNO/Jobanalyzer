// Node-downtime takes CSV output from `sonalyze uptime` and postprocesses it into a listing of
// nodes with their cumulative downtime.
//
// Run as:
//
//	go run node-downtime.go [options] input-file
//
// where the input-file is a csv with five fields: junk, hostname, junk, start, end (ie, the default
// output from `sonalyze uptime`).
//
// Run with -h to see the options.
//
// See cluster-downtime.go for information about how to generate the input-file, and some other
// considerations.
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
	HostOffs      = 1
	StartTimeOffs = 3
	EndTimeOffs   = 4
)

// CSV time format
const TimeFmt = "2006-01-02 15:04"

//// ===== Internal ======

var (
	topn = flag.Int("top", -1, "Print only the top `n` entries")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "node-downtime [options] inputfile\nOptions:\n")
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
	down := make(map[string]int64)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		start, err := time.Parse(TimeFmt, record[StartTimeOffs])
		if err != nil {
			log.Fatal(err)
		}
		end, err := time.Parse(TimeFmt, record[EndTimeOffs])
		if err != nil {
			log.Fatal(err)
		}
		down[record[HostOffs]] += (end.Unix() - start.Unix())
	}
	type entry struct {
		hostname string
		time     int64
	}
	var entries []entry
	for k, v := range down {
		entries = append(entries, entry{k, v})
	}
	slices.SortFunc(entries, func(a, b entry) int {
		if a.time == b.time {
			return cmp.Compare(a.hostname, b.hostname)
		}
		return -cmp.Compare(a.time, b.time)
	})
	for i, e := range entries {
		if *topn != -1 && i >= *topn {
			break
		}
		secs := e.time
		secs /= 60
		mins := secs % 60
		secs /= 60
		hours := secs
		fmt.Printf("%s %dh %dm\n", e.hostname, hours, mins)
	}
}
