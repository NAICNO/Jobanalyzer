// Output format:
//
// For HTML, the "fields" list must have a single string from the set
// "cpu","mem","res","gpu","gpumem".  This is the per-process value we print, along with a
// timestamp.  But the layout differs among these three formats:
//
//  - For html+javascript, each row consists of the data field for a process at a time, that is,
//    time increases along the x axis.  The data are emitted into a JS `DATASETS` array where each
//    element is an object with a `label` and `data` field, the former carrying a string value that
//    is a label for the row and the latter being an array of numbers.  This is accompanied by a
//    `LABELS` array that contains one value per value of the arrays of `DATASETS`, a string
//    representing the timestamp for the colum.

package profile

import (
	"fmt"
	"io"
	"strings"

	"sonalyze/data/sample"
	. "sonalyze/table"
)

func (pc *ProfileCommand) printHtml(
	out io.Writer,
	jobId uint32,
	m *profData,
	processes []sample.SampleStream,
	pif *processIndexFactory,
	host, user string,
) error {
	labels, rows, err := pc.collectHtml(m, processes, pif)
	if err != nil {
		return err
	}
	quant := pc.PrintFields[0].Name
	formatHtml(out, jobId, quant, host, user, int(pc.Bucket), labels, rows)
	return nil
}

// This returns a rows along with column labels (timestamps).
//
// TODO: IMPROVEME: This is exactly the same as collectCsvOrAwk except it does some catenation and
// the data are transposed.  The two could and should be merged, and the catenation logic moved into
// the HTML formatter.

func (pc *ProfileCommand) collectHtml(
	m *profData,
	processes []sample.SampleStream,
	pif *processIndexFactory,
) (labels []string, rows []string, err error) {
	var formatter func(*profDatum) string
	formatter, err = lookupSingleFormatter(pc.PrintFields, true)
	if err != nil {
		return
	}
	rowNames := m.rows()

	labels = make([]string, len(rowNames))
	for i, rn := range rowNames {
		labels[i] = "\"" + formatTime(rn) + "\""
	}

	// Iterate by processes rather than m.cols() since process order is what we care about.

	rowLabels := make([]string, len(processes))
	for i, p := range processes {
		// Here, use the raw pid for compatibility with the Rust code
		pid := pif.indexFor(p[0])
		rowLabels[i] = fmt.Sprintf("%s (%s)", p[0].Cmd.String(), pif.nameFor(pid))
	}

	rows = make([]string, len(processes))
	for i, p := range processes {
		cn := pif.indexFor(p[0])
		s := ""
		sep := ""
		for _, rn := range rowNames {
			entry := m.get(rn, cn)
			s += sep
			if entry != nil {
				s += formatter(entry)
			}
			sep = ","
		}
		label := rowLabels[i]
		rows[i] = fmt.Sprintf("{label: \"%s\", data: [%s]}", label, s)
	}

	return
}

var htmlCaptions = map[string]string{
	"cpu":           "Y axis: Number of CPU cores (1.0 = 1 core at 100%)",
	"CpuUtilPct":    "Y axis: Number of CPU cores (1.0 = 1 core at 100%)",
	"mem":           "Y axis: Virtual primary memory in GB",
	"VirtualMemGB":  "Y axis: Virtual primary memory in GB",
	"rss":           "Y axis: Resident primary memory in GB",
	"res":           "Y axis: Resident primary memory in GB",
	"ResidentMemGB": "Y axis: Resident primary memory in GB",
	"gpu":           "Y axis: Number of GPU cards in use (1.0 = 1 card at 100%)",
	"Gpu":           "Y axis: Number of GPU cards in use (1.0 = 1 card at 100%)",
	"gpumem":        "Y axis: Real GPU memory in GB",
	"GpuMemGB":      "Y axis: Real GPU memory in GB",
}

func formatHtml(
	unbufOut io.Writer,
	job uint32,
	quant, host, user string,
	bucket int,
	labels []string,
	rows []string,
) {
	out := Buffered(unbufOut)
	defer out.Flush()

	title := fmt.Sprintf("`%s` profile of job %d on `%s`, user `%s`", quant, job, host, user)
	if bucket > 1 {
		title += fmt.Sprintf(", bucketing=%d", bucket)
	}
	fmt.Fprintf(out, `
<html>
 <head>
  <title>%s</title>
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
  <center><h1>%s</h1></center>
  <div><canvas id="chart_node"></canvas></div>
  <center><b>X axis: UTC timestamp</b><br><b>%s</b></center>
 </body>
<html>
`,
		title,
		strings.Join(labels, ","),
		strings.Join(rows, ","),
		title,
		htmlCaptions[quant],
	)
}
