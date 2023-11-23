// Superstructure for stateful naic reporting.
//
// Run `naicreport help` for help.

package main

import (
	"fmt"
	"os"

	"naicreport/deadweight"
	"naicreport/glance"
	"naicreport/hostnames"
	"naicreport/load"
	"naicreport/mlcpuhog"
)

func main() {
	if len(os.Args) < 2 {
		toplevelUsage(1)
	}
	var err error
	switch os.Args[1] {
	case "help":
		toplevelUsage(0)

	case "at-a-glance":
		err = glance.Report(os.Args[0], os.Args[2:])

	case "deadweight", "ml-deadweight":
		err = deadweight.Deadweight(os.Args[0], os.Args[2:])

	case "hostnames":
		err = hostnames.Hostnames(os.Args[0], os.Args[2:])

	case "load", "ml-webload":
		err = load.Load(os.Args[0], os.Args[2:])

	case "ml-cpuhog":
		err = mlcpuhog.MlCpuhog(os.Args[0], os.Args[2:])

	default:
		toplevelUsage(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "NAICREPORT FAILED\n%v\n\n", err)
		toplevelUsage(1)
	}
}

func toplevelUsage(code int) {
	fmt.Fprintf(
		os.Stderr,
		`Usage of %s:
  %s <verb> <option> ... [-- filename ...]

where <verb> is one of

  at-a-glance
    Produce a summary report from many parts
  deadweight
    Analyze the deadweight logs and generate a report of new violations
  help
    Print help
  hostnames
	Analyze the names of log files to generate a list of host names
  load
    Run sonalyze to generate plottable (JSON) load reports
  ml-cpuhog
    Analyze the cpuhog logs and generate a report of new violations
  ml-deadweight
    Obsolete name for "deadweight"
  ml-webload
    Obsolete name for "load"

All verbs accept -h to print verb-specific help.
Explicit filenames override any --data-path argument, when sensible
`,
		os.Args[0],
		os.Args[0])
	os.Exit(code)
}
