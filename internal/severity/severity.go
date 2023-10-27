package severity

import (
	"strings"
)

type Severity int32 // sync/atomic int32

// These constants identify the log levels in order of increasing severity.
// A message written to a high-severity log file is also written to each
// lower-severity log file.
const (
	InfoLog Severity = iota
	WarningLog
	ErrorLog
	FatalLog
	NumSeverity = 4
)

// Char contains one shortcut letter per severity level.
const Char = "IWEF"

// Name contains one name per severity level.
var Name = []string{
	InfoLog:    "INFO",
	WarningLog: "WARNING",
	ErrorLog:   "ERROR",
	FatalLog:   "FATAL",
}

// ByName looks up a severity level by name.
func ByName(s string) (Severity, bool) {
	s = strings.ToUpper(s)
	for i, name := range Name {
		if name == s {
			return Severity(i), true
		}
	}
	return 0, false
}

func Flag(s Severity) string {
	return string(Char[s])
}
