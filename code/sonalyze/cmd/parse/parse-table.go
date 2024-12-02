// DO NOT EDIT.  Generated from parse.go by generate-table

package parse

import (
	"go-utils/gpuset"
	. "sonalyze/common"
	"sonalyze/sonarlog"
	. "sonalyze/table"
)

import (
	"fmt"
	"io"
)

var (
	_ fmt.Formatter
	_ = io.SeekStart
)

// MT: Constant after initialization; immutable
var parseFormatters = map[string]Formatter[sonarlog.Sample]{
	"Version": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatUstr(Ustr(d.Version), ctx)
		},
		Help: "(string) Semver string (MAJOR.MINOR.BUGFIX)",
	},
	"Timestamp": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatDateTimeValue(DateTimeValue(d.Timestamp), ctx)
		},
		Help: "(DateTimeValue) Timestamp of record ",
	},
	"time": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatIsoDateTimeValue(IsoDateTimeValue(d.Timestamp), ctx)
		},
		Help: "(IsoDateTimeValue) Timestamp of record",
	},
	"Host": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatUstr(Ustr(d.Host), ctx)
		},
		Help: "(string) Host name (FQDN)",
	},
	"Cores": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatInt(int(d.Cores), ctx)
		},
		Help: "(int) Total number of cores (including hyperthreads)",
	},
	"MemtotalKB": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatInt(int(d.MemtotalKB), ctx)
		},
		Help: "(int) Installed main memory",
	},
	"memtotal": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatIntDiv1M(IntDiv1M(d.MemtotalKB), ctx)
		},
		Help: "(int) Installed main memory (GB)",
	},
	"User": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatUstr(Ustr(d.User), ctx)
		},
		Help: "(string) Username of process owner",
	},
	"Pid": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatInt(int(d.Pid), ctx)
		},
		Help: "(int) Process ID",
	},
	"Ppid": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatInt(int(d.Ppid), ctx)
		},
		Help: "(int) Process parent ID",
	},
	"Job": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatInt(int(d.Job), ctx)
		},
		Help: "(int) Job ID",
	},
	"Cmd": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatUstr(Ustr(d.Cmd), ctx)
		},
		Help: "(string) Command name",
	},
	"CpuPct": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatFloat32(float32(d.CpuPct), ctx)
		},
		Help: "(float) cpu% reading (CONSULT DOCUMENTATION)",
	},
	"CpuKB": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatInt(int(d.CpuKB), ctx)
		},
		Help: "(int) Virtual memory reading",
	},
	"mem_gb": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatIntDiv1M(IntDiv1M(d.CpuKB), ctx)
		},
		Help: "(int) Virtual memory reading",
	},
	"RssAnonKB": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatInt(int(d.RssAnonKB), ctx)
		},
		Help: "(int) RssAnon reading",
	},
	"res_gb": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatIntDiv1M(IntDiv1M(d.RssAnonKB), ctx)
		},
		Help: "(int) RssAnon reading",
	},
	"Gpus": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatGpuSet(gpuset.GpuSet(d.Gpus), ctx)
		},
		Help: "(GpuSet) GPU set (`none`,`unknown`,list)",
	},
	"GpuPct": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatFloat32(float32(d.GpuPct), ctx)
		},
		Help: "(float) GPU utilization reading",
	},
	"GpuMemPct": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatFloat32(float32(d.GpuMemPct), ctx)
		},
		Help: "(float) GPU memory percentage reading",
	},
	"GpuKB": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatInt(int(d.GpuKB), ctx)
		},
		Help: "(int) GPU memory utilization reading",
	},
	"gpumem_gb": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatIntDiv1M(IntDiv1M(d.GpuKB), ctx)
		},
		Help: "(int) GPU memory utilization reading",
	},
	"GpuFail": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatInt(int(d.GpuFail), ctx)
		},
		Help: "(int) GPU status flag (0=ok, 1=error state)",
	},
	"CpuTimeSec": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatInt(int(d.CpuTimeSec), ctx)
		},
		Help: "(int) CPU time since last reading (seconds, CONSULT DOCUMENTATION)",
	},
	"Rolledup": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatInt(int(d.Rolledup), ctx)
		},
		Help: "(int) Number of rolled-up processes, minus 1",
	},
	"Flags": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatInt(int(d.Flags), ctx)
		},
		Help: "(int) Bit vector of flags, UTSL",
	},
	"CpuUtilPct": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatFloat32(float32(d.CpuUtilPct), ctx)
		},
		Help: "(float) CPU utilization since last reading (percent, CONSULT DOCUMENTATION)",
	},
}

func init() {
	DefAlias(parseFormatters, "Version", "version")
	DefAlias(parseFormatters, "Version", "v")
	DefAlias(parseFormatters, "Timestamp", "localtime")
	DefAlias(parseFormatters, "Host", "host")
	DefAlias(parseFormatters, "Cores", "cores")
	DefAlias(parseFormatters, "User", "user")
	DefAlias(parseFormatters, "Pid", "pid")
	DefAlias(parseFormatters, "Ppid", "ppid")
	DefAlias(parseFormatters, "Job", "job")
	DefAlias(parseFormatters, "Cmd", "cmd")
	DefAlias(parseFormatters, "CpuPct", "cpu_pct")
	DefAlias(parseFormatters, "CpuPct", "cpu%")
	DefAlias(parseFormatters, "CpuKB", "cpukib")
	DefAlias(parseFormatters, "Gpus", "gpus")
	DefAlias(parseFormatters, "GpuPct", "gpu_pct")
	DefAlias(parseFormatters, "GpuPct", "gpu%")
	DefAlias(parseFormatters, "GpuMemPct", "gpumem_pct")
	DefAlias(parseFormatters, "GpuMemPct", "gpumem%")
	DefAlias(parseFormatters, "GpuKB", "gpukib")
	DefAlias(parseFormatters, "GpuFail", "gpu_status")
	DefAlias(parseFormatters, "GpuFail", "gpufail")
	DefAlias(parseFormatters, "CpuTimeSec", "cputime_sec")
	DefAlias(parseFormatters, "Rolledup", "rolledup")
	DefAlias(parseFormatters, "CpuUtilPct", "cpu_util_pct")
}

func (c *ParseCommand) Summary(out io.Writer) {
	fmt.Fprint(out, `Export sample data in various formats, after optional preprocessing.

This facility is mostly for debugging and experimentation, as the data
volume is typically significant and the data are not necessarily
postprocessed in a way useful to the consumer.

The -merge and -clean options perform some postprocessing, but you need to
know what you're looking at to find these useful.
`)
}

const parseHelp = `
parse
  Read raw Sonar data and present it in whole or part.  Default output format
  is 'csv'.
`

func (c *ParseCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(c.Fmt, parseHelp, parseFormatters, parseAliases, parseDefaultFields)
}

// MT: Constant after initialization; immutable
var parseAliases = map[string][]string{
	"default":   []string{"job", "user", "cmd"},
	"Default":   []string{"Job", "User", "Cmd"},
	"all":       []string{"version", "localtime", "host", "cores", "memtotal", "user", "pid", "job", "cmd", "cpu_pct", "mem_gb", "res_gb", "gpus", "gpu_pct", "gpumem_pct", "gpumem_gb", "gpu_status", "cputime_sec", "rolledup", "cpu_util_pct"},
	"All":       []string{"Version", "Timestamp", "Host", "Cores", "MemtotalKB", "User", "Pid", "Ppid", "Job", "Cmd", "CpuPct", "CpuKB", "RssAnonKB", "Gpus", "GpuPct", "GpuMemPct", "GpuKB", "GpuFail", "CpuTimeSec", "Rolledup", "CpuUtilPct"},
	"roundtrip": []string{"v", "time", "host", "cores", "user", "job", "pid", "cmd", "cpu%", "cpukib", "gpus", "gpu%", "gpumem%", "gpukib", "gpufail", "cputime_sec", "rolledup"},
}

const parseDefaultFields = "default"
