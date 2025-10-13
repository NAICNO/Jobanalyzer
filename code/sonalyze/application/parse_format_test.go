package application

import (
	"strings"
	"testing"

	"sonalyze/cmd"
	"sonalyze/cmd/parse"
	"sonalyze/db/special"
)

// Basic unit test that all field names are working in the printer.  The input record has all the
// fields with nonzero values.
func TestParseOldFieldNames(t *testing.T) {
	fields := "localtime,host,cores,memtotal,user,pid,job,cmd,cpu_pct,mem_gb,res_gb," +
		"gpus,gpu_pct,gpumem_pct,gpumem_gb,gpu_status,cputime_sec,rolledup,version"
	lines := strings.Split(mockitParse(t, fields), "\n")

	if lines[0] != fields {
		t.Fatalf("Header: Got %s wanted %s", lines[0], fields)
	}

	// The next line should be the lowest timestamped record, but in the order of fields
	// Note GB for KB
	expect := "localtime=2024-10-31 00:00,host=ml6.hpc.uio.no,cores=64,memtotal=251," +
		"user=testuser,pid=2811127,job=1999327,cmd=testprog.cuda,cpu_pct=96.8,mem_gb=8," +
		"res_gb=0,gpus=5,gpu_pct=85,gpumem_pct=16,gpumem_gb=0,gpu_status=3,cputime_sec=1454," +
		"rolledup=4,version=0.9.0"
	if lines[1] != expect {
		t.Fatalf("Line: Got %s wanted %s", lines[1], expect)
	}
}

func TestParseNewFieldNames(t *testing.T) {
	fields := "Timestamp,Hostname,Cores,MemtotalKB,User,Pid,Ppid,Job,Cmd,CpuPct,CpuKB,RssAnonKB," +
		"Gpus,GpuPct,GpuMemPct,GpuKB,GpuFail,CpuTimeSec,Rolledup,Flags,Version"
	lines := strings.Split(mockitParse(t, fields), "\n")

	if lines[0] != fields {
		t.Fatalf("Header: Got %s wanted %s", lines[0], fields)
	}

	// The next line should be the lowest timestamped record, but in the order of fields
	expect := "Timestamp=2024-10-31 00:00,Hostname=ml6.hpc.uio.no,Cores=64,MemtotalKB=263419260," +
		"User=testuser,Pid=2811127,Ppid=1234,Job=1999327,Cmd=testprog.cuda,CpuPct=96.8,CpuKB=9361016," +
		"RssAnonKB=476264,Gpus=5,GpuPct=85,GpuMemPct=16,GpuKB=581632,GpuFail=3,CpuTimeSec=1454," +
		"Rolledup=4,Flags=0,Version=0.9.0"
	if lines[1] != expect {
		t.Fatalf("Line: Got %s wanted %s", lines[1], expect)
	}
}

func mockitParse(t *testing.T, fields string) string {
	var pc parse.ParseCommand
	pc.DatabaseArgs.SetLogFiles([]string{"testdata/parse_format_test/test.csv"}, "logfiles.cluster")
	pc.FormatArgs.Fmt = "csvnamed,header," + fields
	err := pc.Validate()
	if err != nil {
		t.Fatal(err)
	}
	err = cmd.OpenDataStoreFromCommand(&pc)
	if err != nil {
		t.Fatal(err)
	}
	defer special.CloseDataStore()
	var stdout, stderr strings.Builder
	err = LocalSampleOperation(cmd.NewMetaFromCommand(&pc), &pc, nil, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	return stdout.String()
}
