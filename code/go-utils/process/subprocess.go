// Abstractions for running subprocesses and capturing their output.

package process

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Run the program with the arguments, collecting its output and returning it.  If there is an error
// in running the program or the program exits with a nonzero code then an error is returned along
// with stderr and stdout is empty, otherwise stdout and stderr are returned but the assumption is
// that the command exited with code zero.

func RunSubprocess(programPath string, arguments []string) (string, string, error) {
	cmd := exec.Command(programPath, arguments...)
	var stdout strings.Builder
	var stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	errs := stderr.String()
	if err != nil {
		return "", errs, errors.Join(fmt.Errorf("While running %s", programPath), err)
	}
	outs := stdout.String()
	return outs, errs, nil
}
