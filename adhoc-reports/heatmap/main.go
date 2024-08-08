// Given some text input, extract two fields and produce a square heatmap with the values for one
// field running along the x axis and the other down the y axis.
//
// The default grid size (not currently overridable) is 20x20 boxes.  Input values are quantized
// into those boxes and each box is incremented when a value falls in it.  Before quantization the
// input values are clamped.  The min and max values for clamping are 0 and 100 ie percentages (not
// currently overridable).
//
// The number of quantities for quantization is the number of colors in the color table (currently
// also 20 but this is accidental).
//
// In the heat map, boxes with zero values are displayed as white and all other values are given a
// color from the color map; the colors are selected by dividing the count by the ratio of max cell
// value to the number of colors.
//
//
// Eg to produce a png format heat map from the default output of `sonalyze sacct`, where the fifth
// column (one-based) has relative cpu usage and the sixth column has relative memory usage:
//
// sonalyze sacct -data-dir ~/sonalyze-test/data/fox.educloud.no -f 3d -fmt awk,default | ./heatmap -a 5 -b 6 -ppm | pnmtopng > fox-default-3d.png
//
// (Without pnmtopng, save the ppm and view it in emacs and in image viewers.)
//
// The default text output displays raw non-zero cell values and is usually fine in a terminal window.
//
// TODO:
//
// - The text heatmap will not be formatted properly once a cell value goes over 999
//
// - Allowing the size of the grid to be specified may be useful (the size should divide 100 evenly).
//
// - It's pretty brittle to depend on the default fields for `sonalyze sacct`, would be better to
//   specify that we want specifically rcpu and rmem, take the first and second fields for example.
//   That would make this program more general - we could plot other quantities?
//
// - There are useful variations here that collect other information and place that in the grid, for
//   subsequent analysis - part of this program or part of something else?

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

const (
	N               = 20
	width           = N
	height          = N
	pixelsPerSquare = 20
)

var grid [width * height]int

var (
	a   = flag.Int("a", 1, "One-based index of first column")
	b   = flag.Int("b", 2, "One-based index of second column")
	ppm = flag.Bool("ppm", false, "Dump output in ppm format rather than text")
)

func main() {
	flag.Parse()
	if *a < 1 || *b < 1 {
		fmt.Fprintf(os.Stderr, "%s: Nonsense column(s): %d %d\n", os.Args[0], *a, *b)
		os.Exit(1)
	}

	scan := bufio.NewScanner(os.Stdin)
	var any bool
	var maxval int
	for scan.Scan() {
		any = true
		fields := strings.Split(scan.Text(), " ")
		if *a > len(fields) || *b > len(fields) {
			fmt.Fprintf(os.Stderr, "%s: Nonsense column(s): %d %d\n", os.Args[0], *a, *b)
			os.Exit(1)
		}
		v1, e1 := strconv.ParseUint(fields[int(*a-1)], 10, 64)
		v2, e2 := strconv.ParseUint(fields[int(*b-1)], 10, 64)
		if e1 != nil || e2 != nil {
			fmt.Fprintf(os.Stderr, "%s: Unparseable field(s): '%s' '%s'\n", os.Args[0], fields[int(*a-1)], fields[int(*b-1)])
			os.Exit(1)
		}
		v1 = min(100, max(1, v1))
		v2 = min(100, max(1, v2))
		const quant = 100 / N
		xLoc := ((v1 - 1) / quant)
		yLoc := ((v2 - 1) / quant)
		grid[yLoc*width+xLoc]++
		maxval = max(maxval, grid[yLoc*width+xLoc])
	}

	if any {
		switch {
		case *ppm:
			printPpm(maxval)
		default:
			printText()
		}
	}
}

func printText() {
	fmt.Printf(" ")
	for x := 0; x < width; x++ {
		fmt.Printf(" ---")
	}
	fmt.Println()
	for y := 0; y < height; y++ {
		fmt.Printf("|")
		for x := 0; x < width; x++ {
			v := grid[y*width+x]
			if v == 0 {
				fmt.Printf("    ")
			} else {
				fmt.Printf("%4d", v)
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
	realWidth := width*pixelsPerSquare + 2
	realHeight := height*pixelsPerSquare + 2
	fmt.Printf("P6 %d %d 255\n", realWidth, realHeight)
	for x := 0; x < realWidth; x++ {
		os.Stdout.Write(black)
	}
	for y := 0; y < height; y++ {
		for j := 0; j < pixelsPerSquare; j++ {
			os.Stdout.Write(black)
			for x := 0; x < width; x++ {
				val := grid[y*width+x]
				var c = white
				if val > 0 {
					c = colors[int(math.Floor(float64(val)/q))]
				}
				for i := 0; i < pixelsPerSquare; i++ {
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
