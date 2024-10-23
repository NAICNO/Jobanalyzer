// Handle local and remote data analysis commands

package main

import (
	"fmt"
	"io"
	"math"
	"os"

	"go-utils/config"
	"go-utils/hostglob"
	"go-utils/maps"
	"go-utils/slices"
	. "sonalyze/command"
	. "sonalyze/common"
	"sonalyze/db"
	"sonalyze/profile"
	"sonalyze/sonarlog"
)

func localAnalysis(cmd SampleAnalysisCommand, _ io.Reader, stdout, stderr io.Writer) error {
	args := cmd.SharedFlags()

	cfg, err := db.MaybeGetConfig(cmd.ConfigFile())
	if err != nil {
		return err
	}

	hostGlobber, recordFilter, err := buildRecordFilters(cmd, cfg, args.Verbose)
	if err != nil {
		return fmt.Errorf("Failed to create record filter\n%w", err)
	}

	var theLog db.SampleCluster
	if len(args.LogFiles) > 0 {
		theLog, err = db.OpenTransientSampleCluster(args.LogFiles, cfg)
	} else {
		theLog, err = db.OpenPersistentCluster(args.DataDir, cfg)
	}
	if err != nil {
		return fmt.Errorf("Failed to open log store\n%w", err)
	}

	streams, bounds, read, dropped, err :=
		sonarlog.ReadSampleStreamsAndMaybeBounds(
			theLog,
			args.FromDate,
			args.ToDate,
			hostGlobber,
			recordFilter,
			cmd.NeedsBounds(),
			args.Verbose,
		)
	if err != nil {
		return fmt.Errorf("Failed to read log records\n%w", err)
	}
	if args.Verbose {
		Log.Infof("%d records read + %d dropped\n", read, dropped)
		UstrStats(stderr, false)
	}

	sonarlog.ComputePerSampleFields(streams)
	err = cmd.Perform(stdout, cfg, theLog, streams, bounds, hostGlobber, recordFilter)

	if err != nil {
		return fmt.Errorf("Failed to perform operation\n%w", err)
	}

	return nil
}

func buildRecordFilters(
	cmd SampleAnalysisCommand,
	cfg *config.ClusterConfig,
	verbose bool,
) (*hostglob.HostGlobber, *db.SampleFilter, error) {
	args := cmd.SharedFlags()

	// Temporary limitation.

	if _, ok := cmd.(*profile.ProfileCommand); ok {
		if len(args.RecordFilterArgs.Job) != 1 || len(args.RecordFilterArgs.ExcludeJob) != 0 {
			return nil, nil, fmt.Errorf("Exactly one specific job number is required by `profile`")
		}
	}

	// Included host set, empty means "all"

	includeHosts, err := hostglob.NewGlobber(true, args.RecordFilterArgs.Host)
	if err != nil {
		return nil, nil, err
	}

	// Included job numbers, empty means "all"

	includeJobs := make(map[uint32]bool)
	for _, j := range args.RecordFilterArgs.Job {
		includeJobs[j] = true
	}

	// Excluded job numbers.

	excludeJobs := make(map[uint32]bool)
	for _, j := range args.RecordFilterArgs.ExcludeJob {
		excludeJobs[j] = true
	}

	// Command-specific defaults for the record filters.

	allUsers, skipSystemUsers, excludeSystemCommands, excludeHeartbeat := cmd.DefaultRecordFilters()

	// Included users, empty means "all"

	includeUsers := make(map[Ustr]bool)
	if len(args.RecordFilterArgs.User) > 0 {
		allUsers = false
		for _, u := range args.RecordFilterArgs.User {
			if u == "-" {
				allUsers = true
				break
			}
		}
		if !allUsers {
			for _, u := range args.RecordFilterArgs.User {
				includeUsers[StringToUstr(u)] = true
			}
		}
	} else if allUsers {
		// Everyone, so do nothing
	} else {
		// LOGNAME is Posix but may be limited to a user being "logged in"; USER is BSD and a bit
		// more general supposedly.  We prefer the former but will settle for the latter, in
		// particular, Github actions has only USER.
		if name := os.Getenv("LOGNAME"); name != "" {
			includeUsers[StringToUstr(name)] = true
		} else if name := os.Getenv("USER"); name != "" {
			includeUsers[StringToUstr(name)] = true
		} else {
			return nil, nil, fmt.Errorf("Not able to determine user, none given and $LOGNAME and $USER are empty")
		}
	}

	// Excluded users.

	excludeUsers := make(map[Ustr]bool)
	for _, u := range args.RecordFilterArgs.ExcludeUser {
		excludeUsers[StringToUstr(u)] = true
	}

	if skipSystemUsers {
		// This list needs to be configurable somehow, but isn't so in the Rust version either.
		excludeUsers[StringToUstr("root")] = true
		excludeUsers[StringToUstr("zabbix")] = true
	}

	// Included commands.

	includeCommands := make(map[Ustr]bool)
	for _, command := range args.RecordFilterArgs.Command {
		includeCommands[StringToUstr(command)] = true
	}

	// Excluded commands.

	excludeCommands := make(map[Ustr]bool)
	for _, command := range args.RecordFilterArgs.ExcludeCommand {
		excludeCommands[StringToUstr(command)] = true
	}

	if excludeSystemCommands {
		// This list needs to be configurable somehow, but isn't so in the Rust version either.
		excludeCommands[StringToUstr("bash")] = true
		excludeCommands[StringToUstr("zsh")] = true
		excludeCommands[StringToUstr("sshd")] = true
		excludeCommands[StringToUstr("tmux")] = true
		excludeCommands[StringToUstr("systemd")] = true
	}

	// Skip heartbeat records?  It's probably OK to filter only by command name, since we're
	// currently doing full-command-name matching.

	if excludeHeartbeat {
		excludeCommands[StringToUstr("_heartbeat_")] = true
	}

	// System configuration additions, if available

	if cfg != nil {
		for _, user := range cfg.ExcludeUser {
			excludeUsers[StringToUstr(user)] = true
		}
	}

	// Record filtering logic is the same for all commands.  The record filter can use only raw
	// ingested data, it can be applied at any point in the pipeline.  It *must* be thread-safe.

	excludeSystemJobs := args.RecordFilterArgs.ExcludeSystemJobs
	haveFrom := args.SourceArgs.HaveFrom
	haveTo := args.SourceArgs.HaveTo
	var from int64 = 0
	if haveFrom {
		from = args.SourceArgs.FromDate.Unix()
	}
	var to int64 = math.MaxInt64
	if haveTo {
		to = args.SourceArgs.ToDate.Unix()
	}
	var minPid uint32
	if excludeSystemJobs {
		minPid = 1000
	}

	var recordFilter = &db.SampleFilter{
		IncludeUsers:    includeUsers,
		IncludeHosts:    includeHosts,
		IncludeJobs:     includeJobs,
		IncludeCommands: includeCommands,
		ExcludeUsers:    excludeUsers,
		ExcludeJobs:     excludeJobs,
		ExcludeCommands: excludeCommands,
		MinPid:          minPid,
		From:            from,
		To:              to,
	}

	if verbose {
		if haveFrom {
			Log.Infof("Including records starting on or after %s", args.SourceArgs.FromDate)
		}
		if haveTo {
			Log.Infof("Including records ending on or before %s", args.SourceArgs.ToDate)
		}
		if len(includeUsers) > 0 {
			Log.Infof("Including records with users %s", maps.Values(includeUsers))
		}
		if !includeHosts.IsEmpty() {
			Log.Infof("Including records with hosts matching %s", includeHosts)
		}
		if len(includeJobs) > 0 {
			Log.Infof(
				"Including records with job id matching %s",
				slices.Map(maps.Keys(includeJobs), func(x uint32) string {
					return fmt.Sprint(x)
				}))
		}
		if len(includeCommands) > 0 {
			Log.Infof("Including records with commands matching %s", maps.Keys(includeCommands))
		}
		if len(excludeUsers) > 0 {
			Log.Infof("Excluding records with users matching %s", maps.Keys(excludeUsers))
		}
		if len(excludeJobs) > 0 {
			Log.Infof(
				"Excluding records with job ids matching %s",
				slices.Map(maps.Keys(excludeJobs), func(x uint32) string {
					return fmt.Sprint(x)
				}))
		}
		if len(excludeCommands) > 0 {
			Log.Infof("Excluding records with commands matching %s", maps.Keys(excludeCommands))
		}
		if excludeSystemJobs {
			Log.Infof("Excluding records with PID < 1000")
		}
	}

	return includeHosts, recordFilter, nil
}
