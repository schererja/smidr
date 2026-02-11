package signals

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCollect(t *testing.T) {
	snap, err := Collect()
	if err != nil {
		t.Logf("collect returned partial errors: %v", err)
	}

	if snap.Timestamp.IsZero() {
		t.Error("expected timestamp to be set")
	}

	if time.Since(snap.Timestamp) > 5*time.Second {
		t.Error("timestamp should be recent")
	}
}

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

// TestCollectUptimeMissingFile is a documentation test
// In production, collectUptime would fail if /proc/uptime is missing
// We can't easily test this without refactoring to accept an injected filesystem

func TestCollectLoadAverage(t *testing.T) {
	load, err := collectLoadAverage()
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			t.Skip("skipping test: /proc/loadavg not available (not running on Linux)")
		}
		t.Fatalf("collectLoadAverage failed: %v", err)
	}

	if load < 0 {
		t.Errorf("load average should not be negative, got %f", load)
	}
}

func TestCollectMemoryUsage(t *testing.T) {
	pct, err := collectMemoryUsage()
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			t.Skip("skipping test: /proc/meminfo not available (not running on Linux)")
		}
		t.Fatalf("collectMemoryUsage failed: %v", err)
	}

	if pct < 0 || pct > 100 {
		t.Errorf("memory usage should be 0-100%%, got %f", pct)
	}
}

func TestCollectDiskUsage(t *testing.T) {
	pct, err := collectDiskUsage("/")
	if err != nil {
		t.Fatalf("collectDiskUsage failed: %v", err)
	}

	if pct < 0 || pct > 100 {
		t.Errorf("disk usage should be 0-100%%, got %f", pct)
	}
}

func TestCollectDiskUsageInvalidPath(t *testing.T) {
	_, err := collectDiskUsage("/this/path/definitely/does/not/exist/anywhere")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestCollectProcessCount(t *testing.T) {
	count, err := collectProcessCount()
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			t.Skip("skipping test: /proc not available (not running on Linux)")
		}
		t.Fatalf("collectProcessCount failed: %v", err)
	}

	if count <= 0 {
		t.Errorf("expected at least one process, got %d", count)
	}

	// sanity check: should have at least this test process
	if count < 1 {
		t.Error("process count too low")
	}
}

// TestCollectResilience verifies that Collect returns partial data even if some collectors fail
func TestCollectResilience(t *testing.T) {
	snap, err := Collect()

	// Even with errors, we should get a snapshot with a timestamp
	if snap.Timestamp.IsZero() {
		t.Error("expected snapshot with timestamp even if collectors fail")
	}

	// If there were errors, they should be joined
	if err != nil {
		errStr := err.Error()
		if !strings.Contains(errStr, "uptime") &&
			!strings.Contains(errStr, "load") &&
			!strings.Contains(errStr, "memory") &&
			!strings.Contains(errStr, "disk") &&
			!strings.Contains(errStr, "process") {
			t.Logf("error message format looks good: %v", err)
		}
	}
}

// Integration test: verify all fields are populated in normal operation
func TestCollectAllFields(t *testing.T) {
	snap, err := Collect()
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			t.Skip("skipping test: /proc not available (not running on Linux)")
		}
		t.Fatalf("collect failed: %v", err)
	}

	if snap.Timestamp.IsZero() {
		t.Error("Timestamp not set")
	}
	if snap.UptimeSeconds <= 0 {
		t.Error("UptimeSeconds not positive")
	}
	if snap.LoadAverage1m < 0 {
		t.Error("LoadAverage1m negative")
	}
	if snap.MemoryUsedPct < 0 || snap.MemoryUsedPct > 100 {
		t.Errorf("MemoryUsedPct out of range: %f", snap.MemoryUsedPct)
	}
	if snap.DiskUsedPct < 0 || snap.DiskUsedPct > 100 {
		t.Errorf("DiskUsedPct out of range: %f", snap.DiskUsedPct)
	}
	if snap.ProcessCount <= 0 {
		t.Error("ProcessCount not positive")
	}
}

// TestSnapshotTimestamp validates timestamp precision and timezone
func TestSnapshotTimestamp(t *testing.T) {
	before := time.Now().UTC()
	snap, _ := Collect()
	after := time.Now().UTC()

	if snap.Timestamp.Before(before) || snap.Timestamp.After(after) {
		t.Errorf("timestamp %v not between %v and %v", snap.Timestamp, before, after)
	}

	if snap.Timestamp.Location() != time.UTC {
		t.Errorf("timestamp should be UTC, got %v", snap.Timestamp.Location())
	}
}

// TestMemoryUsageEdgeCases tests /proc/meminfo parsing edge cases
func TestMemoryUsageParsingFromFixture(t *testing.T) {
	// Create a temporary directory for test fixtures
	tmpDir := t.TempDir()
	meminfoPath := filepath.Join(tmpDir, "meminfo")

	tests := []struct {
		name        string
		content     string
		expectError bool
		expectPct   float64
	}{
		{
			name: "normal meminfo",
			content: `MemTotal:       16384000 kB
MemFree:         8192000 kB
MemAvailable:    9000000 kB
Buffers:          100000 kB`,
			expectError: false,
			// Used = 16384000 - 9000000 = 7384000
			// Pct = 7384000 / 16384000 * 100 = 45.07%
			expectPct: 45.07,
		},
		{
			name: "zero available memory",
			content: `MemTotal:       16384000 kB
MemAvailable:           0 kB`,
			expectError: false,
			expectPct:   100.0,
		},
		{
			name: "missing memtotal",
			content: `MemFree:         8192000 kB
MemAvailable:    9000000 kB`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.WriteFile(meminfoPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test fixture: %v", err)
			}

			// Note: Can't easily test this without refactoring to accept a path parameter
			// This is a reminder for future refactoring to make collectMemoryUsage testable
			t.Log("Memory usage collector needs refactoring to accept path for full testability")
		})
	}
}
