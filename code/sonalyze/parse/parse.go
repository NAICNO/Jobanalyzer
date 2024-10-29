package parse

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"go-utils/config"
	"go-utils/hostglob"
	"go-utils/maps"
	uslices "go-utils/slices"
	. "sonalyze/command"
	"sonalyze/db"
	"sonalyze/sonarlog"
)

type ParseCommand struct /* implements SampleAnalysisCommand */ {
	SharedArgs
	FormatArgs

	MergeByHostAndJob bool
	MergeByJob        bool
	Clean             bool
}

var _ SampleAnalysisCommand = (*ParseCommand)(nil)

func (_ *ParseCommand) Summary() []string {
	return []string{
		"Export sample data in various formats, after optional preprocessing.",
	}
}

func (pc *ParseCommand) Add(fs *flag.FlagSet) {
	pc.SharedArgs.Add(fs)
	pc.FormatArgs.Add(fs)
	fs.BoolVar(&pc.MergeByHostAndJob, "merge-by-host-and-job", false,
		"Merge streams that have the same host and job ID")
	fs.BoolVar(&pc.MergeByJob, "merge-by-job", false,
		"Merge streams that have the same job ID, across hosts")
	fs.BoolVar(&pc.Clean, "clean", false,
		"Clean the job but perform no merging")
}

func (pc *ParseCommand) ReifyForRemote(x *Reifier) error {
	e1 := errors.Join(
		pc.SharedArgs.ReifyForRemote(x),
		pc.FormatArgs.ReifyForRemote(x),
	)
	x.Bool("merge-by-host-and-job", pc.MergeByHostAndJob)
	x.Bool("merge-by-job", pc.MergeByJob)
	x.Bool("clean", pc.Clean)
	return e1
}

func (pc *ParseCommand) Validate() error {
	return errors.Join(
		pc.SharedArgs.Validate(),
		ValidateFormatArgs(
			&pc.FormatArgs, parseDefaultFields, parseFormatters, parseAliases, DefaultCsv),
	)
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

func (pc *ParseCommand) NeedsBounds() bool {
	return pc.MergeByJob
}

func (pc *ParseCommand) Perform(
	out io.Writer,
	_ *config.ClusterConfig,
	cluster db.SampleCluster,
	streams sonarlog.InputStreamSet,
	bounds sonarlog.Timebounds, // for pc.MergeByJob only
	hostGlobber *hostglob.HostGlobber,
	recordFilter *db.SampleFilter,
) error {
	var mergedSamples []*sonarlog.SampleStream
	var samples sonarlog.SampleStream
	switch {
	default:
		// Reread the data, bypassing all postprocessing, to get the expected raw values.  If it's
		// expensive then so be it - this is special-case code usually used for limited testing, not
		// something you'd use for analysis.
		recordBlobs, _, err := cluster.ReadSamples(pc.FromDate, pc.ToDate, hostGlobber, pc.Verbose)
		if err != nil {
			return err
		}

		// Simulate the normal pipeline, the recordFilter application is expected by the user.
		mapped := make([]sonarlog.Sample, 0)
		for _, records := range recordBlobs {
			mapped = append(mapped, uslices.FilterMap(
				records,
				db.InstantiateSampleFilter(recordFilter),
				func(r *db.Sample) sonarlog.Sample {
					return sonarlog.Sample{Sample: r}
				},
			)...)
		}
		samples = sonarlog.SampleStream(mapped)

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
				pc.PrintFields,
				parseFormatters,
				pc.PrintOpts,
				uslices.Map(*stream, func(x sonarlog.Sample) any { return x }),
				ComputePrintMods(pc.PrintOpts),
			)
		}
	} else {
		FormatData(
			out,
			pc.PrintFields,
			parseFormatters,
			pc.PrintOpts,
			uslices.Map(samples, func(x sonarlog.Sample) any { return x }),
			ComputePrintMods(pc.PrintOpts),
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

const v0ParseDefaultFields = "job,user,cmd"
const v1ParseDefaultFields = "Job,User,Cmd"
const parseDefaultFields = v0ParseDefaultFields

const v0ParseAllFields = "version,localtime,host,cores,memtotal,user,pid,job,cmd,cpu_pct,mem_gb," +
	"res_gb,gpus,gpu_pct,gpumem_pct,gpumem_gb,gpu_status,cputime_sec,rolledup,cpu_util_pct"
const v1ParseAllFields = "Version,Timestamp,Host,Cores,MemtotalKB,User,Pid,Ppid,Job,Cmd,CpuPct," +
	"CpuKB,RssAnonKB,Gpus,GpuPct,GpuMemPct,GpuKB,GpuFail,CpuTimeSec,Rolledup,CpuUtilPct"
const parseAllFields = v0ParseAllFields

// MT: Constant after initialization; immutable
var parseAliases = map[string][]string{
	"default":   strings.Split(parseDefaultFields, ","),
	"v0default": strings.Split(v0ParseDefaultFields, ","),
	"v1default": strings.Split(v1ParseDefaultFields, ","),
	"all":       strings.Split(parseAllFields, ","),
	"v0all":     strings.Split(v0ParseAllFields, ","),
	"v1all":     strings.Split(v1ParseAllFields, ","),
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
}

type SFS = SimpleFormatSpec
type AFS = SimpleFormatSpecWithAttr
type ZFA = SynthesizedFormatSpecWithAttr

// TODO: IMPROVEME: The defaulted fields here follow the Rust code.  Now that it's trivial to do so,
// we should consider adding more.
//
// TODO: IMPROVEME: The use of utc for "localtime" is a bug that comes from the Rust code.

var parseFormatters = ReflectFormattersFromMap(
	reflect.TypeOf((*sonarlog.Sample)(nil)).Elem(),
	map[string]any{
		"Version":    SFS{"Semver string (MAJOR.MINOR.BUGFIX)", "version"},
		"Timestamp":  AFS{"Timestamp (yyyy-mm-dd hh:mm)", "localtime", FmtDateTimeValue},
		"time":       ZFA{"Timestamp (ISO date with seconds)", "Timestamp", FmtIsoDateTimeValue},
		"Host":       SFS{"Host name (FQDN)", "host"},
		"Cores":      SFS{"Total number of cores (including hyperthreads)", "cores"},
		"MemtotalKB": SFS{"Installed main memory", ""},
		"memtotal":   ZFA{"Installed main memory (GB)", "MemtotalKB", FmtDivideBy1M},
		"User":       SFS{"Username of process owner", "user"},
		"Pid":        AFS{"Process ID", "pid", FmtDefaultable},
		"Ppid":       AFS{"Process parent ID", "ppid", FmtDefaultable},
		"Job":        SFS{"Job ID", "job"},
		"Cmd":        SFS{"Command name", "cmd"},
		"CpuPct":     SFS{"cpu% reading (CONSULT DOCUMENTATION)", "cpu_pct"},
		"CpuKB":      SFS{"Virtual memory reading", "cpukib"},
		"mem_gb":     ZFA{"Virtual memory reading", "CpuKB", FmtDivideBy1M},
		"RssAnonKB":  SFS{"RssAnon reading", ""},
		"res_gb":     ZFA{"RssAnon reading", "RssAnonKB", FmtDivideBy1M},
		"Gpus":       AFS{"GPU set (`none`,`unknown`,list)", "gpus", FmtDefaultable},
		"GpuPct":     AFS{"GPU utilization reading", "gpu_pct", FmtDefaultable},
		"GpuMemPct":  AFS{"GPU memory percentage reading", "gpumem_pct", FmtDefaultable},
		"GpuKB":      AFS{"GPU memory utilization reading", "gpukib", FmtDefaultable},
		"gpumem_gb": ZFA{"GPU memory utilization reading", "GpuKB",
			FmtDivideBy1M | FmtDefaultable},
		"GpuFail": AFS{"GPU status flag (0=ok, 1=error state)", "gpu_status", FmtDefaultable},
		"CpuTimeSec": AFS{"CPU time since last reading (seconds, CONSULT DOCUMENTATION)",
			"cputime_sec", FmtDefaultable},
		"Rolledup": AFS{"Number of rolled-up processes, minus 1", "rolledup", FmtDefaultable},
		"Flags":    SFS{"Bit vector of flags, UTSL", ""},
		"CpuUtilPct": AFS{"CPU utilization since last reading (percent, CONSULT DOCUMENTATION)",
			"cpu_util_pct", FmtDefaultable},
	},
)

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
