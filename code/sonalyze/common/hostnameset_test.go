package common

import (
	"fmt"
	"reflect"
	"testing"
)

func TestNodeset(t *testing.T) {
	var n nodeset
	if !n.empty() {
		t.Fatal("empty")
	}
	if n.size() != 0 {
		t.Fatal(n.size())
	}
	if n.member(7) {
		t.Fatal("Member")
	}
	for v := range n.iter() {
		t.Fatal(v)
	}
	n.ranges = []nrange{nrange{1, 1}}
	if s := n.singletonName(); s != "1" {
		t.Fatal(s)
	}
	n.ranges = []nrange{nrange{1, 1}, nrange{3, 5}}
	if n.empty() {
		t.Fatal("not empty")
	}
	if n.size() != 4 {
		t.Fatal(n.size())
	}
	if s := n.compressedName(); s != "[1,3-5]" {
		t.Fatal(s)
	}
	if n.member(7) {
		t.Fatal("Member")
	}
	if !n.member(5) {
		t.Fatal("Member")
	}
	if !n.member(1) {
		t.Fatal("Member")
	}
	m := map[int]bool{1: true, 3: true, 4: true, 5: true}
	for v := range n.iter() {
		if !m[v] {
			t.Fatal(v)
		}
		delete(m, v)
	}
	if len(m) != 0 {
		t.Fatal(m)
	}
	var o nodeset
	o.insertRange(nrange{1, 5})
	o.insertRange(nrange{5, 8})
	o.insertRange(nrange{10, 10})
	if !reflect.DeepEqual(o.ranges, []nrange{nrange{1, 8}, nrange{10, 10}}) {
		t.Fatal(o)
	}
	var p nodeset
	for _, r := range []nrange{nrange{2, 3}, nrange{4, 5}, nrange{2, 4}, nrange{9, 10}, nrange{6, 7}, nrange{1, 3}} {
		p.insertRange(r)
	}
	if !reflect.DeepEqual(p.ranges, []nrange{nrange{1, 7}, nrange{9, 10}}) {
		t.Fatal(p)
	}
}

func TestHostnameSetBasic(t *testing.T) {
	hs, err := parseConcretePattern("a[9-12,7]b.c")
	if err != nil {
		t.Fatal(err)
	}
	if hs.IsSingle() {
		t.Fatal("Single")
	}
	m := map[string]bool{
		"a9b.c":  true,
		"a10b.c": true,
		"a11b.c": true,
		"a12b.c": true,
		"a7b.c":  true,
	}
	for n := range hs.Expand() {
		if !m[n] {
			t.Fatal("Not found: " + n)
		}
		delete(m, n)
	}
	if len(m) != 0 {
		t.Fatal(m)
	}
	s := hs.String()
	if s != "a[7,9-12]b.c" {
		t.Fatal(s)
	}
	h, err := parseHostname("a11b.c")
	if err != nil {
		t.Fatal(err)
	}
	s = h.String()
	if s != "a11b.c" {
		t.Fatal(s)
	}
	if !hs.PrefixMatch(h) {
		t.Fatal("Should have matched")
	}
	h2, _ := parseHostname("a1b.c")
	if hs.PrefixMatch(h2) {
		t.Fatal("Should not have matched")
	}
	hs2, _ := parseConcretePattern("a[3-5]b.c")
	hs3, _ := parseConcretePattern("a[14].x")
	hs4, _ := parseConcretePattern("a17.x")
	xs := UnionHostnameSets([]HostnameSet{hs, hs3, hs2, hs4})
	fmt.Println(xs)
	m = map[string]bool{
		"a[3-5,7,9-12]b.c": true,
		"a[14,17].x":       true,
	}
	for _, x := range xs {
		n := x.String()
		if !m[n] {
			t.Fatal(n)
		}
		delete(m, n)
	}
	if len(m) != 0 {
		t.Fatal(m)
	}
}
