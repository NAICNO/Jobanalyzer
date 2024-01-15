package filesys

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"time"
)

// Given the (relative) name of a root directory, a start date, a date past the end date, and a glob
// pattern, find and return all files that match the pattern in the data store, filtering by the
// start date.  The returned names are relative to the data_path.
//
// The path shall be a clean, absolute path that ends in `/` only if the entire path is `/`.
//
// For the dates, only year/month/day are considered, and timestamps should be passed as UTC times
// with hour, minute, second, and nsec as zero.
//
// The pattern shall have no path components and is typically a glob

func EnumerateFiles(data_path string, from time.Time, to time.Time, pattern string) ([]string, error) {
	filesys := os.DirFS(data_path)
	result := []string{}
	for from.Before(to) {
		probe_fn := fmt.Sprintf("%4d/%02d/%02d/%s", from.Year(), from.Month(), from.Day(), pattern)
		matches, err := fs.Glob(filesys, probe_fn)
		if err != nil {
			return nil, err
		}
		result = append(result, matches...)
		from = from.AddDate(0, 0, 1)
	}
	return result, nil
}

type TestFile struct {
	Dir  string
	Name string
	Data []byte
}

/// Create a temp directory, and subdirectories of that and files in those directories as directed,
/// and return the name of the temp directory.  The caller should remove the directory when it
/// is no longer useful, normally by way of `defer`.  If a nil error is returned there will
/// be no directory.

func PopulateTestData(tag string, data ...TestFile) (string, error) {
	tempdir, err := os.MkdirTemp("", tag+"_test")
	if err != nil {
		return "", err
	}
	for _, d := range data {
		err = os.MkdirAll(path.Join(tempdir, d.Dir), 0700)
		if err != nil {
			os.RemoveAll(tempdir)
			return "", err
		}
		err = os.WriteFile(path.Join(tempdir, d.Dir, d.Name), d.Data, 0600)
		if err != nil {
			os.RemoveAll(tempdir)
			return "", err
		}
	}
	return tempdir, nil
}

func CopyFile(from, to string) error {
	data, err := os.ReadFile(from)
	if err != nil {
		return err
	}
	return os.WriteFile(to, data, 0644)
}
