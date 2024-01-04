// Superstructure for stateful naic reporting.
//
// Run `naicreport help` for help.

package main

import (
	"fmt"
	"os"
	"sort"

	"naicreport/deadweight"
	"naicreport/glance"
	"naicreport/hostnames"
	"naicreport/load"
	"naicreport/mlcpuhog"
)

type command struct {
	help    string
	handler func(arg0 string, args []string) error
}

var commandSummary = "<verb> <option> ... [-- filename ...]"

var commands = map[string]command{
	"at-a-glance": command{
		"Produce a summary report from many parts",
		glance.Report,
	},
	"deadweight": command{
		"Analyze the deadweight logs and generate a report of new violations",
		deadweight.Deadweight,
	},
	"ml-deadweight": command{
		"Obsolete name for \"deadweight\"",
		deadweight.Deadweight,
	},
	"hostnames": command{
		"Analyze the names of log files to generate a list of host names",
		hostnames.Hostnames,
	},
	"load": command{
		"Run sonalyze to generate plottable (JSON) load reports",
		load.Load,
	},
	"ml-webload": command{
		"Obsolete name for \"load\"",
		load.Load,
	},
	"ml-cpuhog": command{
		"Analyze the cpuhog logs and generate a report of new violations",
		mlcpuhog.MlCpuhog,
	},
}

func main() {
	if len(os.Args) < 2 {
		usage(1)
	}
	if entry, found := commands[os.Args[1]]; found {
		err := entry.handler(os.Args[0], os.Args[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "NAICREPORT FAILED\n%v\n\n", err)
			usage(1)
		}
	} else if os.Args[1] == "help" {
		usage(0)
	} else {
		usage(1)
	}
}

func usage(code int) {
	out := os.Stdout
	if code != 0 {
		out = os.Stderr
	}
	fmt.Fprintf(out, "Usage of %s:\n\n  %s %s\n\n", os.Args[0], os.Args[0], commandSummary)
	fmt.Fprintf(out, "where <verb> is one of\n\n")
	entries := make(sort.StringSlice, 0)
	for name, command := range commands {
		entries = append(entries, "  "+name+"\n    "+command.help)
	}
	sort.Sort(entries)
	for _, e := range entries {
		fmt.Fprintln(out, e)
	}
	fmt.Fprintln(out, "\nAll verbs accept -h to print verb-specific help.")
	fmt.Fprintln(out, "Explicit filenames override any -data-dir or -state-dir argument, when sensible.")
	os.Exit(code)
}
