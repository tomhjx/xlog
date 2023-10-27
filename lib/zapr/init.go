package zapr

import (
	"github.com/tomhjx/xlog/internal/severity"
	"go.uber.org/zap/zapcore"
)

var (
	SeverityLevel func(severity.Severity) zapcore.Level
	LevelSeverity func(zapcore.Level) severity.Severity
)

func init() {

	severityLevels := map[severity.Severity]zapcore.Level{
		severity.InfoLog:    zapcore.InfoLevel,
		severity.WarningLog: zapcore.WarnLevel,
		severity.ErrorLog:   zapcore.ErrorLevel,
		severity.FatalLog:   zapcore.FatalLevel,
	}

	levelSeveritys := map[zapcore.Level]severity.Severity{}
	for s, l := range severityLevels {
		levelSeveritys[l] = s
	}

	SeverityLevel = func(s severity.Severity) zapcore.Level {
		l, ok := severityLevels[s]
		if !ok {
			return zapcore.InfoLevel
		}
		return l
	}
	LevelSeverity = func(l zapcore.Level) severity.Severity {
		s, ok := levelSeveritys[l]
		if !ok {
			return severity.InfoLog
		}
		return s
	}
}
