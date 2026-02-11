//go:build darwin

package signals

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func collectUptime() (float64, error) {
	// Use sysctl to get boot time
	out, err := exec.Command("sysctl", "-n", "kern.boottime").Output()
	if err != nil {
		return 0, fmt.Errorf("sysctl kern.boottime: %w", err)
	}

	// Output format: { sec = 1234567890, usec = 123456 } ...
	str := strings.TrimSpace(string(out))
	secIdx := strings.Index(str, "sec = ")
	if secIdx == -1 {
		return 0, errors.New("failed to parse kern.boottime")
	}

	secStart := secIdx + 6
	secEnd := strings.Index(str[secStart:], ",")
	if secEnd == -1 {
		return 0, errors.New("failed to parse kern.boottime sec")
	}

	bootTimeSec, err := strconv.ParseInt(str[secStart:secStart+secEnd], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse boot time: %w", err)
	}

	bootTime := time.Unix(bootTimeSec, 0)
	uptime := time.Since(bootTime).Seconds()
	return uptime, nil
}

func collectLoadAverage() (float64, error) {
	// Use sysctl to get load average
	out, err := exec.Command("sysctl", "-n", "vm.loadavg").Output()
	if err != nil {
		return 0, fmt.Errorf("sysctl vm.loadavg: %w", err)
	}

	// Output format: { 1.23 4.56 7.89 }
	str := strings.TrimSpace(string(out))
	str = strings.Trim(str, "{}")
	parts := strings.Fields(str)
	if len(parts) == 0 {
		return 0, errors.New("empty vm.loadavg")
	}

	load, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("parse load: %w", err)
	}

	return load, nil
}

func collectMemoryUsage() (float64, error) {
	// Use vm_stat to get memory statistics
	out, err := exec.Command("vm_stat").Output()
	if err != nil {
		return 0, fmt.Errorf("vm_stat: %w", err)
	}

	lines := bytes.Split(out, []byte("\n"))
	stats := make(map[string]uint64)

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		parts := bytes.SplitN(line, []byte(":"), 2)
		if len(parts) != 2 {
			continue
		}
		key := string(bytes.TrimSpace(parts[0]))
		valStr := string(bytes.TrimSpace(bytes.Trim(parts[1], ".")))
		val, err := strconv.ParseUint(valStr, 10, 64)
		if err != nil {
			continue
		}
		stats[key] = val
	}

	// vm_stat reports in pages, get page size
	pageSize := uint64(syscall.Getpagesize())

	// Calculate memory usage
	// Active + Wired + Compressed = Used
	// Free + Inactive + Purgeable = Available (roughly)
	pagesActive := stats["Pages active"]
	pagesWired := stats["Pages wired down"]
	pagesCompressed := stats["Pages occupied by compressor"]

	bytesUsed := (pagesActive + pagesWired + pagesCompressed) * pageSize

	// Get total physical memory
	out, err = exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err != nil {
		return 0, fmt.Errorf("sysctl hw.memsize: %w", err)
	}

	totalBytes, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse hw.memsize: %w", err)
	}

	if totalBytes == 0 {
		return 0, errors.New("total memory is zero")
	}

	pct := (float64(bytesUsed) / float64(totalBytes)) * 100.0
	return pct, nil
}

func collectProcessCount() (int, error) {
	// Use ps to count processes
	out, err := exec.Command("ps", "-A").Output()
	if err != nil {
		return 0, fmt.Errorf("ps -A: %w", err)
	}

	lines := bytes.Split(out, []byte("\n"))
	// Subtract 1 for header line, subtract empty lines
	count := 0
	for i, line := range lines {
		if i == 0 || len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		count++
	}

	return count, nil
}
