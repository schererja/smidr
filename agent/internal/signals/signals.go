package signals

import (
	"errors"
	"fmt"
	"syscall"
	"time"
)

// Snapshot represents all collected system signals at a point in time.
type Snapshot struct {
	Timestamp    time.Time `json:"timestamp"`
	UptimeSeconds float64   `json:"uptimeSeconds"`
	LoadAverage1m float64   `json:"loadAverage1m"`
	MemoryUsedPct float64   `json:"memoryUsedPct"`
	DiskUsedPct   float64   `json:"diskUsedPct"`
	ProcessCount  int       `json:"processCount"`
}

// Collect gathers all system signals and returns a snapshot.
func Collect() (Snapshot, error) {
	snap := Snapshot{
		Timestamp: time.Now().UTC(),
	}

	var errs []error

	uptime, err := collectUptime()
	if err != nil {
		errs = append(errs, fmt.Errorf("uptime: %w", err))
	} else {
		snap.UptimeSeconds = uptime
	}

	load, err := collectLoadAverage()
	if err != nil {
		errs = append(errs, fmt.Errorf("load: %w", err))
	} else {
		snap.LoadAverage1m = load
	}

	memPct, err := collectMemoryUsage()
	if err != nil {
		errs = append(errs, fmt.Errorf("memory: %w", err))
	} else {
		snap.MemoryUsedPct = memPct
	}

	diskPct, err := collectDiskUsage("/")
	if err != nil {
		errs = append(errs, fmt.Errorf("disk: %w", err))
	} else {
		snap.DiskUsedPct = diskPct
	}

	procCount, err := collectProcessCount()
	if err != nil {
		errs = append(errs, fmt.Errorf("process: %w", err))
	} else {
		snap.ProcessCount = procCount
	}

	if len(errs) > 0 {
		return snap, errors.Join(errs...)
	}

	return snap, nil
}

// collectDiskUsage computes disk usage percentage for the given path using statfs.
func collectDiskUsage(path string) (float64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, fmt.Errorf("statfs %s: %w", path, err)
	}

	totalBlocks := stat.Blocks
	availBlocks := stat.Bavail
	usedBlocks := totalBlocks - availBlocks

	if totalBlocks == 0 {
		return 0, errors.New("total blocks is zero")
	}

	pct := (float64(usedBlocks) / float64(totalBlocks)) * 100.0
	return pct, nil
}
