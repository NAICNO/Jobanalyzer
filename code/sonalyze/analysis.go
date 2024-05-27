// Handle local and remote data analysis commands

package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"go-utils/config"
	"go-utils/hostglob"
	. "sonalyze/command"
	"sonalyze/profile"
	"sonalyze/sonarlog"
)

func localAnalysis(cmd AnalysisCommand, _ io.Reader, stdout, stderr io.Writer) error {
	args := cmd.SharedFlags()

	// TODO: Instead of requiring every cmd to have ConfigFile(), we could introduce a ConfigFileAPI
	// interface and test if the command responds to that.
	var cfg *config.ClusterConfig
	if configName := cmd.ConfigFile(); configName != "" {
		var err error
		cfg, err = config.ReadConfig(configName)
		if err != nil {
			return err
		}
	}

	hostGlobber, recordFilter, err := buildFilters(cmd, cfg)
	if err != nil {
		return fmt.Errorf("Failed to create record filter\n%w", err)
	}

	var theLog *sonarlog.LogDir
	if len(args.LogFiles) > 0 {
		theLog, err = sonarlog.OpenFiles(args.LogFiles)
	} else {
		theLog, err = sonarlog.OpenDir(args.DataDir)
	}
	if err != nil {
		return fmt.Errorf("Failed to open log store\n%w", err)
	}
	samples, dropped, err := theLog.ReadLogEntries(args.FromDate, args.ToDate, hostGlobber, args.Verbose)
	if err != nil {
		return fmt.Errorf("Failed to read log records\n%w", err)
	}
	if args.Verbose {
		log.Printf("%d records read + %d dropped\n", len(samples), dropped)
		sonarlog.UstrStats(stderr, false)
	}

	err = cmd.Perform(stdout, cfg, theLog, samples, hostGlobber, recordFilter)
	if err != nil {
		return fmt.Errorf("Failed to perform operation\n%w", err)
	}

	return nil
}

func buildFilters(
	cmd AnalysisCommand,
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

	includeHosts := hostglob.NewGlobber(true)
	for _, h := range args.RecordFilterArgs.Host {
		err := includeHosts.Insert(h)
		if err != nil {
			return nil, nil, err
		}
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

	includeUsers := make(map[sonarlog.Ustr]bool)
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
				includeUsers[sonarlog.StringToUstr(u)] = true
			}
		}
	} else if allUsers {
		// Everyone, so do nothing
	} else {
		if name := os.Getenv("LOGNAME"); name != "" {
			includeUsers[sonarlog.StringToUstr(name)] = true
		} else {
			return nil, nil, fmt.Errorf("Not able to determine user, none given and $LOGNAME is empty")
		}
	}

	// Excluded users.

	excludeUsers := make(map[sonarlog.Ustr]bool)
	for _, u := range args.RecordFilterArgs.ExcludeUser {
		excludeUsers[sonarlog.StringToUstr(u)] = true
	}

	if skipSystemUsers {
		// This list needs to be configurable somehow, but isn't so in the Rust version either.
		excludeUsers[sonarlog.StringToUstr("root")] = true
		excludeUsers[sonarlog.StringToUstr("zabbix")] = true
	}

	// Included commands.

	includeCommands := make(map[sonarlog.Ustr]bool)
	for _, command := range args.RecordFilterArgs.Command {
		includeCommands[sonarlog.StringToUstr(command)] = true
	}

	// Excluded commands.

	excludeCommands := make(map[sonarlog.Ustr]bool)
	for _, command := range args.RecordFilterArgs.ExcludeCommand {
		excludeCommands[sonarlog.StringToUstr(command)] = true
	}

	if excludeSystemCommands {
		// This list needs to be configurable somehow, but isn't so in the Rust version either.
		excludeCommands[sonarlog.StringToUstr("bash")] = true
		excludeCommands[sonarlog.StringToUstr("zsh")] = true
		excludeCommands[sonarlog.StringToUstr("sshd")] = true
		excludeCommands[sonarlog.StringToUstr("tmux")] = true
		excludeCommands[sonarlog.StringToUstr("systemd")] = true
	}

	// Skip heartbeat records?  It's probably OK to filter only by command name, since we're
	// currently doing full-command-name matching.

	if excludeHeartbeat {
		excludeCommands[sonarlog.StringToUstr("_heartbeat_")] = true
	}

	// System configuration additions, if available

	if cfg != nil {
		for _, user := range cfg.ExcludeUser {
			excludeUsers[sonarlog.StringToUstr(user)] = true
		}
	}

	// Record filtering logic is the same for all commands.

	excludeSystemJobs := args.RecordFilterArgs.ExcludeSystemJobs
	haveFrom := args.SourceArgs.HaveFrom
	haveTo := args.SourceArgs.HaveTo
	from := args.SourceArgs.FromDate.Unix()
	to := args.SourceArgs.ToDate.Unix()
	recordFilter := func(e *sonarlog.Sample) bool {
		return (len(includeUsers) == 0 || includeUsers[e.User]) &&
			(includeHosts.IsEmpty() || includeHosts.Match(e.Host.String())) &&
			(len(includeJobs) == 0 || includeJobs[e.Job]) &&
			(len(includeCommands) == 0 || includeCommands[e.Cmd]) &&
			!excludeUsers[e.User] &&
			!excludeJobs[e.Job] &&
			!excludeCommands[e.Cmd] &&
			(!excludeSystemJobs || e.Pid >= 1000) &&
			(!haveFrom || from <= e.Timestamp) &&
			(!haveTo || e.Timestamp <= to)
	}

	return includeHosts, recordFilter, nil
}
