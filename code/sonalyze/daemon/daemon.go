// `sonalyze daemon` - HTTP server that runs sonalyze on behalf of a remote client
//
// This server responds to GET and POST requests carrying parameters that specify how to run
// sonalyze against a local data store.  The path for analysis commands is the sonalyze command
// name, eg, `GET /jobs?...` will run `sonalyze jobs`.  The path for add commands is either `POST
// /add?...` with the appropriate arguments, or for backward compatibility with existing infra, a
// keyword describing the data, eg `POST /sonar-freecsv?...` will run `sonalyze add -sample`.
//
// A query parameter `cluster=clusterName` is required for all requests, it names the cluster we're
// operating within and determines a bunch of file paths.
//
// Other parameter names are always the long parameter names for sonalyze and the parameter values
// are always urlencoded as necessary; parameter-less flags default to not-present.  Most parameters
// and names are forwarded to sonalyze, with eg --data-path and --config-file supplied by this code.
// The returned output is the raw output from sonalyze, whether for success or error.  A successful
// runs yields 2xx and an error yields 4xx or 5xx.
//
// Arguments:
//
// -jobanalyzer-dir <jobanalyzer-root-directory>
// -jobanalyzer-path <jobanalyzer-root-directory>
//
//  This is a required argument.  In the named directory there shall be:
//
//   - subdirectories `data` and `cluster-config`
//   - for each cluster CLUSTERNAME, a subdirectory `data/CLUSTERNAME` that has the sonar data
//     tree for the cluster
//   - for each cluster CLUSTERNAME, a file `cluster-config/CLUSTERNAME-config.json, which holds
//     the cluster description (machine configuration) for the cluster
//   - optionally a file `cluster-config/cluster-aliases.json`
//
//  The CLUSTERNAME is always the canonical cluster name.  Cluster names and the the json files are
//  described in production/jobanalyzer-server/cluster-config/README.md.
//
// -port <port-number>
//
//   This is an optional argument.  It is the port number on which to listen, the default is 8087.
//
// -analysis-auth <filename>
// -password-file <filename>
//
//   This is an optional argument.  It names a file with username:password pairs, one per line, to
//   be matched with values in an incoming HTTP basic authentication header for a GET operation.
//   (Note, if the connection is not HTTPS then the password may have been intercepted in transit.)
//
// -upload-auth <filename>
//
//   This is an optional but *strongly* recommended argument.  If provided then the file named must
//   provide username:password combinations, to be matched with one in an HTTP basic authentication
//   header.  (If the connection is not HTTPS then the password may have been intercepted in
//   transit.)
//
// -match-user-and-cluster
//
//   Optional but *strongly* recommended argument.  If set, and -upload-auth is also provided, then
//   the user name provided by the HTTP connection must match the cluster name in the data packet or
//   query string.  The effect is to make it possible for each cluster to have its own
//   username:password pair and for one cluster not to be able to upload data for another.
//
// -cache <size>
//
//   Cache raw or parboiled data in memory between operations.  The size is expressed as nnM
//   (megabytes) or nnG (gigabytes).  A sensible size *might* be about 256MB per 100 (slurm) nodes
//   per week.
//
// -no-add
//
//   This disables the /add, /sysinfo and /sonar-freecsv endpoints and the options -upload-auth and
//   -match-user-and-cluster.  The implication is that we're either running on a read-only database
//   or we're using -kafka to handle all ingestion.
//
// -kafka <broker-address>
//
//   EXPERIMENTAL.  The daemon will attempt to ingest data over a unencrypted and unauthenticated
//   Kafka channel for the clusters found in the data directory.  It should be the only consumer
//   for those data.  The broker-address is normally on the form hostname:port.
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

package daemon

import (
	_ "embed"
	"errors"
	"fmt"
	"io"

	"go-utils/auth"
	. "sonalyze/cmd"
)

const (
	defaultListenPort = 8087
	logTag            = "jobanalyzer/sonalyze"
	authRealm         = "Jobanalyzer remote access"
)

// MT: Immutable (no mutator operations) and thread-safe.
//
// This *will* be accessed concurrently b/c every HTTP handler runs as a separate goroutine and the
// handler invocations all share this.
type DaemonCommand struct {
	DevArgs
	VerboseArgs
	DatabaseArgs
	port                uint
	getAuthFile         string
	postAuthFile        string
	matchUserAndCluster bool
	kafkaBroker         string
	noAdd               bool

	getAuthenticator  *auth.Authenticator
	postAuthenticator *auth.Authenticator
	cmdlineHandler    CommandLineHandler
}

func New(cmdlineHandler CommandLineHandler) *DaemonCommand {
	dc := new(DaemonCommand)
	dc.cmdlineHandler = cmdlineHandler
	return dc
}

func (dc *DaemonCommand) Add(fs *CLI) {
	dc.DevArgs.Add(fs)
	dc.VerboseArgs.Add(fs)
	dc.DatabaseArgs.Add(fs, DBArgOptions{RequireFullDatabase: true})
	fs.Group("daemon-configuration")
	fs.UintVar(&dc.port, "port", defaultListenPort, "Listen for connections on `port`")
	fs.StringVar(&dc.getAuthFile, "analysis-auth", "", "Authentication info `filename` for analysis access")
	fs.StringVar(&dc.postAuthFile, "upload-auth", "", "Authentication info `filename` for data upload access")
	fs.BoolVar(&dc.matchUserAndCluster, "match-user-and-cluster", false, "Require user name to match cluster name")
	fs.StringVar(&dc.getAuthFile, "password-file", "", "Alias for -analysis-auth")
	fs.StringVar(&dc.kafkaBroker, "kafka", "", "Ingest data from this broker for all known clusters")
	fs.BoolVar(&dc.noAdd, "no-add", false, "Disable HTTPS ingestion")
}

//go:embed summary.txt
var summary string

func (dc *DaemonCommand) Summary(out io.Writer) {
	fmt.Fprint(out, summary)
}

func (dc *DaemonCommand) Validate() error {
	var e1, e2, e4, e5, e7, e8 error
	e1 = dc.DevArgs.Validate()
	e2 = dc.VerboseArgs.Validate()
	if dc.getAuthFile != "" {
		dc.getAuthenticator, e4 = auth.ReadPasswords(dc.getAuthFile)
		if e4 != nil {
			e4 = fmt.Errorf("Failed to read analysis authentication file: %v", e4)
		}
	}
	if dc.postAuthFile != "" {
		dc.postAuthenticator, e5 = auth.ReadPasswords(dc.postAuthFile)
		if e5 != nil {
			return fmt.Errorf("Failed to read upload authentication file: %v", e5)
		}
	}
	if dc.noAdd {
		if dc.matchUserAndCluster || dc.postAuthFile != "" {
			e8 = errors.New("The -no-add switch precludes https upload parameters")
		}
	}
	return errors.Join(e1, e2, e4, e5, e7, e8)
}
