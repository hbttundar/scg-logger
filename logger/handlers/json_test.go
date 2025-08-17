package handlers_test

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/next-trace/scg-logger/logger/handlers"
)

// TestJSONHandlerEmitsOutput ensures the JSON handler writes structured output.
func TestJSONHandlerEmitsOutput(t *testing.T) {
	var buf bytes.Buffer

	h := handlers.JSON(&buf, slog.HandlerOptions{})
	l := slog.New(h)
	l.Info("json line", "k", "v")

	if buf.Len() == 0 {
		t.Fatal("expected JSON handler to write output")
	}
}
