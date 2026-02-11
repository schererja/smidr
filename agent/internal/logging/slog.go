package logging

import (
	"context"
	"log/slog"
)

// FromSlog creates a Logger from an slog.Logger
func FromSlog(logger *slog.Logger) *Logger {
	return &Logger{Logger: logger}
}

// ToSlog returns the underlying slog.Logger
func (l *Logger) ToSlog() *slog.Logger {
	return l.Logger
}

// Enabled reports whether the logger emits log records at the given level
func (l *Logger) Enabled(ctx context.Context, level Level) bool {
	return l.Logger.Enabled(ctx, level.toSlogLevel())
}

// LogAttrs logs at the given level with attributes
func (l *Logger) LogAttrs(ctx context.Context, level Level, msg string, attrs ...slog.Attr) {
	l.Logger.LogAttrs(ctx, level.toSlogLevel(), msg, attrs...)
}

// Handler returns the logger's Handler
func (l *Logger) Handler() slog.Handler {
	return l.Logger.Handler()
}

// WithAttrs returns a logger with additional attributes
func (l *Logger) WithAttrs(attrs ...slog.Attr) *Logger {
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return l.With(args...)
}
