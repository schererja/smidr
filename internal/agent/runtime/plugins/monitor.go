package plugins

import (
	"context"
	"fmt"
	"time"

	"github.com/intrik8-labs/smidr/internal/agent/job"
	"github.com/intrik8-labs/smidr/internal/logging"
)

// MonitorPlugin handles monitoring job execution
type MonitorPlugin struct{}

// NewMonitorPlugin creates a new monitor plugin instance
func NewMonitorPlugin() *MonitorPlugin {
	return &MonitorPlugin{}
}

// Execute runs a monitoring task
func (p *MonitorPlugin) Execute(ctx context.Context, job job.Job) error {
	logger := logging.FromContext(ctx).With("plugin", "monitor", "job_id", job.ID)

	logger.Info("Starting monitor job execution")

	// Extract monitoring parameters from payload
	durationStr, ok := job.Payload["duration"].(string)
	if !ok {
		durationStr = "30s" // default duration
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return fmt.Errorf("invalid duration format: %s", durationStr)
	}

	target, ok := job.Payload["target"].(string)
	if !ok {
		target = "system" // default target
	}

	logger.Info("Monitoring target", "target", target, "duration", duration)

	// Simulate monitoring for the specified duration
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	endTime := time.Now().Add(duration)
	checkCount := 0

	for time.Now().Before(endTime) {
		select {
		case <-ctx.Done():
			logger.Info("Monitor job cancelled")
			return ctx.Err()
		case <-ticker.C:
			checkCount++
			logger.Info("Monitor check", "check", checkCount, "target", target, "status", "healthy")
		}
	}

	logger.Info("Monitor job execution completed", "checks_performed", checkCount)
	return nil
}
