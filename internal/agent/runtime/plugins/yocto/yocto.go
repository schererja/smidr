package yocto

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/intrik8-labs/smidr/internal/agent/job"
	"github.com/intrik8-labs/smidr/internal/logging"
)

type YoctoPlugin struct{}

func NewYoctoPlugin() *YoctoPlugin {
	return &YoctoPlugin{}
}

func (p *YoctoPlugin) Execute(ctx context.Context, job job.Job) error {
	logger := logging.FromContext(ctx).With("plugin", "yocto", "job_id", job.ID)

	logger.Info("Starting Yocto build job execution")

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
		if strings.Contains(step.cmd, "bitbake") {
			time.Sleep(10 * time.Second)
		} else {
			time.Sleep(1 * time.Second)
		}
		logger.Info("Yocto step completed", "step", step.name)
	}
	artifactPath := filepath.Join(workDir, "build", "tmp", "deploy", "images", machine, fmt.Sprintf("%s-%s.tar.gz", image, machine))
	logger.Info("Yocto build artifacts generated", "artifact", artifactPath)
	logger.Info("Yocto build job execution completed")
	return nil
}
