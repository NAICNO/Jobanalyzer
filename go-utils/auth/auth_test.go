package auth

import (
	"testing"
)

func TestAuth(t *testing.T) {
	u, p, err := ParseAuth("auth_test1.txt")
	if err != nil {
		t.Fatal(err)
	}
	if u != "frobnitz" || p != "fizzbuzz" {
		t.Fatalf("Bad user or password: %s %s", u, p)
	}

	u, p, err = ParseAuth("auth_test2.txt")
	if err != nil {
		t.Fatal(err)
	}
	if u != "grunge" || p != "dirge" {
		t.Fatalf("Bad user or password: %s %s", u, p)
	}
}

func TestPwfile(t *testing.T) {
	oracle, err := ParsePasswdFile("auth_test3.txt")
	if err != nil {
		t.Fatal(err)
	}
	if !oracle("grunge", "dirge") {
		t.Fatalf("Failed #1")
	}
	if oracle("grunge", "blapp") {
		t.Fatalf("Failed #2")
	}
	if !oracle("fuzz", "fizz") {
		t.Fatalf("Failed #3")
	}
	if oracle("blum", "fuzz") {
		t.Fatalf("Failed #4")
	}
}
