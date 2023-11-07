package storage

import (
	"io"
	"os"
	"path"
	"testing"
	"time"
)

func TestEnumerateFiles(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %q", err)
	}
	root := path.Join(wd, "../../tests/naicreport/whitebox-tree")
	files, err := EnumerateFiles(
		root,
		time.Date(2023, 5, 30, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 7, 31, 0, 0, 0, 0, time.UTC),
		"a*.csv")
	if err != nil {
		t.Fatalf("EnumerateFiles returned error %q", err)
	}
	if !same(files, []string{
		"2023/05/30/a0.csv",
		"2023/05/31/a1.csv",
		"2023/06/01/a1.csv",
		"2023/06/02/a2.csv",
		"2023/06/04/a4.csv",
		"2023/06/05/a5.csv",
	}) {
		t.Fatalf("EnumerateFiles returned the wrong files %q", files)
	}
}

func TestReadFreeCSV(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %q", err)
	}
	contents, err := ReadFreeCSV(path.Join(wd, "../../tests/naicreport/whitebox-read-free-csv.csv"))
	if err != nil {
		t.Fatalf("ReadFreeCSV failed: %q", err)
	}
	if len(contents) != 2 {
		t.Fatalf("ReadFreeCSV len failed: %d", len(contents))
	}
	// This is the first record:
	// v=0.7.0,time=2023-08-15T13:00:01+02:00,host=ml3.hpc.uio.no,cores=56,user=joachipo,job=998278,pid=0,cmd=python,cpu%=1578.7,cpukib=257282980,gpus=3,gpu%=1566.9,gpumem%=34,gpukib=3188736,cputime_sec=78770,rolledup=28
	x := contents[0]
	if x["v"] != "0.7.0" ||
		x["time"] != "2023-08-15T13:00:01+02:00" ||
		x["host"] != "ml3.hpc.uio.no" ||
		x["cores"] != "56" ||
		x["user"] != "joachipo" ||
		x["job"] != "998278" ||
		x["pid"] != "0" ||
		x["cmd"] != "python" ||
		x["cpu%"] != "1578.7" ||
		x["cpukib"] != "257282980" ||
		x["gpus"] != "3" ||
		x["gpu%"] != "1566.9" ||
		x["gpumem%"] != "34" ||
		x["gpukib"] != "3188736" ||
		x["cputime_sec"] != "78770" ||
		x["rolledup"] != "28" ||
		len(x) != 16 {
		t.Fatalf("Fields are wrong: %q", x)
	}
}

func TestReadFreeCSVOpenErr(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %q", err)
	}
	_, err = ReadFreeCSV(path.Join(wd, "../../tests/naicreport/no-such-file.csv"))
	if err == nil {
		t.Fatalf("open succeeded??")
	}
	_, ok := err.(*os.PathError)
	if !ok {
		t.Fatalf("Unexpected error from opening nonexistent file: %q", err)
	}
}

func TestWriteFreeCSV(t *testing.T) {
	td_name, err := os.MkdirTemp(os.TempDir(), "naicreport")
	if err != nil {
		t.Fatalf("MkdirTemp failed %q", err)
	}

	filename := path.Join(td_name, "test_write")
	contents := []map[string]string{
		map[string]string{"abra": "10", "zappa": "5", "cadabra": "20"},
		map[string]string{"zappa": "1", "cadabra": "3", "abra": "2"},
	}
	err = WriteFreeCSV(
		filename,
		[]string{"zappa", "abra", "cadabra"},
		contents)
	if err != nil {
		t.Fatalf("WriteFreeCSV failed %q", err)
	}

	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Open failed %q", err)
	}
	all, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("ReadAll failed %q", err)
	}
	expect := "zappa=5,abra=10,cadabra=20\nzappa=1,abra=2,cadabra=3\n"
	if string(all) != expect {
		t.Fatalf("File contents wrong %q", all)
	}
}

func same(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestFieldGetters(t *testing.T) {
	success := true
	if GetString(map[string]string{"hi": "ho"}, "hi", &success) != "ho" || !success {
		t.Fatalf("Failed GetString #1")
	}
	GetString(map[string]string{"hi": "ho"}, "hum", &success)
	if success {
		t.Fatalf("Failed GetString #2")
	}

	success = true
	if GetJobMark(map[string]string{"fixit": "107<"}, "fixit", &success) != 107 || !success {
		t.Fatalf("Failed GetJobMark #1")
	}
	if GetJobMark(map[string]string{"fixit": "107>"}, "fixit", &success) != 107 || !success {
		t.Fatalf("Failed GetJobMark #2")
	}
	if GetJobMark(map[string]string{"fixit": "107!"}, "fixit", &success) != 107 || !success {
		t.Fatalf("Failed GetJobMark #3")
	}
	if GetJobMark(map[string]string{"fixit": "107"}, "fixit", &success) != 107 || !success {
		t.Fatalf("Failed GetJobMark #4")
	}
	GetJobMark(map[string]string{"fixit": "107"}, "flux", &success)
	if success {
		t.Fatalf("Failed GetJobMark #5")
	}
	success = true
	GetJobMark(map[string]string{"fixit": "107+"}, "fixit", &success)
	if success {
		t.Fatalf("Failed GetJobMark #6")
	}

	success = true
	if GetUint32(map[string]string{"fixit": "107"}, "fixit", &success) != 107 || !success {
		t.Fatalf("Failed GetUint32 #1")
	}
	GetUint32(map[string]string{"fixit": "107"}, "flux", &success)
	if success {
		t.Fatalf("Failed GetUint32 #2")
	}
	success = true
	GetUint32(map[string]string{"fixit": "107+"}, "fixit", &success)
	if success {
		t.Fatalf("Failed GetUint32 #3")
	}

	success = true
	if GetBool(map[string]string{"fixit": "TRUE"}, "fixit", &success) != true || !success {
		t.Fatalf("Failed GetBool #1")
	}
	GetBool(map[string]string{"fixit": "TRUE"}, "flux", &success)
	if success {
		t.Fatalf("Failed GetBool #2")
	}
	success = true
	GetBool(map[string]string{"fixit": "TRUISH"}, "fixit", &success)
	if success {
		t.Fatalf("Failed GetBool #3")
	}

	success = true
	if GetFloat64(map[string]string{"oops": "10"}, "oops", &success) != 10 || !success {
		t.Fatalf("Failed GetFloat64 #1")
	}
	if GetFloat64(map[string]string{"oops": "-13.5e7"}, "oops", &success) != -13.5e7 || !success {
		t.Fatalf("Failed GetFloat64 #2")
	}
	GetFloat64(map[string]string{"oops": "1"}, "w", &success)
	if success {
		t.Fatalf("Failed GetFloat64 #3")
	}
	success = true
	GetFloat64(map[string]string{"oops": "-13.5f7"}, "oops", &success)
	if success {
		t.Fatalf("Failed GetFloat64 #4")
	}

	success = true
	if GetDateTime(map[string]string{"now": "2023-09-12 08:37"}, "now", &success) !=
		time.Date(2023, 9, 12, 8, 37, 0, 0, time.UTC) || !success {
		t.Fatalf("Failed GetDateTime #1")
	}
	GetDateTime(map[string]string{"now": "2023-09-12 08:37"}, "then", &success)
	if success {
		t.Fatalf("Failed GetDateTime #2")
	}
	success = true
	GetDateTime(map[string]string{"now": "2023-09-12T08:37"}, "now", &success)
	if success {
		t.Fatalf("Failed GetDateTime #3")
	}

	success = true
	if GetRFC3339(map[string]string{"now": "2023-09-12T08:37:00Z"}, "now", &success) !=
		time.Date(2023, 9, 12, 8, 37, 0, 0, time.UTC) || !success {
		t.Fatalf("Failed GetRFC3339 #1")
	}
	GetRFC3339(map[string]string{"now": "2023-09-12 08:37"}, "then", &success)
	if success {
		t.Fatalf("Failed GetRFC3339 #2")
	}
	success = true
	GetRFC3339(map[string]string{"now": "2023-09-12 08:37Z"}, "now", &success)
	if success {
		t.Fatalf("Failed GetRFC3339 #3")
	}
}

func TestReadConfig(t *testing.T) {
	cfg, err := ReadConfig("../../tests/naicreport/whitebox-config.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg) != 3 {
		t.Fatalf("Expected 3 elements")
	}
	if cfg[0].CpuCores != 64 || cfg[0].MemGB != 256 || cfg[0].GpuCards != 8 || cfg[0].GpuMemGB != 88 || cfg[0].GpuMemPct != false {
		t.Fatalf("element 0: %v", cfg[0])
	}
	if cfg[1].CpuCores != 192 || cfg[1].MemGB != 1024 || cfg[1].GpuCards != 4 || cfg[1].GpuMemGB != 0 || cfg[1].GpuMemPct != true {
		t.Fatalf("element 1: %v", cfg[1])
	}
	if cfg[2].CpuCores != 128 || cfg[2].MemGB != 512 || cfg[2].GpuCards != 0 || cfg[2].GpuMemGB != 0 || cfg[2].GpuMemPct != false {
		t.Fatalf("element 2: %v", cfg[2])
	}
}
