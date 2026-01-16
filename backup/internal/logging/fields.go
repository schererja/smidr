package logging

import (
	"log/slog"
	"time"
)

// Common field constructors for structured logging

// String creates a string field
func String(key, value string) slog.Attr {
	return slog.String(key, value)
}

// Int creates an int field
func Int(key string, value int) slog.Attr {
	return slog.Int(key, value)
}

// Int64 creates an int64 field
func Int64(key string, value int64) slog.Attr {
	return slog.Int64(key, value)
}

// Uint64 creates a uint64 field
func Uint64(key string, value uint64) slog.Attr {
	return slog.Uint64(key, value)
}

// Float64 creates a float64 field
func Float64(key string, value float64) slog.Attr {
	return slog.Float64(key, value)
}

// Bool creates a bool field
func Bool(key string, value bool) slog.Attr {
	return slog.Bool(key, value)
}

// Time creates a time field
func Time(key string, value time.Time) slog.Attr {
	return slog.Time(key, value)
}

// Duration creates a duration field
func Duration(key string, value time.Duration) slog.Attr {
	return slog.Duration(key, value)
}

// Any creates a field with any value
func Any(key string, value any) slog.Attr {
	return slog.Any(key, value)
}

// Err creates an error field
func Err(err error) slog.Attr {
	return slog.Any("error", err)
}

// Group creates a group of fields
func Group(key string, args ...any) slog.Attr {
	return slog.Group(key, args...)
}

// Canonical field names (Smidr Logging Spec)
// These match the canonical log event schema
const (
	// Required fields (always present via Init)
	// ts, level, msg are handled by slog automatically
	FieldService   = "service"   // e.g., "smidr-core", "smidr-agent", "smidrctl"
	FieldComponent = "component" // e.g., "scheduler", "executor", "api"

	// Correlation identifiers (strongly recommended)
	FieldJobID      = "job_id"      // Correlates build logs
	FieldExecutorID = "executor_id" // Trace executor behavior
	FieldRequestID  = "request_id"  // API tracing
	FieldTenant     = "tenant"      // Multi-tenant support

	// Common operational fields
	FieldTaskID    = "task_id"
	FieldAgentID   = "agent_id"
	FieldUserID    = "user_id"
	FieldError     = "error"
	FieldDuration  = "duration"
	FieldStatus    = "status"
	FieldOperation = "operation"
	FieldStep      = "step"
	FieldImage     = "image"
)
