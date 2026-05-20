// See ../doc/TECHNICAL.md for a definition of the protocol.
//
// When adding a new command to the daemon, several points in the API implementations have to be
// updated, see comments in the subdirectories.

package daemon

import (
	"fmt"
	"io"
	"log/syslog"
	"syscall"

	"go-utils/process"
	. "sonalyze/common"
	"sonalyze/daemon/api0"
	"sonalyze/daemon/api1"
	"sonalyze/daemon/api2"
	"sonalyze/daemon/apiutil"
	"sonalyze/db"
	"sonalyze/db/special"
)

// Note, this should *NOT* be called Perform(), so that we can be sure we're not confusing a
// DaemonCommand with a SimpleCommand.

func (dc *DaemonCommand) RunDaemon(_ io.Reader, _, stderr io.Writer) error {
	logger, err := syslog.Dial("", "", syslog.LOG_INFO|syslog.LOG_USER, logTag)
	if err != nil {
		return fmt.Errorf("FATAL ERROR: Failing to open logger: %v", err)
	}
	Log.SetUnderlying(logger)

	if dc.kafkaBroker != "" {
		for _, cl := range special.AllClusters() {
			meta := db.NewContextFromCluster(cl)
			var ds db.AppendablePersistentDataProvider
			var err error
			if meta.HaveDatabaseConnection() {
				ds = db.OpenConnectedDB(meta)
			} else {
				ds, err = db.OpenAppendablePersistentDirectoryDB(meta)
			}
			if err != nil {
				if Verbose {
					Log.Warningf("Failed to open data store for %s", cl.Name)
				}
				continue
			}
			if Verbose {
				Log.Infof("Starting listener for %s", cl.Name)
			}
			go runKafka(dc.kafkaBroker, cl.Name, dc.consumerGroup, ds)
		}
	}

	if dc.restAPI != "" {
		api := apiutil.CreateAPI(dc.restAPI)
		if dc.v0 {
			api0.SetupAPI(
				api,
				dc.JobanalyzerDir(),
				dc.DatabaseURI(),
				dc.cmdlineHandler,
				dc.getAuthenticator,
			)
		}
		if dc.v1 {
			api1.SetupAPI(
				api,
				dc.insert,
				dc.postAuthenticator,
			)
		}
		if dc.v2 {
			api2.SetupAPI(api)
		}
		apiutil.RunAPI()
	}

	process.WaitForSignal(syscall.SIGHUP, syscall.SIGTERM)

	if dc.restAPI != "" {
		apiutil.StopAPI()
	}

	return nil
}
