package application

import (
	"strings"
	"testing"

	"sonalyze/cmd/report"
)

func TestReport(t *testing.T) {
	var expect = []string{
		`["ml1.hpc.uio.no","ml2.hpc.uio.no","ml3.hpc.uio.no","ml4.hpc.uio.no","ml6.hpc.uio.no","ml9.hpc.uio.no"]`,
		``,
	}
	var rc report.ReportCommand
	rc.ReportDir = "testdata/report_test"
	rc.ReportName = "ml-hostnames.json"
	err := rc.Validate()
	if err != nil {
		t.Fatal(err)
	}
	var stdout strings.Builder
	err = rc.Perform(nil, &stdout, nil)
	lines := strings.Split(stdout.String(), "\n")
	if len(lines) != len(expect) {
		t.Fatalf("Length: got %d", len(lines))
	}
	for i, e := range expect {
		if lines[i] != e {
			t.Fatalf("Line %d: got %s", i, lines[i])
		}
	}
}
