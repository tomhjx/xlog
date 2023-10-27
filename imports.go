package xlog

import "github.com/go-logr/logr"

// The reason for providing these aliases is to allow code to work with logr
// without directly importing it.

// Logger in this package is exactly the same as logr.Logger.
type Logger = logr.Logger

// LogSink in this package is exactly the same as logr.LogSink.
type LogSink = logr.LogSink

// Runtimeinfo in this package is exactly the same as logr.RuntimeInfo.
type RuntimeInfo = logr.RuntimeInfo

var (
	// New is an alias for logr.New.
	New = logr.New
)
