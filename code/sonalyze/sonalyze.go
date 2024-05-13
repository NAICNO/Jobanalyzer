// `sonalyze` -- Analyze `sonar` log files
//
// See MANUAL.md for a manual, or run `sonalyze help` for brief help.

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime/pprof"
	"strings"
	"time"

	"go-utils/config"
	"go-utils/hostglob"
	. "sonalyze/command"
	"sonalyze/jobs"
	"sonalyze/load"
	"sonalyze/metadata"
	"sonalyze/parse"
	"sonalyze/profile"
	"sonalyze/sonarlog"
	"sonalyze/uptime"
)

const SonalyzeVersion = "0.1.0"

func main() {
	err := sonalyze()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func sonalyze() error {
	cmd, verb := commandLine()
	args := cmd.Args()

	if args.CpuProfile != "" {
		f, err := os.Create(args.CpuProfile)
		if err != nil {
			return fmt.Errorf("Failed to create profile\n%w", err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if args.Remoting {
		return sonalyzeRemotely(cmd, verb)
	}

	return sonalyzeLocally(cmd)
}

func sonalyzeRemotely(cmd Command, verb string) error {
	r := NewReifier()
	err := cmd.ReifyForRemote(&r)
	if err != nil {
		return err
	}

	args := cmd.Args()
	bs, err := os.ReadFile(args.AuthFile)
	if err != nil {
		// Note, file name is redacted
		return errors.New("Failed to read auth file")
	}
	username, password, ok := strings.Cut(strings.TrimSpace(string(bs)), ":")
	if !ok {
		return errors.New("Invalid auth file syntax")
	}

	// TODO: FIXME: Using -u is broken as the name/passwd will be in clear text on the command line
	// and visible by `ps`.  Better might be to use --netrc-file, but then we have to generate this
	// file carefully for each invocation, also a sensitive issue, and there would have to be a host
	// name.

	cmdArgs := []string{
		"-s",
		"--get",
		args.Remote + "/" + verb + "?" + r.EncodedArguments(),
	}
	if username != "" {
		cmdArgs = append(cmdArgs, "-u", fmt.Sprintf("%s:%s", username, password))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	command := exec.CommandContext(ctx, "curl", cmdArgs...)

	var stdout, stderr strings.Builder
	command.Stdout = &stdout
	command.Stderr = &stderr

	if args.Verbose {
		log.Printf("Executing <%s>", command.String())
	}

	err = command.Run()
	if err != nil {
		// Print this unredacted on the assumption that the remote sonalyzed/sonalyze don't
		// reveal anything they shouldn't.
		return err
	}
	errs := stderr.String()
	if errs != "" {
		return errors.New(errs)
	}
	// print, not println, or we end up adding a blank line that confuses consumers
	fmt.Print(stdout.String())
	return nil
}

func sonalyzeLocally(cmd Command) error {
	args := cmd.Args()

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

	var theLog *sonarlog.LogStore
	if len(args.LogFiles) > 0 {
		theLog, err = sonarlog.OpenFiles(args.LogFiles)
	} else {
		theLog, err = sonarlog.OpenDir(args.DataDir, args.FromDate, args.ToDate, hostGlobber)
	}
	if err != nil {
		return fmt.Errorf("Failed to open log store\n%w", err)
	}
	samples, dropped, err := theLog.ReadLogEntries(args.Verbose)
	if err != nil {
		return fmt.Errorf("Failed to read log records\n%w", err)
	}
	if args.Verbose {
		log.Printf("%d records read + %d dropped\n", len(samples), dropped)
		sonarlog.UstrStats(os.Stderr, false)
	}

	err = cmd.Perform(os.Stdout, cfg, theLog, samples, hostGlobber, recordFilter)
	if err != nil {
		return fmt.Errorf("Failed to perform operation\n%w", err)
	}

	return nil
}

func buildFilters(
	cmd Command,
	cfg *config.ClusterConfig,
) (*hostglob.HostGlobber, func(*sonarlog.Sample) bool, error) {
	args := cmd.Args()

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

func commandLine() (Command, string) {
	out := flag.CommandLine.Output()

	if len(os.Args) < 2 {
		fmt.Fprintf(out, "Required operation missing, try `sonalyze help`\n")
		os.Exit(2)
	}

	var cmd Command
	var verb = os.Args[1]
	switch verb {
	case "help", "-h":
		fmt.Fprintf(out, "Usage: %s command [options] [-- logfile ...]\n", os.Args[0])
		fmt.Fprintf(out, "Commands:\n")
		fmt.Fprintf(out, "  help     - print this message\n")
		fmt.Fprintf(out, "  jobs     - summarize and filter jobs\n")
		fmt.Fprintf(out, "  load     - print system load across time\n")
		fmt.Fprintf(out, "  metadata - parse data, print stats and metadata\n")
		fmt.Fprintf(out, "  parse    - parse, select and reformat input data\n")
		fmt.Fprintf(out, "  profile  - print the profile of a particular job\n")
		fmt.Fprintf(out, "  uptime   - print aggregated information about system uptime\n")
		fmt.Fprintf(out, "  version  - print information about the program\n")
		fmt.Fprintf(out, "Each command accepts -h to further explain options.\n")
		os.Exit(0)
	case "jobs":
		cmd = new(jobs.JobsCommand)
	case "load":
		cmd = new(load.LoadCommand)
	case "meta", "metadata":
		cmd = new(metadata.MetadataCommand)
		verb = "metadata"
	case "parse":
		cmd = new(parse.ParseCommand)
	case "profile":
		cmd = new(profile.ProfileCommand)
	case "uptime":
		cmd = new(uptime.UptimeCommand)
	case "version":
		// Must print version on stdout, and the features() thing is required by some tests.
		// "short" indicates that we're only parsing the first 8 fields (v0.6.0 data).
		fmt.Printf("sonalyze-go version(%s) features(short_untagged_sonar_data)\n", SonalyzeVersion)
		os.Exit(0)
	default:
		fmt.Fprintf(out, "Required operation missing, try `sonalyze help`\n")
		os.Exit(2)
	}

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cmd.Add(fs)

	fs.Usage = func() {
		fmt.Fprintf(
			flag.CommandLine.Output(),
			"Usage: %s %s [options] [-- logfile ...]\nOptions:\n",
			os.Args[0],
			os.Args[1],
		)
		fs.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "  logfile ...\n    \tInput data files\n")
	}
	fs.Parse(os.Args[2:])
	// Trailing arguments are assumed to be files, even if they did not follow `--`.  Since we're
	// just going to open those names for reading it's OK not to vet the values further.
	cmd.Args().LogFiles = fs.Args()

	if h := cmd.MaybeFormatHelp(); h != nil {
		PrintFormatHelp(out, h)
		os.Exit(0)
	}

	err := cmd.Validate()
	if err != nil {
		fmt.Fprintf(out, "Bad arguments, try -h\n%v\n", err.Error())
		os.Exit(2)
	}

	return cmd, verb
}
