package plugins

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/intrik8-labs/smidr/internal/agent/job"
	"github.com/intrik8-labs/smidr/internal/logging"
)

// BuildPlugin handles build job execution
type BuildPlugin struct{}

// NewBuildPlugin creates a new build plugin instance
func NewBuildPlugin() *BuildPlugin {
	return &BuildPlugin{}
}

// Execute runs a build command
func (p *BuildPlugin) Execute(ctx context.Context, job job.Job) error {
	logger := logging.FromContext(ctx).With("plugin", "build", "job_id", job.ID)

	logger.Info("Starting build job execution")

	// Extract build command from payload
	buildCmd, ok := job.Payload["command"].(string)
	if !ok {
		return fmt.Errorf("build job missing command in payload")
	}

	// Extract working directory
	workDir, ok := job.Payload["workdir"].(string)
	if !ok {
		workDir = "." // default to current directory
	}

	logger.Info("Executing build command", "command", buildCmd, "workdir", workDir)

	// Split command into parts
	cmdParts := strings.Fields(buildCmd)
	if len(cmdParts) == 0 {
		return fmt.Errorf("invalid build command: %s", buildCmd)
	}

	// Create command
	cmd := exec.CommandContext(ctx, cmdParts[0], cmdParts[1:]...)
	cmd.Dir = workDir

	// Execute command
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("Build command failed", "error", err, "output", string(output))
		return fmt.Errorf("build command failed: %w", err)
	}

	logger.Info("Build command completed successfully", "output", string(output))

	// Simulate some processing time
	time.Sleep(2 * time.Second)

	logger.Info("Build job execution completed")
	return nil
}
