// `sonalyze daemon` - HTTP server that runs sonalyze on behalf of a remote client
//
// This server responds to GET and POST requests carrying parameters that specify how to run
// sonalyze against a local data store.  See doc/HOWTO-RESTAPI.md.
//
// Arguments:
//
// -jobanalyzer-dir <jobanalyzer-root-directory>
// -jobanalyzer-path <jobanalyzer-root-directory>
//
//  This is a required argument.  In the named directory there shall be:
//
//   - subdirectories `data` and `cluster-config` (though see -database-uri below)
//   - for each cluster CLUSTERNAME, a subdirectory `data/CLUSTERNAME` that has the sonar data
//     tree for the cluster (ditto)
//   - for each cluster CLUSTERNAME, a file `cluster-config/CLUSTERNAME-config.json, which holds
//     the cluster description (machine configuration) for the cluster
//   - optionally a file `cluster-config/cluster-aliases.json`
//
//  The CLUSTERNAME is always the canonical cluster name.  Cluster names and the the json files are
//  described in production/jobanalyzer-server/cluster-config/README.md.
//
// -database-uri <uri>
//
//  If present, this specifies a database access point.  The database is used for data access rather
//  than the data/ subdirectory of the jobanalyzer directory.
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
// -cache <size>
//
//   Cache raw or parboiled data in memory between operations.  The size is expressed as nnM
//   (megabytes) or nnG (gigabytes).  A sensible size *might* be about 256MB per 100 (slurm) nodes
//   per week.
//
// -kafka <broker-address>
//
//   EXPERIMENTAL.  The daemon will attempt to ingest data over a unencrypted and unauthenticated
//   Kafka channel for the clusters found in the data directory.  It should be the only consumer
//   for those data.  The broker-address is normally on the form hostname:port.
//
// -rest-api <interface>
//
//   The daemon will present various APIs on the given interface (in the form interface:port,
//   e.g. "localhost:8888").  Access the /openapi.json or /openapi.yaml endpoint on that interface
//   to retrieve API documentation.  Normally under /api/v0 there will be the old sonalyze API (so
//   /api/v0/jobs corresponds to the old /jobs API), under /api/v1 there will be a "clean" API more
//   or less aligned with the v0 API but with clean JSON output and not the idiosyncrasies of v0,
//   and under /api/v2 there is a subset of the slurm-monitor REST API v2.
//
//   The individual APIs are all disabled by default.  Pass -v0, -v1, and -v2 to enable them.
//
// -v0
//
//   Enable the v0 API.
//
// -v1
//
//   Enable the v1 API.
//
// -v2
//
//   Enable the v2 API.  Note this does not use the analysis-auth, having been set up for OAUTH
//   authentication, and may require additional work to be safe.
//
// -insert
//
//   Enable the /api/v1/insert points in the REST API.  Normally this API is enabled only when
//   running without a -database-uri and without -kafka (though it is not incompatible with the
//   latter).
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
	"fmt"
	"io"

	"go-utils/auth"
	. "sonalyze/cmd"
)

const (
	defaultListenPort = 8087
	logTag            = "jobanalyzer/sonalyze"
)

// MT: Immutable (no mutator operations) and thread-safe.
//
// This *will* be accessed concurrently b/c every HTTP handler runs as a separate goroutine and the
// handler invocations all share this.
type DaemonCommand struct {
	DevArgs
	VerboseArgs
	DatabaseArgs
	getAuthFile  string
	postAuthFile string
	kafkaBroker  string
	restAPI      string
	insert       bool
	v0           bool
	v1           bool
	v2           bool

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
	fs.StringVar(&dc.getAuthFile, "analysis-auth", "", "Authentication info `filename` for analysis access")
	fs.StringVar(&dc.postAuthFile, "upload-auth", "", "Authentication info `filename` for data upload access")
	fs.StringVar(&dc.getAuthFile, "password-file", "", "Alias for -analysis-auth")
	fs.StringVar(&dc.kafkaBroker, "kafka", "", "Ingest data from this broker for all known clusters")
	fs.StringVar(&dc.restAPI, "rest-api", "", "Serve /api/v0, /api/v1 and /api/v2 on this interface:port")
	fs.BoolVar(&dc.insert, "insert", false, "Enable the /api/v1/insert points")
	fs.BoolVar(&dc.v0, "v0", false, "Enable the v0 API")
	fs.BoolVar(&dc.v1, "v1", false, "Enable the v1 API")
	fs.BoolVar(&dc.v2, "v2", false, "Enable the v2 API")
}

//go:embed summary.txt
var summary string

func (dc *DaemonCommand) Summary(out io.Writer) {
	fmt.Fprint(out, summary)
}

func (dc *DaemonCommand) Validate() error {
	if err := dc.DevArgs.Validate(); err != nil {
		return err
	}
	if err := dc.VerboseArgs.Validate(); err != nil {
		return err
	}
	if dc.getAuthFile != "" {
		var err error
		dc.getAuthenticator, err = auth.ReadPasswords(dc.getAuthFile)
		if err != nil {
			return fmt.Errorf("Failed to read analysis authentication file: %v", err)
		}
	}
	if dc.postAuthFile != "" {
		var err error
		dc.postAuthenticator, err = auth.ReadPasswords(dc.postAuthFile)
		if err != nil {
			return fmt.Errorf("Failed to read upload authentication file: %v", err)
		}
	}
	if dc.insert && dc.DatabaseURI() != "" {
		return fmt.Errorf("Can't have both -database-uri and -insert")
	}
	return nil
}

func (dc *DaemonCommand) ReifyForRemote(x *ArgReifier) error {
	panic("Daemon is not remotable")
}
