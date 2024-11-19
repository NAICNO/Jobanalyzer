package uptime

import (
	"cmp"
	"io"
	"reflect"
	"slices"
	"strings"

	uslices "go-utils/slices"

	. "sonalyze/table"
)

func (uc *UptimeCommand) printReports(out io.Writer, reports []*UptimeLine) {
	slices.SortFunc(reports, func(a, b *UptimeLine) int {
		c := cmp.Compare(a.Hostname, b.Hostname)
		if c == 0 {
			c = cmp.Compare(a.Start, b.Start)
			if c == 0 {
				if a.Device != b.Device {
					if a.Device == "host" {
						c = -1
					} else {
						c = 1
					}
				}
				if c == 0 {
					c = cmp.Compare(a.End, b.End)
					if c == 0 {
						c = cmp.Compare(a.State, b.State)
					}
				}
			}
		}
		return c
	})
	FormatData(
		out,
		uc.PrintFields,
		uptimeFormatters,
		uc.PrintOpts,
		uslices.Map(reports, func(x *UptimeLine) any { return x }),
	)
}

func (uc *UptimeCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(
		uc.Fmt, uptimeHelp, uptimeFormatters, uptimeAliases, uptimeDefaultFields)
}

const uptimeHelp = `
uptime
  Compute the status of hosts and GPUs across time.  Default output format
  is 'fixed'.
`

const v0UptimeDefaultFields = "device,host,state,start,end"
const v1UptimeDefaultFields = "Device,Hostname,State,Start,End"
const uptimeDefaultFields = v0UptimeDefaultFields

// MT: Constant after initialization; immutable
var uptimeAliases = map[string][]string{
	"all":       []string{"device", "host", "state", "start", "end"},
	"default":   strings.Split(uptimeDefaultFields, ","),
	"v0default": strings.Split(v0UptimeDefaultFields, ","),
	"v1default": strings.Split(v1UptimeDefaultFields, ","),
}

// MT: Constant after initialization; immutable
var uptimeFormatters = DefineTableFromTags(reflect.TypeFor[UptimeLine](), nil)
