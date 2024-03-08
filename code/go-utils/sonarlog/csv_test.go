package sonarlog

import (
	"bytes"
	"os"
	"testing"
	"time"
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
	if x.Host.String() != "ml4.hpc.uio.no" || x.User.String() != "root" || x.Cmd.String() != "tuned" {
		t.Errorf("First record is bogus: %v", x)
	}
}

func TestCsvnamed1(t *testing.T) {
	now := time.Now().UTC().Unix();
	reading := &SonarReading{
		Version:    StringToUstr("abc"),
		Timestamp:  now,
		Cluster:    StringToUstr("bad"), // This is not currently in the csv representation
		Host:       StringToUstr("hi"),
		Cores:      5,
		MemtotalKib: 10,
		User:       StringToUstr("me"),
		Job:        37,
		Pid:        1337,
		Cmd:        StringToUstr("secret"),
		CpuPct:     0.5,
		CpuKib:     12,
		RssAnonKib: 15,
		Gpus:       StringToUstr("none"),
		GpuPct:     0.25,
		GpuMemPct:  10,
		GpuKib:     14,
		GpuFail:    2,
		CpuTimeSec: 1234,
		Rolledup:   1,
	}
	expected := "v=abc,time=" + time.Unix(now, 0).Format(time.RFC3339) + ",host=hi,user=me,cmd=secret,cores=5,memtotalkib=10,job=37,pid=1337,cpu%=0.5,cpukib=12,rssanonkib=15,gpus=none,gpu%=0.25,gpumem%=10,gpukib=14,gpufail=2,cputime_sec=1234,rolledup=1\n"
	s := string(reading.Csvnamed())
	if s != expected {
		t.Fatalf("Bad csv:\nWant: %s\nGot:  %s", expected, s)
	}
}

func TestCsvnamed2(t *testing.T) {
	now := time.Now().UTC().Unix();
	heartbeat := &SonarHeartbeat{
		Version:   StringToUstr("abc"),
		Timestamp: now,
		Cluster:   StringToUstr("bad"),
		Host:      StringToUstr("hi"),
	}
	expected := "v=abc,time=" + time.Unix(now, 0).Format(time.RFC3339) + ",host=hi,user=_sonar_,cmd=_heartbeat_\n"
	s := string(heartbeat.Csvnamed())
	if s != expected {
		t.Fatalf("Bad csv: %s", s)
	}
}
