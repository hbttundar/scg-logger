package logger

import (
	"context"

	"github.com/next-trace/scg-logger/contract"
)

// noopLogger is a no-op implementation used when no logger is present in context.
// It drops all records.
type noopLogger struct{}

func (n noopLogger) For(_ context.Context) contract.Logger                 { return n }
func (noopLogger) DebugCtx(_ context.Context, _ string, _ ...any)          {}
func (noopLogger) InfoCtx(_ context.Context, _ string, _ ...any)           {}
func (noopLogger) WarnCtx(_ context.Context, _ string, _ ...any)           {}
func (noopLogger) ErrorCtx(_ context.Context, _ string, _ error, _ ...any) {}

var _ contract.Logger = (*noopLogger)(nil)

func getNoop() contract.Logger { return noopLogger{} }
