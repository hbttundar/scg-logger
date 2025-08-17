package contract

import "context"

// Logger is the abstraction microservices depend on.
//
// Code to interface principle: any backend can be used by implementing this interface.
// It is intentionally minimal and context-first for correlation and cancellation support.
//
// Dependency guidance: Microservices must depend on contract.Logger only. The concrete
// implementation (e.g., slog-backed) is internal to this library. Switching to another
// backend (zap, zerolog, logrus, etc.) in the future requires only upgrading this module
// without changing microservice code.
//
// ErrorCtx semantics: when err is nil, the implementation SHOULD NOT emit a misleading
// error field; it may omit the error field or add an auxiliary indicator in a
// backend-specific way. This library omits the error field when err is nil.
//
// All methods accept additional key-value pairs (structured logging). Keys must be strings.
// Values can be of any type but should be JSON-serializable for best results.
//
// For returns a logger derived from the given context. If the context contains
// predefined structured fields (see logger.WithFields), the returned logger is
// enriched with those fields. When no fields are present, For should return the
// original logger instance.
//
// Usage example (middleware and downstream):
//
//	// middleware
//	func mw(next http.Handler) http.Handler {
//	    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	        // Suppose we have a trace id from some source
//	        fields := map[string]any{"trace_id": "abc-123"}
//	        ctx := logger.WithFields(r.Context(), fields)
//	        next.ServeHTTP(w, r.WithContext(ctx))
//	    })
//	}
//
//	// service function
//	func doWork(ctx context.Context, l contract.Logger) {
//	    l.For(ctx).InfoCtx(ctx, "Doing work") // log line will include trace_id
//	}
//
// Note: this library uses Go's slog under the hood (not zap/logrus).
// Import path for WithFields helper: github.com/next-trace/scg-logger/logger
// and interface here is in github.com/next-trace/scg-logger/contract.
// Keep dependencies to the contract in your services.
//
//nolint:interfacebloat // minimal addition for context-aware enrichment.
type Logger interface {
	// Derive a logger with fields from ctx if present.
	For(ctx context.Context) Logger

	DebugCtx(ctx context.Context, msg string, kv ...any)
	InfoCtx(ctx context.Context, msg string, kv ...any)
	WarnCtx(ctx context.Context, msg string, kv ...any)
	ErrorCtx(ctx context.Context, msg string, err error, kv ...any)
}
