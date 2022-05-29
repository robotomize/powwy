package logging

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

const loggerKey = contextKey("logger")

var (
	defaultLogger     *zap.SugaredLogger
	defaultLoggerOnce sync.Once
)

// NewDevelopmentLogger logger with config for development
func NewDevelopmentLogger() *zap.SugaredLogger {
	cfg := &zap.Config{
		Encoding:    "console",
		Level:       zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development: true,
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "T",
			LevelKey:       "L",
			NameKey:        "N",
			CallerKey:      "C",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "M",
			StacktraceKey:  "S",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		ErrorOutputPaths: []string{"stderr"},
		OutputPaths:      []string{"stdout"},
	}

	logger, err := cfg.Build()
	if err != nil {
		logger = zap.NewNop()
	}

	return logger.Sugar()
}

func DefaultDevelopmentLogger() *zap.SugaredLogger {
	defaultLoggerOnce.Do(func() {
		defaultLogger = NewDevelopmentLogger()
	})
	return defaultLogger
}

func WithLogger(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func FromContext(ctx context.Context) *zap.SugaredLogger {
	if logger, ok := ctx.Value(loggerKey).(*zap.SugaredLogger); ok {
		return logger
	}

	return DefaultDevelopmentLogger()
}
