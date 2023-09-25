package jobstate

import (
	"io"
	"os"
	"path"
	"testing"
	"time"
)

func TestWriteState(t *testing.T) {
	s := NewJobDatabase()
	s1 := &JobState{
		Id:                10,
		Host:              "hello",
		StartedOnOrBefore: time.Date(2023, 6, 14, 16, 0, 0, 0, time.UTC),
		FirstViolation:    time.Date(2023, 6, 15, 10, 20, 30, 0, time.UTC),
		LastSeen:          time.Date(2023, 9, 11, 15, 37, 0, 0, time.UTC),
		IsReported:        false,
	}
	s.Active[JobKey{Id: s1.Id, Host: s1.Host}] = s1

	td_name, err := os.MkdirTemp(os.TempDir(), "naicreport")
	if err != nil {
		t.Fatalf("MkdirTemp failed %q", err)
	}
	err = WriteJobDatabase(td_name, "jobstate.csv", s)
	if err != nil {
		t.Fatalf("Could not write: %q", err)
	}

	// First test: read raw text and make sure it looks OK

	f, err := os.Open(path.Join(td_name, "jobstate.csv"))
	if err != nil {
		t.Fatalf("Open failed %q", err)
	}
	all, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("ReadAll failed %q", err)
	}
	expect := "id=10,host=hello,startedOnOrBefore=2023-06-14T16:00:00Z,firstViolation=2023-06-15T10:20:30Z,lastSeen=2023-09-11T15:37:00Z,isReported=false,isExpired=false\n"
	if string(all) != expect {
		t.Fatalf("File contents wrong %q", all)
	}
	f.Close()

	// Second test: read it as state and make sure the contents look good

	newState, err := ReadJobDatabase(td_name, "jobstate.csv")
	if err != nil {
		t.Fatalf("ReadJobDatabase failed %q", err)
	}
	if len(newState.Active) != 1 {
		t.Fatalf("Bad contents")
	}
	for k, v := range newState.Active {
		if k.Id != s1.Id || k.Host != s1.Host {
			t.Fatalf("Bad key %q", k)
		}
		if v.Id != s1.Id || v.Host != s1.Host || !v.StartedOnOrBefore.Equal(s1.StartedOnOrBefore) ||
			!v.FirstViolation.Equal(s1.FirstViolation) || !v.LastSeen.Equal(s1.LastSeen) ||
			v.IsReported != s1.IsReported {
			t.Fatalf("Bad contents")
		}
	}
}
