package sacct

import (
	"io"
	"math"
	"reflect"
	"strings"

	uslices "go-utils/slices"

	. "sonalyze/common"
	. "sonalyze/table"
)

type SacctRegular struct {
	Start               IsoDateTimeOrUnknown `desc:"Start time of job, if any"`
	End                 IsoDateTimeOrUnknown `desc:"End time of job"`
	Submit              IsoDateTimeOrUnknown `desc:"Submit time of job"`
	RequestedCPU        uint64               `desc:"Requested CPU time (elapsed * cores * nodes)"`
	UsedCPU             uint64               `desc:"Used CPU time"`
	RelativeCPU         int                  `alias:"rcpu" desc:"Percent cpu utilization: UsedCPU/RequestedCPU*100"`
	RelativeResidentMem int                  `alias:"rmem" desc:"Percent memory utilization: MaxRSS/ReqMem*100"`
	User                Ustr                 `desc:"Job's user"`
	JobName             UstrMax30            `desc:"Job name"`
	State               Ustr                 `desc:"Job completion state"`
	Account             Ustr                 `desc:"Job's account"`
	Reservation         Ustr                 `desc:"Job's reservation, if any"`
	Layout              Ustr                 `desc:"Job's layout, if any"`
	NodeList            Ustr                 `desc:"Job's node list"`
	JobID               uint32               `desc:"Primary Job ID"`
	MaxRSS              uint32               `desc:"Max resident set size (RSS) across all steps (GB)"`
	ReqMem              uint32               `desc:"Raw requested memory (GB)"`
	ReqCPUS             uint32               `desc:"Raw requested CPU cores"`
	ReqGPUS             Ustr                 `desc:"Raw requested GPU cards"`
	ReqNodes            uint32               `desc:"Raw requested system nodes"`
	Elapsed             uint32               `desc:"Time elapsed"`
	Suspended           uint32               `desc:"Time suspended"`
	Timelimit           uint32               `desc:"Time limit in seconds"`
	ExitCode            uint8                `desc:"Exit code"`
	Wait                int64                `desc:"Wait time of job (start - submit), in seconds"`
	Partition           Ustr                 `desc:"Requested partition"`
	ArrayJobID          uint32               `desc:"ID of the overarching array job"`
	ArrayIndex          uint32               `desc:"Index of this job within an array job"`
}

func (sc *SacctCommand) printRegularJobs(stdout io.Writer, regular []*sacctSummary) {
	// TODO: By and by it may be possible to lift this extra loop into the loop already being run in
	// perform.go to compute the `regular` values, and not allocate extra values here.
	toPrint := make([]*SacctRegular, len(regular))
	for i, r := range regular {
		var relativeCpu, relativeResidentMem int
		var waitTime int64
		if r.requestedCpu > 0 {
			relativeCpu = int(math.Round(100 * float64(r.usedCpu) / float64(r.requestedCpu)))
		}
		if r.Main.ReqMem > 0 {
			relativeResidentMem = int(math.Round(100 * float64(r.maxrss) / float64(r.Main.ReqMem)))
		}
		if r.Main.Start > 0 {
			waitTime = r.Main.Start - r.Main.Submit
		}
		toPrint[i] = &SacctRegular{
			Start:               IsoDateTimeOrUnknown(r.Main.Start),
			End:                 IsoDateTimeOrUnknown(r.Main.End),
			Submit:              IsoDateTimeOrUnknown(r.Main.Submit),
			RequestedCPU:        r.requestedCpu,
			UsedCPU:             r.usedCpu,
			RelativeCPU:         relativeCpu,
			RelativeResidentMem: relativeResidentMem,
			User:                r.Main.User,
			JobName:             UstrMax30(r.Main.JobName),
			State:               r.Main.State,
			Account:             r.Main.Account,
			Reservation:         r.Main.Reservation,
			Layout:              r.Main.Layout,
			NodeList:            r.Main.NodeList,
			JobID:               r.Main.JobID,
			MaxRSS:              r.maxrss,
			ReqMem:              r.Main.ReqMem,
			ReqCPUS:             r.Main.ReqCPUS,
			ReqNodes:            r.Main.ReqNodes,
			Elapsed:             r.Main.ElapsedRaw,
			Suspended:           r.Main.Suspended,
			Timelimit:           r.Main.TimelimitRaw,
			ExitCode:            r.Main.ExitCode,
			Wait:                waitTime,
			Partition:           r.Main.Partition,
			ReqGPUS:             r.Main.ReqGPUS,
			ArrayJobID:          r.Main.ArrayJobID,
			ArrayIndex:          r.Main.ArrayIndex,
		}
	}
	FormatData(
		stdout,
		sc.PrintFields,
		sacctFormatters,
		sc.PrintOpts,
		uslices.Map(toPrint, func(x *SacctRegular) any { return x }),
	)
}

func (sc *SacctCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(sc.Fmt, sacctHelp, sacctFormatters, sacctAliases, sacctDefaultFields)
}

const sacctHelp = `
sacct
  Aggregate SLURM sacct data into data about jobs and present them.
`

const v0SacctDefaultFields = "JobID,JobName,User,Account,rcpu,rmem"
const v1SacctDefaultFields = "JobID,JobName,User,Account,RelativeCPU,RelativeResidentMem"
const sacctDefaultFields = v0SacctDefaultFields

// MT: Constant after initialization; immutable
var sacctAliases = map[string][]string{
	"default":   strings.Split(sacctDefaultFields, ","),
	"v0default": strings.Split(v0SacctDefaultFields, ","),
	"v1default": strings.Split(v1SacctDefaultFields, ","),
}

// MT: Constant after initialization; immutable
var sacctFormatters = DefineTableFromTags(reflect.TypeFor[SacctRegular](), nil)
