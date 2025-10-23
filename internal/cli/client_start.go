package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/schererja/smidr/internal/client"
	"github.com/spf13/cobra"
)

var (
	startConfigPath string
	startTarget     string
	startForceClean bool
	startForceImage bool
)

var clientStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a new build on the daemon",
	Long: `Start a new build on the Smidr daemon.

This command submits a build request to the daemon and returns immediately
with a build ID that can be used to monitor the build status.

Examples:
  smidr client start --config config.yaml --target core-image-minimal
  smidr client start --config config.yaml --target core-image-minimal --force-clean
  smidr client start --address remote-host:50051 --config config.yaml --target my-image`,
	RunE: runClientStart,
}

func init() {
	clientCmd.AddCommand(clientStartCmd)
	
	clientStartCmd.Flags().StringVarP(&startConfigPath, "config", "c", "", "Path to config file (required)")
	clientStartCmd.Flags().StringVarP(&startTarget, "target", "t", "", "Build target/image name (required)")
	clientStartCmd.Flags().BoolVar(&startForceClean, "force-clean", false, "Force a clean build")
	clientStartCmd.Flags().BoolVar(&startForceImage, "force-image", false, "Force image regeneration only")
	
	clientStartCmd.MarkFlagRequired("config")
	clientStartCmd.MarkFlagRequired("target")
}

func runClientStart(cmd *cobra.Command, args []string) error {
	fmt.Printf("🔌 Connecting to daemon at %s...\n", clientDaemonAddress)
	
	c, err := client.NewClient(clientDaemonAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer c.Close()

	fmt.Println("✅ Connected to daemon")
	fmt.Printf("🚀 Starting build: %s\n", startTarget)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	status, err := c.StartBuild(ctx, startConfigPath, startTarget, startForceClean, startForceImage)
	if err != nil {
		return fmt.Errorf("failed to start build: %w", err)
	}

	fmt.Println("\n✅ Build started successfully!")
	fmt.Printf("📝 Build ID: %s\n", status.BuildId)
	fmt.Printf("🎯 Target: %s\n", status.Target)
	fmt.Printf("📊 State: %s\n", status.State)
	fmt.Printf("⏰ Started: %s\n", time.Unix(status.StartedAt, 0).Format(time.RFC3339))
	
	fmt.Printf("\n💡 Monitor the build with:\n")
	fmt.Printf("   smidr client status --build-id %s\n", status.BuildId)
	fmt.Printf("   smidr client logs --build-id %s --follow\n", status.BuildId)

	return nil
}
