package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"

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
	Short: "Build the embedded Linux image",
	Long: `Start building the embedded Linux image according to the
configuration specified in smidr.yaml.

This will:
1. Prepare the container environment
2. Fetch and cache source code
3. Execute the build process
4. Extract build artifacts`,
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
}

func runBuild(cmd *cobra.Command) error {
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

	// Create logger
	verbose := viper.GetBool("verbose")
	logger := source.NewConsoleLogger(os.Stdout, verbose)

	// Determine host downloads dir to mount into the container.
	// Prefer the configured global cache (cfg.Cache.Downloads) so users can
	// share a single DL_DIR across projects. Fall back to cfg.Directories.Source
	// only if cache.downloads is not set. Also expand a leading ~ to the user's
	// home directory so YAML values like "~/.smidr/downloads" work as expected.
	downloadsDir := ""
	if cfg.Cache.Downloads != "" {
		downloadsDir = cfg.Cache.Downloads
	} else if cfg.Directories.Source != "" {
		downloadsDir = cfg.Directories.Source
	}
	// Expand leading ~ to home dir
	if downloadsDir != "" && strings.HasPrefix(downloadsDir, "~") {
		if homedir, err := os.UserHomeDir(); err == nil {
			downloadsDir = strings.Replace(downloadsDir, "~", homedir, 1)
		}
	}

	// Step 1: Fetch layers
	fmt.Println("📦 Fetching required layers...")
	fetcher := source.NewFetcher(cfg.Directories.Source, downloadsDir, logger)
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
	for _, l := range cfg.Layers {
		if l.Path != "" {
			cfgLayerDirs = append(cfgLayerDirs, l.Path)
		} else {
			// default to sources/<layer.Name> under the configured sources dir
			cfgLayerDirs = append(cfgLayerDirs, fmt.Sprintf("%s/%s", cfg.Directories.Source, l.Name))
		}
	}

	// Determine container image - use config, test override, or fallback
	imageToUse := cfg.Container.BaseImage
	if imageToUse == "" {
		imageToUse = "crops/yocto:ubuntu-22.04-builder" // fallback to official Yocto build image
	}

	containerConfig := container.ContainerConfig{
		Image: imageToUse,
		// Use a portable shell invocation. Using ["sh", "-c", "..."] runs the
		// command under /bin/sh when no ENTRYPOINT is set (busybox), and when an
		// image *does* set an ENTRYPOINT (for example /bin/bash) Docker will append
		// these args to the entrypoint; in practice bash will accept 'sh -c ...'
		// and run the given commands. Keep the container alive briefly so tests
		// can exec into it.
		// Provide a single string command; the Docker manager will use /bin/sh -c
		// as the Entrypoint so this will run consistently regardless of whether
		// the image defines its own ENTRYPOINT.
		// Keep container alive for a long time to allow exec commands and debugging
		Cmd:            []string{"echo 'Container ready'; sleep 3600"},
		DownloadsDir:   downloadsDir,           // mount host downloads (DL_DIR) into container
		SstateCacheDir: cfg.Directories.SState, // Wire SSTATE dir from config
		WorkspaceDir:   cfg.Directories.Build,
		LayerDirs:      cfgLayerDirs,
		Name:           testName,
		MemoryLimit:    cfg.Container.Memory,
		CPUCount:       cfg.Container.CPUCount,
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

	// Validate the resolved downloads directory exists and optionally create it
	// Default behavior: auto-create when running non-interactively (CI) or when
	// SMIDR_AUTO_CREATE_DOWNLOADS=1 is set. If running interactively, prompt the
	// user unless auto-create is enabled.
	if containerConfig.DownloadsDir != "" {
		autoCreate := os.Getenv("SMIDR_AUTO_CREATE_DOWNLOADS") == "1"
		// Detect non-interactive environment by checking if stdin is a terminal
		if !isatty(os.Stdin.Fd()) {
			autoCreate = true
		}

		if fi, err := os.Stat(containerConfig.DownloadsDir); err != nil || !fi.IsDir() {
			if autoCreate {
				if err := os.MkdirAll(containerConfig.DownloadsDir, 0755); err != nil {
					return fmt.Errorf("failed to create downloads dir: %w", err)
				}
				fmt.Printf("Created downloads dir: %s\n", containerConfig.DownloadsDir)
			} else {
				fmt.Printf("⚠️  Downloads dir %s does not exist. Create it now? (y/N): ", containerConfig.DownloadsDir)
				reader := bufio.NewReader(os.Stdin)
				resp, _ := reader.ReadString('\n')
				resp = strings.TrimSpace(resp)
				if strings.ToLower(resp) == "y" {
					if err := os.MkdirAll(containerConfig.DownloadsDir, 0755); err != nil {
						return fmt.Errorf("failed to create downloads dir: %w", err)
					}
					fmt.Printf("Created %s\n", containerConfig.DownloadsDir)
				} else {
					fmt.Println("Proceeding without downloads dir. Fetch step may fail.")
				}
			}
		} else {
			// Directory exists - check if empty
			f, _ := os.Open(containerConfig.DownloadsDir)
			names, _ := f.Readdirnames(1)
			f.Close()
			if len(names) == 0 {
				if !autoCreate {
					fmt.Printf("⚠️  Downloads dir %s exists but is empty. Continue? (y/N): ", containerConfig.DownloadsDir)
					reader := bufio.NewReader(os.Stdin)
					resp, _ := reader.ReadString('\n')
					resp = strings.TrimSpace(resp)
					if strings.ToLower(resp) != "y" {
						fmt.Println("Aborting build. Populate the downloads dir or continue with fetch later.")
						return nil
					}
				} else {
					// autoCreate true -> continue silently
				}
			}
		}
	}

	// Print resolved host<->container mappings for clarity
	fmt.Println()
	fmt.Printf("🔧 Resolved mounts:\n")
	fmt.Printf("   Host downloads (DL_DIR): %s -> /home/builder/downloads\n", containerConfig.DownloadsDir)
	fmt.Printf("   Host sstate-cache: %s -> /home/builder/sstate-cache\n", containerConfig.SstateCacheDir)
	fmt.Printf("   Host workspace: %s -> /home/builder/work\n", containerConfig.WorkspaceDir)
	if containerConfig.DownloadsDir == cfg.Directories.Source && containerConfig.DownloadsDir != "" {
		fmt.Println("⚠️  Note: downloads and sources are the same path on host — this can cause confusion. Consider setting cache.downloads and directories.source to distinct paths in smidr.yaml.")
	}

	// Step 4: Pull image (skip pulling if a test image override is provided or image exists locally)
	if testImage == "" {
		// Check if image exists locally first
		if dm.ImageExists(cmd.Context(), containerConfig.Image) {
			fmt.Printf("🐳 Using local image: %s\n", containerConfig.Image)
		} else {
			// Image doesn't exist locally, try to pull it
			fmt.Println("🐳 Pulling container image...")
			if err := dm.PullImage(cmd.Context(), containerConfig.Image); err != nil {
				return fmt.Errorf("failed to pull image: %w", err)
			}
		}
	} else {
		fmt.Printf("🐳 Using prebuilt test image: %s\n", containerConfig.Image)
	}

	// Step 5: Create container
	fmt.Printf("🧭 Mounting host downloads dir: %s -> /home/builder/downloads\n", containerConfig.DownloadsDir)
	fmt.Println("🐳 Creating container...")
	containerID, err := dm.CreateContainer(cmd.Context(), containerConfig)
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
		stopErr := dm.StopContainer(cmd.Context(), containerID, 2*time.Second)
		if stopErr != nil {
			fmt.Printf("⚠️  Failed to stop container: %v\n", stopErr)
		}
		removeErr := dm.RemoveContainer(cmd.Context(), containerID, true)
		if removeErr != nil {
			fmt.Printf("⚠️  Failed to remove container: %v\n", removeErr)
		}
		os.Stdout.Sync() // Flush again after cleanup
	}()

	// Step 7: Start container
	fmt.Println("🐳 Starting container...")
	if err := dm.StartContainer(cmd.Context(), containerID); err != nil {
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

		executor := bitbake.NewBuildExecutor(cfg, dm, containerID, containerConfig.WorkspaceDir)
		buildResult, err := executor.ExecuteBuild(cmd.Context())

		if err != nil {
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
	} else {
		// Test mode - run marker validation logic
		fmt.Println("🚀 Build process running in test mode (SMIDR_TEST_WRITE_MARKERS=1)")
		if err := runTestMarkerValidation(cmd.Context(), dm, containerID, containerConfig); err != nil {
			return fmt.Errorf("test marker validation failed: %w", err)
		}
	}

	// Step 9: Generate build files (unchanged)
	generator := bitbake.NewGenerator(cfg, "./build")
	if err := generator.Generate(); err != nil {
		return fmt.Errorf("error generating build files: %w", err)
	}
	fmt.Println("✅ Build files generated successfully")
	fmt.Println("💡 Use 'smidr artifacts list' to view build artifacts once available")

	return nil // Build completed successfully
}

// getTailString returns the last n characters of a string
func getTailString(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

// runTestMarkerValidation runs the test marker validation logic
func runTestMarkerValidation(ctx context.Context, dm *docker.DockerManager, containerID string, containerConfig container.ContainerConfig) error {
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

	// Probe layer visibility if any provided
	if len(containerConfig.LayerDirs) > 0 {
		// Attempt to list first layer directory
		res, _ := dm.Exec(ctx, containerID, []string{"sh", "-c", "ls -la /home/builder/layers/layer-0 || true"}, 5*time.Second)
		if len(res.Stdout) > 0 {
			fmt.Print(string(res.Stdout))
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
	if cfg.Cache.Downloads == "" {
		if homedir, err := os.UserHomeDir(); err == nil {
			cfg.Cache.Downloads = fmt.Sprintf("%s/.smidr/downloads", homedir)
		} else {
			cfg.Cache.Downloads = fmt.Sprintf("%s/sources", workDir)
		}
	}
	if cfg.Cache.SState == "" {
		if homedir, err := os.UserHomeDir(); err == nil {
			cfg.Cache.SState = fmt.Sprintf("%s/.smidr/sstate-cache", homedir)
		} else {
			cfg.Cache.SState = fmt.Sprintf("%s/sstate-cache", workDir)
		}
	}
}

// isatty returns true if the given file descriptor is a terminal.
// Uses golang.org/x/term for cross-platform detection.
func isatty(fd uintptr) bool {
	return term.IsTerminal(int(fd))
}
