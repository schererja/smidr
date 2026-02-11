---
name: "go-build-tags-platform"
description: "Using Go build tags to create platform-specific implementations with shared interfaces"
domain: "architecture"
confidence: "low"
source: "earned"
---

## Context
When implementing functionality that requires platform-specific APIs (Linux `/proc`, macOS `sysctl`, Windows registry), use Go build tags to separate implementations while maintaining a unified interface. This ensures only the relevant code is compiled for each platform, avoiding runtime checks and reducing binary size.

## Patterns

### Split Implementation Across Platform Files
Create one shared file with the interface and one file per platform with the implementation.

**File structure:**
```
signals/
├── signals.go           # Shared types, public API
├── signals_linux.go     # Linux implementation
├── signals_darwin.go    # macOS implementation
└── signals_windows.go   # Windows implementation (if needed)
```

**signals.go** (shared):
```go
package signals

import "time"

// Snapshot represents collected system data
type Snapshot struct {
    Timestamp     time.Time
    UptimeSeconds float64
    LoadAverage1m float64
    MemoryUsedPct float64
    ProcessCount  int
}

// Collect gathers system data (implementation is platform-specific)
func Collect() (Snapshot, error) {
    // Platform implementations provide this via signals_*.go
    return collectPlatform()
}
```

**signals_linux.go** (Linux-specific):
```go
//go:build linux

package signals

import (
    "os"
    "strconv"
    "strings"
)

func collectPlatform() (Snapshot, error) {
    // Read from /proc filesystem
    data, err := os.ReadFile("/proc/uptime")
    if err != nil {
        return Snapshot{}, err
    }
    
    parts := strings.Fields(string(data))
    uptime, _ := strconv.ParseFloat(parts[0], 64)
    
    return Snapshot{
        UptimeSeconds: uptime,
        // ... other fields
    }, nil
}
```

**signals_darwin.go** (macOS-specific):
```go
//go:build darwin

package signals

import (
    "os/exec"
    "strconv"
    "strings"
)

func collectPlatform() (Snapshot, error) {
    // Use sysctl command
    out, err := exec.Command("sysctl", "-n", "kern.boottime").Output()
    if err != nil {
        return Snapshot{}, err
    }
    
    // Parse sysctl output...
    uptime := parseBootTime(string(out))
    
    return Snapshot{
        UptimeSeconds: uptime,
        // ... other fields
    }, nil
}
```

### Keep Function Signatures Identical
All platform implementations must have identical function signatures. The Go compiler enforces this at compile time.

```go
// Both signatures must match exactly
// signals_linux.go
func collectUptime() (float64, error)

// signals_darwin.go  
func collectUptime() (float64, error)
```

**Why:** Ensures the interface contract is respected. If signatures diverge, the code won't compile.

### Use Negative Build Tags for Fallback Implementations
For unsupported platforms, provide a stub implementation with a clear error.

```go
//go:build !linux && !darwin

package signals

import "errors"

func collectPlatform() (Snapshot, error) {
    return Snapshot{}, errors.New("signal collection not supported on this platform")
}
```

### Avoid Runtime GOOS Checks in Platform-Specific Code
Don't use `if runtime.GOOS == "linux"` when build tags are appropriate.

❌ **Bad** (runtime check):
```go
func Collect() (Snapshot, error) {
    if runtime.GOOS == "linux" {
        return collectLinux()
    } else if runtime.GOOS == "darwin" {
        return collectDarwin()
    }
    return Snapshot{}, errors.New("unsupported")
}
```

✅ **Good** (build tags):
```go
// signals_linux.go
//go:build linux
func collectPlatform() (Snapshot, error) { /* ... */ }

// signals_darwin.go
//go:build darwin
func collectPlatform() (Snapshot, error) { /* ... */ }
```

**Benefits of build tags:**
- Compile-time validation (can't accidentally call Linux code on macOS)
- Smaller binaries (unused platform code isn't included)
- No runtime overhead from platform checks
- Clearer separation of concerns

### Share Helper Functions Across Platforms When Possible
If a helper function works on multiple platforms, put it in the shared file or create a separate file with appropriate build tags.

**Example:** `syscall.Statfs` works on both Linux and macOS:
```go
// signals.go (shared)
func collectDiskUsage(path string) (float64, error) {
    var stat syscall.Statfs_t
    if err := syscall.Statfs(path, &stat); err != nil {
        return 0, err
    }
    
    totalBlocks := stat.Blocks
    availBlocks := stat.Bavail
    usedBlocks := totalBlocks - availBlocks
    
    return (float64(usedBlocks) / float64(totalBlocks)) * 100.0, nil
}
```

### Document Platform Differences in Comments
Explain what each platform implementation does differently.

```go
//go:build darwin

package signals

// macOS-specific signal collection using:
// - sysctl kern.boottime for uptime
// - sysctl vm.loadavg for load average  
// - vm_stat for memory statistics
// - ps -A for process count
```

## Anti-Patterns

- **Runtime GOOS checks for platform features** — Use build tags instead for better compile-time safety
- **Inconsistent function signatures** — Won't compile if signatures differ across platform files
- **Platform-specific logic in shared files** — Keep platform details in `_platform.go` files
- **Forgetting the build tag comment** — Without `//go:build`, the file compiles on all platforms

## When to Apply

Use build tags when:
- Platform-specific APIs are required (filesystem paths, syscalls, commands)
- The logic differs significantly between platforms
- You want compile-time enforcement of platform support
- Binary size matters (exclude unused platform code)

Don't use build tags when:
- Minor differences can be handled with simple if statements
- The same standard library API works everywhere (even if behavior differs slightly)
- You're just checking for feature availability (use runtime checks instead)

## References

- Implementation: `agent/internal/signals/signals.go`, `signals_linux.go`, `signals_darwin.go`
- Go build constraints: https://pkg.go.dev/cmd/go#hdr-Build_constraints
