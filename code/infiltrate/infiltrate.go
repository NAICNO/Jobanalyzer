// `infiltrate` - data receiver for Sonar data, to run on database/analysis host.
//
// Infiltrate receives JSON-formatted data by HTTP/HTTPS POST on several addresses and stores the data
// locally for subsequent analysis.  This agent is always running - it's the only contact point for
// the data producers on the nodes.
//
// For the time being, we use the Sonar CSV format and directory structure to store all Sonar data,
// this allows existing consumers to work without change, just without needing a shared disk with
// the data producers.  The storage format will change later.
//
// The -data-path option is required while -port is optional.
//
// If the -auth-file option is provided then the file named must provide a user name and password
// on the form username/password, to be matched with one in an HTTP basic authentication header.  If
// the connection is not HTTPS then the password may have been intercepted in transit.
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

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"go-utils/auth"
	"go-utils/options"
	"go-utils/sonarlog"
	"go-utils/status"
)

const (
	defaultListenPort        = 8086
	dirPermissions           = 0755
	filePermissions          = 0644
	serverShutdownTimeoutSec = 10
	matchUserAndCluster      = false // This will become true eventually
)

var verbose bool
var programFailed = false

func main() {
	status.Start("jobanalyzer/infiltrate")

	port, httpsKey, httpsCert, dataPath, authFile, err := commandLine()
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
	go runWriter()

	// TODO: We have shared abstractions for the HTTP server and the signal handling, now.  See
	// sonalyzed for examples.

	go runServer(port, httpsKey, httpsCert, dataPath, authenticator)
	// Hang until SIGHUP or SIGTERM, then shut down orderly.
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, syscall.SIGHUP)  // Sent manually and maybe when logging out?
	signal.Notify(stopSignal, syscall.SIGTERM) // Sent during shutdown
	<-stopSignal
	stopServer()
	stopWriter()
	if programFailed {
		os.Exit(1)
	}
}

var serverStopChannel = make(chan bool)
var server *http.Server

func runServer(port int, httpsKey, httpsCert, dataPath string, authenticator *auth.Authenticator) {
	http.HandleFunc(
		"/sonar-reading",
		incomingData(
			authenticator,
			func(payload []byte, clusterName string) (int, string, string) {
				return sonarReading(payload, dataPath, clusterName)
			}),
	)
	http.HandleFunc(
		"/sonar-heartbeat",
		incomingData(
			authenticator,
			func(payload []byte, clusterName string) (int, string, string) {
				return sonarHeartbeat(payload, dataPath, clusterName)
			}),
	)
	if verbose {
		status.Infof("Listening on port %d", port)
	}
	var err error
	if httpsKey != "" {
		hn, err := os.Hostname()
		if err == nil {
			server = &http.Server{Addr: fmt.Sprintf("%s:%d", hn, port)}
			err = server.ListenAndServeTLS(httpsCert, httpsKey)
		}
	} else {
		server = &http.Server{Addr: fmt.Sprintf(":%d", port)}
		err = server.ListenAndServe()
	}
	if err != nil {
		if err != http.ErrServerClosed {
			status.Error(err.Error())
			status.Error("SERVER NOT RUNNING")
			// TODO: This is ugly, though legal
			programFailed = true
		} else {
			status.Info(err.Error())
		}
	}
	serverStopChannel <- true
}

func stopServer() {
	ctx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeoutSec*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		status.Warning(err.Error())
	}
	<-serverStopChannel
}

func incomingData(
	authenticator *auth.Authenticator,
	dataHandler func([]byte, string) (int, string, string),
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
		if !ok || ct[0] != "application/json" {
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
		passed := !ok && authenticator == nil || ok && authenticator != nil && authenticator.Authenticate(user, pass)
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

		code, msg, logmsg := dataHandler(payload, user)

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

func sonarReading(payload []byte, dataPath, clusterName string) (int, string, string) {
	var rs []*sonarlog.SonarReading
	err := json.Unmarshal(payload, &rs)
	if err != nil {
		return 400, "Bad content",
			fmt.Sprintf("Bad content - can't unmarshal SonarReading JSON: %v", err)
	}
	for _, r := range rs {
		if !matchUserAndCluster || clusterName == "" || r.Cluster == clusterName {
			writeRecord(dataPath, r.Cluster, r.Host, r.Timestamp, r.Csvnamed())
		}
	}
	return 200, "", ""
}

func sonarHeartbeat(payload []byte, dataPath, clusterName string) (int, string, string) {
	var rs []*sonarlog.SonarHeartbeat
	err := json.Unmarshal(payload, &rs)
	if err != nil {
		return 400, "Bad content",
			fmt.Sprintf("Bad content - can't unmarshal SonarHeartbeat JSON: %v", err)
	}
	for _, r := range rs {
		if !matchUserAndCluster || clusterName == "" || r.Cluster == clusterName {
			writeRecord(dataPath, r.Cluster, r.Host, r.Timestamp, r.Csvnamed())
		}
	}
	return 200, "", ""
}

// There will be only one copy of `infiltrate` running on the server, so data files have only one
// writer, and we don't need a lock in the file system for concurrent writing.  There is a danger
// that there's a reader while we're writing but that danger is already there and should be fixed
// separately.
//
// However, each HTTP request is served on a separate goroutine and we don't want to have to deal
// with mutexing the log files *internally* in `infiltrate`, so we run the writer on a separate
// goroutine.

type dataRecord struct {
	dataPath string // command line arg, this existed at startup at least
	dirname  string // path underneath dataPath
	filename string // path underneath dataPath
	payload  []byte // text to write
	attempts int    // number of attempts so far
}

const maxAttempts = 6
const channelCapacity = 1000

var dataChannel = make(chan *dataRecord, channelCapacity)
var dataStopChannel = make(chan bool)

// To stop the writer, first the caller must wait for the web server to stop so that no more records
// to will arrive on the dataChannel.  Then send nil on dataChannel to make the writer exit its loop
// in an orderly way.  Wait for a response from the writer on stopChannel, and we are done.
//
// We don't care about any of the pending retry writes - their sleeping goroutines will either be
// terminated when the program exits, or will wake up and send data on a channel nobody's listening
// on before that.  Either way this is invisible.

func stopWriter() {
	dataChannel <- nil
	<-dataStopChannel
}

// writeRecord is infallible, all operations that could fail are performed in the runWriter loop.

func writeRecord(dataPath, cluster, host, timestamp string, payload []byte) {
	// The path will be (below dataPath) cluster/year/month/day/hostname.csv
	tval, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		if verbose {
			status.Warningf("Bad timestamp %s, dropping record", timestamp)
		}
	}
	dirname := fmt.Sprintf("%s/%04d/%02d/%02d", cluster, tval.Year(), tval.Month(), tval.Day())
	filename := fmt.Sprintf("%s/%s.csv", dirname, host)

	dataChannel <- &dataRecord{dataPath, dirname, filename, payload, 0}
}

// TODO: Optimization: Cache open files for a time.

func runWriter() {
	for {
		r := <-dataChannel
		if r == nil {
			break
		}
		r.attempts++
		if verbose {
			fmt.Printf("Storing: %s", string(r.payload))
		}

		err := os.MkdirAll(path.Join(r.dataPath, r.dirname), dirPermissions)
		if err != nil {
			// Could be disk full, fs went away, element of path exists as file, wrong permissions
			maybeRetry(r, fmt.Sprintf("Failed to create path (%v)", err))
			return
		}
		f, err := os.OpenFile(
			path.Join(r.dataPath, r.filename),
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			filePermissions,
		)
		if err != nil {
			// Could be disk full, fs went away, file is directory, wrong permissions
			maybeRetry(r, fmt.Sprintf("Failed to open/create file (%v)", err))
			return
		}
		n, err := f.Write(r.payload)
		f.Close()
		if err != nil {
			if n == 0 {
				// Nothing was written so try to recover by restarting.
				maybeRetry(r, fmt.Sprintf("Failed to write file (%v)", err))
			} else {
				// Partial data were written.
				//
				// The usual and benign reason for a partial write on Unix is that a signal was
				// delivered in the middle of a write and the write needs to restart with the rest
				// of the data; this is signalled with an EINTR error return from write(2).  The Go
				// libraries try to hide that problem - see internal/poll/fd_unix.go in the Go
				// sources, the function Write() is the normal destination for file output.  It
				// calls ignoringEINTRIO(syscall.Write) to perform the write, and that in turn will
				// restart the write in the case of EINTR.
				//
				// Of course, "transparently restarting after writing some data" in O_APPEND mode is
				// a complete fiction, but what can you do.
				//
				// Anyway, if we get here with a partial data write it's going to be something more
				// serious than EINTR, such as a disk full.  Trying to recover is probably not worth
				// our time.  Just log the failure and hope somebody sees it.
				status.Errorf("Write error on log (%v), %d bytes written of %d",
					err, n, len(r.payload))
			}
		}
	}
	dataStopChannel <- true
}

func maybeRetry(r *dataRecord, msg string) {
	if r.attempts < maxAttempts {
		if verbose {
			status.Info(msg + ", retrying later")
		}
		go func() {
			// Obviously some kind of backoff is possible, but do we care?
			time.Sleep(time.Duration(5 * time.Minute))
			dataChannel <- r
		}()
	} else {
		status.Warning(msg + ", too many retries, abandoning")
	}
}

func commandLine() (port int, httpsKey, httpsCert, dataPath, authFile string, err error) {
	flags := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flags.IntVar(&port, "port", defaultListenPort, "Listen for connections on `port`")
	flags.StringVar(&httpsCert, "server-cert", "",
		"Listen for HTTPS connections with server cert `filename` (requires -server-key)")
	flags.StringVar(&httpsKey, "server-key", "",
		"Listen for HTTPS connections with server key `filename` (requires -server-cert)")
	flags.StringVar(&dataPath, "data-path", "", "Path of data store root `directory` (required)")
	flags.StringVar(&authFile, "auth-file", "", "Read user names and passwords from `filename`")
	flags.BoolVar(&verbose, "v", false, "Verbose logging")
	err = flags.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		os.Exit(0)
	}
	if err != nil {
		return
	}
	dataPath, err = options.RequireDirectory(dataPath, "-data-path")
	if err != nil {
		return
	}
	if (httpsCert != "") != (httpsKey != "") {
		err = fmt.Errorf("Need both -https-cert and -https-key, or neither")
	}
	return
}
