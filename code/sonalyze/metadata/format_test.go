// This is logically a unit test for the printing of metadata information, but in actuality an
// integration test that tests the `metadata` verb since it uses the entire machinery underneath that
// verb.  Architecturally, this test case is therefore a peer of the `sonalyze` top-level logic
// even if it's situated inside the `metadata` package.

package metadata

import (
	"strings"
	"testing"

	"sonalyze/application"
)

var (
	expect = []string{
		`ml6.hpc.uio.no,2024-10-31 00:00,2024-10-31 00:30`,
		``,
	}
)

func Test(t *testing.T) {
	testit(t, "host,earliest,latest")
	testit(t, "Hostname,Earliest,Latest")
}

func testit(t *testing.T, fields string) {
	var mc MetadataCommand
	mc.SourceArgs.LogFiles = []string{"testdata/test.csv"}
	mc.FormatArgs.Fmt = "csv,header," + fields
	mc.Bounds = true
	err := mc.Validate()
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr strings.Builder
	err = application.LocalOperation(&mc, nil, &stdout, &stderr)
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
