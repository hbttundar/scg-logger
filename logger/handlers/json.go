package handlers

import (
	"io"
	"log/slog"
)

// JSON returns a slog.Handler configured for JSON structured logging.
func JSON(w io.Writer, opts slog.HandlerOptions) slog.Handler {
	return slog.NewJSONHandler(w, &opts)
}
