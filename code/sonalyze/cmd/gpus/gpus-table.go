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
