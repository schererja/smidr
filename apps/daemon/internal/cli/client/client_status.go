package client

import (
	"context"
	"fmt"
	"time"

	"github.com/schererja/smidr/internal/client"
	"github.com/spf13/cobra"
)

var (
	statusBuildID string
)

var clientStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get the status of a build",
	Long: `Get the current status of a build running on the daemon.

Examples:
  smidr client status --build-id build-123
  smidr client status --build-id build-123 --address remote-host:50051`,
	RunE: runClientStatus,
}

func init() {
	clientStatusCmd.Flags().StringVar(&statusBuildID, "build-id", "", "Build ID to check (required)")
	clientStatusCmd.MarkFlagRequired("build-id")
}

func runClientStatus(cmd *cobra.Command, args []string) error {
	c, err := client.NewClient(clientDaemonAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status, err := c.GetBuildStatus(ctx, statusBuildID)
	if err != nil {
		return fmt.Errorf("failed to get build status: %w", err)
	}

	// Print status
	fmt.Printf("📝 Build ID: %s\n", status.BuildId)
	fmt.Printf("🎯 Target: %s\n", status.Target)
	fmt.Printf("📊 State: %s\n", status.State)
	fmt.Printf("📄 Config: %s\n", status.ConfigPath)
	fmt.Printf("⏰ Started: %s\n", time.Unix(status.StartedAt, 0).Format(time.RFC3339))
	
	if status.CompletedAt > 0 {
		completedTime := time.Unix(status.CompletedAt, 0)
		duration := completedTime.Sub(time.Unix(status.StartedAt, 0))
		fmt.Printf("✅ Completed: %s (took %s)\n", completedTime.Format(time.RFC3339), duration.Round(time.Second))
		fmt.Printf("🔢 Exit Code: %d\n", status.ExitCode)
	}
	
	if status.ErrorMessage != "" {
		fmt.Printf("❌ Error: %s\n", status.ErrorMessage)
	}

	return nil
}
