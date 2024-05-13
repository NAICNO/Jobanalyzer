package parse

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"sort"
	"time"

	"go-utils/config"
	"go-utils/hostglob"
	"go-utils/maps"
	. "sonalyze/command"
	"sonalyze/common"
	"sonalyze/sonarlog"
)

type ParseCommand struct /* implements AnalysisCommand */ {
	SharedArgs
	MergeByHostAndJob bool
	MergeByJob        bool
	Clean             bool
	Fmt               string

	// Synthesized and other
	printFields []string
	printOpts   *FormatOptions
}

func (_ *ParseCommand) Summary() []string {
	return []string{
		"Export sample data in various formats, after optional preprocessing.",
	}
}

func (pc *ParseCommand) Add(fs *flag.FlagSet) {
	pc.SharedArgs.Add(fs)
	fs.BoolVar(&pc.MergeByHostAndJob, "merge-by-host-and-job", false,
		"Merge streams that have the same host and job ID")
	fs.BoolVar(&pc.MergeByJob, "merge-by-job", false,
		"Merge streams that have the same job ID, across hosts")
	fs.BoolVar(&pc.Clean, "clean", false,
		"Clean the job but perform no merging")
	fs.StringVar(&pc.Fmt, "fmt", "",
		"Select `field,...` and format for the output [default: try -fmt=help]")
}

func (pc *ParseCommand) ReifyForRemote(x *Reifier) error {
	e1 := pc.SharedArgs.ReifyForRemote(x)
	x.Bool("merge-by-host-and-job", pc.MergeByHostAndJob)
	x.Bool("merge-by-job", pc.MergeByJob)
	x.Bool("clean", pc.Clean)
	x.String("fmt", pc.Fmt)
	return e1
}

func (pc *ParseCommand) Validate() error {
	var e1, e2 error
	e1 = pc.SharedArgs.Validate()
	spec := parseDefaultFields
	if pc.Fmt != "" {
		spec = pc.Fmt
	}
	var others map[string]bool
	pc.printFields, others, e2 = ParseFormatSpec(spec, parseFormatters, parseAliases)
	if e2 == nil && len(pc.printFields) == 0 {
		e2 = errors.New("No output fields were selected in format string")
	}
	pc.printOpts = StandardFormatOptions(others, DefaultCsv)

	return errors.Join(e1, e2)
}

func (pc *ParseCommand) DefaultRecordFilters() (
	allUsers, skipSystemUsers, excludeSystemCommands, excludeHeartbeat bool,
) {
	// `parse` implies `--user=-` b/c we're interested in raw data.
	allUsers = true
	skipSystemUsers = false
	excludeSystemCommands = false
	excludeHeartbeat = false
	return
}

func (pc *ParseCommand) ConfigFile() string {
	// `sonalyze parse` accepts no config file
	return ""
}

func (pc *ParseCommand) Perform(
	out io.Writer,
	_ *config.ClusterConfig,
	_ *sonarlog.LogStore,
	samples sonarlog.SampleStream,
	_ *hostglob.HostGlobber,
	recordFilter func(*sonarlog.Sample) bool,
) error {
	var mergedSamples []*sonarlog.SampleStream
	switch {
	case pc.Clean:
		streams := sonarlog.PostprocessLog(samples, recordFilter, nil)
		mergedSamples = maps.Values(streams)

	case pc.MergeByJob:
		bounds := sonarlog.ComputeTimeBounds(samples)
		streams := sonarlog.PostprocessLog(samples, recordFilter, nil)
		mergedSamples, _ = sonarlog.MergeByJob(streams, bounds)

	case pc.MergeByHostAndJob:
		streams := sonarlog.PostprocessLog(samples, recordFilter, nil)
		mergedSamples = sonarlog.MergeByHostAndJob(streams)

	default:
		samples = sonarlog.ApplyFilter(samples, recordFilter)
	}

	if mergedSamples != nil {
		// All elements that are part of the InputStreamKey must be part of the sort key here.
		sort.Stable(sonarlog.HostTimeJobCmdSortableSampleStreams(mergedSamples))
		for _, stream := range mergedSamples {
			fmt.Fprintln(out, "*")
			FormatData(
				out,
				pc.printFields,
				parseFormatters,
				pc.printOpts,
				*stream,
				parseCtx(pc.printOpts.NoDefaults),
			)
		}
	} else {
		FormatData(
			out,
			pc.printFields,
			parseFormatters,
			pc.printOpts,
			samples,
			parseCtx(pc.printOpts.NoDefaults),
		)
	}
	return nil
}

func (pc *ParseCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(pc.Fmt, parseHelp, parseFormatters, parseAliases, parseDefaultFields)
}

const parseHelp = `
parse
  Read raw Sonar data and present it in whole or part.  Default output format
  is 'csv'
`

const parseDefaultFields = "job,user,cmd"

var parseAliases = map[string][]string{
	// TODO: IMPROVEME: Roundtripping is actually version-dependent, but this set of fields is
	// compatible with the Rust version.
	"roundtrip": []string{
		"v",
		"time",
		"host",
		"cores",
		"user",
		"job",
		"pid",
		"cmd",
		"cpu%",
		"cpukib",
		"gpus",
		"gpu%",
		"gpumem%",
		"gpukib",
		"gpufail",
		"cputime_sec",
		"rolledup",
	},
	"all": []string{
		"version",
		"localtime",
		"host",
		"cores",
		"memtotal",
		"user",
		"pid",
		"job",
		"cmd",
		"cpu_pct",
		"mem_gb",
		"res_gb",
		"gpus",
		"gpu_pct",
		"gpumem_pct",
		"gpumem_gb",
		"gpu_status",
		"cputime_sec",
		"rolledup",
		"cpu_util_pct",
	},
}

type parseCtx bool

// TODO: IMPROVEME: The defaulted fields here follow the Rust code.  We'll keep it this way until we
// switch over.  Then we can maybe default more fields.
var parseFormatters = map[string]Formatter[*sonarlog.Sample, parseCtx]{
	"version": {
		func(d *sonarlog.Sample, _ parseCtx) string {
			return d.Version.String()
		},
		"Semver string (MAJOR.MINOR.BUGFIX)",
	},
	"localtime": {
		func(d *sonarlog.Sample, _ parseCtx) string {
			// TODO: IMPROVEME: The use of utc here is a bug that comes from the Rust code.
			return common.FormatYyyyMmDdHhMmUtc(d.Timestamp)
		},
		"Timestamp (yyyy-mm-dd hh:mm)",
	},
	"time": {
		func(d *sonarlog.Sample, _ parseCtx) string {
			return time.Unix(d.Timestamp, 0).UTC().Format(time.RFC3339)
		},
		"Timestamp (ISO date with seconds)",
	},
	"host": {
		func(d *sonarlog.Sample, _ parseCtx) string {
			return d.Host.String()
		},
		"Host name (FQDN)",
	},
	"cores": {
		func(d *sonarlog.Sample, _ parseCtx) string {
			return fmt.Sprint(d.Cores)
		},
		"Total number of cores (including hyperthreads)",
	},
	"memtotal": {
		func(d *sonarlog.Sample, _ parseCtx) string {
			return fmt.Sprint(d.MemtotalKib / (1024 * 1024))
		},
		"Installed main memory (GiB)",
	},
	"user": {
		func(d *sonarlog.Sample, _ parseCtx) string {
			return d.User.String()
		},
		"Username of process owner",
	},
	"pid": {
		func(d *sonarlog.Sample, nodefaults parseCtx) string {
			if bool(nodefaults) && d.Pid == 0 {
				return "*skip*"
			}
			return fmt.Sprint(d.Pid)
		},
		"Process ID",
	},
	"job": {
		func(d *sonarlog.Sample, _ parseCtx) string {
			return fmt.Sprint(d.Job)
		},
		"Job ID",
	},
	"cmd": {
		func(d *sonarlog.Sample, _ parseCtx) string {
			return d.Cmd.String()
		},
		"Command name",
	},
	"cpu_pct": {
		func(d *sonarlog.Sample, _ parseCtx) string {
			return fmt.Sprint(d.CpuPct)
		},
		"cpu% reading (CONSULT DOCUMENTATION)",
	},
	"mem_gb": {
		func(d *sonarlog.Sample, _ parseCtx) string {
			return fmt.Sprint(d.CpuKib / (1024 * 1024))
		},
		"Virtual memory reading (GiB)",
	},
	"res_gb": {
		func(d *sonarlog.Sample, _ parseCtx) string {
			return fmt.Sprint(d.RssAnonKib / (1024 * 1024))
		},
		"RssAnon reading (GiB)",
	},
	"cpukib": {
		func(d *sonarlog.Sample, _ parseCtx) string {
			return fmt.Sprint(d.CpuKib)
		},
		"Virtual memory reading (KiB)",
	},
	"gpus": {
		func(d *sonarlog.Sample, nodefaults parseCtx) string {
			if bool(nodefaults) && d.Gpus.IsEmpty() {
				return "*skip*"
			}
			return d.Gpus.String()
		},
		"GPU set (`none`,`unknown`,list)",
	},
	"gpu_pct": {
		func(d *sonarlog.Sample, nodefaults parseCtx) string {
			if bool(nodefaults) && d.GpuPct == 0 {
				return "*skip*"
			}
			return fmt.Sprint(d.GpuPct)
		},
		"GPU utilization reading",
	},
	"gpumem_pct": {
		func(d *sonarlog.Sample, nodefaults parseCtx) string {
			if bool(nodefaults) && d.GpuMemPct == 0 {
				return "*skip*"
			}
			return fmt.Sprint(d.GpuMemPct)
		},
		"GPU memory percentage reading",
	},
	"gpumem_gb": {
		func(d *sonarlog.Sample, nodefaults parseCtx) string {
			if bool(nodefaults) && d.GpuKib == 0 {
				return "*skip*"
			}
			return fmt.Sprint(d.GpuKib / (1024 * 1024))
		},
		"GPU memory utilization reading (GiB)",
	},
	"gpukib": {
		func(d *sonarlog.Sample, nodefaults parseCtx) string {
			if bool(nodefaults) && d.GpuKib == 0 {
				return "*skip*"
			}
			return fmt.Sprint(d.GpuKib)
		},
		"GPU memory utilization reading (KiB)",
	},
	"gpu_status": {
		func(d *sonarlog.Sample, nodefaults parseCtx) string {
			if bool(nodefaults) && d.GpuFail == 0 {
				return "*skip*"
			}
			return fmt.Sprint(d.GpuFail)
		},
		"GPU status flag (0=ok, 1=error state)",
	},
	"cputime_sec": {
		func(d *sonarlog.Sample, nodefaults parseCtx) string {
			if bool(nodefaults) && d.CpuTimeSec == 0 {
				return "*skip*"
			}
			return fmt.Sprint(d.CpuTimeSec)
		},
		"CPU time since last reading (seconds, CONSULT DOCUMENTATION)",
	},
	"rolledup": {
		func(d *sonarlog.Sample, nodefaults parseCtx) string {
			if bool(nodefaults) && d.Rolledup == 0 {
				return "*skip*"
			}
			return fmt.Sprint(d.Rolledup)
		},
		"Number of rolled-up processes, minus 1",
	},
	"cpu_util_pct": {
		func(d *sonarlog.Sample, nodefaults parseCtx) string {
			if bool(nodefaults) && d.CpuUtilPct == 0 {
				return "*skip*"
			}
			return fmt.Sprint(d.CpuUtilPct)
		},
		"CPU utilization since last reading (percent, CONSULT DOCUMENTATION)",
	},
}

func init() {
	// These are needed for true roundtripping but they can't be defined as aliases because the
	// field names would be the underlying names, which is not what we want.  This way it's
	// compatible with the Rust code.
	parseFormatters["v"] = parseFormatters["version"]
	parseFormatters["cpu%"] = parseFormatters["cpu_pct"]
	parseFormatters["gpu%"] = parseFormatters["gpu_pct"]
	parseFormatters["gpumem%"] = parseFormatters["gpumem_pct"]
	parseFormatters["gpufail"] = parseFormatters["gpu_status"]
}
