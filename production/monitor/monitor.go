package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"
)

var (
	remote = flag.String("remote", "", "Remote address and port")
	authFile = flag.String("auth-file", "", "Name of local netrc file with authorization")
	mailTo = flag.String("mailto", "", "Address to send information to")
	logDir = flag.String("logdir", "", "Directory in which to log state information")
	limit = flag.Uint("limit", 180, "Minimum number of minutes between reports")
	verbose = flag.Bool("v", false, "Verbose output")
)

const (
	remoteTimeoutSec = 10
)

func main() {
	var err error

	flag.Parse()
	if *remote == "" || *authFile == "" || *mailTo == "" || *logDir == "" {
		fmt.Fprintf(os.Stderr, "Missing option.  Try -h.")
		os.Exit(2)
	}


	ctx, cancel := context.WithTimeout(context.Background(), remoteTimeoutSec*time.Second)
	defer cancel()

	var newStdout, newStderr strings.Builder
	command := exec.CommandContext(ctx, "curl", curlArgs...)
	command.Stdout = &newStdout
	command.Stderr = &newStderr
	if rCmd.VerboseFlag() {
		Log.Infof("Executing <%s>", command.String())
	}
	err = command.Run()
	if err != nil {
		// Oops
	}

	// Basically:
	// - execute curl -g ${remote}/cluster?fmt=csv,cluster
	// - if curl fails, log the failure somehow (TBD)
	// - expect 200, everything else is an error
	// - expect a list with at least one item in it for the output
	//
	// If the tests don't pass:
	// - lookup ${logDir}/naicmonitor-monitor.json, create it if not present
	// - if this fails, panic
	// - otherwise, read the file, it will have a json object with a "last-report" field
	//   that carries the Unix time for the last report sent
	// - if now minus the timestamp is < ${limit} minutes, exit
	// - otherwise:
	//    - try to send mail to $mailTo using the local mailer in non-interactive mode
	//    - if this fails, log the error (TBD)
	//    - otherwise, update the timestamp file
}
