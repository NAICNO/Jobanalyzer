// `sonalyze report` will serve a static file from the report directory for the cluster, with access
// controls.  Eventually the access control here will continue to require superuser privileges, as
// there are no per-user reports.
//
// The option -report-name selects a file by full name in the data directory.  (The file name
// extension will determine the mime type in the remote case).  Normally these will be csv, txt or
// json files.  Any file in that directory is fair game, even the txt files.  Subdirectories are
// excluded for now.
//
// When running locally, we use -report-dir which should point to a directory that has the report
// files.  -cluster is not allowed.
//
// When running remotely, we will use -cluster to select the cluster, and the daemon code will
// compute the data directory using standard db abstractions.

package report

import (
	"errors"
	"flag"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"

	. "sonalyze/cmd"
	. "sonalyze/common"
)

type ReportCommand struct {
	DevArgs
	RemotingArgs
	VerboseArgs
	ReportDir string
	ReportName string			// This must be a plain filename
}

var _ = (SimpleCommand)((*ReportCommand)(nil))

func (rc *ReportCommand) Summary() []string {
	return []string{
		"Extract pre-rendered reports",
	}
}

func (rc *ReportCommand) Add(fs *flag.FlagSet) {
	rc.DevArgs.Add(fs)
	rc.RemotingArgs.Add(fs)
	rc.VerboseArgs.Add(fs)
	fs.StringVar(
		&rc.ReportDir, "report-dir", "", "`directory-name` containing reports (precludes -remote)")
	fs.StringVar(&rc.ReportName, "report-name", "", "`filename` of the report to extract")
}

func (rc *ReportCommand) ReifyForRemote(x *ArgReifier) error {
	// This is normally done by SourceArgs
	x.String("cluster", rc.RemotingArgs.Cluster)

	// Do not forward ReportDir, though it should be "" anyway.
	x.String("report-name", rc.ReportName)

	// As per normal, do not forward VerboseArgs.
	return rc.DevArgs.ReifyForRemote(x)
}

var filenameRe = regexp.MustCompile(`^[a-zA-Z_0-9.-]+$`)

func (rc *ReportCommand) Validate() error {
	if rc.ReportName == "" {
		return errors.New("A value for -report-name is required")
	}
	// This attempts to reject anything with a path and drive component as well as anything
	// considered a special file by the OS.  It's pretty conservative.  It is possible that some
	// combination of fs.ValidPath + filepath.Localize is just as good and more permissive, it would
	// allow for files in subdirectories of the report directory too.
	if !filenameRe.MatchString(rc.ReportName) || !filepath.IsLocal(rc.ReportName) {
		return errors.New("Illegal report file name")
	}

	if rc.ReportDir == "" {
		ApplyDefault(&rc.Remote, "data-source", "remote")
		ApplyDefault(&rc.AuthFile, "data-source", "auth-file")
		ApplyDefault(&rc.Cluster, "data-source", "cluster")
	}

	return errors.Join(
		rc.DevArgs.Validate(),
		rc.RemotingArgs.Validate(),
		rc.VerboseArgs.Validate(),
	)
}

func (rc *ReportCommand) Perform(_ io.Reader, stdout, _ io.Writer) error {
	// ReportDir will have a value that is safe if from remote invocation
	// ReportName will have a safe value
	fn := path.Join(rc.ReportDir, rc.ReportName)
	file, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer file.Close()
	_, _ = io.Copy(stdout, file)
	return nil
}
