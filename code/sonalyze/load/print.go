package load

import (
	"fmt"
	"io"
	"math"
	"sort"
	"time"

	"go-utils/config"
	"go-utils/gpuset"

	. "sonalyze/command"
	. "sonalyze/common"
	"sonalyze/sonarlog"
)

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

	// The handling of hostname is a hack.
	// The handling of JSON is also a hack.
	explicitHost := false
	for _, f := range lc.PrintFields {
		if f == "host" {
			explicitHost = true
			break
		}
	}
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
		report := GenerateReport(data, time.Now().Unix(), conf)
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

// MT: Constant after initialization; immutable
var loadAliases = map[string][]string{ /* No aliases */ }

// TODO here:
//  - ReportRecord (maybe) and GenerateReport (for sure) should move into perform.go
//  - perform.go should generate the reports
//  - printStreams() should take a set of reports
//  - probably there will be no cfg argument to printStreams()
//  - test all fields esp GpuSet
//  - run test suite
//  - compare against a standard run
//  - define aliases and defaults
//  - should UnixTime be DateTimeValue?
//  - defaults should be defined in this file not in load.go

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

// `sys` may be nil if none of the requested fields use its data, so we must guard against that.
func GenerateReport(input []sonarlog.Sample, now int64, sys *config.NodeConfigRecord) (result []ReportRecord) {
	result = make([]ReportRecord, 0, len(input))
	for _, d := range input {
		var relativeCpu, relativeVirtualMem, relativeResidentMem, relativeGpu, relativeGpuMem int
		if sys != nil {
			if sys.CpuCores > 0 {
				relativeCpu = int(math.Round(float64(d.CpuUtilPct) / float64(sys.CpuCores)))
			}
			if sys.MemGB > 0 {
				relativeVirtualMem = int(math.Round(float64(d.S.CpuKib) / (1024 * 1024) / float64(sys.MemGB) * 100.0))
				relativeResidentMem = int(math.Round(float64(d.S.RssAnonKib) / (1024 * 1024) / float64(sys.MemGB) * 100.0))
			}
			if sys.GpuCards > 0 {
				// GpuPct is already scaled by 100 so don't do it again
				relativeGpu = int(math.Round(float64(d.S.GpuPct) / float64(sys.GpuCards)))
			}
			if sys.GpuMemGB > 0 {
				relativeGpuMem = int(math.Round(float64(d.S.GpuKib) / (1024 * 1024) / float64(sys.GpuMemGB) * 100))
			}
		}
		result = append(result, ReportRecord{
			Now:         UnixTime(now),
			DateTime:    UnixTime(d.S.Timestamp),
			Date:        DateValue(d.S.Timestamp),
			Time:        TimeValue(d.S.Timestamp),
			Cpu:         int(d.CpuUtilPct),
			RelativeCpu: relativeCpu,
			VirtualGB:   int(d.S.CpuKib / (1024 * 1024)),
			RelativeVirtualMem: relativeVirtualMem,
			ResidentGB:  int(d.S.RssAnonKib / (1024 * 1024)),
			RelativeResidentMem: relativeResidentMem,
			Gpu: int(d.S.GpuPct),
			RelativeGpu: relativeGpu,
			GpuGB: int(d.S.GpuKib / (1024 * 1024)),
			RelativeGpuMem: relativeGpuMem,
			Gpus: d.S.Gpus,
			Hostname: d.S.Host,
		})
	}
	return
}

// MT: Constant after initialization; immutable
var loadFormatters = ReflectFormatters[ReportRecord](nil)
