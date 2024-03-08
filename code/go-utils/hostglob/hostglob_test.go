package hostglob

import (
	"testing"
)

func TestSplitMultiPattern(t *testing.T) {
	xs, err := SplitMultiPattern("yes.no,ml[1-3].hi,ml[1,2],zappa")
	if err != nil {
		t.Fatalf("Hostnames #1: %s", err.Error())
	}
	if len(xs) != 4 || xs[0] != "yes.no" || xs[1] != "ml[1-3].hi" || xs[2] != "ml[1,2]" || xs[3] != "zappa" {
		t.Fatalf("Hostnames #2: %v", xs)
	}
	// Empty input is allowed
	xs, err = SplitMultiPattern("")
	if err != nil {
		t.Fatalf("Hostnames #3: %s", err.Error())
	}
	if len(xs) != 0 {
		t.Fatalf("Hostnames #4: %v", xs)
	}
	// No closing bracket
	xs, err = SplitMultiPattern("yes[hi")
	if err == nil {
		t.Fatalf("Should fail #1: %v", xs)
	}
	// Nested opening bracket
	xs, err = SplitMultiPattern("yes[hi[]")
	if err == nil {
		t.Fatalf("Should fail #2: %v", xs)
	}
	// No opening bracket
	xs, err = SplitMultiPattern("yes]")
	if err == nil {
		t.Fatalf("Should fail #3: %v", xs)
	}
	// Empty at beginning
	xs, err = SplitMultiPattern(",yes")
	if err == nil {
		t.Fatalf("Should fail #4: %v", xs)
	}
	// Empty at end
	xs, err = SplitMultiPattern("yes,")
	if err == nil {
		t.Fatalf("Should fail #5: %v", xs)
	}
	// Empty in the middle
	xs, err = SplitMultiPattern("yes,,no")
	if err == nil {
		t.Fatalf("Should fail #6: %v", xs)
	}
}

func TestExpandPattern(t *testing.T) {
	x, err := ExpandPattern("ab[1-2,4].cd[3]")
	if err != nil {
		t.Fatal(err)
	}
	if len(x) != 3 || x[0] != "ab1.cd3" || x[1] != "ab2.cd3" || x[2] != "ab4.cd3" {
		t.Fatalf("Pattern: %v", x)
	}
	x, err = ExpandPattern("ab[1-2].cd[3].ef[1-2]")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"ab1.cd3.ef1": true,
		"ab2.cd3.ef1": true,
		"ab1.cd3.ef2": true,
		"ab2.cd3.ef2": true,
	}
	if len(x) != 4 || !want[x[0]] || !want[x[1]] || !want[x[2]] || !want[x[3]] {
		t.Fatalf("Pattern: %v", x)
	}
	x, err = ExpandPattern("ab[].cd")
	if err == nil {
		t.Fatal("Expected failure")
	}
	x, err = ExpandPattern("ab*.cd")
	if err == nil {
		t.Fatal("Expected failure")
	}
	x, err = ExpandPattern("ab[1-2]cd")
	if len(x) != 2 || x[0] != "ab1cd" || x[1] != "ab2cd" {
		t.Fatal("Embedded range")
	}
}

func TestCompressHostnames(t *testing.T) {
	testCompress(
		t,
		[]string{
			"c6-1",
			"c6-2",
			"c6-3",
			"c66-4",
			"cesium",				// No number
			"c6-1234567890123456789012345678901234567890", // Numbers out of range
			"c6-1234567890123456789012345678901234567891", // Numbers out of range
		},
		map[string]bool{
			"c6-[1-3]": true,
			"c66-4":    true,
			"cesium":   true,
			"c6-1234567890123456789012345678901234567890": true,
			"c6-1234567890123456789012345678901234567891": true,
		})
	testCompress(
		t,
		[]string{
			"c6-1.e1",
			"c6-2.e1",
			"c6-1.e2",
			"c6-2.e2",
		},
		map[string]bool{
			"c6-[1-2].e1": true,
			"c6-[1-2].e2": true,
		})
	testCompress(
		t,
		[]string{
			"gpu-4-ib.fox",
			"gpu-5-ib.fox",
			"gpu-6-ib.fox",
		},
		map[string]bool{
			"gpu-[4-6]-ib.fox": true,
		})
}

func testCompress(t *testing.T, hosts []string, expect map[string]bool) {
	cs := CompressHostnames(hosts)
	if len(cs) != len(expect) {
		t.Fatal(cs)
	}
	for i := 0 ; i < len(cs) ; i++ {
		if !expect[cs[i]] {
			t.Fatal(cs)
		}
	}
}

func TestGlobber1(t *testing.T) {
    hf := NewGlobber(true)
    if hf.Insert("ml8") != nil {
		t.Fatal("Insert 1")
	}
    if hf.Insert("ml3.hpc") != nil {
		t.Fatal("Insert 2")
	}

    // Single-element prefix match against this
    if !hf.Match("ml8.hpc.uio.no") {
		t.Fatal("Match 1")
	}

    // Multi-element prefix match against this
    if !hf.Match("ml3.hpc.uio.no") {
		t.Fatal("Match 2")
	}

    hf = NewGlobber(false)
    if hf.Insert("ml4.hpc.uio.no") != nil {
		t.Fatal("Insert 3")
	}

    // Exhaustive match against this
    if !hf.Match("ml4.hpc.uio.no") {
		t.Fatal("Match 3")
	}
    if hf.Match("ml4.hpc.uio.no.yes") {
		t.Fatal("Match 4")
	}
}

func TestGlobber2(t *testing.T) {
    hf := NewGlobber(true)
	err := hf.Insert("ml[1-3]*")
    if err != nil {
		t.Fatal(err)
	}
    if !hf.Match("ml1") {
		t.Fatal("Match 1")
	}
    if !hf.Match("ml1x") {
		t.Fatal("Match 2")
	}
    if !hf.Match("ml1.uio") {
		t.Fatal("Match 3")
	}
}

func TestGlobber3(t *testing.T) {
    hf := NewGlobber(false)
	err := hf.Insert("c[1-3]-[2,4]")
    if err != nil {
		t.Fatal(err)
	}
    if !hf.Match("c1-2") {
		t.Fatal("Match 1")
	}
    if !hf.Match("c2-2") {
		t.Fatal("Match 2")
	}
    if hf.Match("c2-3") {
		t.Fatal("Match 3")
	}
}

