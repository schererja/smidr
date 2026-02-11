//go:build linux

package signals

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func collectUptime() (float64, error) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, err
	}

	parts := strings.Fields(string(data))
	if len(parts) == 0 {
		return 0, errors.New("empty /proc/uptime")
	}

	uptime, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("parse uptime: %w", err)
	}

	return uptime, nil
}

func collectLoadAverage() (float64, error) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, err
	}

	parts := strings.Fields(string(data))
	if len(parts) == 0 {
		return 0, errors.New("empty /proc/loadavg")
	}

	load, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("parse load: %w", err)
	}

	return load, nil
}

func collectMemoryUsage() (float64, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	defer f.Close()

	var memTotal, memAvailable uint64
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			memTotal = parseMemInfoValue(line)
		} else if strings.HasPrefix(line, "MemAvailable:") {
			memAvailable = parseMemInfoValue(line)
		}
		if memTotal > 0 && memAvailable > 0 {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	if memTotal == 0 {
		return 0, errors.New("MemTotal not found in /proc/meminfo")
	}

	memUsed := memTotal - memAvailable
	pct := (float64(memUsed) / float64(memTotal)) * 100.0
	return pct, nil
}

func parseMemInfoValue(line string) uint64 {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return 0
	}
	val, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return 0
	}
	return val
}

func collectProcessCount() (int, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0, err
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, err := strconv.Atoi(entry.Name()); err == nil {
			count++
		}
	}

	return count, nil
}
