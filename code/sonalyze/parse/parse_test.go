package parse

import (
	"cmp"
	"slices"
	"strings"
	"testing"
	"time"

	uslices "go-utils/slices"

	. "sonalyze/command"
	"sonalyze/db"
	"sonalyze/sonarlog"
)

// Basic unit test that all field names are working in the printer.  The input record has all the
// fields with nonzero values.
func TestOldFieldNames(t *testing.T) {
	samples := setup(t)
	var myOut strings.Builder
	fields := "localtime,host,cores,memtotal,user,pid,job,cmd,cpu_pct,mem_gb,res_gb," +
		"gpus,gpu_pct,gpumem_pct,gpumem_gb,gpu_status,cputime_sec,rolledup,version"
	// The data are unmerged and do not have CpuUtilPct, oh well
	FormatData(
		&myOut,
		strings.Split(fields, ","),
		parseFormatters,
		&FormatOptions{Csv: true, Named: true, Header: true},
		uslices.Map(samples, func(x sonarlog.Sample) any { return x }),
		PrintMods(0),
	)

	// The first line should be the header
	lines := strings.Split(myOut.String(), "\n")
	if lines[0] != fields {
		t.Fatalf("Header: Got %s wanted %s", lines[0], fields)
	}

	// The next line should be the lowest timestamped record, but in the order of fields
	// Note GB for KB
	expect := "localtime=2024-10-31 00:00,host=ml6.hpc.uio.no,cores=64,memtotal=251," +
		"user=testuser,pid=2811127,job=1999327,cmd=testprog.cuda,cpu_pct=96.8,mem_gb=8," +
		"res_gb=0,gpus=5,gpu_pct=85,gpumem_pct=16,gpumem_gb=0,gpu_status=3,cputime_sec=1454,rolledup=4,version=0.9.0"
	if lines[1] != expect {
		t.Fatalf("Line: Got %s wanted %s", lines[1], expect)
	}
}

func TestNewFieldNames(t *testing.T) {
	samples := setup(t)
	var myOut strings.Builder
	fields := "Timestamp,Host,Cores,MemtotalKB,User,Pid,Ppid,Job,Cmd,CpuPct,CpuKB,RssAnonKB," +
		"Gpus,GpuPct,GpuMemPct,GpuKB,GpuFail,CpuTimeSec,Rolledup,Flags,Version"
	// The data are unmerged and do not have CpuUtilPct, oh well
	FormatData(
		&myOut,
		strings.Split(fields, ","),
		parseFormatters,
		&FormatOptions{Csv: true, Named: true, Header: true},
		uslices.Map(samples, func(x sonarlog.Sample) any { return x }),
		PrintMods(0),
	)

	// The first line should be the header
	lines := strings.Split(myOut.String(), "\n")
	if lines[0] != fields {
		t.Fatalf("Header: Got %s wanted %s", lines[0], fields)
	}

	// The next line should be the lowest timestamped record, but in the order of fields
	expect := "Timestamp=2024-10-31 00:00,Host=ml6.hpc.uio.no,Cores=64,MemtotalKB=263419260," +
		"User=testuser,Pid=2811127,Ppid=1234,Job=1999327,Cmd=testprog.cuda,CpuPct=96.8,CpuKB=9361016," +
		"RssAnonKB=476264,Gpus=5,GpuPct=85,GpuMemPct=16,GpuKB=581632,GpuFail=3,CpuTimeSec=1454," +
		"Rolledup=4,Flags=0,Version=0.9.0"
	if lines[1] != expect {
		t.Fatalf("Line: Got %s wanted %s", lines[1], expect)
	}
}

func setup(t *testing.T) []sonarlog.Sample {
	// This emulates the setup of the non-merging variant of parse.Perform()

	c, err := db.OpenTransientSampleCluster([]string{"testdata/test.csv"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	var notime time.Time
	recordBlobs, _, err := c.ReadSamples(notime, notime, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	samples := make([]sonarlog.Sample, 0)
	for _, records := range recordBlobs {
		samples = append(samples, uslices.Map(
			records,
			func(r *db.Sample) sonarlog.Sample {
				return sonarlog.Sample{Sample: r}
			},
		)...)
	}

	slices.SortFunc(samples, func(a, b sonarlog.Sample) int {
		return cmp.Compare(a.Timestamp, b.Timestamp)
	})

	return samples
}
