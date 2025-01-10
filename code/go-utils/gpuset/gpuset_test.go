package gpuset

import (
	"reflect"
	"testing"
)

func TestGpuset(t *testing.T) {
	s, err := NewGpuSet("unknown")
	if err != nil {
		t.Fatal(err)
	}
	if !s.IsUnknown() {
		t.Fatalf("Unknown set")
	}
	s, err = NewGpuSet("none")
	if err != nil {
		t.Fatal(err)
	}
	if !s.IsEmpty() {
		t.Fatalf("Empty set")
	}
	s, err = NewGpuSet("1,3,2,5,0")
	if err != nil {
		t.Fatal(err)
	}
	if s.Size() != 5 {
		t.Fatalf("Length-5 set")
	}
	if !reflect.DeepEqual(s.AsSlice(), []int{0, 1, 2, 3, 5}) {
		t.Fatalf("Set values")
	}
	v, err := NewGpuSet("1,2,5")
	if err != nil {
		t.Fatal(err)
	}
	u, err := NewGpuSet("0,1,2,3,5")
	if err != nil {
		t.Fatal(err)
	}
	x, _ := NewGpuSet("unknown")

	// Sets equal themselves, even when unknown
	if !s.Equal(s) {
		t.Fatal("Equal")
	}
	if !x.Equal(x) {
		t.Fatal("Equal")
	}

	// Sets equal other sets that are equal
	if !s.Equal(u) {
		t.Fatal("Equal")
	}
	if !u.Equal(s) {
		t.Fatal("Equal")
	}

	// Sets do not equal other sets that are not equal
	if s.Equal(v) {
		t.Fatal("Equal")
	}
	if v.Equal(s) {
		t.Fatal("Equal")
	}

	// Equal sets are improper subsets
	if !s.HasSubset(u, false) {
		t.Fatal("Improper subset")
	}
	if !u.HasSubset(s, false) {
		t.Fatal("Improper subset")
	}

	// Equal sets are not proper subsets
	if s.HasSubset(u, true) {
		t.Fatal("Proper subset")
	}

	// Actual subsets are proper and improper subsets
	if !s.HasSubset(v, true) {
		t.Fatal("Proper subset")
	}
	if !s.HasSubset(v, false) {
		t.Fatal("Improper subset")
	}
	// This is not reflexive
	if v.HasSubset(s, true) {
		t.Fatal("Proper subset")
	}
	if v.HasSubset(s, false) {
		t.Fatal("Improper subset")
	}

	// Unknown sets
	if s.HasSubset(x, true) {
		t.Fatal("Unknown")
	}
	if s.HasSubset(x, false) {
		t.Fatal("Unknown")
	}
	if x.HasSubset(s, true) {
		t.Fatal("Unknown")
	}
	if x.HasSubset(s, false) {
		t.Fatal("Unknown")
	}

	// Unequal, overlapping - neither a subset of the other
	v, _ = NewGpuSet("0,3,5,7")
	if v.HasSubset(u, false) {
		t.Fatal("Overlapping")
	}
	if u.HasSubset(v, false) {
		t.Fatal("Overlapping")
	}
}
