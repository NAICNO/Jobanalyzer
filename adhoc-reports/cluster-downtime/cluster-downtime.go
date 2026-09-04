// Cluster-downtime takes CSV output from `sonalyze uptime` and postprocesses it into a time-based histogram.
//
// Run as:
//
//	go run cluster-downtime.go csv.go flag.go [options] input-file
//
// where the input-file is a csv with five fields: "host", hostname, "down", start, end (ie, the default
// output fields from `sonalyze uptime`).
//
// Run with -h to see the options.
//
// Timestamps are rounded to 10 minutes by default.  Usually you want neither -hour or -day, and
// -day can be tricky to interpret in maintenance windows as nodes bounce down and up again.
//
// NOTE, if the histogram does not show 0 then some nodes are down; on large clusters, it may be
// that the histogram never shows 0.  When you have a floor like that you'll need to process the
// data further to compute, say, downtime per node.  See node-downtime.go for this.
//
// A typical sonalyze command line to generate the input-file might be this:
//
//	sonalyze uptime \
//	    -data-dir ../data/betzy.sigma2.no \
//	    -f 2025-01-01 -t 2025-09-01 \
//	    -host 'b[4101-4396]' \
//	    -interval 60 \
//	    -only-down \
//	    -fmt csv,default \
//	| grep '^host' \
//	> betzy-4xxx-2025-jan-aug.csv
//
// The grep is necessary because these nodes also have GPUs and there's no sonalyze switch to only show
// the nodes themselves in this case.  (We could have added that logic in this script instead.)
//
// Obviously the code in this script is easily adapted to other output fields, just hack the constants below.
package main

import (
	"cmp"
	"flag"
	"fmt"
	"log"
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
	hour  = flag.Bool("hour", false, "Round times to the closest hour")
	day   = flag.Bool("day", false, "Round times to the closest day")
	scale = flag.Int("scale", 10, "Scale histogram by `factor`")
)

func main() {
	rest := FlagParse("cluster-downtime", []string{"inputfile"})
	if *hour && *day {
		FlagFail("Can't have both -hour and -day")
	}
	var events []event
	CsvLines(rest[0], func(record []string) {
		// We round every time to the nearest 10 minute
		start, err := time.Parse(TimeFmt, record[StartTimeOffs])
		if err != nil {
			log.Fatal(err)
		}
		start = adjust(start, true)
		end, err := time.Parse(TimeFmt, record[EndTimeOffs])
		if err != nil {
			log.Fatal(err)
		}
		end = adjust(end, false)
		events = append(events, event{start.Unix(), Down}, event{end.Unix(), Up})
	})
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

func adjust(t time.Time, down bool) time.Time {
	// If rounded components overflow they will be normalized by the constructor
	if *day {
		d := t.Day()
		if !down && (t.Hour() > 0 || t.Minute() > 0) {
			d++
		}
		return time.Date(t.Year(), t.Month(), d, 0, 0, 0, 0, t.Location())
	}
	if *hour {
		h := t.Hour()
		if !down && t.Minute() > 0 {
			h++
		}
		return time.Date(t.Year(), t.Month(), t.Day(), h, 0, 0, 0, t.Location())
	}
	// 10 minutes
	m := t.Minute()
	if down {
		m = m / 10 * 10
	} else {
		m = ((m + 9) / 10) * 10
	}
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), m, 0, 0, t.Location())
}

func stars(n int) string {
	s := ""
	for n > 0 {
		s += "*"
		n--
	}
	return s
}
