// Produce a per-node timeline of gpu data.
//
// See summary.txt for info.
//
// TODO (coming with computed queries):
//  - selection by every field ("FanPct > 50", "PowerDrawW > 150")

package gpus

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"go-utils/hostglob"
	uslices "go-utils/slices"

	. "sonalyze/cmd"
	. "sonalyze/common"
	"sonalyze/db"
	"sonalyze/sonarlog"
	. "sonalyze/table"
)

type GpuCommand struct /* implements AnalysisCommand */ {
	// Almost SharedArgs, but HostArgs instead of RecordFilterArgs
	DevArgs
	SourceArgs
	HostArgs
	VerboseArgs
	FormatArgs
	ConfigFileArgs

	Gpu int
}

var _ = AnalysisCommand((*GpuCommand)(nil))

//go:embed summary.txt
var summary string

func (gc *GpuCommand) Summary() string {
	return summary
}

func (gc *GpuCommand) Add(fs *CLI) {
	gc.DevArgs.Add(fs)
	gc.SourceArgs.Add(fs)
	gc.HostArgs.Add(fs)
	gc.VerboseArgs.Add(fs)
	gc.FormatArgs.Add(fs)
	gc.ConfigFileArgs.Add(fs)
	fs.Group("record-filter")
	fs.IntVar(&gc.Gpu, "gpu", -1, "Select single GPU")
}

func (gc *GpuCommand) Validate() error {
	return errors.Join(
		gc.DevArgs.Validate(),
		gc.SourceArgs.Validate(),
		gc.HostArgs.Validate(),
		gc.VerboseArgs.Validate(),
		gc.ConfigFileArgs.Validate(),
		ValidateFormatArgs(
			&gc.FormatArgs, gpusDefaultFields, gpusFormatters, gpusAliases, DefaultFixed),
	)
}

func (gc *GpuCommand) ReifyForRemote(x *ArgReifier) error {
	// gc.Verbose is not reified, as for SharedArgs.
	if gc.Gpu != -1 {
		x.IntUnchecked("gpu", gc.Gpu)
	}
	return errors.Join(
		gc.DevArgs.ReifyForRemote(x),
		gc.SourceArgs.ReifyForRemote(x),
		gc.HostArgs.ReifyForRemote(x),
		gc.FormatArgs.ReifyForRemote(x),
		gc.ConfigFileArgs.ReifyForRemote(x),
	)
}

type ReportLine struct {
	Timestamp DateTimeValue
	Hostname  Ustr
	Gpu       int
	*sonarlog.PerGpuDatum
}

func (gc *GpuCommand) Perform(stdin io.Reader, stdout, stderr io.Writer) error {
	cfg, err := db.MaybeGetConfig(gc.ConfigFile())
	if err != nil {
		return err
	}

	hostGlobber, err := hostglob.NewGlobber(true, gc.Host)
	if err != nil {
		return err
	}

	var theLog db.SampleCluster
	if len(gc.LogFiles) > 0 {
		theLog, err = db.OpenTransientSampleCluster(gc.LogFiles, cfg)
	} else {
		theLog, err = db.OpenPersistentCluster(gc.DataDir, cfg)
	}
	if err != nil {
		return fmt.Errorf("Failed to open log store: %v", err)
	}

	streams, _, read, dropped, err :=
		sonarlog.ReadGpuDataStreams(
			theLog,
			gc.FromDate,
			gc.ToDate,
			hostGlobber,
			gc.Verbose,
		)
	if err != nil {
		return fmt.Errorf("Failed to read log records: %v", err)
	}
	if gc.Verbose {
		Log.Infof("%d records read + %d dropped\n", read, dropped)
		UstrStats(stderr, false)
	}

	reports := make([]*ReportLine, 0)
	for _, s := range streams {
		for _, d := range s.Data {
			for i, gpu := range d.Decoded {
				if gc.Gpu != -1 && i == gc.Gpu {
					var r ReportLine
					r.Timestamp = DateTimeValue(d.Time)
					r.Hostname = s.Host
					r.Gpu = i
					r.PerGpuDatum = &gpu
					reports = append(reports, &r)
				}
			}
		}
	}

	FormatData(
		stdout,
		gc.PrintFields,
		gpusFormatters,
		gc.PrintOpts,
		uslices.Map(reports, func(x *ReportLine) any { return x }),
	)

	return nil
}

func (gc *GpuCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(gc.Fmt, gpusHelp, gpusFormatters, gpusAliases, gpusDefaultFields)
}

type SFS = SimpleFormatSpec

var (
	gpusFormatters = DefineTableFromMap(
		reflect.TypeFor[ReportLine](),
		map[string]any{
			"Timestamp":   SFS{"Timestamp of when the reading was taken", ""},
			"Hostname":    SFS{"Name that host is known by on the cluster", ""},
			"Gpu":         SFS{"Card index on the host",""},
			"FanPct":      SFS{"Fan speed in percent of max",""},
			"PerfMode":    SFS{"Numeric performance mode",""},
			"MemUsedKB":   SFS{"Amount of memory in use",""},
			"TempC":       SFS{"Card temperature in degrees C",""},
			"PowerDrawW":  SFS{"Current power draw in Watts",""},
			"PowerLimitW": SFS{"Current power limit in Watts", ""},
			"CeClockMHz":  SFS{"Current compute element clock", ""},
			"MemClockMHz": SFS{"Current memory clock", ""},
		},
	)

	gpusAliases = map[string][]string{
		"default":   []string{"Hostname","Gpu","Timestamp","MemUsedKB","PowerDrawW"},
		"Default":   []string{"Hostname","Gpu","Timestamp","MemUsedKB","PowerDrawW"},
		"All":       []string{
			"Timestamp","Hostname","Gpu","FanPct","PerfMode","MemUsedKB","TempC",
			"PowerDrawW","PowerLimitW","CeClockMHz","MemClockMHz",
		},
	}

	gpusDefaultFields = strings.Join(gpusAliases["default"], ",")
)

const gpusHelp = `
gpu
  Extract information about individual gpus on the cluster from sample data.
`
