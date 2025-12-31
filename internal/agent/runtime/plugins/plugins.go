package plugins

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/intrik8-labs/smidr/internal/agent/job"
	"github.com/intrik8-labs/smidr/internal/logging"
)

// Plugin defines the interface for job execution plugins
type Plugin interface {
	Execute(ctx context.Context, job job.Job) error
}

// BuildPlugin implements the Plugin interface for build jobs
type BuildPlugin struct{}

// NewBuildPlugin creates a new build plugin
func NewBuildPlugin() *BuildPlugin {
	return &BuildPlugin{}
}

// Execute runs the build job
func (p *BuildPlugin) Execute(ctx context.Context, job job.Job) error {
	log := logging.FromContext(ctx)

	log.InfoContext(ctx, "executing build job",
		logging.String("job_id", job.ID),
		logging.String("target", job.Payload["target"].(string)),
	)

	// Simulate build by running a command (e.g., make build)
	cmd := exec.CommandContext(ctx, "echo", "Building target:", job.Payload["target"].(string))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build failed: %w, output: %s", err, string(output))
	}

	log.InfoContext(ctx, "build completed",
		logging.String("job_id", job.ID),
		logging.String("output", string(output)),
	)

	return nil
}

// MonitorPlugin implements the Plugin interface for monitor jobs
type MonitorPlugin struct{}

// NewMonitorPlugin creates a new monitor plugin
func NewMonitorPlugin() *MonitorPlugin {
	return &MonitorPlugin{}
}

// Execute runs the monitor job
func (p *MonitorPlugin) Execute(ctx context.Context, job job.Job) error {
	log := logging.FromContext(ctx)

	log.InfoContext(ctx, "executing monitor job",
		logging.String("job_id", job.ID),
	)

	// Simulate monitoring
	time.Sleep(5 * time.Second)

	log.InfoContext(ctx, "monitor completed",
		logging.String("job_id", job.ID),
	)

	return nil
}
