package logging

import (
	"log/slog"
	"strings"
)

// Level represents a log level
type Level int

const (
	LevelDebug Level = iota - 1
	LevelInfo
	LevelWarn
	LevelError
)

// String returns the string representation of the level
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// toSlogLevel converts our Level to slog.Level
func (l Level) toSlogLevel() slog.Level {
	switch l {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// ParseLevel parses a level string into a Level
func ParseLevel(s string) Level {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN", "WARNING":
		return LevelWarn
	case "ERROR":
		return LevelError
	default:
		return LevelInfo
	}
}

// SetLevel sets the log level for the default logger
func SetLevel(level Level) {
	if defaultLogger != nil {
		// Note: slog doesn't support dynamic level changes easily
		// This would require reinitializing with a new handler
		// For now, this is a placeholder that could be implemented
		// with a custom handler that supports dynamic levels
	}
}
