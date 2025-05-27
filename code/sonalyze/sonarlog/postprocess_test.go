package sonarlog

import (
	"math"
	"os"
	"reflect"
	"testing"
	"time"

	"go-utils/config"
	. "sonalyze/common"
	"sonalyze/db"
	"sonalyze/db/repr"
)

func TestRectifyGpuMem(t *testing.T) {
	// The input file has gpukib fields, but by using a config that says to use gpumem% instead we
	// should force the values in the record to something else.
	memsize := 400
	numcards := 1
	cfg := config.NewClusterConfig(
		2,
		"Test",
		"Test",
		[]string{},
		[]string{},
		[]*config.NodeConfigRecord{
			&config.NodeConfigRecord{
				Timestamp: "2024-06-12T11:20:00+02.00",
				Hostname:  "ml4.hpc.uio.no",
				GpuMemPct: true,
				GpuCards:  numcards,
				GpuMemGB:  memsize,
			},
		},
	)
	c, err := db.OpenTransientSampleCluster(
		[]string{"../../tests/sonarlog/whitebox-logclean.csv"},
		cfg,
	)
	if err != nil {
		t.Fatal(err)
	}
	// Reading the file applies the gpu memory rectification.  The records for job 1249151 all say
	// gpumem%=12 so we should see a computed value for gpukib which is different from the gpukib
	// figure in the data.
	var notime time.Time
	sampleBlobs, _, err := c.ReadSamples(notime, notime, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	expect := uint64((memsize * 12) / 100 * 1024 * 1024)
	for _, samples := range sampleBlobs {
		for _, s := range samples {
			if s.Job == 1249151 {
				found++
				if s.GpuKB != expect {
					t.Errorf("GpuKB %v expected %v (%v %v)", s.GpuKB, expect, s.GpuPct, s.GpuMemPct)
				}
			}
		}
	}
	if found == 0 {
		t.Fatalf("No records")
	}
}

func TestPostprocessLogCpuUtilPct(t *testing.T) {
	// This file has field names, cputime_sec, pid, and rolledup.  There are two hosts.  One of the
	// records is invalid: it's missing the user name.  Another record has a field with an unknown
	// tag.  Both are counted together as "discarded" which is sort of nuts.
	f, err := os.Open("../../tests/sonarlog/whitebox-logclean.csv")
	if err != nil {
		t.Fatal(err)
	}
	entries, _, _, discarded, err := db.ParseSampleCSV(f, NewUstrFacade(), true)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 7 {
		t.Fatalf("Expected 7, got %d", len(entries))
	}
	if discarded != 2 {
		t.Fatalf("Expected 2 discarded, got %d", discarded)
	}

	streams, _ := createInputStreams(
		[][]*repr.Sample{
			entries,
		},
		&db.SampleFilter{
			ExcludeUsers: map[Ustr]bool{
				StringToUstr("root"): true,
			},
			To: math.MaxInt64,
		},
		false,
	)
	ComputePerSampleFields(streams)

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
	if (*s2)[0].Timestamp >= (*s2)[1].Timestamp {
		t.Fatal("Timestamp 0")
	}
	if (*s2)[1].Timestamp >= (*s2)[2].Timestamp {
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
	a, b, c = parseVersion(StringToUstr("1.2.3-devel"))
	if a != 1 || b != 2 || c != 103 {
		t.Fatal("Could not parse version")
	}
}

func TestDecodeBase45Delta(t *testing.T) {
	// This is the test from the Sonar sources, it's pretty basic.  The string should represent the
	// array [*1, *0, *29, *43, 1, *11] with * denoting an INITIAL char.
	xs, err := decodeLoadData(repr.EncodedLoadDataFromBytes([]byte(")(t*1b")))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(xs, []uint64{1, 30, 89, 12}) {
		t.Fatal("Failed decode")
	}
}
