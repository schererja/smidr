package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// artifactsCmd represents the artifacts command
var artifactsCmd = &cobra.Command{
	Use:   "artifacts",
	Short: "Manage build artifacts",
	Long: `List, download, or manage build artifacts generated
by successful builds.`,
}

// listCmd represents the artifacts list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available artifacts",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📦 Available Artifacts")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━")

		// TODO: Implement actual artifact listing
		fmt.Println("🔍 No artifacts found")
		fmt.Println("💡 Run 'smidr build' to generate artifacts")
	},
}

func init() {
	rootCmd.AddCommand(artifactsCmd)
	artifactsCmd.AddCommand(listCmd)
}
