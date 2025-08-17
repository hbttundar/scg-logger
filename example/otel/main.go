package main

import (
	"context"
	"log"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
 "go.opentelemetry.io/otel/semconv/v1.28.0"

	scglogger "github.com/next-trace/scg-logger/logger"
)

func main() {
	// Initialize a local stdout exporter for traces. No network calls.
	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatalf("failed to create stdout exporter: %v", err)
	}

	// Minimal tracer provider with AlwaysSample for demo purposes.
	res, _ := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("scg-logger-demo"),
		),
	)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	defer func() {
		_ = tp.Shutdown(context.Background())
	}()
	otel.SetTracerProvider(tp)

	// Create the scg-logger (JSON to stdout).
	l := scglogger.New(
		scglogger.WithService("demo"),
		scglogger.WithLevel("debug"),
		// scglogger.WithPretty(true), // Uncomment to see text output.
		// scglogger.WithCaller(true), // Uncomment to include source info.
	)
	// Optional: also set as slog default if you want third-party libs to use it.
	_ = setSlogDefault(l)

	// Start a span and log within its context.
	tr := otel.Tracer("scg-logger-demo")
	ctx, span := tr.Start(context.Background(), "demo-operation")
	defer span.End()

	l.InfoCtx(ctx, "hello from scg-logger with OTel span", "k", 1)

	// You should see "trace_id" and "span_id" in the log output.
}

func setSlogDefault(l scglogger.Logger) error {
	// This helper is only for demo convenience. We do not encourage globals in production.
	// If your scg-logger exposes Core() to retrieve underlying *slog.Logger, you can wire it here.
	type coreProvider interface {
		Core() *slog.Logger
	}
	if cp, ok := l.(coreProvider); ok {
		slog.SetDefault(cp.Core())
		return nil
	}
	// No-op if not available; scg-logger API may deliberately not expose it.
	return nil
}
