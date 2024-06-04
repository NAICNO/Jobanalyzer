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
	"go-utils/slices"
	. "sonalyze/command"
	. "sonalyze/common"
	"sonalyze/db"
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
	var others map[string]bool
	pc.printFields, others, e2 = ParseFormatSpec(parseDefaultFields, pc.Fmt, parseFormatters, parseAliases)
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
	cluster db.SampleCluster,
	streams sonarlog.InputStreamSet,
	bounds sonarlog.Timebounds,
	hostGlobber *hostglob.HostGlobber,
) error {
	var mergedSamples []*sonarlog.SampleStream
	var samples sonarlog.SampleStream
	switch {
	default:
		// Reread the data, bypassing all postprocessing, to get the expected raw values.  If it's
		// expensive then so be it - this is special-case code usually used for limited testing, not
		// something you'd use for analysis.
		records, _, err := cluster.ReadSamples(pc.FromDate, pc.ToDate, hostGlobber, pc.Verbose)
		if err != nil {
			return err
		}
		samples = sonarlog.SampleStream(slices.Map(
			records,
			func(r *db.Sample) sonarlog.Sample {
				return sonarlog.Sample{S: r}
			},
		))

	case pc.Clean:
		mergedSamples = maps.Values(streams)

	case pc.MergeByJob:
		mergedSamples, _ = sonarlog.MergeByJob(streams, bounds)

	case pc.MergeByHostAndJob:
		mergedSamples = sonarlog.MergeByHostAndJob(streams)
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

// MT: Constant after initialization; immutable
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
// MT: Constant after initialization; immutable
var parseFormatters = map[string]Formatter[sonarlog.Sample, parseCtx]{
	"version": {
		func(d sonarlog.Sample, _ parseCtx) string {
			return d.S.Version.String()
		},
		"Semver string (MAJOR.MINOR.BUGFIX)",
	},
	"localtime": {
		func(d sonarlog.Sample, _ parseCtx) string {
			// TODO: IMPROVEME: The use of utc here is a bug that comes from the Rust code.
			return FormatYyyyMmDdHhMmUtc(d.S.Timestamp)
		},
		"Timestamp (yyyy-mm-dd hh:mm)",
	},
	"time": {
		func(d sonarlog.Sample, _ parseCtx) string {
			return time.Unix(d.S.Timestamp, 0).UTC().Format(time.RFC3339)
		},
		"Timestamp (ISO date with seconds)",
	},
	"host": {
		func(d sonarlog.Sample, _ parseCtx) string {
			return d.S.Host.String()
		},
		"Host name (FQDN)",
	},
	"cores": {
		func(d sonarlog.Sample, _ parseCtx) string {
			return fmt.Sprint(d.S.Cores)
		},
		"Total number of cores (including hyperthreads)",
	},
	"memtotal": {
		func(d sonarlog.Sample, _ parseCtx) string {
			return fmt.Sprint(d.S.MemtotalKib / (1024 * 1024))
		},
		"Installed main memory (GiB)",
	},
	"user": {
		func(d sonarlog.Sample, _ parseCtx) string {
			return d.S.User.String()
		},
		"Username of process owner",
	},
	"pid": {
		func(d sonarlog.Sample, nodefaults parseCtx) string {
			if bool(nodefaults) && d.S.Pid == 0 {
				return "*skip*"
			}
			return fmt.Sprint(d.S.Pid)
		},
		"Process ID",
	},
	"job": {
		func(d sonarlog.Sample, _ parseCtx) string {
			return fmt.Sprint(d.S.Job)
		},
		"Job ID",
	},
	"cmd": {
		func(d sonarlog.Sample, _ parseCtx) string {
			return d.S.Cmd.String()
		},
		"Command name",
	},
	"cpu_pct": {
		func(d sonarlog.Sample, _ parseCtx) string {
			return fmt.Sprint(d.S.CpuPct)
		},
		"cpu% reading (CONSULT DOCUMENTATION)",
	},
	"mem_gb": {
		func(d sonarlog.Sample, _ parseCtx) string {
			return fmt.Sprint(d.S.CpuKib / (1024 * 1024))
		},
		"Virtual memory reading (GiB)",
	},
	"res_gb": {
		func(d sonarlog.Sample, _ parseCtx) string {
			return fmt.Sprint(d.S.RssAnonKib / (1024 * 1024))
		},
		"RssAnon reading (GiB)",
	},
	"cpukib": {
		func(d sonarlog.Sample, _ parseCtx) string {
			return fmt.Sprint(d.S.CpuKib)
		},
		"Virtual memory reading (KiB)",
	},
	"gpus": {
		func(d sonarlog.Sample, nodefaults parseCtx) string {
			if bool(nodefaults) && d.S.Gpus.IsEmpty() {
				return "*skip*"
			}
			return d.S.Gpus.String()
		},
		"GPU set (`none`,`unknown`,list)",
	},
	"gpu_pct": {
		func(d sonarlog.Sample, nodefaults parseCtx) string {
			if bool(nodefaults) && d.S.GpuPct == 0 {
				return "*skip*"
			}
			return fmt.Sprint(d.S.GpuPct)
		},
		"GPU utilization reading",
	},
	"gpumem_pct": {
		func(d sonarlog.Sample, nodefaults parseCtx) string {
			if bool(nodefaults) && d.S.GpuMemPct == 0 {
				return "*skip*"
			}
			return fmt.Sprint(d.S.GpuMemPct)
		},
		"GPU memory percentage reading",
	},
	"gpumem_gb": {
		func(d sonarlog.Sample, nodefaults parseCtx) string {
			if bool(nodefaults) && d.S.GpuKib == 0 {
				return "*skip*"
			}
			return fmt.Sprint(d.S.GpuKib / (1024 * 1024))
		},
		"GPU memory utilization reading (GiB)",
	},
	"gpukib": {
		func(d sonarlog.Sample, nodefaults parseCtx) string {
			if bool(nodefaults) && d.S.GpuKib == 0 {
				return "*skip*"
			}
			return fmt.Sprint(d.S.GpuKib)
		},
		"GPU memory utilization reading (KiB)",
	},
	"gpu_status": {
		func(d sonarlog.Sample, nodefaults parseCtx) string {
			if bool(nodefaults) && d.S.GpuFail == 0 {
				return "*skip*"
			}
			return fmt.Sprint(d.S.GpuFail)
		},
		"GPU status flag (0=ok, 1=error state)",
	},
	"cputime_sec": {
		func(d sonarlog.Sample, nodefaults parseCtx) string {
			if bool(nodefaults) && d.S.CpuTimeSec == 0 {
				return "*skip*"
			}
			return fmt.Sprint(d.S.CpuTimeSec)
		},
		"CPU time since last reading (seconds, CONSULT DOCUMENTATION)",
	},
	"rolledup": {
		func(d sonarlog.Sample, nodefaults parseCtx) string {
			if bool(nodefaults) && d.S.Rolledup == 0 {
				return "*skip*"
			}
			return fmt.Sprint(d.S.Rolledup)
		},
		"Number of rolled-up processes, minus 1",
	},
	"cpu_util_pct": {
		func(d sonarlog.Sample, nodefaults parseCtx) string {
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
