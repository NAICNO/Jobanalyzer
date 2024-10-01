package uptime

import (
	"io"
	"reflect"
	"sort"
	"strings"

	uslices "go-utils/slices"

	. "sonalyze/command"
)

func (uc *UptimeCommand) printReports(out io.Writer, reports []*UptimeLine) {
	sort.Sort(sortableReports(reports))
	FormatData(
		out,
		uc.PrintFields,
		uptimeFormatters,
		uc.PrintOpts,
		uslices.Map(reports, func(x *UptimeLine) any { return x }),
		ComputePrintMods(uc.PrintOpts),
	)
}

type sortableReports []*UptimeLine

func (sr sortableReports) Len() int {
	return len(sr)
}

func (sr sortableReports) Swap(i, j int) {
	sr[i], sr[j] = sr[j], sr[i]
}

func (sr sortableReports) Less(i, j int) bool {
	// Sort the reports by hostname, start time, device type, end time, and state
	if sr[i].Hostname == sr[j].Hostname {
		if sr[i].Start == sr[j].Start {
			if sr[i].Device == sr[j].Device {
				if sr[i].End == sr[j].End {
					return sr[i].State < sr[j].State
				}
				return sr[i].End < sr[j].End
			}
			return sr[i].Device == "host"
		}
		return sr[i].Start < sr[j].Start
	}
	return sr[i].Hostname < sr[j].Hostname
}

func (uc *UptimeCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(uc.Fmt, uptimeHelp, uptimeFormatters, uptimeAliases, uptimeDefaultFields)
}

const uptimeHelp = `
print
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
var uptimeFormatters map[string]Formatter[any, PrintMods] = ReflectFormattersFromTags(
	// TODO: Go 1.22, reflect.TypeFor[UptimeLine]
	reflect.TypeOf((*UptimeLine)(nil)).Elem(),
	nil,
)
