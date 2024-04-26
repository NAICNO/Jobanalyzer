package sonarlog

import (
	"bytes"
	"math"
	"os"
	"testing"
)

func TestParseUint(t *testing.T) {
	n, err := parseUint([]byte{})
	if err == nil {
		t.Fatal("Expected error for empty string")
	}
	n, err = parseUint([]byte("a"))
	if err == nil {
		t.Fatal("Expected error for non-digit string")
	}
	n, err = parseUint([]byte("a9"))
	if err == nil {
		t.Fatal("Expected error for non-digit in string")
	}
	n, err = parseUint([]byte("9a8"))
	if err == nil {
		t.Fatal("Expected error for non-digit in string")
	}
	n, err = parseUint([]byte("378"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 378 {
		t.Fatalf("Expected 378, got %v", n)
	}
	n, err = parseUint([]byte("12345678901234567891234567"))
	if err == nil {
		t.Fatal("Expected out-of-range error")
	}
}

func TestParseFloat(t *testing.T) {
	n, err := parseFloat([]byte{}, false)
	if err == nil {
		t.Fatal("Expected error for empty string")
	}
	n, err = parseFloat([]byte("-12"), false)
	if err == nil {
		t.Fatal("Expected error for negative number")
	}
	n, err = parseFloat([]byte("+12"), false)
	if err == nil {
		t.Fatal("Expected error for positive sign")
	}
	n, err = parseFloat([]byte("+iNf"), false)
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsInf(n, 1) {
		t.Fatal("+inf should be positive infinity")
	}
	n, err = parseFloat([]byte("Inf"), false)
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsInf(n, 1) {
		t.Fatal("+inf should be positive infinity")
	}
	n, err = parseFloat([]byte("-inF"), false)
	if err == nil {
		t.Fatal("Expected error for negative infinity")
	}
	n, err = parseFloat([]byte("+infinitY"), false)
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsInf(n, 1) {
		t.Fatal("+inf should be positive infinity")
	}
	n, err = parseFloat([]byte("infInity"), false)
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsInf(n, 1) {
		t.Fatal("+inf should be positive infinity")
	}
	n, err = parseFloat([]byte("nAn"), false)
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsNaN(n) {
		t.Fatal("nAn should be NaN")
	}
	n, err = parseFloat([]byte("12"), false)
	if err != nil {
		t.Fatal(err)
	}
	if n != 12 {
		t.Fatalf("Expected 12 got %v", n)
	}
	n, err = parseFloat([]byte("12.25"), false)
	if err != nil {
		t.Fatal(err)
	}
	if n != 12.25 {
		t.Fatalf("Expected 12.25 got %v", n)
	}
	n, err = parseFloat([]byte("12a"), false)
	if err == nil {
		t.Fatal("Expected error (non-digit)")
	}
	// Is this too harsh?
	n, err = parseFloat([]byte("12."), false)
	if err == nil {
		t.Fatal("Expected error (empty fraction)")
	}
	n, err = parseFloat([]byte(".25"), false)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0.25 {
		t.Fatalf("Expected 0.25 got %v", n)
	}
	n, err = parseFloat([]byte("12.a"), false)
	if err == nil {
		t.Fatal("Expected error (non-digit in fraction)")
	}
	n, err = parseFloat([]byte("12e+7"), false)
	if err == nil {
		t.Fatal("Expected error (non-digit)")
	}
	n, err = parseFloat([]byte("12e7"), false)
	if err == nil {
		t.Fatal("Expected error (non-digit)")
	}
	n, err = parseFloat([]byte("12e-7"), false)
	if err == nil {
		t.Fatal("Expected error (non-digit)")
	}
}

func TestParseSonarLogTagged(t *testing.T) {
	// This test file has a blank line that should be skipped
	bs, err := os.ReadFile("../../tests/sonarlog/whitebox-intermingled.csv")
	uf := NewUstrFacade()
	readings, bad, err := ParseSonarLog(bytes.NewReader(bs), uf, true)
	if err != nil {
		t.Fatalf("Unexpected fatal error during parsing: %v", err)
	}
	if bad != 4 {
		// One dropped record because missing 'user'
		// One bad field name 'blague' but record retained
		// One bad field syntax 'cores192' but record retained
		// One dropped record because bad field value '1x92'
		t.Errorf("Expected 4 irritants, got %d", bad)
	}
	if len(readings) != 5 {
		t.Errorf("Expected 5 readings, got %d", len(readings))
	}
	x := readings[0]
	if x.Host.String() != "ml4.hpc.uio.no" || x.User.String() != "root" || x.Cmd.String() != "tuned" {
		t.Errorf("First record is bogus: %v", x)
	}
	if (x.Flags & FlagHeartbeat) != 0 {
		t.Errorf("Expected heartbeat flag to be clear")
	}

	// The third record should be a heartheat
	x = readings[2]
	if x.User.String() != "_sonar_" || x.Cmd.String() != "_heartbeat_" {
		t.Errorf("Expected heartbeat record")
	}
	if (x.Flags & FlagHeartbeat) == 0 {
		t.Errorf("Expected heartbeat flag to be set")
	}
}

func TestParseSonarLogUntagged(t *testing.T) {
	// This test file has a blank line that should be skipped
	bs, err := os.ReadFile("../../tests/sonarlog/whitebox-untagged-intermingled.csv")
	uf := NewUstrFacade()
	readings, bad, err := ParseSonarLog(bytes.NewReader(bs), uf, true)
	if err != nil {
		t.Fatalf("Unexpected fatal error during parsing: %v", err)
	}
	if bad != 5 {
		// First record is missing user
		// Second record has blank field for cores
		// Fourth record has bad syntax for cores
		// Sixth record has a spurious field and so the others are shifted and fail syntax check
		// Seventh record is missing a field at the end
		t.Errorf("Expected 4 irritants, got %d", bad)
	}
	if len(readings) != 2 {
		t.Errorf("Expected 2 readings, got %d", len(readings))
	}
	x := readings[0]
	if x.Host.String() != "ml3.hpc.uio.no" || x.User.String() != "larsbent" || x.Cmd.String() != "python" {
		t.Errorf("First record is bogus: %v", x)
	}
	if (x.Flags & FlagHeartbeat) != 0 {
		t.Errorf("Expected heartbeat flag to be clear")
	}
}
