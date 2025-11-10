package logs

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs [build-id] [image]",
	Short: "Show build logs for a specific build (plain text or JSONL)",
	Long: `Show build logs for a specific build. If no build ID is given, shows the most recent build log.

	By default, shows the plain text log. Use --json to show structured JSONL log entries.
	If multiple images exist for a build, specify the image name as the second argument.

	Flags:
		--json               Show structured JSONL log entries
		--customer <name>    Customer/project name for artifact scoping

	Examples:
		smidr logs
		smidr logs core-image-minimal-20251016-022334 --customer acme
		smidr logs --json
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
		jsonMode, _ := cmd.Flags().GetBool("json")
		customer, _ := cmd.Flags().GetString("customer")
		if err := showBuildLogs(buildID, imageName, customer, jsonMode); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

// New returns the logs command for registration with the root command
func New() *cobra.Command {
	logsCmd.Flags().Bool("json", false, "Show structured JSONL log entries")
	logsCmd.Flags().String("customer", "", "Customer name to narrow search to artifact-<customer> directory")
	return logsCmd
}

// Update showBuildLogs to accept customer
func showBuildLogs(buildID, imageName, customer string, jsonMode bool) error {
	homedir, _ := os.UserHomeDir()
	artifactsDir := filepath.Join(homedir, ".smidr", "artifacts")
	var searchBase string
	if customer != "" {
		searchBase = filepath.Join(artifactsDir, "artifact-"+customer)
	} else {
		searchBase = artifactsDir
	}
	var artifactDirs []string
	if buildID != "" {
		// Find all artifact directories matching the buildID (may be nested)
		err := filepath.Walk(searchBase, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() && strings.Contains(path, buildID) {
				artifactDirs = append(artifactDirs, path)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("could not search artifacts: %w", err)
		}
		if len(artifactDirs) == 0 {
			return fmt.Errorf("no artifact directories found for build ID: %s", buildID)
		}
	} else {
		// Find most recent artifact dir
		var newest string
		var newestTime int64
		_ = filepath.Walk(searchBase, func(path string, info os.FileInfo, err error) error {
			if err != nil || !info.IsDir() {
				return nil
			}
			fi, err := os.Stat(path)
			if err != nil {
				return nil
			}
			if fi.ModTime().Unix() > newestTime {
				newest = path
				newestTime = fi.ModTime().Unix()
			}
			return nil
		})
		if newest == "" {
			return fmt.Errorf("no artifacts found")
		}
		artifactDirs = []string{newest}
	}

	// If multiple artifactDirs, require imageName or pick the most recent
	var targetDir string
	if imageName != "" {
		// Look for a subdir matching the image name
		found := false
		for _, dir := range artifactDirs {
			entries, _ := os.ReadDir(dir)
			for _, e := range entries {
				if e.IsDir() && strings.HasPrefix(e.Name(), imageName) {
					targetDir = filepath.Join(dir, e.Name())
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return fmt.Errorf("could not find image directory for image: %s", imageName)
		}
	} else if len(artifactDirs) == 1 {
		// If only one artifact dir, and only one image subdir, use it
		entries, _ := os.ReadDir(artifactDirs[0])
		var imageSubdirs []string
		for _, e := range entries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				imageSubdirs = append(imageSubdirs, e.Name())
			}
		}
		if len(imageSubdirs) == 1 {
			targetDir = filepath.Join(artifactDirs[0], imageSubdirs[0])
		} else {
			return fmt.Errorf("multiple image directories found in artifact: %s. Please specify the image name. Options: %v", artifactDirs[0], imageSubdirs)
		}
	} else {
		return fmt.Errorf("multiple artifact directories found for build ID: %s. Please specify the image name.", buildID)
	}

	var logPath string
	if jsonMode {
		logPath = filepath.Join(targetDir, "build-log.jsonl")
	} else {
		logPath = filepath.Join(targetDir, "build-log.txt")
	}
	f, err := os.Open(logPath)
	if err != nil {
		return fmt.Errorf("could not open log file: %w", err)
	}
	defer f.Close()
	if jsonMode {
		dec := json.NewDecoder(f)
		for {
			var entry map[string]interface{}
			if err := dec.Decode(&entry); err == io.EOF {
				break
			} else if err != nil {
				return fmt.Errorf("error decoding JSONL: %w", err)
			}
			b, _ := json.MarshalIndent(entry, "", "  ")
			fmt.Println(string(b))
		}
	} else {
		io.Copy(os.Stdout, f)
	}
	return nil
}
