// DO NOT EDIT.  Generated from diskprof.go by generate-table

package diskprof

import "sonalyze/db/repr"

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
var diskprofFormatters = map[string]Formatter[*repr.DiskSample]{
	"Timestamp": {
		Fmt: func(d *repr.DiskSample, ctx PrintMods) string {
			return FormatDateTimeValue((d.Timestamp), ctx)
		},
		Xtract: func(d *repr.DiskSample) any {
			return d.Timestamp
		},
		Help: "(DateTimeValue) Full ISO timestamp of when the reading was taken",
	},
	"Hostname": {
		Fmt: func(d *repr.DiskSample, ctx PrintMods) string {
			return FormatUstr((d.Hostname), ctx)
		},
		Xtract: func(d *repr.DiskSample) any {
			return d.Hostname
		},
		Help: "(string) Name that host is known by on the cluster",
	},
	"Name": {
		Fmt: func(d *repr.DiskSample, ctx PrintMods) string {
			return FormatUstr((d.Name), ctx)
		},
		Xtract: func(d *repr.DiskSample) any {
			return d.Name
		},
		Help: "(string) Name of disk",
	},
	"Major": {
		Fmt: func(d *repr.DiskSample, ctx PrintMods) string {
			return FormatUint64((d.Major), ctx)
		},
		Xtract: func(d *repr.DiskSample) any {
			return d.Major
		},
		Help: "(uint64) Major device number",
	},
	"Minor": {
		Fmt: func(d *repr.DiskSample, ctx PrintMods) string {
			return FormatUint64((d.Minor), ctx)
		},
		Xtract: func(d *repr.DiskSample) any {
			return d.Minor
		},
		Help: "(uint64) Minor device number",
	},
	"MsReading": {
		Fmt: func(d *repr.DiskSample, ctx PrintMods) string {
			return FormatUint64((d.MsReading), ctx)
		},
		Xtract: func(d *repr.DiskSample) any {
			return d.MsReading
		},
		Help: "(uint64) ms spent reading",
	},
	"MsWriting": {
		Fmt: func(d *repr.DiskSample, ctx PrintMods) string {
			return FormatUint64((d.MsWriting), ctx)
		},
		Xtract: func(d *repr.DiskSample) any {
			return d.MsWriting
		},
		Help: "(uint64) ms spent writing",
	},
}

func init() {
	DefAlias(diskprofFormatters, "Timestamp", "timestamp")
	DefAlias(diskprofFormatters, "Hostname", "host")
	DefAlias(diskprofFormatters, "Name", "name")
	DefAlias(diskprofFormatters, "Major", "major")
	DefAlias(diskprofFormatters, "Minor", "minor")
	DefAlias(diskprofFormatters, "MsReading", "ms-reading")
	DefAlias(diskprofFormatters, "MsWriting", "ms-writing")
}

// MT: Constant after initialization; immutable
var diskprofPredicates = map[string]Predicate[*repr.DiskSample]{
	"Timestamp": Predicate[*repr.DiskSample]{
		Convert: CvtString2DateTimeValue,
		Compare: func(d *repr.DiskSample, v any) int {
			return cmp.Compare((d.Timestamp), v.(DateTimeValue))
		},
	},
	"Hostname": Predicate[*repr.DiskSample]{
		Convert: CvtString2Ustr,
		Compare: func(d *repr.DiskSample, v any) int {
			return cmp.Compare((d.Hostname), v.(Ustr))
		},
	},
	"Name": Predicate[*repr.DiskSample]{
		Convert: CvtString2Ustr,
		Compare: func(d *repr.DiskSample, v any) int {
			return cmp.Compare((d.Name), v.(Ustr))
		},
	},
	"Major": Predicate[*repr.DiskSample]{
		Convert: CvtString2Uint64,
		Compare: func(d *repr.DiskSample, v any) int {
			return cmp.Compare((d.Major), v.(uint64))
		},
	},
	"Minor": Predicate[*repr.DiskSample]{
		Convert: CvtString2Uint64,
		Compare: func(d *repr.DiskSample, v any) int {
			return cmp.Compare((d.Minor), v.(uint64))
		},
	},
	"MsReading": Predicate[*repr.DiskSample]{
		Convert: CvtString2Uint64,
		Compare: func(d *repr.DiskSample, v any) int {
			return cmp.Compare((d.MsReading), v.(uint64))
		},
	},
	"MsWriting": Predicate[*repr.DiskSample]{
		Convert: CvtString2Uint64,
		Compare: func(d *repr.DiskSample, v any) int {
			return cmp.Compare((d.MsWriting), v.(uint64))
		},
	},
}

func (c *DiskProfCommand) Summary(out io.Writer) {
	fmt.Fprint(out, `  Display disk sample data
`)
}

const diskprofHelp = `
diskprof
  Extract disk profiling data from sample data and present it in primitive form.  Output
  records are sorted by node name and time.  The default format is 'fixed'.
`

func (c *DiskProfCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(c.Fmt, diskprofHelp, diskprofFormatters, diskprofAliases, diskprofDefaultFields)
}

// MT: Constant after initialization; immutable
var diskprofAliases = map[string][]string{
	"Default": []string{"Timestamp", "Hostname", "Name", "MsReading", "MsWriting"},
	"All":     []string{"Timestamp", "Hostname", "Name", "Major", "Minor", "MsReading", "MsWriting"},
}

const diskprofDefaultFields = "Default"
