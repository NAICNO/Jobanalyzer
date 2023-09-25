/// Utilities for tests

package util

import (
	"os"
	"path"
)

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
