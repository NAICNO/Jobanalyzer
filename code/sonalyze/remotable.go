// Handle remotable data analysis commands

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"sonalyze/add"
	. "sonalyze/command"
	. "sonalyze/common"
	"sonalyze/sacct"
)

func remoteOperation(rCmd RemotableCommand, verb string, stdin io.Reader, stdout, stderr io.Writer) error {
	r := NewReifier()
	err := rCmd.ReifyForRemote(&r)
	if err != nil {
		return err
	}

	args := rCmd.RemotingFlags()
	var username, password string
	if args.AuthFile != "" {
		bs, err := os.ReadFile(args.AuthFile)
		if err != nil {
			// Note, file name is redacted
			return errors.New("Failed to read auth file")
		}
		var ok bool
		username, password, ok = strings.Cut(strings.TrimSpace(string(bs)), ":")
		if !ok {
			return errors.New("Invalid auth file syntax")
		}
	}

	curlArgs := []string{
		"--silent",
		"--fail-with-body",
		"--location",
	}

	// TODO: IMPROVEME: Using -u is broken as the name/passwd will be in clear text on the command line
	// and visible by `ps`.  Better might be to use --netrc-file, but then we have to generate this
	// file carefully for each invocation, also a sensitive issue, and there would have to be a host
	// name.  (But the underlying problem is that we're using curl and not making the request
	// directly.)

	if username != "" {
		curlArgs = append(curlArgs, "-u", fmt.Sprintf("%s:%s", username, password))
	}

	switch cmd := rCmd.(type) {
	case SampleAnalysisCommand:
		curlArgs = append(curlArgs, "--get")
	case *sacct.SacctCommand:
		curlArgs = append(curlArgs, "--get")
	case *add.AddCommand:
		// This turns into a POST with data coming from the standard DataSource
		var contentType string
		switch {
		case cmd.Sample, cmd.SlurmSacct:
			contentType = "text/csv"
		case cmd.Sysinfo:
			contentType = "application/json"
		default:
			panic("Unknown state of AddCommand")
		}
		curlArgs = append(
			curlArgs,
			"--data-binary", "@-",
			"-H", "Content-Type: "+contentType,
		)
	default:
		panic("Unimplemented")
	}
	curlArgs = append(curlArgs, args.Remote+"/"+verb+"?"+r.EncodedArguments())

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	command := exec.CommandContext(ctx, "curl", curlArgs...)
	command.Stdin = stdin

	var newStdout, newStderr strings.Builder
	command.Stdout = &newStdout
	command.Stderr = &newStderr

	if rCmd.VerboseFlag() {
		Log.Infof("Executing <%s>", command.String())
	}

	err = command.Run()
	if err != nil {
		if rCmd.VerboseFlag() {
			outs := newStdout.String()
			if outs != "" {
				fmt.Fprint(stdout, outs)
			}
			errs := newStderr.String()
			if errs != "" {
				fmt.Fprint(stdout, errs)
			}
		}
		// Print this unredacted on the assumption that the remote sonalyzed/sonalyze don't
		// reveal anything they shouldn't.
		return err
	}
	errs := newStderr.String()
	if errs != "" {
		return errors.New(errs)
	}
	// print, not println, or we end up adding a blank line that confuses consumers
	fmt.Fprint(stdout, newStdout.String())
	return nil
}
