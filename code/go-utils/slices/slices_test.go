package slices

import (
	"reflect"
	"testing"
)

func TestMap(t *testing.T) {
	s := []int{1,2,3,4,5}
	q := Map(s, func(x int) float64 { return -float64(x)/2 })
	r := []float64{-0.5, -1, -1.5, -2, -2.5}
	if !reflect.DeepEqual(q, r) {
		t.Fatal("Bad ", q)
	}
}

func TestCopy(t *testing.T) {
	s := []int{1,2,3,4,5}
	q := Copy(s)
	if !reflect.DeepEqual(s, q) {
		t.Fatal("Bad ", q)
	}
}

// Insert moves elements to the side
func TestInsertInplace(t *testing.T) {
	s := make([]int, 5, 10)
	copy(s, []int{1,2,3,4,5})
	q := Insert(s, 3, 6, 7, 8)
	if !reflect.DeepEqual(q, []int{1,2,3,6,7,8,4,5}) {
		t.Fatal("Bad ", q)
	}
	if !reflect.DeepEqual(s, []int{1,2,3,6,7}) {
		t.Fatal("Bad ", s)
	}
}

// Insert creates a new array
func TestInsertNew(t *testing.T) {
	s := make([]int, 5, 7)
	copy(s, []int{1,2,3,4,5})
	q := Insert(s, 3, 6, 7, 8)
	if !reflect.DeepEqual(q, []int{1,2,3,6,7,8,4,5}) {
		t.Fatal("Bad ", q)
	}
	if !reflect.DeepEqual(s, []int{1,2,3,4,5}) {
		t.Fatal("Bad ", s)
	}
}

func TestBinarySearchFunc(t *testing.T) {
	n, found := BinarySearchFunc([]int{}, 3, func(x, y int)int { return 0 })
	if n != 0 {
		t.Fatal("Bad", n)
	}
	if found {
		t.Fatal("Bad", found)
	}
	xs := []int{1,3,5,7,9,11}
	for i := 1 ; i < len(xs) ; i++ {
		s := xs[:i]
		for k := s[0]-1; k <= s[len(s)-1]+1 ; k++ {
			n, found := BinarySearchFunc(s, k, func(x, y int) int {
				if x == y {
					return 0
				}
				return x - y
			})
			if found {
				if s[n] != k {
					t.Fatal("Bad", n)
				}
			} else {
				if n < len(s) && s[n] == k {
					t.Fatal("Bad", n)
				}
				if n > 0 && s[n-1] >= k {
					t.Fatal("Bad", n)
				}
				if n < len(s)-1 && s[n+1] <= k {
					t.Fatal("Bad", n)
				}
			}
		}
	}
}
