// hostnames - compute a set of host names on a cluster.
//
// The output is a json array of strings, lexicographically sorted.
//
// NOTE: Clients should no longer use the output from this, but instead go directly to sonalyze and
// run eg `sonalyze node -remote ... -cluster ... -auth-file ... -from 14d -newest -fmt csv,host` to
// get an unsorted list of host names, one per line (or ask for JSON and get the "host" field from
// each object of the resulting array).
//
// End-user options:
//
//  -remote url
//  -auth-file filename
//    Required: The server (with optional authorization) that will serve data to us
//
//  -cluster clustername
//    Required: the cluster for which we want information from the server
//
//  -sonalyze filename
//    The `sonalyze` executable.
//
// Debugging / development options:
//
//  -- filename ...
//    Test input files - these must be sysinfo-*.json type files.
//
//  -v
//    Print various (verbose) debugging output

package hostnames

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"

	"go-utils/options"
	"go-utils/process"

	"naicreport/util"
)

func Hostnames(progname string, args []string) (err error) {
	opts := flag.NewFlagSet(progname+" hostnames", flag.ContinueOnError)
	sonalyzePath := opts.String("sonalyze", "", "Sonalyze executable `filename` (required)")
	remoteOpts := util.AddRemoteOptions(opts)
	verbose := opts.Bool("v", false, "Verbose (debugging) output")
	err = opts.Parse(os.Args[2:])
	if err == flag.ErrHelp {
		os.Exit(0)
	}
	if err != nil {
		return
	}
	testFiles := opts.Args()
	if len(testFiles) > 0 {
		if remoteOpts.Server != "" || remoteOpts.Cluster != "" {
			err = errors.New("Can't combine remote options and local test files")
		} else {
			testFiles, err = util.CleanRestArgs(testFiles)
		}
	} else {
		if remoteOpts.Server == "" || remoteOpts.Cluster == "" {
			err = errors.New("Both -remote and -cluster are required")
		}
	}
	if err != nil {
		return
	}
	*sonalyzePath, err = options.RequireCleanPath(*sonalyzePath, "-sonalyze")
	if err != nil {
		return
	}

	nodeArgs := []string{"node", "-from", "14d", "-newest", "-fmt", "csv,host"}
	if *verbose {
		nodeArgs = append(nodeArgs, "-v")
	}
	// Files must come last.
	if len(testFiles) > 0 {
		nodeArgs = append(nodeArgs, "--")
		nodeArgs = append(nodeArgs, testFiles...)
	} else {
		nodeArgs = util.ForwardRemoteOptions(nodeArgs, remoteOpts)
	}
	if *verbose {
		fmt.Fprintf(os.Stderr, "Sonalyze node arguments\n%v\n", nodeArgs)
	}
	hnOutput, hnErrOutput, err := process.RunSubprocess(
		"sonalyze",
		*sonalyzePath,
		nodeArgs,
	)
	if err != nil {
		if hnErrOutput != "" {
			err = errors.Join(err, fmt.Errorf("With stderr:\n%s", hnErrOutput))
		}
	}
	if *verbose && hnErrOutput != "" {
		fmt.Fprint(os.Stderr, hnErrOutput)
	}
	hosts := strings.Fields(hnOutput)
	slices.Sort(hosts)
	encoding, err := json.Marshal(hosts)
	if err == nil {
		fmt.Println(string(encoding))
	}
	return
}
