// Run sonar repeatedly with sensible options
//
// Usage:
//   sonard [-i interval-in-seconds] [-m min-cpu-time-in-seconds] [-v] -s path-to-sonar logfile
//
// TODO: Figure out some way of getting --batchless right.  --batchless is right for the ML nodes but
// would be wrong for systems running Slurm, for example.  It may be that we can look for slurm
// executables to determine whether to include that option or not, and provide some override for
// obscure.  But slurm executables may or may not be available on compute nodes on HPC systems.
//
// TODO: Maybe it should be possible to ask for --rollup, too, or maybe --rollup and !--batchless
// go hand in hand: On HPC systems the option is "--rollup", on the ML nodes it is "--batchless".

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"time"

	"go-utils/process"
	"go-utils/status"
)

const (
	defMinCpu   = 30
	defInterval = 60
	minInterval = 1
)

var (
	interval  = flag.Uint("i", defInterval, "Interval in `seconds` at which to run sonar")
	minCpu    = flag.Uint("m", defMinCpu, "Minimum CPU time consumption in `seconds` for a job before sonar records it")
	sonarName = flag.String("s", "", "Sonar executable `filename`")
	verbose   = flag.Bool("v", false, "Print informational messages")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] output-logfile\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "  output-logfile\n    \tDestination for sonar log records\n")
	}
	flag.Parse()
	if *verbose {
		status.Default().LowerLevelTo(status.LogLevelInfo)
	}
	if *sonarName == "" {
		status.Fatalf("-s is required")
	}
	if *interval < minInterval {
		status.Fatalf("Minimum -i value is %d seconds, have %d", minInterval, *interval)
	}
	rest := flag.Args()
	if len(rest) != 1 {
		status.Fatalf("There must be exactly one logfile argument at the end, I see %v", rest)
	}
	logfileName := path.Clean(rest[0])

	logfile, err := os.OpenFile(logfileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		status.Fatalf("Could not open log file %s for appending", logfileName)
	}
	defer logfile.Close()

	arguments := []string{
		"ps",
		"--exclude-system-jobs",
		"--exclude-commands=bash,ssh,zsh,tmux,systemd",
		"--batchless",
	}
	if *minCpu > 0 {
		arguments = append(arguments, "--min-cpu-time", fmt.Sprint(*minCpu))
	}

	go func() {
		for {
			cmd := exec.Command(*sonarName, arguments...)
			var stderr strings.Builder
			cmd.Stdout = logfile
			cmd.Stderr = &stderr
			if *verbose {
				status.Infof("Running %s %v", *sonarName, arguments)
			}
			err := cmd.Run()
			errout := stderr.String()
			if err != nil || len(errout) != 0 {
				status.Fatalf("Sonar exited with an error\n%v", errors.Join(err, errors.New(errout)))
			}
			time.Sleep(time.Duration(*interval) * time.Second)
		}
	}()

	// Catch sensible signals and terminate normally.
	process.WaitForSignal(syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
}
