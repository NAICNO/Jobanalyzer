package daemon

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/syslog"
	"net/http"
	"os"
	"path"
	"strings"
	"syscall"

	"go-utils/auth"
	"go-utils/httpsrv"
	"go-utils/process"
)

func (dc *DaemonCommand) RunDaemon(_ io.Reader, _, stderr io.Writer) /* never returns */ {
	if err := statusStart("jobanalyzer/sonalyzed", stderr); err != nil {
		fmt.Fprintf(stderr, "FATAL ERROR: Failing to open logger: %s", err.Error())
		os.Exit(1)
	}

	// Note "daemon" is not a command here
	http.HandleFunc("/add", httpAddHandler(dc))
	http.HandleFunc("/jobs", httpGetHandler(dc, "jobs"))
	http.HandleFunc("/load", httpGetHandler(dc, "load"))
	http.HandleFunc("/uptime", httpGetHandler(dc, "uptime"))
	http.HandleFunc("/profile", httpGetHandler(dc, "profile"))
	http.HandleFunc("/parse", httpGetHandler(dc, "parse"))
	http.HandleFunc("/metadata", httpGetHandler(dc, "metadata"))
	// These request names are compatible with the older `infiltrate` and with the upload infra
	// already running on the clusters.
	http.HandleFunc("/sonar-freecsv", httpPostHandler(dc, "sample", "text/csv"))
	http.HandleFunc("/sysinfo", httpPostHandler(dc, "sysinfo", "application/json"))

	var programFailed bool
	s := httpsrv.New(dc.Verbose, int(dc.port), func(err error) {
		programFailed = true
	})
	go s.Start()

	// Wait here until we're stopped by SIGHUP (manual) or SIGTERM (from OS during shutdown).
	//
	// TODO: IMPROVEME: For SIGHUP, we should not exit but should instead reread the password file
	// and the cluster aliases file.
	process.WaitForSignal(syscall.SIGHUP, syscall.SIGTERM)
	s.Stop()

	if programFailed {
		os.Exit(1)
	}
	os.Exit(0)
}

// HTTP handlers
//
// Documented behavior: the server will close the request body, we don't need to do it.
//
// I can find no documentation about needing to consume the body in case of an early (error)
// return, nor anything obvious in the net/http source code to indicate this, nor has google
// turned up anything.  So request handler code assumes it's not necessary.

func httpGetHandler(
	dc *DaemonCommand,
	command string,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, _, clusterName, ok := requestPreamble(dc, w, r, "GET", dc.getAuthenticator, authRealm, "")
		if !ok {
			return
		}

		verb := command
		arguments := []string{
			"--data-path",
			path.Join(dc.jobanalyzerDir, "data", clusterName),
		}
		for name, vs := range r.URL.Query() {
			if name == "cluster" {
				continue
			}

			if !argOk(command, name) {
				w.WriteHeader(400)
				fmt.Fprintf(w, "Bad parameter %s", name)
				if dc.Verbose {
					statusWarningf("Bad parameter %s", name)
				}
				return
			}

			// Repeats are OK, sonalyze allows them in a number of cases.  Booleans carry the magic
			// boolean value, allowing us to construct a boolean option without a value without
			// tracking name->type mappings.  This is a hack, but it works.  The reason we have to
			// exclude the value for boolean options (`--some-gpu` instead of `--some-gpu=true`) is
			// a limitation of Rust's `clap` library.  The reason the value carried is not simply
			// "true" is that that is a more likely value for some other parameters (host names?)
			// and we can't exclude it here without risk.

			for _, v := range vs {
				arguments = append(arguments, "--"+name)
				if v != magicBoolean {
					arguments = append(arguments, v)
				}
			}
		}

		switch command {
		case "jobs", "load", "uptime":
			arguments = append(
				arguments,
				"--config-file",
				path.Join(dc.jobanalyzerDir, "scripts", clusterName, clusterName+"-config.json"),
			)
		}

		stdout, ok := runSonalyze(dc, w, verb, arguments, []byte{})
		if !ok {
			return
		}

		w.WriteHeader(200)
		fmt.Fprint(w, stdout)
	}
}

func httpAddHandler(dc *DaemonCommand) func(http.ResponseWriter, *http.Request) {
	forSample := httpPostHandler(dc, "sample", "text/csv")
	forSysinfo := httpPostHandler(dc, "sysinfo", "application/json")
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		vs, isSample := query["sample"]
		var e1, e2, e3 error
		if isSample && (len(vs) != 1 || vs[0] != magicBoolean) {
			e1 = errors.New("Bad `sample` parameter")
		}
		ws, isSysinfo := query["sysinfo"]
		if isSysinfo && (len(ws) != 1 || ws[0] != magicBoolean) {
			e2 = errors.New("Bad `sysinfo` parameter")
		}
		if isSample == isSysinfo {
			e3 = errors.New("Need `-sample` or `-sysinfo` but not both")
		}
		if err := errors.Join(e1, e2, e3); err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Bad operation: %s", err.Error())
			if dc.Verbose {
				statusWarningf("Bad operation: %s", err.Error())
			}
			return
		}
		switch {
		case isSample:
			forSample(w, r)
		case isSysinfo:
			forSysinfo(w, r)
		default:
			panic("Unexpected")
		}
	}
}

func httpPostHandler(
	dc *DaemonCommand,
	dataType string,
	contentType string,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, userName, clusterName, ok := requestPreamble(dc, w, r, "POST", dc.postAuthenticator, "", contentType)
		if !ok {
			return
		}

		if dc.matchUserAndCluster && userName != "" && clusterName != userName {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Upload not authorized")
			if dc.Verbose {
				statusWarningf("Upload not authorized")
			}
			return
		}

		verb := "add"
		arguments := []string{
			"--" + dataType,
			"--data-path",
			path.Join(dc.jobanalyzerDir, "data", clusterName),
		}

		stdout, ok := runSonalyze(dc, w, verb, arguments, payload)
		if !ok {
			return
		}

		w.WriteHeader(200)
		fmt.Fprint(w, stdout)
	}
}

func requestPreamble(
	dc *DaemonCommand,
	w http.ResponseWriter,
	r *http.Request,
	method string,
	authenticator *auth.Authenticator,
	realm string,
	contentType string,
) (payload []byte, userName, clusterName string, ok bool) {
	if dc.Verbose {
		// Header reveals auth info, don't put it into logs
		statusInfof("Request from %s: %v", r.RemoteAddr, r.URL.String())
	}

	if !httpsrv.AssertMethod(w, r, method, dc.Verbose) {
		return
	}

	authOk, userName := httpsrv.Authenticate(w, r, authenticator, realm, dc.Verbose)
	if !authOk {
		return
	}

	payload, havePayload := httpsrv.ReadPayload(w, r, dc.Verbose)
	if !havePayload {
		return
	}

	if contentType != "" {
		if !httpsrv.AssertContentType(w, r, contentType, dc.Verbose) {
			return
		}
	}

	clusterValues, found := r.URL.Query()["cluster"]
	if !found || len(clusterValues) != 1 || clusterValues[0] == "" {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Bad parameters - missing or empty or repeated 'cluster'")
		if dc.Verbose {
			statusWarningf("Bad parameters - missing or empty or repeated 'cluster'")
		}
		return
	}

	clusterName = clusterValues[0]
	if dc.aliasResolver != nil {
		clusterName = dc.aliasResolver.Resolve(clusterName)
	}

	ok = true
	return
}

func runSonalyze(dc *DaemonCommand, w http.ResponseWriter, verb string, arguments []string, input []byte) (stdout string, ok bool) {
	cmdName := "<sonalyze>"

	// Run the command and report the result

	if dc.Verbose {
		statusInfof(
			"Command: %s %s",
			path.Join(dc.jobanalyzerDir, cmdName),
			verb+" "+strings.Join(arguments, " "),
		)
	}

	anyCmd, _ := dc.cmdlineHandler.ParseVerb(cmdName, verb)
	if anyCmd == nil {
		errResponse(w, 400, fmt.Errorf("Bad verb in remote command: %s", verb), "", dc.Verbose)
		return
	}
	fs := flag.NewFlagSet(cmdName, flag.ContinueOnError)
	err := dc.cmdlineHandler.ParseArgs(verb, arguments, anyCmd, fs)
	if err != nil {
		errResponse(w, 400, err, "", dc.Verbose)
		return
	}
	stop, err := dc.cmdlineHandler.StartCPUProfile(anyCmd.CpuProfileFile())
	if err != nil {
		statusWarningf("Failed to start CPU profile")
	}
	if stop != nil {
		defer stop()
	}

	var stdoutBuf, stderrBuf strings.Builder
	err = dc.cmdlineHandler.HandleCommand(anyCmd, bytes.NewReader(input), &stdoutBuf, &stderrBuf)
	stdout = stdoutBuf.String()
	stderr := stderrBuf.String()
	if err != nil {
		errResponse(w, 400, err, stderr, dc.Verbose)
		return
	}
	if stderr != "" {
		statusWarningf(stderr, "")
	}

	ok = true
	return
}

func errResponse(w http.ResponseWriter, code int, err error, stderr string, verbose bool) {
	w.WriteHeader(code)
	fmt.Fprint(w, err.Error())
	if stderr != "" {
		fmt.Fprint(w, "\n", stderr)
	}
	if verbose {
		statusWarningf("ERROR: %v", err)
	}
}

// Disallow argument names that are malformed or are specific values.  This is not fabulous but
// maintaining a whitelist is a lot of work.

func argOk(command, arg string) bool {
	// Args are alphabetic and lower-case only, except - is allowed except in the first position
	for i, c := range arg {
		switch {
		case c >= 'a' && c <= 'z':
			// OK
		case c == '-' && i > 0:
			// OK
		default:
			return false
		}
	}

	// Disallow short options (pretty primitive)
	if len(arg) <= 1 {
		return false
	}

	// Specific names are excluded, for now, the names in the comments relate to structure names in
	// sonalyze/src/sonalyze.rs or sonalyze/command/args.go.
	switch arg {
	case "cpuprofile":
		// DevArgs (go)
		return false
	case "data-path", "data-dir":
		// SourceArgs (rust), DataDirArgs (go)
		return false
	case "cluster", "remote", "auth-file":
		// SourceArgs
		return false
	case "config-file":
		// ConfigFileArgs
		return false
	case "verbose":
		// VerboseArgs
		return false
	case "raw":
		// MetaArgs (rust)
		return false
	default:
		return true
	}
}

var logger *syslog.Writer
var logStderr io.Writer

func statusStart(logTag string, stderr io.Writer) error {
	// Currently logStderr is unused but the vision is that it can be used for some debugging logging
	logStderr = stderr

	// The "","" address connects us to the Unix syslog daemon.  The priority (INFO) is a
	// placeholder, it will be overridden by all the logger functions below.
	var err error
	logger, err = syslog.Dial("", "", syslog.LOG_INFO|syslog.LOG_USER, logTag)
	return err
}

func statusWarningf(format string, args ...any) {
	if logger != nil {
		logger.Warning(fmt.Sprintf(format, args...))
	}
}

func statusInfof(format string, args ...any) {
	if logger != nil {
		logger.Info(fmt.Sprintf(format, args...))
	}
}
