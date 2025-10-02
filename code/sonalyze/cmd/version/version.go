package version

import (
	_ "embed"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	. "sonalyze/cmd"
)

type VersionCommand struct {
	// Add these to make this into a SimpleCommand
	RemotingArgsNoCluster
	DevArgs
	VerboseArgs
}

var _ = (SimpleCommand)((*VersionCommand)(nil))
var _ = (RemotableCommand)((*VersionCommand)(nil))

func (vc *VersionCommand) Add(fs *CLI) {
	vc.RemotingArgsNoCluster.Add(fs)
	vc.DevArgs.Add(fs)
	vc.VerboseArgs.Add(fs)
}

func (vc *VersionCommand) ReifyForRemote(x *ArgReifier) error {
	return vc.DevArgs.ReifyForRemote(x)
}

func (vc *VersionCommand) Validate() error {
	return vc.RemotingArgsNoCluster.Validate()
}

func (vc *VersionCommand) Summary(out io.Writer) {
	fmt.Fprintf(out, "Display the version number.")
}

// The version data are version,description
// They are newest-first; we always want the first line.
//
//go:embed version.csv
var versionData string

func (_ *VersionCommand) Perform(_ io.Reader, stdout, _ io.Writer) error {
	version := "0.0.0"

	rdr := csv.NewReader(strings.NewReader(versionData))
	rdr.FieldsPerRecord = -1 // Free form, though should not matter
	fields, err := rdr.Read()
	if err == nil && len(fields) >= 1 {
		version = fields[0]
	}

	// Must print version on stdout, and the features() thing is required by some tests.
	// "short" indicates that we're only parsing the first 8 fields (v0.6.0 data).
	fmt.Fprintf(stdout, "sonalyze-go version(%s) features(short_untagged_sonar_data)\n", version)
	return nil
}
