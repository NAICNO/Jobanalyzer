// DO NOT EDIT.  Generated from gpus.go by generate-table

package gpus

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
var gpuFormatters = map[string]Formatter[*ReportLine]{
	"Timestamp": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatDateTimeValue((d.Timestamp), ctx)
		},
		Help: "(DateTimeValue) Timestamp of when the reading was taken",
	},
	"Hostname": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatUstr((d.Hostname), ctx)
		},
		Help: "(string) Name that host is known by on the cluster",
	},
	"Gpu": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatInt((d.Gpu), ctx)
		},
		Help: "(int) Card index on the host",
	},
	"FanPct": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatInt((d.FanPct), ctx)
		},
		Help: "(int) Fan speed in percent of max",
	},
	"PerfMode": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatInt((d.PerfMode), ctx)
		},
		Help: "(int) Numeric performance mode",
	},
	"MemUsedKB": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatInt64((d.MemUsedKB), ctx)
		},
		Help: "(int64) Amount of memory in use",
	},
	"TempC": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatInt((d.TempC), ctx)
		},
		Help: "(int) Card temperature in degrees C",
	},
	"PowerDrawW": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatInt((d.PowerDrawW), ctx)
		},
		Help: "(int) Current power draw in Watts",
	},
	"PowerLimitW": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatInt((d.PowerLimitW), ctx)
		},
		Help: "(int) Current power limit in Watts",
	},
	"CeClockMHz": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatInt((d.CeClockMHz), ctx)
		},
		Help: "(int) Current compute element clock",
	},
	"MemClockMHz": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatInt((d.MemClockMHz), ctx)
		},
		Help: "(int) Current memory clock",
	},
}

// MT: Constant after initialization; immutable
var gpuPredicates = map[string]Predicate[*ReportLine]{
	"Timestamp": Predicate[*ReportLine]{
		Convert: CvtString2DateTimeValue,
		Compare: func(d *ReportLine, v any) int {
			return cmp.Compare((d.Timestamp), v.(DateTimeValue))
		},
	},
	"Hostname": Predicate[*ReportLine]{
		Convert: CvtString2Ustr,
		Compare: func(d *ReportLine, v any) int {
			return cmp.Compare((d.Hostname), v.(Ustr))
		},
	},
	"Gpu": Predicate[*ReportLine]{
		Convert: CvtString2Int,
		Compare: func(d *ReportLine, v any) int {
			return cmp.Compare((d.Gpu), v.(int))
		},
	},
	"FanPct": Predicate[*ReportLine]{
		Convert: CvtString2Int,
		Compare: func(d *ReportLine, v any) int {
			return cmp.Compare((d.FanPct), v.(int))
		},
	},
	"PerfMode": Predicate[*ReportLine]{
		Convert: CvtString2Int,
		Compare: func(d *ReportLine, v any) int {
			return cmp.Compare((d.PerfMode), v.(int))
		},
	},
	"MemUsedKB": Predicate[*ReportLine]{
		Convert: CvtString2Int64,
		Compare: func(d *ReportLine, v any) int {
			return cmp.Compare((d.MemUsedKB), v.(int64))
		},
	},
	"TempC": Predicate[*ReportLine]{
		Convert: CvtString2Int,
		Compare: func(d *ReportLine, v any) int {
			return cmp.Compare((d.TempC), v.(int))
		},
	},
	"PowerDrawW": Predicate[*ReportLine]{
		Convert: CvtString2Int,
		Compare: func(d *ReportLine, v any) int {
			return cmp.Compare((d.PowerDrawW), v.(int))
		},
	},
	"PowerLimitW": Predicate[*ReportLine]{
		Convert: CvtString2Int,
		Compare: func(d *ReportLine, v any) int {
			return cmp.Compare((d.PowerLimitW), v.(int))
		},
	},
	"CeClockMHz": Predicate[*ReportLine]{
		Convert: CvtString2Int,
		Compare: func(d *ReportLine, v any) int {
			return cmp.Compare((d.CeClockMHz), v.(int))
		},
	},
	"MemClockMHz": Predicate[*ReportLine]{
		Convert: CvtString2Int,
		Compare: func(d *ReportLine, v any) int {
			return cmp.Compare((d.MemClockMHz), v.(int))
		},
	},
}

func (c *GpuCommand) Summary(out io.Writer) {
	fmt.Fprint(out, `Experimental: Print per-gpu data across time for one or more cards on one or more nodes.
`)
}

const gpuHelp = `
gpu
  Extract information about individual gpus on the cluster from sample data.  The default
  format is 'fixed'.
`

func (c *GpuCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(c.Fmt, gpuHelp, gpuFormatters, gpuAliases, gpuDefaultFields)
}

// MT: Constant after initialization; immutable
var gpuAliases = map[string][]string{
	"default": []string{"Hostname", "Gpu", "Timestamp", "MemUsedKB", "PowerDrawW"},
	"Default": []string{"Hostname", "Gpu", "Timestamp", "MemUsedKB", "PowerDrawW"},
	"All":     []string{"Timestamp", "Hostname", "Gpu", "FanPct", "PerfMode", "MemUsedKB", "TempC", "PowerDrawW", "PowerLimitW", "CeClockMHz", "MemClockMHz"},
}

const gpuDefaultFields = "default"
