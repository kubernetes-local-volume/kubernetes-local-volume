package logging

import (
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

