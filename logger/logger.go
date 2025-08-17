package logger

import (
	"context"
	"log/slog"

	"github.com/next-trace/scg-logger/contract"
	ih "github.com/next-trace/scg-logger/logger/handlers"
	"github.com/next-trace/scg-logger/utils"
)

// Ensure slogLogger implements the contract.Logger at compile time.
var _ contract.Logger = (*slogLogger)(nil)

// slogLogger is the default implementation wrapping *slog.Logger.
type slogLogger struct {
	core *slog.Logger
	svc  string
}

// New creates a new Logger using functional options.
// Defaults: JSON output, level=info, no caller.
func New(opts ...Option) contract.Logger {
	cfg := applyOptions(opts...)

	lvl, err := mapLevel(cfg.Level)
	if err != nil {
		// keep going with info, but include an attribute to signal configuration issue
		lvl = slog.LevelInfo
	}

	var h slog.Handler

	options := slog.HandlerOptions{Level: lvl, AddSource: cfg.WithCaller}
	writer := cfg.Writer

	if cfg.Pretty {
		h = ih.Text(writer, options)
	} else {
		h = ih.JSON(writer, options)
	}

	core := slog.New(h)
	// Attach service if provided
	if cfg.Service != "" {
		core = core.With("service", cfg.Service)
	}

	return &slogLogger{core: core, svc: cfg.Service}
}

// MustInitDefault initializes and returns a logger, panicking on failure.
// Note: This library intentionally avoids global defaults; inject the returned logger via context.
func MustInitDefault(opts ...Option) contract.Logger {
	l := New(opts...)
	if l == nil {
		panic("logger initialization failed")
	}

	return l
}

// For checks the context for structured fields and returns an enriched logger.
func (l *slogLogger) For(ctx context.Context) contract.Logger {
	if ctx == nil {
		return l
	}

	if v := ctx.Value(logFieldsKey); v != nil {
		if m, ok := v.(map[string]interface{}); ok && len(m) > 0 {
			attrs := make([]any, 0, len(m)*2)
			for k, val := range m {
				attrs = append(attrs, k, val)
			}
			core := l.core.With(attrs...)
			return &slogLogger{core: core, svc: l.svc}
		}
	}
	return l
}

func (l *slogLogger) withCtx(ctx context.Context, kv []any) (context.Context, []any) {
	// Correlate OTel trace/span if available
	kv = addTraceKV(ctx, kv)
	return ctx, kv
}

func (l *slogLogger) DebugCtx(ctx context.Context, msg string, kv ...any) {
	kv = utils.SanitizeKV(kv)

	ctx, kv = l.withCtx(ctx, kv)

	l.core.DebugContext(ctx, msg, kv...)
}

func (l *slogLogger) InfoCtx(ctx context.Context, msg string, kv ...any) {
	kv = utils.SanitizeKV(kv)

	ctx, kv = l.withCtx(ctx, kv)

	l.core.InfoContext(ctx, msg, kv...)
}

func (l *slogLogger) WarnCtx(ctx context.Context, msg string, kv ...any) {
	kv = utils.SanitizeKV(kv)

	ctx, kv = l.withCtx(ctx, kv)

	l.core.WarnContext(ctx, msg, kv...)
}

func (l *slogLogger) ErrorCtx(ctx context.Context, msg string, err error, kv ...any) {
	kv = utils.SanitizeKV(kv)

	ctx, kv = l.withCtx(ctx, kv)

	if err != nil {
		kv = append(kv, slog.String("error", err.Error()))
	}

	l.core.ErrorContext(ctx, msg, kv...)
}
