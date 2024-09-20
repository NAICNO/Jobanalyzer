package metadata

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"sort"
	"time"

	"go-utils/config"
	"go-utils/hostglob"
	. "sonalyze/command"
	. "sonalyze/common"
	"sonalyze/db"
	"sonalyze/sonarlog"
)

type MetadataCommand struct /* implements SampleAnalysisCommand */ {
	SharedArgs
	MergeByHostAndJob bool // Inert, but compatible
	MergeByJob        bool
	Times             bool
	Files             bool
	Bounds            bool
	Fmt               string

	// Synthesized and other
	printFields []string
	printOpts   *FormatOptions
}

var _ SampleAnalysisCommand = (*MetadataCommand)(nil)

func (_ *MetadataCommand) Summary() []string {
	return []string{
		"Display metadata about the sample streams in the database.",
		"One or more of -files, -times and -bounds must be selected to produce",
		"output.",
	}
}

func (mdc *MetadataCommand) Add(fs *flag.FlagSet) {
	mdc.SharedArgs.Add(fs)

	fs.BoolVar(&mdc.MergeByHostAndJob, "merge-by-host-and-job", false,
		"Merge streams that have the same host and job ID")
	fs.BoolVar(&mdc.MergeByJob, "merge-by-job", false,
		"Merge streams that have the same job ID, across hosts")
	fs.BoolVar(&mdc.Files, "files", false, "List selected files")
	fs.BoolVar(&mdc.Times, "times", false, "Show parsed from/to timestamps")
	fs.BoolVar(&mdc.Bounds, "bounds", false, "Show host with earliest/latest timestamp")
	fs.StringVar(&mdc.Fmt, "fmt", "",
		"Select `field,...` and format for the output [default: try -fmt=help]")
}

func (mdc *MetadataCommand) ReifyForRemote(x *Reifier) error {
	e1 := mdc.SharedArgs.ReifyForRemote(x)
	x.Bool("merge-by-host-and-job", mdc.MergeByHostAndJob)
	x.Bool("merge-by-job", mdc.MergeByJob)
	x.Bool("files", mdc.Files)
	x.Bool("times", mdc.Times)
	x.Bool("bounds", mdc.Bounds)
	x.String("fmt", mdc.Fmt)
	return e1
}

func (mdc *MetadataCommand) Validate() error {
	e1 := mdc.SharedArgs.Validate()

	var e2 error
	var others map[string]bool
	mdc.printFields, others, e2 = ParseFormatSpec(metadataDefaultFields, mdc.Fmt, metadataFormatters, metadataAliases)
	if e2 == nil && len(mdc.printFields) == 0 {
		e2 = errors.New("No output fields were selected in format string")
	}
	mdc.printOpts = StandardFormatOptions(others, DefaultCsv)

	return errors.Join(e1, e2)
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

type metadataItem struct {
	host             string
	earliest, latest int64
}

type HostTimeSortableItems []metadataItem

func (xs HostTimeSortableItems) Len() int {
	return len(xs)
}

func (xs HostTimeSortableItems) Less(i, j int) bool {
	if xs[i].host == xs[j].host {
		return xs[i].earliest < xs[j].earliest
	}
	return xs[i].host < xs[j].host
}

func (xs HostTimeSortableItems) Swap(i, j int) {
	xs[i], xs[j] = xs[j], xs[i]
}

func (mdc *MetadataCommand) NeedsBounds() bool {
	return mdc.Bounds
}

func (mdc *MetadataCommand) Perform(
	out io.Writer,
	_ *config.ClusterConfig,
	cluster db.SampleCluster,
	streams sonarlog.InputStreamSet,
	bounds sonarlog.Timebounds,	// for mdc.Bounds only
	hostGlobber *hostglob.HostGlobber,
	_ *db.SampleFilter,
) error {
	if mdc.Times {
		fmt.Fprintf(out, "From: %s\n", mdc.FromDate.Format(time.RFC3339))
		fmt.Fprintf(out, "To:   %s\n", mdc.ToDate.Format(time.RFC3339))
	}

	if mdc.Files {
		if sampleDir, ok := cluster.(db.SampleCluster); ok {
			// For -files, print the full paths all the input files as presented to os.Open.
			files, err := sampleDir.SampleFilenames(mdc.FromDate, mdc.ToDate, hostGlobber)
			if err != nil {
				return err
			}
			for _, name := range files {
				fmt.Fprintln(out, name)
			}
		} else {
			panic("Bad cluster type")
		}
	}

	if mdc.Bounds {
		if mdc.MergeByJob {
			_, bounds = sonarlog.MergeByJob(streams, bounds)
		}

		// Print the bounds
		items := make([]metadataItem, 0)
		for k, v := range bounds {
			items = append(items, metadataItem{
				host:     k.String(),
				earliest: v.Earliest,
				latest:   v.Latest,
			})
		}
		sort.Sort(HostTimeSortableItems(items))
		FormatData(out, mdc.printFields, metadataFormatters, mdc.printOpts, items, metadataCtx(false))
	}

	return nil
}

func (mdc *MetadataCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(mdc.Fmt, metadataHelp, metadataFormatters, metadataAliases, metadataDefaultFields)
}

const metadataHelp = `
metadata
  Compute time bounds and file names for the run.
`

const metadataDefaultFields = "all"

// MT: Constant after initialization; immutable
var metadataAliases = map[string][]string{
	"all": []string{"host", "earliest", "latest"},
}

type metadataCtx bool

// MT: Constant after initialization; immutable
var metadataFormatters = map[string]Formatter[metadataItem, metadataCtx]{
	"host": {
		func(d metadataItem, _ metadataCtx) string {
			return d.host
		},
		"The host name for a sample stream",
	},
	"earliest": {
		func(d metadataItem, _ metadataCtx) string {
			return FormatYyyyMmDdHhMmUtc(d.earliest)
		},
		"The earliest time in a sample stream (yyyy-mm-dd hh:mm)",
	},
	"latest": {
		func(d metadataItem, _ metadataCtx) string {
			return FormatYyyyMmDdHhMmUtc(d.latest)
		},
		"The latest time in a sample stream (yyyy-mm-dd hh:mm)",
	},
}
