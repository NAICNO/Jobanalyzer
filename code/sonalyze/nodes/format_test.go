// This is logically a unit test for the printing of node information, but in actuality an
// integration test that tests the `node` verb since it uses the entire machinery underneath that
// verb.  Architecturally, this test case is therefore a peer of the `sonalyze` top-level logic
// even if it's situated inside the `nodes` package.

package nodes

import (
	"strings"
	"testing"
)

var (
	logFiles = []string{"testdata/sysinfo1.json", "testdata/sysinfo2.json"}
	expect   = []string{
		`2024-11-01T00:00:01+01:00,ml1.hpc.uio.no,"2x14 (hyperthreaded) Intel(R) Xeon(R) Gold 5120 CPU @ 2.20GHz, 125 GiB, 3x NVIDIA GeForce RTX 2080 Ti @ 11GiB",56,125,3,33,no`,
		`2024-11-01T00:00:01+01:00,ml6.hpc.uio.no,"2x16 (hyperthreaded) AMD EPYC 7282 16-Core Processor, 251 GiB, 8x NVIDIA GeForce RTX 2080 Ti @ 11GiB",64,251,8,88,no`,
		``,
	}
)

func Test(t *testing.T) {
	testit(t, "timestamp,host,desc,cores,mem,gpus,gpumem,gpumempct")
	testit(t, "Timestamp,Hostname,Description,CpuCores,MemGB,GpuCards,GpuMemGB,GpuMemPct")
}

func testit(t *testing.T, fields string) {
	// Mock the command and run it
	var nc NodeCommand
	nc.SourceArgs.LogFiles = logFiles
	nc.FormatArgs.Fmt = "csv,header," + fields
	err := nc.Validate()
	if err != nil {
		t.Fatal(err)
	}
	var stdout strings.Builder
	err = nc.Nodes(nil, &stdout, nil)
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
