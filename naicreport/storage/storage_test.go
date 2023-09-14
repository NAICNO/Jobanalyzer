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
	root := path.Join(wd, "../../sonar_test_data0")
	files, err := EnumerateFiles(
		root,
		time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 7, 31, 0, 0, 0, 0, time.UTC),
		"ml8*.csv")
	if err != nil {
		t.Fatalf("EnumerateFiles returned error %q", err)
	}
	if !same(files, []string {
		"2023/05/30/ml8.hpc.uio.no.csv",
		"2023/05/31/ml8.hpc.uio.no.csv",
		"2023/06/01/ml8.hpc.uio.no.csv",
		"2023/06/02/ml8.hpc.uio.no.csv",
		"2023/06/03/ml8.hpc.uio.no.csv",
		"2023/06/04/ml8.hpc.uio.no.csv",
	}) {
		t.Fatalf("EnumerateFiles returned the wrong files %q", files)
	}
}

func TestReadFreeCSV(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %q", err)
	}
	contents, err := ReadFreeCSV(path.Join(wd, "../../sonar_test_data0/2023/08/15/ml3.hpc.uio.no.csv"))
	if err != nil {
		t.Fatalf("ReadFreeCSV failed: %q", err)
	}
	if len(contents) != 33 {
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
	_, err = ReadFreeCSV(path.Join(wd, "../../sonar_test_data0/abracadabra.csv"))
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
	contents := []map[string]string	{
		map[string]string { "abra": "10", "zappa": "5", "cadabra": "20" },
		map[string]string { "zappa": "1", "cadabra": "3", "abra": "2" },
	}
	err = WriteFreeCSV(
		filename,
		[]string { "zappa", "abra", "cadabra" },
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
	if GetString(map[string]string { "hi": "ho" }, "hi", &success) != "ho" || !success {
		t.Fatalf("Failed GetString #1")
	}
	GetString(map[string]string { "hi": "ho" }, "hum", &success)
	if success {
		t.Fatalf("Failed GetString #2")
	}

	success = true
	if GetJobMark(map[string]string {"fixit": "107<"}, "fixit", &success) != 107 || !success {
		t.Fatalf("Failed GetJobMark #1")
	}
	if GetJobMark(map[string]string {"fixit": "107>"}, "fixit", &success) != 107 || !success {
		t.Fatalf("Failed GetJobMark #2")
	}
	if GetJobMark(map[string]string {"fixit": "107!"}, "fixit", &success) != 107 || !success {
		t.Fatalf("Failed GetJobMark #3")
	}
	if GetJobMark(map[string]string {"fixit": "107"}, "fixit", &success) != 107 || !success {
		t.Fatalf("Failed GetJobMark #4")
	}
	GetJobMark(map[string]string {"fixit": "107"}, "flux", &success)
	if success {
		t.Fatalf("Failed GetJobMark #5")
	}
	success = true
	GetJobMark(map[string]string {"fixit": "107+"}, "fixit", &success)
	if success {
		t.Fatalf("Failed GetJobMark #6")
	}
	
	success = true
	if GetFloat64(map[string]string {"oops": "10"}, "oops", &success) != 10 || !success {
		t.Fatalf("Failed GetFloat64 #1")
	}
	if GetFloat64(map[string]string {"oops": "-13.5e7"}, "oops", &success) != -13.5e7 || !success {
		t.Fatalf("Failed GetFloat64 #2")
	}
	GetFloat64(map[string]string {"oops": "1"}, "w", &success)
	if success {
		t.Fatalf("Failed GetFloat64 #3")
	}
	success = true
	GetFloat64(map[string]string {"oops": "-13.5f7"}, "oops", &success)
	if success {
		t.Fatalf("Failed GetFloat64 #4")
	}
	
	success = true
	if GetDateTime(map[string]string {"now": "2023-09-12 08:37"}, "now", &success) !=
		time.Date(2023, 9, 12, 8, 37, 0, 0, time.UTC) || !success {
		t.Fatalf("Failed GetDateTime #1")
	}
	GetDateTime(map[string]string {"now": "2023-09-12 08:37"}, "then", &success)
	if success {
		t.Fatalf("Failed GetDateTime #2")
	}
	success = true
	GetDateTime(map[string]string {"now": "2023-09-12T08:37"}, "now", &success)
	if success {
		t.Fatalf("Failed GetDateTime #3")
	}
}
