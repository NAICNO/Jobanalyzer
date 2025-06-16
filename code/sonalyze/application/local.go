// Application logic for analysis of local Sample data.

package application

import (
	"fmt"
	"io"

	. "sonalyze/cmd"
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

	return command.Perform(stdout, cfg, theLog, filter, hosts, recordFilter)
}
