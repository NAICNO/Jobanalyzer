// DO NOT EDIT.  Generated from print.go by generate-table

package uptime

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
var uptimeFormatters = map[string]Formatter[*UptimeLine]{
	"Device": {
		Fmt: func(d *UptimeLine, ctx PrintMods) string {
			return FormatString(d.Device, ctx)
		},
		Help: "(string) Device type: 'host' or 'gpu'",
	},
	"Hostname": {
		Fmt: func(d *UptimeLine, ctx PrintMods) string {
			return FormatString(d.Hostname, ctx)
		},
		Help: "(string) Host name for the device",
	},
	"State": {
		Fmt: func(d *UptimeLine, ctx PrintMods) string {
			return FormatString(d.State, ctx)
		},
		Help: "(string) Device state: 'up' or 'down'",
	},
	"Start": {
		Fmt: func(d *UptimeLine, ctx PrintMods) string {
			return FormatDateTimeValue(d.Start, ctx)
		},
		Help: "(DateTimeValue) Start time of 'up' or 'down' window",
	},
	"End": {
		Fmt: func(d *UptimeLine, ctx PrintMods) string {
			return FormatDateTimeValue(d.End, ctx)
		},
		Help: "(DateTimeValue) End time of 'up' or 'down' window",
	},
}

func init() {
	DefAlias(uptimeFormatters, "Device", "device")
	DefAlias(uptimeFormatters, "Hostname", "host")
	DefAlias(uptimeFormatters, "State", "state")
	DefAlias(uptimeFormatters, "Start", "start")
	DefAlias(uptimeFormatters, "End", "end")
}

// MT: Constant after initialization; immutable
var uptimePredicates = map[string]Predicate[*UptimeLine]{
	"Device": Predicate[*UptimeLine]{
		Compare: func(d *UptimeLine, v any) int {
			return cmp.Compare(d.Device, v.(string))
		},
	},
	"Hostname": Predicate[*UptimeLine]{
		Compare: func(d *UptimeLine, v any) int {
			return cmp.Compare(d.Hostname, v.(string))
		},
	},
	"State": Predicate[*UptimeLine]{
		Compare: func(d *UptimeLine, v any) int {
			return cmp.Compare(d.State, v.(string))
		},
	},
	"Start": Predicate[*UptimeLine]{
		Convert: CvtString2DateTimeValue,
		Compare: func(d *UptimeLine, v any) int {
			return cmp.Compare(d.Start, v.(DateTimeValue))
		},
	},
	"End": Predicate[*UptimeLine]{
		Convert: CvtString2DateTimeValue,
		Compare: func(d *UptimeLine, v any) int {
			return cmp.Compare(d.End, v.(DateTimeValue))
		},
	},
}

type UptimeLine struct {
	Device   string
	Hostname string
	State    string
	Start    DateTimeValue
	End      DateTimeValue
}

const uptimeHelp = `
uptime
  Compute the status of hosts and GPUs across time.  Default output format
  is 'fixed'.
`

func (c *UptimeCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(c.Fmt, uptimeHelp, uptimeFormatters, uptimeAliases, uptimeDefaultFields)
}

// MT: Constant after initialization; immutable
var uptimeAliases = map[string][]string{
	"default": []string{"device", "host", "state", "start", "end"},
	"Default": []string{"Device", "Hostname", "State", "Start", "End"},
	"all":     []string{"default"},
	"All":     []string{"Default"},
}

const uptimeDefaultFields = "default"
