package filesys

import (
	"io/fs"
	"os"
	"path"
	"reflect"
	"sort"
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
	if !reflect.DeepEqual(files, []string{
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

func checkSameTree(t *testing.T, d1, d2 string) {
	a := os.DirFS(d1).(fs.StatFS)
	b := os.DirFS(d2).(fs.StatFS)
	m1, err := fs.Glob(a, "*")
	if err != nil {
		t.Fatal("Glob a", d1)
	}
	m2, err := fs.Glob(b, "*")
	if err != nil {
		t.Fatal("Glob b", d2)
	}
	sort.Sort(sort.StringSlice(m1))
	sort.Sort(sort.StringSlice(m2))
	if !reflect.DeepEqual(m1, m2) {
		t.Fatal("Not the same names", m1, m2)
	}
	for _, m := range m1 {
		i1, err := a.Stat(m)
		if err != nil {
			t.Fatal("Stat", d1)
		}
		i2, err := b.Stat(m)
		if err != nil {
			t.Fatal("Stat", d2)
		}
		if i1.IsDir() != i2.IsDir() {
			t.Fatal("Not both directories or files")
		}
		if i1.IsDir() {
			checkSameTree(t, path.Join(d1, m), path.Join(d2, m))
		} else {
			x, _ := os.ReadFile(path.Join(d1, m))
			y, _ := os.ReadFile(path.Join(d2, m))
			if !reflect.DeepEqual(x, y) {
				t.Fatal("Not same contents", m)
			}
		}
	}
}

func TestCopyDir(t *testing.T) {
	tmp, _ := os.MkdirTemp("", "filesys")
	defer os.RemoveAll(tmp)
	src := "../../tests/sonarlog/whitebox-tree"
	CopyDir(src, tmp)
	checkSameTree(t, src, tmp)
}
