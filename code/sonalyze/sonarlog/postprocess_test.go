package sonarlog

import (
	"os"
	"testing"

	. "sonalyze/common"
	"sonalyze/db"
)

func TestPostprocessLogCpuUtilPct(t *testing.T) {
	// This file has field names, cputime_sec, pid, and rolledup.  There are two hosts.  One of the
	// records is invalid: it's missing the user name.  Another record has a field with an unknown
	// tag.  Both are counted together as "discarded" which is sort of nuts.
	f, err := os.Open("../../tests/sonarlog/whitebox-logclean.csv")
	if err != nil {
		t.Fatal(err)
	}
	entries, discarded, err := db.ParseSonarLog(f, NewUstrFacade(), true)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 7 {
		t.Fatalf("Expected 7, got %d", len(entries))
	}
	if discarded != 2 {
		t.Fatalf("Expected 2 discarded, got %d", discarded)
	}

	root := StringToUstr("root")
	streams, _ := createInputStreams(entries)
	PostprocessLogPart2(streams, func(r *Sample) bool { return r.S.User != root })

	if len(streams) != 4 {
		t.Fatalf("Expected 4 streams, got %d", len(streams))
	}
	s1, found := streams[InputStreamKey{
		StringToUstr("ml4.hpc.uio.no"),
		JobIdTag + 4093,
		StringToUstr("zabbix_agentd"),
	}]
	if !found {
		t.Fatalf("No entry s1")
	}
	if len(*s1) != 1 {
		t.Fatalf("Expected one element in stream s1")
	}

	s2, found := streams[InputStreamKey{
		StringToUstr("ml4.hpc.uio.no"),
		1090,
		StringToUstr("python"),
	}]
	if !found {
		t.Fatalf("No entry s2")
	}
	if len(*s2) != 3 {
		t.Fatalf("Expected three elements in stream s2")
	}
	if (*s2)[0].S.Timestamp >= (*s2)[1].S.Timestamp {
		t.Fatal("Timestamp 0")
	}
	if (*s2)[1].S.Timestamp >= (*s2)[2].S.Timestamp {
		t.Fatal("Timestamp 1")
	}

	// For this pid (1090) there are three records for ml4, pairwise 300 seconds apart (and
	// disordered in the input), and the cputime_sec field for the second record is 300 seconds
	// higher, giving us 100% utilization for that time window, and for the third record 150 seconds
	// higher, giving us 50% utilization for that window.

	if (*s2)[0].CpuUtilPct != 1473.7 {
		t.Fatal("Util s2 0")
	}
	if (*s2)[1].CpuUtilPct != 100.0 {
		t.Fatal("Util s2 1")
	}
	if (*s2)[2].CpuUtilPct != 50.0 {
		t.Fatal("Util s2 2")
	}

	// This has the same pid *but* a different host, so the utilization for the first record should
	// once again be set to the cpu_pct value.

	s3, found := streams[InputStreamKey{
		StringToUstr("ml5.hpc.uio.no"),
		1090,
		StringToUstr("python"),
	}]
	if !found {
		t.Fatalf("No entry s3")
	}
	if len(*s3) != 1 {
		t.Fatalf("Expected three elements in stream s3")
	}
	if (*s3)[0].CpuUtilPct != 128.0 {
		t.Fatal("Util s3 0")
	}

	s4, found := streams[InputStreamKey{
		StringToUstr("ml4.hpc.uio.no"),
		1089,
		StringToUstr("python"),
	}]
	if !found {
		t.Fatalf("No entry s4")
	}
	if len(*s4) != 1 {
		t.Fatalf("Expected three elements in stream s4")
	}
}

func TestParseVersion(t *testing.T) {
	a, b, c := parseVersion(StringToUstr("1.2.3"))
	if a != 1 || b != 2 || c != 3 {
		t.Fatal("Could not parse version")
	}
	a, b, c = parseVersion(StringToUstr("1.2"))
	if a != 0 || b != 0 || c != 0 {
		t.Fatal("Incorrectly parsed version")
	}
	a, b, c = parseVersion(StringToUstr("x.y.z"))
	if a != 0 || b != 0 || c != 0 {
		t.Fatal("Incorrectly parsed version")
	}
}
