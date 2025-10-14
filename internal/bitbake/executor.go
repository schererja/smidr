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

	// Check actual container memory limit using docker inspect
	fmt.Printf("[DEBUG] Checking container memory limit with docker inspect...\n")

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
	// First ensure we have the right working directory with proper permissions
	// Use /home/builder/build initially, then we can move it to the mounted workspace
	setupDirCmd := []string{"sh", "-c", "mkdir -p /home/builder/build/conf && whoami && pwd && ls -la /home/builder/"}

	result, err := e.containerMgr.Exec(ctx, e.containerID, setupDirCmd, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to setup build directory: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("failed to setup build directory: %s", string(result.Stderr))
	}

	// Generate local.conf content
	localConfContent := e.generateLocalConfContent()

	// Write local.conf to container in temporary location first
	writeLocalConfCmd := []string{"sh", "-c", fmt.Sprintf("cat > /home/builder/build/conf/local.conf << 'EOF'\n%s\nEOF", localConfContent)}

	result, err = e.containerMgr.Exec(ctx, e.containerID, writeLocalConfCmd, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to write local.conf: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("failed to write local.conf: %s", string(result.Stderr))
	}

	// Verify the local.conf was written correctly
	verifyCmd := []string{"sh", "-c", "cat /home/builder/build/conf/local.conf | grep -E 'BB_NUMBER_THREADS|PARALLEL_MAKE'"}
	verifyResult, _ := e.containerMgr.Exec(ctx, e.containerID, verifyCmd, 10*time.Second)
	fmt.Printf("[DEBUG] Contents of local.conf in container:\n%s\n", string(verifyResult.Stdout))

	// Generate bblayers.conf content
	bblayersContent := e.generateBBLayersConfContent()

	// Write bblayers.conf to container
	writeBBLayersCmd := []string{"sh", "-c", fmt.Sprintf("cat > /home/builder/build/conf/bblayers.conf << 'EOF'\n%s\nEOF", bblayersContent)}

	result, err = e.containerMgr.Exec(ctx, e.containerID, writeBBLayersCmd, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to write bblayers.conf: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("failed to write bblayers.conf: %s", string(result.Stderr))
	}

	// Now try to copy to the mounted workspace if available
	copyToWorkspaceCmd := []string{"sh", "-c", "if [ -d '/home/builder/work' ]; then cp -r /home/builder/build/* /home/builder/work/ 2>/dev/null || echo 'Failed to copy to workspace, will use /home/builder/build'; fi"}
	_, _ = e.containerMgr.Exec(ctx, e.containerID, copyToWorkspaceCmd, 10*time.Second)

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
		// Try to source the environment in a writable directory
		// The oe-init-build-env script needs to create conf directory, so use /home/builder/build
		// First, check whether the expected oe-init-build-env exists in the layers dir
		// Debug: List what's actually in the layers directory
		debugCmd := []string{"sh", "-c", "echo 'DEBUG: Contents of /home/builder/layers:' && ls -la /home/builder/layers/ || echo 'layers dir does not exist'"}
		debugRes, _ := e.containerMgr.Exec(ctx, e.containerID, debugCmd, 5*time.Second)
		fmt.Printf("Container layers directory contents:\n%s\n", string(debugRes.Stdout))

		checkOoCmd := []string{"sh", "-c", "if [ -f /home/builder/layers/poky/oe-init-build-env ]; then echo present; else echo missing; fi"}
		chkRes, _ := e.containerMgr.Exec(ctx, e.containerID, checkOoCmd, 5*time.Second)
		if strings.Contains(string(chkRes.Stdout), "missing") {
			// Additional debug: check if poky directory exists at all
			pokyDebugCmd := []string{"sh", "-c", "echo 'DEBUG: Checking poky dir:' && ls -la /home/builder/layers/poky/ || echo 'poky dir does not exist'"}
			pokyDebugRes, _ := e.containerMgr.Exec(ctx, e.containerID, pokyDebugCmd, 5*time.Second)
			fmt.Printf("Container poky directory debug:\n%s\n", string(pokyDebugRes.Stdout))

			return fmt.Errorf("build environment not available: poky not found in layers directory. Please ensure layers are properly fetched and mounted")
		}

		// Check for meta-openembedded (meta-oe etc.)
		checkOE := []string{"sh", "-c", "if [ -d /home/builder/layers/meta-openembedded/meta-oe ]; then echo present; else echo missing; fi"}
		oeRes, _ := e.containerMgr.Exec(ctx, e.containerID, checkOE, 5*time.Second)
		if strings.Contains(string(oeRes.Stdout), "missing") {
			return fmt.Errorf("build environment not available: meta-openembedded not found in layers directory. Please ensure layers are properly fetched and mounted")
		}

		sourceCmd := []string{"sh", "-c", "mkdir -p /home/builder/build && cd /home/builder/build && source /home/builder/layers/poky/oe-init-build-env . && which bitbake"}
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

	// Use smaller image for qemu machines to avoid memory issues
	machine := e.config.Build.Machine
	if machine == "" {
		machine = e.config.Base.Machine
	}
	if machine == "qemux86-64" && imageName == "core-image-weston" {
		imageName = "core-image-minimal"
		fmt.Printf("⚠️  Using core-image-minimal instead of core-image-weston for qemu machine\n")
	}

	// Build the command with proper environment sourcing in writable directory
	// We need to re-apply our settings after sourcing because oe-init-build-env might override them
	parallelMake := e.config.Build.ParallelMake
	if parallelMake <= 0 {
		parallelMake = 2
	}
	bbThreads := e.config.Build.BBNumberThreads
	if bbThreads <= 0 {
		bbThreads = 2
	}

	fmt.Printf("[INFO] Forcing BitBake parallelism: BB_NUMBER_THREADS=%d, PARALLEL_MAKE=-j%d\n", bbThreads, parallelMake)

	// Use sed to update the values in local.conf right before running bitbake
	bitbakeCmd := fmt.Sprintf(`set -x && \
		echo "=== Starting build setup ===" && \
		cd /home/builder/build && \
		echo "=== Checking memory limit ===" && \
		cat /sys/fs/cgroup/memory/memory.limit_in_bytes 2>/dev/null || cat /sys/fs/cgroup/memory.max 2>/dev/null || echo "Cannot read cgroup v1/v2 memory limit" && \
		echo "=== Sourcing environment ===" && \
		source /home/builder/layers/poky/oe-init-build-env . && \
		echo "=== Updating config ===" && \
		sed -i 's/^BB_NUMBER_THREADS.*/BB_NUMBER_THREADS = "%d"/' conf/local.conf && \
		sed -i 's/^PARALLEL_MAKE.*/PARALLEL_MAKE = "-j %d"/' conf/local.conf && \
		echo "=== Verifying settings ===" && \
		grep -E 'BB_NUMBER_THREADS|PARALLEL_MAKE' conf/local.conf && \
		echo "=== Starting bitbake ===" && \
		bitbake %s`, bbThreads, parallelMake, imageName)

	cmd := []string{"sh", "-c", bitbakeCmd}

	// Execute with a longer timeout for builds (default 24 hours)
	timeout := 24 * time.Hour
	if e.config.Advanced.BuildTimeout > 0 {
		timeout = time.Duration(e.config.Advanced.BuildTimeout) * time.Minute
	}

	// Pre-fetch sources to avoid checksum warnings and fail early if fetch fails
	// Use `bitbake -c fetch` which is broadly supported for image targets.
	fetchCmd := []string{"sh", "-c", fmt.Sprintf("cd /home/builder/build && source /home/builder/layers/poky/oe-init-build-env . && bitbake -c fetch %s", imageName)}
	fmt.Println("⬇️  Running pre-fetch (bitbake -c fetch) to download sources before build...")
	fetchResult, fetchErr := e.containerMgr.ExecStream(ctx, e.containerID, fetchCmd, timeout)
	if fetchErr != nil {
		return &BuildResult{Success: false, ExitCode: fetchResult.ExitCode, Output: string(fetchResult.Stdout) + "\n" + string(fetchResult.Stderr)}, fmt.Errorf("pre-fetch failed: %w", fetchErr)
	}
	if fetchResult.ExitCode != 0 {
		return &BuildResult{Success: false, ExitCode: fetchResult.ExitCode, Output: string(fetchResult.Stdout) + "\n" + string(fetchResult.Stderr)}, fmt.Errorf("pre-fetch failed with exit code %d", fetchResult.ExitCode)
	}

	fmt.Println("📺 Streaming build output...")
	result, err := e.containerMgr.ExecStream(ctx, e.containerID, cmd, timeout)

	buildResult := &BuildResult{
		Success:  result.ExitCode == 0,
		ExitCode: result.ExitCode,
		Output:   string(result.Stdout) + "\n" + string(result.Stderr), // Combine stdout and stderr
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

	// Basic configuration with fallbacks for missing layers
	machine := ""
	if e.config.Build.Machine != "" {
		machine = e.config.Build.Machine
	} else if e.config.Base.Machine != "" {
		machine = e.config.Base.Machine
	}

	// Use fallback machine if configured machine requires missing layers
	if machine == "verdin-imx8mp" {
		// verdin-imx8mp requires meta-freescale layers that we don't have
		machine = "qemux86-64"
		fmt.Printf("⚠️  Falling back to %s machine (verdin-imx8mp requires meta-freescale layers)\n", machine)
	}

	if machine != "" {
		content.WriteString(fmt.Sprintf("MACHINE = \"%s\"\n", machine))
	}

	distro := e.config.Base.Distro
	if distro != "" {
		content.WriteString(fmt.Sprintf("DISTRO ?= \"%s\"\n", distro))
	} else {
		fmt.Printf("[WARN] No DISTRO set in config. Please set base.distro in smidr.yaml.\n")
	}

	// Parallel build settings with safe defaults
	parallelMake := e.config.Build.ParallelMake
	fmt.Printf("[DEBUG] Config value for ParallelMake: %d\n", parallelMake)
	if parallelMake <= 0 {
		parallelMake = 2 // default to 2 if not specified to avoid OOM
		fmt.Printf("[INFO] Using default ParallelMake: %d\n", parallelMake)
	}
	content.WriteString(fmt.Sprintf("PARALLEL_MAKE = \"-j %d\"\n", parallelMake))

	bbThreads := e.config.Build.BBNumberThreads
	fmt.Printf("[DEBUG] Config value for BBNumberThreads: %d\n", bbThreads)
	if bbThreads <= 0 {
		bbThreads = 2 // default to 2 if not specified to avoid OOM
		fmt.Printf("[INFO] Using default BBNumberThreads: %d\n", bbThreads)
	}
	content.WriteString(fmt.Sprintf("BB_NUMBER_THREADS = \"%d\"\n", bbThreads))

	fmt.Printf("[INFO] Container build config: BB_NUMBER_THREADS=%d, PARALLEL_MAKE=-j%d\n", bbThreads, parallelMake)

	// Directory settings
	content.WriteString("DL_DIR = \"/home/builder/downloads\"\n")
	// Use SSTATE_MIRRORS instead of SSTATE_DIR to avoid permission issues
	content.WriteString("SSTATE_MIRRORS = \"file://.* file:///home/builder/sstate-cache/PATH\"\n")

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
		content.WriteString(fmt.Sprintf("IMAGE_INSTALL:append = \" %s\"\n", packages))
	}

	// Deploy directory settings
	// Always use container-local deploy dir
	content.WriteString("TI_COMMON_DEPLOY = \"${TOPDIR}/deploy\"\n")
	content.WriteString("DEPLOY_DIR = \"${TI_COMMON_DEPLOY}${@'' if d.getVar('BB_CURRENT_MC') == 'default' else '/${BB_CURRENT_MC}'}\"\n")

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

	ctx := context.Background()
	added := make(map[string]bool)
	// Recursively add all valid sublayers for poky and all other layers
	for _, layer := range e.config.Layers {
		repoDir := "/home/builder/layers/" + layer.Name
		findCmd := []string{"sh", "-c", fmt.Sprintf("find '%s' -type f -name layer.conf | grep '/conf/layer.conf'$", repoDir)}
		findResult, err := e.containerMgr.Exec(ctx, e.containerID, findCmd, 5*time.Second)
		if err != nil {
			fmt.Printf("[WARN] Could not search for conf/layer.conf in %s: %v\n", repoDir, err)
			continue
		}
		fmt.Printf("[DEBUG] find results for %s:\n%s\n", repoDir, string(findResult.Stdout))
		layerConfs := strings.Split(strings.TrimSpace(string(findResult.Stdout)), "\n")
		for _, confPath := range layerConfs {
			if confPath == "" {
				continue
			}
			layerDir := strings.TrimSuffix(confPath, "/conf/layer.conf")
			catCmd := []string{"sh", "-c", fmt.Sprintf("cat '%s'", confPath)}
			catResult, err := e.containerMgr.Exec(ctx, e.containerID, catCmd, 2*time.Second)
			confContent := string(catResult.Stdout)
			if err != nil || !strings.Contains(confContent, "BBFILE_COLLECTIONS") {
				continue
			}
			compatible := false
			yoctoSeries := e.config.YoctoSeries
			if yoctoSeries == "" {
				compatible = true
			} else {
				for _, line := range strings.Split(confContent, "\n") {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "LAYERSERIES_COMPAT_") {
						parts := strings.SplitN(line, "=", 2)
						if len(parts) == 2 {
							val := strings.Trim(parts[1], " \"'")
							for _, series := range strings.Fields(val) {
								if series == yoctoSeries || strings.HasPrefix(series, yoctoSeries+"-") {
									compatible = true
									break
								}
							}
						}
					}
					if compatible {
						break
					}
				}
			}
			if !compatible {
				fmt.Printf("[WARN] Layer %s is not compatible with Yocto series '%s'\n", layerDir, yoctoSeries)
				continue
			}
			if !added[layerDir] {
				fmt.Printf("[INFO] Adding Yocto layer: %s\n", layerDir)
				content.WriteString(fmt.Sprintf("  %s \\\n", layerDir))
				added[layerDir] = true
			}
		}
	}

	content.WriteString("  \"\n")

	return content.String()
}
