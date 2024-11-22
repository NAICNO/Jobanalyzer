package metadata

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"time"

	"go-utils/config"
	"go-utils/hostglob"
	uslices "go-utils/slices"

	. "sonalyze/cmd"
	"sonalyze/db"
	"sonalyze/sonarlog"
	. "sonalyze/table"
)

type MetadataCommand struct /* implements SampleAnalysisCommand */ {
	SharedArgs
	FormatArgs

	MergeByHostAndJob bool // Inert, but compatible
	MergeByJob        bool
	Times             bool
	Files             bool
	Bounds            bool
}

var _ SampleAnalysisCommand = (*MetadataCommand)(nil)

func (_ *MetadataCommand) Summary() []string {
	return []string{
		"Display metadata about the sample streams in the database.",
		"One or more of -files, -times and -bounds must be selected to produce",
		"output.",
	}
}

func (mdc *MetadataCommand) Add(fs *CLI) {
	mdc.SharedArgs.Add(fs)
	mdc.FormatArgs.Add(fs)

	fs.Group("aggregation")
	fs.BoolVar(&mdc.MergeByHostAndJob, "merge-by-host-and-job", false,
		"Merge streams that have the same host and job ID")
	fs.BoolVar(&mdc.MergeByJob, "merge-by-job", false,
		"Merge streams that have the same job ID, across hosts")

	fs.Group("operation-selection")
	fs.BoolVar(&mdc.Files, "files", false, "List selected files")
	fs.BoolVar(&mdc.Times, "times", false, "Show parsed from/to timestamps")
	fs.BoolVar(&mdc.Bounds, "bounds", false, "Show host with earliest/latest timestamp")
}

func (mdc *MetadataCommand) ReifyForRemote(x *ArgReifier) error {
	e1 := errors.Join(
		mdc.SharedArgs.ReifyForRemote(x),
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
		mdc.SharedArgs.Validate(),
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

type metadataItem struct {
	Hostname string        `alias:"host"     desc:"Name that host is known by on the cluster"`
	Earliest DateTimeValue `alias:"earliest" desc:"Timestamp of earliest sample for host"`
	Latest   DateTimeValue `alias:"latest"   desc:"Timestamp of latest sample for host"`
}

type HostTimeSortableItems []*metadataItem

func (xs HostTimeSortableItems) Len() int {
	return len(xs)
}

func (xs HostTimeSortableItems) Less(i, j int) bool {
	if xs[i].Hostname == xs[j].Hostname {
		return xs[i].Earliest < xs[j].Earliest
	}
	return xs[i].Hostname < xs[j].Hostname
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
	bounds sonarlog.Timebounds, // for mdc.Bounds only
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

		// Print the bounds.
		items := make([]*metadataItem, 0)
		for k, v := range bounds {
			items = append(items, &metadataItem{
				Hostname: k.String(),
				Earliest: DateTimeValue(v.Earliest),
				Latest:   DateTimeValue(v.Latest),
			})
		}
		sort.Sort(HostTimeSortableItems(items))
		FormatData(
			out,
			mdc.PrintFields,
			metadataFormatters,
			mdc.PrintOpts,
			uslices.Map(items, func(x *metadataItem) any { return x }),
		)
	}

	return nil
}

func (mdc *MetadataCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(
		mdc.Fmt, metadataHelp, metadataFormatters, metadataAliases, metadataDefaultFields)
}

const metadataHelp = `
metadata
  Compute time bounds and file names for the run.
`

const v0MetadataDefaultFields = "host,earliest,latest"
const v1MetadataDefaultFields = "Hostname,Earliest,Latest"
const metadataDefaultFields = v0MetadataDefaultFields

// MT: Constant after initialization; immutable
var metadataAliases = map[string][]string{
	"all":       []string{"host", "earliest", "latest"},
	"default":   strings.Split(metadataDefaultFields, ","),
	"v0default": strings.Split(v0MetadataDefaultFields, ","),
	"v1default": strings.Split(v1MetadataDefaultFields, ","),
}

// MT: Constant after initialization; immutable
var metadataFormatters = DefineTableFromTags(reflect.TypeFor[metadataItem](), nil)
