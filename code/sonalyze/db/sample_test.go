// Some simple (and old) test cases for Sonar `sample` data.

package db

import (
	"bytes"
	"os"
	"testing"

	. "sonalyze/common"
)

func TestParseSonarLogTagged(t *testing.T) {
	// This test file has a blank line that should be skipped
	bs, err := os.ReadFile("../../tests/sonarlog/whitebox-intermingled.csv")
	if err != nil {
		t.Fatalf("Unexpected fatal error during parsing: %v", err)
	}
	uf := NewUstrFacade()
	readings, _, _, bad, err := ParseSampleCSV(bytes.NewReader(bs), uf, true)
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
	if x.Hostname.String() != "ml4.hpc.uio.no" || x.User.String() != "root" || x.Cmd.String() != "tuned" {
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
	if err != nil {
		t.Fatalf("Unexpected fatal error during parsing: %v", err)
	}
	uf := NewUstrFacade()
	readings, _, _, bad, err := ParseSampleCSV(bytes.NewReader(bs), uf, true)
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
	if x.Hostname.String() != "ml3.hpc.uio.no" || x.User.String() != "larsbent" || x.Cmd.String() != "python" {
		t.Errorf("First record is bogus: %v", x)
	}
	if (x.Flags & FlagHeartbeat) != 0 {
		t.Errorf("Expected heartbeat flag to be clear")
	}
}
