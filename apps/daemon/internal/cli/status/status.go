package status

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/schererja/smidr/internal/artifacts"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status [build-id] [image]",
	Short: "Show build status and artifact summary for a build",
	Long: `Show build status and artifact summary for a specific build. If no build ID is given, shows the most recent build status.

	By default, prints a summary from build-metadata.json. Use --list-artifacts to print a detailed artifact listing.

	Flags:
		--customer <name>     Customer/project name for artifact scoping
		--list-artifacts      List all artifact files and sizes in detail

	Examples:
		smidr status
		smidr status core-image-minimal-20251016-022334 --customer acme
		smidr status --list-artifacts
	`,
	Args: cobra.RangeArgs(0, 2),
	Run: func(cmd *cobra.Command, args []string) {
		buildID := ""
		imageName := ""
		if len(args) > 0 {
			buildID = args[0]
		}
		if len(args) > 1 {
			imageName = args[1]
		}
		customer, _ := cmd.Flags().GetString("customer")
		listArtifacts, _ := cmd.Flags().GetBool("list-artifacts")
		if err := showBuildStatus(buildID, imageName, customer, listArtifacts); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

// New returns the status command for registration with the root command
func New() *cobra.Command {
	statusCmd.Flags().String("customer", "", "Customer name to narrow search to artifact-<customer> directory")
	statusCmd.Flags().Bool("list-artifacts", false, "List all artifact files and sizes in detail")
	return statusCmd
}

func showBuildStatus(buildID, imageName, customer string, listArtifacts bool) error {

	// Use ArtifactManager and ListBuilds for artifact lookup (like artifacts list)
	var baseDir string
	if customer != "" {
		homedir, _ := os.UserHomeDir()
		baseDir = filepath.Join(homedir, ".smidr", "artifacts", "artifact-"+customer)
	} else {
		baseDir = ""
	}
	am, err := artifacts.NewArtifactManager(baseDir)
	if err != nil {
		return fmt.Errorf("failed to create artifact manager: %w", err)
	}

	builds, err := am.ListBuilds()
	if err != nil {
		return fmt.Errorf("failed to list builds: %w", err)
	}
	if len(builds) == 0 {
		return fmt.Errorf("no builds found")
	}

	// Find the build to show
	var meta *artifacts.BuildMetadata
	if buildID != "" {
		for _, b := range builds {
			if b.BuildID == buildID {
				meta = &b
				break
			}
		}
		if meta == nil {
			return fmt.Errorf("no build found with build ID: %s", buildID)
		}
	} else {
		// Show the most recent build (by timestamp)
		latest := builds[0]
		for _, b := range builds[1:] {
			if b.Timestamp.After(latest.Timestamp) {
				latest = b
			}
		}
		meta = &latest
	}

	// If metadata doesn't include artifact sizes, compute them from disk (align with artifacts list logic)
	if len(meta.ArtifactSizes) == 0 {
		artifactList, err := am.ListArtifacts(meta.BuildID)
		if err == nil && len(artifactList) > 0 {
			meta.ArtifactSizes = make(map[string]int64)
			for _, name := range artifactList {
				fpath := filepath.Join(am.GetArtifactPath(meta.BuildID), name)
				if fi, err := os.Stat(fpath); err == nil {
					meta.ArtifactSizes[name] = fi.Size()
				}
			}
		}
	}

	fmt.Println("\n📦 Build Status Summary")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("project_name    : %s\n", meta.ProjectName)
	fmt.Printf("user            : %s\n", meta.User)
	fmt.Printf("timestamp       : %s\n", meta.Timestamp.Format(time.RFC3339))
	fmt.Printf("config_used     : ")
	if len(meta.ConfigUsed) > 0 {
		keys := make([]string, 0, len(meta.ConfigUsed))
		for k := range meta.ConfigUsed {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, k := range keys {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("%s:%s", k, meta.ConfigUsed[k])
		}
		fmt.Println()
	} else {
		fmt.Println("none")
	}
	fmt.Printf("build_duration  : %s\n", meta.BuildDuration)
	fmt.Printf("build_id        : %s\n", meta.BuildID)
	fmt.Printf("target_image    : %s\n", meta.TargetImage)
	// Summarize artifacts rather than listing them all
	artifactList, err := am.ListArtifacts(meta.BuildID)
	if err != nil {
		fmt.Printf("artifacts       : error reading artifacts: %v\n", err)
	} else if len(artifactList) == 0 {
		fmt.Printf("artifacts       : none\n")
	} else {
		var totalSize int64
		// Map of artifact sizes for optional detailed listing
		resolvedSizes := make(map[string]int64, len(artifactList))
		for _, name := range artifactList {
			if size, ok := meta.ArtifactSizes[name]; ok {
				totalSize += size
				resolvedSizes[name] = size
			} else {
				fpath := filepath.Join(am.GetArtifactPath(meta.BuildID), name)
				if fi, err := os.Stat(fpath); err == nil {
					sz := fi.Size()
					totalSize += sz
					resolvedSizes[name] = sz
				}
			}
		}
		fmt.Printf("artifacts       : %d files (%s)\n", len(artifactList), artifacts.FormatSize(totalSize))

		// Optional detailed listing when requested
		if listArtifacts {
			fmt.Printf("\n📦 Artifacts (%d files):\n", len(artifactList))
			for _, name := range artifactList {
				size := resolvedSizes[name]
				fmt.Printf("  %-60s %s\n", name, artifacts.FormatSize(size))
			}
			fmt.Printf("\nTotal size: %s\n", artifacts.FormatSize(totalSize))
			fmt.Printf("Location: %s\n", am.GetArtifactPath(meta.BuildID))
		}
	}
	fmt.Printf("status          : %s\n", meta.Status)
	return nil
}
