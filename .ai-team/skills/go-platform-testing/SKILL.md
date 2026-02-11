---
name: "go-platform-testing"
description: "Platform-aware testing patterns for Go code that uses OS-specific features"
domain: "testing"
confidence: "low"
source: "earned"
---

## Context
When writing Go tests for code that depends on platform-specific features (like Linux's `/proc` filesystem), tests should run successfully on all platforms without requiring build tags. This skill documents patterns for writing tests that skip gracefully on unsupported platforms while still validating behavior on supported ones.

## Patterns

### Skip Platform-Specific Tests via Error Detection
Instead of using build tags (`//go:build linux`), detect platform incompatibility at runtime by checking error messages. This keeps all tests visible and runnable on all platforms.

```go
func TestCollectUptime(t *testing.T) {
    uptime, err := collectUptime()
    if err != nil {
        if strings.Contains(err.Error(), "no such file") {
            t.Skip("skipping test: /proc/uptime not available (not running on Linux)")
        }
        t.Fatalf("collectUptime failed: %v", err)
    }
    
    if uptime <= 0 {
        t.Errorf("expected positive uptime, got %f", uptime)
    }
}
```

**Why this works:**
- Tests run on all platforms (CI can validate syntax/compilation)
- Platform-specific tests skip with clear messages
- No need to maintain separate `_linux_test.go` files
- Test visibility: developers see what *would* run on other platforms

### Test File Permissions on Security-Sensitive Files
Always verify file permissions for keys, certificates, configs, and other sensitive files.

```go
func TestWritePrivateKey(t *testing.T) {
    keyPath := filepath.Join(t.TempDir(), "key.pem")
    
    err := writePrivateKey(keyPath)
    if err != nil {
        t.Fatalf("writePrivateKey failed: %v", err)
    }
    
    info, err := os.Stat(keyPath)
    if err != nil {
        t.Fatalf("failed to stat key: %v", err)
    }
    
    if info.Mode().Perm() != 0600 {
        t.Errorf("key permissions = %o, want 0600", info.Mode().Perm())
    }
}
```

### Use t.TempDir() for Filesystem Isolation
Go 1.15+ provides `t.TempDir()` which creates isolated temporary directories that are automatically cleaned up.

```go
func TestConfigWrite(t *testing.T) {
    tmpDir := t.TempDir()  // Auto-cleaned after test
    configPath := filepath.Join(tmpDir, "config.yaml")
    
    // Write and test...
}
```

**Benefits:**
- No manual cleanup needed
- Tests can't interfere with each other
- Works on all platforms

### Table-Driven Tests for Parsing and Validation
Use subtests with test tables for parsing functions, validators, and formatters.

```go
func TestParseMemInfoValue(t *testing.T) {
    tests := []struct {
        name string
        line string
        want uint64
    }{
        {
            name: "valid meminfo line",
            line: "MemTotal:       16384000 kB",
            want: 16384000,
        },
        {
            name: "invalid format",
            line: "InvalidLine",
            want: 0,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := parseMemInfoValue(tt.line)
            if got != tt.want {
                t.Errorf("got %d, want %d", got, tt.want)
            }
        })
    }
}
```

### Assert Error Keywords, Not Exact Strings
Error messages may change over time or vary by platform. Assert on keywords that convey intent.

```go
func TestLoadConfigEmptyPath(t *testing.T) {
    _, err := LoadConfig("")
    if err == nil {
        t.Error("expected error for empty path")
    }
    // Good: checks for keyword
    if !strings.Contains(err.Error(), "required") {
        t.Errorf("expected 'required' in error, got: %v", err)
    }
    // Bad: brittle exact match
    // if err.Error() != "config path is required" { ... }
}
```

### Test Idempotency for State-Changing Operations
Operations that create files, generate keys, or modify config should be idempotent.

```go
func TestEnsureKeyIdempotent(t *testing.T) {
    tmpDir := t.TempDir()
    keyPath := filepath.Join(tmpDir, "key.pem")
    
    // First call
    err := EnsureKey(keyPath)
    if err != nil {
        t.Fatalf("first call failed: %v", err)
    }
    
    key1, _ := os.ReadFile(keyPath)
    
    // Second call
    err = EnsureKey(keyPath)
    if err != nil {
        t.Fatalf("second call failed: %v", err)
    }
    
    key2, _ := os.ReadFile(keyPath)
    
    // Key should not be regenerated
    if string(key1) != string(key2) {
        t.Error("key was regenerated on second call")
    }
}
```

## Anti-Patterns

- **Build tags for platform tests** — Makes tests invisible on other platforms. Use runtime skips instead.
- **Hardcoded paths** — Use `t.TempDir()` or test fixtures, never `/tmp/test` or `C:\temp\test`.
- **Ignoring file permissions** — Security-sensitive files must be tested for correct permissions.
- **Manual cleanup** — Use `t.TempDir()` and `t.Cleanup()` instead of manual `defer os.RemoveAll()`.
- **Exact error string matching** — Brittle and breaks when error messages change.

## When to Apply

Use these patterns when:
- Testing code that reads `/proc`, `/sys`, or other Linux-specific paths
- Testing code that creates files with security requirements (keys, certs, tokens)
- Writing tests that should run in CI on multiple platforms
- Testing parsers, validators, or other data transformation functions

## References

- Test files: `agent/internal/signals/signals_test.go`
- Test files: `agent/internal/agent/config_test.go`
- Test files: `agent/internal/agent/certificates_test.go`
