package application

import (
	"strings"
	"testing"

	"sonalyze/cmd"
)

func testSimpleCommand(t *testing.T, command cmd.SimpleCommand, fields string, expect []string) {
	err := command.Validate()
	if err != nil {
		t.Fatal(err)
	}
	var stdout strings.Builder
	err = command.Perform(cmd.NewMeta(command), nil, &stdout, nil)
	if err != nil {
		t.Fatal(err)
	}
	checkTestOutput(t, stdout.String(), fields, expect)
}

func testSampleAnalysisCommand(t *testing.T, command cmd.SampleAnalysisCommand, fields string, expect []string) {
	err := command.Validate()
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr strings.Builder
	err = LocalSampleOperation(cmd.NewMeta(command), command, nil, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	checkTestOutput(t, stdout.String(), fields, expect)
}

func checkTestOutput(t *testing.T, stdout, fields string, expect []string) {
	lines := strings.Split(stdout, "\n")
	if lines[0] != fields {
		t.Fatalf("Header: got %s wanted %s", lines[0], fields)
	}
	if len(lines) != len(expect)+1 {
		t.Fatalf("Length: got %d", len(lines))
	}
	for i, e := range expect {
		if lines[i+1] != e {
			t.Fatalf("Line %d:\ngot    %s\nexpect %s", i+1, lines[i+1], e)
		}
	}
}
