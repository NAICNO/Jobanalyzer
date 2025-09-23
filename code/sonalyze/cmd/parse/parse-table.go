// DO NOT EDIT.  Generated from parse.go by generate-table

package parse

import "sonalyze/data/sample"

import (
	"cmp"
	"fmt"
	"go-utils/gpuset"
	"io"
	. "sonalyze/common"
	. "sonalyze/table"
)

var (
	_ = cmp.Compare(0, 0)
	_ fmt.Formatter
	_ = io.SeekStart
	_ = UstrEmpty
	_ gpuset.GpuSet
)

// MT: Constant after initialization; immutable
var parseFormatters = map[string]Formatter[sample.Sample]{
	"Version": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatUstr((d.Version), ctx)
		},
		Help: "(string) Semver string (MAJOR.MINOR.BUGFIX)",
	},
	"Timestamp": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatDateTimeValue((d.Timestamp), ctx)
		},
		Help: "(DateTimeValue) Timestamp of record ",
	},
	"time": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatIsoDateTimeValue((d.Timestamp), ctx)
		},
		Help: "(IsoDateTimeValue) Timestamp of record",
	},
	"Hostname": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatUstr((d.Hostname), ctx)
		},
		Help: "(string) Host name (FQDN)",
	},
	"Cores": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatUint32((d.Cores), ctx)
		},
		Help: "(uint32) Total number of cores (including hyperthreads)",
	},
	"Threads": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatUint32((d.Threads), ctx)
		},
		Help: "(uint32) Number of threads active",
	},
	"MemtotalKB": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatUint64((d.MemtotalKB), ctx)
		},
		Help: "(uint64) Installed main memory",
	},
	"memtotal": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatU64Div1M((d.MemtotalKB), ctx)
		},
		Help: "(int) Installed main memory (GB)",
	},
	"User": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatUstr((d.User), ctx)
		},
		Help: "(string) Username of process owner",
	},
	"Pid": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatUint32((d.Pid), ctx)
		},
		Help: "(uint32) Process ID",
	},
	"Ppid": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatUint32((d.Ppid), ctx)
		},
		Help: "(uint32) Process parent ID",
	},
	"Job": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatUint32((d.Job), ctx)
		},
		Help: "(uint32) Job ID",
	},
	"Cmd": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatUstr((d.Cmd), ctx)
		},
		Help: "(string) Command name",
	},
	"CpuPct": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatFloat32((d.CpuPct), ctx)
		},
		Help: "(float32) cpu% reading (CONSULT DOCUMENTATION)",
	},
	"CpuKB": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatUint64((d.CpuKB), ctx)
		},
		Help: "(uint64) Virtual memory reading",
	},
	"mem_gb": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatU64Div1M((d.CpuKB), ctx)
		},
		Help: "(int) Virtual memory reading",
	},
	"RssAnonKB": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatUint64((d.RssAnonKB), ctx)
		},
		Help: "(uint64) RssAnon reading",
	},
	"res_gb": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatU64Div1M((d.RssAnonKB), ctx)
		},
		Help: "(int) RssAnon reading",
	},
	"Gpus": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatGpuSet((d.Gpus), ctx)
		},
		Help: "(GpuSet) GPU set (`none`,`unknown`,list)",
	},
	"GpuPct": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatFloat32((d.GpuPct), ctx)
		},
		Help: "(float32) GPU utilization reading",
	},
	"GpuMemPct": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatFloat32((d.GpuMemPct), ctx)
		},
		Help: "(float32) GPU memory percentage reading",
	},
	"GpuKB": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatUint64((d.GpuKB), ctx)
		},
		Help: "(uint64) GPU memory utilization reading",
	},
	"gpumem_gb": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatU64Div1M((d.GpuKB), ctx)
		},
		Help: "(int) GPU memory utilization reading",
	},
	"GpuFail": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatUint8((d.GpuFail), ctx)
		},
		Help: "(uint8) GPU status flag (0=ok, 1=error state)",
	},
	"CpuTimeSec": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatUint64((d.CpuTimeSec), ctx)
		},
		Help: "(uint64) CPU time since last reading (seconds, CONSULT DOCUMENTATION)",
	},
	"Rolledup": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatUint32((d.Rolledup), ctx)
		},
		Help: "(uint32) Number of rolled-up processes, minus 1",
	},
	"Flags": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatUint8((d.Flags), ctx)
		},
		Help: "(uint8) Bit vector of flags, UTSL",
	},
	"CpuUtilPct": {
		Fmt: func(d sample.Sample, ctx PrintMods) string {
			return FormatFloat32((d.CpuUtilPct), ctx)
		},
		Help: "(float32) CPU utilization since last reading (percent, CONSULT DOCUMENTATION)",
	},
}

func init() {
	DefAlias(parseFormatters, "Version", "version")
	DefAlias(parseFormatters, "Version", "v")
	DefAlias(parseFormatters, "Timestamp", "localtime")
	DefAlias(parseFormatters, "Hostname", "host")
	DefAlias(parseFormatters, "Cores", "cores")
	DefAlias(parseFormatters, "Threads", "threads")
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

// MT: Constant after initialization; immutable
var parsePredicates = map[string]Predicate[sample.Sample]{
	"Version": Predicate[sample.Sample]{
		Convert: CvtString2Ustr,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.Version), v.(Ustr))
		},
	},
	"Timestamp": Predicate[sample.Sample]{
		Convert: CvtString2DateTimeValue,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.Timestamp), v.(DateTimeValue))
		},
	},
	"time": Predicate[sample.Sample]{
		Convert: CvtString2IsoDateTimeValue,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.Timestamp), v.(IsoDateTimeValue))
		},
	},
	"Hostname": Predicate[sample.Sample]{
		Convert: CvtString2Ustr,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.Hostname), v.(Ustr))
		},
	},
	"Cores": Predicate[sample.Sample]{
		Convert: CvtString2Uint32,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.Cores), v.(uint32))
		},
	},
	"Threads": Predicate[sample.Sample]{
		Convert: CvtString2Uint32,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.Threads), v.(uint32))
		},
	},
	"MemtotalKB": Predicate[sample.Sample]{
		Convert: CvtString2Uint64,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.MemtotalKB), v.(uint64))
		},
	},
	"memtotal": Predicate[sample.Sample]{
		Convert: CvtString2Uint64,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.MemtotalKB), v.(U64Div1M))
		},
	},
	"User": Predicate[sample.Sample]{
		Convert: CvtString2Ustr,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.User), v.(Ustr))
		},
	},
	"Pid": Predicate[sample.Sample]{
		Convert: CvtString2Uint32,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.Pid), v.(uint32))
		},
	},
	"Ppid": Predicate[sample.Sample]{
		Convert: CvtString2Uint32,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.Ppid), v.(uint32))
		},
	},
	"Job": Predicate[sample.Sample]{
		Convert: CvtString2Uint32,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.Job), v.(uint32))
		},
	},
	"Cmd": Predicate[sample.Sample]{
		Convert: CvtString2Ustr,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.Cmd), v.(Ustr))
		},
	},
	"CpuPct": Predicate[sample.Sample]{
		Convert: CvtString2Float32,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.CpuPct), v.(float32))
		},
	},
	"CpuKB": Predicate[sample.Sample]{
		Convert: CvtString2Uint64,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.CpuKB), v.(uint64))
		},
	},
	"mem_gb": Predicate[sample.Sample]{
		Convert: CvtString2Uint64,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.CpuKB), v.(U64Div1M))
		},
	},
	"RssAnonKB": Predicate[sample.Sample]{
		Convert: CvtString2Uint64,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.RssAnonKB), v.(uint64))
		},
	},
	"res_gb": Predicate[sample.Sample]{
		Convert: CvtString2Uint64,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.RssAnonKB), v.(U64Div1M))
		},
	},
	"Gpus": Predicate[sample.Sample]{
		Convert: CvtString2GpuSet,
		SetCompare: func(d sample.Sample, v any, op int) bool {
			return SetCompareGpuSets((d.Gpus), v.(gpuset.GpuSet), op)
		},
	},
	"GpuPct": Predicate[sample.Sample]{
		Convert: CvtString2Float32,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.GpuPct), v.(float32))
		},
	},
	"GpuMemPct": Predicate[sample.Sample]{
		Convert: CvtString2Float32,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.GpuMemPct), v.(float32))
		},
	},
	"GpuKB": Predicate[sample.Sample]{
		Convert: CvtString2Uint64,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.GpuKB), v.(uint64))
		},
	},
	"gpumem_gb": Predicate[sample.Sample]{
		Convert: CvtString2Uint64,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.GpuKB), v.(U64Div1M))
		},
	},
	"GpuFail": Predicate[sample.Sample]{
		Convert: CvtString2Uint8,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.GpuFail), v.(uint8))
		},
	},
	"CpuTimeSec": Predicate[sample.Sample]{
		Convert: CvtString2Uint64,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.CpuTimeSec), v.(uint64))
		},
	},
	"Rolledup": Predicate[sample.Sample]{
		Convert: CvtString2Uint32,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.Rolledup), v.(uint32))
		},
	},
	"Flags": Predicate[sample.Sample]{
		Convert: CvtString2Uint8,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.Flags), v.(uint8))
		},
	},
	"CpuUtilPct": Predicate[sample.Sample]{
		Convert: CvtString2Float32,
		Compare: func(d sample.Sample, v any) int {
			return cmp.Compare((d.CpuUtilPct), v.(float32))
		},
	},
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
	"all":       []string{"version", "localtime", "host", "cores", "threads", "memtotal", "user", "pid", "job", "cmd", "cpu_pct", "mem_gb", "res_gb", "gpus", "gpu_pct", "gpumem_pct", "gpumem_gb", "gpu_status", "cputime_sec", "rolledup", "cpu_util_pct"},
	"All":       []string{"Version", "Timestamp", "Hostname", "Cores", "Threads", "MemtotalKB", "User", "Pid", "Ppid", "Job", "Cmd", "CpuPct", "CpuKB", "RssAnonKB", "Gpus", "GpuPct", "GpuMemPct", "GpuKB", "GpuFail", "CpuTimeSec", "Rolledup", "CpuUtilPct"},
	"roundtrip": []string{"v", "time", "host", "cores", "user", "job", "pid", "cmd", "cpu%", "cpukib", "gpus", "gpu%", "gpumem%", "gpukib", "gpufail", "cputime_sec", "rolledup"},
}

const parseDefaultFields = "default"
