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
	l.Logger.Info(msg, slog.Any("data", attrs))
}

// InfoContext logs a message at INFO level with context
func (l *Logger) InfoContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.Logger.InfoContext(ctx, msg, slog.Any("data", attrs))
}

// Warn logs a message at WARN level without context
func (l *Logger) Warn(msg string, attrs ...slog.Attr) {
	l.Logger.Warn(msg, slog.Any("data", attrs))
}

// WarnContext logs a message at WARN level with context
func (l *Logger) WarnContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.Logger.WarnContext(ctx, msg, slog.Any("data", attrs))
}

// Error logs a message at ERROR level with error details without context
func (l *Logger) Error(msg string, err error, attrs ...slog.Attr) {
	l.Logger.Error(msg, slog.Any("data", withError(err, attrs)))
}

// ErrorContext logs a message at ERROR level with error details and context
func (l *Logger) ErrorContext(ctx context.Context, msg string, err error, attrs ...slog.Attr) {
	l.Logger.ErrorContext(ctx, msg, slog.Any("data", withError(err, attrs)))
}

// Error logs a message at ERROR level with error details without context
func (l *Logger) Fatal(msg string, err error, attrs ...slog.Attr) {
	l.Logger.Error(msg, slog.Any("data", withError(err, attrs)))
	os.Exit(1)

}

// ErrorContext logs a message at ERROR level with error details and context
func (l *Logger) FatalContext(ctx context.Context, msg string, err error, attrs ...slog.Attr) {
	l.Logger.ErrorContext(ctx, msg, slog.Any("data", withError(err, attrs)))
	os.Exit(1)
}

// Debug logs a message at DEBUG level without context
func (l *Logger) Debug(msg string, attrs ...slog.Attr) {
	l.Logger.Debug(msg, slog.Any("data", attrs))
}

// DebugContext logs a message at DEBUG level with context
func (l *Logger) DebugContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.Logger.DebugContext(ctx, msg, slog.Any("data", attrs))
}

// With creates a new Logger with the given attributes that will be included in all log messages
func (l *Logger) With(attrs ...slog.Attr) *Logger {
	return &Logger{Logger: l.Logger.With(slog.Any("context", attrs))}
}
