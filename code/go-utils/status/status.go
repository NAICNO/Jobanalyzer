// Basic logging infrastructure that we can share and evolve.

package status

import (
	"fmt"
	"io"
	"log/syslog"
	"os"
	"sync"
)

// LogLevel indicates the level of logging that should be done.

type LogLevel int
const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
	LogLevelCritical
)

// Implementations of this must be thread-safe.
type Logger interface {
	// Print only messages at level l or above
	SetLevel(l LogLevel)

	// Lower log level at least to l
	LowerLevelTo(l LogLevel)

	// Print on this stream, if installed
	SetStderr(w io.Writer)

	// Print on this underlying (simpler) logger, if installed - often syslog.
	SetUnderlying(w UnderlyingLogger)

	// Print at various levels.  None of these must exit or panic, the name indicates the log level
	// only.
	Debug(xs ...any)
	Debugf(format string, args ...any)

	Info(xs ...any)
	Infof(format string, args ...any)

	Warning(xs ...any)
	Warningf(format string, args ...any)

	Error(xs ...any)
	Errorf(format string, args ...any)

	Critical(xs ...any)
	Criticalf(format string, args ...any)
}

// Typically the underlying logger would be a syslog thing, and it has a simpler interface.  In
// particular, log/syslog implements UnderlyingLogger.  An underlying logger must be thread-safe.
type UnderlyingLogger interface {
	Debug(m string) error
	Info(m string) error
	Warning(m string) error
	Err(m string) error
	Crit(m string) error
}

type StandardLogger struct {
	sync.Mutex
	level LogLevel
	stderr io.Writer
	underlying UnderlyingLogger
}

// MT: Constant after initialization, thread-safe.
var defaultLogger Logger = &StandardLogger{
	level: LogLevelError,
	stderr: os.Stderr,
	underlying: nil,
}

func Default() Logger {
	return defaultLogger
}

func (sl *StandardLogger) SetLevel(l LogLevel) {
	sl.Lock()
	defer sl.Unlock()

	sl.level = l
}

func (sl *StandardLogger) LowerLevelTo(l LogLevel) {
	sl.Lock()
	defer sl.Unlock()

	if sl.level > l {
		sl.level = l
	}
}

func (sl *StandardLogger) SetStderr(stderr io.Writer) {
	sl.Lock()
	defer sl.Unlock()

	sl.stderr = stderr
}

func (sl *StandardLogger) SetUnderlying(underlying UnderlyingLogger) {
	sl.Lock()
	defer sl.Unlock()

	sl.underlying = underlying
}

func (sl *StandardLogger) Critical(xs ...any) {
	sl.Lock()
	defer sl.Unlock()

	if sl.level >= LogLevelCritical {
		s := fmt.Sprint(xs...)
		if sl.stderr != nil {
			fmt.Fprintln(sl.stderr, s)
		}
		if sl.underlying != nil {
			sl.underlying.Crit(s)
		}
	}
}

func (sl *StandardLogger) Criticalf(format string, args ...any) {
	sl.Lock()
	defer sl.Unlock()

	if sl.level >= LogLevelCritical {
		s := fmt.Sprintf(format, args...)
		if sl.stderr != nil {
			fmt.Fprintln(sl.stderr, s)
		}
		if sl.underlying != nil {
			sl.underlying.Crit(s)
		}
	}
}

func (sl *StandardLogger) Error(xs ...any) {
	sl.Lock()
	defer sl.Unlock()

	if sl.level >= LogLevelError {
		s := fmt.Sprint(xs...)
		if sl.stderr != nil {
			fmt.Fprintln(sl.stderr, s)
		}
		if sl.underlying != nil {
			sl.underlying.Err(s)
		}
	}
}

func (sl *StandardLogger) Errorf(format string, args ...any) {
	sl.Lock()
	defer sl.Unlock()

	if sl.level >= LogLevelError {
		s := fmt.Sprintf(format, args...)
		if sl.stderr != nil {
			fmt.Fprintln(sl.stderr, s)
		}
		if sl.underlying != nil {
			sl.underlying.Err(s)
		}
	}
}

func (sl *StandardLogger) Warning(xs ...any) {
	sl.Lock()
	defer sl.Unlock()

	if sl.level >= LogLevelWarning {
		s := fmt.Sprint(xs...)
		if sl.stderr != nil {
			fmt.Fprintln(sl.stderr, s)
		}
		if sl.underlying != nil {
			sl.underlying.Warning(s)
		}
	}
}

func (sl *StandardLogger) Warningf(format string, args ...any) {
	sl.Lock()
	defer sl.Unlock()

	if sl.level >= LogLevelWarning {
		s := fmt.Sprintf(format, args...)
		if sl.stderr != nil {
			fmt.Fprintln(sl.stderr, s)
		}
		if sl.underlying != nil {
			sl.underlying.Warning(s)
		}
	}
}

func (sl *StandardLogger) Info(xs ...any) {
	sl.Lock()
	defer sl.Unlock()

	if sl.level >= LogLevelInfo {
		s := fmt.Sprint(xs...)
		if sl.stderr != nil {
			fmt.Fprintln(sl.stderr, s)
		}
		if sl.underlying != nil {
			sl.underlying.Info(s)
		}
	}
}

func (sl *StandardLogger) Infof(format string, args ...any) {
	sl.Lock()
	defer sl.Unlock()

	if sl.level >= LogLevelInfo {
		s := fmt.Sprintf(format, args...)
		if sl.stderr != nil {
			fmt.Fprintln(sl.stderr, s)
		}
		if sl.underlying != nil {
			sl.underlying.Info(s)
		}
	}
}

func (sl *StandardLogger) Debug(xs ...any) {
	sl.Lock()
	defer sl.Unlock()

	if sl.level >= LogLevelDebug {
		s := fmt.Sprint(xs...)
		if sl.stderr != nil {
			fmt.Fprintln(sl.stderr, s)
		}
		if sl.underlying != nil {
			sl.underlying.Debug(s)
		}
	}
}

func (sl *StandardLogger) Debugf(format string, args ...any) {
	sl.Lock()
	defer sl.Unlock()

	if sl.level >= LogLevelDebug {
		s := fmt.Sprintf(format, args...)
		if sl.stderr != nil {
			fmt.Fprintln(sl.stderr, s)
		}
		if sl.underlying != nil {
			sl.underlying.Debug(s)
		}
	}
}

// Older API, still useful

func Start(logTag string) {
	// The "","" address connects us to the Unix syslog daemon.  The priority (INFO) is a
	// placeholder, it will be overridden by all the logger functions below.
	logger, err := syslog.Dial("", "", syslog.LOG_INFO|syslog.LOG_USER, logTag)
	if err != nil {
		Fatal(err.Error())
	}
	defaultLogger.SetUnderlying(logger)
}

func Fatal(msg string) {
	defaultLogger.Critical(msg)
	os.Exit(1)
}

func Fatalf(format string, args ...any) {
	defaultLogger.Criticalf(format, args...)
	os.Exit(1)
}

func Critical(msg string) {
	defaultLogger.Critical(msg)
}

func Criticalf(format string, args ...any) {
	defaultLogger.Criticalf(format, args...)
}

func Error(msg string) {
	defaultLogger.Error(msg)
}

func Errorf(format string, args ...any) {
	defaultLogger.Errorf(format, args...)
}

func Warning(msg string) {
	defaultLogger.Warning(msg)
}

func Warningf(format string, args ...any) {
	defaultLogger.Warningf(format, args...)
}

func Info(msg string) {
	defaultLogger.Info(msg)
}

func Infof(format string, args ...any) {
	defaultLogger.Infof(format, args...)
}
