package sonarlog

import (
	"bytes"
	"math"
	"os"
	"testing"
	"time"
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
	n, err := parseFloat([]byte{})
	if err == nil {
		t.Fatal("Expected error for empty string")
	}
	n, err = parseFloat([]byte("-12"))
	if err == nil {
		t.Fatal("Expected error for negative number")
	}
	n, err = parseFloat([]byte("+12"))
	if err == nil {
		t.Fatal("Expected error for positive sign")
	}
	n, err = parseFloat([]byte("+iNf"))
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsInf(n, 1) {
		t.Fatal("+inf should be positive infinity")
	}
	n, err = parseFloat([]byte("Inf"))
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsInf(n, 1) {
		t.Fatal("+inf should be positive infinity")
	}
	n, err = parseFloat([]byte("-inF"))
	if err == nil {
		t.Fatal("Expected error for negative infinity")
	}
	n, err = parseFloat([]byte("+infinitY"))
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsInf(n, 1) {
		t.Fatal("+inf should be positive infinity")
	}
	n, err = parseFloat([]byte("infInity"))
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsInf(n, 1) {
		t.Fatal("+inf should be positive infinity")
	}
	n, err = parseFloat([]byte("nAn"))
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsNaN(n) {
		t.Fatal("nAn should be NaN")
	}
	n, err = parseFloat([]byte("12"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 12 {
		t.Fatalf("Expected 12 got %v", n)
	}
	n, err = parseFloat([]byte("12.25"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 12.25 {
		t.Fatalf("Expected 12.25 got %v", n)
	}
	n, err = parseFloat([]byte("12a"))
	if err == nil {
		t.Fatal("Expected error (non-digit)")
	}
	// Is this too harsh?
	n, err = parseFloat([]byte("12."))
	if err == nil {
		t.Fatal("Expected error (empty fraction)")
	}
	n, err = parseFloat([]byte(".25"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 0.25 {
		t.Fatalf("Expected 0.25 got %v", n)
	}
	n, err = parseFloat([]byte("12.a"))
	if err == nil {
		t.Fatal("Expected error (non-digit in fraction)")
	}
	n, err = parseFloat([]byte("12e+7"))
	if err == nil {
		t.Fatal("Expected error (non-digit)")
	}
	n, err = parseFloat([]byte("12e7"))
	if err == nil {
		t.Fatal("Expected error (non-digit)")
	}
	n, err = parseFloat([]byte("12e-7"))
	if err == nil {
		t.Fatal("Expected error (non-digit)")
	}
}

func TestParseSonarLogTagged(t *testing.T) {
	// This test file has a blank line that should be skipped
	bs, err := os.ReadFile("../../tests/sonarlog/whitebox-intermingled.csv")
	uf := NewUstrFacade()
	readings, bad, err := ParseSonarLog(bytes.NewReader(bs), uf)
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
	readings, bad, err := ParseSonarLog(bytes.NewReader(bs), uf)
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

func TestFormatCsvnamed(t *testing.T) {
	now := time.Now().UTC().Unix()
	noSet, _ := NewGpuSet("unknown")
	reading := &Sample{
		Version:     StringToUstr("abc"),
		Timestamp:   now,
		Cluster:     StringToUstr("bad"), // This is not currently in the csv representation
		Host:        StringToUstr("hi"),
		Cores:       5,
		MemtotalKib: 10,
		User:        StringToUstr("me"),
		Job:         37,
		Pid:         1337,
		Cmd:         StringToUstr("secret"),
		CpuPct:      0.5,
		CpuKib:      12,
		RssAnonKib:  15,
		Gpus:        noSet,
		GpuPct:      0.25,
		GpuMemPct:   10,
		GpuKib:      14,
		GpuFail:     2,
		CpuTimeSec:  1234,
		Rolledup:    1,
		Flags:       0,
	}
	expected := "v=abc,time=" + time.Unix(now, 0).Format(time.RFC3339) + ",host=hi,user=me,cmd=secret,cores=5,memtotalkib=10,job=37,pid=1337,cpu%=0.5,cpukib=12,rssanonkib=15,gpus=unknown,gpu%=0.25,gpumem%=10,gpukib=14,gpufail=2,cputime_sec=1234,rolledup=1\n"
	s := string(reading.Csvnamed())
	if s != expected {
		t.Fatalf("Bad csv:\nWant: %s\nGot:  %s", expected, s)
	}
}
