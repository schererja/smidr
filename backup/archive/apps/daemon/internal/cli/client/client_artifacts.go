package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/schererja/smidr/internal/client"
)

var clientArtifactsCmd = &cobra.Command{
	Use:   "artifacts <build-id>",
	Short: "List artifacts from a completed build",
	Long: `List all artifacts from a completed build.

This command shows all available artifacts including their names, paths,
sizes, and checksums from the specified build.

Examples:
  smidr client artifacts build-123
  smidr client artifacts build-123 --address remote-host:50051`,
	Args: cobra.ExactArgs(1),
	RunE: runClientArtifacts,
}

func runClientArtifacts(cmd *cobra.Command, args []string) error {
	buildID := args[0]

	c, err := client.NewClient(clientDaemonAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get artifacts list
	resp, err := c.ListArtifacts(ctx, buildID)
	if err != nil {
		return fmt.Errorf("failed to list artifacts: %w", err)
	}

	if len(resp.Artifacts) == 0 {
		fmt.Printf("No artifacts found for build %s\n", buildID)
		return nil
	}

	fmt.Printf("📦 Artifacts for build %s:\n\n", buildID)

	// Calculate max widths for formatting
	maxNameWidth := 4 // "Name"
	for _, artifact := range resp.Artifacts {
		if len(artifact.Name) > maxNameWidth {
			maxNameWidth = len(artifact.Name)
		}
	}

	// Print header
	fmt.Printf("%-*s %10s %s\n",
		maxNameWidth, "Name",
		"Size",
		"Checksum")
	fmt.Printf("%s %s %s\n",
		strings.Repeat("-", maxNameWidth),
		strings.Repeat("-", 10),
		strings.Repeat("-", 16))

	// Print artifacts
	for _, artifact := range resp.Artifacts {
		sizeStr := formatSize(artifact.SizeBytes)
		checksumStr := artifact.Checksum
		if len(checksumStr) > 16 {
			checksumStr = checksumStr[:16] + "..."
		}

		fmt.Printf("%-*s %10s %s\n",
			maxNameWidth, artifact.Name,
			sizeStr,
			checksumStr)
	}

	fmt.Printf("\n📊 Total: %d artifacts\n", len(resp.Artifacts))

	return nil
}

// formatSize formats bytes into human-readable format
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1fGB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.1fMB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1fKB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
