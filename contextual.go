package xlog

import (
	"context"
	"sync"

	"github.com/go-logr/logr"
	"github.com/tomhjx/xlog/lib/zapr"
	"github.com/tomhjx/xlog/option"
)

type logWriter struct {
	Logger
}

// TODO can be used as a last resort by code that has no means of
// receiving a logger from its caller. FromContext or an explicit logger
// parameter should be used instead.
func TODO() Logger {
	return Background()
}

// Background retrieves the fallback logger. It should not be called before
// that logger was initialized by the program and not by code that should
// better receive a logger via its parameters. TODO can be used as a temporary
// solution for such code.
func Background() Logger {
	return GlobalLogger().Logger
}

// LoggerWithValues returns logger.WithValues(...kv) when
// contextual logging is enabled, otherwise the logger.
func LoggerWithValues(logger Logger, kv ...interface{}) Logger {
	if logging.contextualLoggingEnabled {
		return logger.WithValues(kv...)
	}
	return logger
}

// LoggerWithName returns logger.WithName(name) when contextual logging is
// enabled, otherwise the logger.
func LoggerWithName(logger Logger, name string) Logger {
	if logging.contextualLoggingEnabled {
		return logger.WithName(name)
	}
	return logger
}

// NewContext returns logr.NewContext(ctx, logger) when
// contextual logging is enabled, otherwise ctx.
func NewContext(ctx context.Context, logger Logger) context.Context {
	if logging.contextualLoggingEnabled {
		return logr.NewContext(ctx, logger)
	}
	return ctx
}

// To remove a backing logr implemention, use ClearLogger. Setting an
// empty logger with SetLogger(logr.Logger{}) does not work.
//
// Modifying the logger is not thread-safe and should be done while no other
// goroutines invoke log calls, usually during program initialization.
func SetLogger(logger logr.Logger) {
	logging.logger = newLogWriter(logger)
}

// ClearLogger removes a backing Logger implementation if one was set earlier
// with SetLogger.
//
// Modifying the logger is not thread-safe and should be done while no other
// goroutines invoke log calls, usually during program initialization.
func ClearLogger() {
	logging.logger = nil
}

func InitGlobalLogger() {

	c := option.LogOption{}
	c.OutputPath = logging.file
	SetLogger(zapr.New(c))
}

var loggingLoggerOnce sync.Once

func GlobalLogger() *logWriter {
	if logging.logger == nil {
		loggingLoggerOnce.Do(InitGlobalLogger)
	}
	return logging.logger
}

func newLogWriter(l Logger) *logWriter {
	return &logWriter{Logger: l}
}

// FromContext retrieves a logger set by the caller or, if not set,
// falls back to the program's global logger (a Logger instance or xlog
// itself).
func FromContext(ctx context.Context) Logger {
	if logging.contextualLoggingEnabled {
		if logger, err := logr.FromContext(ctx); err == nil {
			return logger
		}
	}

	return Background()
}
