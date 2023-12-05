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
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"

	"go-utils/auth"
	"go-utils/sonarlog"
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
	window, cluster, inputSource, inputType, outputType, authFile, target := commandLine()

	var err error
	var authUser, authPass string
	if authFile != "" {
		authUser, authPass, err = auth.ParseAuth(authFile)
		if err != nil {
			log.Fatalf("Failed to read authentication file: %v", err)
		}
	}

	bs, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("Failed to read from stdin: %v", err)
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
			log.Fatalf("Failed to parse input as csvnamed: %v", err)
		}
		if badRecords > 0 {
			log.Printf("Bad records and/or fields: %d", badRecords)
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
		log.Fatalf("Only http / https targets for now")
	}

	switch outputType {
	case "json":
		// These loops send multiple records together so as to optimize network traffic.

		targetStr := target.String() + "/sonar-reading"
		rs := make([]*sonarlog.SonarReading, 0, maxRecordsPerMessage)
		i := 0
		for {
			if len(rs) == cap(rs) || i == len(readings) && len(rs) > 0 {
				buf, err := json.Marshal(&rs)
				if err != nil {
					log.Fatalf("Failed to marshal: %v", err)
				}
				postDataByHTTP(0, buf, authUser, authPass, targetStr)
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

		targetStr = target.String() + "/sonar-heartbeat"
		hs := make([]*sonarlog.SonarHeartbeat, 0, maxRecordsPerMessage)
		i = 0
		for {
			if len(hs) == cap(hs) || i == len(heartbeats) && len(hs) > 0 {
				buf, err := json.Marshal(&hs)
				if err != nil {
					log.Fatalf("Failed to marshal: %v", err)
				}
				postDataByHTTP(0, buf, authUser, authPass, targetStr)
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

	processRetries()
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

func postDataByHTTP(prevAttempts int, buf []byte, authUser, authPass, target string) {
	if verbose {
		fmt.Printf("Trying to send %s\n", string(buf))
	}

	// Go down a level from http.Post() in order to be able to set authentication header.
	req, err := http.NewRequest("POST", target, bytes.NewReader(buf))
	if err != nil {
		log.Printf("Failed to post: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if authUser != "" {
		req.SetBasicAuth(authUser, authPass)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// There doesn't seem to be any good way to determine that a host is currently unreachable
		// vs all sorts of other errors that can happen along the way.  So when a sending error
		// occurs, always retry.
		if prevAttempts+1 <= maxAttempts {
			addRetry(prevAttempts+1, buf, authUser, authPass, target)
		} else {
			log.Printf("Failed to post to %s after max retries: %v", target, err)
		}
		return
	}

	if verbose {
		fmt.Printf("Response %s\n", resp.Status)
	}

	// Codes in the 200 range indicate everything is OK, for now.
	// Really we should expect
	//  202 (StatusAccepted) for when a new record is created
	//  208 (StatusAlreadyReported) for when the record is a dup
	//
	// TODO: Possibly for codes in the 500 range we should retry?
	if resp.StatusCode >= 300 {
		log.Printf("Failed to post: HTTP status=%d", resp.StatusCode)
		// Fall through: must read response body
	}

	// API requires that we read and close the body
	_, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
}

// Retries are a bit crude.  Data for a packet we could not deliver go into a queue and a re-send
// attempt is made after some minutes.  The exfiltrate process stays alive, but the sonar process
// that created the data should be able to exit on its own as we've read all the data.  Another run
// of Sonar may start up another exfiltrate meanwhile.  This is OK, as records may arrive
// out-of-order at the destination.  There is a hard limit on the number of retries, after which the
// record is dropped on the floor.  The process exits when the retry queue is empty.

type retry struct {
	prevAttempts       int    // number of attempts that have been performed
	buf                []byte // the content
	authUser, authPass string // authorization
	target             string // target address
}

var retries = make([]retry, 0)

func processRetries() {
	for len(retries) > 0 {
		time.Sleep(resendIntervalMin * time.Minute)
		rs := retries
		retries = make([]retry, 0)
		for _, r := range rs {
			postDataByHTTP(r.prevAttempts, r.buf, r.authUser, r.authPass, r.target)
		}
	}
}

func addRetry(prevAttempts int, buf []byte, authUser, authPass, target string) {
	retries = append(retries, retry{prevAttempts, buf, authUser, authPass, target})
}

func commandLine() (
	window int,
	cluster, inputSource, inputType, outputType, authFile string,
	target *url.URL,
) {
	flag.IntVar(&window, "window", -1, "Sending window in seconds")
	flag.StringVar(&cluster, "cluster", "", "Name of cluster")
	sourceArg := flag.String("source", "", "Source and format (eg `sonar/csvnamed`)")
	flag.StringVar(&outputType, "output", "", "Format of output (transmitted data)")
	targetArg := flag.String("target", "", "Target address")
	flag.StringVar(&authFile, "auth-file", "", "Authentication file")
	flag.BoolVar(&verbose, "v", false, "Verbose information")
	flag.Parse()

	if window < 0 {
		badArg("Argument -window is required")
	}

	if cluster == "" {
		badArg("Argument -cluster is required")
	}

	if *sourceArg == "" {
		badArg("Argument -source is required")
	}
	if *sourceArg != "sonar/csvnamed" {
		// This is expected to change
		badArg("Unknown --source value")
	}
	inputSource = "sonar"
	inputType = "csvnamed"

	if outputType == "" {
		badArg("Argument -output is required")
	}
	if outputType != "json" {
		// This is expected to change
		badArg("-output must be `json`")
	}

	if *targetArg == "" {
		badArg("Argument -target is required")
	}
	// TODO: Validation.  The parser seems to accept pretty much anything.  Probably we require
	// scheme://host:port and no path on the host and no query.  What about userinfo in the host
	// field?
	target, err := url.Parse(*targetArg)
	if err != nil || target.Scheme == "" || target.Host == "" || target.Path != "" {
		errmsg := ""
		if err != nil {
			errmsg = fmt.Sprintf(": %v", err)
		}
		badArg(fmt.Sprintf("Failed to parse target URL %s%s", target, errmsg))
	}

	return
}

func badArg(msg string) {
	fmt.Fprintln(os.Stderr, msg+"\nTry -h")
	os.Exit(1)
}
