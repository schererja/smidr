package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

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

	// Step 1: Fetch layers
	fmt.Println("📦 Fetching required layers...")
	fetcher := source.NewFetcher(cfg.Directories.Source, logger)
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

	containerConfig := container.ContainerConfig{
		Image: "busybox:latest", // TODO: Use from config or flag
		// Use a portable shell invocation. Using ["sh", "-c", "..."] runs the
		// command under /bin/sh when no ENTRYPOINT is set (busybox), and when an
		// image *does* set an ENTRYPOINT (for example /bin/bash) Docker will append
		// these args to the entrypoint; in practice bash will accept 'sh -c ...'
		// and run the given commands. Keep the container alive briefly so tests
		// can exec into it.
		// Provide a single string command; the Docker manager will use /bin/sh -c
		// as the Entrypoint so this will run consistently regardless of whether
		// the image defines its own ENTRYPOINT.
		Cmd:            []string{"echo 'Container ready'; sleep 5"},
		DownloadsDir:   cfg.Directories.Source, // Example: mount sources as downloads
		SstateCacheDir: cfg.Directories.SState, // Wire SSTATE dir from config
		WorkspaceDir:   cfg.Directories.Build,
		LayerDirs:      cfgLayerDirs,
		Name:           testName,
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

	// Step 4: Pull image (skip pulling if a test image override is provided)
	if testImage == "" {
		fmt.Println("🐳 Pulling container image...")
		if err := dm.PullImage(cmd.Context(), containerConfig.Image); err != nil {
			return fmt.Errorf("failed to pull image: %w", err)
		}
	} else {
		fmt.Printf("🐳 Using prebuilt test image: %s\n", containerConfig.Image)
	}

	// Step 5: Create container
	fmt.Println("🐳 Creating container...")
	containerID, err := dm.CreateContainer(cmd.Context(), containerConfig)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	// Step 6: Ensure cleanup (stop/remove) always runs
	defer func() {
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

	// Step 8: (Placeholder) Run build logic inside container here
	fmt.Println("🚀 Build process would start here (not yet implemented)")
	if os.Getenv("SMIDR_TEST_WRITE_MARKERS") == "1" {
		// Write marker files into mounts to validate wiring
		if containerConfig.DownloadsDir != "" {
			res, err := dm.Exec(cmd.Context(), containerID, []string{"sh", "-c", "echo itest > /home/builder/downloads/itest.txt"}, 5*time.Second)
			if err != nil {
				fmt.Printf("⚠️  Failed to write downloads marker: %v\n", err)
			}
			if res.ExitCode != 0 {
				fmt.Printf("⚠️  downloads marker exec failed: stdout=%s stderr=%s\n", string(res.Stdout), string(res.Stderr))
				// Run debug script to diagnose permission issues
				debugRes, _ := dm.Exec(cmd.Context(), containerID, []string{"/usr/local/bin/smidr-debug"}, 10*time.Second)
				fmt.Printf("🔍 Debug info after downloads marker failure:\n%s", string(debugRes.Stdout))
			}
		}
		if containerConfig.WorkspaceDir != "" {
			res, err := dm.Exec(cmd.Context(), containerID, []string{"sh", "-c", "echo itest > /home/builder/work/itest.txt"}, 5*time.Second)
			if err != nil {
				fmt.Printf("⚠️  Failed to write workspace marker: %v\n", err)
			}
			if res.ExitCode != 0 {
				fmt.Printf("⚠️  workspace marker exec failed: stdout=%s stderr=%s\n", string(res.Stdout), string(res.Stderr))
				// Run debug script to diagnose permission issues (only if downloads didn't already run it)
				if containerConfig.DownloadsDir == "" {
					debugRes, _ := dm.Exec(cmd.Context(), containerID, []string{"/usr/local/bin/smidr-debug"}, 10*time.Second)
					fmt.Printf("🔍 Debug info after workspace marker failure:\n%s", string(debugRes.Stdout))
				}
			}
		}
		if containerConfig.SstateCacheDir != "" {
			res, err := dm.Exec(cmd.Context(), containerID, []string{"sh", "-c", "echo itest > /home/builder/sstate-cache/itest.txt"}, 5*time.Second)
			if err != nil {
				fmt.Printf("⚠️  Failed to write sstate marker: %v\n", err)
			}
			if res.ExitCode != 0 {
				fmt.Printf("⚠️  sstate marker exec failed: stdout=%s stderr=%s\n", string(res.Stdout), string(res.Stderr))
				// Run debug script to diagnose permission issues (only if others didn't already run it)
				if containerConfig.DownloadsDir == "" && containerConfig.WorkspaceDir == "" {
					debugRes, _ := dm.Exec(cmd.Context(), containerID, []string{"/usr/local/bin/smidr-debug"}, 10*time.Second)
					fmt.Printf("🔍 Debug info after sstate marker failure:\n%s", string(debugRes.Stdout))
				}
			}
		}
		// Probe layer visibility if any provided
		if len(containerConfig.LayerDirs) > 0 {
			// Attempt to list first layer directory
			res, _ := dm.Exec(cmd.Context(), containerID, []string{"sh", "-c", "ls -la /home/builder/layers/layer-0 || true"}, 5*time.Second)
			if len(res.Stdout) > 0 {
				fmt.Print(string(res.Stdout))
			}
		}
	}

	// Step 9: Generate build files (unchanged)
	generator := bitbake.NewGenerator(cfg, "./build")
	if err := generator.Generate(); err != nil {
		return fmt.Errorf("error generating build files: %w", err)
	}
	fmt.Println("✅ Build files generated successfully")
	fmt.Println("💡 Use 'smidr artifacts list' to view build artifacts once available")
	// Return a sentinel error to keep current non-zero behavior while still flushing defers
	return fmt.Errorf("build step not yet implemented")
}

// setDefaultDirs ensures default directory paths are populated on the config
// when the user did not provide them. This centralizes defaulting and makes
// it testable.
func setDefaultDirs(cfg *config.Config, workDir string) {
	if cfg.Directories.Source == "" {
		cfg.Directories.Source = fmt.Sprintf("%s/sources", workDir)
	}
	if cfg.Directories.Build == "" {
		cfg.Directories.Build = fmt.Sprintf("%s/build", workDir)
	}
	if cfg.Directories.SState == "" {
		cfg.Directories.SState = fmt.Sprintf("%s/sstate-cache", workDir)
	}
}
