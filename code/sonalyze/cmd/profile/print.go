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
	"slices"
	"strings"

	"sonalyze/data/sample"
	. "sonalyze/table"
)

func (pc *ProfileCommand) printProfile(
	out io.Writer,
	jobId uint32,
	host, user string,
	hasRolledup bool,
	m *profData,
	processes sample.SampleStreams,
	pif *processIndexFactory,
) error {
	// Add the "nproc" field if it is required (we can't do it until rollup has happened) and the
	// fields are still the defaults.  This is not quite compatible with older sonalyze: here we
	// have default fields only if no fields were specified (defaults were applied), while in older
	// code a field list identical to the default would be taken as having default fields too.  The
	// new logic is better.
	//
	// Anyway, this hack depends on nothing interesting having happened to pc.Fmt or pc.PrintFields
	// after the initial parsing.
	//
	// The situation is the same for the "host" field: it gets added late if there is more than one
	// host in the profile.  This field could be anywhere for json output, but for fixed output we
	// want it second, after the timestamp.
	hasDefaultFields := pc.Fmt == ""
	if hasDefaultFields {
		fields := slices.Clone(profileAliases["default"])
		changed := false
		if hasRolledup {
			fields = slices.Insert(fields, len(fields)-1, "nproc")
			changed = true
		}
		if pif.isMultiHost() {
			fields = slices.Insert(fields, 1, "host")
			changed = true
		}
		if changed {
			pc.PrintFields, _, _ = ParseFormatSpec(
				strings.Join(fields, ","),
				"",
				profileFormatters,
				profileAliases,
			)
		}
	}

	if pc.PrintOpts.Csv || pc.PrintOpts.Awk {
		header, matrix, err := pc.collectCsvOrAwk(m, processes, pif)
		if err != nil {
			return err
		}
		if pc.PrintOpts.Csv {
			FormatRawRowmajorCsv(out, header, matrix)
		} else {
			FormatRawRowmajorAwk(out, header, matrix)
		}
	} else if pc.htmlOutput {
		labels, rows, err := pc.collectHtml(m, processes, pif)
		if err != nil {
			return err
		}
		quant := pc.PrintFields[0].Name
		formatHtml(out, jobId, quant, host, user, int(pc.Bucket), labels, rows)
	} else if pc.PrintOpts.Fixed {
		data, err := pc.collectFixed(m, processes, pif)
		if err != nil {
			return err
		}
		FormatData(
			out,
			pc.PrintFields,
			profileFormatters,
			pc.PrintOpts,
			data,
		)
	} else if pc.PrintOpts.Json {
		formatJson(out, m, processes, pif, pc.testNoMemory)
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
	processes sample.SampleStreams,
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

// This returns a row-of-columns representation, and the header will be nil if a header is not
// explicitly requested in the options.
//
// NOTE, for multi-host jobs the host is encoded in the header names, this is a tricky matter and
// really not a super happy outcome (esp for awk, since the header is not printed by default).  It's
// no worse than the PID or command, really, but the header field is becoming dangerously
// overloaded.  The syntax of each header field is always <something>(pid@host) or <something>(pid)
// so it is not a crisis, but we could consider cleaning this up.

func (pc *ProfileCommand) collectCsvOrAwk(
	m *profData,
	processes sample.SampleStreams,
	pif *processIndexFactory,
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
			pid := pif.indexFor(process[0])
			header = append(header,
				fmt.Sprintf(
					"%s%s(%s)", process[0].Cmd, sep, pif.nameFor(pid)))
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
			pid := pif.indexFor(p[0])
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

// TODO: Should the derivation of fixedLine data be lifted to perform.go?
// TODO: Merge fixed formatting with JSON-formatting logic somehow?

//go:generate ../../../generate-table/generate-table -o profile-table.go print.go

/*TABLE profile

package profile

%%

FIELDS *fixedLine

 Timestamp     DateTimeValueOrBlank alias:"time"    desc:"Time of the start of the profiling bucket"
 Hostname      Ustr                 alias:"host"    desc:"Host on which process ran"
 CpuUtilPct    int                  alias:"cpu"     desc:"CPU utilization in percent, 100% = 1 core (except for HTML)"
 VirtualMemGB  int                  alias:"mem"     desc:"Main virtual memory usage in GiB"
 ResidentMemGB int                  alias:"res,rss" desc:"Main resident memory usage in GiB"
 Gpu           int                  alias:"gpu"     desc:"GPU utilization in percent, 100% = 1 card (except for HTML)"
 GpuMemGB      int                  alias:"gpumem"  desc:"GPU resident memory usage in GiB (across all cards)"
 Command       Ustr                 alias:"cmd"     desc:"Name of executable starting the process"
 NumProcs      IntOrEmpty           alias:"nproc"   desc:"Number of rolled-up processes, blank for zero"

GENERATE fixedLine

SUMMARY ProfileCommand

Experimental: Print profile information for one aspect of a particular job.

This prints a table across time of utilization of various resources of the
processes in a job.  The job can be on multiple nodes.  For fixed formatting,
all resources are printed on one line per process per time step; similarly for
json all resources for a process at a time step are embedded in a single object.
For CSV, AWK and HTML output a single resource must be selected with -fmt, and
its utilization across processes per time step is printed; start with the CSV
output to understand this (eg -fmt csv,gpu will show the table for the gpu
resource in CSV form).  Commands, process IDs and host names are encoded in the
output header in an idiosyncratic, but useful, form.  Note that no header is
printed by default for AWK.  Be sure to file bugs for missing functionality.

HELP ProfileCommand

  Compute aggregate job behavior across processes by time step, for some job
  attributes.  Default output format is 'fixed'.

ALIASES

  default time,cpu,mem,gpu,gpumem,cmd
  Default Timestamp,CpuUtilPct,VirtualMemGB,Gpu,GpuMemGB,Command

DEFAULTS default

ELBAT*/

func (pc *ProfileCommand) collectFixed(
	m *profData,
	processes sample.SampleStreams,
	pif *processIndexFactory,
) (data []*fixedLine, err error) {
	rowNames := m.rows()

	// The length of this will eventually be the number of defined elements in m
	data = make([]*fixedLine, 0)

	// Iterate by processes rather than m.cols() since process order is what we care about.
	for _, rn := range rowNames {
		first := true
		for _, p := range processes {
			cn := pif.indexFor(p[0])
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
					Hostname:      entry.s.Hostname,
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
			err = fmt.Errorf("Not a printable field for this output format: %s", f.Name)
			return
		}
	}
	if n != 1 {
		err = errors.New("Formatted output needs exactly one valid field")
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
	processes sample.SampleStreams,
	pif *processIndexFactory,
	noMemory bool,
) {
	type jsonPoint struct {
		Command    string `json:"command"`
		Host       string `json:"host,omitempty"`
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
			cn := pif.indexFor(p[0])
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
			hn, pid := pif.hostAndPid(cn)
			var hostname string
			if pif.isMultiHost() {
				hostname = hn.String()
			}
			points = append(points, jsonPoint{
				Command:    entry.s.Cmd.String(),
				Host:       hostname,
				Pid:        pid,
				CpuUtilPct: int(math.Round(float64(entry.cpuUtilPct))),
				CpuGB:      cpuGB,
				RssAnonGB:  uint64(math.Round(float64(entry.rssAnonKB) / (1024 * 1024))),
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

func formatTime(t int64) string {
	return FormatYyyyMmDdHhMmUtc(t)
}
