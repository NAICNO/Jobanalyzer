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

	cfg, err := MaybeGetConfig(cmd.ConfigFile())
	if err != nil {
		return err
	}

	hostGlobber, recordFilter, err := buildFilters(cmd, cfg, args.Verbose)
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

type filterFunc = func (*sonarlog.Sample) bool

func buildFilters(
	cmd SampleAnalysisCommand,
	cfg *config.ClusterConfig,
	verbose bool,
) (*hostglob.HostGlobber, filterFunc, error) {
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

	// Record filtering logic is the same for all commands.  The filter is applied to every record
	// and can be quite expensive.  The typical case for most attributes is that there is no filter,
	// and for the rest that there is one value to filter by (one job, one user, one host).
	// Probably the common case for from/to is that they are present.  It may be better to use a
	// closure style here.

	excludeSystemJobs := args.RecordFilterArgs.ExcludeSystemJobs
	haveFrom := args.SourceArgs.HaveFrom
	haveTo := args.SourceArgs.HaveTo
	from := args.SourceArgs.FromDate.Unix()
	to := args.SourceArgs.ToDate.Unix()

	var recordFilter filterFunc
	// Time filter must be first, it does not call `next`, it's an optimization.  The others must
	// consider that `next` may be nil.
	if haveFrom || haveTo {
		if !haveFrom {
			from = 0
		}
		if !haveTo {
			to = math.MaxInt64
		}
		recordFilter = filterByTimeRange(from, to)
	}
	recordFilter = filterByUser(includeUsers, true, recordFilter)
	recordFilter = filterByHost(includeHosts, true, recordFilter)
	recordFilter = filterByJob(includeJobs, true, recordFilter)
	recordFilter = filterByCommand(includeCommands, true, recordFilter)
	recordFilter = filterByUser(excludeUsers, false, recordFilter)
	recordFilter = filterByJob(excludeJobs, false, recordFilter)
	recordFilter = filterByCommand(excludeCommands, false, recordFilter)
	if excludeSystemJobs {
		recordFilter = filterSystemJobs(recordFilter)
	}

	/*
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
	*/

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

func filterByTimeRange(from, to int64) filterFunc {
	return func (e *sonarlog.Sample) bool {
		return from <= e.S.Timestamp && e.S.Timestamp <= to
	}
}

func filterByUser(users map[Ustr]bool, answer bool, next filterFunc) filterFunc {
	if len(users) == 0 {
		return next
	}
	// TODO: Maybe optimize for the 1 case, see below for an example.
	return func(e *sonarlog.Sample) bool {
		if _, found := users[e.S.User]; found != answer {
			return false
		}
		return next == nil || next(e)
	}
}

func filterByJob(jobs map[uint32]bool, answer bool, next filterFunc) filterFunc {
	switch len(jobs) {
	case 0:
		return next
	case 1:
		var theJob uint32
		for j := range(jobs) {
			theJob = j
		}
		return func(e *sonarlog.Sample) bool {
			if (e.S.Job == theJob) != answer {
				return false
			}
			return next == nil || next(e)
		}
	default:
		return func(e *sonarlog.Sample) bool {
			if _, found := jobs[e.S.Job]; found != answer {
				return false
			}
			return next == nil || next(e)
		}
	}
}

func filterByHost(hosts *hostglob.HostGlobber, answer bool, next filterFunc) filterFunc {
	if hosts.IsEmpty() {
		return next
	}
	return func (e *sonarlog.Sample) bool {
		if hosts.Match(e.S.Host.String()) != answer {
			return false
		}
		return next == nil || next(e)
	}
}

func filterByCommand(commands map[Ustr]bool, answer bool, next filterFunc) filterFunc {
	if len(commands) == 0 {
		return next
	}
	// TODO: Maybe optimize for the 1 case, see below for an example.
	return func(e *sonarlog.Sample) bool {
		if _, found := commands[e.S.Cmd]; found != answer {
			return false
		}
		return next == nil || next(e)
	}
}

func filterSystemJobs(next filterFunc) filterFunc {
	return func(e *sonarlog.Sample) bool {
		if e.S.Pid < 1000 {
			return false
		}
		return next == nil || next(e)
	}
}
