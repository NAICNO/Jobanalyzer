// This is logically a unit test for the printing of cluster information, but in actuality an
// integration test that tests the `cluster` verb since it uses the entire machinery underneath that
// verb.

package application

import (
	"testing"

	"sonalyze/cmd/clusters"
)

func TestClusters(t *testing.T) {
	testitCluster(t, "cluster,aliases,desc")
	testitCluster(t, "Name,Aliases,Description")
}

func testitCluster(t *testing.T, fields string) {
	var (
		jobanalyzerDir = "testdata/cluster_format_test"
		expect         = []string{
			`cluster1,c1,Cluster 1`,
			`cluster2,"c2,d2",Cluster 2`,
			``,
		}
	)
	var cc clusters.ClusterCommand
	cc.JobanalyzerDir = jobanalyzerDir
	cc.FormatArgs.Fmt = "csv,header," + fields
	testSimpleCommand(t, &cc, fields, expect)
}
