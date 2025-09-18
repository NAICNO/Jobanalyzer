package report

import (
	_ "embed"
	"errors"
	"fmt"
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
	ReportDir  string
	ReportName string // This must be a plain filename
}

var _ = (SimpleCommand)((*ReportCommand)(nil))

//go:embed summary.txt
var summary string

func (rc *ReportCommand) Summary(out io.Writer) {
	fmt.Fprint(out, summary)
}

func (rc *ReportCommand) Add(fs *CLI) {
	rc.DevArgs.Add(fs)
	rc.RemotingArgs.Add(fs)
	rc.VerboseArgs.Add(fs)

	fs.Group("local-data-source")
	fs.StringVar(
		&rc.ReportDir, "report-dir", "", "`directory-name` containing reports (precludes -remote)")

	fs.Group("application-control")
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
		return errors.New("Illegal file name for -report-name")
	}

	if rc.ReportDir == "" {
		ApplyDefault(&rc.Remote, DataSourceRemote)
		if os.Getenv("SONALYZE_AUTH") == "" {
			ApplyDefault(&rc.AuthFile, DataSourceAuthFile)
		}
		ApplyDefault(&rc.Cluster, DataSourceCluster)
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
