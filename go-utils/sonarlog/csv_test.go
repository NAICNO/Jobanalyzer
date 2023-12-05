package sonarlog

import (
	"bytes"
	"os"
	"testing"
)

func TestParseSonarCSVNamed(t *testing.T) {
	bs, err := os.ReadFile("../../tests/sonarlog/whitebox-intermingled.csv")
	readings, heartbeats, bad, err := ParseSonarCsvnamed(bytes.NewReader(bs))
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
	if len(readings) != 4 {
		t.Errorf("Expected 4 readings, got %d", len(readings))
	}
	if len(heartbeats) != 1 {
		t.Errorf("Expected 1 heartbeats, got %d", len(heartbeats))
	}
	x := readings[0]
	if x.Host != "ml4.hpc.uio.no" || x.User != "root" || x.Cmd != "tuned" {
		t.Errorf("First record is bogus: %v", x)
	}
}
