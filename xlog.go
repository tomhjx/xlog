package xlog

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/tomhjx/xlog/internal/severity"
)

// severityValue identifies the sort of log: info, warning etc. It also implements
// the flag.Value interface. The -stderrthreshold flag is of type severity and
// should be modified only through the flag.Value interface. The values match
// the corresponding constants in C++.
type severityValue struct {
	severity.Severity
}

// get returns the value of the severity.
func (s *severityValue) get() severity.Severity {
	return severity.Severity(atomic.LoadInt32((*int32)(&s.Severity)))
}

// set sets the value of the severity.
func (s *severityValue) set(val severity.Severity) {
	atomic.StoreInt32((*int32)(&s.Severity), int32(val))
}

// String is part of the flag.Value interface.
func (s *severityValue) String() string {
	return strconv.FormatInt(int64(s.Severity), 10)
}

// Get is part of the flag.Getter interface.
func (s *severityValue) Get() interface{} {
	return s.Severity
}

// Set is part of the flag.Value interface.
func (s *severityValue) Set(value string) error {
	var threshold severity.Severity
	// Is it a known name?
	if v, ok := severity.ByName(value); ok {
		threshold = v
	} else {
		v, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return err
		}
		threshold = severity.Severity(v)
	}
	logging.severity.set(threshold)
	return nil
}

// Level is treated as a sync/atomic int32.

// Level specifies a level of verbosity for V logs. *Level implements
// flag.Value; the -v flag is of type Level and should be modified
// only through the flag.Value interface.
type Level int32

// get returns the value of the Level.
func (l *Level) get() Level {
	return Level(atomic.LoadInt32((*int32)(l)))
}

// set sets the value of the Level.
func (l *Level) set(val Level) {
	atomic.StoreInt32((*int32)(l), int32(val))
}

// String is part of the flag.Value interface.
func (l *Level) String() string {
	return strconv.FormatInt(int64(*l), 10)
}

// Get is part of the flag.Getter interface.
func (l *Level) Get() interface{} {
	return *l
}

// Set is part of the flag.Value interface.
func (l *Level) Set(value string) error {
	v, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return err
	}
	logging.mu.Lock()
	defer logging.mu.Unlock()
	logging.setVState(Level(v))
	return nil
}

// setVState sets a consistent state for V logging.
// l.mu is held.
func (l *loggingT) setVState(verbosity Level) {
	l.verbosity.set(verbosity)
}

func (l *loggingT) println(s severity.Severity, logger *logWriter, args ...interface{}) {
	l.printlnDepth(s, logger, 1, args...)
}

func (l *loggingT) printlnDepth(s severity.Severity, logger *logWriter, depth int, args ...interface{}) {
	l.output(s, logger, depth, fmt.Sprintln(args...))
}

func (l *loggingT) print(s severity.Severity, logger *logWriter, args ...interface{}) {
	l.printDepth(s, logger, 1, args...)
}

func (l *loggingT) printDepth(s severity.Severity, logger *logWriter, depth int, args ...interface{}) {
	l.output(s, logger, depth, fmt.Sprint(args...))
}

func (l *loggingT) output(s severity.Severity, logger *logWriter, depth int, msg string) {
	if s < l.severity.Severity {
		return
	}
	depth += 2
	if s == severity.ErrorLog || s == severity.FatalLog {
		logger.WithCallDepth(depth).Error(nil, msg)
	} else {
		logger.WithCallDepth(depth).Info(msg)
	}

	if s == severity.FatalLog {

		// If we got here via Exit rather than Fatal, print no stacks.
		if atomic.LoadUint32(&fatalNoStacks) > 0 {
			OsExit(1)
		}

		OsExit(255) // C++ uses -1, which is silly because it's anded with 255 anyway.
	}

}

func (l *loggingT) printf(s severity.Severity, logger *logWriter, format string, args ...interface{}) {
	l.printfDepth(s, logger, 1, format, args...)
}

func (l *loggingT) printfDepth(s severity.Severity, logger *logWriter, depth int, format string, args ...interface{}) {
	l.output(s, logger, depth, fmt.Sprintf(format, args...))
}

// if loggr is specified, will call loggr.Error, otherwise output with logging module.
func (l *loggingT) errorS(err error, logger *logWriter, depth int, msg string, keysAndValues ...interface{}) {
	if severity.ErrorLog < l.severity.Severity {
		return
	}
	logger.WithCallDepth(depth+1).Error(err, msg, keysAndValues...)
}

// if loggr is specified, will call loggr.Info, otherwise output with logging module.
func (l *loggingT) infoS(logger *logWriter, depth int, msg string, keysAndValues ...interface{}) {
	if severity.InfoLog < l.severity.Severity {
		return
	}
	logger.WithCallDepth(depth+1).Info(msg, keysAndValues...)
}

func V(level Level) Verbose {
	return Verbose{logging.verbosity.get() >= level, newLogWriter(GlobalLogger().Logger)}
}

// Verbose is a boolean type that implements Infof (like Printf) etc.
// See the documentation of V for more information.
type Verbose struct {
	enabled bool
	logger  *logWriter
}

// Info is equivalent to the global Info function, guarded by the value of v.
// See the documentation of V for usage.
func (v Verbose) Info(args ...interface{}) {
	if !v.enabled {
		return
	}
	logging.print(severity.InfoLog, v.logger, args...)
}

// InfoDepth is equivalent to the global InfoDepth function, guarded by the value of v.
// See the documentation of V for usage.
func (v Verbose) InfoDepth(depth int, args ...interface{}) {
	if !v.enabled {
		return
	}
	logging.printDepth(severity.InfoLog, v.logger, depth, args...)
}

// Infoln is equivalent to the global Infoln function, guarded by the value of v.
// See the documentation of V for usage.
func (v Verbose) Infoln(args ...interface{}) {
	if !v.enabled {
		return
	}
	logging.println(severity.InfoLog, v.logger, args...)
}

// InfolnDepth is equivalent to the global InfolnDepth function, guarded by the value of v.
// See the documentation of V for usage.
func (v Verbose) InfolnDepth(depth int, args ...interface{}) {
	if !v.enabled {
		return
	}
	logging.printlnDepth(severity.InfoLog, v.logger, depth, args...)
}

// Infof is equivalent to the global Infof function, guarded by the value of v.
// See the documentation of V for usage.
func (v Verbose) Infof(format string, args ...interface{}) {
	if !v.enabled {
		return
	}
	logging.printf(severity.InfoLog, v.logger, format, args...)
}

// InfofDepth is equivalent to the global InfofDepth function, guarded by the value of v.
// See the documentation of V for usage.
func (v Verbose) InfofDepth(depth int, format string, args ...interface{}) {
	if !v.enabled {
		return
	}
	logging.printfDepth(severity.InfoLog, v.logger, depth, format, args...)
}

// InfoS is equivalent to the global InfoS function, guarded by the value of v.
// See the documentation of V for usage.
func (v Verbose) InfoS(msg string, keysAndValues ...interface{}) {
	if !v.enabled {
		return
	}
	logging.infoS(v.logger, 0, msg, keysAndValues...)
}

// InfoSDepth is equivalent to the global InfoSDepth function, guarded by the value of v.
// See the documentation of V for usage.
func (v Verbose) InfoSDepth(depth int, msg string, keysAndValues ...interface{}) {
	if !v.enabled {
		return
	}
	logging.infoS(v.logger, depth, msg, keysAndValues...)
}

// ErrorS is equivalent to the global Error function, guarded by the value of v.
// See the documentation of V for usage.
func (v Verbose) ErrorS(err error, msg string, keysAndValues ...interface{}) {
	if !v.enabled {
		return
	}
	logging.errorS(err, v.logger, 0, msg, keysAndValues...)
}

// Info logs to the INFO log.
// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
func Info(args ...interface{}) {
	logging.print(severity.InfoLog, GlobalLogger(), args...)
}

// InfoDepth acts as Info but uses depth to determine which call frame to log.
// InfoDepth(0, "msg") is the same as Info("msg").
func InfoDepth(depth int, args ...interface{}) {
	logging.printDepth(severity.InfoLog, GlobalLogger(), depth, args...)
}

// Infoln logs to the INFO log.
// Arguments are handled in the manner of fmt.Println; a newline is always appended.
func Infoln(args ...interface{}) {
	logging.println(severity.InfoLog, GlobalLogger(), args...)
}

// InfolnDepth acts as Infoln but uses depth to determine which call frame to log.
// InfolnDepth(0, "msg") is the same as Infoln("msg").
func InfolnDepth(depth int, args ...interface{}) {
	logging.printlnDepth(severity.InfoLog, GlobalLogger(), depth, args...)
}

// Infof logs to the INFO log.
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func Infof(format string, args ...interface{}) {
	logging.printf(severity.InfoLog, GlobalLogger(), format, args...)
}

// InfofDepth acts as Infof but uses depth to determine which call frame to log.
// InfofDepth(0, "msg", args...) is the same as Infof("msg", args...).
func InfofDepth(depth int, format string, args ...interface{}) {
	logging.printfDepth(severity.InfoLog, GlobalLogger(), depth, format, args...)
}

// InfoS structured logs to the INFO log.
// The msg argument used to add constant description to the log line.
// The key/value pairs would be join by "=" ; a newline is always appended.
//
// Basic examples:
// >> klog.InfoS("Pod status updated", "pod", "kubedns", "status", "ready")
// output:
// >> I1025 00:15:15.525108       1 controller_utils.go:116] "Pod status updated" pod="kubedns" status="ready"
func InfoS(msg string, keysAndValues ...interface{}) {
	logging.infoS(GlobalLogger(), 0, msg, keysAndValues...)
}

// Warning logs to the WARNING and INFO logs.
// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
func Warning(args ...interface{}) {
	logging.print(severity.WarningLog, GlobalLogger(), args...)
}

// WarningDepth acts as Warning but uses depth to determine which call frame to log.
// WarningDepth(0, "msg") is the same as Warning("msg").
func WarningDepth(depth int, args ...interface{}) {
	logging.printDepth(severity.WarningLog, GlobalLogger(), depth, args...)
}

// Warningln logs to the WARNING and INFO logs.
// Arguments are handled in the manner of fmt.Println; a newline is always appended.
func Warningln(args ...interface{}) {
	logging.println(severity.WarningLog, GlobalLogger(), args...)
}

// WarninglnDepth acts as Warningln but uses depth to determine which call frame to log.
// WarninglnDepth(0, "msg") is the same as Warningln("msg").
func WarninglnDepth(depth int, args ...interface{}) {
	logging.printlnDepth(severity.WarningLog, GlobalLogger(), depth, args...)
}

// Warningf logs to the WARNING and INFO logs.
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func Warningf(format string, args ...interface{}) {
	logging.printf(severity.WarningLog, GlobalLogger(), format, args...)
}

// WarningfDepth acts as Warningf but uses depth to determine which call frame to log.
// WarningfDepth(0, "msg", args...) is the same as Warningf("msg", args...).
func WarningfDepth(depth int, format string, args ...interface{}) {
	logging.printfDepth(severity.WarningLog, GlobalLogger(), depth, format, args...)
}

// Error logs to the ERROR, WARNING, and INFO logs.
// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
func Error(args ...interface{}) {
	logging.print(severity.ErrorLog, GlobalLogger(), args...)
}

// ErrorDepth acts as Error but uses depth to determine which call frame to log.
// ErrorDepth(0, "msg") is the same as Error("msg").
func ErrorDepth(depth int, args ...interface{}) {
	logging.printDepth(severity.ErrorLog, GlobalLogger(), depth, args...)
}

// Errorln logs to the ERROR, WARNING, and INFO logs.
// Arguments are handled in the manner of fmt.Println; a newline is always appended.
func Errorln(args ...interface{}) {
	logging.println(severity.ErrorLog, GlobalLogger(), args...)
}

// ErrorlnDepth acts as Errorln but uses depth to determine which call frame to log.
// ErrorlnDepth(0, "msg") is the same as Errorln("msg").
func ErrorlnDepth(depth int, args ...interface{}) {
	logging.printlnDepth(severity.ErrorLog, GlobalLogger(), depth, args...)
}

// Errorf logs to the ERROR, WARNING, and INFO logs.
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func Errorf(format string, args ...interface{}) {
	logging.printf(severity.ErrorLog, GlobalLogger(), format, args...)
}

// ErrorfDepth acts as Errorf but uses depth to determine which call frame to log.
// ErrorfDepth(0, "msg", args...) is the same as Errorf("msg", args...).
func ErrorfDepth(depth int, format string, args ...interface{}) {
	logging.printfDepth(severity.ErrorLog, GlobalLogger(), depth, format, args...)
}

// ErrorS structured logs to the ERROR, WARNING, and INFO logs.
// the err argument used as "err" field of log line.
// The msg argument used to add constant description to the log line.
// The key/value pairs would be join by "=" ; a newline is always appended.
//
// Basic examples:
// >> klog.ErrorS(err, "Failed to update pod status")
// output:
// >> E1025 00:15:15.525108       1 controller_utils.go:114] "Failed to update pod status" err="timeout"
func ErrorS(err error, msg string, keysAndValues ...interface{}) {
	logging.errorS(err, GlobalLogger(), 0, msg, keysAndValues...)
}

// ErrorSDepth acts as ErrorS but uses depth to determine which call frame to log.
// ErrorSDepth(0, "msg") is the same as ErrorS("msg").
func ErrorSDepth(depth int, err error, msg string, keysAndValues ...interface{}) {
	logging.errorS(err, GlobalLogger(), depth, msg, keysAndValues...)
}

// fatalNoStacks is non-zero if we are to exit without dumping goroutine stacks.
// It allows Exit and relatives to use the Fatal logs.
var fatalNoStacks uint32

// Fatal logs to the FATAL, ERROR, WARNING, and INFO logs,
// prints stack trace(s), then calls OsExit(255).
//
// Stderr only receives a dump of the current goroutine's stack trace. Log files,
// if there are any, receive a dump of the stack traces in all goroutines.
//
// Callers who want more control over handling of fatal events may instead use a
// combination of different functions:
//   - some info or error logging function, optionally with a stack trace
//     value generated by github.com/go-logr/lib/dbg.Backtrace
//   - Flush to flush pending log data
//   - panic, os.Exit or returning to the caller with an error
//
// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
func Fatal(args ...interface{}) {
	logging.print(severity.FatalLog, GlobalLogger(), args...)
}

// FatalDepth acts as Fatal but uses depth to determine which call frame to log.
// FatalDepth(0, "msg") is the same as Fatal("msg").
func FatalDepth(depth int, args ...interface{}) {
	logging.printDepth(severity.FatalLog, GlobalLogger(), depth, args...)
}

// Fatalln logs to the FATAL, ERROR, WARNING, and INFO logs,
// including a stack trace of all running goroutines, then calls OsExit(255).
// Arguments are handled in the manner of fmt.Println; a newline is always appended.
func Fatalln(args ...interface{}) {
	logging.println(severity.FatalLog, GlobalLogger(), args...)
}

// FatallnDepth acts as Fatalln but uses depth to determine which call frame to log.
// FatallnDepth(0, "msg") is the same as Fatalln("msg").
func FatallnDepth(depth int, args ...interface{}) {
	logging.printlnDepth(severity.FatalLog, GlobalLogger(), depth, args...)
}

// Fatalf logs to the FATAL, ERROR, WARNING, and INFO logs,
// including a stack trace of all running goroutines, then calls OsExit(255).
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func Fatalf(format string, args ...interface{}) {
	logging.printf(severity.FatalLog, GlobalLogger(), format, args...)
}

// FatalfDepth acts as Fatalf but uses depth to determine which call frame to log.
// FatalfDepth(0, "msg", args...) is the same as Fatalf("msg", args...).
func FatalfDepth(depth int, format string, args ...interface{}) {
	logging.printfDepth(severity.FatalLog, GlobalLogger(), depth, format, args...)
}

// InfoSDepth acts as InfoS but uses depth to determine which call frame to log.
// InfoSDepth(0, "msg") is the same as InfoS("msg").
func InfoSDepth(depth int, msg string, keysAndValues ...interface{}) {
	logging.infoS(GlobalLogger(), depth, msg, keysAndValues...)
}

// // fatalNoStacks is non-zero if we are to exit without dumping goroutine stacks.
// // It allows Exit and relatives to use the Fatal logs.
// var fatalNoStacks uint32

// // Exit logs to the FATAL, ERROR, WARNING, and INFO logs, then calls OsExit(1).
// // Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
// func Exit(args ...interface{}) {
// 	atomic.StoreUint32(&fatalNoStacks, 1)
// 	logging.print(severity.FatalLog, logging.logger, logging.filter, args...)
// }

// // ExitDepth acts as Exit but uses depth to determine which call frame to log.
// // ExitDepth(0, "msg") is the same as Exit("msg").
// func ExitDepth(depth int, args ...interface{}) {
// 	atomic.StoreUint32(&fatalNoStacks, 1)
// 	logging.printDepth(severity.FatalLog, logging.logger, logging.filter, depth, args...)
// }

// // Exitln logs to the FATAL, ERROR, WARNING, and INFO logs, then calls OsExit(1).
// func Exitln(args ...interface{}) {
// 	atomic.StoreUint32(&fatalNoStacks, 1)
// 	logging.println(severity.FatalLog, logging.logger, logging.filter, args...)
// }

// // ExitlnDepth acts as Exitln but uses depth to determine which call frame to log.
// // ExitlnDepth(0, "msg") is the same as Exitln("msg").
// func ExitlnDepth(depth int, args ...interface{}) {
// 	atomic.StoreUint32(&fatalNoStacks, 1)
// 	logging.printlnDepth(severity.FatalLog, logging.logger, logging.filter, depth, args...)
// }

// // Exitf logs to the FATAL, ERROR, WARNING, and INFO logs, then calls OsExit(1).
// // Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
// func Exitf(format string, args ...interface{}) {
// 	atomic.StoreUint32(&fatalNoStacks, 1)
// 	logging.printf(severity.FatalLog, logging.logger, logging.filter, format, args...)
// }

// // ExitfDepth acts as Exitf but uses depth to determine which call frame to log.
// // ExitfDepth(0, "msg", args...) is the same as Exitf("msg", args...).
//
//	func ExitfDepth(depth int, format string, args ...interface{}) {
//		atomic.StoreUint32(&fatalNoStacks, 1)
//		logging.printfDepth(severity.FatalLog, logging.logger, logging.filter, depth, format, args...)
//	}

type loggingT struct {
	settings
	// mu protects the remaining elements of this structure and the fields
	// in settingsT which need a mutex lock.
	mu sync.Mutex
}

// Flush flushes all pending log I/O.
func Flush() {
}

type settings struct {
	// contextualLoggingEnabled controls whether contextual logging is
	// active. Disabling it may have some small performance benefit.
	contextualLoggingEnabled bool
	logger                   *logWriter

	severity severityValue

	verbosity      Level // V logging level
	file           string
	fileMaxSizeMB  int
	fileMaxAgeDay  int
	fileMaxBackups int
}

var logging loggingT

func GetSeverityNames() []string {
	return severity.Name
}
