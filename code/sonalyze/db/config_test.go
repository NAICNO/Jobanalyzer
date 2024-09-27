// Simple test cases for the "config" cluster

package db

import (
	"cmp"
	"slices"
	"testing"

	"go-utils/config"
)

func TestConfig(t *testing.T) {
	cfg, err := ReadConfigData(MakeConfigFilePath("testdata", "cluster1.uio.no"))
	if err != nil {
		t.Fatal(err)
	}
	hosts := cfg.Hosts()
	if len(hosts) != 2 {
		t.Fatal("Hosts ", hosts)
	}
	slices.SortFunc(hosts, func(a, b *config.NodeConfigRecord) int {
		return cmp.Compare(a.Hostname, b.Hostname)
	})
	if hosts[0].Hostname != "n1.cluster1.uio.no" {
		t.Fatal("Name ", hosts[0])
	}
}