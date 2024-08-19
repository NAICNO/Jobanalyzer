package db

import (
	"math"
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

func TestParseSlurmElapsed(t *testing.T) {
	// Minutes and seconds
	n, err := parseSlurmElapsed64([]byte("1:10"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 1*60+10 {
		t.Fatal(n)
	}

	// Hours, minutes, seconds
	n, err = parseSlurmElapsed64([]byte("7:1:10"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 7*3600+1*60+10 {
		t.Fatal(n)
	}

	// Days, hours, minutes, seconds
	n, err = parseSlurmElapsed64([]byte("2-07:1:10"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 2*3600*24+7*3600+1*60+10 {
		t.Fatal(n)
	}

	// Hours are optional even if we have days
	// Micros are ignored
	n, err = parseSlurmElapsed64([]byte("2-1:10.12"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 2*3600*24+1*60+10 {
		t.Fatal(n)
	}

	// Both minutes and seconds required
	_, err = parseSlurmElapsed64([]byte("10"))
	if err == nil {
		t.Fatal("Expected failure")
	}

	// Both minutes and seconds required
	_, err = parseSlurmElapsed64([]byte("3-10"))
	if err == nil {
		t.Fatal("Expected failure")
	}

	// Both minutes and seconds required
	_, err = parseSlurmElapsed64([]byte("3-"))
	if err == nil {
		t.Fatal("Expected failure")
	}

	// H:M:S but number must be nonblank
	_, err = parseSlurmElapsed64([]byte("10:20:"))
	if err == nil {
		t.Fatal("Expected failure")
	}

	// Micros must be nonblank
	_, err = parseSlurmElapsed64([]byte("3:10."))
	if err == nil {
		t.Fatal("Expected failure")
	}

	// Junk at the end
	_, err = parseSlurmElapsed64([]byte("3:10x"))
	if err == nil {
		t.Fatal("Expected failure")
	}
}

func TestParseSlurmBytes(t *testing.T) {
	type test struct {
		s string
		r uint32
	}
	tests := []test{
		{"125", 1},
		{"125G", 125},
		{"12.5G", 13},
		{"0.01G", 1},
		{"125M", 1},
		{"1250M", 2},
		{"125K", 1},
		{"1250000.0001K", 2},
	}
	for _, x := range tests {
		n, err := parseSlurmBytes([]byte(x.s))
		if err != nil {
			t.Fatal(err)
		}
		if n != x.r {
			t.Fatalf("Bad size %d for %s", n, x.s)
		}
	}
}
