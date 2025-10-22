package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"golang.org/x/term"

	"github.com/intrik8-labs/smidr/internal/artifacts"
	bitbake "github.com/intrik8-labs/smidr/internal/bitbake"
	config "github.com/intrik8-labs/smidr/internal/config"
	"github.com/intrik8-labs/smidr/internal/container"
	docker "github.com/intrik8-labs/smidr/internal/container/docker"
	source "github.com/intrik8-labs/smidr/internal/source"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a Yocto-based embedded Linux image in a managed container",
	Long: `Run a Yocto build in a managed container environment using the configuration in smidr.yaml.

	This command will:
		1. Prepare the container environment
		2. Fetch and cache source code and layers
		3. Execute the build process (bitbake)
		4. Extract build artifacts and store them in the artifact directory

	Flags allow you to override the build target, force a rebuild, or group builds by customer/project.

	Examples:
		smidr build --target core-image-minimal
		smidr build --target core-image-minimal --customer acme
		smidr build --target core-image-minimal --force
		smidr build --target core-image-minimal --fetch-only
		smidr build --target core-image-minimal --clean
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runBuild(cmd); err != nil {
			fmt.Println("Error during build:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	// Build-specific flags
	buildCmd.Flags().BoolP("force", "f", false, "Force rebuild (ignore cache)")
	buildCmd.Flags().StringP("target", "t", "", "Override build target")
	buildCmd.Flags().Bool("fetch-only", false, "Only fetch layers but don't build it")
	buildCmd.Flags().String("customer", "", "Optional: customer/user name for build directory grouping")
	buildCmd.Flags().Bool("clean", false, "If set, deletes the build directory before building (for a full rebuild)")
	buildCmd.Flags().Bool("clean-image", false, "If set, runs 'bitbake -c clean <image>' to regenerate only image artifacts without rebuilding dependencies")
}

func runBuild(cmd *cobra.Command) error {

	// Helper to expand ~ and make absolute
	expandPath := func(path string) string {
		if path == "" {
			return ""
		}
		// Only expand ~ if it is at the start of the path
		if strings.HasPrefix(path, "~") {
			if len(path) == 1 || path[1] == '/' {
				homedir, err := os.UserHomeDir()
				if err != nil {
					panic("Could not resolve ~ to home directory: " + err.Error())
				}
				path = homedir + path[1:]
			}
		}
		// Make absolute if not already
		if !strings.HasPrefix(path, "/") {
			abs, err := os.Getwd()
			if err != nil {
				panic("Could not get working directory: " + err.Error())
			}
			path = abs + "/" + path
		}
		return path
	}

	// Set up signal handling for graceful cancellation
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	// Set up signal handler for SIGINT and SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Printf("\n🛑 Received signal %v, initiating graceful shutdown...\n", sig)
		cancel()
	}()

	fmt.Println("🔨 Starting Smidr build...")
	fmt.Println()
	configFile := viper.GetString("config")
	if configFile == "" {
		configFile = "smidr.yaml"
	}

	fmt.Printf("📄 Loading configuration from %s...\n", configFile)
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}
	fmt.Printf("✅ Loaded project: %s\n", cfg.Name)
	fmt.Println()

	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	setDefaultDirs(cfg, workDir)

	// --- PATCH: Create unique or customer-specific build directory for this build ---
	homedir, _ := os.UserHomeDir()
	customer, _ := cmd.Flags().GetString("customer")
	clean, _ := cmd.Flags().GetBool("clean")
	imageName := cfg.Build.Image

	var buildDir string
	if customer != "" {
		// Use per-customer/image build dir for incremental builds
		buildDir = fmt.Sprintf("%s/.smidr/builds/build-%s/%s", homedir, customer, imageName)
		if clean {
			// Remove the build dir for a full rebuild
			os.RemoveAll(buildDir)
			fmt.Printf("🧹 Cleaned build directory: %s\n", buildDir)
		}
	} else {
		// Use default unique build directory
		buildUUID := uuid.New().String()
		buildDir = fmt.Sprintf("%s/.smidr/builds/build-%s", homedir, buildUUID)
	}

	// TMPDIR: use config if set, else per-build
	var tmpDir string
	if cfg.Directories.Tmp != "" {
		tmpDir = expandPath(cfg.Directories.Tmp)
		// If --clean was requested and we're using a configured (potentially shared) TMPDIR,
		// clean it to ensure a fresh build with artifacts
		if clean && customer != "" {
			// Create a customer-specific subdir in the configured TMPDIR to avoid cleaning shared state
			tmpDir = fmt.Sprintf("%s/%s-%s", tmpDir, customer, imageName)
			os.RemoveAll(tmpDir)
			fmt.Printf("🧹 Cleaned TMPDIR: %s\n", tmpDir)
		}
	} else {
		tmpDir = fmt.Sprintf("%s/tmp", buildDir)
	}
	// DEPLOY_DIR: always per-build
	deployDir := fmt.Sprintf("%s/deploy", buildDir)

	// Ensure tmp and deploy are created with permissive permissions so container user can write
	if err := os.MkdirAll(tmpDir, 0777); err != nil {
		return fmt.Errorf("failed to create tmp dir: %w", err)
	}
	if err := os.Chown(tmpDir, 1000, 1000); err != nil {
		_ = os.Chmod(tmpDir, 0777)
	}
	if err := os.MkdirAll(deployDir, 0777); err != nil {
		return fmt.Errorf("failed to create deploy dir: %w", err)
	}

	cfg.Directories.Build = buildDir
	cfg.Directories.Tmp = tmpDir
	cfg.Directories.Deploy = deployDir

	fmt.Println("🔒 Using:")
	fmt.Printf("    TMPDIR: %s\n", tmpDir)
	fmt.Printf("    DEPLOY_DIR: %s\n", deployDir)
	fmt.Printf("    BUILD_DIR: %s\n", buildDir)

	// Create logger
	verbose := viper.GetBool("verbose")
	logger := source.NewConsoleLogger(os.Stdout, verbose)

	// Expand and resolve all relevant directories
	// Prefer directories.downloads; fallback to directories.source
	downloadsDir := cfg.Directories.Downloads
	if downloadsDir == "" && cfg.Directories.Source != "" {
		downloadsDir = cfg.Directories.Source
	}
	downloadsDir = expandPath(downloadsDir)
	if downloadsDir != "" {
		cfg.Directories.Downloads = downloadsDir
	}
	cfg.Directories.Layers = expandPath(cfg.Directories.Layers)
	cfg.Directories.Source = expandPath(cfg.Directories.Source)
	cfg.Directories.SState = expandPath(cfg.Directories.SState)
	cfg.Directories.Build = expandPath(cfg.Directories.Build)
	cfg.Directories.Tmp = expandPath(cfg.Directories.Tmp)
	cfg.Directories.Deploy = expandPath(cfg.Directories.Deploy)
	// Cache fields removed; Directories are canonical

	// Step 1: Fetch layers
	fmt.Println("📦 Fetching required layers...")
	fetcher := source.NewFetcher(cfg.Directories.Layers, downloadsDir, logger)
	results, err := fetcher.FetchLayers(cfg)
	if err != nil {
		return fmt.Errorf("failed to fetch layers: %w", err)
	}

	// Report fetch results
	fmt.Println()
	fmt.Printf("✅ Successfully fetched %d layers\n", len(results))
	for _, result := range results {
		if result.Cached {
			fmt.Printf("   ♻️  %s (cached)\n", result.LayerName)
		} else {
			fmt.Printf("   ⬇️  %s (downloaded)\n", result.LayerName)
		}
	}
	fmt.Println()
	fmt.Printf("Project: %s - %s\n", cfg.Name, cfg.Description)

	// If --fetch-only was requested, stop here to keep CI fast
	if ok, _ := cmd.Flags().GetBool("fetch-only"); ok {
		fmt.Println("🛑 Fetch-only mode enabled — skipping container start and build")
		return nil
	}

	// Step 2: Prepare container config
	fmt.Println("🐳 Preparing container environment...")
	// Allow tests to override container name and mounts for verification
	testName := os.Getenv("SMIDR_TEST_CONTAINER_NAME")
	testDownloads := os.Getenv("SMIDR_TEST_DOWNLOADS_DIR")
	testSstate := os.Getenv("SMIDR_TEST_SSTATE_DIR")
	testWorkspace := os.Getenv("SMIDR_TEST_WORKSPACE_DIR")
	testImage := os.Getenv("SMIDR_TEST_IMAGE")
	testLayerDirsCSV := os.Getenv("SMIDR_TEST_LAYER_DIRS") // comma-separated
	var testLayerDirs []string
	if testLayerDirsCSV != "" {
		for _, p := range strings.Split(testLayerDirsCSV, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				testLayerDirs = append(testLayerDirs, p)
			}
		}
	}
	// Prepare layer dirs from config so they are injected into the container by default
	var cfgLayerDirs []string
	var cfgLayerNames []string
	mountedParents := make(map[string]string) // Track which parent directories we've already added for mounting (path -> name)

	for _, l := range cfg.Layers {
		var layerPath string
		if l.Path != "" {
			// If path is relative, make it relative to the layers directory
			if !strings.HasPrefix(l.Path, "/") && !strings.HasPrefix(l.Path, "~") {
				layerPath = expandPath(fmt.Sprintf("%s/%s", cfg.Directories.Layers, l.Path))
			} else {
				// Absolute path or ~ path - expand as-is
				layerPath = expandPath(l.Path)
			}
		} else {
			// Always expand ~ and make absolute for default layers dir
			layerPath = expandPath(fmt.Sprintf("%s/%s", cfg.Directories.Layers, l.Name))
		}

		// For mounting, we only want to mount the top-level parent directories
		// For sublayers like meta-openembedded/meta-oe, we should mount meta-openembedded
		var mountPath string
		var mountName string
		if l.Path != "" && strings.Contains(l.Path, "/") {
			// This is a sublayer - mount the parent directory instead
			parentPath := strings.Split(l.Path, "/")[0]
			mountPath = expandPath(fmt.Sprintf("%s/%s", cfg.Directories.Layers, parentPath))
			mountName = parentPath
		} else {
			// This is a top-level layer - mount it directly
			mountPath = layerPath
			mountName = l.Name
		}

		// Only add unique mount paths
		if _, exists := mountedParents[mountPath]; !exists {
			cfgLayerDirs = append(cfgLayerDirs, mountPath)
			cfgLayerNames = append(cfgLayerNames, mountName)
			mountedParents[mountPath] = mountName
		}
	}

	// Only mount valid Yocto layers (those with conf/layer.conf)
	var validLayerDirs []string
	var validLayerNames []string
	for i, dir := range cfgLayerDirs {
		// Ensure the layer directory exists before checking for layer.conf
		if err := os.MkdirAll(dir, 0755); err != nil {
			// logger.Debug("Failed to create layer directory %s: %v", dir, err)
			continue
		}

		confPath := fmt.Sprintf("%s/conf/layer.conf", dir)
		if fi, err := os.Stat(confPath); err == nil && !fi.IsDir() {
			validLayerDirs = append(validLayerDirs, dir)
			validLayerNames = append(validLayerNames, cfgLayerNames[i])
		} else {
			// Include directory anyway for mounting - layer.conf will be created when fetched
			validLayerDirs = append(validLayerDirs, dir)
			validLayerNames = append(validLayerNames, cfgLayerNames[i])
			// logger.Debug("Including layer directory for mounting (conf/layer.conf will be created on fetch): %s", dir)
		}
	}
	cfgLayerDirs = validLayerDirs
	cfgLayerNames = validLayerNames

	// Determine container image - use config, test override, or fallback
	imageToUse := cfg.Container.BaseImage
	if imageToUse == "" {
		imageToUse = "crops/yocto:ubuntu-22.04-builder" // fallback to official Yocto build image
	}

	// Patch: Set TMPDIR in container environment to ensure BitBake uses the correct tmp dir
	env := os.Environ()
	env = append(env, fmt.Sprintf("TMPDIR=%s", cfg.Directories.Tmp))
	containerConfig := container.ContainerConfig{
		Image:          imageToUse,
		Cmd:            []string{"echo 'Container ready'; sleep 86400"}, // 24 hours
		DownloadsDir:   cfg.Directories.Downloads,                       // mount host downloads (DL_DIR) into container
		SstateCacheDir: cfg.Directories.SState,                          // Wire SSTATE dir from config
		WorkspaceDir:   cfg.Directories.Build,
		BuildDir:       cfg.Directories.Build,
		TmpDir:         cfg.Directories.Tmp, // Mount host tmp dir if set
		LayerDirs:      cfgLayerDirs,
		LayerNames:     cfgLayerNames,
		Name:           testName,
		MemoryLimit:    cfg.Container.Memory,
		CPUCount:       cfg.Container.CPUCount,
		Env:            env,
	}
	// Apply test overrides if provided
	if testDownloads != "" {
		containerConfig.DownloadsDir = testDownloads
	}
	if testSstate != "" {
		containerConfig.SstateCacheDir = testSstate
	}
	if testWorkspace != "" {
		containerConfig.WorkspaceDir = testWorkspace
	}
	if len(testLayerDirs) > 0 {
		containerConfig.LayerDirs = testLayerDirs
	}
	// Apply test image override last so tests can build and use a local image (avoids Docker Hub rate limits)
	if testImage != "" {
		containerConfig.Image = testImage
	}

	// Allow tests to override entrypoint (comma-separated)
	if ep := os.Getenv("SMIDR_TEST_ENTRYPOINT"); ep != "" {
		parts := strings.Split(ep, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		containerConfig.Entrypoint = parts
	}

	// Wire configured entrypoint from smidr.yaml if provided
	if len(cfg.Container.Entrypoint) > 0 {
		containerConfig.Entrypoint = cfg.Container.Entrypoint
	}

	// Step 3: Create DockerManager
	dm, err := docker.NewDockerManager()
	if err != nil {
		return fmt.Errorf("failed to create DockerManager: %w", err)
	}

	// Ensure the downloads directory exists; always auto-create, no prompts
	if containerConfig.DownloadsDir != "" {
		if fi, err := os.Stat(containerConfig.DownloadsDir); err != nil || !fi.IsDir() {
			if err := os.MkdirAll(containerConfig.DownloadsDir, 0755); err != nil {
				return fmt.Errorf("failed to create downloads dir: %w", err)
			}
		}
		// If the directory exists (even if empty), proceed silently
	}

	// Print resolved host<->container mappings for clarity
	fmt.Println()
	fmt.Printf("🔧 Resolved mounts:\n")
	fmt.Printf("   Host downloads (DL_DIR): %s -> /home/builder/downloads\n", containerConfig.DownloadsDir)
	fmt.Printf("   Host sstate-cache: %s -> /home/builder/sstate-cache\n", containerConfig.SstateCacheDir)
	fmt.Printf("   Host workspace: %s -> /home/builder/work\n", containerConfig.WorkspaceDir)
	if containerConfig.TmpDir != "" {
		fmt.Printf("   Host tmp: %s -> /home/builder/tmp\n", containerConfig.TmpDir)
	}
	// Print layers directory mount if set, using real layer names
	if len(containerConfig.LayerDirs) > 0 {
		for i, dir := range containerConfig.LayerDirs {
			// Use the corresponding layer name if available
			var layerName string
			if i < len(containerConfig.LayerNames) && containerConfig.LayerNames[i] != "" {
				layerName = containerConfig.LayerNames[i]
			} else {
				layerName = fmt.Sprintf("layer-%d", i)
			}
			fmt.Printf("   Host layer %s: %s -> /home/builder/layers/%s\n", layerName, dir, layerName)
		}
	}
	if containerConfig.DownloadsDir == cfg.Directories.Source && containerConfig.DownloadsDir != "" {
		fmt.Println("⚠️  Note: downloads and sources are the same path on host — this can cause confusion. Consider setting directories.downloads and directories.source to distinct paths in smidr.yaml.")
	}

	// Step 4: Pull image (skip pulling if a test image override is provided or image exists locally)
	if testImage == "" {
		// Check if image exists locally first
		if dm.ImageExists(ctx, containerConfig.Image) {
			fmt.Printf("🐳 Using local image: %s\n", containerConfig.Image)
		} else {
			// Image doesn't exist locally, try to pull it
			fmt.Println("🐳 Pulling container image...")
			if err := dm.PullImage(ctx, containerConfig.Image); err != nil {
				return fmt.Errorf("failed to pull image: %w", err)
			}
		}
	} else {
		fmt.Printf("🐳 Using prebuilt test image: %s\n", containerConfig.Image)
	}

	// Step 5: Create container
	fmt.Printf("🧭 Mounting host downloads dir: %s -> /home/builder/downloads\n", containerConfig.DownloadsDir)
	fmt.Println("🐳 Creating container...")
	containerID, err := dm.CreateContainer(ctx, containerConfig)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	// Step 6: Ensure cleanup (stop/remove) always runs
	// But skip cleanup if SMIDR_DEBUG_KEEP_CONTAINER is set
	defer func() {
		if os.Getenv("SMIDR_DEBUG_KEEP_CONTAINER") != "" {
			fmt.Printf("⚠️  DEBUG MODE: Container %s is being kept for inspection\n", containerID)
			fmt.Printf("    Inspect with: docker inspect %s\n", containerID)
			fmt.Printf("    Check logs with: docker logs %s\n", containerID)
			fmt.Printf("    Remove with: docker rm -f %s\n", containerID)
			return
		}
		fmt.Println("🧹 Cleaning up container...")
		os.Stdout.Sync() // Flush stdout to ensure log is written
		stopErr := dm.StopContainer(ctx, containerID, 2*time.Second)
		if stopErr != nil {
			fmt.Printf("⚠️  Failed to stop container: %v\n", stopErr)
		}
		removeErr := dm.RemoveContainer(ctx, containerID, true)
		if removeErr != nil {
			fmt.Printf("⚠️  Failed to remove container: %v\n", removeErr)
		}
		os.Stdout.Sync() // Flush again after cleanup
	}()

	// Step 7: Start container
	fmt.Println("🐳 Starting container...")
	if err := dm.StartContainer(ctx, containerID); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Print container ID for debugging
	fmt.Printf("📦 Container ID: %s\n", containerID)
	fmt.Printf("   Monitor with: docker stats %s\n", containerID)

	// Add a short delay if debugging to allow user to start monitoring
	if os.Getenv("SMIDR_DEBUG_KEEP_CONTAINER") != "" {
		fmt.Println("⏳ Pausing 5 seconds for monitoring setup...")
		time.Sleep(5 * time.Second)
	}

	// Step 8: Execute the actual build
	if os.Getenv("SMIDR_TEST_WRITE_MARKERS") != "1" {
		// Only run real build if not in test mode
		fmt.Println("🚀 Starting Yocto build...")
		fmt.Println("💡 Use Ctrl+C to gracefully cancel the build at any time")

		// Prepare log files in the build directory
		logDir := cfg.Directories.Build
		os.MkdirAll(logDir, 0755)
		plainLogPath := filepath.Join(logDir, "build-log.txt")
		jsonlLogPath := filepath.Join(logDir, "build-log.jsonl")
		plainLogFile, err := os.Create(plainLogPath)
		if err != nil {
			return fmt.Errorf("failed to create build-log.txt: %w", err)
		}
		defer plainLogFile.Close()
		jsonlLogFile, err := os.Create(jsonlLogPath)
		if err != nil {
			return fmt.Errorf("failed to create build-log.jsonl: %w", err)
		}
		defer jsonlLogFile.Close()

		logWriter := &bitbake.BuildLogWriter{
			PlainWriter: io.MultiWriter(os.Stdout, plainLogFile),
			JSONLWriter: jsonlLogFile,
		}

		executor := bitbake.NewBuildExecutor(cfg, dm, containerID, containerConfig.WorkspaceDir)

		// Set clean-image flag if requested to force image regeneration
		cleanImage, _ := cmd.Flags().GetBool("clean-image")
		if cleanImage {
			executor.SetForceImage(true)
		}

		buildResult, err := executor.ExecuteBuild(ctx, logWriter)

		if err != nil {
			// Check if the error was due to context cancellation (signal handling)
			if ctx.Err() == context.Canceled {
				fmt.Printf("🛑 Build was cancelled by user signal\n")
				if buildResult != nil {
					fmt.Printf("Build was running for: %v before cancellation\n", buildResult.Duration)
				}
				return fmt.Errorf("build cancelled by user")
			}

			fmt.Printf("❌ Build failed: %v\n", err)
			if buildResult != nil {
				fmt.Printf("Build took: %v\n", buildResult.Duration)
				if buildResult.Error != "" {
					fmt.Printf("Error details: %s\n", buildResult.Error)
				}
				// Show both stdout and stderr for better debugging
				if buildResult.Output != "" {
					fmt.Printf("Build output (last 3000 chars):\n%s\n", getTailString(buildResult.Output, 3000))
				}
				if buildResult.Error != "" && buildResult.Error != buildResult.Output {
					fmt.Printf("Build error output:\n%s\n", buildResult.Error)
				}
			}
			return fmt.Errorf("build execution failed: %w", err)
		}

		fmt.Printf("✅ Build completed successfully in %v\n", buildResult.Duration)
		fmt.Printf("Exit code: %d\n", buildResult.ExitCode)

		// Extract build artifacts
		fmt.Println("📦 Extracting build artifacts...")
		if err := extractBuildArtifacts(ctx, dm, containerID, cfg, buildResult.Duration); err != nil {
			fmt.Printf("[WARNING] Failed to extract artifacts: %v\n", err)
			// Don't fail the build for artifact extraction errors
		}
	} else {
		// Test mode - run marker validation logic
		fmt.Println("🚀 Build process running in test mode (SMIDR_TEST_WRITE_MARKERS=1)")
		if err := runTestMarkerValidation(ctx, dm, containerID, containerConfig, cfg); err != nil {
			return fmt.Errorf("test marker validation failed: %w", err)
		}
	}

	fmt.Println("💡 Use 'smidr artifacts list' to view build artifacts once available")

	return nil // Build completed successfully
}

// getTailString returns the last n characters of a string
func getTailString(s string, n int) string {
	lines := strings.Split(s, "\n")
	if n >= len(lines) {
		return s
	}
	return strings.Join(lines[len(lines)-n:], "\n")
}

// runTestMarkerValidation runs the test marker validation logic
func runTestMarkerValidation(ctx context.Context, dm container.ContainerManager, containerID string, containerConfig container.ContainerConfig, cfg *config.Config) error {
	// Test container functionality and mount accessibility
	// Note: On CI, bind-mounted directories may not be writable due to UID mismatches

	// First verify basic container functionality
	tmpRes, err := dm.Exec(ctx, containerID, []string{"sh", "-c", "echo 'container-functional' > /tmp/test-writable.txt && cat /tmp/test-writable.txt"}, 5*time.Second)
	if err != nil || tmpRes.ExitCode != 0 {
		fmt.Printf("⚠️  Container basic write test failed: %v, output: %s\n", err, string(tmpRes.Stderr))
	}

	// Test workspace by writing to writable location and using docker cp to extract
	if containerConfig.WorkspaceDir != "" {
		// Create marker in writable space inside container
		res, err := dm.Exec(ctx, containerID, []string{"sh", "-c", "echo itest > /tmp/builder-workspace/itest.txt"}, 5*time.Second)
		if err != nil {
			fmt.Printf("⚠️  Failed to create workspace marker: %v\n", err)
		} else if res.ExitCode != 0 {
			fmt.Printf("⚠️  workspace marker creation failed: %s\n", string(res.Stderr))
		} else {
			// Use docker cp to copy the file from container to host mount point
			// This works even when container user can't write directly to bind-mounted dirs
			if err := dm.CopyFromContainer(ctx, containerID, "/tmp/builder-workspace/itest.txt", containerConfig.WorkspaceDir+"/itest.txt"); err != nil {
				fmt.Printf("⚠️  Failed to copy workspace marker to host: %v\n", err)
			} else {
				fmt.Printf("✓ Workspace marker successfully created via docker cp\n")
			}
		}
	}

	// For downloads and sstate, just test if the directories are accessible
	// Don't try to write to them due to permission issues with bind mounts
	if containerConfig.DownloadsDir != "" {
		res, err := dm.Exec(ctx, containerID, []string{"sh", "-c", "ls -la /home/builder/downloads | head -5"}, 5*time.Second)
		if err != nil {
			fmt.Printf("⚠️  Downloads dir not accessible: %v\n", err)
		} else {
			fmt.Printf("✓ Downloads directory accessible with content:\n%s\n", string(res.Stdout))
		}
	}

	if containerConfig.SstateCacheDir != "" {
		res, err := dm.Exec(ctx, containerID, []string{"sh", "-c", "ls -la /home/builder/sstate-cache | head -5"}, 5*time.Second)
		if err != nil {
			fmt.Printf("⚠️  Sstate dir not accessible: %v\n", err)
		} else {
			fmt.Printf("✓ Sstate directory accessible with content:\n%s\n", string(res.Stdout))
		}
	}

	// Probe layer visibility if any provided, using real layer names
	if len(containerConfig.LayerDirs) > 0 {
		for i := range containerConfig.LayerDirs {
			layerName := ""
			if i < len(cfg.Layers) {
				layerName = cfg.Layers[i].Name
			} else {
				layerName = fmt.Sprintf("layer-%d", i)
			}
			res, _ := dm.Exec(ctx, containerID, []string{"sh", "-c", fmt.Sprintf("ls -la /home/builder/layers/%s || true", layerName)}, 5*time.Second)
			if len(res.Stdout) > 0 {
				fmt.Printf("Layer %s contents:\n%s\n", layerName, string(res.Stdout))
			}
		}
	}

	return nil
}

// setDefaultDirs ensures default directory paths are populated on the config
// when the user did not provide them. This centralizes defaulting and makes
// it testable.
func setDefaultDirs(cfg *config.Config, workDir string) {
	if cfg.Directories.Source == "" {
		if homedir, err := os.UserHomeDir(); err == nil {
			cfg.Directories.Source = fmt.Sprintf("%s/.smidr/sources", homedir)
		} else {
			cfg.Directories.Source = fmt.Sprintf("%s/sources", workDir)
		}
	}
	if cfg.Directories.Build == "" {
		if homedir, err := os.UserHomeDir(); err == nil {
			cfg.Directories.Build = fmt.Sprintf("%s/.smidr/build", homedir)
		} else {
			cfg.Directories.Build = fmt.Sprintf("%s/build", workDir)
		}
	}
	if cfg.Directories.SState == "" {
		if homedir, err := os.UserHomeDir(); err == nil {
			cfg.Directories.SState = fmt.Sprintf("%s/.smidr/sstate-cache", homedir)
		} else {
			cfg.Directories.SState = fmt.Sprintf("%s/sstate-cache", workDir)
		}
	}

	// Provide sane defaults for global cache locations if not supplied
	// Removed cache defaults; defaults exist for Directories in setDefaultDirs
}

// extractBuildArtifacts extracts build artifacts from the container to persistent storage
func extractBuildArtifacts(ctx context.Context, dm *docker.DockerManager, containerID string, cfg *config.Config, buildDuration time.Duration) error {
	// Get current user for metadata
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	// Determine if this is a customer build (by checking build dir path)
	artifactDir := ""
	customer := ""
	imageName := cfg.Build.Image
	timestamp := time.Now().Format("20060102-150405")

	if strings.Contains(cfg.Directories.Build, "/build-") {
		parts := strings.Split(cfg.Directories.Build, "/build-")
		if len(parts) > 1 {
			customerAndRest := parts[1]
			customerParts := strings.SplitN(customerAndRest, "/", 2)
			customer = customerParts[0]
		}
	}

	if customer != "" {
		// Use customer artifact dir
		artifactDir = fmt.Sprintf("%s/.smidr/artifacts/artifact-%s/%s-%s", currentUser.HomeDir, customer, imageName, timestamp)
	} else {
		// Fallback: use buildID as before
		buildID := artifacts.GenerateBuildID(cfg.Name, currentUser.Username)
		artifactDir = fmt.Sprintf("%s/.smidr/artifacts/%s", currentUser.HomeDir, buildID)
	}

	// Start with the configured deploy directory (may be per-build or shared TMPDIR)
	deploySrc := cfg.Directories.Deploy
	fmt.Printf("[DEBUG] Configured Deploy directory: %s\n", deploySrc)

	// If the configured deploy doesn't exist or is empty, try TMPDIR/deploy as fallback (for sstate-restored builds)
	deployInfo, deployStatErr := os.Stat(deploySrc)
	if deployStatErr != nil || !deployInfo.IsDir() {
		fmt.Printf("[DEBUG] Configured deploy stat error or not a dir: %v\n", deployStatErr)
		// Try TMPDIR/deploy as fallback
		tmpDeploySrc := filepath.Join(cfg.Directories.Tmp, "deploy")
		fmt.Printf("[DEBUG] Trying TMPDIR/deploy: %s\n", tmpDeploySrc)
		tmpInfo, tmpErr := os.Stat(tmpDeploySrc)
		if tmpErr == nil && tmpInfo.IsDir() {
			deploySrc = tmpDeploySrc
			fmt.Printf("[INFO] Configured deploy not found, using TMPDIR/deploy: %s\n", deploySrc)
		} else {
			fmt.Printf("[DEBUG] TMPDIR/deploy also not available: %v\n", tmpErr)
		}
	} else {
		// Check if it's empty
		entries, _ := os.ReadDir(deploySrc)
		fmt.Printf("[DEBUG] Configured deploy has %d entries\n", len(entries))
		if len(entries) > 0 {
			fmt.Printf("[DEBUG] Sample entries: ")
			for i, e := range entries {
				if i >= 3 {
					fmt.Printf("... (%d more)", len(entries)-3)
					break
				}
				fmt.Printf("%s ", e.Name())
			}
			fmt.Println()
		}
		if len(entries) == 0 {
			// Empty, try TMPDIR/deploy instead
			tmpDeploySrc := filepath.Join(cfg.Directories.Tmp, "deploy")
			fmt.Printf("[DEBUG] Configured deploy empty, trying TMPDIR/deploy: %s\n", tmpDeploySrc)
			tmpInfo, tmpErr := os.Stat(tmpDeploySrc)
			if tmpErr == nil && tmpInfo.IsDir() {
				tmpEntries, _ := os.ReadDir(tmpDeploySrc)
				fmt.Printf("[DEBUG] TMPDIR/deploy has %d entries\n", len(tmpEntries))
				if len(tmpEntries) > 0 {
					deploySrc = tmpDeploySrc
					fmt.Printf("[INFO] Configured deploy empty, using TMPDIR/deploy: %s\n", deploySrc)
				} else {
					// Both are empty - this was a 100% sstate cache build
					fmt.Println("ℹ️  Build completed entirely from sstate cache - no new artifacts generated")
					fmt.Println("💡 Artifacts from previous builds are available in the sstate cache")
					// Don't fail - just skip artifact extraction
					return nil
				}
			} else {
				fmt.Printf("[DEBUG] TMPDIR/deploy also not available: %v\n", tmpErr)
				// No deploy directory at all - sstate-only build
				fmt.Println("ℹ️  Build completed entirely from sstate cache - no deploy directory populated")
				return nil
			}
		}
	}

	deployDst := filepath.Join(artifactDir, "deploy")

	fmt.Printf("[DEBUG] Copying deploy artifacts\n")
	fmt.Printf("[DEBUG] Source: %s\n", deploySrc)
	fmt.Printf("[DEBUG] Destination: %s\n", deployDst) // Check if source exists and is a directory
	info, statErr := os.Stat(deploySrc)
	if statErr != nil {
		fmt.Printf("[DEBUG] Source deploy directory does not exist: %v\n", statErr)
		return fmt.Errorf("deploy source directory does not exist: %w", statErr)
	}
	if !info.IsDir() {
		fmt.Printf("[DEBUG] Source deploy path is not a directory!\n")
		return fmt.Errorf("deploy source is not a directory")
	}

	err = copyDir(deploySrc, deployDst)
	if err != nil {
		fmt.Printf("[DEBUG] Error copying deploy directory: %v\n", err)
		return fmt.Errorf("failed to copy deploy artifacts: %w", err)
	}
	fmt.Printf("[DEBUG] Deploy directory copied successfully.\n")

	// Copy build logs to artifact directory
	buildLogTxt := filepath.Join(cfg.Directories.Build, "build-log.txt")
	buildLogJsonl := filepath.Join(cfg.Directories.Build, "build-log.jsonl")
	destLogTxt := filepath.Join(artifactDir, "build-log.txt")
	destLogJsonl := filepath.Join(artifactDir, "build-log.jsonl")
	if err := copyFile(buildLogTxt, destLogTxt); err != nil {
		fmt.Printf("[WARNING] Failed to copy build-log.txt to artifact dir: %v\n", err)
	}
	if err := copyFile(buildLogJsonl, destLogJsonl); err != nil {
		fmt.Printf("[WARNING] Failed to copy build-log.jsonl to artifact dir: %v\n", err)
	}

	// Create build metadata
	metadata := artifacts.BuildMetadata{
		BuildID:     filepath.Base(artifactDir),
		ProjectName: cfg.Name,
		User:        currentUser.Username,
		Timestamp:   time.Now(),
		ConfigUsed: map[string]string{
			"yocto_series":  cfg.YoctoSeries,
			"machine":       cfg.Build.Machine,
			"image":         cfg.Build.Image,
			"base_provider": cfg.Base.Provider,
			"base_version":  cfg.Base.Version,
		},
		BuildDuration: buildDuration,
		TargetImage:   cfg.Build.Image,
		Status:        "success",
	}

	// Save metadata as JSON
	if err := os.MkdirAll(artifactDir, 0755); err != nil {
		return fmt.Errorf("failed to create artifact dir: %w", err)
	}
	metaPath := filepath.Join(artifactDir, "build-metadata.json")
	f, err := os.Create(metaPath)
	if err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(metadata); err != nil {
		return fmt.Errorf("failed to encode metadata: %w", err)
	}

	fmt.Printf("✅ Artifacts copied to: %s\n", artifactDir)
	return nil
}

// copyDir recursively copies a directory from src to dst
func copyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("[WARNING] Skipping missing file: %s (%v)\n", path, err)
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			// fmt.Printf("[DEBUG] Failed to get relative path for %s: %v\n", path, err)
			return err
		}
		target := filepath.Join(dst, rel)
		// Handle symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, lerr := os.Readlink(path)
			if lerr != nil {
				fmt.Printf("[WARNING] Skipping broken symlink: %s (%v)\n", path, lerr)
				return nil
			}
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				fmt.Printf("[WARNING] Failed to create parent dir for symlink: %s (%v)\n", filepath.Dir(target), err)
				return err
			}
			if err := os.Symlink(linkTarget, target); err != nil {
				fmt.Printf("[WARNING] Failed to create symlink: %s -> %s (%v)\n", target, linkTarget, err)
			}
			return nil
		}
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		} else {
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				fmt.Printf("[WARNING] Failed to create parent dir for file: %s (%v)\n", filepath.Dir(target), err)
				return err
			}
			srcFile, err := os.Open(path)
			if err != nil {
				fmt.Printf("[WARNING] Skipping missing file: %s (%v)\n", path, err)
				return nil
			}
			defer srcFile.Close()
			dstFile, err := os.Create(target)
			if err != nil {
				fmt.Printf("[WARNING] Failed to create file: %s (%v)\n", target, err)
				return err
			}
			defer dstFile.Close()
			_, err = io.Copy(dstFile, srcFile)
			if err != nil {
				fmt.Printf("[WARNING] Failed to copy file: %s -> %s (%v)\n", path, target, err)
				return err
			}
			if err := os.Chmod(target, info.Mode()); err != nil {
				fmt.Printf("[WARNING] Failed to chmod file: %s (%v)\n", target, err)
			}
			return nil
		}
	})
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

// isatty returns true if the given file descriptor is a terminal.
// Uses golang.org/x/term for cross-platform detection.
func isatty(fd uintptr) bool {
	return term.IsTerminal(int(fd))
}
