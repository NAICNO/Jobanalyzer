// `sonalyzed` - HTTP server that runs sonalyze on behalf of a remote client
//
// This server responds to GET and POST requests carrying parameters that specify how to run
// sonalyze against a local data store.  The path for analysis commands is the sonalyze command
// name, eg, `GET /jobs?...` will run `sonalyze jobs`.  The path for add commands is a keyword
// describing the data (this is compatible with existing infra), eg `POST /sonar-freecsv?...` will
// run `sonalyze add -sample`.
//
// A query parameter `cluster=clusterName` is required for all requests, it names the cluster we're
// operating within and determines a bunch of file paths.
//
// Other parameter names are always the long parameter names for sonalyze and the parameter values
// are always urlencoded as necessary; boolean values must be the value defined as `magicBoolean`
// below.  Most parameters and names are forwarded to sonalyze, with eg --data-path and
// --config-file supplied by sonalyzed.  The returned output is the raw output from sonalyze,
// whether for success or error.  A successful runs yields 2xx and an error yields 4xx or 5xx.
//
// Arguments:
//
// -jobanalyzer-dir <jobanalyzer-root-directory>
// -jobanalyzer-path <jobanalyzer-root-directory>
//
//  This is a required argument.  In the named directory there shall be:
//
//   - the `sonalyze` executable
//   - optionally a file `cluster-aliases.json`, described below
//   - subdirectories `data` and `scripts`
//   - for each cluster, subdirectories `data/CLUSTERNAME` and `scripts/CLUSTERNAME`
//   - each subdirectory of `data` has the sonar data tree for the cluster
//   - each subdirectory of `scripts` has a file `CLUSTERNAME-config.json`, which holds the cluster
//     description (machine configuration).
//
// -port <port-number>
//
//  This is an optional argument.  It is the port number on which to listen, the default is 8087.
//
// -analysis-auth <filename>
// -password-file <filename>
//
//  This is an optional argument.  It names a file with username:password pairs, one per line, to be
//  matched with values in an incoming HTTP basic authentication header for a GET operation.  (Note,
//  if the connection is not HTTPS then the password may have been intercepted in transit.)
//
// -upload-auth <filename>
//
//  This is an optional but *strongly* recommended argument.  If provided then the file named must
//   provide username:password combinations, to be matched with one in an HTTP basic authentication
//   header.  (If the connection is not HTTPS then the password may have been intercepted in
//   transit.)
//
// -match-user-and-cluster
//   Optional but *strongly* recommended argument.  If set, and -upload-auth is also provided, then
//   the user name provided by the HTTP connection must match the cluster name in the data packet or
//   query string.  The effect is to make it possible for each cluster to have its own
//   username:password pair and for one cluster not to be able to upload data for another.
//
// Termination:
//
//  Sending SIGHUP or SIGTERM to sonalyzed will shut it down in an orderly manner.
//
//  sonalyzed is usually run in the background and exit codes are not easily examined, but when
//  sonalyzed exits it will deliver a non-zero exit code if an error was discovered during startup
//  or shutdown.
//
//  This server needs to stay up because it's the only contact point for all Sonalyze queries, and
//  it tries hard to avoid exiting or panicking.  However, this can happen.  Infrastructure should
//  exist to restart it if it crashes.
//
// Logging:
//
//  sonalyzed logs everything to the syslog with the tag defined below ("logTag").  Errors
//  encountered during startup are also logged to stderr.
//
// Cluster names and aliases:
//
//  Cluster names are the aliases of login nodes (fox.educloud.no for the UiO Fox supercomputer) or
//  synthesized names for a group of machines in the same family (mlx.hpc.uio.no for the UiO ML
//  nodes cluster).
//
//  The cluster alias file is a JSON array containing objects with "alias" and "value" fields:
//
//    [{"alias":"ml","value":"mlx.hpc.uio.no"}, ...]
//
//  so that the short name "ml" can be used to name the cluster "mlx.hpc.uio.no" in requests.

package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"syscall"

	"go-utils/alias"
	"go-utils/auth"
	"go-utils/httpsrv"
	"go-utils/options"
	"go-utils/process"
	"go-utils/status"
)

const (
	defaultListenPort      = 8087
	clusterAliasesFilename = "cluster-aliases.json"
	logTag                 = "jobanalyzer/sonalyzed"
	authRealm              = "Jobanalyzer remote access"
	magicBoolean           = "xxxxxtruexxxxx"
)

var (
	jobanalyzerDir      = flag.String("jobanalyzer-dir", "", "Jobanalyzer root `directory` (required)")
	port                = flag.Int("port", defaultListenPort, "Listen for connections on `port`")
	getAuthFile         = flag.String("analysis-auth", "", "Authentication info `filename` for analysis access")
	postAuthFile        = flag.String("upload-auth", "", "Authentication info `filename` for data upload access")
	matchUserAndCluster = flag.Bool("match-user-and-cluster", false, "Require user name to match cluster name")
	verbose             = flag.Bool("v", false, "Verbose logging")
)

func init() {
	flag.StringVar(jobanalyzerDir, "jobanalyzer-path", "", "Alias for -jobanalyzer-dir")
	flag.StringVar(getAuthFile, "password-file", "", "Alias for -analysis-auth")
}

var (
	aliasResolver     *alias.Aliases
	getAuthenticator  *auth.Authenticator
	postAuthenticator *auth.Authenticator
)

func main() {
	status.Start(logTag)
	flag.Parse()

	var err error
	*jobanalyzerDir, err = options.RequireDirectory(*jobanalyzerDir, "-jobanalyzer-path")
	if err != nil {
		status.Fatal(err.Error())
	}
	if *getAuthFile != "" {
		getAuthenticator, err = auth.ReadPasswords(*getAuthFile)
		if err != nil {
			status.Fatalf("Failed to read analysis authentication file: %v\n", err)
		}
	}
	if *postAuthFile != "" {
		postAuthenticator, err = auth.ReadPasswords(*postAuthFile)
		if err != nil {
			status.Fatalf("Failed to read upload authentication file: %v\n", err)
		}
	}
	aliasResolver, err = alias.ReadAliases(path.Join(*jobanalyzerDir, clusterAliasesFilename))
	if err != nil {
		status.Warning(err.Error())
	}

	http.HandleFunc("/add", httpAddHandler())
	http.HandleFunc("/jobs", httpGetHandler("jobs"))
	http.HandleFunc("/load", httpGetHandler("load"))
	http.HandleFunc("/uptime", httpGetHandler("uptime"))
	http.HandleFunc("/profile", httpGetHandler("profile"))
	http.HandleFunc("/parse", httpGetHandler("parse"))
	http.HandleFunc("/metadata", httpGetHandler("metadata"))
	// These request names are compatible with the older `infiltrate` and with the upload infra
	// already running on the clusters.
	http.HandleFunc("/sonar-freecsv", httpPostHandler("sample", "text/csv"))
	http.HandleFunc("/sysinfo", httpPostHandler("sysinfo", "application/json"))

	var programFailed bool
	s := httpsrv.New(*verbose, *port, func(err error) {
		programFailed = true
	})
	go s.Start()

	// Wait here until we're stopped by SIGHUP (manual) or SIGTERM (from OS during shutdown).
	//
	// TODO: For SIGHUP, we should not exit but should instead reread the password file and the
	// cluster aliases file.
	process.WaitForSignal(syscall.SIGHUP, syscall.SIGTERM)
	s.Stop()

	if programFailed {
		os.Exit(1)
	}
}

// HTTP handlers
//
// Documented behavior: the server will close the request body, we don't need to do it.
//
// I can find no documentation about needing to consume the body in case of an early (error)
// return, nor anything obvious in the net/http source code to indicate this, nor has google
// turned up anything.  So request handler code assumes it's not necessary.

func httpGetHandler(
	command string,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, _, clusterName, ok := requestPreamble(w, r, "GET", getAuthenticator, authRealm, "")
		if !ok {
			return
		}

		arguments := []string{
			command,
			"--data-path",
			path.Join(*jobanalyzerDir, "data", clusterName),
		}
		for name, vs := range r.URL.Query() {
			if name == "cluster" {
				continue
			}

			if !argOk(command, name) {
				w.WriteHeader(400)
				fmt.Fprintf(w, "Bad parameter %s", name)
				if *verbose {
					status.Warningf("Bad parameter %s", name)
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
				path.Join(*jobanalyzerDir, "scripts", clusterName, clusterName+"-config.json"),
			)
		}

		stdout, ok := runSonalyze(w, arguments, []byte{})
		if !ok {
			return
		}

		w.WriteHeader(200)
		fmt.Fprint(w, stdout)
	}
}

func httpAddHandler() func(http.ResponseWriter, *http.Request) {
	forSample := httpPostHandler("sample", "text/csv")
	forSysinfo := httpPostHandler("sysinfo", "application/json")
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
			if *verbose {
				status.Warningf("Bad operation: %s", err.Error())
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
	dataType string,
	contentType string,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, userName, clusterName, ok := requestPreamble(w, r, "POST", postAuthenticator, "", contentType)
		if !ok {
			return
		}

		if *matchUserAndCluster && userName != "" && clusterName != userName {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Upload not authorized")
			if *verbose {
				status.Warningf("Upload not authorized")
			}
			return
		}

		arguments := []string{
			"add",
			"--" + dataType,
			"--data-path",
			path.Join(*jobanalyzerDir, "data", clusterName),
		}

		stdout, ok := runSonalyze(w, arguments, payload)
		if !ok {
			return
		}

		w.WriteHeader(200)
		fmt.Fprint(w, stdout)
	}
}

func requestPreamble(
	w http.ResponseWriter,
	r *http.Request,
	method string,
	authenticator *auth.Authenticator,
	realm string,
	contentType string,
) (payload []byte, userName, clusterName string, ok bool) {
	if *verbose {
		// Header reveals auth info, don't put it into logs
		status.Infof("Request from %s: %v", r.RemoteAddr, r.URL.String())
	}

	if !httpsrv.AssertMethod(w, r, method, *verbose) {
		return
	}

	authOk, userName := httpsrv.Authenticate(w, r, authenticator, realm, *verbose)
	if !authOk {
		return
	}

	payload, havePayload := httpsrv.ReadPayload(w, r, *verbose)
	if !havePayload {
		return
	}

	if contentType != "" {
		if !httpsrv.AssertContentType(w, r, contentType, *verbose) {
			return
		}
	}

	clusterValues, found := r.URL.Query()["cluster"]
	if !found || len(clusterValues) != 1 || clusterValues[0] == "" {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Bad parameters - missing or empty or repeated 'cluster'")
		if *verbose {
			status.Warningf("Bad parameters - missing or empty or repeated 'cluster'")
		}
		return
	}

	clusterName = clusterValues[0]
	if aliasResolver != nil {
		clusterName = aliasResolver.Resolve(clusterName)
	}

	ok = true
	return
}

func runSonalyze(w http.ResponseWriter, arguments []string, input []byte) (stdout string, ok bool) {
	// Run the command and report the result

	if *verbose {
		status.Infof(
			"Command: %s %s",
			path.Join(*jobanalyzerDir, "sonalyze"),
			strings.Join(arguments, " "),
		)
	}
	stdout, stderr, err := process.RunSubprocessWithInput(
		"sonalyze",
		path.Join(*jobanalyzerDir, "sonalyze"),
		arguments,
		bytes.NewReader(input),
	)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, err.Error())
		if stderr != "" {
			fmt.Fprint(w, "\n", stderr)
		}
		if *verbose {
			status.Warningf("ERROR: %v", err)
		}
		return
	}
	if stderr != "" {
		status.Warning(stderr)
	}

	ok = true
	return
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
