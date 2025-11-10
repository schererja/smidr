package artifacts

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	artifactsmgr "github.com/schererja/smidr/internal/artifacts"
	"github.com/spf13/cobra"
)

// artifactsCmd represents the artifacts command
var artifactsCmd = &cobra.Command{
	Use:   "artifacts",
	Short: "Manage build artifacts (list, copy, clean, show)",
	Long: `Manage build artifacts including listing, copying, and cleaning up old builds.

	Subcommands:
		list                List all builds
		copy <build-id>     Copy artifacts to current directory or destination
		clean               Clean up old builds (retention policy)
		show <build-id>     Show detailed build information

	Global Flags:
		--customer <name>   Filter artifacts for a specific customer/project

	Examples:
		smidr artifacts list --customer acme
		smidr artifacts copy core-image-minimal-20251016-022334 ./outdir --customer acme
		smidr artifacts clean --keep 5 --days 14 --dry-run
		smidr artifacts show core-image-minimal-20251016-022334 --customer acme
	`,
}

var artifactsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all build artifacts",
	Long: `List all available build artifacts with metadata including:
		- Build ID and timestamp
		- Project name and user
		- Build duration and status
		- Artifact sizes

	Flags:
		--customer <name>   Filter artifacts for a specific customer/project

	Example:
		smidr artifacts list --customer acme
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runArtifactsList(cmd); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var artifactsCopyCmd = &cobra.Command{
	Use:   "copy <build-id> [destination]",
	Short: "Copy artifacts from a specific build",
	Long: `Copy build artifacts from a specific build to a destination directory.
	If no destination is provided, artifacts are copied to the current directory.

	Flags:
		--customer <name>   Filter artifacts for a specific customer/project

	Examples:
		smidr artifacts copy core-image-minimal-20251016-022334 ./outdir --customer acme
		smidr artifacts copy core-image-minimal-20251016-022334
	`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		buildID := args[0]
		destination := "."
		if len(args) > 1 {
			destination = args[1]
		}

		if err := runArtifactsCopy(cmd, buildID, destination); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var artifactsCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up old build artifacts",
	Long: `Clean up old build artifacts based on retention policies.
	Default policy: keep last 10 builds and delete builds older than 30 days.

	Flags:
		--customer <name>   Filter artifacts for a specific customer/project
		--keep, -k <n>      Number of recent builds to keep (default 10)
		--days, -d <n>      Delete builds older than this many days (default 30)
		--dry-run, -n       Show what would be deleted without actually deleting

	Examples:
		smidr artifacts clean --keep 5 --days 14 --dry-run --customer acme
		smidr artifacts clean
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runArtifactsClean(cmd); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var artifactsShowCmd = &cobra.Command{
	Use:   "show <build-id>",
	Short: "Show detailed information about a specific build",
	Long: `Display detailed information about a specific build including metadata and artifact listing.

	Flags:
		--customer <name>   Filter artifacts for a specific customer/project

	Example:
		smidr artifacts show core-image-minimal-20251016-022334 --customer acme
	`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		buildID := args[0]
		if err := runArtifactsShow(cmd, buildID); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// New returns the artifacts command for registration with the root command
func New() *cobra.Command {
	artifactsCmd.AddCommand(artifactsListCmd)
	artifactsCmd.AddCommand(artifactsCopyCmd)
	artifactsCmd.AddCommand(artifactsCleanCmd)
	artifactsCmd.AddCommand(artifactsShowCmd)

	// Add --customer flag to all artifact subcommands
	artifactsListCmd.Flags().String("customer", "", "Filter artifacts for a specific customer")
	artifactsCopyCmd.Flags().String("customer", "", "Filter artifacts for a specific customer")
	artifactsCleanCmd.Flags().String("customer", "", "Filter artifacts for a specific customer")
	artifactsShowCmd.Flags().String("customer", "", "Filter artifacts for a specific customer")

	// Flags for clean command
	artifactsCleanCmd.Flags().IntP("keep", "k", 10, "Number of recent builds to keep")
	artifactsCleanCmd.Flags().IntP("days", "d", 30, "Delete builds older than this many days")
	artifactsCleanCmd.Flags().BoolP("dry-run", "n", false, "Show what would be deleted without actually deleting")

	return artifactsCmd
}

func runArtifactsList(cmd *cobra.Command) error {
	customer, _ := cmd.Flags().GetString("customer")
	var baseDir string
	if customer != "" {
		usr, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home dir: %w", err)
		}
		baseDir = filepath.Join(usr, ".smidr", "artifacts", "artifact-"+customer)
	} else {
		baseDir = ""
	}
	am, err := artifactsmgr.NewArtifactManager(baseDir)
	if err != nil {
		return fmt.Errorf("failed to create artifact manager: %w", err)
	}

	builds, err := am.ListBuilds()
	if err != nil {
		return fmt.Errorf("failed to list builds: %w", err)
	}

	if len(builds) == 0 {
		fmt.Println("No builds found")
		return nil
	}

	fmt.Printf("Found %d builds:\n\n", len(builds))

	for _, build := range builds {
		fmt.Printf("� Build ID: %s\n", build.BuildID)
		fmt.Printf("   Project: %s\n", build.ProjectName)
		fmt.Printf("   User: %s\n", build.User)
		fmt.Printf("   Timestamp: %s\n", build.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("   Duration: %v\n", build.BuildDuration)
		fmt.Printf("   Status: %s\n", build.Status)
		fmt.Printf("   Target: %s\n", build.TargetImage)

		// List actual artifact files
		artifactList, err := am.ListArtifacts(build.BuildID)
		if err != nil {
			fmt.Printf("   Artifacts: [error reading files: %v]\n", err)
		} else if len(artifactList) > 0 {
			// Calculate total size if possible
			var totalSize int64
			for _, name := range artifactList {
				if size, ok := build.ArtifactSizes[name]; ok {
					totalSize += size
				} else {
					// Try to stat the file if not in metadata
					fpath := filepath.Join(am.GetArtifactPath(build.BuildID), name)
					if fi, err := os.Stat(fpath); err == nil {
						totalSize += fi.Size()
					}
				}
			}
			fmt.Printf("   Artifacts: %d files (%s)\n", len(artifactList), artifactsmgr.FormatSize(totalSize))
		} else {
			fmt.Printf("   Artifacts: No artifacts found\n")
		}
		fmt.Println()
	}

	return nil
}

func runArtifactsCopy(cmd *cobra.Command, buildID, destination string) error {
	customer, _ := cmd.Flags().GetString("customer")
	var baseDir string
	if customer != "" {
		usr, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home dir: %w", err)
		}
		baseDir = filepath.Join(usr, ".smidr", "artifacts", "artifact-"+customer)
	} else {
		baseDir = ""
	}
	am, err := artifactsmgr.NewArtifactManager(baseDir)
	if err != nil {
		return fmt.Errorf("failed to create artifact manager: %w", err)
	}

	// Check if build exists
	metadata, err := am.LoadMetadata(buildID)
	if err != nil {
		return fmt.Errorf("build %s not found: %w", buildID, err)
	}

	// List artifacts for this build
	artifactList, err := am.ListArtifacts(buildID)
	if err != nil {
		return fmt.Errorf("failed to list artifacts: %w", err)
	}

	if len(artifactList) == 0 {
		fmt.Printf("No artifacts found for build %s\n", buildID)
		return nil
	}

	// Create destination directory
	destDir := filepath.Join(destination, fmt.Sprintf("smidr-artifacts-%s", buildID))
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	fmt.Printf("Copying %d artifacts from build %s to %s\n", len(artifactList), buildID, destDir)
	fmt.Printf("Build: %s (%s)\n", metadata.ProjectName, metadata.Timestamp.Format("2006-01-02 15:04:05"))

	// Copy each artifact
	for _, artifactName := range artifactList {
		destPath := filepath.Join(destDir, artifactName)
		if err := am.CopyArtifact(buildID, artifactName, destPath); err != nil {
			fmt.Printf("[WARNING] Failed to copy %s: %v\n", artifactName, err)
			continue
		}
		fmt.Printf("  ✓ %s\n", artifactName)
	}

	fmt.Printf("\n✅ Artifacts copied successfully to %s\n", destDir)
	return nil
}

func runArtifactsClean(cmd *cobra.Command) error {
	customer, _ := cmd.Flags().GetString("customer")
	var baseDir string
	if customer != "" {
		usr, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home dir: %w", err)
		}
		baseDir = filepath.Join(usr, ".smidr", "artifacts", "artifact-"+customer)
	} else {
		baseDir = ""
	}
	am, err := artifactsmgr.NewArtifactManager(baseDir)
	if err != nil {
		return fmt.Errorf("failed to create artifact manager: %w", err)
	}

	// Get flags
	keepLast, _ := cmd.Flags().GetInt("keep")
	maxDays, _ := cmd.Flags().GetInt("days")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	policy := artifactsmgr.RetentionPolicy{
		KeepLastN: keepLast,
		MaxAge:    time.Duration(maxDays) * 24 * time.Hour,
	}

	if dryRun {
		fmt.Printf("� Dry run mode - showing what would be deleted\n")
		fmt.Printf("Retention policy: keep last %d builds, delete builds older than %d days\n\n", keepLast, maxDays)

		// For dry run, we'd need to implement a preview version of cleanup
		// For now, just list current builds
		builds, err := am.ListBuilds()
		if err != nil {
			return fmt.Errorf("failed to list builds: %w", err)
		}

		fmt.Printf("Current builds: %d\n", len(builds))
		if len(builds) > keepLast {
			fmt.Printf("Would delete %d builds based on count policy\n", len(builds)-keepLast)
		}

		// Check age policy
		now := time.Now()
		oldCount := 0
		for _, build := range builds {
			if now.Sub(build.Timestamp) > policy.MaxAge {
				oldCount++
			}
		}
		if oldCount > 0 {
			fmt.Printf("Would delete %d builds based on age policy\n", oldCount)
		}

		return nil
	}

	fmt.Printf("🧹 Cleaning up artifacts with policy: keep last %d builds, delete builds older than %d days\n", keepLast, maxDays)

	if err := am.CleanupArtifacts(policy); err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	return nil
}

func runArtifactsShow(cmd *cobra.Command, buildID string) error {
	customer, _ := cmd.Flags().GetString("customer")
	var baseDir string
	if customer != "" {
		usr, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home dir: %w", err)
		}
		baseDir = filepath.Join(usr, ".smidr", "artifacts", "artifact-"+customer)
	} else {
		baseDir = ""
	}
	am, err := artifactsmgr.NewArtifactManager(baseDir)
	if err != nil {
		return fmt.Errorf("failed to create artifact manager: %w", err)
	}

	// Load metadata
	metadata, err := am.LoadMetadata(buildID)
	if err != nil {
		return fmt.Errorf("build %s not found: %w", buildID, err)
	}

	// List artifacts
	artifactList, err := am.ListArtifacts(buildID)
	if err != nil {
		return fmt.Errorf("failed to list artifacts: %w", err)
	}

	fmt.Printf("🔨 Build Details: %s\n", buildID)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Project:      %s\n", metadata.ProjectName)
	fmt.Printf("User:         %s\n", metadata.User)
	fmt.Printf("Timestamp:    %s\n", metadata.Timestamp.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("Duration:     %v\n", metadata.BuildDuration)
	fmt.Printf("Status:       %s\n", metadata.Status)
	fmt.Printf("Target Image: %s\n", metadata.TargetImage)

	fmt.Printf("\n📋 Configuration Used:\n")
	for key, value := range metadata.ConfigUsed {
		fmt.Printf("  %s: %s\n", key, value)
	}

	fmt.Printf("\n📦 Artifacts (%d files):\n", len(artifactList))
	if len(artifactList) == 0 {
		fmt.Printf("  No artifacts found\n")
	} else {
		var totalSize int64
		for _, artifactName := range artifactList {
			var size int64
			if s, ok := metadata.ArtifactSizes[artifactName]; ok {
				size = s
			} else {
				// Try to stat the file if not in metadata
				fpath := filepath.Join(am.GetArtifactPath(buildID), artifactName)
				if fi, err := os.Stat(fpath); err == nil {
					size = fi.Size()
				}
			}
			totalSize += size
			fmt.Printf("  %-40s %s\n", artifactName, artifactsmgr.FormatSize(size))
		}
		fmt.Printf("\nTotal size: %s\n", artifactsmgr.FormatSize(totalSize))

		artifactPath := am.GetArtifactPath(buildID)
		fmt.Printf("Location: %s\n", artifactPath)
	}

	return nil
}
