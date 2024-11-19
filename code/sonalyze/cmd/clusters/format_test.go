// This is logically a unit test for the printing of cluster information, but in actuality an
// integration test that tests the `cluster` verb since it uses the entire machinery underneath that
// verb.  Architecturally, this test case is therefore a peer of the `sonalyze` top-level logic.
// even if it's situated inside the `clusters` package.

package clusters

import (
	"strings"
	"testing"
)

var (
	jobanalyzerDir = "testdata"
	expect         = []string{
		`cluster1,c1,Cluster 1`,
		`cluster2,"c2,d2",Cluster 2`,
		``,
	}
)

func Test(t *testing.T) {
	testit(t, "cluster,aliases,desc")
	testit(t, "Name,Aliases,Description")
}

func testit(t *testing.T, fields string) {
	// Mock the command and run it
	var cc ClusterCommand
	cc.jobanalyzerDir = jobanalyzerDir
	cc.FormatArgs.Fmt = "csv,header," + fields
	err := cc.Validate()
	if err != nil {
		t.Fatal(err)
	}
	var stdout strings.Builder
	err = cc.Clusters(nil, &stdout, nil)
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
