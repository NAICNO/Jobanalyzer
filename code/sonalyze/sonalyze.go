// `sonalyze` -- Analyze `sonar` log files
//
// See MANUAL.md for a manual, or run `sonalyze help` for brief help.

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime/pprof"

	"sonalyze/add"
	. "sonalyze/command"
	"sonalyze/jobs"
	"sonalyze/load"
	"sonalyze/metadata"
	"sonalyze/parse"
	"sonalyze/profile"
	"sonalyze/uptime"
)

// v0.1.0 - translation from Rust
// v0.2.0 - added 'add' verb

const SonalyzeVersion = "0.2.0"

func main() {
	err := sonalyze()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func sonalyze() error {
	anyCmd, verb := commandLine()

	if anyCmd.CpuProfileFile() != "" {
		f, err := os.Create(anyCmd.CpuProfileFile())
		if err != nil {
			return fmt.Errorf("Failed to create profile\n%w", err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if cmd := anyCmd.(RemotableCommand); cmd.RemotingFlags().Remoting {
		return remoteOperation(cmd, verb)
	}

	switch cmd := anyCmd.(type) {
	case AnalysisCommand:
		return localAnalysis(cmd)
	case *add.AddCommand:
		return cmd.AddData()
	default:
		return errors.New("NYI command")
	}
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
		fmt.Fprintf(out, "  add      - add data to the database\n")
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
	case "add":
		cmd = new(add.AddCommand)
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
		restargs := ""
		if _, ok := cmd.(SetRestArgumentsAPI); ok {
			restargs = " [-- logfile ...]"
		}
		fmt.Fprintf(
			out,
			"Usage: %s %s [options]%s\n\n",
			os.Args[0],
			os.Args[1],
			restargs,
		)
		for _, s := range cmd.Summary() {
			fmt.Fprintln(out, "  ", s)
		}
		fmt.Fprintln(out, "\nOptions:\n")
		fs.PrintDefaults()
		if restargs != "" {
			fmt.Fprintf(out, "  logfile ...\n    \tInput data files\n")
		}
	}
	fs.Parse(os.Args[2:])

	rest := fs.Args()
	if len(rest) > 0 {
		if lfCmd, ok := cmd.(SetRestArgumentsAPI); ok {
			lfCmd.SetRestArguments(rest)
		} else {
			fmt.Fprintf(out, "Rest arguments not accepted by `%s`.\n", verb)
			os.Exit(2)
		}
	}

	if fhCmd, ok := cmd.(FormatHelpAPI); ok {
		if h := fhCmd.MaybeFormatHelp(); h != nil {
			PrintFormatHelp(out, h)
			os.Exit(0)
		}
	}

	err := cmd.Validate()
	if err != nil {
		fmt.Fprintf(out, "Bad arguments, try -h\n%v\n", err.Error())
		os.Exit(2)
	}

	return cmd, verb
}
