package zapr

import (
	"log"
	"net/url"

	"github.com/go-logr/logr"
	"github.com/tomhjx/xlog/internal/severity"
	"github.com/tomhjx/xlog/option"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type lumberjackSink struct {
	*lumberjack.Logger
}

// Sync implements zap.Sink. The remaining methods are implemented
// by the embedded *lumberjack.Logger.
func (lumberjackSink) Sync() error { return nil }

func New(op option.LogOption) logr.Logger {

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(severity.Flag(LevelSeverity(l)))
	}

	zap.RegisterSink("file", func(u *url.URL) (zap.Sink, error) {
		return lumberjackSink{&lumberjack.Logger{
			Filename:   u.Opaque,
			MaxSize:    op.MaxSizeMB,
			MaxAge:     op.MaxAgeDay,
			MaxBackups: op.MaxBackups,
			LocalTime:  true,
		}}, nil
	})

	zc := zap.NewProductionConfig()
	zc.EncoderConfig = encoderConfig
	if op.OutputPath != "" {
		zc.OutputPaths = append(zc.OutputPaths, op.OutputPath)
	}
	zl, err := zc.Build()
	if err != nil {
		log.Fatal(err)
	}
	return NewLogger(zl)
}
