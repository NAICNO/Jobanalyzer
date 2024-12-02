// DO NOT EDIT.  Generated from parse.go by generate-table

package parse

import (
	"go-utils/gpuset"
	. "sonalyze/common"
	"sonalyze/sonarlog"
	. "sonalyze/table"
)

type GpuSet = gpuset.GpuSet
type float = float32

// MT: Constant after initialization; immutable
var parseFormatters = map[string]Formatter[sonarlog.Sample]{
	"Version": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatUstr(Ustr(d.Version), ctx)
		},
		Help: "Semver string (MAJOR.MINOR.BUGFIX)",
	},
	"Timestamp": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatDateTimeValue(DateTimeValue(d.Timestamp), ctx)
		},
		Help: "Timestamp (yyyy-mm-dd hh:mm)",
	},
	"time": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatIsoDateTimeValue(IsoDateTimeValue(d.Timestamp), ctx)
		},
		Help: "Timestamp (ISO date with seconds)",
	},
	"Host": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatUstr(Ustr(d.Host), ctx)
		},
		Help: "Host name (FQDN)",
	},
	"Cores": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatInt(int(d.Cores), ctx)
		},
		Help: "Total number of cores (including hyperthreads)",
	},
	"MemtotalKB": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatInt(int(d.MemtotalKB), ctx)
		},
		Help: "Installed main memory",
	},
	"memtotal": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatIntDiv1M(IntDiv1M(d.MemtotalKB), ctx)
		},
		Help: "Installed main memory (GB)",
	},
	"User": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatUstr(Ustr(d.User), ctx)
		},
		Help: "Username of process owner",
	},
	"Pid": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatInt(int(d.Pid), ctx)
		},
		Help: "Process ID",
	},
	"Ppid": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatInt(int(d.Ppid), ctx)
		},
		Help: "Process parent ID",
	},
	"Job": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatInt(int(d.Job), ctx)
		},
		Help: "Job ID",
	},
	"Cmd": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatUstr(Ustr(d.Cmd), ctx)
		},
		Help: "Command name",
	},
	"CpuPct": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatFloat32(float(d.CpuPct), ctx)
		},
		Help: "cpu% reading (CONSULT DOCUMENTATION)",
	},
	"CpuKB": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatInt(int(d.CpuKB), ctx)
		},
		Help: "Virtual memory reading",
	},
	"mem_gb": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatIntDiv1M(IntDiv1M(d.CpuKB), ctx)
		},
		Help: "Virtual memory reading",
	},
	"RssAnonKB": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatInt(int(d.RssAnonKB), ctx)
		},
		Help: "RssAnon reading",
	},
	"res_gb": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatIntDiv1M(IntDiv1M(d.RssAnonKB), ctx)
		},
		Help: "RssAnon reading",
	},
	"Gpus": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatGpuSet(GpuSet(d.Gpus), ctx)
		},
		Help: "GPU set (`none`,`unknown`,list)",
	},
	"GpuPct": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatFloat32(float(d.GpuPct), ctx)
		},
		Help: "GPU utilization reading",
	},
	"GpuMemPct": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatFloat32(float(d.GpuMemPct), ctx)
		},
		Help: "GPU memory percentage reading",
	},
	"GpuKB": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatInt(int(d.GpuKB), ctx)
		},
		Help: "GPU memory utilization reading",
	},
	"gpumem_gb": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatIntDiv1M(IntDiv1M(d.GpuKB), ctx)
		},
		Help: "GPU memory utilization reading",
	},
	"GpuFail": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatInt(int(d.GpuFail), ctx)
		},
		Help: "GPU status flag (0=ok, 1=error state)",
	},
	"CpuTimeSec": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatInt(int(d.CpuTimeSec), ctx)
		},
		Help: "CPU time since last reading (seconds, CONSULT DOCUMENTATION)",
	},
	"Rolledup": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatInt(int(d.Rolledup), ctx)
		},
		Help: "Number of rolled-up processes, minus 1",
	},
	"Flags": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatInt(int(d.Flags), ctx)
		},
		Help: "Bit vector of flags, UTSL",
	},
	"CpuUtilPct": {
		Fmt: func(d sonarlog.Sample, ctx PrintMods) string {
			return FormatFloat32(float(d.CpuUtilPct), ctx)
		},
		Help: "CPU utilization since last reading (percent, CONSULT DOCUMENTATION)",
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
