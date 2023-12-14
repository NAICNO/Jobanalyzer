package options

import (
	"fmt"
	"io/fs"
	"os"
	"path"
)

// Require a non-empty value that names an existing directory.

func RequireDirectory(optval, optname string) (string, error) {
	if optval == "" {
		return "", fmt.Errorf("Required argument: %s", optname)
	}

	optval = path.Clean(optval)
	info, err := os.DirFS(optval).(fs.StatFS).Stat(".")
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("Bad %s directory %s", optname, optval)
	}

	return optval, nil
}

// Simple utility: Clean a required path and make it absolute

func RequireCleanPath(optval, optname string) (string, error) {
	if optval == "" {
		return "", fmt.Errorf("%s requires a value", optname)
	}

	optval = path.Clean(optval)
	if path.IsAbs(optval) {
		return optval, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return path.Join(wd, optval), nil
}

