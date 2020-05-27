package logging

import (
	"context"

	"go.uber.org/zap"
)

const (
	TraceId = "traceid"
	Key     = "key"
)

type loggerKey struct{}

// This logger is used when there is no logger attached to the context.
// Rather than returning nil and causing a panic, we will use the fallback
// logger. Fallback logger is tagged with logger=fallback to make sure
// that code that doesn't set the logger correctly can be caught at runtime.
var fallbackLogger *zap.SugaredLogger

func init() {
	if logger, err := zap.NewProduction(); err != nil {
		// We failed to create a fallback logger. Our fallback
		// unfortunately falls back to noop.
		fallbackLogger = zap.NewNop().Sugar()
	} else {
		fallbackLogger = logger.Named("fallback").Sugar()
	}
}

func GetLogger() *zap.SugaredLogger {
	return fallbackLogger
}

// WithLogger returns a copy of parent context in which the
// value associated with logger key is the supplied logger.
func WithLogger(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// FromContext returns the logger stored in context.
// Returns nil if no logger is set in context, or if the stored value is
// not of correct type.
func FromContext(ctx context.Context) *zap.SugaredLogger {
	if logger, ok := ctx.Value(loggerKey{}).(*zap.SugaredLogger); ok {
		return logger
	}
	return fallbackLogger
}
