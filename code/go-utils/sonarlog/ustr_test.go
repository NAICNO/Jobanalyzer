package sonarlog

import (
	"fmt"
	"testing"
	"unsafe"
)

func TestUstr(t *testing.T) {
	var v Ustr
	// Comparison is possible
	_ = v == v
	_ = v != v
	if unsafe.Sizeof(v) != 4 {
		t.Fatal("Size must be 4")
	}
	// Make sure there's stuff in the pool before we ask for empty string
	_ = StringToUstr("x")
	_ = StringToUstr("y")
	_ = StringToUstr("z")
	v = StringToUstr("")
	if int(v) != 0 {
		t.Fatal("Empty string must be zero")
	}
	if v != UstrEmpty {
		t.Fatal("UstrEmpty must be zero")
	}
	if StringToUstr("hi there") != StringToUstr("hi there") {
		t.Fatal("Equality")
	}
	if StringToUstr("hi") == StringToUstr("ho") {
		t.Fatal("Disequality")
	}
	if StringToUstr("hi").String() != "hi" {
		t.Fatal("Roundtripping")
	}
	if Ustr(0).String() != "" {
		t.Fatal("Roundtripping empty")
	}
}

func TestUstrAllocator(t *testing.T) {
	c := NewUstrCache()
	v := c.Alloc("hi")
	w := c.Alloc("hi")
	if v != w {
		t.Fatal("Alloc cache")
	}
	x := StringToUstr("hi")
	if x != v {
		t.Fatal("Transparent cache")
	}
	zappa(t, v, c)
}

func TestUstrFacade(t *testing.T) {
	c := NewUstrFacade()
	v := c.Alloc("hi")
	w := c.Alloc("hi")
	if v != w {
		t.Fatal("Alloc facade")
	}
	x := StringToUstr("hi")
	if x != v {
		t.Fatal("Transparent facade")
	}
	zappa(t, v, c)
}

func zappa(t *testing.T, hi Ustr, a UstrAllocator) {
	if a.Alloc("hi") != hi {
		t.Fatal("Interface")
	}
}

func TestUstrJoin(t *testing.T) {
	c := NewUstrFacade()
	v := c.Alloc("hi")
	w := c.Alloc("ho")
	x := c.Alloc("hum")
	j := c.Alloc("-")
	z := UstrJoin([]Ustr{v, w, x}, j)
	if z.String() != "hi-ho-hum" {
		t.Fatalf("Join: %s", z.String())
	}
}

func TestUstrSort(t *testing.T) {
	c := NewUstrFacade()
	v := c.Alloc("hi")
	w := c.Alloc("ho")
	x := c.Alloc("hum")
	xs := []Ustr{w, x, v}
	j := c.Alloc("-")
	z := UstrJoin(xs, j)
	if z.String() != "ho-hum-hi" {
		t.Fatalf("Join: %s", z.String())
	}
	UstrSortAscending(xs)
	z = UstrJoin(xs, j)
	if z.String() != "hi-ho-hum" {
		t.Fatalf("Join: %s", z.String())
	}
}

func TestBytesToUstr(t *testing.T) {
	a := StringToUstr("hello world")
	b := BytesToUstr([]byte("hello world"))
	if a != b {
		t.Fatal("BytesToUstr")
	}
}

func TestAllocBytesFacade(t *testing.T) {
	a := StringToUstr("hello world")
	c := NewUstrFacade()
	b := c.AllocBytes([]byte("hello world"))
	if a != b {
		t.Fatal("AllocBytes (facade)")
	}
}

func TestAllocBytesCache(t *testing.T) {
	a := StringToUstr("hello world")
	c := NewUstrCache()
	b := c.AllocBytes([]byte("hello world"))
	if a != b {
		t.Fatal("AllocBytes (cache)")
	}
}

func TestHashFunctions(t *testing.T) {
	// h and j will tend to diverge if hashString operates on runes and hashBytes operates on bytes.
	h := hashString("abcæøå")
	j := hashBytes([]byte("abcæøå"))
	if h != j {
		t.Fatal("Hash function is inconsistent")
	}
}

func TestHashtable(t *testing.T) {
	ht := newHashtable()
	s := "supercalifragilistic"
	hcs := hashString(s)
	ht.insert(hcs, s, Ustr(37))

	bs := []byte(s)
	hcb := hashBytes(bs)
	hnb := ht.getBytes(hcb, bs)

	// getString and getBytes should return the same node for same input
	if hnb == nil {
		t.Fatal("Nil node")
	}

	hns := ht.getString(hcs, s)
	if hns == nil {
		t.Fatal("Nil node")
	}

	if hnb != hns {
		t.Fatal("Nodes unequal")
	}

	// getString should return the same node before and after rehashing
	k := ht.rehashes
	for i := 1000; ht.rehashes == k; i++ {
		s := fmt.Sprintf("x%d", i)
		hcs := hashString(s)
		ht.insert(hcs, s, Ustr(i))
	}
	hns2 := ht.getString(hcs, s)
	if hns != hns2 {
		t.Fatal("Nodes unequal")
	}
}
