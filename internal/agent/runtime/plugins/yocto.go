package plugins

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/intrik8-labs/smidr/internal/agent/job"
	"github.com/intrik8-labs/smidr/internal/logging"
)

// YoctoPlugin handles Yocto build job execution
type YoctoPlugin struct{}

// NewYoctoPlugin creates a new Yocto plugin instance
func NewYoctoPlugin() *YoctoPlugin {
	return &YoctoPlugin{}
}

// Execute runs a Yocto build
func (p *YoctoPlugin) Execute(ctx context.Context, job job.Job) error {
	logger := logging.FromContext(ctx).With("plugin", "yocto", "job_id", job.ID)

	logger.Info("Starting Yocto build job execution")

	// Extract Yocto parameters from payload
	image, ok := job.Payload["image"].(string)
	if !ok {
		image = "core-image-minimal" // default image
	}

	machine, ok := job.Payload["machine"].(string)
	if !ok {
		machine = "qemux86-64" // default machine
	}

	workDir, ok := job.Payload["workdir"].(string)
	if !ok {
		return fmt.Errorf("yocto job missing workdir in payload")
	}

	logger.Info("Yocto build parameters", "image", image, "machine", machine, "workdir", workDir)

	// Simulate Yocto build steps
	steps := []struct {
		name string
		cmd  string
	}{
		{"Initialize build environment", fmt.Sprintf("cd %s && source oe-init-build-env build", workDir)},
		{"Configure local.conf", fmt.Sprintf("echo 'MACHINE = \"%s\"' >> %s/build/conf/local.conf", machine, workDir)},
		{"Configure bblayers.conf", fmt.Sprintf("echo 'BBLAYERS += \" %s/meta-custom \"' >> %s/build/conf/bblayers.conf", workDir, workDir)},
		{"Start build", fmt.Sprintf("cd %s/build && bitbake %s", workDir, image)},
	}

	for _, step := range steps {
		logger.Info("Executing Yocto step", "step", step.name)

		// For simulation, we'll just log and sleep instead of running actual commands
		// In a real implementation, you'd execute these commands
		if strings.Contains(step.cmd, "bitbake") {
			// Simulate longer build time for the actual build step
			time.Sleep(10 * time.Second)
		} else {
			time.Sleep(1 * time.Second)
		}

		logger.Info("Yocto step completed", "step", step.name)
	}

	// Simulate artifact generation
	artifactPath := filepath.Join(workDir, "build", "tmp", "deploy", "images", machine, fmt.Sprintf("%s-%s.tar.gz", image, machine))
	logger.Info("Yocto build artifacts generated", "artifact", artifactPath)

	logger.Info("Yocto build job execution completed")
	return nil
}
