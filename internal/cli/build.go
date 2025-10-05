package cli

import (
	"fmt"
	"os"

	"github.com/intrik8-labs/smidr/internal/bitbake"
	"github.com/intrik8-labs/smidr/internal/config"
	"github.com/intrik8-labs/smidr/internal/source"
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
		err := runBuild(cmd)
		if err != nil {
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
		fmt.Println("Error loading configuration:", err)
		os.Exit(1)
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
	generator := bitbake.NewGenerator(cfg, "./build")
	if err := generator.Generate(); err != nil {
		fmt.Println("Error generating build files:", err)
		os.Exit(1)
	}
	fmt.Println("✅ Build files generated successfully")
	fmt.Println("🚀 Build process would start here (not yet implemented)")
	fmt.Println("💡 Use 'smidr artifacts list' to view build artifacts once available")
	os.Exit(1)
	return nil
}
