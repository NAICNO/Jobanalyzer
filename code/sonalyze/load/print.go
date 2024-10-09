package load

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"go-utils/config"
	"go-utils/gpuset"

	. "sonalyze/command"
	. "sonalyze/common"
	"sonalyze/sonarlog"
)

// TODO here:
//  - perform.go should generate the reports
//  - printStreams() should take a list of reports
//  - probably there will be no cfg argument to printStreams() then
//  - printRequiresConfig must also test canonical names
//
//  - test all fields esp GpuSet
//  - run test suite
//  - compare against a standard run
//
//  - should "UnixTime" be "DateTimeValue"?  Almost certainly!  It's useful to separate formatting from representation.
//  - should DateTimeValue, DateValue, TimeValue be defined in the formatting code then?  Possibly.
//
//  - in the formatting code, can we make better use of "this thing implements fmt.Stringer" as a catchall?

type ReportRecord struct {
	Now                 UnixTime      `alias:"now"      desc:"The current time (yyyy-mm-dd hh:mm)"`
	DateTime            UnixTime      `alias:"datetime" desc:"The starting date and time of the aggregation window (yyyy-mm-dd hh:mm)"`
	Date                DateValue     `alias:"date"     desc:"The starting date of the aggregation window (yyyy-mm-dd)"`
	Time                TimeValue     `alias:"time"     desc:"The startint time of the aggregation window (hh:mm)"`
	Cpu                 int           `alias:"cpu"      desc:"Average CPU utilization in percent in the aggregation window (100% = 1 core)"`
	RelativeCpu         int           `alias:"rcpu"     desc:"Average relative CPU utilization in percent in the aggregation window (100% = all cores)"`
	VirtualGB           int           `alias:"mem"      desc:"Average virtual memory utilization in GiB in the aggregation window"`
	RelativeVirtualMem  int           `alias:"rmem"     desc:"Relative virtual memory utilization in GiB in the aggregation window (100% = system RAM)"`
	ResidentGB          int           `alias:"res"      desc:"Average resident memory utilization in GiB in the aggregation window"`
	RelativeResidentMem int           `alias:"rres"     desc:"Relative resident memory utilization in GiB in the aggregation window (100% = system RAM)"`
	Gpu                 int           `alias:"gpu"      desc:"Average GPU utilization in percent in the aggregation window (100% = 1 card)"`
	RelativeGpu         int           `alias:"rgpu"     desc:"Average relative GPU utilization in percent in the aggregation window (100% = all cards)"`
	GpuGB               int           `alias:"gpumem"   desc:"Average gpu memory utilization in GiB in the aggregation window"`
	RelativeGpuMem      int           `alias:"rgpumem"  desc:"Average relative gpu memory utilization in GiB in the aggregation window (100% = all GPU RAM)"`
	Gpus                gpuset.GpuSet `alias:"gpus"     desc:"GPU device numbers used by the job, 'none' if none or 'unknown' in error states"`
	Hostname            Ustr          `alias:"host"     desc:"Combined host names of jobs active in the aggregation window"`
}

func (lc *LoadCommand) printRequiresConfig() bool {
	for _, f := range lc.PrintFields {
		switch f {
		case "rcpu", "rmem", "rres", "rgpu", "rgpumem":
			return true
		}
	}
	return false
}

func (lc *LoadCommand) printStreams(
	out io.Writer,
	cfg *config.ClusterConfig,
	mergedConf *config.NodeConfigRecord,
	mergedStreams sonarlog.SampleStreams,
) {

	// Sort hosts lexicographically.  This is not ideal because hosts like c1-10 vs c1-5 are not in
	// the order we expect but at least it's predictable.
	sort.Stable(sonarlog.HostSortableSampleStreams(mergedStreams))

	// The handling of hostname for "fixed" formatting is a hack: if the records don't contain the
	// host name, print the host above each run of records for a host.
	explicitHost := false
	for _, f := range lc.PrintFields {
		if f == "host" || f == "Hostname" {
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

	for _, stream := range mergedStreams {
		if lc.PrintOpts.Json {
			fmt.Fprint(out, jsonSep)
			jsonSep = ","
		}
		hostname := (*stream)[0].S.Host.String()
		if lc.PrintOpts.Fixed && !explicitHost {
			fmt.Fprintf(out, "HOST: %s\n", hostname)
		}
		conf := mergedConf
		if conf == nil && cfg != nil {
			conf = cfg.LookupHost(hostname)
		}

		// For JSON, add richer information about the host so that the client does not have to
		// synthesize this information itself.
		if lc.PrintOpts.Json {
			description := "Unknown"
			gpuCards := 0
			if conf != nil {
				description = conf.Description
				gpuCards = conf.GpuCards
			}
			fmt.Fprintf(
				out,
				"{\"system\":{"+
					"\"hostname\":\"%s\","+
					"\"description\":\"%s\","+
					"\"gpucards\":\"%d\""+
					"},\"records\":",
				hostname,
				QuoteJson(description),
				gpuCards,
			)
		}

		var data sonarlog.SampleStream = *stream
		if !lc.All {
			// Invariant: there's always at least one record
			data = data[len(data)-1:]
		}
		report := generateReport(data, time.Now().Unix(), conf)
		FormatData(out, lc.PrintFields, loadFormatters, lc.PrintOpts, report, PrintMods(0))

		if lc.PrintOpts.Json {
			fmt.Fprint(out, "}")
		}
	}
	if lc.PrintOpts.Json {
		fmt.Fprint(out, "]")
	}
}

func (lc *LoadCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(lc.Fmt, loadHelp, loadFormatters, loadAliases, loadDefaultFields)
}

const loadHelp = `
load
  Aggregate process data across users and commands on a host and bucket into
  time slots, producing a view of system load.  Default output format is 'fixed'.
`

const (
	// Note the v1 default switches from virtual to real memory; we keep virtual in v0 default for
	// backward compatibility.
	v0LoadDefaultFields = "date,time,cpu,mem,gpu,gpumem,gpumask"
	v1LoadDefaultFields = "Date,Time,Cpu,ResidentGB,Gpu,GpuGB,Gpus"
	loadDefaultFields   = v0LoadDefaultFields
)

// MT: Constant after initialization; immutable
var loadAliases = map[string][]string{
	"default":   strings.Split(loadDefaultFields, ","),
	"v0default": strings.Split(v0LoadDefaultFields, ","),
	"v1default": strings.Split(v1LoadDefaultFields, ","),
}

// MT: Constant after initialization; immutable
var loadFormatters = ReflectFormatters[ReportRecord](nil)
