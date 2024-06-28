package slices

import (
	"reflect"
	"testing"
)

func TestMap(t *testing.T) {
	s := []int{1, 2, 3, 4, 5}
	q := Map(s, func(x int) float64 { return -float64(x) / 2 })
	r := []float64{-0.5, -1, -1.5, -2, -2.5}
	if !reflect.DeepEqual(q, r) {
		t.Fatal("Bad ", q)
	}
}

func TestCatenate(t *testing.T) {
	r := Catenate([][]int{[]int{1, 2, 3}, []int{4, 5, 6}, []int{7, 8, 9}})
	if !reflect.DeepEqual(r, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}) {
		t.Fatal("Bad")
	}
	if cap(r) != 9 {
		t.Fatal("Too much cap")
	}
}

func TestCatenateP(t *testing.T) {
	r := CatenateP([]*[]int{&[]int{1, 2, 3}, &[]int{4, 5, 6}, &[]int{7, 8, 9}})
	if !reflect.DeepEqual(r, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}) {
		t.Fatal("Bad")
	}
	if cap(r) != 9 {
		t.Fatal("Too much cap")
	}
}
