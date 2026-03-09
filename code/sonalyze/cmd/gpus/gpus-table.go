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
		Xtract: func(d *ReportLine) any {
			return d.Timestamp
		},
		Help: "(DateTimeValue) Timestamp of when the reading was taken",
	},
	"Hostname": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatUstr((d.Hostname), ctx)
		},
		Xtract: func(d *ReportLine) any {
			return d.Hostname
		},
		Help: "(string) Name that host is known by on the cluster",
	},
	"Index": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatUint64((d.Index), ctx)
		},
		Xtract: func(d *ReportLine) any {
			return d.Index
		},
		Help: "(uint64) Card index on the host",
	},
	"Fan": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatUint64((d.Fan), ctx)
		},
		Xtract: func(d *ReportLine) any {
			return d.Fan
		},
		Help: "(uint64) Fan speed in percent of max",
	},
	"Memory": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatUint64((d.Memory), ctx)
		},
		Xtract: func(d *ReportLine) any {
			return d.Memory
		},
		Help: "(uint64) Amount of memory in use",
	},
	"Temperature": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatInt64((d.Temperature), ctx)
		},
		Xtract: func(d *ReportLine) any {
			return d.Temperature
		},
		Help: "(int64) Card temperature in degrees C",
	},
	"Power": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatUint64((d.Power), ctx)
		},
		Xtract: func(d *ReportLine) any {
			return d.Power
		},
		Help: "(uint64) Current power draw in Watts",
	},
	"PowerLimit": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatUint64((d.PowerLimit), ctx)
		},
		Xtract: func(d *ReportLine) any {
			return d.PowerLimit
		},
		Help: "(uint64) Current power limit in Watts",
	},
	"CEClock": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatUint64((d.CEClock), ctx)
		},
		Xtract: func(d *ReportLine) any {
			return d.CEClock
		},
		Help: "(uint64) Current compute element clock in MHz",
	},
	"MemoryClock": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatUint64((d.MemoryClock), ctx)
		},
		Xtract: func(d *ReportLine) any {
			return d.MemoryClock
		},
		Help: "(uint64) Current memory clock in MHz",
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
	"Index": Predicate[*ReportLine]{
		Convert: CvtString2Uint64,
		Compare: func(d *ReportLine, v any) int {
			return cmp.Compare((d.Index), v.(uint64))
		},
	},
	"Fan": Predicate[*ReportLine]{
		Convert: CvtString2Uint64,
		Compare: func(d *ReportLine, v any) int {
			return cmp.Compare((d.Fan), v.(uint64))
		},
	},
	"Memory": Predicate[*ReportLine]{
		Convert: CvtString2Uint64,
		Compare: func(d *ReportLine, v any) int {
			return cmp.Compare((d.Memory), v.(uint64))
		},
	},
	"Temperature": Predicate[*ReportLine]{
		Convert: CvtString2Int64,
		Compare: func(d *ReportLine, v any) int {
			return cmp.Compare((d.Temperature), v.(int64))
		},
	},
	"Power": Predicate[*ReportLine]{
		Convert: CvtString2Uint64,
		Compare: func(d *ReportLine, v any) int {
			return cmp.Compare((d.Power), v.(uint64))
		},
	},
	"PowerLimit": Predicate[*ReportLine]{
		Convert: CvtString2Uint64,
		Compare: func(d *ReportLine, v any) int {
			return cmp.Compare((d.PowerLimit), v.(uint64))
		},
	},
	"CEClock": Predicate[*ReportLine]{
		Convert: CvtString2Uint64,
		Compare: func(d *ReportLine, v any) int {
			return cmp.Compare((d.CEClock), v.(uint64))
		},
	},
	"MemoryClock": Predicate[*ReportLine]{
		Convert: CvtString2Uint64,
		Compare: func(d *ReportLine, v any) int {
			return cmp.Compare((d.MemoryClock), v.(uint64))
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
	"default": []string{"Hostname", "Gpu", "Timestamp", "Memory", "PowerDraw"},
	"Default": []string{"Hostname", "Gpu", "Timestamp", "Memory", "PowerDraw"},
	"All":     []string{"Timestamp", "Hostname", "Index", "Fan", "Memory", "Temperature", "PowerDraw", "PowerLimit", "CEClock", "MemoryClock"},
}

const gpuDefaultFields = "default"
