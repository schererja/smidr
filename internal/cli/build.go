package cli

import (
	"fmt"
	"os"

	"github.com/intrik8-labs/smidr/internal/bitbake"
	"github.com/intrik8-labs/smidr/internal/config"
	"github.com/spf13/cobra"
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
		fmt.Println("🔨 Starting Smidr build...")
		cfg, err := config.Load("smidr.yaml")
		if err != nil {
			fmt.Println("Error loading configuration:", err)
			os.Exit(1)
		}
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
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	// Build-specific flags
	buildCmd.Flags().BoolP("force", "f", false, "Force rebuild (ignore cache)")
	buildCmd.Flags().StringP("target", "t", "", "Override build target")
}
