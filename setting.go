package xlog

import (
	"github.com/tomhjx/xlog/internal/severity"
)

func SetVerbosity(v int) {
	logging.setVState(Level(v))
}

func SetSeverity(s severity.Severity) {
	logging.severity = severityValue{
		Severity: s,
	}
}

func SetFile(p string) {
	logging.file = p
}

func SetFileMaxSizeMB(p int) {
	logging.fileMaxSizeMB = p
}

func SetFileMaxAgeDay(p int) {
	logging.fileMaxAgeDay = p
}

func SetFileMaxBackups(p int) {
	logging.fileMaxBackups = p
}

func SwitchContextual(b bool) {
	logging.settings.contextualLoggingEnabled = b
}
