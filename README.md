# scg-logger

A small, production-ready logging abstraction for Go microservices. It wraps Go's standard log/slog and exposes a minimal Logger interface, enabling future backends (zap, zerolog, logrus, …) with zero changes in microservices code.

Key features:
- Minimal Logger interface (code to interface).
- Functional options for configuration (OCP, extensible).
- JSON by default; optional pretty text.
- Optional caller info.
- Context-first, with optional OpenTelemetry correlation (trace_id, span_id).
- Lint-, security-, and test-friendly (≥90% coverage).

## Install

```
go get github.com/next-trace/scg-logger
```

## Quickstart

```go
package main

import (
    "context"
    "github.com/next-trace/scg-logger/logger"
)

func main() {
    // Initialize a logger instance (JSON by default, level=info). Override via options.
    l := logger.MustInitDefault(
        logger.WithService("auth-api"),
        logger.WithLevel("debug"),
        // logger.WithPretty(true),    // optional: human-readable text
        // logger.WithCaller(true),    // optional: include source file/line
    )

    // Recommended: inject into context and pass ctx along your call graph.
    ctx := context.Background()
    ctx = logger.IntoContext(ctx, l)

    // Optionally attach request-scoped fields to ctx (e.g., request ID, user ID).
    ctx = logger.WithFields(ctx, map[string]any{"request_id": "req-001", "user_id": "123"})

    // Retrieve and use the logger downstream.
    // Use l.For(ctx) (or logger.FromContext(ctx).For(ctx)) to enrich with ctx fields automatically.
    logger.FromContext(ctx).For(ctx).InfoCtx(ctx, "user login", "role", "admin")
}
```

## API

- contract.Logger (interface) in package contract
  - For(ctx) Logger
  - DebugCtx(ctx, msg, kv...)
  - InfoCtx(ctx, msg, kv...)
  - WarnCtx(ctx, msg, kv...)
  - ErrorCtx(ctx, msg, err, kv...)

- Initialize (package logger)
  - New(opts ...Option) contract.Logger
  - MustInitDefault(opts ...Option) contract.Logger

- Options
  - WithService(name string)
  - WithLevel("debug"|"info"|"warn"|"error")
  - WithPretty(bool)
  - WithCaller(bool)
  - WithWriter(io.Writer)

- Context helpers
  - IntoContext(ctx, l)
  - FromContext(ctx)
  - WithFields(ctx, fields map[string]any) // attach request-scoped fields; used with l.For(ctx)
    - Note: If no logger is stored in ctx, FromContext returns a no-op logger that emits no output.

## Request-scoped fields and Logger.For
Attach fields into context in middleware and enrich logs downstream using l.For(ctx).

```go
// middleware
func mw(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := logger.WithFields(r.Context(), map[string]any{
            "request_id": "abc-123",
        })
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// service
func doWork(ctx context.Context, l contract.Logger) {
    l.For(ctx).InfoCtx(ctx, "doing work") // includes request_id automatically
}
```

## OpenTelemetry correlation
If a span exists in the provided context, the logger will add `trace_id` and `span_id` to log records automatically.

```go
tr := otel.Tracer("auth")
ctx, span := tr.Start(ctx, "Authenticate")
defer span.End()

logger.FromContext(ctx).InfoCtx(ctx, "processing request")
```

## Development

The repository includes a helper script:

```
./scg lint
./scg security
./scg test
```

CI also runs `golangci-lint`, `govulncheck`, `gosec`, and tests with race detector and coverage.

## Optional: OTel correlation example
The OpenTelemetry span correlation demo lives in a separate sub-module to keep the main library slim and free of exporter dependencies.

Location: example/otel

How to run:

```
cd example/otel
go run .
```

Expected: log lines emitted by scg-logger include "trace_id" and "span_id" when a span is active in the context.