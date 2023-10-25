// Run sonar repeatedly with sensible options
//
// Usage:
//   sonard [-i interval-in-seconds] [-m min-cpu-time-in-seconds] -s path-to-sonar logfile
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
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

func main() {
	intervalFlag := flag.Int("i", 60, "Interval (seconds) at which to run sonar")
	minCpuFlag := flag.Int("m", 30, "Minimum CPU time consumption for a job before sonar records it")
	sonarName := flag.String("s", "", "Path to sonar executable")
	verboseFlag := flag.Bool("v", false, "Print informational messages")
	flag.Parse()
	if *intervalFlag < 5 {
		log.Fatalf("Minimum -i value is 5 seconds, have %d", *intervalFlag)
	}
	if *minCpuFlag < 1 {
		log.Fatalf("Minimum -m value is 1 second, have %d", *minCpuFlag)
	}
	if *minCpuFlag >= *intervalFlag {
		*minCpuFlag = *intervalFlag / 2
		if *verboseFlag {
			log.Printf("Adjusting -m value to %d to fit value of -i (which is %d)", *minCpuFlag, *intervalFlag)
		}
	}
	rest := flag.Args()
	if len(rest) != 1 {
		log.Fatalf("There must be exactly one logfile argument at the end, I see %v", rest)
	}
	logfileName := path.Clean(rest[0])

	// There's an assumption here that when this process receives SIGHUP or SIGINT it will not need
	// to catch the signal and specifically close this file; the file will be closed for it, and the
	// data written by the subprocesses will have been written to the file properly.

	logfile, err := os.OpenFile(logfileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Could not open log file %s for appending", logfileName)
	}
	arguments := []string{
		"ps",
		"--exclude-system-jobs",
		"--exclude-commands=bash,ssh,zsh,tmux,systemd",
		"--min-cpu-time", fmt.Sprint(*minCpuFlag),
		"--batchless",
	}
	for {
		cmd := exec.Command(*sonarName, arguments...)
		var stderr strings.Builder
		cmd.Stdout = logfile
		cmd.Stderr = &stderr
		if *verboseFlag {
			log.Printf("Running %s %v", *sonarName, arguments)
		}
		err := cmd.Run()
		errout := stderr.String()
		if err != nil || len(errout) != 0 {
			log.Fatalf("Sonar exited with an error\n%v", errors.Join(err, errors.New(errout)))
		}
		time.Sleep(time.Duration(*intervalFlag) * time.Second)
	}
}
