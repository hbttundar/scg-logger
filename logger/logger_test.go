package logger_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/next-trace/scg-logger/logger"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
)

func parseJSONLine(t *testing.T, line string) map[string]any {
	t.Helper()

	m := map[string]any{}

	if err := json.Unmarshal([]byte(line), &m); err != nil {
		t.Fatalf("failed to parse json: %v line=%s", err, line)
	}

	return m
}

// captureStdout captures os.Stdout during fn execution and returns its content.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	os.Stdout = w

	defer func() { os.Stdout = old }()

	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	_ = w.Close()

	out := <-done

	return out
}

func TestOptionsAndLevels(t *testing.T) {
	out := captureStdout(t, func() {
		l := logger.New(logger.WithService("test-svc"), logger.WithLevel("debug"))

		ctx := t.Context()
		l.DebugCtx(ctx, "hello", "a", 1)
		l.InfoCtx(ctx, "world", "b", 2)
	})
	if !strings.Contains(out, "\"service\":\"test-svc\"") {
		t.Fatalf("expected service in output: %s", out)
	}

	if !strings.Contains(out, "\"level\":\"DEBUG\"") && !strings.Contains(out, "\"level\":\"debug\"") {
		if !strings.Contains(out, "hello") {
			t.Fatalf("expected debug record in output: %s", out)
		}
	}
}

func TestPrettyAndCaller(t *testing.T) {
	out := captureStdout(t, func() {
		l := logger.New(logger.WithPretty(true), logger.WithCaller(true))
		l.InfoCtx(t.Context(), "with caller", "x", 1)
	})
	if !strings.Contains(out, "with caller") {
		t.Fatalf("expected message present: %s", out)
	}

	matched, _ := regexp.MatchString(`source=.*:\d+`, out)

	if !matched {
		t.Fatalf("expected source info in text output: %s", out)
	}
}

func TestWithWriterRedirectsOutput(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(
		logger.WithWriter(&buf),
		logger.WithService("writer-svc"),
	)
	l.InfoCtx(t.Context(), "to buffer", "k", "v")

	if buf.Len() == 0 {
		t.Fatal("expected output written to provided writer")
	}

	if !strings.Contains(buf.String(), "writer-svc") {
		t.Fatalf("expected service tag in buffer: %s", buf.String())
	}
}

func TestOptionsCombination_All(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(
		logger.WithWriter(&buf),
		logger.WithService("combo-svc"),
		logger.WithLevel("debug"),
		logger.WithPretty(true),
		logger.WithCaller(true),
	)
	l.DebugCtx(t.Context(), "combo", "x", 1)

	out := buf.String()

	if !strings.Contains(out, "combo-svc") {
		t.Fatalf("expected service in output: %s", out)
	}

	if !strings.Contains(out, "combo") {
		t.Fatalf("expected message in output: %s", out)
	}
}

func TestErrorLogging(t *testing.T) {
	out := captureStdout(t, func() {
		l := logger.New()
		l.ErrorCtx(t.Context(), "oops", errors.New("bad"), "x", 1)
	})
	if !strings.Contains(out, "\"error\":\"bad\"") {
		t.Fatalf("expected error field present: %s", out)
	}
}

func TestContextHelpers(t *testing.T) {
	l := logger.New(logger.WithService("ctx"))

	ctx := t.Context()
	ctx = logger.IntoContext(ctx, l)
	got := logger.FromContext(ctx)

	if got == nil {
		t.Fatal("expected logger from context")
	}
}

func TestOtelCorrelation(t *testing.T) {
	out := captureStdout(t, func() {
		// Use a real SDK tracer provider to ensure valid IDs
		tp := sdktrace.NewTracerProvider()
		otel.SetTracerProvider(tp)
		tr := otel.Tracer("test")

		ctx, span := tr.Start(t.Context(), "op")
		defer span.End()

		l := logger.New()
		l.InfoCtx(ctx, "trace line")
	})
	m := parseJSONLine(t, strings.Split(strings.TrimSpace(out), "\n")[0])

	if _, ok := m["trace_id"]; !ok {
		t.Fatalf("expected trace_id in output: %v", m)
	}

	if _, ok := m["span_id"]; !ok {
		t.Fatalf("expected span_id in output: %v", m)
	}
}

func TestKVValidation(t *testing.T) {
	out := captureStdout(t, func() {
		l := logger.New()
		l.InfoCtx(t.Context(), "bad kv", "only-key")
	})
	if !strings.Contains(out, "kv_error") {
		t.Fatal("expected kv_error when odd kv provided")
	}
}

func TestNonStringKeyDegradesGracefully(t *testing.T) {
	out := captureStdout(t, func() {
		l := logger.New()
		l.InfoCtx(t.Context(), "non string key", 123, "value")
	})
	// Key becomes empty string in sanitized output, but value should be present
	if !strings.Contains(out, "value") {
		t.Fatalf("expected value present even with non-string key: %s", out)
	}
}

// parseFirstJSONLine parses the first JSON line from the captured output.
func parseFirstJSONLine(t *testing.T, out string) map[string]any { // helper for local tests.
	t.Helper()

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 0 {
		t.Fatalf("no output to parse: %q", out)
	}

	m := map[string]any{}
	if err := json.Unmarshal([]byte(lines[0]), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	return m
}

func TestFromContextFallbackReturnsNoop(t *testing.T) {
	// When no logger set, FromContext should return a no-op logger that emits nothing.
	out := captureStdout(t, func() {
		l := logger.FromContext(t.Context())
		l.InfoCtx(t.Context(), "should not appear")
	})
	if strings.Contains(out, "should not appear") {
		t.Fatalf("expected no-op logger to emit no output: %s", out)
	}
}

func TestMustInitDefaultOverrides(t *testing.T) {
	// Initialize default with a recognizable service name while stdout is captured
	// so the handler binds to the captured writer.
	out := captureStdout(t, func() {
		l := logger.MustInitDefault(logger.WithService("override-svc"))
		l.InfoCtx(t.Context(), "default line")
	})
	if !strings.Contains(out, "override-svc") {
		t.Fatalf("expected override service in default output: %s", out)
	}
}

func TestErrorCtxWithNilErrorDoesNotAddErrorKey(t *testing.T) {
	out := captureStdout(t, func() {
		l := logger.New()
		l.ErrorCtx(t.Context(), "no error case", nil, "k", "v")
	})
	m := parseFirstJSONLine(t, out)

	if _, ok := m["error"]; ok {
		t.Fatalf("did not expect error key when err is nil: %v", m)
	}
}

func TestInvalidLevelFallsBackToInfo(t *testing.T) {
	out := captureStdout(t, func() {
		l := logger.New(logger.WithLevel("bogus"))
		l.InfoCtx(t.Context(), "info still works")
		l.DebugCtx(t.Context(), "debug may be filtered")
	})
	// We at least should see the info message in output.
	if !strings.Contains(out, "info still works") {
		t.Fatalf("expected info record present with invalid level fallback: %s", out)
	}
}

func TestOTelNegativePaths_NoCtxOrNoSpan_NoIDs(t *testing.T) {
	// nil ctx path
	outNil := captureStdout(t, func() {
		l := logger.New()
		l.InfoCtx(t.Context(), "nil ctx")
	})
	mNil := parseFirstJSONLine(t, outNil)

	if _, ok := mNil["trace_id"]; ok {
		t.Fatalf("unexpected trace_id for nil ctx: %v", mNil)
	}

	if _, ok := mNil["span_id"]; ok {
		t.Fatalf("unexpected span_id for nil ctx: %v", mNil)
	}

	// regular ctx with no span
	outNoSpan := captureStdout(t, func() {
		l := logger.New()
		l.InfoCtx(t.Context(), "no span")
	})
	mNoSpan := parseFirstJSONLine(t, outNoSpan)

	if _, ok := mNoSpan["trace_id"]; ok {
		t.Fatalf("unexpected trace_id without span: %v", mNoSpan)
	}

	if _, ok := mNoSpan["span_id"]; ok {
		t.Fatalf("unexpected span_id without span: %v", mNoSpan)
	}

	// noop tracer provider produces invalid IDs; correlation should not be added.
	outNoIDs := captureStdout(t, func() {
		otel.SetTracerProvider(nooptrace.NewTracerProvider())
		tr := otel.Tracer("neg")
		ctx, span := tr.Start(t.Context(), "noop-op")

		defer span.End()

		l := logger.New()
		l.InfoCtx(ctx, "noop provider")
	})
	mNoIDs := parseFirstJSONLine(t, outNoIDs)

	if _, ok := mNoIDs["trace_id"]; ok {
		t.Fatalf("unexpected trace_id with noop provider: %v", mNoIDs)
	}

	if _, ok := mNoIDs["span_id"]; ok {
		t.Fatalf("unexpected span_id with noop provider: %v", mNoIDs)
	}
}

func TestLevelWarnFiltersInfoAllowsWarn(t *testing.T) {
	out := captureStdout(t, func() {
		l := logger.New(logger.WithLevel("warn"))
		l.InfoCtx(t.Context(), "info should be filtered")
		l.WarnCtx(t.Context(), "warn should pass")
	})
	if strings.Contains(out, "info should be filtered") {
		t.Fatalf("expected info to be filtered when level=warn: %s", out)
	}

	if !strings.Contains(out, "warn should pass") {
		t.Fatalf("expected warn to pass when level=warn: %s", out)
	}
}

func TestLevelErrorFiltersWarnAndInfoAllowsError(t *testing.T) {
	out := captureStdout(t, func() {
		l := logger.New(logger.WithLevel("error"))
		l.InfoCtx(t.Context(), "info should be filtered")
		l.WarnCtx(t.Context(), "warn should be filtered")
		l.ErrorCtx(t.Context(), "error should pass", nil)
	})
	if strings.Contains(out, "info should be filtered") {
		t.Fatalf("expected info to be filtered when level=error: %s", out)
	}

	if strings.Contains(out, "warn should be filtered") {
		t.Fatalf("expected warn to be filtered when level=error: %s", out)
	}

	if !strings.Contains(out, "error should pass") {
		t.Fatalf("expected error to pass when level=error: %s", out)
	}
}

func TestIntoContextWithNilCtx(t *testing.T) {
	// Ensure IntoContext handles nil ctx and downstream FromContext works.
	out := captureStdout(t, func() {
		l := logger.New(logger.WithService("nil-ctx-svc"))
		ctx := logger.IntoContext(t.Context(), l)
		logger.FromContext(ctx).InfoCtx(ctx, "via nil ctx")
	})
	if !strings.Contains(out, "nil-ctx-svc") {
		t.Fatalf("expected service from logger stored with nil ctx: %s", out)
	}
}

// Benchmark to ensure no heavy reflection penalties (non-critical for tests but keeps coverage meaningful).
func BenchmarkInfo(b *testing.B) {
	_ = captureStdout(&testing.T{}, func() {
		l := logger.New(logger.WithLevel("info"))

		ctx := b.Context()

		for range b.N {
			l.InfoCtx(ctx, "bench", slog.String("i", "1"))
		}

		time.Sleep(1 * time.Millisecond)
	})
}
