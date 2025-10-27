// Simple test cases for the "table" cluster

package special

import (
	"testing"
)

func TestCluster(t *testing.T) {
	err := OpenFullDataStore("../filedb/testdata")
	if err != nil {
		t.Fatal(err)
	}
	if len(AllClusters()) != 2 {
		t.Fatal("Cluster length", AllClusters())
	}

	c1 := LookupCluster("cluster1.uio.no")
	if c1 == nil {
		t.Fatal("cluster1")
	}
	if c1.Name != "cluster1.uio.no" {
		t.Fatal("cluster1 name", c1)
	}
	if c1.Description != "Cluster 1" {
		t.Fatal("cluster1 desc", c1)
	}

	c2 := LookupCluster("cluster2.uio.no")
	if c2 == nil {
		t.Fatal("cluster2")
	}
	if c2.Name != "cluster2.uio.no" {
		t.Fatal("cluster2 name", c2)
	}
	if c2.Description != "Cluster 2" {
		t.Fatal("cluster2 desc", c2)
	}

	r1 := ResolveClusterName("c1")
	if r1 != "cluster1.uio.no" {
		t.Fatal("c1", r1)
	}

	r2 := ResolveClusterName("c2")
	if r2 != "cluster2.uio.no" {
		t.Fatal("c2", r2)
	}
}
