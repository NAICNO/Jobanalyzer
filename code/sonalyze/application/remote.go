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

func RemoteOperation(rCmd Command, verb string, stdin io.Reader, stdout, stderr io.Writer) error {
	r := NewArgReifier()
	err := rCmd.ReifyForRemote(&r)
	if err != nil {
		return err
	}

	authFile := rCmd.AuthFile()
	var username, password string
	var netrcFile string // "" if not netrc
	if it := os.Getenv("SONALYZE_AUTH"); it != "" {
		var ok bool
		username, password, ok = strings.Cut(strings.TrimSpace(it), ":")
		if !ok {
			return errors.New("Invalid SONALYZE_AUTH syntax")
		}
	} else if authFile != "" {
		f, err := os.Open(authFile)
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
			netrcFile = authFile
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
	curlArgs = append(curlArgs, rCmd.RemoteHost()+"/"+verb+"?"+r.EncodedArguments())

	if rCmd.VerboseFlag() {
		Log.Infof(
			"NOTE, we will kill the request if no response after %d seconds", remoteTimeoutSec)
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

	// If there is a processing error on the remote end then the server will respond with a 400 code
	// and the text that would otherwise go to stderr, see runSonalyze() in daemon/perform.go.  That
	// translates as a non-nil error with code 22 here, and the error message is on our local
	// stdout.
	//
	// However if there is a problem with contacting the host or other curl failure, then the exit
	// code is king and stderr may have some text.
	//
	// Curl has an elaborate set of exit codes, we could be more precise here but for most remote
	// cases the user would just look them up.

	if err != nil {
		if xe, ok := err.(*exec.ExitError); ok {
			switch xe.ExitCode() {
			case 22:
				return fmt.Errorf("Remote: %s", newStdout.String())
			case 5, 6, 7:
				return fmt.Errorf("Failed to resolve remote host (or proxy).  Exit code %v, stderr=%s",
					xe.ExitCode(), string(xe.Stderr))
			default:
				return fmt.Errorf("Curl problem: exit code %v, stderr=%s", xe.ExitCode(), string(xe.Stderr))
			}
		}
		return fmt.Errorf("Could not start curl: %v", err)
	}

	// print, not println, or we end up adding a blank line that confuses consumers
	fmt.Fprint(stdout, newStdout.String())
	return nil
}
