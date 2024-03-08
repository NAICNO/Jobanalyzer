// `infiltrate` - data receiver for node data, to run on database/analysis host.
//
// Infiltrate receives data (on various formats) by HTTP/HTTPS POST on several service addresses and
// stores the data locally for subsequent analysis.  This agent is always running - it's the only
// contact point for the data producers on the nodes - and there can be only one.
//
// Options:
//
// -data-dir `filename`
//   Required argument.  The root of the data store.
//
// -port `port-number`
//   Optional argument.  Port on which to listen, default is 8086.
//
// -auth-file `filename`
//   Optional but *strongly* recommended argument.  If provided then the file named must provide
//   username:password combinations, to be matched with one in an HTTP basic authentication header.
//   (If the connection is not HTTPS then the password may have been intercepted in transit.)
//
// -match-user-and-cluster
//   Optional but *strongly* recommended argument.  If set, and -auth-file is also provided, then
//   the user name provided by the HTTP connection must match the cluster name in the data packet or
//   query string.  The effect is to make it possible for each cluster to have its own
//   username:password pair and for one cluster not to be able to upload data for another.
//
// -server-cert `filename`
//   Optional unless -server-key is provided.  Path of a file holding the public certificate of the
//   HTTPS server.  Only HTTPS traffic will be accepted.
//
// -server-key `filename`
//   Optional unless -server-cert is provided.  Path of a file holding the private key of the
//   HTTPS server.  Only HTTPS traffic will be accepted.
//
// Sending SIGHUP or SIGTERM to infiltrate will shut it down in an orderly manner.
//
// About exits: Infiltrate is usually run in the background and exit codes are not easily examined,
// but when infiltrate exits it will deliver a non-zero exit code if an error was discovered during
// startup or shutdown.
//
// About panics: This server really, really needs to stay up because it's the only contact point for
// all Sonar instances on all nodes.  But we're not going to engage in heroics within the server to
// keep it running - infrastructure should restart it if it crashes (due to a panic).  Also, the
// http framework catches panics within the request handler and tries to keep the server up.
//
// About logging: Infiltrate logs everything to the syslog.  Errors encountered during startup are
// also logged to stderr.
//
// Services and input formats:
//
// /sonar-reading
//    Input is JSON-format Sonar monitoring data: array of go-utils/sonarlog.SonarReading.  Input
//    fields are checked against that record type: fields that are not known to the type are
//    dropped.
//
// /sonar-heartbeat
//    Input is JSON-format Sonar heartbeat data: array of go-utils/sonarlog.SonarHeartbeat.  Input
//    fields are checked against that record type: fields that are not known to the type are
//    dropped.
//
// /sonar-freecsv?cluster=clusterName
//    Input is "free CSV" format Sonar monitoring and heartbeat data, intermixed, one record per
//    line.  If -match-user-and-cluster is used then the the user name provided by the HTTP
//    connection must match the "cluster" parameter, which is always required.  Most input fields
//    are not checked - the records are always stored verbatim in the database, after checking the
//    cluster.  However the record must have sensible `time` and `host` fields, or it will be
//    rejected.
//
// /sysinfo?cluster=clusterName
//    Input is JSON-format system information data: a single record of
//    go-utils/config.NodeConfigRecord.  The cluster parameter is required; cluster name checking
//    is as for /sonar-freecsv.  The record must have sensible `timestamp` and `hostname` fields, or
//    it will be rejected.
//
// Notes about services:
//
// /sonar-reading and /sonar-heartbeat serve the same purpose as /sonar-freecsv but are more tightly
// coupled and probably ahead of their time.  They will make more sense when we are no longer adding
// functionality to Sonar.

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"syscall"
	"time"

	"go-utils/auth"
	"go-utils/config"
	"go-utils/datastore"
	"go-utils/httpsrv"
	"go-utils/options"
	"go-utils/process"
	"go-utils/sonarlog"
	"go-utils/status"
)

const (
	defaultListenPort = 8086
)

func main() {
	status.Start("jobanalyzer/infiltrate")
	mainLogic()
	if programFailed {
		os.Exit(1)
	}
}

// Command-line parameters
var (
	matchUserAndCluster bool
	port                int
	httpsKey            string
	httpsCert           string
	dataDir             string
	authFile            string
	verbose             bool
)

var programFailed bool
var ds *datastore.Store

func mainLogic() {
	err := commandLine()
	if err != nil {
		status.Fatalf("Command line: %v", err)
	}

	var authenticator *auth.Authenticator
	if authFile != "" {
		authenticator, err = auth.ReadPasswords(authFile)
		if err != nil {
			status.Fatalf("Failed to read authentication file: %v\n", err)
		}
	}

	// Before stopping the writer we must wait for the web server to stop, so that no more records
	// will arrive in the waiter.
	ds = datastore.Open(dataDir, verbose)
	defer ds.Close()

	http.HandleFunc("/sonar-reading", incomingData(authenticator, "application/json", sonarReading))
	http.HandleFunc("/sonar-heartbeat", incomingData(authenticator, "application/json", sonarHeartbeat))
	http.HandleFunc("/sonar-freecsv", incomingData(authenticator, "text/csv", sonarFreeCsv))
	http.HandleFunc("/sysinfo", incomingData(authenticator, "application/json", sysinfo))
	if verbose {
		status.Infof("Listening on port %d", port)
	}

	s := httpsrv.New(verbose, port, func(err error) {
		programFailed = true
	})
	go s.Start()
	defer s.Stop()

	// Wait here until we're stopped by SIGHUP (manual) or SIGTERM (from OS during shutdown).
	// TODO: For SIGHUP, we should not exit but should instead reread any config files.
	process.WaitForSignal(syscall.SIGHUP, syscall.SIGTERM)
}

func incomingData(
	authenticator *auth.Authenticator,
	contentType string,
	dataHandler func(url.Values, []byte, string) (int, string, string),
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if verbose {
			// Header reveals auth info, don't put it into logs
			status.Infof("Request from %s: %v", r.RemoteAddr, r.URL.String())
		}

		// Error logging during the preparatory steps -- until we know we have a full request that
		// is also authenticated -- is under -v in order to avoid logging storms: if some attacker
		// spews garbage at us we may otherwise DDoS ourselves with log data.
		//
		// Documented behavior: the server will close the request body, we don't need to do it.
		//
		// I can find no documentation about needing to consume the body in case of an early (error)
		// return, nor anything obvious in the net/http source code to indicate this, nor has google
		// turned up anything.  So this code assumes it's not necessary.

		if r.Method != "POST" {
			w.WriteHeader(403)
			fmt.Fprintf(w, "Bad method")
			if verbose {
				status.Warningf("Bad method: %s", r.Method)
			}
			return
		}

		ct, ok := r.Header["Content-Type"]
		if !ok || ct[0] != contentType {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Bad content-type")
			contentType := "(no type)"
			if ok {
				contentType = ct[0]
			}
			if verbose {
				status.Warningf("Bad content-type %s", contentType)
			}
			return
		}

		user, pass, ok := r.BasicAuth()
		passed := !ok && authenticator == nil ||
			ok && authenticator != nil && authenticator.Authenticate(user, pass)
		if !passed {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Unauthorized")
			if verbose {
				status.Warning("Authorization failed")
			}
			return
		}

		payload := make([]byte, r.ContentLength)
		haveRead := 0
		for haveRead < int(r.ContentLength) {
			n, err := r.Body.Read(payload[haveRead:])
			if err != nil && err != io.EOF {
				w.WriteHeader(400)
				fmt.Fprintf(w, "Bad content")
				if verbose {
					status.Warning("Bad content - can't read the body")
				}
				return
			}
			haveRead += n
		}

		code, msg, logmsg := dataHandler(r.URL.Query(), payload, user)

		// If we don't do anything then the result will just be 200 OK.
		if code != 200 {
			w.WriteHeader(code)
			fmt.Fprintf(w, msg)
			if verbose {
				status.Info(logmsg)
			}
		}
	}
}

func sonarReading(_ url.Values, payload []byte, clusterName string) (int, string, string) {
	var rs []*sonarlog.SonarReading
	err := json.Unmarshal(payload, &rs)
	if err != nil {
		return 400, "Bad content",
			fmt.Sprintf("Bad content - can't unmarshal SonarReading JSON: %v", err)
	}
	clusterUstr := sonarlog.StringToUstr(clusterName)
	for _, r := range rs {
		if !matchUserAndCluster || clusterUstr == sonarlog.UstrEmpty || r.Cluster == clusterUstr {
			ds.Write(
				r.Cluster.String(), r.Host.String(), time.Unix(r.Timestamp, 0).Format(time.RFC3339),
				"%s.csv",
				r.Csvnamed(),
			)
		}
	}
	return 200, "", ""
}

func sonarHeartbeat(_ url.Values, payload []byte, clusterName string) (int, string, string) {
	var rs []*sonarlog.SonarHeartbeat
	err := json.Unmarshal(payload, &rs)
	if err != nil {
		return 400, "Bad content",
			fmt.Sprintf("Bad content - can't unmarshal SonarHeartbeat JSON: %v", err)
	}
	clusterUstr := sonarlog.StringToUstr(clusterName)
	for _, r := range rs {
		if !matchUserAndCluster || clusterUstr == sonarlog.UstrEmpty || r.Cluster == clusterUstr {
			ds.Write(
				r.Cluster.String(), r.Host.String(), time.Unix(r.Timestamp, 0).Format(time.RFC3339),
				"%s.csv",
				r.Csvnamed(),
			)
		}
	}
	return 200, "", ""
}

func sonarFreeCsv(query url.Values, payload []byte, clusterName string) (int, string, string) {
	vs, found := query["cluster"]
	if !found || len(vs) != 1 {
		return 400, "Bad parameters", "Bad parameters - missing or repeated 'cluster'"
	}
	cluster := vs[0]
	scanner := bufio.NewScanner(bytes.NewReader(payload))
	for scanner.Scan() {
		text := scanner.Text()
		fields, err := sonarlog.GetCsvFields(text)
		if err != nil {
			return 400, "Bad content",
				fmt.Sprintf("Bad content - can't unmarshal Sonar free CSV: %v", err)
		}
		host := fields["host"]
		time := fields["time"]
		if host == "" || time == "" {
			return 400, "Bad content",
				fmt.Sprintf("Bad content - missing fields in Sonar free CSV")
		}
		if !matchUserAndCluster || clusterName == "" || cluster == clusterName {
			ds.Write(cluster, host, time, "%s.csv", []byte(text))
		}
	}
	return 200, "", ""
}

func sysinfo(query url.Values, payload []byte, clusterName string) (int, string, string) {
	vs, found := query["cluster"]
	if !found || len(vs) != 1 {
		return 400, "Bad parameters", "Bad parameters - missing or repeated 'cluster'"
	}
	cluster := vs[0]
	var info config.NodeConfigRecord
	err := json.Unmarshal(payload, &info)
	if err != nil {
		return 400, "Bad content",
			fmt.Sprintf("Bad content - can't unmarshal Sysinfo JSON: %v", err)
	}
	if info.Timestamp == "" || info.Hostname == "" {
		// Older versions of `sysinfo`
		return 400, "Bad content",
			fmt.Sprintf("Bad content - no timestamp")
	}
	if !matchUserAndCluster || clusterName == "" || cluster == clusterName {
		ds.Write(cluster, info.Hostname, info.Timestamp, "sysinfo-%s.json", payload)
	}
	return 200, "", ""
}

func commandLine() error {
	flags := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flags.IntVar(&port, "port", defaultListenPort, "Listen for connections on `port`")
	flags.StringVar(&httpsCert, "server-cert", "",
		"Listen for HTTPS connections with server cert `filename` (requires -server-key)")
	flags.StringVar(&httpsKey, "server-key", "",
		"Listen for HTTPS connections with server key `filename` (requires -server-cert)")
	flags.StringVar(&dataDir, "data-dir", "", "Root `directory` of data store (required)")
	flags.StringVar(&authFile, "auth-file", "", "Read user names and passwords from `filename`")
	flags.BoolVar(&matchUserAndCluster, "match-user-and-cluster", false,
		"Require user name to match cluster name")
	var dataPath string
	flags.StringVar(&dataPath, "data-path", "", "Obsolete name for -data-dir")
	flags.BoolVar(&verbose, "v", false, "Verbose logging")
	err := flags.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		os.Exit(0)
	}
	if err != nil {
		return err
	}
	if dataDir == "" {
		dataDir = dataPath
	}
	dataDir, err = options.RequireDirectory(dataDir, "-data-dir")
	if err != nil {
		return err
	}
	if (httpsCert != "") != (httpsKey != "") {
		return fmt.Errorf("Need both -https-cert and -https-key, or neither")
	}
	return nil
}
