package status

import (
	"fmt"
	"log/syslog"
	"os"
)

func Fatal(msg string) {
	Critical(msg)
	fmt.Fprintf(os.Stderr, "FATAL: %s\n", msg)
	os.Exit(1)
}

var logger *syslog.Writer

func Start(logTag string) {
	var err error
	// The "","" address connects us to the Unix syslog daemon.  The priority (INFO) is a
	// placeholder, it will be overridden by all the logger functions below.
	logger, err = syslog.Dial("", "", syslog.LOG_INFO|syslog.LOG_USER, logTag)
	if err != nil {
		Fatal(err.Error())
	}
}

func Critical(msg string) {
	if logger != nil {
		logger.Crit(msg)
	}
}

func Error(msg string) {
	if logger != nil {
		logger.Err(msg)
	}
}

func Warning(msg string) {
	if logger != nil {
		logger.Warning(msg)
	}
}

func Info(msg string) {
	if logger != nil {
		logger.Info(msg)
	}
}
