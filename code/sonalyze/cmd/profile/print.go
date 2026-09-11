// For one particular job, break it down in its component processes and print individual stats for
// each process for each time slot.
//
// The data fields are raw data items from Sample records.  There are no average or peak values
// because we only have samples at the start and beginning of the time slot.  There is a
// straightforward derivation from the Sample values to relative values should we need that.
//
//   cpu: the cpu_util_pct field
//   mem: the mem_gb field
//   res: the res_gb field
//   gpu: the gpu_pct field
//   gpumem: the gpumem_gb field
//   nproc: the rolledup field + 1, this is not printed if every process has rolledup=0
//   command: the command field
//
// Output formats: See the individual print-*.go files.

package profile

import (
	"errors"
	"fmt"
	"io"
	"math"
	"slices"
	"strings"

	"sonalyze/data/sample"
	. "sonalyze/table"
)

func (pc *ProfileCommand) printProfile(
	out io.Writer,
	jobId uint32,
	host, user string,
	hasRolledup bool,
	m *profData,
	processes []sample.SampleStream,
	pif *processIndexFactory,
) error {
	// Add the "nproc" field if it is required (we can't do it until rollup has happened) and the
	// fields are still the defaults.  This is not quite compatible with older sonalyze: here we
	// have default fields only if no fields were specified (defaults were applied), while in older
	// code a field list identical to the default would be taken as having default fields too.  The
	// new logic is better.
	//
	// Anyway, this hack depends on nothing interesting having happened to pc.Fmt or pc.PrintFields
	// after the initial parsing.
	//
	// The situation is the same for the "host" field: it gets added late if there is more than one
	// host in the profile.  This field could be anywhere for json output, but for fixed output we
	// want it second, after the timestamp.
	hasDefaultFields := pc.Fmt == ""
	if hasDefaultFields {
		fields := slices.Clone(profileAliases["default"])
		changed := false
		if hasRolledup {
			fields = slices.Insert(fields, len(fields)-1, "nproc")
			changed = true
		}
		if pif.isMultiHost() {
			fields = slices.Insert(fields, 1, "host")
			changed = true
		}
		if changed {
			pc.PrintFields, _, _ = ParseFormatSpec(
				strings.Join(fields, ","),
				"",
				profileFormatters,
				profileAliases,
			)
		}
	}
	if pc.PrintOpts.Csv || pc.PrintOpts.Awk {
		return pc.printCsvOrAwk(out, m, processes, pif)
	}
	if pc.htmlOutput {
		return pc.printHtml(out, jobId, m, processes, pif, host, user)
	}
	if pc.PrintOpts.Fixed {
		return pc.printFixed(out, m, processes, pif)
	}
	if pc.PrintOpts.Json {
		return pc.printJson(out, m, processes, pif)
	}
	panic("Unknown print format")
}

// TODO: Ideally this function would just delegate to the derived formatters where it can, so that
// we can have canonical formatters.  Indeed, the "scaled" thing is really part of the computation,
// not formatting anyway, so should be applied earlier, and we should probably derive some table of
// values to be printed.

func lookupSingleFormatter(
	fields []FieldSpec,
	scaleCpuGpu bool,
) (formatter func(*profDatum) string, err error) {
	n := 0
	for _, f := range fields {
		switch f.Name {
		case "cpu", "CpuUtilPct":
			if scaleCpuGpu {
				formatter = formatCpuUtilPctScaled
			} else {
				formatter = formatCpuUtilPct
			}
			n++
		case "mem", "VirtualMemGB":
			formatter = formatMem
			n++
		case "res", "rss", "ResidentMemGB":
			formatter = formatRes
			n++
		case "gpu", "Gpu":
			if scaleCpuGpu {
				formatter = formatGpuPctScaled
			} else {
				formatter = formatGpuPct
			}
			n++
		case "gpumem", "GpuMemGB":
			formatter = formatGpuMem
			n++
		default:
			err = fmt.Errorf("Not a printable field for this output format: %s", f.Name)
			return
		}
	}
	if n != 1 {
		err = errors.New("Formatted output needs exactly one valid field")
	}
	return
}

// TODO: These formatters are now only used by lookupSingleFormatter().  It's bad that the
// formatting logic also does things like conversion and truncation - we could have generated data
// first and then formatted.

func formatCpuUtilPct(s *profDatum) string {
	return fmt.Sprint(math.Round(float64(s.cpuUtilPct)))
}

func formatCpuUtilPctScaled(s *profDatum) string {
	return fmt.Sprintf("%.1f", float64(s.cpuUtilPct)/100)
}

func formatMem(s *profDatum) string {
	return fmt.Sprint(math.Round(float64(s.cpuKB) / (1024 * 1024)))
}

func formatRes(s *profDatum) string {
	return fmt.Sprint(math.Round(float64(s.rssAnonKB) / (1024 * 1024)))
}

func formatGpuPct(s *profDatum) string {
	return fmt.Sprint(math.Round(float64(s.gpuPct)))
}

func formatGpuPctScaled(s *profDatum) string {
	return fmt.Sprintf("%.1f", float64(s.gpuPct)/100)
}

func formatGpuMem(s *profDatum) string {
	return fmt.Sprint(math.Round(float64(s.gpuKB) / (1024 * 1024)))
}

func formatTime(t int64) string {
	return FormatYyyyMmDdHhMmUtc(t)
}
