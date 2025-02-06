// This program assumes jobanalyzer profile csv-format input on stdin and will try to print a
// sensible stacked profile.
//
// Normally you'd run eg
//
//   sonalyze profile -cluster fox -f 4d -j 1307357 -fmt csv,gpu | go run stack.go
//
// The output will be a sequence of lines in increasing timestamp order:
//
//   timestamp xxxxyyyyzzzz
//   timestamp xxyyyyyyyyyzzzz
//   timestamp xz
//
// where the xyz are characters for a particular process.  The mapping from profile to char is
// printed at the start of the profile.
//
// By default, profile values are scaled by 10 b/c this is "roughly right" for many use cases, but
// sometimes this will not work out well; use -scale to adjust.

package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
)

const candidates = "-+*:=&^%$#@!abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
var (
	scale = flag.Float64("scale", 10.0, "Value scale factor")
)

func main() {
	flag.Parse()
	r := csv.NewReader(os.Stdin)
	input, err := r.ReadAll()
	if err != nil {
		panic(err)
	}
	hdr := input[0]
	chars := make([]string, len(hdr)-1)
	s := ""
	for i, h := range hdr[1:] {
		chars[i] = string(candidates[i])
		fmt.Printf(" %s   %s\n", chars[i], h)
		for range int(100 / *scale) {
			s += "_"
		}
	}
	// This line indicates "100% utilization" for all processes.  This is not quite what we want
	// since coordination processes, watch, etc are usually at 0% and that's the way it should be.
	// To do better, we should not allocate any chars for processes that are always at 100%.
	fmt.Println("100%              " + s)
	for _, l := range input[1:] {
		s := l[0] + "  "
		for i, x := range l[1:] {
			if x != "" {
				n, err := strconv.Atoi(x)
				if err != nil {
					panic(err)
				}
				for range int(math.Ceil(float64(n) / *scale)) {
					s += chars[i]
				}
			}
		}
		fmt.Println(s)
	}
}
