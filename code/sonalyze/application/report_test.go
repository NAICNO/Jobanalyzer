package application

import (
	"strings"
	"testing"

	"sonalyze/cmd"
	"sonalyze/cmd/report"
	"sonalyze/db/special"
)

func TestReport(t *testing.T) {
	var expect = []string{
		`["ml1.hpc.uio.no","ml2.hpc.uio.no","ml3.hpc.uio.no","ml4.hpc.uio.no","ml6.hpc.uio.no","ml9.hpc.uio.no"]`,
		``,
	}
	var rc report.ReportCommand
	rc.DatabaseArgs.SetReportDir("testdata/report_test", "report.cluster")
	rc.ReportName = "ml-hostnames.json"
	err := rc.Validate()
	if err != nil {
		t.Fatal(err)
	}
	err = cmd.OpenDataStoreFromCommand(&rc)
	if err != nil {
		t.Fatal(err)
	}
	defer special.CloseDataStore()
	var stdout strings.Builder
	err = rc.Perform(cmd.NewMetaFromCommand(&rc), nil, &stdout, nil)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(stdout.String(), "\n")
	if len(lines) != len(expect) {
		t.Log(lines)
		t.Fatalf("Length: got %d expected %d", len(lines), len(expect))
	}
	for i, e := range expect {
		if lines[i] != e {
			t.Fatalf("Line %d: got %s", i, lines[i])
		}
	}
}
