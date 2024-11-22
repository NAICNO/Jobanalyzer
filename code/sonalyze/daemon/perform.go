// See ../TECHNICAL.md for a definition of the protocol.
//
// When adding a new command to the daemon, several points in this file have to be updated:
//
// - a new handler has to be installed in RunDaemon()
// - any special argument construction has to be created in httpGetHandler() (several places) or
//   httpPostHandler()
// - any local-only arguments that should never be forwarded need to be added to the blacklist
//   in argOk()
//
// In addition, due to the structure of the URL syntax, a new command point may need to be added to
// the HTTP server's configuration file.

package daemon

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/syslog"
	"net/http"
	"net/url"
	"path"
	"strings"
	"syscall"

	"go-utils/auth"
	"go-utils/httpsrv"
	"go-utils/process"
	. "sonalyze/cmd"
	. "sonalyze/common"
	"sonalyze/db"
)

// Note, this should *NOT* be called Perform(), so that we can be sure we're not confusing a
// DaemonCommand with a SimpleCommand.

func (dc *DaemonCommand) RunDaemon(_ io.Reader, _, stderr io.Writer) error {
	logger, err := syslog.Dial("", "", syslog.LOG_INFO|syslog.LOG_USER, logTag)
	if err != nil {
		return fmt.Errorf("FATAL ERROR: Failing to open logger: %v", err)
	}
	Log.SetUnderlying(logger)

	if dc.cacheSize > 0 {
		db.CacheInit(dc.cacheSize)
	}

	// Note "daemon" is not a command here
	http.HandleFunc("/add", httpAddHandler(dc))
	http.HandleFunc("/cluster", httpGetHandler(dc, "cluster"))
	http.HandleFunc("/config", httpGetHandler(dc, "config"))
	http.HandleFunc("/node", httpGetHandler(dc, "node"))
	http.HandleFunc("/jobs", httpGetHandler(dc, "jobs"))
	http.HandleFunc("/load", httpGetHandler(dc, "load"))
	http.HandleFunc("/uptime", httpGetHandler(dc, "uptime"))
	http.HandleFunc("/profile", httpGetHandler(dc, "profile"))
	http.HandleFunc("/parse", httpGetHandler(dc, "sample"))
	http.HandleFunc("/sample", httpGetHandler(dc, "sample"))
	http.HandleFunc("/report", httpGetHandler(dc, "report"))
	http.HandleFunc("/metadata", httpGetHandler(dc, "metadata"))
	http.HandleFunc("/sacct", httpGetHandler(dc, "sacct"))
	// These request names are compatible with the older `infiltrate` and `sonalyzed`, and with the
	// upload infra already running on the clusters.  We'd like to get rid of them eventually.
	http.HandleFunc("/sonar-freecsv", httpPostHandler(dc, "sample", "text/csv"))
	http.HandleFunc("/sysinfo", httpPostHandler(dc, "sysinfo", "application/json"))

	var programFailed bool
	s := httpsrv.New(dc.Verbose, int(dc.port), func(err error) {
		programFailed = true
	})
	go s.Start()

	// Wait here until we're stopped by SIGHUP (manual) or SIGTERM (from OS during shutdown).
	//
	// TODO: IMPROVEME: For SIGHUP, we should not exit but should instead reread the password file,
	// the cluster aliases file, and the configuration files (we could purge the config object
	// cache).  Really we must be purging the entire LogFile cache in this case too.
	process.WaitForSignal(syscall.SIGHUP, syscall.SIGTERM)
	s.Stop()

	if programFailed {
		return errors.New("HTTP server failed to start, or errored out")
	}
	return nil
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
		_, _, clusterName, ok :=
			requestPreamble(dc, command, w, r, "GET", dc.getAuthenticator, authRealm, "")
		if !ok {
			return
		}

		verb := command
		arguments := []string{}

		// For `cluster` we want only the jobanalyzerDir
		switch command {
		case "cluster":
			arguments = append(arguments, "--jobanalyzer-dir", dc.jobanalyzerDir)
		case "report":
			arguments = append(
				arguments,
				"--report-dir",
				db.MakeReportDirPath(dc.jobanalyzerDir, clusterName),
			)
		case "config":
			// Nothing, we add the config-file below
		default:
			arguments = append(
				arguments,
				"--data-path",
				db.MakeClusterDataPath(dc.jobanalyzerDir, clusterName),
			)
		}

		for name, vs := range r.URL.Query() {
			if name == "cluster" {
				continue
			}

			if !argOk(command, name) {
				w.WriteHeader(400)
				fmt.Fprintf(w, "Bad parameter %s", name)
				if dc.Verbose {
					Log.Warningf("Bad parameter %s", name)
				}
				return
			}

			// Repeats are OK, the commands allow them in a number of cases.
			//
			// Booleans carry the regular true/false values or, for backward compatibility, the old
			// MagicBoolean value.  See comments in ../command/reify.go.

			for _, v := range vs {
				// The MagicBoolean is handled by transforming it to "true", for uniformity.
				if v == MagicBoolean {
					v = "true"
				}
				// Go requires "=" between parameter and name for boolean params, but allows it for
				// every type, so do it uniformly.
				arguments = append(arguments, "--"+name+"="+v)
			}
		}

		// Everyone except `cluster` and `report` gets a config, which they will need for caching
		// things properly.
		if command != "cluster" && command != "report" {
			arguments = append(
				arguments,
				"--config-file",
				db.MakeConfigFilePath(dc.jobanalyzerDir, clusterName),
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

func parseAddQuery(query url.Values, name string) (isSet bool, err error) {
	vs, isName := query[name]
	if !isName {
		return
	}
	if len(vs) == 1 {
		switch vs[0] {
		case "true", MagicBoolean:
			isSet = true
			return
		case "false":
			return
		}
	}
	err = fmt.Errorf("Bad `%s` parameter", name)
	return
}

func httpAddHandler(dc *DaemonCommand) func(http.ResponseWriter, *http.Request) {
	forSample := httpPostHandler(dc, "sample", "text/csv")
	forSlurmSacct := httpPostHandler(dc, "slurm-sacct", "text/csv")
	forSysinfo := httpPostHandler(dc, "sysinfo", "application/json")
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		isSample, e1 := parseAddQuery(query, "sample")
		isSysinfo, e2 := parseAddQuery(query, "sysinfo")
		isSlurmSacct, e3 := parseAddQuery(query, "slurm-sacct")
		n := 0
		if isSample {
			n++
		}
		if isSysinfo {
			n++
		}
		if isSlurmSacct {
			n++
		}
		var e4 error
		if n != 1 {
			e4 = errors.New("Need exactly one of `-sample`, `-sysinfo`, or `-slurm-sacct`")
		}
		if err := errors.Join(e1, e2, e3, e4); err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Bad operation: %s", err.Error())
			if dc.Verbose {
				Log.Warningf("Bad operation: %s", err.Error())
			}
			return
		}
		switch {
		case isSample:
			forSample(w, r)
		case isSysinfo:
			forSysinfo(w, r)
		case isSlurmSacct:
			forSlurmSacct(w, r)
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
		payload, userName, clusterName, ok :=
			requestPreamble(dc, "add", w, r, "POST", dc.postAuthenticator, "", contentType)
		if !ok {
			return
		}

		if dc.matchUserAndCluster && userName != "" && clusterName != userName {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Upload not authorized")
			if dc.Verbose {
				Log.Warningf("Upload not authorized")
			}
			return
		}

		verb := "add"
		arguments := []string{
			"--" + dataType,
			"--data-path",
			db.MakeClusterDataPath(dc.jobanalyzerDir, clusterName),
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
	command string,
	w http.ResponseWriter,
	r *http.Request,
	method string,
	authenticator *auth.Authenticator,
	realm string,
	contentType string,
) (payload []byte, userName, clusterName string, ok bool) {
	if dc.Verbose {
		// Header reveals auth info, don't put it into logs
		Log.Infof("Request from %s: %v", r.RemoteAddr, r.URL.String())
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
	if command != "cluster" {
		if !found || len(clusterValues) != 1 || clusterValues[0] == "" {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Bad parameters - missing or empty or repeated 'cluster'")
			if dc.Verbose {
				Log.Warningf("Bad parameters - missing or empty or repeated 'cluster'")
			}
			return
		}

		clusterName = clusterValues[0]
		if dc.aliasResolver != nil {
			clusterName = dc.aliasResolver.Resolve(clusterName)
		}
	} else {
		if found {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Bad parameters - illegal 'cluster'")
			if dc.Verbose {
				Log.Warningf("Bad parameters - illegal 'cluster'")
			}
			return
		}
	}

	ok = true
	return
}

func runSonalyze(
	dc *DaemonCommand,
	w http.ResponseWriter,
	verb string,
	arguments []string,
	input []byte,
) (stdout string, ok bool) {
	cmdName := "<sonalyze>"

	// Run the command and report the result

	if dc.Verbose {
		Log.Infof(
			"Command: %s %s",
			path.Join(dc.jobanalyzerDir, cmdName),
			verb+" "+strings.Join(arguments, " "),
		)
	}

	anyCmd, _ := dc.cmdlineHandler.ParseVerb(cmdName, verb)
	if anyCmd == nil {
		errResponse(w, 400, fmt.Errorf("Bad verb in daemon-dispatched command: %s", verb), "", dc.Verbose)
		return
	}
	fs := NewCLI(verb, anyCmd, cmdName, false)
	err := dc.cmdlineHandler.ParseArgs(verb, arguments, anyCmd, fs)
	if err != nil {
		errResponse(w, 400, err, "", dc.Verbose)
		return
	}

	// The -cpuprofile option is ignored here, it should have forced ParseArgs to error out.

	var stdoutBuf, stderrBuf strings.Builder
	err = dc.cmdlineHandler.HandleCommand(anyCmd, bytes.NewReader(input), &stdoutBuf, &stderrBuf)
	stdout = stdoutBuf.String()
	stderr := stderrBuf.String()
	if err != nil {
		errResponse(w, 400, err, stderr, dc.Verbose)
		return
	}
	if stderr != "" {
		Log.Warningf(stderr, "")
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
		Log.Warningf("ERROR: %v", err)
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
	case "report-dir":
		// ReportCommand
		return false
	case "verbose", "v":
		// VerboseArgs
		return false
	case "raw":
		// MetaArgs (rust)
		return false
	default:
		return true
	}
}
