// DO NOT EDIT.  Generated from gpus.go by generate-table

package gpus

import (
	. "sonalyze/common"
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
var gpuFormatters = map[string]Formatter[*ReportLine]{
	"Timestamp": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatDateTimeValue(DateTimeValue(d.Timestamp), ctx)
		},
		Help: "(DateTimeValue) Timestamp of when the reading was taken",
	},
	"Hostname": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatUstr(Ustr(d.Hostname), ctx)
		},
		Help: "(string) Name that host is known by on the cluster",
	},
	"Gpu": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatInt(int(d.Gpu), ctx)
		},
		Help: "(int) Card index on the host",
	},
	"FanPct": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatInt(int(d.FanPct), ctx)
		},
		Help: "(int) Fan speed in percent of max",
	},
	"PerfMode": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatInt(int(d.PerfMode), ctx)
		},
		Help: "(int) Numeric performance mode",
	},
	"MemUsedKB": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatInt(int(d.MemUsedKB), ctx)
		},
		Help: "(int) Amount of memory in use",
	},
	"TempC": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatInt(int(d.TempC), ctx)
		},
		Help: "(int) Card temperature in degrees C",
	},
	"PowerDrawW": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatInt(int(d.PowerDrawW), ctx)
		},
		Help: "(int) Current power draw in Watts",
	},
	"PowerLimitW": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatInt(int(d.PowerLimitW), ctx)
		},
		Help: "(int) Current power limit in Watts",
	},
	"CeClockMHz": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatInt(int(d.CeClockMHz), ctx)
		},
		Help: "(int) Current compute element clock",
	},
	"MemClockMHz": {
		Fmt: func(d *ReportLine, ctx PrintMods) string {
			return FormatInt(int(d.MemClockMHz), ctx)
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
