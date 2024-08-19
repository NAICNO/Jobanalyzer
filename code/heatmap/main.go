// Given some space-separated text input, extract two fields and produce a square heatmap with the
// values for one field running along the x axis and the other down the y axis.
//
// The default grid size (override with -n) is 20x20 boxes.  Input values are quantized into those
// boxes and each box is incremented when a value falls in it.  Before quantization the input values
// are clamped.  The min and max values for clamping are 0 and 100 ie percentages (not currently
// overridable).
//
// The number of quantities for quantization is the number of colors in the color table.
//
// In the heat map, boxes with zero values are displayed as white and all other values are given a
// color from the color map; the colors are selected by dividing the count by the ratio of max cell
// value to the number of colors.
//
//
// Eg to produce a png format heat map from the default output of `sonalyze sacct`, where the fifth
// column (one-based) has relative cpu usage and the sixth column has relative memory usage:
//
// sonalyze sacct -data-dir ~/sonalyze-test/data/fox.educloud.no -f 3d -fmt awk,default | \
//   ./heatmap -a 5 -b 6 -ppm | \
//   pnmtopng > fox-default-3d.png
//
// (Without pnmtopng, save the ppm and view it in emacs and in image viewers.)
//
// The default text output displays raw non-zero cell values and is usually fine in a terminal window.
//
// TODO:
//
// - There are useful variations here that collect other information and place that in the grid, for
//   subsequent analysis - part of this program or part of something else?
//
// - Incorporate the png generation into this program to simplify pipelines

package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

var (
	a   = flag.Int("a", 1, "One-based index of first column")
	b   = flag.Int("b", 2, "One-based index of second column")
	n   = flag.Int("n", 20, "Number of boxes along each dimension, 1-100, must divide 100")
	ppm = flag.Bool("ppm", false, "PPM format output")
	csv = flag.Bool("csv", false, "Text output, CSV format (comma-separated)")
	awk = flag.Bool("awk", false, "Text output, awk format (single-space-separated)")
	px  = flag.Int("px", 20, "Pixels per square")
	v   = flag.Bool("v", false, "Verbose")
)

var (
	width, height int
	grid          []int
)

func main() {
	flag.Parse()
	if *a < 1 || *b < 1 {
		fmt.Fprintf(os.Stderr, "%s: Nonsense column(s): %d %d\n", os.Args[0], *a, *b)
		os.Exit(1)
	}
	if *n < 1 || *n > 100 || 100%*n != 0 {
		fmt.Fprintf(
			os.Stderr,
			"%s: the grid size n must divide 100 and be in the range [1..100]: %d\n",
			os.Args[0],
			*n)
		os.Exit(1)
	}
	formats := 0
	if *ppm {
		formats++
	}
	if *csv {
		formats++
	}
	if *awk {
		formats++
	}
	if formats > 1 {
		fmt.Fprintf(os.Stderr, "%s: Choose at most one output format\n", os.Args[0])
		os.Exit(1)
	}
	if *px < 1 {
		fmt.Fprintf(os.Stderr, "%s: Pixels per square must be at least 1: %d\n", *px)
		os.Exit(1)
	}

	height = *n
	width = *n
	grid = make([]int, height*width)
	scan := bufio.NewScanner(os.Stdin)
	var count int
	var maxval int
	for scan.Scan() {
		count++
		fields := strings.Split(scan.Text(), " ")
		if *a > len(fields) || *b > len(fields) {
			fmt.Fprintf(os.Stderr, "%s: Nonsense column(s): %d %d\n", os.Args[0], *a, *b)
			os.Exit(1)
		}
		v1, e1 := strconv.ParseUint(fields[int(*a-1)], 10, 64)
		v2, e2 := strconv.ParseUint(fields[int(*b-1)], 10, 64)
		if e1 != nil || e2 != nil {
			fmt.Fprintf(
				os.Stderr,
				"%s: Unparseable field(s): '%s' '%s'\n",
				os.Args[0],
				fields[int(*a-1)],
				fields[int(*b-1)])
			os.Exit(1)
		}
		v1 = min(100, max(1, v1))
		v2 = min(100, max(1, v2))
		quant := uint64(100 / *n)
		xLoc := int((v1 - 1) / quant)
		yLoc := int((v2 - 1) / quant)
		grid[yLoc*width+xLoc]++
		maxval = max(maxval, grid[yLoc*width+xLoc])
	}

	if *v {
		fmt.Fprintf(os.Stderr, "record count = %d\n", count)
	}

	if count > 0 {
		switch {
		case *ppm:
			printPpm(maxval)
		default:
			printText(maxval)
		}
	}
}

func printText(maxval int) {
	diglen := max(len(fmt.Sprint(maxval)), 3)
	theFmt := fmt.Sprintf("%%%dd", diglen+1)
	theSpaces := " " + strings.Join(make([]string, diglen+1), " ")
	theHeader := " " + strings.Join(make([]string, diglen+1), "-")
	quiet := *csv || *awk
	sep := " "
	if *csv {
		sep = ","
	}
	if !quiet {
		fmt.Printf(" ")
		for x := 0; x < width; x++ {
			fmt.Print(theHeader)
		}
		fmt.Println()
	}
	for y := 0; y < height; y++ {
		if !quiet {
			fmt.Printf("|")
		}
		for x := 0; x < width; x++ {
			v := grid[y*width+x]
			switch {
			case v == 0 && !quiet:
				fmt.Printf(theSpaces)
			case quiet:
				if x > 0 {
					fmt.Print(sep)
				}
				fmt.Print(v)
			default:
				fmt.Printf(theFmt, v)
			}
		}
		fmt.Println()
	}
}

/* "magma" from https://waldyrious.net/viridis-palette-generator/ */
var (
	colors = [][]byte{
		[]byte{252, 253, 191},
		[]byte{253, 229, 167},
		[]byte{254, 205, 144},
		[]byte{254, 180, 123},
		[]byte{253, 155, 107},
		[]byte{250, 129, 95},
		[]byte{244, 105, 92},
		[]byte{232, 83, 98},
		[]byte{214, 69, 108},
		[]byte{192, 58, 118},
		[]byte{171, 51, 124},
		[]byte{148, 44, 128},
		[]byte{128, 37, 130},
		[]byte{106, 28, 129},
		[]byte{86, 20, 125},
		[]byte{63, 15, 114},
		[]byte{41, 17, 90},
		[]byte{21, 14, 56},
		[]byte{7, 6, 28},
		[]byte{0, 0, 4},
	}
	black = []byte{0, 0, 0}
	white = []byte{255, 255, 255}
)

func printPpm(maxval int) {
	q := float64(maxval) / float64(len(colors)-1)
	realWidth := width*(*px) + 2
	realHeight := height*(*px) + 2
	fmt.Printf("P6 %d %d 255\n", realWidth, realHeight)
	for x := 0; x < realWidth; x++ {
		os.Stdout.Write(black)
	}
	for y := 0; y < height; y++ {
		for j := 0; j < (*px); j++ {
			os.Stdout.Write(black)
			for x := 0; x < width; x++ {
				val := grid[y*width+x]
				var c = white
				if val > 0 {
					c = colors[int(math.Floor(float64(val)/q))]
				}
				for i := 0; i < (*px); i++ {
					os.Stdout.Write(c)
				}
			}
			os.Stdout.Write(black)
		}
	}
	for x := 0; x < realWidth; x++ {
		os.Stdout.Write(black)
	}
}
