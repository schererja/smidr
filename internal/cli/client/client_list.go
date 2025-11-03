package client

import (
	"context"
	"fmt"
	"time"

	v1 "github.com/schererja/smidr/api/proto"
	"github.com/schererja/smidr/internal/client"
	"github.com/spf13/cobra"
)

var (
	listLimit int32
)

var clientListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all builds on the daemon",
	Long: `List all builds (active and completed) on the daemon.

Examples:
  smidr client list
  smidr client list --limit 10
  smidr client list --address remote-host:50051`,
	RunE: runClientList,
}

func init() {
	clientListCmd.Flags().Int32Var(&listLimit, "limit", 0, "Maximum number of builds to list (0 = all)")
}

func runClientList(cmd *cobra.Command, args []string) error {
	c, err := client.NewClient(clientDaemonAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	buildsList, err := c.ListBuilds(ctx, nil, listLimit)
	if err != nil {
		return fmt.Errorf("failed to list builds: %w", err)
	}

	if len(buildsList.Builds) == 0 {
		fmt.Println("No builds found")
		return nil
	}

	fmt.Printf("Found %d build(s):\n\n", len(buildsList.Builds))

	for _, build := range buildsList.Builds {
		fmt.Printf("🔹 Build ID: %s\n", build.BuildId)
		fmt.Printf("   Target: %s\n", build.Target)
		fmt.Printf("   State: %s\n", formatBuildState(build.State))
		fmt.Printf("   Started: %s\n", time.Unix(build.StartedAt, 0).Format(time.RFC3339))

		if build.CompletedAt > 0 {
			completedTime := time.Unix(build.CompletedAt, 0)
			duration := completedTime.Sub(time.Unix(build.StartedAt, 0))
			fmt.Printf("   Completed: %s (took %s)\n", completedTime.Format(time.RFC3339), duration.Round(time.Second))
			fmt.Printf("   Exit Code: %d\n", build.ExitCode)
		} else if build.State == v1.BuildState_BUILD_STATE_BUILDING {
			duration := time.Since(time.Unix(build.StartedAt, 0))
			fmt.Printf("   Duration: %s\n", duration.Round(time.Second))
		}

		if build.ErrorMessage != "" {
			fmt.Printf("   Error: %s\n", build.ErrorMessage)
		}

		fmt.Println()
	}

	return nil
}

func formatBuildState(state v1.BuildState) string {
	switch state {
	case v1.BuildState_BUILD_STATE_QUEUED:
		return "⏳ QUEUED"
	case v1.BuildState_BUILD_STATE_PREPARING:
		return "🔧 PREPARING"
	case v1.BuildState_BUILD_STATE_BUILDING:
		return "🔨 BUILDING"
	case v1.BuildState_BUILD_STATE_EXTRACTING_ARTIFACTS:
		return "📦 EXTRACTING"
	case v1.BuildState_BUILD_STATE_COMPLETED:
		return "✅ COMPLETED"
	case v1.BuildState_BUILD_STATE_FAILED:
		return "❌ FAILED"
	case v1.BuildState_BUILD_STATE_CANCELLED:
		return "🚫 CANCELLED"
	default:
		return "❓ UNKNOWN"
	}
}
