// Superstructure for stateful naic reporting.
//
// Run `naicreport help` for help.

package main

import (
	"fmt"
	"os"

	"naicreport/glance"
	"naicreport/mlcpuhog"
	"naicreport/mldeadweight"
	"naicreport/mlwebload"
)

func main() {
	if len(os.Args) < 2 {
		toplevelUsage(1)
	}
	var err error
	switch os.Args[1] {
	case "help":
		toplevelUsage(0)

	case "ml-deadweight":
		err = mldeadweight.MlDeadweight(os.Args[0], os.Args[2:])

	case "ml-cpuhog":
		err = mlcpuhog.MlCpuhog(os.Args[0], os.Args[2:])

	case "ml-webload":
		err = mlwebload.MlWebload(os.Args[0], os.Args[2:])

	case "at-a-glance":
		err = glance.Report(os.Args[0], os.Args[2:])

	default:
		toplevelUsage(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n\n", err)
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
  ml-deadweight
   Analyze the deadweight logs and generate a report of new violations
  ml-cpuhog
    Analyze the cpuhog logs and generate a report of new violations
  ml-webload
    Run sonalyze to generate plottable (JSON) load reports

All verbs accept -h to print verb-specific help.
Explicit filenames override any --data-path argument, when sensible
`,
		os.Args[0],
		os.Args[0])
	os.Exit(code)
}
