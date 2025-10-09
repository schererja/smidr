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
	if cfg.Directories.Source == "" {
		cfg.Directories.Source = fmt.Sprintf("%s/sources", workDir)
	}
	if cfg.Directories.Build == "" {
		cfg.Directories.Build = fmt.Sprintf("%s/build", workDir)
	}

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
	containerConfig := container.ContainerConfig{
		Image:          "busybox:latest",                                        // TODO: Use from config or flag
		Cmd:            []string{"sh", "-c", "echo 'Container ready'; sleep 2"}, // Placeholder
		DownloadsDir:   cfg.Directories.Source,                                  // Example: mount sources as downloads
		SstateCacheDir: cfg.Directories.SState,                                  // Wire SSTATE dir from config
		WorkspaceDir:   cfg.Directories.Build,
		LayerDirs:      nil, // TODO: wire up if needed
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

	// Step 3: Create DockerManager
	dm, err := docker.NewDockerManager()
	if err != nil {
		return fmt.Errorf("failed to create DockerManager: %w", err)
	}

	// Step 4: Pull image
	fmt.Println("🐳 Pulling container image...")
	if err := dm.PullImage(cmd.Context(), containerConfig.Image); err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
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
			_, _ = dm.Exec(cmd.Context(), containerID, []string{"sh", "-c", "echo itest > /home/builder/downloads/itest.txt"}, 5*time.Second)
		}
		if containerConfig.WorkspaceDir != "" {
			_, _ = dm.Exec(cmd.Context(), containerID, []string{"sh", "-c", "echo itest > /home/builder/work/itest.txt"}, 5*time.Second)
		}
		if containerConfig.SstateCacheDir != "" {
			_, _ = dm.Exec(cmd.Context(), containerID, []string{"sh", "-c", "echo itest > /home/builder/sstate-cache/itest.txt"}, 5*time.Second)
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
