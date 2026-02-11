---
name: "go-stdlib-testing"
description: "Comprehensive testing with Go's standard library — no external test frameworks"
domain: "testing"
confidence: "low"
source: "earned"
---

## Context
Go's standard library `testing` package is sufficient for comprehensive unit testing without external frameworks like testify, ginkgo, or gomega. This skill documents patterns for writing complete, maintainable tests using only `testing` and standard library packages.

## Patterns

### Table-Driven Tests with Subtests
The idiomatic Go pattern for testing multiple cases efficiently.

```go
func TestParseConfig(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    Config
        wantErr bool
    }{
        {
            name:  "valid config",
            input: "key: value",
            want:  Config{Key: "value"},
            wantErr: false,
        },
        {
            name:  "invalid syntax",
            input: "malformed",
            want:  Config{},
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseConfig(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseConfig() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("ParseConfig() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

**Benefits:**
- Clear test organization with descriptive names
- Easy to add new cases
- Parallel execution with `t.Parallel()` if needed
- Go test runner shows which cases pass/fail

### Testing with Temporary Directories
Use `t.TempDir()` for isolated filesystem tests (Go 1.15+).

```go
func TestWriteFile(t *testing.T) {
    dir := t.TempDir()  // Automatically cleaned up
    path := filepath.Join(dir, "test.txt")
    
    err := WriteFile(path, "content")
    if err != nil {
        t.Fatalf("WriteFile() failed: %v", err)
    }
    
    data, err := os.ReadFile(path)
    if err != nil {
        t.Fatalf("failed to read file: %v", err)
    }
    
    if string(data) != "content" {
        t.Errorf("got %q, want %q", data, "content")
    }
}
```

### Error Checking Without Assertion Libraries
Standard library patterns are sufficient for error validation.

```go
// Test error is returned
if err == nil {
    t.Error("expected error, got nil")
}

// Test error contains keyword
if err != nil && !strings.Contains(err.Error(), "required") {
    t.Errorf("expected 'required' in error, got: %v", err)
}

// Test error is specific type
var pathErr *os.PathError
if !errors.As(err, &pathErr) {
    t.Errorf("expected PathError, got %T", err)
}

// Test error wraps another
if !errors.Is(err, os.ErrNotExist) {
    t.Error("expected ErrNotExist wrapped in error")
}
```

### Helper Functions for Test Setup
Extract common setup into helper functions.

```go
func createTestConfig(t *testing.T, content string) string {
    t.Helper()  // Mark as helper for better error reporting
    
    dir := t.TempDir()
    path := filepath.Join(dir, "config.yaml")
    
    if err := os.WriteFile(path, []byte(content), 0600); err != nil {
        t.Fatalf("failed to create test config: %v", err)
    }
    
    return path
}

func TestLoadConfig(t *testing.T) {
    path := createTestConfig(t, "key: value")
    cfg, err := LoadConfig(path)
    // ... test assertions
}
```

### Comparing Complex Structures
Without testify's `assert.Equal`, use comparison helpers.

```go
func TestSnapshot(t *testing.T) {
    got := Collect()
    want := Snapshot{
        Timestamp: time.Now().UTC(),
        Uptime: 123.45,
    }
    
    // For simple types
    if got.Uptime != want.Uptime {
        t.Errorf("Uptime = %f, want %f", got.Uptime, want.Uptime)
    }
    
    // For time with tolerance
    if got.Timestamp.Sub(want.Timestamp).Abs() > time.Second {
        t.Errorf("Timestamp = %v, want %v (±1s)", got.Timestamp, want.Timestamp)
    }
    
    // For slices
    if len(got.Values) != len(want.Values) {
        t.Fatalf("got %d values, want %d", len(got.Values), len(want.Values))
    }
    for i := range got.Values {
        if got.Values[i] != want.Values[i] {
            t.Errorf("Values[%d] = %v, want %v", i, got.Values[i], want.Values[i])
        }
    }
}
```

### Skipping Tests Conditionally
Skip tests that require specific conditions.

```go
func TestLinuxOnly(t *testing.T) {
    if runtime.GOOS != "linux" {
        t.Skip("skipping: test requires Linux")
    }
    // ... test logic
}

func TestRequiresDocker(t *testing.T) {
    if _, err := exec.LookPath("docker"); err != nil {
        t.Skip("skipping: docker not found")
    }
    // ... test logic
}
```

### Testing Concurrent Code
Use channels and goroutines without external frameworks.

```go
func TestConcurrentAccess(t *testing.T) {
    cache := NewCache()
    
    var wg sync.WaitGroup
    errors := make(chan error, 10)
    
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            if err := cache.Set(id, "value"); err != nil {
                errors <- err
            }
        }(i)
    }
    
    wg.Wait()
    close(errors)
    
    for err := range errors {
        t.Errorf("concurrent Set() failed: %v", err)
    }
}
```

### Benchmarks
Standard library includes benchmarking support.

```go
func BenchmarkParseConfig(b *testing.B) {
    input := "key: value\nfoo: bar"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = ParseConfig(input)
    }
}

// Run with: go test -bench=. -benchmem
```

## Anti-Patterns

- **External assertion libraries** — Go's standard patterns are sufficient and more explicit.
- **BDD-style frameworks** — Go's table-driven tests are more idiomatic and readable.
- **Complex test harnesses** — Keep setup/teardown simple with helpers and `t.Cleanup()`.
- **Over-mocking** — Prefer testing against real implementations or simple fakes over heavy mocking frameworks.

## When to Apply

Use stdlib-only testing when:
- Building applications or libraries where minimizing dependencies is important
- Writing tests for system-level code (filesystems, processes, networking)
- Teaching others Go testing patterns
- Contributing to projects with strict dependency policies

## Trade-offs

**Advantages:**
- Zero test dependencies (faster builds, simpler maintenance)
- Tests are pure Go (no DSL to learn)
- Explicit error checking (no magic assertions)
- Compatible with all Go tooling

**Disadvantages:**
- More verbose than assertion libraries (but explicit)
- No built-in HTTP mocking (use `httptest` from stdlib)
- Manual comparison of complex structures

## References

- Test files: `agent/internal/signals/signals_test.go`
- Test files: `agent/internal/agent/config_test.go`
- Test files: `agent/internal/agent/certificates_test.go`
- Test files: `agent/internal/agent/uuid_test.go`
- Official docs: https://pkg.go.dev/testing
