// Handle remotable data analysis commands

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"sonalyze/add"
	. "sonalyze/command"
)

func remoteOperation(rCmd RemotableCommand, verb string) error {
	r := NewReifier()
	err := rCmd.ReifyForRemote(&r)
	if err != nil {
		return err
	}

	args := rCmd.RemotingFlags()
	bs, err := os.ReadFile(args.AuthFile)
	if err != nil {
		// Note, file name is redacted
		return errors.New("Failed to read auth file")
	}
	username, password, ok := strings.Cut(strings.TrimSpace(string(bs)), ":")
	if !ok {
		return errors.New("Invalid auth file syntax")
	}

	curlArgs := []string{
		"--silent",
		"--fail-with-body",
	}

	// TODO: IMPROVEME: Using -u is broken as the name/passwd will be in clear text on the command line
	// and visible by `ps`.  Better might be to use --netrc-file, but then we have to generate this
	// file carefully for each invocation, also a sensitive issue, and there would have to be a host
	// name.  (But the underlying problem is that we're using curl and not making the request
	// directly.)

	if username != "" {
		curlArgs = append(curlArgs, "-u", fmt.Sprintf("%s:%s", username, password))
	}

	var stdin io.Reader
	switch cmd := rCmd.(type) {
	case AnalysisCommand:
		curlArgs = append(curlArgs, "--get")
	case *add.AddCommand:
		// This turns into a POST with data coming from the standard DataSource
		var contentType string
		switch {
		case cmd.Sample:
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
		stdin = cmd.DataSource()
	default:
		panic("Unimplemented")
	}
	curlArgs = append(curlArgs, args.Remote+"/"+verb+"?"+r.EncodedArguments())

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	command := exec.CommandContext(ctx, "curl", curlArgs...)
	command.Stdin = stdin

	var stdout, stderr strings.Builder
	command.Stdout = &stdout
	command.Stderr = &stderr

	if rCmd.VerboseFlag() {
		log.Printf("Executing <%s>", command.String())
	}

	err = command.Run()
	if err != nil {
		if rCmd.VerboseFlag() {
			outs := stdout.String()
			if outs != "" {
				fmt.Print(outs)
			}
			errs := stderr.String()
			if errs != "" {
				fmt.Print(errs)
			}
		}
		// Print this unredacted on the assumption that the remote sonalyzed/sonalyze don't
		// reveal anything they shouldn't.
		return err
	}
	errs := stderr.String()
	if errs != "" {
		return errors.New(errs)
	}
	// print, not println, or we end up adding a blank line that confuses consumers
	fmt.Print(stdout.String())
	return nil
}
