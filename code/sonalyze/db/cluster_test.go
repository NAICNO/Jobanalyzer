// Simple test cases for the "table" cluster

package db

import (
	"testing"
)

func TestCluster(t *testing.T) {
	clusters, aliases, err := ReadClusterData("testdata")
	if err != nil {
		t.Fatal(err)
	}
	if len(clusters) != 2 {
		t.Fatal("Cluster length", clusters)
	}

	c1, found := clusters["cluster1.uio.no"]
	if !found {
		t.Fatal("cluster1")
	}
	if c1.Name != "cluster1.uio.no" {
		t.Fatal("cluster1 name", c1)
	}
	if c1.Description != "Cluster 1" {
		t.Fatal("cluster1 desc", c1)
	}

	c2, found := clusters["cluster2.uio.no"]
	if !found {
		t.Fatal("cluster2")
	}
	if c2.Name != "cluster2.uio.no" {
		t.Fatal("cluster2 name", c2)
	}
	if c2.Description != "Cluster 2" {
		t.Fatal("cluster2 desc", c2)
	}

	r1 := aliases.Resolve("c1")
	if r1 != "cluster1.uio.no" {
		t.Fatal("c1", r1)
	}

	r2 := aliases.Resolve("c2")
	if r2 != "cluster2.uio.no" {
		t.Fatal("c2", r2)
	}
}
