// DO NOT EDIT.  Generated from metadata.go by generate-table

package metadata

import (
	"cmp"
	"fmt"
	"io"
	. "sonalyze/common"
	. "sonalyze/table"
)

var (
	_ = cmp.Compare(0, 0)
	_ fmt.Formatter
	_ = io.SeekStart
	_ = UstrEmpty
)

// MT: Constant after initialization; immutable
var metadataFormatters = map[string]Formatter[*metadataItem]{
	"Hostname": {
		Fmt: func(d *metadataItem, ctx PrintMods) string {
			return FormatString(d.Hostname, ctx)
		},
		Help: "(string) Name that host is known by on the cluster",
	},
	"Earliest": {
		Fmt: func(d *metadataItem, ctx PrintMods) string {
			return FormatDateTimeValue(d.Earliest, ctx)
		},
		Help: "(DateTimeValue) Timestamp of earliest sample for host",
	},
	"Latest": {
		Fmt: func(d *metadataItem, ctx PrintMods) string {
			return FormatDateTimeValue(d.Latest, ctx)
		},
		Help: "(DateTimeValue) Timestamp of latest sample for host",
	},
}

func init() {
	DefAlias(metadataFormatters, "Hostname", "host")
	DefAlias(metadataFormatters, "Earliest", "earliest")
	DefAlias(metadataFormatters, "Latest", "latest")
}

type metadataItem struct {
	Hostname string
	Earliest DateTimeValue
	Latest   DateTimeValue
}

func (c *MetadataCommand) Summary(out io.Writer) {
	fmt.Fprint(out, `Display metadata about the sample streams in the database.

One or more of -files, -times and -bounds must be selected to produce
output.

Mostly this command is useful for debugging, but -bounds can be used to
detect whether a node is up more cheaply than the "uptime" operation.
`)
}

const metadataHelp = `
metadata
  Compute time bounds and file names for the run.
`

func (c *MetadataCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(c.Fmt, metadataHelp, metadataFormatters, metadataAliases, metadataDefaultFields)
}

// MT: Constant after initialization; immutable
var metadataAliases = map[string][]string{
	"default": []string{"host", "earliest", "latest"},
	"Default": []string{"Hostname", "Earliest", "Latest"},
	"all":     []string{"default"},
	"All":     []string{"Default"},
}

const metadataDefaultFields = "default"
