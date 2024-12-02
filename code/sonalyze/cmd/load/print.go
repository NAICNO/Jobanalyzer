package load

import (
	"cmp"
	"fmt"
	"io"
	"slices"

	"go-utils/config"
	. "sonalyze/table"
)

//go:generate ../../../generate-table/generate-table -o load-table.go print.go

/*TABLE load

package load

import (
    "go-utils/gpuset"
	. "sonalyze/common"
	. "sonalyze/table"
)

type GpuSet = gpuset.GpuSet

%%

FIELDS *ReportRecord

	Now                 DateTimeValue alias:"now"      desc:"The current time (yyyy-mm-dd hh:mm)"
	DateTime            DateTimeValue alias:"datetime" desc:"The starting date and time of the aggregation window (yyyy-mm-dd hh:mm)"
	Date                DateValue     alias:"date"     desc:"The starting date of the aggregation window (yyyy-mm-dd)"
	Time                TimeValue     alias:"time"     desc:"The startint time of the aggregation window (hh:mm)"
	Cpu                 int           alias:"cpu"      desc:"Average CPU utilization in percent in the aggregation window (100% = 1 core)"
	RelativeCpu         int           alias:"rcpu"     desc:"Average relative CPU utilization in percent in the aggregation window (100% = all cores)"
	VirtualGB           int           alias:"mem"      desc:"Average virtual memory utilization in GiB in the aggregation window"
	RelativeVirtualMem  int           alias:"rmem"     desc:"Relative virtual memory utilization in GiB in the aggregation window (100% = system RAM)"
	ResidentGB          int           alias:"res"      desc:"Average resident memory utilization in GiB in the aggregation window"
	RelativeResidentMem int           alias:"rres"     desc:"Relative resident memory utilization in GiB in the aggregation window (100% = system RAM)"
	Gpu                 int           alias:"gpu"      desc:"Average GPU utilization in percent in the aggregation window (100% = 1 card)"
	RelativeGpu         int           alias:"rgpu"     desc:"Average relative GPU utilization in percent in the aggregation window (100% = all cards)"
	GpuGB               int           alias:"gpumem"   desc:"Average gpu memory utilization in GiB in the aggregation window"
	RelativeGpuMem      int           alias:"rgpumem"  desc:"Average relative gpu memory utilization in GiB in the aggregation window (100% = all GPU RAM)"
	Gpus                GpuSet        alias:"gpus"     desc:"GPU device numbers used by the job, 'none' if none or 'unknown' in error states"
	Hostname            Ustr          alias:"host"     desc:"Combined host names of jobs active in the aggregation window"

GENERATE ReportRecord

HELP LoadCommand

  Aggregate process data across users and commands on a host and bucket into
  time slots, producing a view of system load.  Default output format is 'fixed'.

ALIASES

  default date,time,cpu,mem,gpu,gpumem,gpumask
  Default Date,Time,Cpu,ResidentGB,Gpu,GpuGB,Gpus

DEFAULTS default

ELBAT*/

type LoadReport struct {
	hostname string
	conf     *config.NodeConfigRecord // may be nil, beware
	records  []*ReportRecord
}

func (lc *LoadCommand) printStreams(
	out io.Writer,
	reports []LoadReport,
) {

	// Sort hosts lexicographically.  This is not ideal because hosts like c1-10 vs c1-5 are not in
	// the order we expect but at least it's predictable.
	slices.SortStableFunc(reports, func(a, b LoadReport) int {
		return cmp.Compare(a.hostname, b.hostname)
	})

	// The handling of hostname for "fixed" formatting is a hack: if the records don't contain the
	// host name, print the host above each run of records for a host.
	explicitHost := false
	for _, f := range lc.PrintFields {
		if f.Name == "host" || f.Name == "Hostname" {
			explicitHost = true
			break
		}
	}

	// The handling of JSON is a hack: we synthesize an array of objects where each object describes
	// a system and its load data.  It caters to an older structure of the system.  These days, the
	// clients would get the load data for all hosts as a simple array, then do a second query on
	// `config` to get host information, and then join the results.
	//
	// TODO: It may be useful to introduce another formatting option to ask for the raw JSON array,
	// or to (painfully) migrate all clients from old JSON output to a cleaner form.
	jsonSep := ""
	if lc.PrintOpts.Json {
		fmt.Fprint(out, "[")
	}

	for _, report := range reports {
		if lc.PrintOpts.Json {
			fmt.Fprint(out, jsonSep)
			jsonSep = ","
		}
		if lc.PrintOpts.Fixed && !explicitHost {
			fmt.Fprintf(out, "HOST: %s\n", report.hostname)
		}

		// For JSON, add richer information about the host so that the client does not have to
		// synthesize this information itself.
		if lc.PrintOpts.Json {
			description := "Unknown"
			gpuCards := 0
			if report.conf != nil {
				description = report.conf.Description
				gpuCards = report.conf.GpuCards
			}
			fmt.Fprintf(
				out,
				"{\"system\":{"+
					"\"hostname\":\"%s\","+
					"\"description\":\"%s\","+
					"\"gpucards\":\"%d\""+
					"},\"records\":",
				report.hostname,
				QuoteJson(description),
				gpuCards,
			)
		}

		records := report.records
		if !lc.All {
			// Invariant: there's always at least one record
			records = records[len(records)-1:]
		}
		FormatData(
			out,
			lc.PrintFields,
			loadFormatters,
			lc.PrintOpts,
			records,
		)

		if lc.PrintOpts.Json {
			fmt.Fprint(out, "}")
		}
	}
	if lc.PrintOpts.Json {
		fmt.Fprint(out, "]")
	}
}
