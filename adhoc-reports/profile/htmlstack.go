// This program assumes jobanalyzer profile csv-format input on stdin and will try to print a
// sensible stacked profile as an HTML+JS program that will render the profile when loaded in
// a browser.
//
// Normally you'd run eg
//
//   sonalyze profile -cluster fox -f 4d -j 1307357 -fmt csv,gpu | go run stack.go > test.html
//
// Normally you'd then open test.html in a browser.
//
// See stack.go for other information.

package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

const (
	keeper     = 5
	defaultMax = 100
)

var (
	cull     = flag.Bool("cull", true, "Remove jobs with utilization near zero")
	verbose  = flag.Bool("v", false, "Verbose output")
	separate = flag.Bool("separate", false, "Give each process its own graph")
)

func main() {
	flag.Parse()
	input, err := csv.NewReader(os.Stdin).ReadAll()
	check(err)

	// Process the header, categorize processes, compute maximum values.
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
			if n >= keeper {
				keep = true
			}
		}
		if keep || !*cull {
			selected = append(selected, hdr)
			indices = append(indices, i)
			ave = append(ave, int(math.Round(float64(sum)/float64((len(input)-1)))))
		} else {
			culled = append(culled, hdr)
		}
	}

	// The labels are the timestamps, the x axis labels for the overall plot
	labels := make([]string, len(input[1:]))
	for proc, i := range input[1:] {
		labels[proc] = fmt.Sprintf(`"%s"`, i[0])
	}

	// array[process] - one label per process
	datalabels := make([]string, len(indices))
	for proc, i := range indices {
		datalabels[proc] = input[0][1:][i]
	}

	if *separate {
		separatePlots(indices, selected, labels, datalabels, input[1:])
	} else {
		singlePlot(indices, selected, labels, datalabels, input[1:])
	}
}

// Stack them to reveal total utilization

func singlePlot(indices []int, selected, labels, datalabels []string, input [][]string) {
	// array[timestamp][process] - one datum at each point
	datasets := make([][]string, len(labels))
	for i := range len(labels) {
		datasets[i] = make([]string, len(indices))
	}

	// Create profile grid
	for time, l := range input {
		l = l[1:]
		v := 0
		for proc, i := range indices {
			if x := l[i]; x != "" {
				n, err := strconv.Atoi(x)
				check(err)
				v += n
				datasets[time][proc] = fmt.Sprint(v)
			}
		}
	}

	// create rows from grid
	//
	// Each row is a json object, {label: l, data: [d,...]} where the l identifies the process and
	// each d is a data point.  There must be exactly as many data points as there are outer labels,
	// but a data point can be the empty string if there are no data; this will be common as
	// processes come and go (and data can also be missing).  The processes are stacked vertically;
	// the bottom-most data set are the raw data, then the next set has its raw data added to the
	// data of the row below, and so on.
	rows := make([]string, 0)
	for i := range len(indices) {
		xs := make([]string, 0, len(labels))
		for j := range len(labels) {
			xs = append(xs, datasets[j][i])
		}
		rows = append(rows, fmt.Sprintf(`{label: "%s", data: [%s]}`, selected[i], strings.Join(xs, ",")))
	}

	fmt.Printf(`
<html>
 <head>
  <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
  <script>
var LABELS = [%s];
var DATASETS = [%s];
function render() {
  new Chart(document.getElementById("chart_node"), {
    type: 'line',
    data: {
      labels: LABELS,
      datasets: DATASETS
    },
    options: { scales: { x: { beginAtZero: true }, y: { beginAtZero: true } } }
  })
}
  </script>
 </head>
 <body onload="render()">
  <div><canvas id="chart_node"></canvas></div>
 </body>
<html>
`,
		strings.Join(labels, ","),
		strings.Join(rows, ","),
	)
}

func separatePlots(indices []int, selected, labels, datalabels []string, input [][]string) {
	canvasIx := 0

	// Header - load chart lib, define render function
	fmt.Printf(`
<html>
 <head>
  <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
  <script>LABELS = [%s]</script>
  <script>
function render() {

`,
		strings.Join(labels, ","))

	for k, proc := range indices {
		data := make([]string, 0)
		for _, l := range input {
			data = append(data, l[1:][proc])
		}
		fmt.Printf(`
  new Chart(document.getElementById("chart_node_%d"), {
    type: 'line',
    data: {
      datasets: [{data:[%s], label: "%s"}],
      labels: LABELS,
    },
    options: { scales: { x: { beginAtZero: true }, y: { beginAtZero: true } } }
  })
`, canvasIx, strings.Join(data, ","), datalabels[k])
		canvasIx++
	}

	// End header
	fmt.Printf(`
  }
  </script>
 </head>
`)

	// Body - define canvases, call the render function
	fmt.Printf(`
 <body onload="render()">
`)
	for c := range canvasIx {
		fmt.Printf(`<div><canvas id="chart_node_%d"></canvas></div>
`, c)
	}
	fmt.Printf(`
 </body>
<html>
`)
}


func check(err error) {
	if err != nil {
		panic(err)
	}
}
