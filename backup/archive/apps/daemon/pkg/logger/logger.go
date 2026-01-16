package logger

import (
	"context"
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
}

var isDebug = os.Getenv("DEBUG")

// New creates a new Logger instance
func NewLogger() *Logger {
	var prettyHandler *Handler
	if isDebug == "1" {
		prettyHandler = NewHandler(&slog.HandlerOptions{
			Level:       slog.LevelDebug,
			AddSource:   true,
			ReplaceAttr: nil,
		})
	} else {
		prettyHandler = NewHandler(&slog.HandlerOptions{
			Level:       slog.LevelInfo,
			AddSource:   false,
			ReplaceAttr: nil,
		})
	}
	logger := slog.New(prettyHandler)
	return &Logger{Logger: logger}
}

// withError enhances log attributes with error details if present
func withError(err error, attrs []slog.Attr) []slog.Attr {

	if err == nil {
		return attrs
	}

	return append(attrs, slog.String("error", err.Error()))
}

// Info logs a message at INFO level without context
func (l *Logger) Info(msg string, attrs ...slog.Attr) {
	if l == nil || l.Logger == nil {
		return
	}
	// Convert slog.Attr to interface{} for variadic call
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	l.Logger.Info(msg, args...)
}

// InfoContext logs a message at INFO level with context
func (l *Logger) InfoContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	if l == nil || l.Logger == nil {
		return
	}
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	l.Logger.InfoContext(ctx, msg, args...)
}

// Warn logs a message at WARN level without context
func (l *Logger) Warn(msg string, attrs ...slog.Attr) {
	if l == nil || l.Logger == nil {
		return
	}
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	l.Logger.Warn(msg, args...)
}

// WarnContext logs a message at WARN level with context
func (l *Logger) WarnContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	if l == nil || l.Logger == nil {
		return
	}
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	l.Logger.WarnContext(ctx, msg, args...)
}

// Error logs a message at ERROR level with error details without context
func (l *Logger) Error(msg string, err error, attrs ...slog.Attr) {
	if l == nil || l.Logger == nil {
		return
	}
	allAttrs := withError(err, attrs)
	args := make([]any, len(allAttrs))
	for i, attr := range allAttrs {
		args[i] = attr
	}
	l.Logger.Error(msg, args...)
}

// ErrorContext logs a message at ERROR level with error details and context
func (l *Logger) ErrorContext(ctx context.Context, msg string, err error, attrs ...slog.Attr) {
	if l == nil || l.Logger == nil {
		return
	}
	allAttrs := withError(err, attrs)
	args := make([]any, len(allAttrs))
	for i, attr := range allAttrs {
		args[i] = attr
	}
	l.Logger.ErrorContext(ctx, msg, args...)
}

// Fatal logs a message at ERROR level with error details without context and exits
func (l *Logger) Fatal(msg string, err error, attrs ...slog.Attr) {
	if l == nil || l.Logger == nil {
		os.Exit(1)
		return
	}
	allAttrs := withError(err, attrs)
	args := make([]any, len(allAttrs))
	for i, attr := range allAttrs {
		args[i] = attr
	}
	l.Logger.Error(msg, args...)
	os.Exit(1)
}

// FatalContext logs a message at ERROR level with error details and context and exits
func (l *Logger) FatalContext(ctx context.Context, msg string, err error, attrs ...slog.Attr) {
	if l == nil || l.Logger == nil {
		os.Exit(1)
		return
	}
	allAttrs := withError(err, attrs)
	args := make([]any, len(allAttrs))
	for i, attr := range allAttrs {
		args[i] = attr
	}
	l.Logger.ErrorContext(ctx, msg, args...)
	os.Exit(1)
}

// Debug logs a message at DEBUG level without context
func (l *Logger) Debug(msg string, attrs ...slog.Attr) {
	if l == nil || l.Logger == nil {
		return
	}
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	l.Logger.Debug(msg, args...)
}

// DebugContext logs a message at DEBUG level with context
func (l *Logger) DebugContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	if l == nil || l.Logger == nil {
		return
	}
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	l.Logger.DebugContext(ctx, msg, args...)
}

// With creates a new Logger with the given attributes that will be included in all log messages
func (l *Logger) With(attrs ...slog.Attr) *Logger {
	if l == nil || l.Logger == nil {
		return nil
	}
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return &Logger{Logger: l.Logger.With(args...)}
}
