package handlers_test

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/next-trace/scg-logger/logger/handlers"
)

// TestTextHandlerEmitsOutput ensures the Text handler writes human-readable output.
func TestTextHandlerEmitsOutput(t *testing.T) {
	var buf bytes.Buffer

	h := handlers.Text(&buf, slog.HandlerOptions{})
	l := slog.New(h)
	l.Info("text line", "k", "v")

	if buf.Len() == 0 {
		t.Fatal("expected Text handler to write output")
	}
}
