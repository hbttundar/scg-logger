package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

// Config holds logger configuration.
type Config struct {
	Service    string
	Level      string    // "debug" | "info" | "warn" | "error"
	Pretty     bool      // text vs JSON
	WithCaller bool      // add source info
	Writer     io.Writer // optional, default stdout
}

// Option is a functional option to modify Config.
type Option func(*Config)

// WithService sets the service name.
func WithService(name string) Option {
	return func(c *Config) { c.Service = name }
}

// WithLevel sets the logging level ("debug", "info", "warn", "error").
func WithLevel(level string) Option {
	return func(c *Config) { c.Level = level }
}

// WithPretty toggles human-readable text output (true) vs JSON (false).
func WithPretty(pretty bool) Option {
	return func(c *Config) { c.Pretty = pretty }
}

// WithCaller toggles inclusion of caller/source information.
func WithCaller(enabled bool) Option {
	return func(c *Config) { c.WithCaller = enabled }
}

// WithWriter sets the output writer; defaults to os.Stdout when nil.
func WithWriter(w io.Writer) Option {
	return func(c *Config) { c.Writer = w }
}

// applyOptions builds a Config with defaults then applies options.
func applyOptions(opts ...Option) Config {
	cfg := Config{
		Level:  "info",
		Pretty: false,
		Writer: os.Stdout,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	if cfg.Writer == nil {
		cfg.Writer = os.Stdout
	}

	return cfg
}

// mapLevel converts the string level to slog.Level. Defaults to info.
func mapLevel(lvl string) (slog.Level, error) {
	switch lvl {
	case "debug":
		return slog.LevelDebug, nil
	case "info", "":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid log level: %s", lvl)
	}
}
