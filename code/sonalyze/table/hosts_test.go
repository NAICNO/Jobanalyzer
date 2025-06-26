package table

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

var (
	_ fmt.Formatter
)

func TestSetBasics(t *testing.T) {
	s := makeSet()
	s.addNames("a.b")
	if !s.lookup([]string{"a"}) {
		t.Fatal("lookup")
	}
	if !s.lookup([]string{"a", "b"}) {
		t.Fatal("lookup")
	}
	if s.lookup([]string{"b"}) {
		t.Fatal("lookup")
	}
	s.addNames("a", "a.b.c", "a.b.d")
	if !s.lookup([]string{"a"}) {
		t.Fatal("lookup")
	}
	if !s.lookup([]string{"a", "b"}) {
		t.Fatal("lookup")
	}
	if !s.lookup([]string{"a", "b", "c"}) {
		t.Fatal("lookup")
	}
	if !s.lookup([]string{"a", "b", "d"}) {
		t.Fatal("lookup")
	}
	if s.lookup([]string{"b"}) {
		t.Fatal("lookup")
	}
	if s.lookup([]string{"a", "c"}) {
		t.Fatal("lookup")
	}
	sinks := s.sinks()
	if len(sinks) != 2 {
		t.Fatal("sinks")
	}
	counts := make(map[string]int)
	for _, x := range sinks {
		counts[x.me]++
	}
	if counts["c"] != 1 {
		t.Fatal("count")
	}
	if counts["d"] != 1 {
		t.Fatal("count")
	}
}

func TestSetEquality(t *testing.T) {
	s := makeSet()
	s.addNames("a.b", "a", "a.b.c", "a.b.d")

	// s = s always
	if compare(s.sources, s.sources) != 0 {
		t.Fatal("equal")
	}

	// a is a prefix of a.b.c and a.b.d so s = r
	// s = {a.b.c, a.b.d}
	// r = {a}
	r := makeSet()
	r.addNames("a")
	if compare(s.sources, r.sources) != 0 {
		t.Fatal("Equal")
	}

	// a.b is a prefix of a.b.c and a.b.d so s = q
	// s = {a.b.c, a.b.d}
	// q = {a.b}
	q := makeSet()
	q.addNames("a.b")
	if compare(s.sources, q.sources) != 0 {
		t.Fatal("Equal")
	}

	// but not the reverse - s is not a covering prefix of q
	if compare(q.sources, s.sources) != 1 {
		t.Fatal("Unequal")
	}
}

func TestSetInequality(t *testing.T) {
	{
		s := makeSet()
		s.addNames("a.b.c", "a.b.d", "b.e")
		q := makeSet()
		q.addNames("a.b")

		// q < s
		// s = {a.b.c, a.b.d, b.e}
		// q = {a.b}
		s.addNames("b.e")
		if compare(s.sources, q.sources) >= 0 {
			t.Fatal("Less")
		}

		// Same, but reversed - easy
		if compare(q.sources, s.sources) != 1 {
			t.Fatal("Not less")
		}
	}

	{
		// Harder: we diverge deeper down
		s := makeSet()
		s.addNames("a.b.c", "a.b.d")
		q := makeSet()
		q.addNames("a.b.c")

		if compare(s.sources, q.sources) >= 0 {
			t.Fatal("Less")
		}
	}
}

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
	for _, x := range []string{"a", "a.b", "a.b.c", "x", "x.b", "x.b.c"} {
		if !h.HasElement(x) {
			t.Fatal(x)
		}
	}

	// Add something that was there, data should not change
	x := h.serial
	h.Add("a.b.c")
	n = h.FormatFull()
	if n != "a.b.c,x.b.c,y.c" {
		t.Fatal(n)
	}
	h.Add("a.b")
	n = h.FormatFull()
	if n != "a.b.c,x.b.c,y.c" {
		t.Fatal(n)
	}
	h.Add("a")
	n = h.FormatFull()
	if n != "a.b.c,x.b.c,y.c" {
		t.Fatal(n)
	}
	if h.serial != x {
		t.Fatal("Changed")
	}

	// Add a longer name, it replaces the existing entry
	x = h.serial
	h.Add("a.b.c.d")
	if h.serial == x {
		t.Fatal("Unchanged")
	}
	n = h.FormatFull()
	if n != "a.b.c.d,x.b.c,y.c" {
		t.Fatal(n)
	}
	if !h.HasElement("a.b.c.d") {
		t.Fatal("a.b.c.d")
	}

	// Add a new name, it creates a new entry
	x = h.serial
	h.Add("a.b.c.e")
	if h.serial == x {
		t.Fatal("Unchanged")
	}
	n = h.FormatFull()
	if n != "a.b.c.d,a.b.c.e,x.b.c,y.c" {
		t.Fatal(n)
	}
	n = h.FormatBrief()
	if n != "a,x,y" {
		t.Fatal(n)
	}
	if !h.HasElement("a.b.c.e") {
		t.Fatal("a.b.c.e")
	}

	{
		lhs := NewHostnames()
		lhs.Add("a.b.c")
		lhs.Add("a.b.d")
		rhs := NewHostnames()
		rhs.Add("a.b")
		if !lhs.Equal(rhs) {
			t.Fatal("Equal")
		}
		if !lhs.HasSubset(rhs, false) {
			t.Fatal("Equal")
		}
		if rhs.Equal(lhs) {
			t.Fatal("Unequal")
		}
		if rhs.HasSubset(lhs, false) {
			t.Fatal("Unequal")
		}
	}

	{
		lhs := NewHostnames()
		lhs.Add("a.b.c")
		lhs.Add("a.b.d")
		rhs := NewHostnames()
		rhs.Add("a.b.c")

		if lhs.Equal(rhs) {
			t.Fatal("Unequal")
		}
		if !lhs.HasSubset(rhs, true) {
			t.Fatal("Less")
		}
		if rhs.HasSubset(lhs, true) {
			t.Fatal("Less")
		}
	}
}

func TestEnumerate(t *testing.T) {
	h := NewHostnames()
	h.Add("a.b.c")
	h.Add("x.b.c")
	h.Add("y.c")
	names := slices.Collect(h.FullNames)
	slices.Sort(names)
	if strings.Join(names, ",") != "a.b.c,x.b.c,y.c" {
		t.Fatal(names)
	}
}

func TestAddCompressed(t *testing.T) {
	h := NewHostnames()
	h.AddCompressed("a[9-11].foo,bar.foo,baz.foo")
	names := slices.Collect(h.FullNames)
	slices.Sort(names)
	if strings.Join(names, ",") != "a10.foo,a11.foo,a9.foo,bar.foo,baz.foo" {
		t.Fatal(names)
	}
}
