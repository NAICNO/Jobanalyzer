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
	"sonalyze/data/cluster"
	"sonalyze/db"
	"sonalyze/db/special"
	. "sonalyze/table"
)

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
		anyCmd, verb := OneShotParseVerb(cmdName, maybeVerb)
		if anyCmd == nil {
			fmt.Fprintf(out, "Unknown operation: %s\nTry `sonalyze help`\n", maybeVerb)
			os.Exit(2)
		}

		fs := cmd.NewCLI(maybeVerb, anyCmd, cmdName, true)
		err := OneShotParseArgs(verb, args, anyCmd, fs)
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

		stop, err := OneShotStartCPUProfile(anyCmd.CpuProfileFile())
		if err != nil {
			return err
		}
		if stop != nil {
			defer stop()
		}

		if anyCmd.Remoting() {
			return application.RemoteOperation(anyCmd, verb, os.Stdin, os.Stdout, out)
		}

		// We are running against a local cluster store.
		//
		// On return, close all open directories after flushing any pending output, cancel all
		// pending input and return errors from blocked reading operations.
		//
		// Note, we are dependent on nobody calling Exit() after this point.
		defer db.Close()

		return OneShotHandleCommand(anyCmd, os.Stdin, os.Stdout, out)
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

func OneShotParseVerb(cmdName, maybeVerb string) (command cmd.Command, verb string) {
	switch maybeVerb {
	case "daemon":
		command = daemon.New(cmd.CommandLineHandler{
			ParseVerb:       DaemonParseVerb,
			ParseArgs:       DaemonParseArgs,
			StartCPUProfile: DaemonStartCPUProfile,
			HandleCommand:   DaemonHandleCommand,
		})
	default:
		command, maybeVerb = application.ConstructCommand(maybeVerb)
	}
	verb = maybeVerb
	return
}

func OneShotParseArgs(
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

func OneShotStartCPUProfile(profileFile string) (func(), error) {
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

func OneShotHandleCommand(
	anyCmd cmd.Command,
	stdin io.Reader,
	stdout, stderr io.Writer,
) (err error) {
	if !anyCmd.Remoting() {
		// Initialize local data store.
		err = cmd.OpenDataStoreFromCommand(anyCmd)
		if err != nil {
			return fmt.Errorf("Could not initialize data store: %v", err)
		}
		if command, ok := anyCmd.(*daemon.DaemonCommand); ok {
			return command.RunDaemon(stdin, stdout, stderr)
		}
	}
	return OneShotHandleSingleCommand(anyCmd, stdin, stdout, stderr)
}

func OneShotHandleSingleCommand(
	anyCmd cmd.Command,
	stdin io.Reader,
	stdout, stderr io.Writer,
) error {
	if command, ok := anyCmd.(cmd.PrimitiveCommand); ok {
		return command.Perform(stdin, stdout, stderr)
	}

	var cluzter *special.ClusterEntry
	if anyCmd.ClusterName() != "" {
		cluzter = special.LookupCluster(anyCmd.ClusterName())
		if cluzter == nil {
			return errors.New("Cluster " + anyCmd.ClusterName() + " not found")
		}
	} else {
		cluzter = special.GetSingleCluster()
		if cluzter == nil {
			return errors.New("No cluster target, and multiple clusters defined")
		}
	}
	meta := cluster.NewMetaFromCluster(cluzter)
	switch command := anyCmd.(type) {
	case cmd.SampleAnalysisCommand:
		return application.LocalSampleOperation(meta, command, stdin, stdout, stderr)
	case cmd.SimpleCommand:
		return command.Perform(meta, stdin, stdout, stderr)
	default:
		return errors.New("NYI command")
	}
}

// No profiling, no recursive running of daemon when running commands remotely with `sonalyze daemon`.

func DaemonParseVerb(cmdName, maybeVerb string) (command cmd.Command, verb string) {
	if maybeVerb == "daemon" {
		return
	}
	return OneShotParseVerb(cmdName, maybeVerb)
}

func DaemonParseArgs(
	verb string,
	args []string,
	command cmd.Command,
	fs *cmd.CLI,
) error {
	err := OneShotParseArgs(verb, args, command, fs)
	if err != nil {
		return err
	}
	if command.CpuProfileFile() != "" {
		return errors.New("The -cpuprofile cannot be run remotely")
	}
	return nil
}

func DaemonStartCPUProfile(string) (func(), error) {
	panic("Should not happen")
}

func DaemonHandleCommand(
	anyCmd cmd.Command,
	stdin io.Reader,
	stdout, stderr io.Writer,
) error {
	if _, ok := anyCmd.(*daemon.DaemonCommand); ok {
		panic("Should not happen")
	}
	return OneShotHandleSingleCommand(anyCmd, stdin, stdout, stderr)
}
