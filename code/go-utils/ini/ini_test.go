package ini

import (
	"regexp"
	"strings"
	"testing"
)

func TestIni(t *testing.T) {
	x, err := ParseIni(strings.NewReader(`
# This is a test

[abra]
x=10
 y =20 and some more
#hi
[zappa]  
[zuppa]
z=z
[cadabra]`))
	if err != nil {
		t.Fatal(err)
	}
	if len(x) != 4 {
		t.Fatalf("Expected length 2: %v", x)
	}
	if x["abra"].Name != "abra" ||
		x["zappa"].Name != "zappa" ||
		x["zuppa"].Name != "zuppa" ||
		x["cadabra"].Name != "cadabra" {
		t.Fatalf("Names are wrong: %v", x)
	}

	m := x["abra"].Vars
	if len(m) != 2 {
		t.Fatalf("abra is wrong: %v", x)
	}
	if m["x"] != "10" {
		t.Fatalf("x is wrong: %v", x)
	}
	if m["y"] != "20 and some more" {
		t.Fatalf("y is wrong: %v", x)
	}

	if len(x["zappa"].Vars) > 0 {
		t.Fatalf("zappa is wrong: %v", x)
	}

	m = x["zuppa"].Vars
	if len(m) != 1 {
		t.Fatalf("zuppa is wrong: %v", x)
	}
	if m["z"] != "z" {
		t.Fatalf("z is wrong: %v", x)
	}

	if len(x["cadabra"].Vars) > 0 {
		t.Fatalf("cadabra is wrong: %v", x)
	}

	// Junk before first one
	x, err = ParseIni(strings.NewReader(`
# Another test

junk=10
`))
	if err == nil {
		t.Fatal("Should have failed for junk before first header")
	}
	if matched, _ := regexp.MatchString(`Missing section header`, err.Error()); !matched {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	// Duplicated section name
	x, err = ParseIni(strings.NewReader(`
[hi]
[hi]`))
	if err == nil {
		t.Fatal("Should have failed for duplicated header")
	}
	if matched, _ := regexp.MatchString(`Duplicated section name`, err.Error()); !matched {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	// Duplicated variable name
	x, err = ParseIni(strings.NewReader(`
[hi]
x=5
x=10`))
	if err == nil {
		t.Fatal("Should have failed for duplicated variable name")
	}
	if matched, _ := regexp.MatchString(`Duplicated variable name`, err.Error()); !matched {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	// Malformed content in section
	x, err = ParseIni(strings.NewReader(`
[hi]
x10
x=10`))
	if err == nil {
		t.Fatal("Should have failed for Malformed content")
	}
	if matched, _ := regexp.MatchString(`Malformed content`, err.Error()); !matched {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
}
