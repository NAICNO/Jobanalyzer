// `exfiltrate` - data forwarder for all sorts of data, to run on HPC nodes.
//
// Exfiltrate receives data on stdin and forwards it to a network service, with some features:
//
// - Input is obtained first: Exfiltrate reads all its input before trying to send, so the
//   originator of the data will be allowed to exit before exfiltrate is finished.
// - Forwarding is reliable: if forwarding fails with some recoverable error, exfiltrate will
//   retry the sending later.
// - Sending is randomized: Exfiltrate will randomly pick a sending time in a specified sending
//   window, in order to distribute load on the server without impacting the quality or timing of
//   measurement data.
//
// In truth, though, this is just `( sleep( ((RANDOM % $window)) ); curl --data-binary @- ... )`
// with a fairly fancy set of retry options to curl, and it may eventually be replaced by just such
// a script.
//
// Usage:
//
//   exfiltrate [options] target-url
//
//   The target-url is mandatory: It is an HTTP or HTTPS URL, including path and arguments, to which
//   data are sent by POST.
//
// The mimetype can and usually should be specified:
//
// -mimetype <type>
//   The default type is text/plain.
//
// Sending can be randomized:
//
// -window <seconds>
//   Specifies the sending window in seconds, a random time within the window is chosen for the first
//   sending attempt.  The default window is 0.
//
// Network transport can be HTTPS:
//
// -ca-cert <filename>
//   The URL must be an HTTPS URL, and the argument is a filename holding the certificate for a
//   Certificate Authority that exfiltrate will use to validate the identity of the server.
//
// Upload can be authenticated:
//
// -auth-file <filename>
//   The file named must provide a user name and password on the form username:password, to be used
//   in an HTTP basic authentication header.  If the connection is not HTTPS then the password may
//   be intercepted in transit.
//
// There is online help:
//
// -h
//   Print help
//
// Debugging options:
//
// -v
//   Verbose diagnostics
//
// For example:
//
//   sonar ps ... | \
//      exfiltrate -window 300 -ca-cert secrets/server-ca.crt -auth-file my-password.txt \
//          https://naic-monitor.uio.no/sonar-freecsv?cluster=mlx.hpc.uio.no
//
// About logging:
//
// Exfiltrate currently logs to stdout/stderr with the expectation that it will be run by cron and
// that there is a sensible MAILTO set up in the crontab to route any output to a responsible user.
// All output (for runs without -v) will pertain to errors.
//
// TODO: Could parameterize the resend attempts and interval.

package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"os"
	"time"

	"go-utils/auth"
	"go-utils/httpclient"
	"go-utils/status"
)

const (
	// The number of attempts and the interval gives the server about a 30 minute interval to come
	// alive in case it's down when the first attempt is made.
	maxAttempts       = 6
	resendIntervalMin = 5
)

// Command line parameters
var (
	window     int
	mimetype   string
	authFile   string
	caCertFile string
	target     *url.URL
	verbose    bool
)

func main() {
	status.Start("jobanalyzer/exfiltrate")

	err := commandLine()
	if err != nil {
		status.Fatalf("Command line: %v", err)
	}

	var authUser, authPass string
	if authFile != "" {
		authUser, authPass, err = auth.ParseAuth(authFile)
		if err != nil {
			status.Fatalf("Failed to read authentication file: %v", err)
		}
	}

	bs, err := io.ReadAll(os.Stdin)
	if err != nil {
		status.Fatalf("Failed to read from stdin: %v", err)
	}
	if verbose {
		fmt.Printf("Bytes of data read: %d\n", len(bs))
	}

	if window > 0 {
		secs := rand.Intn(window)
		if verbose {
			fmt.Printf("Sleeping %d seconds\n", secs)
		}
		time.Sleep(time.Duration(secs) * time.Second)
	}

	if target.Scheme != "http" && target.Scheme != "https" {
		status.Fatal("Only http / https targets for now")
	}
	if target.Scheme == "https" && caCertFile == "" {
		status.Fatal("HTTPS requires a -ca-cert")
	}
	if target.Scheme == "http" && caCertFile != "" {
		status.Fatal("HTTP needs no -ca-cert; did you mean HTTPS?")
	}

	client, err := httpclient.NewClient(target, caCertFile, authUser, authPass, maxAttempts, resendIntervalMin, verbose)
	if err != nil {
		status.Fatalf("Failed to create client: %v", err)
	}
	client.PostDataByHttp("", mimetype, bs)
	client.ProcessRetries()
}

func commandLine() error {
	flags := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flags.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] target-url\nOptions:\n", os.Args[0])
		flags.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "  target-url\n    \tDestination for data\n")
	}
	flags.StringVar(&mimetype, "mimetype", "text/plain", "The `mime type` of the data")
	flags.IntVar(&window, "window", 0, "Send data inside a window of this many `seconds`")
	flags.StringVar(&authFile, "auth-file", "", "Read upload credentials from `filename`")
	flags.StringVar(&caCertFile, "ca-cert", "", "Connect over HTTPS with CA cert `filename`")
	flags.BoolVar(&verbose, "v", false, "Verbose information")
	err := flags.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		os.Exit(0)
	}
	if err != nil {
		return err
	}
	if len(flags.Args()) != 1 {
		return fmt.Errorf("Exactly one target argument is required")
	}
	targetArg := flags.Args()[0]
	target, err = url.Parse(targetArg)
	if err != nil || target.Scheme == "" || target.Host == "" {
		errmsg := ""
		if err != nil {
			errmsg = fmt.Sprintf(": %v", err)
		}
		return fmt.Errorf("Failed to parse target URL %s%s", target, errmsg)
	}

	return nil
}
