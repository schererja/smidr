package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/schererja/smidr/internal/client"
	"github.com/spf13/cobra"
)

var (
	cancelBuildID string
)

var clientCancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel a running build",
	Long: `Cancel a running build on the daemon.

Examples:
  smidr client cancel --build-id build-123
  smidr client cancel --build-id build-123 --address remote-host:50051`,
	RunE: runClientCancel,
}

func init() {
	clientCmd.AddCommand(clientCancelCmd)

	clientCancelCmd.Flags().StringVar(&cancelBuildID, "build-id", "", "Build ID to cancel (required)")
	clientCancelCmd.MarkFlagRequired("build-id")
}

func runClientCancel(cmd *cobra.Command, args []string) error {
	c, err := client.NewClient(clientDaemonAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := c.CancelBuild(ctx, cancelBuildID)
	if err != nil {
		return fmt.Errorf("failed to cancel build: %w", err)
	}

	if result.Success {
		fmt.Printf("✅ Build %s cancelled successfully\n", cancelBuildID)
		fmt.Printf("   %s\n", result.Message)
	} else {
		fmt.Printf("❌ Failed to cancel build %s\n", cancelBuildID)
		fmt.Printf("   %s\n", result.Message)
	}

	return nil
}
