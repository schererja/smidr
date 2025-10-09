package bitbake

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/intrik8-labs/smidr/internal/config"
	"github.com/intrik8-labs/smidr/internal/container"
)

// BuildExecutor handles bitbake build execution
type BuildExecutor struct {
	config       *config.Config
	containerMgr container.ContainerManager
	containerID  string
	workspaceDir string
}

// NewBuildExecutor creates a new build executor
func NewBuildExecutor(cfg *config.Config, containerMgr container.ContainerManager, containerID string, workspaceDir string) *BuildExecutor {
	return &BuildExecutor{
		config:       cfg,
		containerMgr: containerMgr,
		containerID:  containerID,
		workspaceDir: workspaceDir,
	}
}

// BuildResult contains the results of a build execution
type BuildResult struct {
	Success  bool
	ExitCode int
	Duration time.Duration
	Output   string
	Error    string
}

// ExecuteBuild runs the complete bitbake build process
func (e *BuildExecutor) ExecuteBuild(ctx context.Context) (*BuildResult, error) {
	startTime := time.Now()

	fmt.Println("🔧 Setting up Yocto build environment...")

	// Step 1: Generate bitbake configuration files
	if err := e.setupBuildEnvironment(ctx); err != nil {
		return &BuildResult{
			Success:  false,
			Duration: time.Since(startTime),
			Error:    fmt.Sprintf("Failed to setup build environment: %v", err),
		}, err
	}

	// Step 2: Source the build environment
	if err := e.sourceEnvironment(ctx); err != nil {
		return &BuildResult{
			Success:  false,
			Duration: time.Since(startTime),
			Error:    fmt.Sprintf("Failed to source build environment: %v", err),
		}, err
	}

	// Step 3: Execute the bitbake build
	fmt.Printf("🚀 Starting bitbake build: %s\n", e.config.Build.Image)

	buildResult, err := e.executeBitbake(ctx)
	buildResult.Duration = time.Since(startTime)

	if err != nil {
		buildResult.Error = err.Error()
		return buildResult, err
	}

	fmt.Printf("✅ Build completed successfully in %v\n", buildResult.Duration)
	return buildResult, nil
}

// setupBuildEnvironment generates the necessary configuration files
func (e *BuildExecutor) setupBuildEnvironment(ctx context.Context) error {
	// Generate local.conf content
	localConfContent := e.generateLocalConfContent()

	// Write local.conf to container
	writeLocalConfCmd := []string{"sh", "-c", fmt.Sprintf("mkdir -p /home/builder/work/conf && cat > /home/builder/work/conf/local.conf << 'EOF'\n%s\nEOF", localConfContent)}

	result, err := e.containerMgr.Exec(ctx, e.containerID, writeLocalConfCmd, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to write local.conf: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("failed to write local.conf: %s", string(result.Stderr))
	}

	// Generate bblayers.conf content
	bblayersContent := e.generateBBLayersConfContent()

	// Write bblayers.conf to container
	writeBBLayersCmd := []string{"sh", "-c", fmt.Sprintf("cat > /home/builder/work/conf/bblayers.conf << 'EOF'\n%s\nEOF", bblayersContent)}

	result, err = e.containerMgr.Exec(ctx, e.containerID, writeBBLayersCmd, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to write bblayers.conf: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("failed to write bblayers.conf: %s", string(result.Stderr))
	}

	return nil
}

// sourceEnvironment sources the Yocto build environment
func (e *BuildExecutor) sourceEnvironment(ctx context.Context) error {
	// This step is typically handled by the bitbake command itself in most setups
	// The oe-init-build-env script is usually sourced as part of the build process
	// For now, we'll verify that the required tools are available

	checkCmd := []string{"sh", "-c", "which bitbake && echo 'Build environment ready'"}
	result, err := e.containerMgr.Exec(ctx, e.containerID, checkCmd, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to check build environment: %w", err)
	}
	if result.ExitCode != 0 {
		// Try to source the environment
		sourceCmd := []string{"sh", "-c", "cd /home/builder/work && source /sources/poky/oe-init-build-env . && which bitbake"}
		result, err = e.containerMgr.Exec(ctx, e.containerID, sourceCmd, 30*time.Second)
		if err != nil {
			return fmt.Errorf("failed to source build environment: %w", err)
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("build environment not available: %s", string(result.Stderr))
		}
	}

	return nil
}

// executeBitbake runs the actual bitbake command
func (e *BuildExecutor) executeBitbake(ctx context.Context) (*BuildResult, error) {
	// Construct the bitbake command
	imageName := e.config.Build.Image
	if imageName == "" {
		imageName = "core-image-minimal" // default fallback
	}

	// Build the command with proper environment sourcing
	bitbakeCmd := fmt.Sprintf("cd /home/builder/work && source /sources/poky/oe-init-build-env . && bitbake %s", imageName)

	cmd := []string{"sh", "-c", bitbakeCmd}

	// Execute with a longer timeout for builds (default 2 hours)
	timeout := 2 * time.Hour
	if e.config.Advanced.BuildTimeout > 0 {
		timeout = time.Duration(e.config.Advanced.BuildTimeout) * time.Minute
	}

	result, err := e.containerMgr.Exec(ctx, e.containerID, cmd, timeout)

	buildResult := &BuildResult{
		Success:  result.ExitCode == 0,
		ExitCode: result.ExitCode,
		Output:   string(result.Stdout),
		Error:    string(result.Stderr),
	}

	if err != nil {
		buildResult.Success = false
		return buildResult, fmt.Errorf("bitbake execution failed: %w", err)
	}

	if result.ExitCode != 0 {
		buildResult.Success = false
		return buildResult, fmt.Errorf("bitbake build failed with exit code %d", result.ExitCode)
	}

	return buildResult, nil
}

// generateLocalConfContent creates the content for local.conf
func (e *BuildExecutor) generateLocalConfContent() string {
	var content strings.Builder

	content.WriteString("# Generated by smidr\n")
	content.WriteString("# This file is automatically generated. Do not edit manually.\n\n")

	// Basic configuration
	if e.config.Build.Machine != "" {
		content.WriteString(fmt.Sprintf("MACHINE = \"%s\"\n", e.config.Build.Machine))
	} else if e.config.Base.Machine != "" {
		content.WriteString(fmt.Sprintf("MACHINE = \"%s\"\n", e.config.Base.Machine))
	}

	if e.config.Base.Distro != "" {
		content.WriteString(fmt.Sprintf("DISTRO = \"%s\"\n", e.config.Base.Distro))
	}

	// Parallel build settings
	if e.config.Build.ParallelMake > 0 {
		content.WriteString(fmt.Sprintf("PARALLEL_MAKE = \"-j %d\"\n", e.config.Build.ParallelMake))
	}

	if e.config.Build.BBNumberThreads > 0 {
		content.WriteString(fmt.Sprintf("BB_NUMBER_THREADS = \"%d\"\n", e.config.Build.BBNumberThreads))
	}

	// Directory settings
	content.WriteString("DL_DIR = \"/home/builder/downloads\"\n")
	content.WriteString("SSTATE_DIR = \"/home/builder/sstate-cache\"\n")

	// Package management
	if e.config.Build.PackageClasses != "" {
		content.WriteString(fmt.Sprintf("PACKAGE_CLASSES = \"%s\"\n", e.config.Build.PackageClasses))
	} else {
		content.WriteString("PACKAGE_CLASSES = \"package_rpm\"\n")
	}

	// Extra image features
	if e.config.Build.ExtraImageFeatures != "" {
		content.WriteString(fmt.Sprintf("EXTRA_IMAGE_FEATURES = \"%s\"\n", e.config.Build.ExtraImageFeatures))
	}

	// Add extra packages if specified
	if len(e.config.Build.ExtraPackages) > 0 {
		packages := strings.Join(e.config.Build.ExtraPackages, " ")
		content.WriteString(fmt.Sprintf("IMAGE_INSTALL_append = \" %s\"\n", packages))
	}

	// Standard settings
	content.WriteString("\n# Standard settings\n")
	content.WriteString("CONF_VERSION = \"2\"\n")
	content.WriteString("USER_CLASSES ?= \"buildstats\"\n")
	content.WriteString("PATCHRESOLVE = \"noop\"\n")

	return content.String()
}

// generateBBLayersConfContent creates the content for bblayers.conf
func (e *BuildExecutor) generateBBLayersConfContent() string {
	var content strings.Builder

	content.WriteString("# Generated by smidr\n")
	content.WriteString("# This file is automatically generated. Do not edit manually.\n\n")

	content.WriteString("LCONF_VERSION = \"7\"\n\n")
	content.WriteString("BBPATH = \"${TOPDIR}\"\n")
	content.WriteString("BBFILES ?= \"\"\n\n")

	content.WriteString("BBLAYERS ?= \" \\\n")

	// Add core layers first
	content.WriteString("  /sources/poky/meta \\\n")
	content.WriteString("  /sources/poky/meta-poky \\\n")
	content.WriteString("  /sources/poky/meta-yocto-bsp \\\n")

	// Add configured layers
	for i := range e.config.Layers {
		layerPath := fmt.Sprintf("/home/builder/layers/layer-%d", i)
		content.WriteString(fmt.Sprintf("  %s \\\n", layerPath))
	}

	content.WriteString("  \"\n")

	return content.String()
}
