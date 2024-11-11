// This is logically a unit test for the printing of config information, but in actuality an
// integration test that tests the `config` verb since it uses the entire machinery underneath that
// verb.  Architecturally, this test case is therefore a peer of the `sonalyze` top-level logic
// even if it's situated inside the `configs` package.

package configs

import (
	"strings"
	"testing"
)

var (
	configFilename = "testdata/test-config.json"
	expect         = []string{
		`,n1,"2x14 Intel Xeon Gold 5120 (hyperthreaded), 128GB, 4x NVIDIA RTX 2080 Ti @ 11GB",no,56,128,4,44,no`,
		`2024-10-31T12:00:00Z,n2,"1x14 Intel Xeon Gold 5120 (hyperthreaded), 256GB, 2x NVIDIA RTX 2080 Ti @ 11GB",yes,28,256,2,22,no`,
		``,
	}
)

func Test(t *testing.T) {
	testit(t, "timestamp,host,desc,xnode,cores,mem,gpus,gpumem,gpumempct")
	testit(t, "Timestamp,Hostname,Description,CrossNodeJobs,CpuCores,MemGB,GpuCards,GpuMemGB,GpuMemPct")
}

func testit(t *testing.T, fields string) {
	// Mock the command and run it
	var cc ConfigCommand
	cc.ConfigFileArgs.ConfigFilename = configFilename
	cc.FormatArgs.Fmt = "csv,header," + fields
	err := cc.Validate()
	if err != nil {
		t.Fatal(err)
	}
	var stdout strings.Builder
	err = cc.Configs(nil, &stdout, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Check the output.
	lines := strings.Split(stdout.String(), "\n")
	if lines[0] != fields {
		t.Fatalf("Header: got %s wanted %s", lines[0], fields)
	}
	if len(lines) != len(expect)+1 {
		t.Fatalf("Length: got %d", len(lines))
	}
	for i, e := range expect {
		if lines[i+1] != e {
			t.Fatalf("Line %d: got %s", i+1, lines[i+1])
		}
	}
}
