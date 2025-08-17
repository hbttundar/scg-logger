package logger

import (
	"context"

	"github.com/next-trace/scg-logger/contract"
)

// contextKey is an unexported type to avoid collisions for storing the logger instance.
type contextKey struct{}

var loggerKey = contextKey{}

// fieldsContextKey is an unexported string-based key type to avoid collisions for log fields.
type fieldsContextKey string

// logFieldsKey is the private key under which a map[string]interface{} may be stored in ctx.
var logFieldsKey = fieldsContextKey("log_fields")

// IntoContext stores the Logger in the context.
func IntoContext(ctx context.Context, l contract.Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, loggerKey, l)
}

// FromContext retrieves the Logger from context. If not present, returns a no-op logger.
func FromContext(ctx context.Context) contract.Logger {
	if ctx != nil {
		if v := ctx.Value(loggerKey); v != nil {
			if l, ok := v.(contract.Logger); ok && l != nil {
				return l
			}
		}
	}

	return getNoop()
}

// WithFields returns a context containing structured log fields at a private key.
// The expected value is a map[string]interface{}; if ctx already has fields, it merges them.
// This allows middleware to attach correlation fields (e.g., trace_id) to be picked up by Logger.For.
func WithFields(ctx context.Context, fields map[string]any) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(fields) == 0 {
		return ctx
	}
	var base map[string]any
	if v := ctx.Value(logFieldsKey); v != nil {
		if m, ok := v.(map[string]interface{}); ok && m != nil {
			// copy existing to avoid mutating parent context values
			base = make(map[string]any, len(m)+len(fields))
			for k, val := range m {
				base[k] = val
			}
		}
	}
	if base == nil {
		base = make(map[string]any, len(fields))
	}
	for k, val := range fields {
		base[k] = val
	}
	return context.WithValue(ctx, logFieldsKey, base)
}

// Example usage:
//   // In middleware:
//   // func mw(next http.Handler) http.Handler {
//   //     return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//   //         // Add a trace_id (or any fields) to the request context
//   //         ctx := WithFields(r.Context(), map[string]any{"trace_id": "abc-123"})
//   //         next.ServeHTTP(w, r.WithContext(ctx))
//   //     })
//   // }
//   //
//   // In a downstream service function:
//   // func doWork(ctx context.Context, l contract.Logger) {
//   //     l.For(ctx).InfoCtx(ctx, "Doing work") // will include trace_id automatically
//   // }
