// `exfiltrate` - data forwarder for Sonar data, to run on HPC nodes.
//
// Exfiltrate receives data on stdin, parses it, reformats it for transmission, and forwards it to a
// network agent.  It attempts to do so reliably: if forwarding fails with some recoverable error it
// will retry the sending later.  It will batch outputs when possible.  It will randomly pick a
// sending time in a specified sending window, in order to offload the server without impacting the
// quality or timing of measurement data.  Input formats, transmission formats, and agent addresses
// are controlled by options.
//
// Run with -h for help.
//
// The arguments -cluster, -output, -source, -target, and -window are all mandatory.  -cluster is
// used to name the cluster from which the data comes, if that information is not part of the data
// (a weakness of the current Sonar data).
//
// If the -auth-file option is provided then the file named must provide a user name and password on
// the form username/password, to be used in an HTTP basic authentication header.  If the connection
// is not HTTPS then the password may be intercepted in transit.
//
// For example, with a sending window of 300s:
//
//   sonar ps ... | exfiltrate --window=300 --cluster=ml --source=sonar/csvnamed --output=json --target=https://...
//
// Source formats supported
//
//   "sonar/csvnamed"
//     Mixed reading and heartbeat records from sonar on "csvnamed" format (CSV syntax with
//     name=value syntax for each field)
//
// Output (transmission) formats supported
//
//   "json"
//     The data are transmitted as JSON arrays-of-objects where each object is a reading or
//     heartbeat record.
//
// Target URL schemes: "http", "https"
//
// For HTTP POST transmission, measurement data are posted to <target-address>/sonar-reading and
// heartbeat data are posted to <target-address>/sonar-heartbeat.  The cluster tag is injected as a
// new field "cluster" in all the Sonar records of both kinds.
//
// For future formats and schemes, see README.md.
//
// About logging: Exfiltrate currently logs to stdout/stderr with the expectation that it will be
// run by cron and that there is a sensible MAILTO set up in the crontab to route any output to a
// responsible user.  All output (for runs without -v) will pertain to errors.

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/url"
	"os"
	"time"

	"go-utils/auth"
	"go-utils/httpclient"
	"go-utils/sonarlog"
	"go-utils/status"
)

const (
	// 10 should be more than enough for most HPC systems, which tend to have a few large jobs
	// running per node at any point in time.
	maxRecordsPerMessage = 10

	// The number of attempts and the interval gives the server about a 30 minute interval to come
	// alive in case it's down when the first attempt is made.
	maxAttempts       = 6
	resendIntervalMin = 5
)

var verbose bool

func main() {
	status.Start("jobanalyzer/exfiltrate")

	window, cluster, inputSource, inputType, outputType, authFile, target, err := commandLine()
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

	var readings []*sonarlog.SonarReading
	var heartbeats []*sonarlog.SonarHeartbeat
	if inputSource == "sonar" && inputType == "csvnamed" {
		var badRecords int
		readings, heartbeats, badRecords, err = sonarlog.ParseSonarCsvnamed(bytes.NewReader(bs))
		if err != nil {
			status.Fatalf("Failed to parse input as csvnamed: %v", err)
		}
		if badRecords > 0 {
			status.Infof("Bad records and/or fields: %d", badRecords)
		}
	} else {
		panic("Unexpected input type / source combination")
	}

	if verbose {
		fmt.Printf("Readings: %d, heartbeats: %d\n", len(readings), len(heartbeats))
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

	client := httpclient.NewClient(target, authUser, authPass, maxAttempts, resendIntervalMin, verbose)
	switch outputType {
	case "json":
		// These loops send multiple records together so as to optimize network traffic, but not too
		// many of them, as to keep packet size sensible.  This may be more complexity than it's
		// worth.

		rs := make([]*sonarlog.SonarReading, 0, maxRecordsPerMessage)
		i := 0
		for {
			if len(rs) == cap(rs) || i == len(readings) && len(rs) > 0 {
				buf, err := json.Marshal(&rs)
				if err != nil {
					status.Fatalf("Failed to marshal: %v", err)
				}
				client.PostDataByHttp("/sonar-reading", buf)
				rs = rs[0:0]
			}
			if i == len(readings) {
				break
			}

			// Tag the record with the cluster name if it isn't already tagged.  Remove
			// JSON-unrepresentable Infinity and NaN values, as they appear in some (mostly older)
			// CSV records.
			r := readings[i]
			if r.Cluster == "" {
				r.Cluster = cluster
			}
			r.CpuPct = cleanFloat(r.CpuPct)
			r.GpuPct = cleanFloat(r.GpuPct)
			r.GpuMemPct = cleanFloat(r.GpuMemPct)
			rs = append(rs, r)
			i++
		}

		hs := make([]*sonarlog.SonarHeartbeat, 0, maxRecordsPerMessage)
		i = 0
		for {
			if len(hs) == cap(hs) || i == len(heartbeats) && len(hs) > 0 {
				buf, err := json.Marshal(&hs)
				if err != nil {
					status.Fatalf("Failed to marshal: %v", err)
				}
				client.PostDataByHttp("/sonar-heartbeat", buf)
				hs = hs[0:0]
			}
			if i == len(heartbeats) {
				break
			}
			r := heartbeats[i]
			r.Cluster = cluster
			hs = append(hs, r)
			i++
		}
	default:
		panic("Bad output type")
	}

	// Data for a packet that could not be delivered go into a queue and a re-send attempt is made
	// after some minutes by ProcessRetries.  The exfiltrate process stays alive, but the sonar
	// process that created the data should be able to exit on its own as we've read all the data.
	// Another run of Sonar may start up another exfiltrate meanwhile.  This is OK, as records may
	// arrive out-of-order at the destination.  There is a hard limit on the number of retries per
	// packet, after which the record is dropped on the floor.  The exfiltrate process exits when
	// the retry queue is empty.

	client.ProcessRetries()
}

func cleanFloat(f float64) float64 {
	if math.IsInf(f, 1) {
		return math.MaxFloat64
	}
	if math.IsInf(f, -1) {
		return -math.MaxFloat64
	}
	if math.IsNaN(f) {
		return 0
	}
	return f
}

func commandLine() (
	window int,
	cluster, inputSource, inputType, outputType, authFile string,
	target *url.URL,
	err error,
) {
	flags := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flags.IntVar(&window, "window", 0, "Send data inside a window of this many `seconds`")
	flags.StringVar(&cluster, "cluster", "", "Tag the data as coming from `cluster-name` (required)")
	var sourceArg string
	flags.StringVar(&sourceArg, "source", "", "Assume input data are in this `format` (required)")
	flags.StringVar(&outputType, "output", "", "Transmit data in this `format` (required)")
	var targetArg string
	flags.StringVar(&targetArg, "target", "", "Connect to `url` to upload data (required)")
	flags.StringVar(&authFile, "auth-file", "", "Read upload credentials from `filename`")
	flags.BoolVar(&verbose, "v", false, "Verbose information")
	err = flags.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		os.Exit(0)
	}
	if err != nil {
		return
	}

	if cluster == "" {
		err = errors.New("Argument -cluster is required")
		return
	}

	if sourceArg == "" {
		err = errors.New("Argument -source is required")
		return
	}
	if sourceArg != "sonar/csvnamed" {
		// This is expected to change
		err = errors.New("Unknown --source value")
		return
	}
	inputSource = "sonar"
	inputType = "csvnamed"

	if outputType == "" {
		err = errors.New("Argument -output is required")
		return
	}
	if outputType != "json" {
		// This is expected to change
		err = errors.New("-output must be `json`")
		return
	}

	if targetArg == "" {
		err = errors.New("Argument -target is required")
		return
	}
	// TODO: Validation.  The parser seems to accept pretty much anything.  Probably we require
	// scheme://host:port and no path on the host and no query.  What about userinfo in the host
	// field?
	target, err = url.Parse(targetArg)
	if err != nil || target.Scheme == "" || target.Host == "" || target.Path != "" {
		errmsg := ""
		if err != nil {
			errmsg = fmt.Sprintf(": %v", err)
		}
		err = fmt.Errorf("Failed to parse target URL %s%s", target, errmsg)
	}

	return
}
