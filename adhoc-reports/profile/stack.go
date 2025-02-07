// This program assumes jobanalyzer profile csv-format input on stdin and will try to print a
// sensible stacked profile.
//
// Normally you'd run eg
//
//   sonalyze profile -cluster fox -f 4d -j 1307357 -fmt csv,gpu | go run stack.go
//
// The default output will be a sequence of lines in increasing timestamp order:
//
//   100%      ---------------------------
//   timestamp xxxxyyyyzzzz
//   timestamp xxyyyyyyyyyzzzz
//   timestamp xz
//
// where the xyz are characters for a particular process.  The mapping from profile to char is
// printed at the start of the profile.  The 100% line indicates the max extent of the profile if
// all processes are at 100%.
//
// OPTIONS
//
// *Columns*: The profile can be printed padded with blanks so that it makes proper columns, one for
// each process, use -pad to enable this.  Then you'll see the following; the 100% bar is not
// printed b/c redundant:
//
//   timestamp xxxx     yyyy      zzzz
//   timestamp xx       yyyyyyyyy zzzz
//   timestamp x                  z
//
// *Noise*: Jobs that are mostly idle can be removed from the profile with -cull, this is on by
// default, use -cull=false to disable.  Culled processes are listed at the start of the profile.
//
// *Bar height*: By default, profile values (see below) are scaled by (divided by) 10 b/c this is
// "roughly right" for many use cases, but sometimes this will not work out well; use -scale to
// adjust.
//
// *Value ranges*: To plot charts like these there needs to be a per-profile maximum value to chart
// against.  This is not so easy.  Below, "the value" is the value provided by the profiler for the
// quantity in question:
//
//  - For CPU, the value is percent of one core.  The correct maximum value is therefore allocated
//    cores * 100, where "allocated" cores is easy for Slurm data but hard otherwise.
//
//  - For GPU, the value is percent of one card, so a sensible maximum is 100, but in principle a
//    single process can use multiple cards and the true maximum value may be higher, again
//    Slurm data can advise, otherwise you must know.
//
//  - For CPU and GPU memory the value is GB and the max value is the allocated value for CPU RAM
//    and the size of the card for GPU RAM.
//
// To keep this program simple, it's possible to specify the max value using the -max parameter.
// The user is responsible (for now) for figuring out what that value should be.
//
// If the largest observed value is larger than -max then -max is silently set to the largest
// observed value.
//
// *Debugging*: Use -v to see some computed internal values.

package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
)

const (
	marks      = "-+*:=&^%$#@!abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	keeper     = 5
	defaultMax = 100
)

var (
	scale   = flag.Float64("scale", 10.0, "Value scale factor")
	maxval  = flag.Int("max", defaultMax, "Default maximum data value")
	pad     = flag.Bool("pad", false, "Pad with blanks")
	cull    = flag.Bool("cull", true, "Remove jobs with utilization near zero")
	verbose = flag.Bool("v", false, "Verbose output")
)

func main() {
	flag.Parse()
	input, err := csv.NewReader(os.Stdin).ReadAll()
	check(err)

	// Process the header, categorize processes, compute maximum values.
	maxobserved := 0
	indices := make([]int, 0)
	selected := make([]string, 0)
	culled := make([]string, 0)
	ave := make([]int, 0)
	for i, hdr := range input[0][1:] {
		keep := false
		sum := 0
		for _, l := range input[1:] {
			n := 0
			if x := l[i+1]; x != "" {
				n, err = strconv.Atoi(x)
				check(err)
			}
			sum += n
			maxobserved = max(maxobserved, n)
			if n >= keeper {
				keep = true
			}
		}
		if keep || !*cull {
			selected = append(selected, fmt.Sprintf(" %c   %s", marks[i], hdr))
			indices = append(indices, i)
			ave = append(ave, int(math.Round(float64(sum)/float64((len(input)-1)))))
		} else {
			culled = append(culled, hdr)
		}
	}
	if *maxval == defaultMax {
		*maxval = max(*maxval, maxobserved)
	}
	maxTicks := int(math.Ceil(100 / *scale))
	valsPerTick := float64(*maxval) / float64(maxTicks)

	// Print header and debug information
	for _, c := range selected {
		fmt.Println(c)
	}
	for _, c := range culled {
		fmt.Printf("     %s  (mostly idle)\n", c)
	}
	if *verbose {
		fmt.Printf("maxval = %d\n", *maxval)
		fmt.Printf("valsPerTick = %g\n", valsPerTick)
		fmt.Printf("maxTicks = %d\n", maxTicks)
		fmt.Printf("relevant = %d\n", len(indices))
	}
	if !*pad {
		maxutil := ""
		for range maxTicks * len(indices) {
			maxutil += "_"
		}
		fmt.Println("100%              " + maxutil)
	}

	// Print profile
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
			for range int(math.Ceil(float64(n) / valsPerTick)) {
				if k < maxTicks {
					s += string(marks[i])
					k++
				}
			}
			if *pad {
				for k < maxTicks {
					s += " "
					k++
				}
			}
		}
		fmt.Println(s)
	}

	// Print footer
	foot := "Average           "
	footFmt := fmt.Sprintf("%%-%ds", maxTicks)
	for _, a := range ave {
		foot += fmt.Sprintf(footFmt, fmt.Sprint(a))
	}
	fmt.Println(foot)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
