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
	"sonalyze/common"
	"sonalyze/sonarlog"
)

type MetadataCommand struct /* implements Command */ {
	SharedArgs
	MergeByHostAndJob bool // Inert, but compatible
	MergeByJob        bool
	Times             bool
	Files             bool
	Fmt               string

	// Synthesized and other
	printFields []string
	printOpts   *FormatOptions
}

func (mdc *MetadataCommand) ConfigFile() string {
	return ""
}

func (mdc *MetadataCommand) Add(fs *flag.FlagSet) {
	mdc.SharedArgs.Add(fs)

	fs.BoolVar(&mdc.MergeByHostAndJob, "merge-by-host-and-job", false,
		"Merge streams that have the same host and job ID")
	fs.BoolVar(&mdc.MergeByJob, "merge-by-job", false,
		"Merge streams that have the same job ID, across hosts")
	fs.BoolVar(&mdc.Files, "files", false, "List selected files")
	fs.BoolVar(&mdc.Times, "times", false, "Show from/to timestamps")
	fs.StringVar(&mdc.Fmt, "fmt", "",
		"Select `field,...` and format for the output [default: try -fmt=help]")
}

func (mdc *MetadataCommand) ReifyForRemote(x *Reifier) error {
	e1 := mdc.SharedArgs.ReifyForRemote(x)
	x.Bool("merge-by-host-and-job", mdc.MergeByHostAndJob)
	x.Bool("merge-by-job", mdc.MergeByJob)
	x.Bool("files", mdc.Files)
	x.Bool("times", mdc.Times)
	x.String("fmt", mdc.Fmt)
	return e1
}

func (mdc *MetadataCommand) Validate() error {
	e1 := mdc.SharedArgs.Validate()

	var e2 error
	spec := metadataDefaultFields
	if mdc.Fmt != "" {
		spec = mdc.Fmt
	}
	var others map[string]bool
	mdc.printFields, others, e2 = ParseFormatSpec(spec, metadataFormatters, metadataAliases)
	if e2 == nil && len(mdc.printFields) == 0 {
		e2 = errors.New("No output fields were selected in format string")
	}
	mdc.printOpts = StandardFormatOptions(others)
	// TODO: Defaulting like this is done many places, common?
	if !mdc.printOpts.Fixed && !mdc.printOpts.Csv && !mdc.printOpts.Json && !mdc.printOpts.Awk {
		mdc.printOpts.Csv = true
		mdc.printOpts.Header = false
	}

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

func (mdc *MetadataCommand) Perform(
	out io.Writer,
	_ *config.ClusterConfig,
	logStore *sonarlog.LogStore,
	samples sonarlog.SampleStream,
	hostGlobber *hostglob.HostGlobber,
	recordFilter func(*sonarlog.Sample) bool,
) error {
	if mdc.Times {
		fmt.Printf("From: %s\n", mdc.FromDate.Format(time.RFC3339))
		fmt.Printf("To:   %s\n", mdc.ToDate.Format(time.RFC3339))
	}

	if mdc.Files {
		// For -files, print the full paths all the input files as presented to os.Open.
		files, err := logStore.Files()
		if err != nil {
			return err
		}
		for _, name := range files {
			fmt.Println(name)
		}
	}

	bounds := sonarlog.ComputeTimeBounds(samples)
	if mdc.MergeByJob {
		streams := sonarlog.PostprocessLog(samples, recordFilter, nil)
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

var metadataAliases = map[string][]string{
	"all": []string{"host", "earliest", "latest"},
}

type metadataCtx bool

var metadataFormatters = map[string]Formatter[metadataItem, metadataCtx]{
	"host": {
		func(d metadataItem, _ metadataCtx) string {
			return d.host
		},
		"The host name for a sample stream",
	},
	"earliest": {
		func(d metadataItem, _ metadataCtx) string {
			return common.FormatYyyyMmDdHhMmUtc(d.earliest)
		},
		"The earliest time in a sample stream (yyyy-mm-dd hh:mm)",
	},
	"latest": {
		func(d metadataItem, _ metadataCtx) string {
			return common.FormatYyyyMmDdHhMmUtc(d.latest)
		},
		"The latest time in a sample stream (yyyy-mm-dd hh:mm)",
	},
}
