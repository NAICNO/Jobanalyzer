package hostglob

import (
	"testing"
)

func TestSplitHostnames(t *testing.T) {
	xs, err := SplitHostnames("yes.no,ml[1-3].hi,ml[1,2],zappa")
	if err != nil {
		t.Fatalf("Hostnames #1: %s", err.Error())
	}
	if len(xs) != 4 || xs[0] != "yes.no" || xs[1] != "ml[1-3].hi" || xs[2] != "ml[1,2]" || xs[3] != "zappa" {
		t.Fatalf("Hostnames #2: %v", xs)
	}
	// Empty input is allowed
	xs, err = SplitHostnames("")
	if err != nil {
		t.Fatalf("Hostnames #3: %s", err.Error())
	}
	if len(xs) != 0 {
		t.Fatalf("Hostnames #4: %v", xs)
	}
	// No closing bracket
	xs, err = SplitHostnames("yes[hi")
	if err == nil {
		t.Fatalf("Should fail #1: %v", xs)
	}
	// Nested opening bracket
	xs, err = SplitHostnames("yes[hi[]")
	if err == nil {
		t.Fatalf("Should fail #2: %v", xs)
	}
	// No opening bracket
	xs, err = SplitHostnames("yes]")
	if err == nil {
		t.Fatalf("Should fail #3: %v", xs)
	}
	// Empty at beginning
	xs, err = SplitHostnames(",yes")
	if err == nil {
		t.Fatalf("Should fail #4: %v", xs)
	}
	// Empty at end
	xs, err = SplitHostnames("yes,")
	if err == nil {
		t.Fatalf("Should fail #5: %v", xs)
	}
	// Empty in the middle
	xs, err = SplitHostnames("yes,,no")
	if err == nil {
		t.Fatalf("Should fail #6: %v", xs)
	}
}

func TestExpandPatterns(t *testing.T) {
	x := ExpandPatterns("ab[1-2,4].cd[3]")
	if len(x) != 3 || x[0] != "ab1.cd3" || x[1] != "ab2.cd3" || x[2] != "ab4.cd3" {
		t.Fatalf("Pattern: %v", x)
	}
	x = ExpandPatterns("ab[].cd")
	if len(x) != 1 || x[0] != "ab[].cd" {
		t.Fatalf("Pattern: %v", x)
	}
}
