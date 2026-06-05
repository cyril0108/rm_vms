package logger

import (
	"context"
	"log/slog"
)

// Logger wraps the standard Go slog.Logger to provide chainable data.
type Logger struct {
	prefix string
	lin    *slog.Logger
}

func NewLogger(pre string, v ...any) *Logger {
	return &Logger{
		prefix: pre,
		lin: slog.Default().With(v...),
	}
}

func (l *Logger) Lin(v ...any) *Logger {
	return &Logger{
		lin: l.lin.With(v...),
	}
}

func (l *Logger) Prefix(s string) *Logger {
	return &Logger{
		prefix: l.prefix + s,
		lin: l.lin,
	}
}

/**
 * ======================================================
 * Core Logging Functions
 * ======================================================
 */

func (l *Logger) Debug(msg string, v ...any) {
	l.lin.Debug(l.prefix+msg, v...)
}

func (l *Logger) Info(msg string, v ...any) {
	l.lin.Info(l.prefix+msg, v...)
}

func (l *Logger) Warn(msg string, v ...any) {
	l.lin.Warn(l.prefix+msg, v...)
}

func (l *Logger) Error(msg string, v ...any) {
	l.lin.Error(l.prefix+msg, v...)
}

/**
 * ======================================================
 * Context-Aware Logging Functions
 * ======================================================
 * These are crucial for Go web servers and concurrent workers
 * so slog can extract trace IDs or cancellation states from the context.
 */

func (l *Logger) DebugContext(ctx context.Context, msg string, v ...any) {
	l.lin.DebugContext(ctx, l.prefix+msg, v...)
}

func (l *Logger) InfoContext(ctx context.Context, msg string, v ...any) {
	l.lin.InfoContext(ctx, l.prefix+msg, v...)
}

func (l *Logger) WarnContext(ctx context.Context, msg string, v ...any) {
	l.lin.WarnContext(ctx, l.prefix+msg, v...)
}

func (l *Logger) ErrorContext(ctx context.Context, msg string, v ...any) {
	l.lin.ErrorContext(ctx, l.prefix+msg, v...)
}