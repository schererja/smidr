package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show build status and information",
	Long: `Display the current status of builds, including:
- Active builds and their progress
- Recent build history
- Cache usage statistics
- Container information`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📊 Smidr Status")
		fmt.Println("━━━━━━━━━━━━━━━━")

		// TODO: Implement actual status logic
		fmt.Println("🚧 No active builds")
		fmt.Println("📁 Cache: Not initialized")
		fmt.Println("🐳 Containers: None running")
		fmt.Println("")
		fmt.Println("💡 Run 'smidr init' to create a project")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
