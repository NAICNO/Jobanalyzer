// `sonalyze daemon` - HTTP server that runs sonalyze on behalf of a remote client
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
// --config-file supplied by this code.  The returned output is the raw output from sonalyze,
// whether for success or error.  A successful runs yields 2xx and an error yields 4xx or 5xx.
//
// Arguments:
//
// -jobanalyzer-dir <jobanalyzer-root-directory>
// -jobanalyzer-path <jobanalyzer-root-directory>
//
//  This is a required argument.  In the named directory there shall be:
//
//   - subdirectories `data` and `scripts`
//   - for each cluster, subdirectories `data/CLUSTERNAME` and `scripts/CLUSTERNAME`
//   - each subdirectory of `data` has the sonar data tree for the cluster
//   - each subdirectory of `scripts` has a file `CLUSTERNAME-config.json`, which holds the cluster
//     description (machine configuration).
//   - optionally a file `cluster-aliases.json`, described below
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
//  Sending SIGHUP or SIGTERM to `sonalyze daemon` will shut it down in an orderly manner.
//
//  The daemon is usually run in the background and exit codes are not easily examined, but when
//  the daemon exits it will deliver a non-zero exit code if an error was discovered during startup
//  or shutdown.
//
//  This server needs to stay up because it's the only contact point for all Sonalyze queries, and
//  it tries hard to avoid exiting or panicking.  However, this can happen.  Infrastructure should
//  exist to restart it if it crashes.
//
// Logging:
//
//  The daemon logs everything to the syslog with the tag defined below ("logTag").  Errors
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

package daemon

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"

	"go-utils/alias"
	"go-utils/auth"
	"go-utils/options"
	. "sonalyze/command"
)

const (
	defaultListenPort      = 8087
	clusterAliasesFilename = "cluster-aliases.json"
	logTag                 = "jobanalyzer/sonalyze"
	authRealm              = "Jobanalyzer remote access"
	magicBoolean           = "xxxxxtruexxxxx"
)

// Immutable (no mutator operations) and thread-safe.  It *will* be accessed concurrently b/c every
// HTTP handler runs as a separate goroutine.
type DaemonCommand struct {
	DevArgs
	VerboseArgs
	jobanalyzerDir      string
	port                uint
	getAuthFile         string
	postAuthFile        string
	matchUserAndCluster bool
	cache               string

	aliasResolver     *alias.Aliases
	getAuthenticator  *auth.Authenticator
	postAuthenticator *auth.Authenticator
	cmdlineHandler    CommandLineHandler
	cacheSize         int64
}

func New(cmdlineHandler CommandLineHandler) *DaemonCommand {
	dc := new(DaemonCommand)
	dc.cmdlineHandler = cmdlineHandler
	return dc
}

func (dc *DaemonCommand) Add(fs *flag.FlagSet) {
	dc.DevArgs.Add(fs)
	dc.VerboseArgs.Add(fs)
	fs.StringVar(&dc.jobanalyzerDir, "jobanalyzer-dir", "", "Jobanalyzer root `directory` (required)")
	fs.UintVar(&dc.port, "port", defaultListenPort, "Listen for connections on `port`")
	fs.StringVar(&dc.getAuthFile, "analysis-auth", "", "Authentication info `filename` for analysis access")
	fs.StringVar(&dc.postAuthFile, "upload-auth", "", "Authentication info `filename` for data upload access")
	fs.BoolVar(&dc.matchUserAndCluster, "match-user-and-cluster", false, "Require user name to match cluster name")
	fs.StringVar(&dc.jobanalyzerDir, "jobanalyzer-path", "", "Alias for -jobanalyzer-dir")
	fs.StringVar(&dc.getAuthFile, "password-file", "", "Alias for -analysis-auth")
	fs.StringVar(&dc.cache, "cache", "", "Enable data caching with this size (nM for megs, nG for gigs)")
}

func (dc *DaemonCommand) Summary() []string {
	return []string{
		"Run sonalyze as an HTTP server that responds to GET and POST for data",
		"extraction and update.  See manual for more information.",
	}
}

func (dc *DaemonCommand) Validate() error {
	var e1, e2, e3, e4, e5, e6, e7 error
	e1 = dc.DevArgs.Validate()
	e2 = dc.VerboseArgs.Validate()
	dc.jobanalyzerDir, e3 = options.RequireDirectory(dc.jobanalyzerDir, "-jobanalyzer-path")
	if dc.getAuthFile != "" {
		dc.getAuthenticator, e4 = auth.ReadPasswords(dc.getAuthFile)
		if e4 != nil {
			e4 = fmt.Errorf("Failed to read analysis authentication file %w", e4)
		}
	}
	if dc.postAuthFile != "" {
		dc.postAuthenticator, e5 = auth.ReadPasswords(dc.postAuthFile)
		if e5 != nil {
			return fmt.Errorf("Failed to read upload authentication file %w", e5)
		}
	}
	// The aliases file is optional, but if something with that name is there it is an error to fail
	// to open it.
	aliasesFile := path.Join(dc.jobanalyzerDir, clusterAliasesFilename)
	if info, err := os.Stat(aliasesFile); err == nil {
		if info.Mode()&fs.ModeType != 0 {
			e6 = errors.New("Cluster alias file is not a regular file")
		} else {
			dc.aliasResolver, e6 = alias.ReadAliases(aliasesFile)
		}
	}
	if dc.cache != "" {
		var size int64
		if n, _ := fmt.Sscanf(dc.cache, "%dM", &size); n == 1 && size > 0 {
			dc.cacheSize = size * 1024 * 1024
		} else if n, _ := fmt.Sscanf(dc.cache, "%dG", &size); n == 1 && size > 0 {
			dc.cacheSize = size * 1024 * 1024 * 1024
		} else {
			e7 = errors.New("Bad -cache value")
		}
	}
	return errors.Join(e1, e2, e3, e4, e5, e6, e7)
}
