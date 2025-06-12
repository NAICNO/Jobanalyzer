// Application logic for analysis of local Sample data.

package application

import (
	"fmt"
	"io"
	"time"

	. "sonalyze/cmd"
	"sonalyze/cmd/jobs"
	. "sonalyze/common"
	"sonalyze/data/sample"
	"sonalyze/db"
	"sonalyze/db/repr"
	"sonalyze/db/special"
)

// Clearly, for `jobs` the file list thing is tricky b/c the list can be *either* sample data *or*
// sacct data, but not both.  We probably need to require it to be sample data.  More generally
// anything requiring a join of two kinds of data will break down with a file list and will need the
// disambiguation *unless* the data come from the same files.

func LocalSampleOperation(command SampleAnalysisCommand, _ io.Reader, stdout, stderr io.Writer) error {
	args := command.SampleAnalysisFlags()
	filter := BuildSampleFilter(command)

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

type FileListJobsDataProvider struct {
	provider db.DataProvider
	isSample bool
}

func (fljdp *FileListJobsDataProvider) ReadSamples(
	fromDate, toDate time.Time,
	hosts *Hosts,
	verbose bool,
) (
	sampleBlobs [][]*repr.Sample,
	softErrors int,
	err error,
) {
	if fljdp.isSample {
		return fljdp.provider.ReadSamples(fromDate, toDate, hosts, verbose)
	}
	return nil, 0, nil
}

func (fljdp *FileListJobsDataProvider) ReadSacctData(
	fromDate, toDate time.Time,
	verbose bool,
) (
	recordBlobs [][]*repr.SacctInfo,
	softErrors int,
	err error,
) {
	if !fljdp.isSample {
		return fljdp.provider.ReadSacctData(fromDate, toDate, verbose)
	}
	return nil, 0, nil
}

var _ = jobs.JobsDataProvider((*FileListJobsDataProvider)(nil))

func LocalJobsOperation(command *jobs.JobsCommand, _ io.Reader, stdout, stderr io.Writer) error {
	args := command.SampleAnalysisFlags()
	filter := BuildSampleFilter(command)
	cfg, err := special.MaybeGetConfig(command.ConfigFile())
	if err != nil {
		return err
	}
	hosts, recordFilter, err := sample.BuildSampleFilter(cfg, filter, args.Verbose)
	if err != nil {
		return fmt.Errorf("Failed to create record filter: %v", err)
	}

	var theLog jobs.JobsDataProvider
	if len(args.LogFiles) > 0 {
		// We default to sample data, fall back to sacct data under a switch.
		fljdp := &FileListJobsDataProvider{}
		theLog = fljdp
		if command.SlurmJobData {
			fljdp.provider, err = db.OpenFileListDB(db.FileListSlurmJobData, args.LogFiles, cfg)
		} else {
			fljdp.isSample = true
			fljdp.provider, err = db.OpenFileListDB(db.FileListSampleData, args.LogFiles, cfg)
		}
	} else {
		if args.DataDir == "" {
			return fmt.Errorf("Must have either dataDir or logFiles")
		}
		theLog, err = db.OpenPersistentDirectoryDB(args.DataDir, cfg)
	}
	if err != nil {
		return fmt.Errorf("Failed to open log store: %v", err)
	}

	return command.Perform(stdout, cfg, theLog, filter, hosts, recordFilter)
}

func BuildSampleFilter(params SampleAnalysisParameters) sample.QueryFilter {
	args := params.SampleAnalysisFlags()

	var filter sample.QueryFilter
	filter.AllUsers, filter.SkipSystemUsers, filter.ExcludeSystemCommands, filter.ExcludeHeartbeat =
		params.DefaultRecordFilters()
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

	return filter
}
