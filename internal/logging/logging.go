package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
)

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
}

// Config holds configuration for logger initialization
type Config struct {
	Service   string // e.g., "smidr-core", "smidr-agent", "smidrctl"
	Component string // e.g., "scheduler", "executor", "api"
	Level     Level
	JSON      bool // Always use JSON (recommended)
	Output    io.Writer
	AddSource bool
	Pretty    bool // Pretty-print for development (ignored if JSON is false)
}

var (
	defaultLogger *Logger
	defaultConfig Config
)

// Init initializes the default logger with the given configuration
// Service and Component are injected into every log entry
func Init(cfg Config) {
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}

	// JSON is the canonical format
	if !cfg.JSON {
		// Allow text for local dev, but warn
		slog.Warn("text format not recommended for production")
	}

	opts := &slog.HandlerOptions{
		Level:     cfg.Level.toSlogLevel(),
		AddSource: cfg.AddSource,
	}

	var handler slog.Handler
	if cfg.JSON {
		if cfg.Pretty {
			// Pretty JSON for development debugging
			handler = NewPrettyHandler(cfg.Output, opts)
		} else {
			handler = slog.NewJSONHandler(cfg.Output, opts)
		}
	} else {
		handler = slog.NewTextHandler(cfg.Output, opts)
	}

	// Create base logger with canonical fields
	baseLogger := slog.New(handler)

	// Inject service and component into every log entry
	if cfg.Service != "" {
		baseLogger = baseLogger.With("service", cfg.Service)
	}
	if cfg.Component != "" {
		baseLogger = baseLogger.With("component", cfg.Component)
	}

	defaultLogger = &Logger{
		Logger: baseLogger,
	}
	defaultConfig = cfg
	slog.SetDefault(defaultLogger.Logger)
}

// Default returns the default logger
func Default() *Logger {
	if defaultLogger == nil {
		Init(Config{
			Level: LevelInfo,
			JSON:  true,
		})
	}
	return defaultLogger
}

// FromContext extracts a logger from context with all correlation IDs
// This is the primary way to get a logger in application code
func FromContext(ctx context.Context) *Logger {
	return Default().WithContext(ctx)
}

// New creates a new logger with the given handler
func New(handler slog.Handler) *Logger {
	return &Logger{
		Logger: slog.New(handler),
	}
}

// With returns a logger with the given attributes
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger: l.Logger.With(args...),
	}
}

// WithGroup returns a logger with the given group
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{
		Logger: l.Logger.WithGroup(name),
	}
}

// WithContext returns a logger with fields from context
func (l *Logger) WithContext(ctx context.Context) *Logger {
	fields := FieldsFromContext(ctx)
	if len(fields) == 0 {
		return l
	}
	return l.With(fields...)
}

// Global convenience functions

// Debug logs at debug level
func Debug(msg string, args ...any) {
	Default().Debug(msg, args...)
}

// Info logs at info level
func Info(msg string, args ...any) {
	Default().Info(msg, args...)
}

// Warn logs at warn level
func Warn(msg string, args ...any) {
	Default().Warn(msg, args...)
}

// Error logs at error level
func Error(msg string, args ...any) {
	Default().Error(msg, args...)
}

// With returns a logger with the given attributes
func With(args ...any) *Logger {
	return Default().With(args...)
}

// WithContext returns a logger with fields from context
func WithContext(ctx context.Context) *Logger {
	return Default().WithContext(ctx)
}
