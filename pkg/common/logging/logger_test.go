package logging

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestContext(t *testing.T) {
	logger1 := zap.NewNop().Sugar()
	logger2 := zap.NewExample().Sugar()

	// Happy path tests
	ctx := WithLogger(context.Background(), logger1)
	checkFromContext(ctx, logger1, t)

	ctx = WithLogger(ctx, logger2)
	checkFromContext(ctx, logger2, t)

	ctx = WithLogger(ctx, logger1)
	checkFromContext(ctx, logger1, t)

	// Empty logger
	ctx = context.Background()
	checkFromContext(ctx, fallbackLogger, t)

	// Logger with a wrong type
	ctx = context.WithValue(context.Background(), loggerKey{}, zap.NewNop())
	checkFromContext(ctx, fallbackLogger, t)
}

func checkFromContext(ctx context.Context, want *zap.SugaredLogger, t *testing.T) {
	if got := FromContext(ctx); want != got {
		t.Errorf("unexpected logger in context. want: %v, got: %v", want, got)
	}
}
