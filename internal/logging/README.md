# Smidr Logging Package

Structured, context-aware logging for Smidr using Go's standard `log/slog` package.

## Design Principles

1. **Structured, machine-readable logs** - JSON output for database ingestion
2. **Context-aware** - Correlation IDs flow through the call stack
3. **Consistent** - Same format across all binaries
4. **Boring and predictable** - Standard library, no magic

## Quick Start

### Initialize the Logger

Each binary should initialize the logger once at startup:

```go
logging.Init(logging.Config{
    Service:   "smidr-agent",     // Binary name
    Component: "executor",         // Component within binary
    Level:     logging.LevelInfo,
    JSON:      true,              // Always true for production
})
```

### Context-Based Logging (Recommended Pattern)

**Never pass loggers manually. Context is the contract.**

```go
// Add correlation IDs to context
ctx := logging.WithJob(ctx, jobID)
ctx = logging.WithExecutor(ctx, executorID)
ctx = logging.WithTenant(ctx, "default")

// Get logger from context
log := logging.FromContext(ctx)
log.Info("job started")
```

### Structured Fields

```go
log.Info("step started",
    logging.String("step", "build"),
    logging.String("image", "yocto-builder:latest"),
    logging.Duration("timeout", 30*time.Minute),
)
```

## Canonical Log Schema

Every log entry includes:

### Required Fields (Automatic)

- `time` - Timestamp (UTC, RFC3339)
- `level` - Log severity (DEBUG, INFO, WARN, ERROR)
- `msg` - Human-readable message
- `service` - Binary name (set via Init)
- `component` - Component name (set via Init)

### Correlation IDs (Context-Based)

- `job_id` - Links all logs for a build
- `executor_id` - Traces executor behavior
- `request_id` - API request tracing
- `tenant` - Multi-tenant identifier

### Example Output

```json
{
  "time": "2025-01-22T19:48:00.123Z",
  "level": "INFO",
  "msg": "job started",
  "service": "smidr-agent",
  "component": "executor",
  "job_id": "job-123",
  "executor_id": "exec-abc",
  "tenant": "default",
  "step": "build",
  "image": "yocto-builder:latest"
}
```

## Context Helpers

### Adding Correlation IDs

```go
// Primary correlation identifiers
ctx = logging.WithJob(ctx, "job-123")
ctx = logging.WithExecutor(ctx, "exec-abc")
ctx = logging.WithRequestID(ctx, "req-456")
ctx = logging.WithTenant(ctx, "default")
ctx = logging.WithStep(ctx, "build")

// Generic fields
ctx = logging.ContextWithField(ctx, "key", "value")
ctx = logging.ContextWithFields(ctx, "k1", "v1", "k2", "v2")
```

### Context-Aware Logging

```go
// Get logger from context (recommended)
log := logging.FromContext(ctx)
log.Info("message", "key", "value")

// Or use convenience functions
logging.InfoContext(ctx, "message", "key", "value")
logging.ErrorContext(ctx, "error occurred", logging.Err(err))
```

## Field Constructors

Type-safe field constructors for structured logging:

```go
logging.String(key, value)
logging.Int(key, value)
logging.Int64(key, value)
logging.Uint64(key, value)
logging.Float64(key, value)
logging.Bool(key, value)
logging.Time(key, value)
logging.Duration(key, value)
logging.Any(key, value)
logging.Err(err)           // Creates "error" field
logging.Group(key, ...)    // Group related fields
```

## Log Levels

### Semantic Usage

| Level | Use Case |
|-------|----------|
| DEBUG | Development/debugging only |
| INFO  | State transitions, normal operations |
| WARN  | Recoverable issues, retries |
| ERROR | Job/executor failures |
| FATAL | Process crash (use with care) |

### Setting Levels

```go
// In Init
logging.Init(logging.Config{
    Level: logging.LevelInfo,  // or LevelDebug, LevelWarn, LevelError
})

// Parse from string
level := logging.ParseLevel("INFO")  // Supports: DEBUG, INFO, WARN, WARNING, ERROR
```

## Best Practices

### ✅ DO

```go
// Initialize once per binary
logging.Init(logging.Config{
    Service:   "smidr-agent",
    Component: "executor",
    Level:     logging.LevelInfo,
    JSON:      true,
})

// Use context for correlation
ctx = logging.WithJob(ctx, jobID)
log := logging.FromContext(ctx)
log.Info("job started")

// Log state transitions at INFO
log.Info("job state changed", "from", "pending", "to", "running")

// Use structured fields
log.Info("request completed",
    logging.Duration("duration", elapsed),
    logging.Int("status", 200),
)
```

### ❌ DON'T

```go
// Don't pass loggers as parameters
func ProcessJob(log *logging.Logger, jobID string) { ... }  // ❌

// Don't log secrets
log.Info("connecting", "password", pwd)  // ❌

// Don't log full environment
log.Debug("env", os.Environ())  // ❌

// Don't use text format in production
logging.Init(logging.Config{JSON: false})  // ❌
```

## Integration Examples

### smidr-agent Executor

```go
func (e *Executor) ExecuteJob(ctx context.Context, job *Job) error {
    // Add job context
    ctx = logging.WithJob(ctx, job.ID)
    ctx = logging.WithExecutor(ctx, e.ID)

    log := logging.FromContext(ctx)
    log.Info("job execution started")

    for _, step := range job.Steps {
        ctx := logging.WithStep(ctx, step.Name)
        if err := e.executeStep(ctx, step); err != nil {
            log := logging.FromContext(ctx)
            log.Error("step failed",
                logging.Err(err),
                logging.String("image", step.Image),
            )
            return err
        }
    }

    log.Info("job execution completed",
        logging.Duration("duration", time.Since(job.StartTime)),
    )
    return nil
}
```

### smidr-core API Handler

```go
func (h *Handler) CreateJob(w http.ResponseWriter, r *http.Request) {
    // Add request ID
    requestID := generateRequestID()
    ctx := logging.WithRequestID(r.Context(), requestID)
    ctx = logging.WithTenant(ctx, getTenant(r))

    log := logging.FromContext(ctx)
    log.Info("create job request received")

    job, err := h.scheduler.CreateJob(ctx, parseRequest(r))
    if err != nil {
        log.Error("failed to create job", logging.Err(err))
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    log.Info("job created",
        logging.String("job_id", job.ID),
        logging.String("status", job.Status),
    )

    writeJSON(w, job)
}
```

### smidr-core Scheduler

```go
func (s *Scheduler) ScheduleJob(ctx context.Context, job *Job) error {
    log := logging.FromContext(ctx)

    executor, err := s.findExecutor(ctx, job.Requirements)
    if err != nil {
        log.Warn("no executor found",
            logging.String("reason", err.Error()),
        )
        return err
    }

    log.Info("job scheduled",
        logging.String("executor_id", executor.ID),
        logging.Any("requirements", job.Requirements),
    )

    return s.assignJob(ctx, job, executor)
}
```

## Database Ingestion

This format ingests cleanly into:

- **DynamoDB** - Store as JSON blobs
- **Elasticsearch/OpenSearch** - Direct JSON ingestion
- **ClickHouse** - JSON column type
- **PostgreSQL** - JSONB column
- **Loki** - Label extraction from JSON
- **BigQuery** - Nested/repeated fields

Index by: `job_id + time` or `service + component + time`

## Testing

Use custom output for testing:

```go
func TestLogging(t *testing.T) {
    var buf bytes.Buffer

    logging.Init(logging.Config{
        Service:   "test-service",
        Component: "test",
        Level:     logging.LevelDebug,
        JSON:      true,
        Output:    &buf,
    })

    logging.Info("test message", "key", "value")

    // Parse and validate JSON output
    var logEntry map[string]any
    json.Unmarshal(buf.Bytes(), &logEntry)

    assert.Equal(t, "INFO", logEntry["level"])
    assert.Equal(t, "test-service", logEntry["service"])
}
```

## Migration from Other Loggers

If migrating from another logger:

1. Replace logger initialization with `logging.Init()`
2. Add context parameters to functions
3. Use `logging.FromContext(ctx)` instead of passing loggers
4. Convert log calls to structured format
5. Add correlation IDs via context helpers

## File Structure

```
internal/logging/
├── logging.go       # Core logger, Init, Config
├── context.go       # Context helpers (WithJob, etc.)
├── fields.go        # Field constructors and constants
├── levels.go        # Log level types
├── slog.go          # slog integration utilities
└── example_test.go  # Usage examples
```

## FAQ

**Q: Why slog instead of zap/zerolog?**
A: Standard library stability, no dependencies, first-class JSON support, and ecosystem convergence.

**Q: Can I use text format for local development?**
A: Yes, set `JSON: false` in Config, but JSON is recommended everywhere.

**Q: How do I add custom fields to all logs?**
A: Use `logger.With()` or add fields via context with `ContextWithFields()`.

**Q: Should I log STDOUT/STDERR from jobs?**
A: No. Treat job output as artifacts/streams, not logs. Wrap with job_id and step_id.

**Q: How do I change log level at runtime?**
A: Currently requires re-initialization. Dynamic level changes can be added if needed.

## References

- [Go slog documentation](https://pkg.go.dev/log/slog)
- Smidr Logging Design Doc (original specification)
