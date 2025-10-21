package bitbake

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

// BuildLogWriter allows streaming log output to both plain text and JSONL
type BuildLogWriter struct {
	PlainWriter io.Writer
	JSONLWriter io.Writer
}

// WriteLog writes a log line to both plain and JSONL outputs
func (w *BuildLogWriter) WriteLog(stream, line string) {
	ts := time.Now().Format(time.RFC3339Nano)
	if w.PlainWriter != nil {
		fmt.Fprintf(w.PlainWriter, "%s\n", line)
	}
	if w.JSONLWriter != nil {
		entry := map[string]interface{}{
			"timestamp": ts,
			"stream":    stream,
			"message":   line,
		}
		b, _ := json.Marshal(entry)
		fmt.Fprintf(w.JSONLWriter, "%s\n", b)
	}
}

// ExecuteBuild runs the complete bitbake build process
// ExecuteBuild runs the complete bitbake build process, optionally streaming logs
func (e *BuildExecutor) ExecuteBuild(ctx context.Context, logWriter *BuildLogWriter) (*BuildResult, error) {
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

	buildResult, err := e.executeBitbake(ctx, logWriter)
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

	// Patch: Ensure TMPDIR is set in local.conf to match the mounted tmp dir
	localConfContent := e.generateLocalConfContent()
	// Add TMPDIR override to local.conf if not present
	if !strings.Contains(localConfContent, "TMPDIR") {
		localConfContent += fmt.Sprintf("\nTMPDIR = \"%s\"\n", e.config.Directories.Tmp)
	}

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

		sourceCmd := []string{"bash", "-c", "mkdir -p /home/builder/build && cd /home/builder/build && source /home/builder/layers/poky/oe-init-build-env . && which bitbake"}
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
// executeBitbake runs the actual bitbake command, streaming logs if logWriter is provided
func (e *BuildExecutor) executeBitbake(ctx context.Context, logWriter *BuildLogWriter) (*BuildResult, error) {
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

	cmd := []string{"bash", "-c", bitbakeCmd}

	// Execute with a longer timeout for builds (default 24 hours)
	timeout := 24 * time.Hour
	if e.config.Advanced.BuildTimeout > 0 {
		timeout = time.Duration(e.config.Advanced.BuildTimeout) * time.Minute
	}

	// Pre-fetch sources to avoid checksum warnings and fail early if fetch fails
	// Use `bitbake -c fetch` which is broadly supported for image targets.
	fetchCmd := []string{"bash", "-c", fmt.Sprintf("cd /home/builder/build && source /home/builder/layers/poky/oe-init-build-env . && bitbake -c fetch %s", imageName)}
	fmt.Println("⬇️  Running pre-fetch (bitbake -c fetch) to download sources before build...")
	fetchResult, fetchErr := e.containerMgr.ExecStream(ctx, e.containerID, fetchCmd, timeout)
	if logWriter != nil {
		for _, line := range strings.Split(string(fetchResult.Stdout), "\n") {
			if line != "" {
				logWriter.WriteLog("stdout", line)
			}
		}
		for _, line := range strings.Split(string(fetchResult.Stderr), "\n") {
			if line != "" {
				logWriter.WriteLog("stderr", line)
			}
		}
	}
	if fetchErr != nil {
		return &BuildResult{Success: false, ExitCode: fetchResult.ExitCode, Output: string(fetchResult.Stdout) + "\n" + string(fetchResult.Stderr)}, fmt.Errorf("pre-fetch failed: %w", fetchErr)
	}
	if fetchResult.ExitCode != 0 {
		return &BuildResult{Success: false, ExitCode: fetchResult.ExitCode, Output: string(fetchResult.Stdout) + "\n" + string(fetchResult.Stderr)}, fmt.Errorf("pre-fetch failed with exit code %d", fetchResult.ExitCode)
	}

	fmt.Println("📺 Streaming build output...")
	// Stream build output and write to logWriter if provided
	result, err := e.containerMgr.ExecStream(ctx, e.containerID, cmd, timeout)
	if logWriter != nil {
		for _, line := range strings.Split(string(result.Stdout), "\n") {
			if line != "" {
				logWriter.WriteLog("stdout", line)
			}
		}
		for _, line := range strings.Split(string(result.Stderr), "\n") {
			if line != "" {
				logWriter.WriteLog("stderr", line)
			}
		}
	}

	buildResult := &BuildResult{
		Success:  result.ExitCode == 0,
		ExitCode: result.ExitCode,
		Output:   string(result.Stdout) + "\n" + string(result.Stderr),
		Error:    string(result.Stderr),
	}

	if err != nil {
		buildResult.Success = false
		return buildResult, fmt.Errorf("bitbake execution failed: %w", err)
	}

	if result.ExitCode != 0 {
		buildResult.Success = false
		fmt.Println("🧹 BitBake build failed. Attempting targeted cleanup, cleansstate, and retry...")
		stderrText := string(result.Stderr)
		failedRecipe := extractFailedRecipe(stderrText)
		if failedRecipe != "" {
			// Targeted cleanup for pseudo path mismatch
			if strings.Contains(strings.ToLower(stderrText), "pseudo") && strings.Contains(strings.ToLower(stderrText), "path mismatch") {
				fmt.Printf("🧹 Detected pseudo path mismatch for recipe '%s' — cleaning workdir artifacts before cleansstate...\n", failedRecipe)
				if err := e.targetedCleanup(ctx, failedRecipe, logWriter); err != nil {
					fmt.Printf("[WARN] Targeted cleanup had issues: %v (continuing)\n", err)
				}
			}

			cleansstateCmd := []string{"bash", "-c", fmt.Sprintf("cd /home/builder/build && source /home/builder/layers/poky/oe-init-build-env . && bitbake -c cleansstate %s", failedRecipe)}
			fmt.Printf("🧹 Running bitbake -c cleansstate %s\n", failedRecipe)
			cleanResult, cleanErr := e.containerMgr.ExecStream(ctx, e.containerID, cleansstateCmd, 2*time.Hour)
			if logWriter != nil {
				for _, line := range strings.Split(string(cleanResult.Stdout), "\n") {
					if line != "" {
						logWriter.WriteLog("stdout", line)
					}
				}
				for _, line := range strings.Split(string(cleanResult.Stderr), "\n") {
					if line != "" {
						logWriter.WriteLog("stderr", line)
					}
				}
			}
			if cleanErr != nil || cleanResult.ExitCode != 0 {
				fmt.Printf("[ERROR] cleansstate failed for %s: %v\n", failedRecipe, cleanErr)
				return buildResult, fmt.Errorf("bitbake build failed, cleansstate also failed for %s", failedRecipe)
			}

			// Retry build once
			fmt.Println("🔁 Retrying bitbake build after cleansstate...")
			retryResult, retryErr := e.containerMgr.ExecStream(ctx, e.containerID, cmd, timeout)
			if logWriter != nil {
				for _, line := range strings.Split(string(retryResult.Stdout), "\n") {
					if line != "" {
						logWriter.WriteLog("stdout", line)
					}
				}
				for _, line := range strings.Split(string(retryResult.Stderr), "\n") {
					if line != "" {
						logWriter.WriteLog("stderr", line)
					}
				}
			}
			buildResult.Output += "\n--- Cleansstate and retry output ---\n" + string(retryResult.Stdout) + "\n" + string(retryResult.Stderr)
			if retryErr != nil || retryResult.ExitCode != 0 {
				fmt.Printf("[ERROR] Retry build failed for %s: %v\n", failedRecipe, retryErr)
				return buildResult, fmt.Errorf("bitbake build failed after cleansstate/retry for %s", failedRecipe)
			}
			fmt.Println("✅ Build succeeded after cleansstate and retry.")
			buildResult.Success = true
			buildResult.ExitCode = retryResult.ExitCode
			return buildResult, nil
		} else {
			fmt.Println("[WARN] Could not extract failed recipe for cleansstate. No retry attempted.")
			return buildResult, fmt.Errorf("bitbake build failed with exit code %d", result.ExitCode)
		}
	}

	return buildResult, nil
}

// targetedCleanup removes per-recipe workdir artifacts that commonly cause pseudo path mismatch
// It targets only the failing recipe directories under /home/builder/tmp/work/*/<recipe>/*
func (e *BuildExecutor) targetedCleanup(ctx context.Context, recipe string, logWriter *BuildLogWriter) error {
	if strings.TrimSpace(recipe) == "" {
		return nil
	}
	// Bash script to cleanup common problematic subdirs
	script := fmt.Sprintf(`set -e
echo "=== Targeted cleanup for recipe: %s ==="
found=0
for d in /home/builder/tmp/work/*/%s/*; do
  if [ -d "$d" ]; then
	found=1
	echo "Cleaning $d/{packages-split,sstate-build-package,pseudo}"
	rm -rf "$d/packages-split" "$d/sstate-build-package" "$d/pseudo" || true
  fi
done
if [ "$found" = "0" ]; then
  echo "No workdir found for recipe %s (this may be fine)"
fi
`, recipe, recipe, recipe)

	cmd := []string{"bash", "-c", script}
	res, err := e.containerMgr.ExecStream(ctx, e.containerID, cmd, 10*time.Minute)
	if logWriter != nil {
		for _, line := range strings.Split(string(res.Stdout), "\n") {
			if line != "" {
				logWriter.WriteLog("stdout", line)
			}
		}
		for _, line := range strings.Split(string(res.Stderr), "\n") {
			if line != "" {
				logWriter.WriteLog("stderr", line)
			}
		}
	}
	if err != nil {
		return err
	}
	if res.ExitCode != 0 {
		return fmt.Errorf("targeted cleanup exited with code %d", res.ExitCode)
	}
	return nil
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

	// Use fallback machine only if required Toradex/NXP layers are missing
	if machine == "verdin-imx8mp" {
		// Check presence of required layers by name in the loaded config
		has := func(name string) bool {
			for _, l := range e.config.Layers {
				if strings.EqualFold(l.Name, name) {
					return true
				}
			}
			return false
		}
		required := []string{"meta-freescale", "meta-freescale-3rdparty", "meta-toradex-bsp-common", "meta-toradex-nxp"}
		missing := false
		for _, r := range required {
			if !has(r) {
				missing = true
				break
			}
		}
		if missing {
			// Fall back only if we don't have the necessary BSP layers
			fmt.Printf("⚠️  Required Toradex/NXP layers not all present; falling back MACHINE to qemux86-64 for portability\n")
			machine = "qemux86-64"
		}
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
	// Use SSTATE_MIRRORS instead of SSTATE_DIR to avoid permission issues. Allow override via config.
	if strings.TrimSpace(e.config.Advanced.SStateMirrors) != "" {
		content.WriteString(fmt.Sprintf("SSTATE_MIRRORS = \"%s\"\n", e.config.Advanced.SStateMirrors))
	} else {
		content.WriteString("SSTATE_MIRRORS = \"file://.* file:///home/builder/sstate-cache/PATH\"\n")
	}

	// If a host tmp directory is mounted, direct TMPDIR to it so BitBake writes under a writable path
	if strings.TrimSpace(e.config.Directories.Tmp) != "" {
		content.WriteString("TMPDIR = \"/home/builder/tmp\"\n")
	}

	// Optional premirror configuration and network controls
	if strings.TrimSpace(e.config.Advanced.PreMirrors) != "" {
		content.WriteString(fmt.Sprintf("PREMIRRORS = \"%s\"\n", e.config.Advanced.PreMirrors))
	}
	if e.config.Advanced.NoNetwork {
		content.WriteString("BB_NO_NETWORK = \"1\"\n")
	}
	if e.config.Advanced.FetchPremirrorOnly {
		content.WriteString("BB_FETCH_PREMIRRORONLY = \"1\"\n")
	}

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

// extractFailedRecipe tries to parse the failed recipe from BitBake stderr output
func extractFailedRecipe(stderr string) string {
	lines := strings.Split(stderr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "ERROR: Task (") && strings.Contains(line, ":do_") {
			// Example: ERROR: Task (/home/builder/layers/poky/meta/recipes-devtools/strace/strace_5.16.bb:do_package) failed with exit code '1'
			start := strings.Index(line, "(")
			end := strings.Index(line, ":do_")
			if start >= 0 && end > start {
				path := line[start+1 : end]
				parts := strings.Split(path, "/")
				if len(parts) > 0 {
					bbFile := parts[len(parts)-1]
					// strace_5.16.bb -> strace
					if strings.HasSuffix(bbFile, ".bb") {
						nameVer := strings.TrimSuffix(bbFile, ".bb")
						// Take the base name before the first underscore (recipe name)
						base := nameVer
						if idx := strings.Index(base, "_"); idx > 0 {
							base = base[:idx]
						}
						return base
					}
				}
			}
		}
	}
	return ""
}
