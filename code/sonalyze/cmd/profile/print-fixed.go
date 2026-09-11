// Output format:
//
// For fixed output, the output is presented in blocks, one block per timestamp (time increases
// along the y axis).  The first line of a block has the timestamp and data for the first process;
// subsequent lines of the block have only data for subsequent processes at that time.  On each
// line, all the requested fields are printed.  Essentially, the fixed output is the flattened JSON
// output: the intermediate job objects are not printed.

package profile

import (
	"io"
	"math"

	"sonalyze/data/sample"
	. "sonalyze/table"
)

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

func (pc *ProfileCommand) printFixed(
	out io.Writer,
	m *profData,
	processes []sample.SampleStream,
	pif *processIndexFactory,
) error {
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
	return nil
}

func (pc *ProfileCommand) collectFixed(
	m *profData,
	processes []sample.SampleStream,
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
