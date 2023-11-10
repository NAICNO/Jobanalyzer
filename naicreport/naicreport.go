// Superstructure for stateful naic reporting.
//
// Run `naicreport help` for help.

package main

import (
	"fmt"
	"os"

	"naicreport/glance"
	"naicreport/load"
	"naicreport/mlcpuhog"
	"naicreport/deadweight"
)

func main() {
	if len(os.Args) < 2 {
		toplevelUsage(1)
	}
	var err error
	switch os.Args[1] {
	case "help":
		toplevelUsage(0)

	case "deadweight-ingest":
		deadweight.Ingest(os.Args[0], os.Args[2:])

	case "deadweight-report":
		deadweight.Report(os.Args[0], os.Args[2:])

	case "ml-deadweight":
		err = deadweight.MlDeadweight(os.Args[0], os.Args[2:])

	case "ml-cpuhog":
		err = mlcpuhog.MlCpuhog(os.Args[0], os.Args[2:])

	case "ml-webload", "load":
		err = load.Load(os.Args[0], os.Args[2:])

	case "at-a-glance":
		err = glance.Report(os.Args[0], os.Args[2:])

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

  help
    Print help
  at-a-glance
    Produce a summary report from many parts
  deadweight-ingest
    Run sonalyze and ingest new deadweight data into databases
  deadweight-report
    Report all unreported deadweight jobs and purge old data
  load
    Run sonalyze to generate plottable (JSON) load reports
  ml-deadweight
   Analyze the deadweight logs and generate a report of new violations
  ml-cpuhog
    Analyze the cpuhog logs and generate a report of new violations
  ml-webload
    Obsolete name for "load"

All verbs accept -h to print verb-specific help.
Explicit filenames override any --data-path argument, when sensible
`,
		os.Args[0],
		os.Args[0])
	os.Exit(code)
}
