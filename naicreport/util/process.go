package util

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Run the program with the arguments, collecting its output and returning it.  If there is an error
// in running the program or if stderr is non-empty then an error is returned.

func RunSubprocess(subpath string, arguments []string) (string, error) {
	cmd := exec.Command(subpath, arguments...)
	var stdout strings.Builder
	var stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	errs := stderr.String()
	if err != nil || errs != "" {
		return "", errors.Join(fmt.Errorf("While running %s", subpath), err, errors.New(errs))
	}
	return stdout.String(), nil
}
