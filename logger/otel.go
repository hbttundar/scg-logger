package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// addTraceKV appends OpenTelemetry correlation IDs to kv if a span exists in ctx.
func addTraceKV(ctx context.Context, kv []any) []any {
	if ctx == nil {
		return kv
	}

	span := trace.SpanFromContext(ctx)

	sc := span.SpanContext()

	if !sc.HasTraceID() || !sc.HasSpanID() {
		return kv
	}

	// Append trace_id and span_id keeping original kv intact.
	return append(kv, "trace_id", sc.TraceID().String(), "span_id", sc.SpanID().String())
}
