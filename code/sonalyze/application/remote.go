// Application logic for analysis of remote data.

package application

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	. "sonalyze/cmd"
	"sonalyze/cmd/add"
	. "sonalyze/common"
)

const (
	// Some reports created for naicreport take a *really* long time to produce, for example the
	// 90-day load report for all of Fox takes several minutes, even against cached data.  It's a
	// little open if we even want a timeout.  But 1h seems like it is an OK compromise for now.
	//
	// TODO: This timeout will do nothing for the job running on the server, if any.  It could be
	// running for days, for all we know.  It may be that server-side actions should have a timeout,
	// too, and it could be that the client should always send some sort of cancellation signal to
	// the server if the client is cancelled (SIGHUP, etc, not just timeout).  Simpler would be if
	// we could reduce the running times of these reports to something sensible...
	remoteTimeoutSec = 3600
)

var netrc = regexp.MustCompile(`^machine\s+\S+\s+login\s+\S+\s+password\s+\S+\s*$`)

func RemoteOperation(rCmd RemotableCommand, verb string, stdin io.Reader, stdout, stderr io.Writer) error {
	r := NewArgReifier()
	err := rCmd.ReifyForRemote(&r)
	if err != nil {
		return err
	}

	args := rCmd.RemotingFlags()
	var username, password string
	var netrcFile string // "" if not netrc
	if args.AuthFile != "" {
		f, err := os.Open(args.AuthFile)
		if err != nil {
			// Note, file name is redacted
			return errors.New("Failed to open auth file")
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		lines := make([]string, 0)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err = scanner.Err(); err != nil {
			return errors.New("Failed to read auth file")
		}
		if len(lines) != 1 {
			return errors.New("Auth file must have exactly one line")
		}
		if netrc.MatchString(lines[0]) {
			netrcFile = args.AuthFile
		} else {
			var ok bool
			username, password, ok = strings.Cut(strings.TrimSpace(lines[0]), ":")
			if !ok {
				return errors.New("Invalid auth file syntax")
			}
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

	if netrcFile != "" {
		curlArgs = append(curlArgs, "--netrc-file", netrcFile)
	} else if username != "" {
		curlArgs = append(curlArgs, "-u", fmt.Sprintf("%s:%s", username, password))
	}

	switch cmd := rCmd.(type) {
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
	case SampleAnalysisCommand:
		curlArgs = append(curlArgs, "--get")
	case SimpleCommand:
		curlArgs = append(curlArgs, "--get")
	default:
		panic("Unimplemented")
	}
	curlArgs = append(curlArgs, args.Remote+"/"+verb+"?"+r.EncodedArguments())

	if rCmd.VerboseFlag() {
		Log.Infof("NOTE, we will kill the request if no response after %s seconds", remoteTimeoutSec)
	}
	ctx, cancel := context.WithTimeout(context.Background(), remoteTimeoutSec*time.Second)
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
				fmt.Fprintf(stdout, "Output from failed (%s) subprocess: %s", err, outs)
			}
			errs := newStderr.String()
			if errs != "" {
				fmt.Fprintf(stdout, "Errors from failed (%s) subprocess: %s", err, errs)
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
