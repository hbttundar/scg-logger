module github.com/next-trace/scg-logger/example/otel

go 1.25

require (
	github.com/next-trace/scg-logger v0.0.0
	go.opentelemetry.io/otel v1.29.0
	go.opentelemetry.io/otel/sdk v1.29.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.29.0
 go.opentelemetry.io/otel/semconv v1.28.0
)

replace github.com/next-trace/scg-logger => ../..
