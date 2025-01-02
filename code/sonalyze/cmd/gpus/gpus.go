// Produce a per-node timeline of gpu data.
//
// See summary.txt for info.
//
// TODO (coming with computed queries):
//  - selection by every field ("FanPct > 50", "PowerDrawW > 150")

package gpus

import (
	"errors"
	"fmt"
	"io"

	"go-utils/hostglob"

	. "sonalyze/cmd"
	. "sonalyze/common"
	"sonalyze/db"
	"sonalyze/sonarlog"
	. "sonalyze/table"
)

//go:generate ../../../generate-table/generate-table -o gpus-table.go gpus.go

/*TABLE gpu

package gpus

import (
	. "sonalyze/common"
    . "sonalyze/table"
)

%%

FIELDS *ReportLine

 Timestamp   DateTimeValue desc:"Timestamp of when the reading was taken"
 Hostname    Ustr          desc:"Name that host is known by on the cluster"
 Gpu         int           desc:"Card index on the host"
 FanPct      int           desc:"Fan speed in percent of max"
 PerfMode    int           desc:"Numeric performance mode"
 MemUsedKB   int           desc:"Amount of memory in use"
 TempC       int           desc:"Card temperature in degrees C"
 PowerDrawW  int           desc:"Current power draw in Watts"
 PowerLimitW int           desc:"Current power limit in Watts"
 CeClockMHz  int           desc:"Current compute element clock"
 MemClockMHz int           desc:"Current memory clock"

SUMMARY GpuCommand

Experimental: Print per-gpu data across time for one or more cards on one or more nodes.

HELP GpuCommand

  Extract information about individual gpus on the cluster from sample data.  The default
  format is 'fixed'.

ALIASES

  default   Hostname,Gpu,Timestamp,MemUsedKB,PowerDrawW
  Default   Hostname,Gpu,Timestamp,MemUsedKB,PowerDrawW
  All       Timestamp,Hostname,Gpu,FanPct,PerfMode,MemUsedKB,TempC,PowerDrawW,\
            PowerLimitW,CeClockMHz,MemClockMHz

DEFAULTS default

ELBAT*/

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
			&gc.FormatArgs, gpuDefaultFields, gpuFormatters, gpuAliases, DefaultFixed),
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
		gpuFormatters,
		gc.PrintOpts,
		reports,
	)

	return nil
}
