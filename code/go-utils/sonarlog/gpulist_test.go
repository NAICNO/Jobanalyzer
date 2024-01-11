package sonarlog

import (
	"testing"
)

func TestGpuset(t *testing.T) {
	s, err := ParseGpulist("unknown")
	if err != nil {
		t.Fatal(err)
	}
	if s != nil {
		t.Fatalf("Unknown set")
	}
	s, err = ParseGpulist("none")
	if err != nil {
		t.Fatal(err)
	}
	if s == nil || len(s) != 0 {
		t.Fatalf("Empty set")
	}
	// Duplicate elements are not supported, really
	s, err = ParseGpulist("1,3,2,5,0")
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 5 {
		t.Fatalf("Length-5 set")
	}
	// Order of elements is unspecified
	xs := []uint32{0, 1, 2, 3, 5}
	for _, x := range xs {
		found := false
		for _, y := range s {
			if x == y {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Not found: %d", x)
		}
	}
}
