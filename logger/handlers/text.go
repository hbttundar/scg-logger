package handlers

import (
	"io"
	"log/slog"
)

// Text returns a slog.Handler configured for human-readable text logging.
func Text(w io.Writer, opts slog.HandlerOptions) slog.Handler {
	return slog.NewTextHandler(w, &opts)
}
