// `infiltrate` - data receiver for Sonar data, to run on database/analysis host.
//
// Infiltrate receives JSON-formatted data by HTTP POST on several addresses and stores the data
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
// About logging: Infiltrate logs everything to the syslog with the tag defined below ("logTag").
// Errors encountered during startup are also logged to stderr.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log/syslog"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"go-utils/auth"
	"go-utils/sonarlog"
)

const (
	defaultListenPort        = 8086
	dirPermissions           = 0755
	filePermissions          = 0644
	serverShutdownTimeoutSec = 10
	logTag                   = "jobanalyzer/infiltrate"
)

var verbose bool
var programFailed = false

func main() {
	startLogger()
	port, dataPath, authFile := commandLine()
	var err error
	var authUser, authPass string
	if authFile != "" {
		authUser, authPass, err = auth.ParseAuth(authFile)
		if err != nil {
			fatal(fmt.Sprintf("Failed to read authentication file: %v\n", err))
		}
	}
	go runWriter()
	go runServer(port, dataPath, authUser, authPass)
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

func runServer(port int, dataPath, authUser, authPass string) {
	http.HandleFunc(
		"/sonar-reading",
		incomingData(
			authUser,
			authPass,
			func(payload []byte) (int, string, string) {
				return sonarReading(payload, dataPath)
			}),
	)
	http.HandleFunc(
		"/sonar-heartbeat",
		incomingData(
			authUser,
			authPass,
			func(payload []byte) (int, string, string) {
				return sonarHeartbeat(payload, dataPath)
			}),
	)
	if verbose {
		logInfo(fmt.Sprintf("Listening on port %d", port))
	}
	server = &http.Server{Addr: fmt.Sprintf(":%d", port)}
	err := server.ListenAndServe()
	if err != nil {
		if err != http.ErrServerClosed {
			logError(err.Error())
			logError("SERVER NOT RUNNING")
			// TODO: This is ugly, though legal
			programFailed = true
		} else {
			logInfo(err.Error())
		}
	}
	serverStopChannel <- true
}

func stopServer() {
	ctx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeoutSec*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logWarning(err.Error())
	}
	<-serverStopChannel
}

func commandLine() (port int, dataPath, authFile string) {
	flag.IntVar(&port, "port", defaultListenPort, "Port to listen on")
	flag.StringVar(&dataPath, "data-path", "", "Path of data store root directory")
	flag.StringVar(&authFile, "auth-file", "", "Authentication file")
	flag.BoolVar(&verbose, "v", false, "Verbose logging")
	flag.Parse()

	if dataPath == "" {
		fatal("Required argument: -data-path")
	}
	dataPath = path.Clean(dataPath)
	info, err := os.DirFS(dataPath).(fs.StatFS).Stat(".")
	if err != nil || !info.IsDir() {
		fatal(fmt.Sprintf("Bad -data-path directory %s", dataPath))
	}

	return
}

func incomingData(
	authUser, authPass string,
	dataHandler func([]byte) (int, string, string),
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if verbose {
			logInfo(fmt.Sprintf("Request from %s: %v", r.RemoteAddr, r.Header))
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
				logWarning(fmt.Sprintf("Bad method: %s", r.Method))
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
				logWarning(fmt.Sprintf("Bad content-type %s", contentType))
			}
			return
		}

		user, pass, ok := r.BasicAuth()
		passed := !ok && authPass == "" || ok && user == authUser && pass == authPass
		if !passed {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Unauthorized")
			if verbose {
				logWarning("Authorization failed")
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
					logWarning("Bad content - can't read the body")
				}
				return
			}
			haveRead += n
		}

		code, msg, logmsg := dataHandler(payload)

		// If we don't do anything then the result will just be 200 OK.
		if code != 200 {
			w.WriteHeader(code)
			fmt.Fprintf(w, msg)
			if verbose {
				logInfo(logmsg)
			}
		}
	}
}

func sonarReading(payload []byte, dataPath string) (int, string, string) {
	var rs []*sonarlog.SonarReading
	err := json.Unmarshal(payload, &rs)
	if err != nil {
		return 400, "Bad content", fmt.Sprintf("Bad content - can't unmarshal SonarReading JSON: %v", err)
	}
	for _, r := range rs {
		writeRecord(dataPath, r.Cluster, r.Host, r.Timestamp, r.Csvnamed())
	}
	return 200, "", ""
}

func sonarHeartbeat(payload []byte, dataPath string) (int, string, string) {
	var rs []*sonarlog.SonarHeartbeat
	err := json.Unmarshal(payload, &rs)
	if err != nil {
		return 400, "Bad content", fmt.Sprintf("Bad content - can't unmarshal SonarHeartbeat JSON: %v", err)
	}
	for _, r := range rs {
		writeRecord(dataPath, r.Cluster, r.Host, r.Timestamp, r.Csvnamed())
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
			logWarning(fmt.Sprintf("Bad timestamp %s, dropping record", timestamp))
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
				logError(fmt.Sprintf("Write error on log (%v), %d bytes written of %d", err, n, len(r.payload)))
			}
		}
	}
	dataStopChannel <- true
}

func maybeRetry(r *dataRecord, msg string) {
	if r.attempts < maxAttempts {
		if verbose {
			logInfo(msg + ", retrying later")
		}
		go func() {
			// Obviously some kind of backoff is possible, but do we care?
			time.Sleep(time.Duration(5 * time.Minute))
			dataChannel <- r
		}()
	} else {
		logWarning(msg + ", too many retries, abandoning")
	}
}

func fatal(msg string) {
	logCritical(msg)
	fmt.Fprintf(os.Stderr, "FATAL: %s\n", msg)
	os.Exit(1)
}

var logger *syslog.Writer

func startLogger() {
	var err error
	// The "","" address connects us to the Unix syslog daemon.  The priority (INFO) is a
	// placeholder, it will be overridden by all the logger functions below.
	logger, err = syslog.Dial("", "", syslog.LOG_INFO|syslog.LOG_USER, logTag)
    if err != nil {
		fatal(err.Error())
    }
}

func logCritical(msg string) {
	if logger != nil {
		logger.Crit(msg)
	}
}

func logError(msg string) {
	if logger != nil {
		logger.Err(msg)
	}
}

func logWarning(msg string) {
	if logger != nil {
		logger.Warning(msg)
	}
}

func logInfo(msg string) {
	if logger != nil {
		logger.Info(msg)
	}
}