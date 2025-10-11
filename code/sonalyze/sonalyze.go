// `sonalyze` -- Analyze `sonar` log files
//
// See MANUAL.md for a manual, or run `sonalyze help` for brief help.
//
// This code is moderately multi-threaded: There are multiple goroutines in the I/O subsystem, and
// every HTTP handler runs on a separate goroutine as well.  Most components and libraries therefore
// need to be thread-safe (by using locks or being immutable).  The exception to that requirement is
// the individual analysis commands (`jobs`, etc) and the `add` command, which are created in
// response to a request and are themselves only used on a single thread.  Global variables are
// invariably annotated with an `MT: Constraint` comment that documents how the global interacts
// with the thread-safety requirement.
//
// Data are cached by the I/O subsystem.  Cached data are shared (to hold memory usage down) but
// must be regarded as completely immutable, including the slices that point to those data.

package main

import (
	"bufio"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime/pprof"
	"strings"

	"go-utils/status"
	"sonalyze/application"
	"sonalyze/cmd"
	. "sonalyze/common"
	"sonalyze/daemon"
	"sonalyze/db"
	. "sonalyze/table"
)

// See end of file for documentation / implementation, and command/command.go for documentation of
// the CommandLineHandler interface.
//
// MT: Constant after initialization; immutable (no fields)
var stdhandler = standardCommandLineHandler{}

func main() {
	err := sonalyze()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Sonalyze failed: %v\n", err)
		os.Exit(1)
	}
}

func sonalyze() error {
	out := cmd.CLIOutput()

	if len(os.Args) < 2 {
		fmt.Fprintf(out, "Required operation missing, try `sonalyze help`\n")
		os.Exit(2)
	}

	cmdName := os.Args[0]
	maybeVerb := os.Args[1]
	args := os.Args[2:]

	switch maybeVerb {
	case "help", "-h":
		if len(args) > 0 && topicalHelp(out, args[0]) {
			os.Exit(0)
		}
		fmt.Fprintf(out, "Usage: %s command [options] [-- logfile ...]\n\n", cmdName)
		fmt.Fprintf(out, "Commands:\n")
		fmt.Fprintf(out, "  daemon   - spin up a server daemon to process requests\n")
		application.CommandHelp(out)
		fmt.Fprintf(out, "Each command accepts -h to further explain options.\n\n")
		fmt.Fprintf(out, "For help on some other topics, try `sonalyze help <topic>`:\n")
		topicalHelpTopics(out)
		os.Exit(0)

	default:
		anyCmd, verb := stdhandler.ParseVerb(cmdName, maybeVerb)
		if anyCmd == nil {
			fmt.Fprintf(out, "Unknown operation: %s\nTry `sonalyze help`\n", maybeVerb)
			os.Exit(2)
		}

		fs := cmd.NewCLI(maybeVerb, anyCmd, cmdName, true)
		err := stdhandler.ParseArgs(verb, args, anyCmd, fs)
		if err != nil {
			fmt.Fprintf(out, "Bad arguments: %v\nTry `sonalyze %s -h`\n", err, maybeVerb)
			os.Exit(2)
		}

		// All verbose messages are printed with Log.Info so for -v the level has to be at least
		// that low.
		if anyCmd.VerboseFlag() {
			Log.LowerLevelTo(status.LogLevelInfo)
		}

		if fhCmd, ok := anyCmd.(cmd.FormatHelpAPI); ok {
			if h := fhCmd.MaybeFormatHelp(); h != nil {
				PrintFormatHelp(out, h)
				os.Exit(0)
			}
		}

		stop, err := stdhandler.StartCPUProfile(anyCmd.CpuProfileFile())
		if err != nil {
			return err
		}
		if stop != nil {
			defer stop()
		}

		if cmd, ok := anyCmd.(cmd.RemotableCommand); ok && cmd.RemotingFlags().Remoting {
			return application.RemoteOperation(cmd, verb, os.Stdin, os.Stdout, out)
		}

		// We are running against a local cluster store.
		//
		// On return, close all open directories after flushing any pending output, cancel all
		// pending input and return errors from blocked reading operations.
		//
		// Note, we are dependent on nobody calling Exit() after this point.
		defer db.Close()

		return stdhandler.HandleCommand(anyCmd, os.Stdin, os.Stdout, out)
	}
	panic("Unreachable")
}

//go:embed help.txt
var help string

type helpText struct {
	kwd    string
	header string
	text   string
}

var helpTopicRe = regexp.MustCompile(`^#\s+(\S+)\s+-\s*(.*)$`)

func topicalText() []helpText {
	topics := make([]helpText, 0)
	scanner := bufio.NewScanner(strings.NewReader(help))
	var current helpText
	for scanner.Scan() {
		s := scanner.Text()
		if m := helpTopicRe.FindStringSubmatch(s); m != nil {
			if current.kwd != "" {
				topics = append(topics, current)
			}
			current.kwd = m[1]
			current.header = m[2]
			current.text = ""
		} else if current.kwd != "" {
			current.text += "\n  " + s
		}
	}
	if current.kwd != "" {
		topics = append(topics, current)
	}
	return topics
}

func topicalHelp(out io.Writer, what string) bool {
	for _, k := range topicalText() {
		if k.kwd == what {
			fmt.Fprintf(out, "%s:", k.header)
			fmt.Fprintln(out, k.text)
			return true
		}
	}
	return false
}

func topicalHelpTopics(out io.Writer) {
	for _, k := range topicalText() {
		fmt.Fprintf(out, "  %s - %s\n", k.kwd, k.header)
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Command line parsing and execution helpers.

type standardCommandLineHandler struct {
}

func (_ *standardCommandLineHandler) ParseVerb(
	cmdName, maybeVerb string,
) (command cmd.Command, verb string) {
	switch maybeVerb {
	case "daemon":
		command = daemon.New(&daemonCommandLineHandler{})
	default:
		command, maybeVerb = application.ConstructCommand(maybeVerb)
	}
	verb = maybeVerb
	return
}

func (_ *standardCommandLineHandler) ParseArgs(
	verb string,
	args []string,
	command cmd.Command,
	fs *cmd.CLI,
) error {
	command.Add(fs)
	err := fs.Parse(args)
	if err != nil {
		return err
	}

	rest := fs.Args()
	if len(rest) > 0 {
		if lfCmd, ok := command.(cmd.SetRestArgumentsAPI); ok {
			lfCmd.SetRestArguments(rest)
		} else {
			return fmt.Errorf("Rest arguments not accepted by `%s`", verb)
		}
	}

	// Skip validation if the command will provide formatting help and formatting help has been
	// requested.  This is a bit of a hack to avoid Validate() erroring out before help is printed,
	// but it is correct on the assumption that the caller will re-acquire the help message, print
	// it, and exit.
	if fhCmd, ok := command.(cmd.FormatHelpAPI); ok && fhCmd.MaybeFormatHelp() != nil {
		return nil
	}

	return command.Validate()
}

func (_ *standardCommandLineHandler) StartCPUProfile(profileFile string) (func(), error) {
	if profileFile == "" {
		return nil, nil
	}

	f, err := os.Create(profileFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to create profile: %v", err)
	}

	pprof.StartCPUProfile(f)
	return func() { pprof.StopCPUProfile() }, nil
}

func (_ *standardCommandLineHandler) HandleCommand(
	anyCmd cmd.Command,
	stdin io.Reader,
	stdout, stderr io.Writer,
) error {
	// This is the one place in the system where we "open the database", which for now means opening
	// the cluster store.  We can have many clusters in the store.  The cluster store must provide
	// enumeration and lookup and other things, and each "cluster" object also provides services.
	//
	// Since sonalyze can be run against a jobanalyzer directory (containing multiple data types for
	// each of multiple clusters, as well as cluster metadata), against a data directory (containing
	// multiple data types for only one cluster and no metadata for it), or against a file list
	// (containing a single data type for one cluster and no metadata for it), there are various
	// modes for opening the cluster store.  These are the rules:
	//
	// - If there is a -jobanalyzer-dir $DIR then we open this as the multi-cluster store, and
	//   -data-dir, -config-file, and file lists are disallowed, and we read the cluster config data
	//   from that directory.
	//
	// - Otherwise, if there is a -cluster-config then this contains (new-style) cluster
	//   configuration data for the single cluster and we will read the file and define a cluster
	//   according to those data.  These configuration data do not contain node data.
	//
	// - Otherwise, if there is a -config-file then this contains (old-style) cluster configuration
	//   data for the single cluster and we will read the file and define a cluster according to
	//   those data.  These configuration data may contain node data.
	//
	// - Otherwise, we will define a single anonymous cluster.
	//
	// Normal argument parsing applies, which is to say, there can be no -data-dir or file list
	// with -jobanalyzer-dir, and there can be no file list with -data-dir either.
	//
	// When a -jobanalyzer-dir is given, all (old-style) node definitions in the data store are
	// ignored; all node data are read from the database.
	//
	// When a -cluster-config is given (with -data-dir or a file list), all node data will be sought
	// in the database, though in the case of a file list this will only be sensible if the files
	// have sysinfo data - not the common case.
	//
	// When a -config-file is given (with -data-dir or a file list), all node data will be sought in
	// the config data.
	//
	// TODO.
	//
	// Most places in the system use command.DataDir to find the data directory for the cluster in
	// question.  This should probably be removed and should be moved to the meta object, which is
	// the appropriate place to locate cluster-specific data.  Indeed everyone who takes the DataDir
	// value also takes a meta object.  The meta object could supply both the data dir and the file
	// list (LogFiles), which would make the most sense.  Although, the meta object is being
	// constructed in a context where maybe not all commands may have data-dir or file list.  But
	// when we are running from a jobanalyzer-dir, the data-dir is computed (somewhere), not given,
	// and must be attached to the meta object anyway I think.
	//
	// It would be nice to have only one -config-file switch, taking two formats, not two different
	// ones?  Depends on how similar they are.
	//
	// APIs in db/special/cluster.go must change.
	//
	// Probably there will be a JobanalyzerDir accessor on the command like there is a ConfigFile
	// accessor, and we can take it from there.
	//
	// There's an argument to be made for one more mode: where we have -data-dir but instead of a
	// -config-file we have just a cluster config file for a single cluster, and the node data
	// should still be gotten from the directory.  (After all this would be normal.)  This config
	// file is different because it must combine various data sources now in several places, such as
	// aliases from one place and cluster high-level info from another.

	var err error
	var cfg *config.ClusterConfig
	var single *special.ClusterEntry
	if jaArg, ok := anyCmd.(interface { JobanalyzerDir() string }); ok {
		err = special.OpenClusterStore(jaArg.JobanalyzerDir())
	} else if gcArg, ok := anyCmd(interface { ConfigFile() string }); ok {
		cfg, err = special.MaybeGetConfig(gcArg.ConfigFile())
		if err = nil {
			single, err = special.OpenClusterStoreFromConfig(cfg)
		}
	} else {
		single = special.OpenEmptyClusterStore()
	}
	if err != nil {
		return err
	}

	if command, ok := anyCmd.(*daemon.DaemonCommand); ok {
		return command.RunDaemon(stdin, stdout, stderr)
	}

	// What should happen here is not NewMetaFromConfig, but get the cluster store to produce the
	// meta for us.  Probably.  Although it amounts to the same thing.  Probably instead of
	// returning a ClusterEntry above, and retaining the cfg, the "single" above is really the "meta".
	// Then in the daemon code, the ClusterMeta for a cluster is looked up in the cluster store
	// instead of being constructed in any way.

	meta := cmd.NewMetaFromConfig(single, cfg)
	switch command := anyCmd.(type) {
	case cmd.SampleAnalysisCommand:
		return application.LocalSampleOperation(meta, command, stdin, stdout, stderr)
	case cmd.SimpleCommand:
		return command.Perform(meta, stdin, stdout, stderr)
	default:
		return errors.New("NYI command")
	}
	panic("Unreachable")
}

// No profiling, no recursive running of daemon when running commands remotely with `sonalyze daemon`.

type daemonCommandLineHandler struct {
}

func (_ *daemonCommandLineHandler) ParseVerb(
	cmdName, maybeVerb string,
) (command cmd.Command, verb string) {
	if maybeVerb == "daemon" {
		return
	}
	return stdhandler.ParseVerb(cmdName, maybeVerb)
}

func (_ *daemonCommandLineHandler) ParseArgs(
	verb string,
	args []string,
	command cmd.Command,
	fs *cmd.CLI,
) error {
	err := stdhandler.ParseArgs(verb, args, command, fs)
	if err != nil {
		return err
	}
	if command.CpuProfileFile() != "" {
		return errors.New("The -cpuprofile cannot be run remotely")
	}
	return nil
}

func (_ *daemonCommandLineHandler) StartCPUProfile(string) (func(), error) {
	panic("Should not happen")
}

func (_ *daemonCommandLineHandler) HandleCommand(
	anyCmd cmd.Command,
	stdin io.Reader,
	stdout, stderr io.Writer,
) error {
	if _, ok := anyCmd.(*daemon.DaemonCommand); ok {
		panic("Should not happen")
	}
	return stdhandler.HandleCommand(anyCmd, stdin, stdout, stderr)
}
