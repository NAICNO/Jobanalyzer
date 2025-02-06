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
//
// The profile can be printed padded with blanks so that it makes columns, one for each process,
// use -pad to enable this.
//
// Jobs that are mostly idle can be removed from the profile with -cull, this is on by default.

package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
)

const marks = "-+*:=&^%$#@!abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var (
	scale = flag.Float64("scale", 10.0, "Value scale factor")
	pad   = flag.Bool("pad", false, "Pad with blanks")
	cull  = flag.Bool("cull", true, "Remove jobs with utilization near zero")
)

func main() {
	flag.Parse()
	input, err := csv.NewReader(os.Stdin).ReadAll()
	check(err)
	maxutil := ""
	indices := make([]int, 0)
	culled := make([]string, 0)
	for i, h := range input[0][1:] {
		keep := !*cull
		if *cull {
			for _, l := range input[1:] {
				n := 0
				if x := l[i+1]; x != "" {
					n, err = strconv.Atoi(x)
					check(err)
				}
				if n >= 5 {
					keep = true
					break
				}
			}
		}
		if keep {
			fmt.Printf(" %c   %s\n", marks[i], h)
			for range int(100 / *scale) {
				maxutil += "_"
			}
			indices = append(indices, i)
		} else {
			culled = append(culled, h)
		}
	}
	for _, c := range culled {
		fmt.Printf("     %s  (mostly idle)\n", c)
	}
	if !*pad {
		fmt.Println("100%              " + maxutil)
	}
	for _, l := range input[1:] {
		s := l[0] + "  "
		l = l[1:]
		for _, i := range indices {
			n := 0
			if x := l[i]; x != "" {
				n, err = strconv.Atoi(x)
				check(err)
			}
			k := 0
			for range int(math.Ceil(float64(n) / *scale)) {
				s += string(marks[i])
				k++
			}
			if *pad {
				for k < int(100 / *scale) {
					s += " "
					k++
				}
			}
		}
		fmt.Println(s)
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
