package table

import (
	"fmt"
	"testing"
)

var (
	_ fmt.Formatter
)

func TestHostnames(t *testing.T) {
	// Basic test
	h := NewHostnames()
	h.Add("a.b.c")
	h.Add("x.b.c")
	h.Add("y.c")
	n := h.FormatBrief()
	if n != "a,x,y" {
		t.Fatal(n)
	}
	n = h.FormatFull()
	if n != "a.b.c,x.b.c,y.c" {
		t.Fatal(n)
	}
}
