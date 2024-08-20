package load

import (
	"fmt"
	"io"
	"math"
	"sort"
	"time"

	"go-utils/config"
	. "sonalyze/command"
	"sonalyze/common"
	"sonalyze/sonarlog"
)

func (lc *LoadCommand) printRequiresConfig() bool {
	for _, f := range lc.printFields {
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
	for _, f := range lc.printFields {
		if f == "host" {
			explicitHost = true
			break
		}
	}
	jsonSep := ""
	if lc.printOpts.Json {
		fmt.Fprint(out, "[")
	}
	for _, stream := range mergedStreams {
		if lc.printOpts.Json {
			fmt.Fprint(out, jsonSep)
			jsonSep = ","
		}
		hostname := (*stream)[0].S.Host.String()
		if lc.printOpts.Fixed && !explicitHost {
			fmt.Fprintf(out, "HOST: %s\n", hostname)
		}
		conf := mergedConf
		if conf == nil && cfg != nil {
			conf = cfg.LookupHost(hostname)
		}

		// For JSON, add richer information about the host so that the client does not have to
		// synthesize this information itself.
		if lc.printOpts.Json {
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

		ctx := loadCtx{
			sys: conf,
			now: time.Now().Unix(),
		}

		var data sonarlog.SampleStream = *stream
		if !lc.All {
			// Invariant: there's always at least one record
			data = data[len(data)-1:]
		}
		FormatData(out, lc.printFields, loadFormatters, lc.printOpts, data, ctx)

		if lc.printOpts.Json {
			fmt.Fprint(out, "}")
		}
	}
	if lc.printOpts.Json {
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

type loadCtx struct {
	sys *config.NodeConfigRecord
	now int64
}

// MT: Constant after initialization; immutable
var loadFormatters = map[string]Formatter[sonarlog.Sample, loadCtx]{
	"now": {
		func(_ sonarlog.Sample, ctx loadCtx) string {
			return common.FormatYyyyMmDdHhMmUtc(ctx.now)
		},
		"The current time (yyyy-mm-dd hh:mm)",
	},
	"datetime": {
		func(d sonarlog.Sample, _ loadCtx) string {
			return common.FormatYyyyMmDdHhMmUtc(d.S.Timestamp)
		},
		"The starting time of the aggregation window (yyyy-mm-dd hh:mm)",
	},
	"date": {
		func(d sonarlog.Sample, _ loadCtx) string {
			return time.Unix(d.S.Timestamp, 0).UTC().Format("2006-01-02")
		},
		"The starting time of the aggregation window (yyyy-mm-dd)",
	},
	"time": {
		func(d sonarlog.Sample, _ loadCtx) string {
			return time.Unix(d.S.Timestamp, 0).UTC().Format("15:04")
		},
		"The starting time of the aggregation window (hh:mm)",
	},
	"cpu": {
		func(d sonarlog.Sample, _ loadCtx) string {
			return fmt.Sprint(int(d.CpuUtilPct))
		},
		"Average CPU utilization in percent in the aggregation window (100% = 1 core)",
	},
	"rcpu": {
		func(d sonarlog.Sample, ctx loadCtx) string {
			if ctx.sys.CpuCores == 0 {
				return "0"
			}
			return fmt.Sprint(math.Round(float64(d.CpuUtilPct) / float64(ctx.sys.CpuCores)))
		},
		"Average relative CPU utilization in percent in the aggregation window (100% = all cores)",
	},
	"mem": {
		func(d sonarlog.Sample, _ loadCtx) string {
			return fmt.Sprint(d.S.CpuKib / (1024 * 1024))
		},
		"Average virtual memory utilization in GiB in the aggregation window",
	},
	"rmem": {
		func(d sonarlog.Sample, ctx loadCtx) string {
			if ctx.sys.MemGB == 0 {
				return "0"
			}
			return fmt.Sprint(math.Round(float64(d.S.CpuKib) / (1024 * 1024) / float64(ctx.sys.MemGB) * 100.0))
		},
		"Relative virtual memory utilization in GiB in the aggregation window (100% = system RAM)",
	},
	"res": {
		func(d sonarlog.Sample, _ loadCtx) string {
			return fmt.Sprint(d.S.RssAnonKib / (1024 * 1024))
		},
		"Average resident memory utilization in GiB in the aggregation window",
	},
	"rres": {
		func(d sonarlog.Sample, ctx loadCtx) string {
			if ctx.sys.MemGB == 0 {
				return "0"
			}
			return fmt.Sprint(math.Round(float64(d.S.RssAnonKib) / (1024 * 1024) / float64(ctx.sys.MemGB) * 100.0))
		},
		"Relative resident memory utilization in GiB in the aggregation window (100% = system RAM)",
	},
	"gpu": {
		func(d sonarlog.Sample, _ loadCtx) string {
			return fmt.Sprint(int(d.S.GpuPct))
		},
		"Average GPU utilization in percent in the aggregation window (100% = 1 card)",
	},
	"rgpu": {
		func(d sonarlog.Sample, ctx loadCtx) string {
			if ctx.sys.GpuCards == 0 {
				return "0"
			}
			// GpuPct is already scaled by 100 so don't do it again
			return fmt.Sprint(math.Round(float64(d.S.GpuPct) / float64(ctx.sys.GpuCards)))
		},
		"Average relative GPU utilization in percent in the aggregation window (100% = all cards)",
	},
	"gpumem": {
		func(d sonarlog.Sample, _ loadCtx) string {
			return fmt.Sprint(int(d.S.GpuKib / (1024 * 1024)))
		},
		"Average gpu memory utilization in GiB in the aggregation window",
	},
	"rgpumem": {
		func(d sonarlog.Sample, ctx loadCtx) string {
			if ctx.sys.GpuMemGB == 0 {
				return "0"
			}
			return fmt.Sprint(math.Round(float64(d.S.GpuKib) / (1024 * 1024) / float64(ctx.sys.GpuMemGB) * 100))
		},
		"Average relative gpu memory utilization in GiB in the aggregation window (100% = all GPU RAM)",
	},
	"gpus": {
		func(d sonarlog.Sample, _ loadCtx) string {
			return d.S.Gpus.String()
		},
		"GPU device numbers used by the job, 'none' if none or 'unknown' in error states",
	},
	"host": {
		func(d sonarlog.Sample, _ loadCtx) string {
			return d.S.Host.String()
		},
		"Combined host names of jobs active in the aggregation window",
	},
}
