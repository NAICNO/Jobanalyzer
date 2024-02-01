package error

import (
	"fmt"
	"os"
)

func Assert(cond bool, msg string) {
	if !cond {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", msg)
		os.Exit(1)
	}
}

func Check(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s: %s\n", msg, err.Error())
		os.Exit(1)
	}
}
