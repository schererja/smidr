package logging

import (
	"context"
)

type contextKey string

const (
	logFieldsKey contextKey = "log_fields"
)

// ContextWithFields adds logging fields to the context
func ContextWithFields(ctx context.Context, fields ...any) context.Context {
	if len(fields) == 0 {
		return ctx
	}

	existing := FieldsFromContext(ctx)
	merged := append(existing, fields...)
	return context.WithValue(ctx, logFieldsKey, merged)
}

// FieldsFromContext extracts logging fields from the context
func FieldsFromContext(ctx context.Context) []any {
	if ctx == nil {
		return nil
	}

	if fields, ok := ctx.Value(logFieldsKey).([]any); ok {
		return fields
	}
	return nil
}

// ContextWithField adds a single logging field to the context
func ContextWithField(ctx context.Context, key string, value any) context.Context {
	return ContextWithFields(ctx, key, value)
}

// Canonical Context Helpers (Correlation IDs)
// These are the primary way to add correlation identifiers

// WithJob adds job_id to the context
func WithJob(ctx context.Context, jobID string) context.Context {
	return ContextWithFields(ctx, FieldJobID, jobID)
}

// WithExecutor adds executor_id to the context
func WithExecutor(ctx context.Context, executorID string) context.Context {
	return ContextWithFields(ctx, FieldExecutorID, executorID)
}

// WithRequestID adds request_id to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return ContextWithFields(ctx, FieldRequestID, requestID)
}

// WithTenant adds tenant to the context
func WithTenant(ctx context.Context, tenant string) context.Context {
	return ContextWithFields(ctx, FieldTenant, tenant)
}

// WithStep adds step information to the context
func WithStep(ctx context.Context, step string) context.Context {
	return ContextWithFields(ctx, FieldStep, step)
}

// Context-aware logging functions

// DebugContext logs at debug level with context fields
func DebugContext(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).Debug(msg, args...)
}

// InfoContext logs at info level with context fields
func InfoContext(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).Info(msg, args...)
}

// WarnContext logs at warn level with context fields
func WarnContext(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).Warn(msg, args...)
}

// ErrorContext logs at error level with context fields
func ErrorContext(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).Error(msg, args...)
}
