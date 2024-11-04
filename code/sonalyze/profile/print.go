// For one particular job, break it down in its component processes and print individual stats for
// each process for each time slot.
//
// The data fields are raw data items from Sample records.  There are no average or peak values
// because we only have samples at the start and beginning of the time slot.  There is a
// straightforward derivation from the Sample values to relative values should we need that.
//
//   cpu: the cpu_util_pct field
//   mem: the mem_gb field
//   res: the res_gb field
//   gpu: the gpu_pct field
//   gpumem: the gpumem_gb field
//   nproc: the rolledup field + 1, this is not printed if every process has rolledup=0
//   command: the command field
//
//
// Output formats:
//
// For HTML, CSV and AWK output the "fields" list must have a single string from the set
// "cpu","mem","res","gpu","gpumem".  This is the per-process value we print, along with a
// timestamp.  But the layout differs among these three formats:
//
//  - For csv, each row has a timestamp and then one data field for each process at that time, that
//    is, time increases along the y axis.  A header row is printed, with "time" in the first field
//    and process information (command and pid) in each subsequent field.  There will be empty
//    fields where there are no data for a process at a time.
//
//  - For awk, the format should be as for csv except the field separator is a space, as normal.
//
//  - For html+javascript, each row consists of the data field for a process at a time, that is,
//    time increases along the x axis.  The data are emitted into a JS `DATASETS` array where each
//    element is an object with a `label` and `data` field, the former carrying a string value that
//    is a label for the row and the latter being an array of numbers.  This is accompanied by a
//    `LABELS` array that contains one value per value of the arrays of `DATASETS`, a string
//    representing the timestamp for the colum.
//
// For JSON output, we print an array of objects, each representing a "job" at a "time" (these are
// fields in each object).  In each object, there is a field "points" which is an array of data
// points.  Each data point has the value for all the data fields (regardless of what was requested)
// for that job at that time.
//
// For fixed output, the output is presented in blocks, one block per timestamp (time increases
// along the y axis).  The first line of a block has the timestamp and data for the first process;
// subsequent lines of the block have only data for subsequent processes at that time.  On each
// line, all the requested fields are printed.  Essentially, the fixed output is the flattened JSON
// output: the intermediate job objects are not printed.

package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"strings"

	uslices "go-utils/slices"

	. "sonalyze/command"
	. "sonalyze/common"
	"sonalyze/sonarlog"
)

func (pc *ProfileCommand) printProfile(
	out io.Writer,
	jobId uint32,
	host, user string,
	hasRolledup bool,
	m *profData,
	processes sonarlog.SampleStreams,
) error {
	if hasRolledup {
		// Add the "nproc" / "NumProcs" field if it is required and the fields are still the
		// defaults.  This is pretty dumb (but compatible with the Rust code); it gets even dumber
		// with two versions of the default fields string.
		currFields := strings.Join(
			uslices.Map(
				pc.PrintFields,
				func(fs FieldSpec) string { return fs.Name },
			),
			",",
		)
		var newFields string
		if currFields == v0ProfileDefaultFields {
			newFields = v0ProfileDefaultFieldsWithNproc
		} else if currFields == v1ProfileDefaultFields {
			newFields = v1ProfileDefaultFieldsWithNproc
		}
		if newFields != "" {
			pc.PrintFields = uslices.Map(
				strings.Split(newFields, ","),
				func(name string) FieldSpec { return FieldSpec{Name: name} },
			)
		}
	}

	if pc.PrintOpts.Csv || pc.PrintOpts.Awk {
		header, matrix, err := pc.collectCsvOrAwk(m, processes)
		if err != nil {
			return err
		}
		if pc.PrintOpts.Csv {
			FormatRawRowmajorCsv(out, header, matrix)
		} else {
			FormatRawRowmajorAwk(out, header, matrix)
		}
	} else if pc.htmlOutput {
		labels, rows, err := pc.collectHtml(m, processes)
		if err != nil {
			return err
		}
		quant := pc.PrintFields[0].Name
		formatHtml(out, jobId, quant, host, user, int(pc.Bucket), labels, rows)
	} else if pc.PrintOpts.Fixed {
		data, err := pc.collectFixed(m, processes)
		if err != nil {
			return err
		}
		FormatData(
			out,
			pc.PrintFields,
			profileFormatters,
			pc.PrintOpts,
			uslices.Map(data, func(x *fixedLine) any { return x }),
		)
	} else if pc.PrintOpts.Json {
		formatJson(out, m, processes, pc.testNoMemory)
	} else {
		panic("Unknown print format")
	}
	return nil
}

// This returns a rows along with column labels (timestamps).
//
// TODO: IMPROVEME: This is exactly the same as collectCsvOrAwk except it does some catenation and
// the data are transposed.  The two could and should be merged, and the catenation logic moved into
// the HTML formatter.

func (pc *ProfileCommand) collectHtml(
	m *profData,
	processes sonarlog.SampleStreams,
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
		rowLabels[i] = fmt.Sprintf("%s (%d)", (*p)[0].Cmd.String(), (*p)[0].Pid)
	}

	rows = make([]string, len(processes))
	for i, p := range processes {
		cn := processId((*p)[0])
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

// This returns a row-of-columns representation, and the header will be nil if a header is not
// explicitly requested in the options.

func (pc *ProfileCommand) collectCsvOrAwk(
	m *profData,
	processes sonarlog.SampleStreams,
) (header []string, matrix [][]string, err error) {
	var formatter func(*profDatum) string
	formatter, err = lookupSingleFormatter(pc.PrintFields, false)
	if err != nil {
		return
	}

	if pc.PrintOpts.Header {
		header = []string{"time"}
		sep := ""
		if pc.PrintOpts.Csv {
			sep = " "
		}
		for _, process := range processes {
			header = append(header,
				fmt.Sprintf(
					"%s%s(%d)", (*process)[0].Cmd, sep, (*process)[0].Pid))
		}
	}

	// Iterate by processes rather than m.cols() since process order is what we care about.

	rowNames := m.rows()
	matrix = make([][]string, len(rowNames))
	for i := range matrix {
		matrix[i] = make([]string, len(processes)+1)
	}

	for y, rn := range rowNames {
		matrix[y][0] = formatTime(rn)
		for x, p := range processes {
			pid := processId((*p)[0])
			entry := m.get(rn, pid)
			if entry != nil {
				matrix[y][x+1] = formatter(entry)
			}
		}
	}

	return
}

// The output is time-sorted but timestamps may be duplicated.  The first profileLine at a timestamp
// has a non-zero time value, the rest are zero.

// TODO: Should the derivation of these data be lifted to perform.go?
// TODO: Merge with JSON-formatting logic somehow?

type fixedLine struct {
	Timestamp     DateTimeValueOrBlank `alias:"time" desc:"Time of the start of the profiling bucket"`
	CpuUtilPct    int                  `alias:"cpu"  desc:"CPU utilization in percent, 100% = 1 core (except for HTML)"`
	VirtualMemGB  int                  `alias:"mem"  desc:"Main virtual memory usage in GiB"`
	ResidentMemGB int                  `alias:"res"  desc:"Main resident memory usage in GiB"`
	Gpu           int                  `alias:"gpu"  desc:"GPU utilization in percent, 100% = 1 card (except for HTML)"`
	GpuMemGB      int                  `alias:"gpumem" desc:"GPU resident memory usage in GiB (across all cards)"`
	Command       Ustr                 `alias:"cmd"  desc:"Name of executable starting the process"`
	NumProcs      IntOrEmpty           `alias:"nproc" desc:"Number of rolled-up processes, blank for zero"`
}

func (pc *ProfileCommand) collectFixed(
	m *profData,
	processes sonarlog.SampleStreams,
) (data []*fixedLine, err error) {
	rowNames := m.rows()

	// The length of this will eventually be the number of defined elements in m
	data = make([]*fixedLine, 0)

	// Iterate by processes rather than m.cols() since process order is what we care about.
	for _, rn := range rowNames {
		first := true
		for _, p := range processes {
			cn := processId((*p)[0])
			entry := m.get(rn, cn)
			if entry != nil {
				var timestamp int64
				var numprocs int
				if first {
					timestamp = entry.s.Timestamp
				}
				if entry.s.Rolledup > 0 {
					numprocs = int(entry.s.Rolledup) + 1
				}
				data = append(data, &fixedLine{
					Timestamp:     DateTimeValueOrBlank(timestamp),
					CpuUtilPct:    int(math.Round(float64(entry.cpuUtilPct))),
					VirtualMemGB:  int(math.Round(float64(entry.cpuKB) / (1024 * 1024))),
					ResidentMemGB: int(math.Round(float64(entry.rssAnonKB) / (1024 * 1024))),
					Gpu:           int(math.Round(float64(entry.gpuPct))),
					GpuMemGB:      int(math.Round(float64(entry.gpuKB) / (1024 * 1024))),
					Command:       entry.s.Cmd,
					NumProcs:      IntOrEmpty(numprocs),
				})
				first = false
			}
		}
	}

	return
}

// TODO: Ideally this function would just delegate to the derived formatters where it can, so that
// we can have canonical formatters.  Indeed, the "scaled" thing is really part of the computation,
// not formatting anyway, so should be applied earlier, and we should probably derive some table of
// values to be printed.

func lookupSingleFormatter(
	fields []FieldSpec,
	scaleCpuGpu bool,
) (formatter func(*profDatum) string, err error) {
	n := 0
	for _, f := range fields {
		switch f.Name {
		case "cpu", "CpuUtilPct":
			if scaleCpuGpu {
				formatter = formatCpuUtilPctScaled
			} else {
				formatter = formatCpuUtilPct
			}
			n++
		case "mem", "VirtualMemGB":
			formatter = formatMem
			n++
		case "res", "rss", "ResidentMemGB":
			formatter = formatRes
			n++
		case "gpu", "Gpu":
			if scaleCpuGpu {
				formatter = formatGpuPctScaled
			} else {
				formatter = formatGpuPct
			}
			n++
		case "gpumem", "GpuMemGB":
			formatter = formatGpuMem
			n++
		default:
			err = fmt.Errorf("Not a known field: %s", f.Name)
			return
		}
	}
	if n != 1 {
		err = errors.New("formatted output needs exactly one valid field")
	}
	return
}

// TODO: These formatters are now only used by lookupSingleFormatter().  It's bad that the
// formatting logic also does things like conversion and truncation - we could have generated data
// first and then formatted.

func formatCpuUtilPct(s *profDatum) string {
	return fmt.Sprint(math.Round(float64(s.cpuUtilPct)))
}

func formatCpuUtilPctScaled(s *profDatum) string {
	return fmt.Sprintf("%.1f", float64(s.cpuUtilPct)/100)
}

func formatMem(s *profDatum) string {
	return fmt.Sprint(math.Round(float64(s.cpuKB) / (1024 * 1024)))
}

func formatRes(s *profDatum) string {
	return fmt.Sprint(math.Round(float64(s.rssAnonKB) / (1024 * 1024)))
}

func formatGpuPct(s *profDatum) string {
	return fmt.Sprint(math.Round(float64(s.gpuPct)))
}

func formatGpuPctScaled(s *profDatum) string {
	return fmt.Sprintf("%.1f", float64(s.gpuPct)/100)
}

func formatGpuMem(s *profDatum) string {
	return fmt.Sprint(math.Round(float64(s.gpuKB) / (1024 * 1024)))
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

// TODO: Canonical names too?
// TODO: Merge with fixed-formatting logic somehow?

func formatJson(
	out io.Writer,
	m *profData,
	processes sonarlog.SampleStreams,
	noMemory bool,
) {
	type jsonPoint struct {
		Command    string `json:"command"`
		Pid        uint32 `json:"pid"`
		CpuUtilPct int    `json:"cpu"`
		CpuGB      uint64 `json:"mem"`
		RssAnonGB  uint64 `json:"res"`
		GpuPct     int    `json:"gpu"`
		GpuMemGB   uint64 `json:"gpumem"`
		Nproc      int    `json:"nproc"`
	}
	type jsonJob struct {
		Time   string      `json:"time"`
		Job    uint32      `json:"job"`
		Points []jsonPoint `json:"points"`
	}
	objects := make([]jsonJob, 0)
	for _, rn := range m.rows() {
		points := make([]jsonPoint, 0)
		var e *profDatum
		for _, p := range processes {
			cn := processId((*p)[0])
			entry := m.get(rn, cn)
			if entry == nil {
				continue
			}
			if e == nil {
				e = entry
			}
			var cpuGB, gpuGB uint64
			if !noMemory {
				cpuGB = entry.cpuKB / (1024 * 1024)
				gpuGB = entry.gpuKB / (1024 * 1024)
			}
			points = append(points, jsonPoint{
				Command:    entry.s.Cmd.String(),
				Pid:        entry.s.Pid, // Hm
				CpuUtilPct: int(math.Round(float64(entry.cpuUtilPct))),
				CpuGB:      cpuGB,
				RssAnonGB:  entry.rssAnonKB / (1024 * 1024),
				GpuPct:     int(math.Round(float64(entry.gpuPct))),
				GpuMemGB:   gpuGB,
				Nproc:      int(entry.s.Rolledup) + 1,
			})
		}
		objects = append(objects, jsonJob{
			Time:   formatTime(rn),
			Job:    e.s.Job,
			Points: points,
		})
	}
	e := json.NewEncoder(out)
	e.SetEscapeHTML(false)
	err := e.Encode(objects)
	if err != nil {
		panic("JSON encoding")
	}
}

func (pc *ProfileCommand) MaybeFormatHelp() *FormatHelp {
	// This is wrong if some of the data have rolled-up fields because in that case the default
	// fields includes the nproc field; so be it.  We're not going to wait until after reading the
	// data to respond to --fmt=help, and this corner case is not worth fixing with more complexity.
	return StandardFormatHelp(pc.Fmt, profileHelp, profileFormatters, profileAliases, profileDefaultFields)
}

const profileHelp = `
profile
  Compute aggregate job behavior across processes by time step, for some job
  attributes.  Default output format is 'fixed'.
`

const v0ProfileDefaultFields = "time,cpu,mem,gpu,gpumem,cmd"
const v1ProfileDefaultFields = "Timestamp,CpuUtilPct,VirtualMemGB,Gpu,GpuMemGB,Command"
const profileDefaultFields = v0ProfileDefaultFields

const v0ProfileDefaultFieldsWithNproc = "time,cpu,mem,gpu,gpumem,nproc,cmd"
const v1ProfileDefaultFieldsWithNproc = "Timestamp,CpuUtilPct,VirtualMemGB,Gpu,GpuMemGB,NumProcs,Command"
const profileDefaultFieldsWithNproc = v0ProfileDefaultFieldsWithNproc

// MT: Constant after initialization; immutable
var profileAliases = map[string][]string{
	"default":   strings.Split(profileDefaultFields, ","),
	"v0default": strings.Split(v0ProfileDefaultFields, ","),
	"v1default": strings.Split(v1ProfileDefaultFields, ","),
	"rss":       []string{"res"},
}

// MT: Constant after initialization; immutable
var profileFormatters = DefineTableFromTags(
	// TODO: Go 1.22, reflect.TypeFor[fixedLine]
	reflect.TypeOf((*fixedLine)(nil)).Elem(),
	nil,
)

func formatTime(t int64) string {
	return FormatYyyyMmDdHhMmUtc(t)
}
