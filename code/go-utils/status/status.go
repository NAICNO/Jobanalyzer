package status

import (
	"fmt"
	"log/syslog"
	"os"
)

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

func Fatal(msg string) {
	if logger != nil {
		logger.Crit(msg)
	}
	fmt.Fprintf(os.Stderr, "FATAL: %s\n", msg)
	os.Exit(1)
}

func Fatalf(format string, args ...any) {
	Fatal(fmt.Sprintf(format, args...))
}

func Critical(msg string) {
	if logger != nil {
		logger.Crit(msg)
	}
	fmt.Fprintf(os.Stderr, "CRITICAL: %s\n", msg)
}

func Criticalf(format string, args ...any) {
	Critical(fmt.Sprintf(format, args...))
}

func Error(msg string) {
	if logger != nil {
		logger.Err(msg)
	}
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", msg)
}

func Errorf(format string, args ...any) {
	Error(fmt.Sprintf(format, args...))
}

func Warning(msg string) {
	if logger != nil {
		logger.Warning(msg)
	}
}

func Warningf(format string, args ...any) {
	Warning(fmt.Sprintf(format, args...))
}

func Info(msg string) {
	if logger != nil {
		logger.Info(msg)
	}
}

func Infof(format string, args ...any) {
	Info(fmt.Sprintf(format, args...))
}
