// Application logic for analysis of local Sample data.

package application

import (
	"fmt"
	"io"

	. "sonalyze/cmd"
	. "sonalyze/common"
	"sonalyze/data/sample"
	"sonalyze/db"
)

// Clearly, for `jobs` the file list thing is tricky b/c the list can be *either* sample data *or*
// sacct data, but not both.  We probably need to require it to be sample data.  More generally
// anything requiring a join of two kinds of data will break down with a file list and will need the
// disambiguation *unless* the data come from the same files.

func LocalSampleOperation(command SampleAnalysisCommand, _ io.Reader, stdout, stderr io.Writer) error {
	args := command.SampleAnalysisFlags()

	var filter sample.QueryFilter
	filter.AllUsers, filter.SkipSystemUsers, filter.ExcludeSystemCommands, filter.ExcludeHeartbeat =
		command.DefaultRecordFilters()
	filter.HaveFrom = args.SourceArgs.HaveFrom
	filter.FromDate = args.SourceArgs.FromDate
	filter.HaveTo = args.SourceArgs.HaveTo
	filter.ToDate = args.SourceArgs.ToDate
	filter.Host = args.HostArgs.Host
	filter.ExcludeSystemJobs = args.RecordFilterArgs.ExcludeSystemJobs
	filter.User = args.RecordFilterArgs.User
	filter.ExcludeUser = args.RecordFilterArgs.ExcludeUser
	filter.Command = args.RecordFilterArgs.Command
	filter.ExcludeCommand = args.RecordFilterArgs.ExcludeCommand
	filter.Job = args.RecordFilterArgs.Job
	filter.ExcludeJob = args.RecordFilterArgs.ExcludeJob

	theLog, err := db.OpenReadOnlyDB(
		command.ConfigFile(),
		args.DataDir,
		db.FileListSampleData,
		args.LogFiles,
	)
	if err != nil {
		return err
	}

	cfg := theLog.Config()
	hosts, recordFilter, err := sample.BuildSampleFilter(cfg, filter, args.Verbose)
	if err != nil {
		return fmt.Errorf("Failed to create record filter: %v", err)
	}

	// This is the cut point.  We want to push the reading into the Perform functions and
	// package up current state and pass it to Perform: stdout, cfg, theLog,
	// filter, hosts, recordFilter, err, command.NeedsBounds(), args.Verbose.  Do not
	// pass 9 separate args.  The point would be to allow each Perform to decide what
	// it is that it reads, and how.

	// TODO: Should not be necessary to pass dates and hosts separately here since we're passing the
	// filter and it should have those data.  This is the only call to this function.
	//
	// There's no reason this couldn't just take `filter` and itself compile the filter.  Except,
	// `Perform` is called on hosts and recordFilter, so those need to be exposed somehow.
	streams, bounds, read, dropped, err :=
		sample.ReadSampleStreamsAndMaybeBounds(
			theLog,
			args.FromDate,
			args.ToDate,
			hosts,
			recordFilter,
			command.NeedsBounds(),
			args.Verbose,
		)
	if err != nil {
		return fmt.Errorf("Failed to read log records: %v", err)
	}
	if args.Verbose {
		Log.Infof("%d records read + %d dropped\n", read, dropped)
		UstrStats(stderr, false)
	}

	// why does this need hosts and recordFilter?
	// The filter is used by parse for its special case.
	// The globber is not a globber, just a host set, and it is used by various for
	//   report filtering.
	return command.Perform(stdout, cfg, theLog, streams, bounds, hosts, recordFilter)
}
