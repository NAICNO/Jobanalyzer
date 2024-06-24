// This produces an NxN grid (N <= 100) with relative cpu usage across x and relative memory across
// y.  Readings from sacct data are quantized into N buckets with 1..k going into the first bucket
// and 100-k+1..100 going into the last (values are clamped to 1..100).  Usefully, N divides 100
// evenly.  The factors of 100 are 2,2,5,5; go wild.
//
// A grid cell gets a point if a job falls into that bucket.
//
// Eg to produce a graphical heat map:
//
// sonalyze sacct -data-dir ~/sonalyze-test/data/fox.educloud.no -f 3d -fmt awk,default | ./heatmap -ppm > foo.ppm
//
// ppms can be opened in emacs and in image viewers, at least.  To convert to easier-to-handle png,
// install netpbm and then run pnmtopng foo.ppm > foo.png.
//
// In principle the ppm format supports embedded comments to allow self-identifying files, but
// neither emacs nor netpbm seems to understand those.
//
// TODO:
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
	N = 20
	width = N
	height = N
	pixelsPerSquare = 20
)

var grid [width*height]int

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
	black = []byte{0,0,0}
	white = []byte{255,255,255}
)

var (
	ppm = flag.Bool("ppm", false, "Dump output in ppm format rather than text")
)

func main() {
	flag.Parse()

	scan := bufio.NewScanner(os.Stdin)
	var errs int
	var any bool
	var maxval int
	for scan.Scan() {
		any = true
		fields := strings.Split(scan.Text(), " ")
		cpu, e1 := strconv.ParseUint(fields[4], 10, 64)
		mem, e2 := strconv.ParseUint(fields[5], 10, 64)
		if e1 != nil || e2 != nil {
			errs++
			continue
		}
		cpu = min(100, max(1, cpu))
		mem = min(100, max(1, mem))
		const quant = 100/N
		cpuLoc := ((cpu-1) / quant)
		memLoc := ((mem-1) / quant)
		grid[memLoc*width+cpuLoc]++
		maxval = max(maxval, grid[memLoc*width+cpuLoc])
	}

	if any {
		if *ppm {
			q := float64(maxval) / float64(len(colors)-1)
			realWidth := width*pixelsPerSquare+2
			realHeight := height*pixelsPerSquare+2
			fmt.Printf("P6 %d %d 255\n", realWidth, realHeight)
			for x := 0 ; x < realWidth ; x++ {
				os.Stdout.Write(black)
			}
			for y := 0 ; y < height ; y++ {
				for j := 0 ; j < pixelsPerSquare ; j++ {
					os.Stdout.Write(black)
					for x := 0 ; x < width ; x++ {
						val := grid[y*width+x]
						var c = white
						if val > 0 {
							c = colors[int(math.Floor(float64(val)/q))]
						}
						for i := 0 ; i < pixelsPerSquare ; i++ {
							os.Stdout.Write(c)
						}
					}
					os.Stdout.Write(black)
				}
			}
			for x := 0 ; x < realWidth ; x++ {
				os.Stdout.Write(black)
			}
		} else {
			fmt.Printf(" ")
			for x := 0 ; x < width ; x++ {
				fmt.Printf(" ---")
			}
			fmt.Println()
			for y := 0 ; y < height ; y++ {
				fmt.Printf("|")
				for x := 0 ; x < width ; x++ {
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
	}
}
