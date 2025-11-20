package application

import (
	"strings"
	"testing"

	"sonalyze/cmd"
	"sonalyze/db"
)

func testPrimitiveCommand(t *testing.T, command cmd.PrimitiveCommand, fields string, expect []string) {
	err := command.Validate()
	if err != nil {
		t.Fatal(err)
	}
	err = cmd.OpenDataStoreFromCommand(command)
	if err != nil {
		t.Fatal(err)
	}
	defer db.CloseDataStore()
	var stdout strings.Builder
	err = command.Perform(nil, &stdout, nil)
	if err != nil {
		t.Fatal(err)
	}
	checkTestOutput(t, "primitive", stdout.String(), fields, expect)
}

func testSimpleCommand(t *testing.T, tag string, command cmd.SimpleCommand, fields string, expect []string) {
	err := command.Validate()
	if err != nil {
		t.Fatal(err)
	}
	err = cmd.OpenDataStoreFromCommand(command)
	if err != nil {
		t.Fatal(err)
	}
	defer db.CloseDataStore()
	var stdout strings.Builder
	err = command.Perform(cmd.NewContextFromCommand(command), nil, &stdout, nil)
	if err != nil {
		t.Fatal(err)
	}
	checkTestOutput(t, tag+"/simple", stdout.String(), fields, expect)
}

func testSampleAnalysisCommand(t *testing.T, command cmd.SampleAnalysisCommand, fields string, expect []string) {
	err := command.Validate()
	if err != nil {
		t.Fatal(err)
	}
	err = cmd.OpenDataStoreFromCommand(command)
	if err != nil {
		t.Fatal(err)
	}
	defer db.CloseDataStore()
	var stdout, stderr strings.Builder
	err = LocalSampleOperation(cmd.NewContextFromCommand(command), command, nil, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	checkTestOutput(t, "sampleAnalysis", stdout.String(), fields, expect)
}

func checkTestOutput(t *testing.T, tag, stdout, fields string, expect []string) {
	lines := strings.Split(stdout, "\n")
	if lines[0] != fields {
		t.Fatalf("%s Header: got %s wanted %s", tag, lines[0], fields)
	}
	if len(lines) != len(expect)+1 {
		t.Fatalf("%s Length: got %d wanted %d", tag, len(lines), len(expect)+1)
	}
	for i, e := range expect {
		if lines[i+1] != e {
			t.Fatalf("%s Line %d:\ngot    %s\nexpect %s", tag, i+1, lines[i+1], e)
		}
	}
}
