package sysinfo

import (
	"testing"
)

func TestUserMap(t *testing.T) {
	um := NewUserMap()
	r := um.LookupUid(0)
	if r != "root" {
		t.Fatalf("Bad name for uid 0: %s", r)
	}
	r = um.LookupUid(123456)
	if r != "_noinfo_" {
		t.Fatalf("Bad name for uid 123456: %s", r)
	}
}

func TestEnumeratePids(t *testing.T) {
	pids, err := EnumeratePids()
	if err != nil {
		t.Fatalf("EnumeratePids failed: %v", err)
	}
	if len(pids) == 0 {
		t.Fatalf("No pids")
	}
	xs := make(map[uint]bool)
	for _, p := range pids {
		if _, found := xs[p.Pid]; found {
			t.Fatalf("Duplicate pid %v", p.Pid)
		}
		xs[p.Pid] = true
	}
}

func TestCpuInfo(t *testing.T) {
	_, _, _, _, err := CpuInfo()
	if err != nil {
		t.Fatalf("CpuInfo failed: %v", err)
	}
}

func TestPhysMem(t *testing.T) {
	sz, err := PhysicalMemoryBy()
	if err != nil {
		t.Fatalf("PhysicalMemoryBy failed: %v", err)
	}
	if sz == 0 {
		t.Fatalf("Physical memory is zero")
	}
}

func TestPageSize(t *testing.T) {
	const KiB = 1024
	sz := PagesizeBy()
	// 4KB normal on Linux on most architectures, 64KB normal on ARM64
	if sz != 4*KiB && sz != 64*KiB {
		t.Fatalf("Physical memory: %d", sz)
	}
}

func TestBootTime(t *testing.T) {
	time, err := BootTime()
	if err != nil {
		t.Fatalf("BootTime failed: %v", err)
	}
	if time == 0 {
		t.Fatalf("BootTime is zero")
	}
	// TODO: Check that it's not in the future
	// TODO: Check that is's not in the far past
}
