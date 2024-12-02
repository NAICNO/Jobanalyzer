// DO NOT EDIT.  Generated from metadata.go by generate-table

package metadata

import (
	. "sonalyze/table"
)

// MT: Constant after initialization; immutable
var metadataFormatters = map[string]Formatter[*metadataItem]{
	"Hostname": {
		Fmt: func(d *metadataItem, ctx PrintMods) string {
			return FormatString(string(d.Hostname), ctx)
		},
		Help: "Name that host is known by on the cluster",
	},
	"Earliest": {
		Fmt: func(d *metadataItem, ctx PrintMods) string {
			return FormatDateTimeValue(DateTimeValue(d.Earliest), ctx)
		},
		Help: "Timestamp of earliest sample for host",
	},
	"Latest": {
		Fmt: func(d *metadataItem, ctx PrintMods) string {
			return FormatDateTimeValue(DateTimeValue(d.Latest), ctx)
		},
		Help: "Timestamp of latest sample for host",
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
