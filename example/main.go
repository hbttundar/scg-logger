//go:build examples

// This example is excluded from normal builds and tests. Run with: go run -tags examples ./example
package main

import (
	"context"
	"errors"
	"time"

	"github.com/next-trace/scg-logger/logger"
)

func main() {
	// Initialize default logger
	l := logger.MustInitDefault(
		logger.WithService("auth-api"),
		logger.WithLevel("debug"),
	)

	ctx := context.Background()

	//nolint:mnd // demo constant value for example output
	l.DebugCtx(ctx, "debug message", "foo", 42)
	l.InfoCtx(ctx, "user login", "user_id", "12345", "role", "admin")
	l.WarnCtx(ctx, "cache miss", "key", "session:12345")
	l.ErrorCtx(ctx, "failed to process", errors.New("boom"), "attempt", 1)

	// Inject into context and use in downstream function
	ctx = logger.IntoContext(ctx, l)
	doWork(ctx)

	//nolint:mnd // small sleep to flush async writes in some environments
	time.Sleep(10 * time.Millisecond)
}

func doWork(ctx context.Context) {
	l := logger.FromContext(ctx)
	l.InfoCtx(ctx, "doing work", "step", "1")
}
