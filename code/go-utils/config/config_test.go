package config

import (
	"testing"
)

func TestReadConfig(t *testing.T) {
	cfg, err := ReadConfig("test-config.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg) != 5 {
		t.Fatalf("Expected 3 elements")
	}
	c0 := cfg["ml7.hpc.uio.no"]
	if c0.CpuCores != 64 || c0.MemGB != 256 || c0.GpuCards != 8 || c0.GpuMemGB != 88 || c0.GpuMemPct != false {
		t.Fatalf("element 0: %v", c0)
	}
	c1 := cfg["ml8.hpc.uio.no"]
	if c1.CpuCores != 192 || c1.MemGB != 1024 || c1.GpuCards != 4 || c1.GpuMemGB != 0 || c1.GpuMemPct != true {
		t.Fatalf("element 1: %v", c1)
	}
	names := []string{"c1-10", "c1-11", "c1-12"}
	for i := 2; i < 5; i++ {
		c := cfg[names[i-2]]
		if c.CpuCores != 128 || c.MemGB != 512 || c.GpuCards != 0 || c.GpuMemGB != 0 || c.GpuMemPct != false {
			t.Fatalf("content element 2+%d: %v", i, c)
		}
		if c.Hostname != names[i-2] {
			t.Fatalf("name element 2+%d: %v", i, c)
		}
	}
}
