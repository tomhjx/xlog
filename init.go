package xlog

import "github.com/tomhjx/xlog/internal/severity"

func init() {
	SetVerbosity(0)
	SetSeverity(severity.InfoLog)
	SwitchContextual(true)
}
