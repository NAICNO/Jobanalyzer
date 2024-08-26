// Handle local and remote data analysis commands

package main

import (
	"fmt"
	"io"
	"os"

	"go-utils/config"
	"go-utils/hostglob"
	. "sonalyze/command"
	. "sonalyze/common"
	"sonalyze/db"
	"sonalyze/profile"
	"sonalyze/sonarlog"
)

func localAnalysis(cmd SampleAnalysisCommand, _ io.Reader, stdout, stderr io.Writer) error {
	args := cmd.SharedFlags()

	cfg, err := MaybeGetConfig(cmd.ConfigFile())
	if err != nil {
		return err
	}

	hostGlobber, recordFilter, err := buildFilters(cmd, cfg)
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
		sonarlog.ReadSampleStreams(
			theLog,
			args.FromDate,
			args.ToDate,
			hostGlobber,
			args.Verbose,
		)
	if err != nil {
		return fmt.Errorf("Failed to read log records\n%w", err)
	}
	if args.Verbose {
		Log.Infof("%d records read + %d dropped\n", read, dropped)
		UstrStats(stderr, false)
	}

	sonarlog.ComputeAndFilter(streams, recordFilter)
	err = cmd.Perform(stdout, cfg, theLog, streams, bounds, hostGlobber, recordFilter)

	if err != nil {
		return fmt.Errorf("Failed to perform operation\n%w", err)
	}

	return nil
}

func buildFilters(
	cmd SampleAnalysisCommand,
	cfg *config.ClusterConfig,
) (*hostglob.HostGlobber, func(*sonarlog.Sample) bool, error) {
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
		if name := os.Getenv("LOGNAME"); name != "" {
			includeUsers[StringToUstr(name)] = true
		} else {
			return nil, nil, fmt.Errorf("Not able to determine user, none given and $LOGNAME is empty")
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

	// Record filtering logic is the same for all commands.

	excludeSystemJobs := args.RecordFilterArgs.ExcludeSystemJobs
	haveFrom := args.SourceArgs.HaveFrom
	haveTo := args.SourceArgs.HaveTo
	from := args.SourceArgs.FromDate.Unix()
	to := args.SourceArgs.ToDate.Unix()
	recordFilter := func(e *sonarlog.Sample) bool {
		return (len(includeUsers) == 0 || includeUsers[e.S.User]) &&
			(includeHosts.IsEmpty() || includeHosts.Match(e.S.Host.String())) &&
			(len(includeJobs) == 0 || includeJobs[e.S.Job]) &&
			(len(includeCommands) == 0 || includeCommands[e.S.Cmd]) &&
			!excludeUsers[e.S.User] &&
			!excludeJobs[e.S.Job] &&
			!excludeCommands[e.S.Cmd] &&
			(!excludeSystemJobs || e.S.Pid >= 1000) &&
			(!haveFrom || from <= e.S.Timestamp) &&
			(!haveTo || e.S.Timestamp <= to)
	}

	return includeHosts, recordFilter, nil
}
