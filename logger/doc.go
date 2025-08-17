// Package logger provides the default slog-based implementation of contract.Logger
// and helpers to work with contexts and configuration.
//
// Key design choices:
//   - Functional options (WithService/WithLevel/WithPretty/WithCaller/WithWriter) for construction.
//   - No global logger: inject contract.Logger via context (IntoContext/FromContext), enabling testability.
//   - Context-aware methods (DebugCtx/InfoCtx/WarnCtx/ErrorCtx) and Logger.For(ctx) to enrich from context.
//   - OpenTelemetry correlation: trace_id and span_id are appended when a valid span is present.
//   - Handlers: thin wrappers around slog JSON/Text handlers for clear defaults and extensibility.
//
// Usage:
//
//	l := logger.New(logger.WithService("payments"), logger.WithLevel("debug"))
//	ctx := logger.IntoContext(context.Background(), l)
//	ctx = logger.WithFields(ctx, map[string]any{"trace_id": "abc-123"})
//	logger.FromContext(ctx).For(ctx).InfoCtx(ctx, "processing", "order_id", 42)
//
// This package aims to be idiomatic Go: small API, clear behavior, no magic globals, and
// consistent context usage. See README for more examples.
package logger
