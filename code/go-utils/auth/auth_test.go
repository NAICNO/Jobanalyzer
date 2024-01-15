package auth

import (
	"os"
	"testing"

	"go-utils/filesys"
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
	err := filesys.CopyFile("auth_test3.txt", "t.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("t.txt")
	oracle, err := ReadPasswords("t.txt")
	if err != nil {
		t.Fatal(err)
	}
	if !oracle.Authenticate("grunge", "dirge") {
		t.Fatalf("Failed #1")
	}
	if oracle.Authenticate("grunge", "blapp") {
		t.Fatalf("Failed #2")
	}
	if !oracle.Authenticate("fuzz", "fizz") {
		t.Fatalf("Failed #3")
	}
	if oracle.Authenticate("blum", "fuzz") {
		t.Fatalf("Failed #4")
	}

	err = filesys.CopyFile("auth_test4.txt", "t.txt")
	if err != nil {
		t.Fatal(err)
	}
	err = oracle.Reread()
	if err != nil {
		t.Fatal(err)
	}
	if !oracle.Authenticate("grunge", "dirge") {
		t.Fatalf("Failed #5")
	}
	if oracle.Authenticate("fuzz", "fizz") {
		t.Fatalf("Failed #6")
	}
	if !oracle.Authenticate("bletch", "blum") {
		t.Fatalf("Failed #7")
	}
}
