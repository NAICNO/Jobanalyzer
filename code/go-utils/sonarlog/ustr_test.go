package sonarlog

import (
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
