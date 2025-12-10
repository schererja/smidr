package client

import (
	"context"
	"fmt"
	"io"

	"github.com/schererja/smidr/internal/client"
	"github.com/spf13/cobra"
)

var (
	logsBuildID string
	logsFollow  bool
)

var clientLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Stream logs from a build",
	Long: `Stream logs from a build running on the daemon.

Use --follow to continue streaming new logs as they are generated.

Examples:
  smidr client logs --build-id build-123
  smidr client logs --build-id build-123 --follow
  smidr client logs --build-id build-123 --follow --address remote-host:50051`,
	RunE: runClientLogs,
}

func init() {
	clientLogsCmd.Flags().StringVar(&logsBuildID, "build-id", "", "Build ID to stream logs from (required)")
	clientLogsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output")
	clientLogsCmd.MarkFlagRequired("build-id")
}

func runClientLogs(cmd *cobra.Command, args []string) error {
	c, err := client.NewClient(clientDaemonAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer c.Close()

	ctx := context.Background()

	stream, err := c.StreamLogs(ctx, logsBuildID, logsFollow)
	if err != nil {
		return fmt.Errorf("failed to stream logs: %w", err)
	}

	fmt.Printf("📜 Streaming logs for build %s...\n", logsBuildID)
	if logsFollow {
		fmt.Println("   (Press Ctrl+C to stop following)")
	}
	fmt.Println()

	for {
		logLine, err := stream.Recv()
		if err == io.EOF {
			// Stream finished
			break
		}
		if err != nil {
			return fmt.Errorf("error receiving logs: %w", err)
		}

		// Print the log line with stream prefix
		if logLine.Stream == "stderr" {
			fmt.Printf("[stderr] %s\n", logLine.Message)
		} else {
			fmt.Printf("%s\n", logLine.Message)
		}
	}

	if !logsFollow {
		fmt.Println("\n✅ All logs received")
	}

	return nil
}
