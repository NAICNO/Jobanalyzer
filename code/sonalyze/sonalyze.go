// `sonalyze` -- Analyze `sonar` log files
//
// See MANUAL.md for a manual, or run `sonalyze help` for brief help.
//
// This code is moderately multi-threaded: There are multiple goroutines in the I/O subsystem, and
// every HTTP handler runs on a separate goroutine as well.  Most components and libraries therefore
// need to be thread-safe (by using locks or being immutable).  The exception to that requirement is
// the individual analysis commands (`jobs`, etc) and the `add` command, which are created in
// response to a request and are themselves only used on a single thread.
//
// Data are cached by the I/O subsystem.  Cached data are shared (to hold memory usage down) but
// must be regarded as completely immutable, including the slices that point to those data.

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/pprof"

	"go-utils/status"
	"sonalyze/add"
	. "sonalyze/command"
	. "sonalyze/common"
	"sonalyze/daemon"
	"sonalyze/db"
	"sonalyze/jobs"
	"sonalyze/load"
	"sonalyze/metadata"
	"sonalyze/parse"
	"sonalyze/profile"
	"sonalyze/uptime"
)

// v0.1.0 - translation from Rust
// v0.2.0 - added 'add' verb
// v0.3.0 - added 'daemon' verb (integrating sonalyzed into sonalyze), added caching

const SonalyzeVersion = "0.3.0"

// See end of file for documentation.
// MT: Constant after initialization; immutable (no fields)
var stdhandler = standardCommandLineHandler{}

func main() {
	err := sonalyze()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func sonalyze() error {
	out := flag.CommandLine.Output()

	if len(os.Args) < 2 {
		fmt.Fprintf(out, "Required operation missing, try `sonalyze help`\n")
		os.Exit(2)
	}

	cmdName := os.Args[0]
	maybeVerb := os.Args[1]
	args := os.Args[2:]

	switch maybeVerb {
	case "help", "-h":
		fmt.Fprintf(out, "Usage: %s command [options] [-- logfile ...]\n", cmdName)
		fmt.Fprintf(out, "Commands:\n")
		fmt.Fprintf(out, "  add      - add data to the database\n")
		fmt.Fprintf(out, "  daemon   - spin up a server daemon to process requests\n")
		fmt.Fprintf(out, "  jobs     - summarize and filter jobs\n")
		fmt.Fprintf(out, "  load     - print system load across time\n")
		fmt.Fprintf(out, "  metadata - parse data, print stats and metadata\n")
		fmt.Fprintf(out, "  parse    - parse, select and reformat input data\n")
		fmt.Fprintf(out, "  profile  - print the profile of a particular job\n")
		fmt.Fprintf(out, "  uptime   - print aggregated information about system uptime\n")
		fmt.Fprintf(out, "  version  - print information about the program\n")
		fmt.Fprintf(out, "  help     - print this message\n")
		fmt.Fprintf(out, "Each command accepts -h to further explain options.\n")
		os.Exit(0)

	case "version":
		// Must print version on stdout, and the features() thing is required by some tests.
		// "short" indicates that we're only parsing the first 8 fields (v0.6.0 data).
		fmt.Printf("sonalyze-go version(%s) features(short_untagged_sonar_data)\n", SonalyzeVersion)
		os.Exit(0)

	default:
		anyCmd, verb := stdhandler.ParseVerb(cmdName, maybeVerb)
		if anyCmd == nil {
			fmt.Fprintf(out, "Required operation missing, try `sonalyze help`\n")
			os.Exit(2)
		}

		fs := flag.NewFlagSet(cmdName, flag.ExitOnError)
		fs.Usage = func() {
			restargs := ""
			if _, ok := anyCmd.(SetRestArgumentsAPI); ok {
				restargs = " [-- logfile ...]"
			}
			fmt.Fprintf(
				out,
				"Usage: %s %s [options]%s\n\n",
				cmdName,
				maybeVerb,
				restargs,
			)
			for _, s := range anyCmd.Summary() {
				fmt.Fprintln(out, "  ", s)
			}
			fmt.Fprint(out, "\nOptions:\n\n")
			fs.PrintDefaults()
			if restargs != "" {
				fmt.Fprintf(out, "  logfile ...\n    \tInput data files\n")
			}
		}

		err := stdhandler.ParseArgs(verb, args, anyCmd, fs)
		if err != nil {
			fmt.Fprint(out, err.Error())
			os.Exit(2)
		}

		// All verbose messages are printed with Log.Info so for -v the level has to be at least
		// that low.
		if anyCmd.VerboseFlag() {
			Log.LowerLevelTo(status.LogLevelInfo)
		}

		if fhCmd, ok := anyCmd.(FormatHelpAPI); ok {
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

		if cmd, ok := anyCmd.(RemotableCommand); ok && cmd.RemotingFlags().Remoting {
			return remoteOperation(cmd, verb, os.Stdin, os.Stdout, out)
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

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Command line parsing and execution helpers.

type standardCommandLineHandler struct {
}

func (_ *standardCommandLineHandler) ParseVerb(cmdName, maybeVerb string) (cmd Command, verb string) {
	switch maybeVerb {
	case "add":
		cmd = new(add.AddCommand)
	case "daemon":
		cmd = daemon.New(&daemonCommandLineHandler{})
	case "jobs":
		cmd = new(jobs.JobsCommand)
	case "load":
		cmd = new(load.LoadCommand)
	case "meta", "metadata":
		cmd = new(metadata.MetadataCommand)
		maybeVerb = "metadata"
	case "parse":
		cmd = new(parse.ParseCommand)
	case "profile":
		cmd = new(profile.ProfileCommand)
	case "uptime":
		cmd = new(uptime.UptimeCommand)
	default:
		return
	}
	verb = maybeVerb
	return
}

func (_ *standardCommandLineHandler) ParseArgs(verb string, args []string, cmd Command, fs *flag.FlagSet) error {
	cmd.Add(fs)
	err := fs.Parse(args)
	if err != nil {
		return err
	}

	rest := fs.Args()
	if len(rest) > 0 {
		if lfCmd, ok := cmd.(SetRestArgumentsAPI); ok {
			lfCmd.SetRestArguments(rest)
		} else {
			return fmt.Errorf("Rest arguments not accepted by `%s`.\n", verb)
		}
	}

	err = cmd.Validate()
	if err != nil {
		return fmt.Errorf("Bad arguments, try -h\n%w\n", err)
	}

	return nil
}

func (_ *standardCommandLineHandler) StartCPUProfile(profileFile string) (func(), error) {
	if profileFile == "" {
		return nil, nil
	}

	f, err := os.Create(profileFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to create profile\n%w", err)
	}

	pprof.StartCPUProfile(f)
	return func() { pprof.StopCPUProfile() }, nil
}

func (_ *standardCommandLineHandler) HandleCommand(anyCmd Command, stdin io.Reader, stdout, stderr io.Writer) error {
	switch cmd := anyCmd.(type) {
	case AnalysisCommand:
		return localAnalysis(cmd, stdin, stdout, stderr)
	case *add.AddCommand:
		return cmd.AddData(stdin, stdout, stderr)
	case *daemon.DaemonCommand:
		return cmd.RunDaemon(stdin, stdout, stderr)
	default:
		return errors.New("NYI command")
	}
	panic("Unreachable")
}

// No profiling, no recursive running of daemon when running commands remotely with `sonalyze daemon`.

type daemonCommandLineHandler struct {
}

func (_ *daemonCommandLineHandler) ParseVerb(cmdName, maybeVerb string) (cmd Command, verb string) {
	if maybeVerb == "daemon" {
		return
	}
	return stdhandler.ParseVerb(cmdName, maybeVerb)
}

func (_ *daemonCommandLineHandler) ParseArgs(verb string, args []string, cmd Command, fs *flag.FlagSet) error {
	err := stdhandler.ParseArgs(verb, args, cmd, fs)
	if err != nil {
		return err
	}
	if cmd.CpuProfileFile() != "" {
		return fmt.Errorf("The -cpuprofile cannot be run remotely")
	}
	return nil
}

func (_ *daemonCommandLineHandler) StartCPUProfile(string) (func(), error) {
	panic("Should not happen")
}

func (_ *daemonCommandLineHandler) HandleCommand(anyCmd Command, stdin io.Reader, stdout, stderr io.Writer) error {
	if _, ok := anyCmd.(*daemon.DaemonCommand); ok {
		panic("Should not happen")
	}
	return stdhandler.HandleCommand(anyCmd, stdin, stdout, stderr)
}
