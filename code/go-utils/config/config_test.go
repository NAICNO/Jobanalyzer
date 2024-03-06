package config

import (
	"bytes"
	"reflect"
	"testing"
)

func TestReadConfigV1(t *testing.T) {
	cfg, err := ReadConfig("test-config.json")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Version != 1 {
		t.Fatalf("Expected Version=1, got %d", cfg.Version)
	}
	c0 := cfg.LookupHost("ml7.hpc.uio.no")
	if c0.CpuCores != 64 || c0.MemGB != 256 || c0.GpuCards != 8 || c0.GpuMemGB != 88 || c0.GpuMemPct != false {
		t.Fatalf("element 0: %v", c0)
	}
	c1 := cfg.LookupHost("ml8.hpc.uio.no")
	if c1.CpuCores != 192 || c1.MemGB != 1024 || c1.GpuCards != 4 || c1.GpuMemGB != 0 || c1.GpuMemPct != true {
		t.Fatalf("element 1: %v", c1)
	}
	names := []string{"c1-10", "c1-11", "c1-12"}
	for i := 2; i < 5; i++ {
		c := cfg.LookupHost(names[i-2])
		if c.CpuCores != 128 || c.MemGB != 512 || c.GpuCards != 0 || c.GpuMemGB != 0 || c.GpuMemPct != false {
			t.Fatalf("content element 2+%d: %v", i, c)
		}
		if c.Hostname != names[i-2] {
			t.Fatalf("name element 2+%d: %v", i, c)
		}
	}

	testRoundtrip(t, cfg)
}

func testRoundtrip(t *testing.T, cfg *ClusterConfig) {
	var buf bytes.Buffer
	err := WriteConfigTo(&buf, cfg)
	if err != nil {
		t.Fatalf("Could not write: %v", err)
	}
	newCfg, err := ReadConfigFrom(&buf)
	if err != nil {
		t.Fatalf("Could not read: %v", err)
	}
	// This depends on implementation details: the "nodes" are represented as a map, and maps are
	// equal if they have the same keys with the same values.  If the nodes were implemented as an
	// array we'd additionally depend on ordering.
	if !reflect.DeepEqual(cfg, newCfg) {
		t.Fatalf("Failed roundtripping")
	}
}

func TestReadConfigV2(t *testing.T) {
	cfg, err := ReadConfig("test-config-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Version != 2 {
		t.Fatalf("Expected Version=2, got %d", cfg.Version)
	}
	if cfg.Name != "mlx.hpc.uio.no" {
		t.Fatalf("Name: %s", cfg.Name)
	}
	if cfg.Description != "UiO machine learning nodes" {
		t.Fatalf("Description: %s", cfg.Description)
	}
	if len(cfg.Aliases) != 2 || cfg.Aliases[0] != "ml" || cfg.Aliases[1] != "mlx" {
		t.Fatalf("Aliases %v", cfg.Aliases)
	}
	if len(cfg.ExcludeUser) != 2 || cfg.ExcludeUser[0] != "root" || cfg.ExcludeUser[1] != "toor" {
		t.Fatalf("ExcludeUser %v", cfg.ExcludeUser)
	}
	c0 := cfg.LookupHost("ml7.hpc.uio.no")
	if c0.CpuCores != 64 || c0.MemGB != 256 || c0.GpuCards != 8 || c0.GpuMemGB != 88 || c0.GpuMemPct != false {
		t.Fatalf("element 0: %v", c0)
	}
	c1 := cfg.LookupHost("ml8.hpc.uio.no")
	if c1.CpuCores != 192 || c1.MemGB != 1024 || c1.GpuCards != 4 || c1.GpuMemGB != 0 || c1.GpuMemPct != true {
		t.Fatalf("element 1: %v", c1)
	}
	names := []string{"c1-10", "c1-11", "c1-12"}
	for i := 2; i < 5; i++ {
		c := cfg.LookupHost(names[i-2])
		if c.CpuCores != 128 || c.MemGB != 512 || c.GpuCards != 0 || c.GpuMemGB != 0 || c.GpuMemPct != false {
			t.Fatalf("content element 2+%d: %v", i, c)
		}
		if c.Hostname != names[i-2] {
			t.Fatalf("name element 2+%d: %v", i, c)
		}
	}
	testRoundtrip(t, cfg)
}
