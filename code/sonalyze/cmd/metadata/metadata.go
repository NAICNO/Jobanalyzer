package metadata

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"slices"
	"time"

	. "sonalyze/cmd"
	. "sonalyze/common"
	"sonalyze/data/sample"
	"sonalyze/db/types"
	. "sonalyze/table"
)

//go:generate ../../../generate-table/generate-table -o metadata-table.go metadata.go

/*TABLE metadata

package metadata

%%

FIELDS *metadataItem

 Hostname string        alias:"host"     desc:"Name that host is known by on the cluster"
 Earliest DateTimeValue alias:"earliest" desc:"Timestamp of earliest sample for host"
 Latest   DateTimeValue alias:"latest"   desc:"Timestamp of latest sample for host"

GENERATE metadataItem

SUMMARY MetadataCommand

Display metadata about the sample streams in the database.

One or more of -files, -times and -bounds must be selected to produce
output.

Mostly this command is useful for debugging, but -bounds can be used to
detect whether a node is up more cheaply than the "uptime" operation.

HELP MetadataCommand

  Compute time bounds and file names for the run.

ALIASES

  default  host,earliest,latest
  Default  Hostname,Earliest,Latest
  all      default
  All      Default

DEFAULTS default

ELBAT*/

type MetadataCommand struct /* implements SampleAnalysisCommand */ {
	SampleAnalysisArgs
	FormatArgs

	MergeByHostAndJob bool // Inert, but compatible
	MergeByJob        bool
	Times             bool
	Files             bool
	Bounds            bool
}

var _ = SampleAnalysisCommand((*MetadataCommand)(nil))

func (mdc *MetadataCommand) Add(fs *CLI) {
	mdc.SampleAnalysisArgs.Add(fs)
	mdc.FormatArgs.Add(fs)

	fs.Group("aggregation")
	fs.BoolVar(&mdc.MergeByHostAndJob, "merge-by-host-and-job", false,
		"Merge streams that have the same host and job ID")
	fs.BoolVar(&mdc.MergeByJob, "merge-by-job", false,
		"Merge streams that have the same job ID, across hosts")

	fs.Group("operation-selection")
	fs.BoolVar(&mdc.Files, "files", false, "List files selected by the record filter")
	fs.BoolVar(&mdc.Times, "times", false, "Parse the -from and -to timestamps")
	fs.BoolVar(&mdc.Bounds, "bounds", false,
		"List each host with its earliest/latest record timestamp")
}

func (mdc *MetadataCommand) ReifyForRemote(x *ArgReifier) error {
	e1 := errors.Join(
		mdc.SampleAnalysisArgs.ReifyForRemote(x),
		mdc.FormatArgs.ReifyForRemote(x),
	)
	x.Bool("merge-by-host-and-job", mdc.MergeByHostAndJob)
	x.Bool("merge-by-job", mdc.MergeByJob)
	x.Bool("files", mdc.Files)
	x.Bool("times", mdc.Times)
	x.Bool("bounds", mdc.Bounds)
	return e1
}

func (mdc *MetadataCommand) Validate() error {
	return errors.Join(
		mdc.SampleAnalysisArgs.Validate(),
		ValidateFormatArgs(
			&mdc.FormatArgs,
			metadataDefaultFields,
			metadataFormatters,
			metadataAliases,
			DefaultCsv,
		),
	)
}

func (mdc *MetadataCommand) DefaultRecordFilters() (
	allUsers, skipSystemUsers, excludeSystemCommands, excludeHeartbeat bool,
) {
	allUsers, skipSystemUsers, determined := mdc.RecordFilterArgs.DefaultUserFilters()
	if !determined {
		allUsers, skipSystemUsers = true, false
	}
	excludeSystemCommands = false
	excludeHeartbeat = false
	return
}

func (mdc *MetadataCommand) Perform(
	out io.Writer,
	meta types.Context,
	filter sample.QueryFilter,
	hosts *Hosts,
	recordFilter *sample.SampleFilter,
) error {
	sdp, err := sample.OpenSampleDataProvider(meta)
	if err != nil {
		return err
	}
	streams, bounds, read, dropped, err :=
		sdp.Query(
			filter.FromDate,
			filter.ToDate,
			hosts,
			recordFilter,
			mdc.Bounds,
			mdc.Verbose,
		)
	if err != nil {
		return fmt.Errorf("Failed to read log records: %v", err)
	}
	if mdc.Verbose {
		Log.Infof("%d records read + %d dropped\n", read, dropped)
		UstrStats(out, false)
	}
	if mdc.Verbose {
		Log.Infof("Streams constructed by postprocessing: %d", len(streams))
		numSamples := 0
		for _, stream := range streams {
			numSamples += len(*stream)
		}
		Log.Infof("Samples retained after filtering: %d", numSamples)
	}

	if mdc.Times {
		fmt.Fprintf(out, "From: %s\n", mdc.FromDate.Format(time.RFC3339))
		fmt.Fprintf(out, "To:   %s\n", mdc.ToDate.Format(time.RFC3339))
	}

	if mdc.Files {
		// For -files, print the full paths all the input files as presented to os.Open.
		files, err := sdp.Filenames(mdc.FromDate, mdc.ToDate, hosts)
		if err != nil {
			return err
		}
		for _, name := range files {
			fmt.Fprintln(out, name)
		}
	}

	if mdc.Bounds {
		if mdc.MergeByJob {
			_, bounds = sample.MergeByJob(streams, bounds)
		}
		items := make([]*metadataItem, 0)
		for k, v := range bounds {
			items = append(items, &metadataItem{
				Hostname: k.String(),
				Earliest: DateTimeValue(v.Earliest),
				Latest:   DateTimeValue(v.Latest),
			})
		}
		slices.SortFunc(items, func(a, b *metadataItem) int {
			c := cmp.Compare(a.Hostname, b.Hostname)
			if c == 0 {
				c = cmp.Compare(a.Earliest, b.Earliest)
			}
			return c
		})
		items, err := ApplyQuery(mdc.ParsedQuery, metadataFormatters, metadataPredicates, items)
		if err != nil {
			return err
		}
		FormatData(out, mdc.PrintFields, metadataFormatters, mdc.PrintOpts, items)
	}

	return nil
}
