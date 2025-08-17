// Package contract defines the logging abstraction used by applications.
//
// Design principles applied:
//   - Code to interface: depend only on contract.Logger; implementation is provided by subpackages.
//   - SOLID: small interface focused on behavior needed by consumers; LSP via clean slog/noop impls.
//   - DRY/KISS: minimal surface area, context-first methods, structured key-value logging.
//   - Go idioms: context-aware APIs (FooCtx), zero globals, functional options live in logger package.
//
// The contract also includes Logger.For(ctx) to derive a context-enriched logger. Middlewares can
// attach fields using logger.WithFields(ctx, map[string]any{"trace_id": "..."}) and downstream code
// can call l.For(ctx).InfoCtx(ctx, "...") to automatically include those fields.
//
// See package logger for the default slog-based implementation and configuration options.
package contract
