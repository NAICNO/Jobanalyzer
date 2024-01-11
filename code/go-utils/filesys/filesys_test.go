package filesys

import (
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
