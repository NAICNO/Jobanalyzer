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
}
