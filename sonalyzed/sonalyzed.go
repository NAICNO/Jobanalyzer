// `sonalyzed` - HTTP server that runs sonalyze on behalf of a remote client
//
// This server responds to GET requests carrying parameters that specify how to run sonalyze against
// a local data store.  The path is the sonalyze command name, eg, `GET /jobs?...` will run
// `sonalyze jobs`.  The parameter names are always the long parameter names for sonalyze and the
// parameter values are always urlencoded as necessary; boolean values must be the value defined as
// `magicBoolean` below.  Most parameters and names are forwarded to sonalyze with eg --data-path
// and --config-file supplied by sonalyzed.  The returned output is the raw output from sonalyze,
// whether for success or error.  A successful runs yields 2xx and an error yields 4xx or 5xx.
//
// Arguments:
//
// -jobanalyzer-path <jobanalyzer-root-directory>
//
//  This is a required argument.  In the named directory there shall be:
//
//   - the `sonalyze` executable
//   - subdirectories `data` and `scripts`
//   - for each cluster, subdirectories `data/CLUSTERNAME` and `scripts/CLUSTERNAME`
//   - each subdirectory of `data` has the sonar data tree for the cluster
//   - each subdirectory of `scripts has a file `CLUSTERNAME-config.json`, which holds the cluster
//     description.
//
// -port <port-number>
//
//  This is an optional argument.  It is the port number on which to listen, the default is 8087.
//
// -password-file <filename>
//
//  This is an optional argument.  It names a file with username:password pairs, one per line, to be
//  matched with values in an incoming HTTP basic authentication header.  (Note, if the connection
//  is not HTTPS then the password may have been intercepted in transit.)
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

package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"syscall"

	"go-utils/auth"
	"go-utils/httpsrv"
	"go-utils/options"
	"go-utils/process"
	"go-utils/status"
)

const (
	defaultListenPort = 8087

	logTag = "jobanalyzer/sonalyzed"

	// This must equal MAGIC_BOOLEAN in the sonalyze sources.
	magicBoolean = "xxxxxtruexxxxx"
)

var verbose bool
var programFailed = false

func main() {
	status.Start(logTag)

	port, jobanalyzerPath, passwordFile, err := commandLine()
	if err != nil {
		status.Fatalf("Command line: %v", err)
	}

	var authenticator func(user, pass string) bool
	if passwordFile != "" {
		authenticator, err = auth.ParsePasswdFile(passwordFile)
		if err != nil {
			status.Fatalf("Failed to read password file: %v\n", err)
		}
	}

	http.HandleFunc("/jobs", requestHandler("jobs", jobanalyzerPath, authenticator))
	http.HandleFunc("/load", requestHandler("load", jobanalyzerPath, authenticator))
	http.HandleFunc("/uptime", requestHandler("uptime", jobanalyzerPath, authenticator))
	http.HandleFunc("/profile", requestHandler("profile", jobanalyzerPath, authenticator))
	http.HandleFunc("/parse", requestHandler("parse", jobanalyzerPath, authenticator))
	http.HandleFunc("/metadata", requestHandler("metadata", jobanalyzerPath, authenticator))

	s := httpsrv.New(verbose, port, func(err error) {
		programFailed = true
	})
	go s.Start()

	// Wait here until we're stopped by SIGHUP (manual) or SIGTERM (from OS during shutdown).
	process.WaitForSignal(syscall.SIGHUP, syscall.SIGTERM)
	s.Stop()

	if programFailed {
		os.Exit(1)
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
	if len(arg) <= 0 {
		return false
	}

	// Specific names are excluded, for now, the names in the comments relate to structure names in
	// sonalyze.rs.
	switch arg {
	case "data-path", "cluster", "remote", "auth-file":
		// SourceArgs
		return false
	case "config-file":
		// ConfigFileArg
		return false
	case "raw":
		// MetaArgs
		return false
	default:
		return true
	}
}

func requestHandler(
	command, jobanalyzerPath string,
	authenticator func(user, pass string) bool,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if verbose {
			status.Infof("Request from %s: %v", r.RemoteAddr, r.Header)
			status.Infof("%v", r.URL)
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

		if r.Method != "GET" {
			w.WriteHeader(403)
			fmt.Fprintf(w, "Bad method")
			if verbose {
				status.Warningf("Bad method: %s", r.Method)
			}
			return
		}

		user, pass, ok := r.BasicAuth()
		passed := !ok && authenticator == nil || ok && authenticator != nil && authenticator(user, pass)
		if !passed {
			if authenticator != nil {
				w.Header().Add("WWW-Authenticate", "Basic realm=\"Jobanalyzer remote access\", charset=\"utf-8\"")
			}
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

		// The parameter `cluster` provides the cluster name, which is needed for the data directory
		// and the config file.

		arguments := []string{command}
		clusterName := ""

		for name, vs := range r.URL.Query() {
			if name == "cluster" && len(vs) == 1 {
				clusterName = vs[0]
				continue
			}

			// This will handle broken "cluster" arguments too.
			if !argOk(command, name) {
				w.WriteHeader(400)
				fmt.Fprintf(w, "Bad parameter %s", name)
				if verbose {
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

		if clusterName == "" {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Missing `cluster`")
			if verbose {
				status.Warning("Missing `cluster`")
			}
			return
		}
		clusterName = resolveClusterAlias(clusterName)

		arguments = append(
			arguments,
			"--data-path",
			path.Join(jobanalyzerPath, "data", clusterName),
		)
		switch command {
		case "job", "load":
			arguments = append(
				arguments,
				"--config-file",
				path.Join(jobanalyzerPath, "scripts", clusterName, clusterName+"-config.json"),
			)
		}

		// Run the command and report the result

		output, err := process.RunSubprocess(path.Join(jobanalyzerPath, "sonalyze"), arguments)
		if err != nil {
			w.WriteHeader(400)
			fmt.Fprint(w, err.Error())
			if verbose {
				status.Warningf("ERROR: %v", err)
			}
			return
		}

		w.WriteHeader(200)
		fmt.Fprint(w, output)
	}
}

func resolveClusterAlias(clusterName string) string {
	// TODO: This expansion should be in a config file
	switch clusterName {
	case "ml":
		return "mlx.hpc.uio.no"
	case "fox":
		return "fox.educloud.no"
	default:
		return clusterName
	}
}

func commandLine() (port int, jobanalyzerPath, passwordFile string, err error) {
	flags := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flags.IntVar(&port, "port", defaultListenPort, "Listen for connections on `port`")
	flags.StringVar(&jobanalyzerPath, "jobanalyzer-path", "", "Path of jobanalyzer root `directory` (required)")
	flags.StringVar(&passwordFile, "password-file", "", "Read user names and passwords from `filename`")
	flags.BoolVar(&verbose, "v", false, "Verbose logging")
	err = flags.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		os.Exit(0)
	}
	if err != nil {
		return
	}
	jobanalyzerPath, err = options.RequireDirectory(jobanalyzerPath, "-jobanalyzer-path")
	return
}
