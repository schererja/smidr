package artifacts

import (
	"fmt"
	"os"
	"sort"
	"time"
)

// RetentionPolicy defines rules for cleaning up old artifacts
type RetentionPolicy struct {
	KeepLastN int           // Keep the last N builds (0 = no limit)
	MaxAge    time.Duration // Delete builds older than this (0 = no limit)
	MaxSizeGB int64         // Delete oldest builds when total size exceeds this (0 = no limit)
}

// DefaultRetentionPolicy returns a sensible default retention policy
func DefaultRetentionPolicy() RetentionPolicy {
	return RetentionPolicy{
		KeepLastN: 10,                  // Keep last 10 builds
		MaxAge:    30 * 24 * time.Hour, // Keep builds for 30 days
		MaxSizeGB: 0,                   // No size limit by default
	}
}

// CleanupArtifacts applies retention policies to clean up old artifacts
func (am *ArtifactManager) CleanupArtifacts(policy RetentionPolicy) error {
	builds, err := am.ListBuilds()
	if err != nil {
		return fmt.Errorf("failed to list builds for cleanup: %w", err)
	}

	if len(builds) == 0 {
		fmt.Println("No builds found to clean up")
		return nil
	}

	// Sort builds by timestamp (newest first)
	sort.Slice(builds, func(i, j int) bool {
		return builds[i].Timestamp.After(builds[j].Timestamp)
	})

	var toDelete []string
	now := time.Now()

	for i, build := range builds {
		shouldDelete := false

		// Check age policy
		if policy.MaxAge > 0 && now.Sub(build.Timestamp) > policy.MaxAge {
			shouldDelete = true
			fmt.Printf("Marking build %s for deletion (age: %v)\n",
				build.BuildID, now.Sub(build.Timestamp))
		}

		// Check count policy (keep last N)
		if policy.KeepLastN > 0 && i >= policy.KeepLastN {
			shouldDelete = true
			fmt.Printf("Marking build %s for deletion (keeping last %d builds)\n",
				build.BuildID, policy.KeepLastN)
		}

		if shouldDelete {
			toDelete = append(toDelete, build.BuildID)
		}
	}

	// TODO: Implement size-based cleanup if needed
	if policy.MaxSizeGB > 0 {
		fmt.Printf("[WARNING] Size-based cleanup not yet implemented (limit: %d GB)\n", policy.MaxSizeGB)
	}

	// Delete marked builds
	deletedCount := 0
	for _, buildID := range toDelete {
		if err := am.DeleteBuild(buildID); err != nil {
			fmt.Printf("[WARNING] Failed to delete build %s: %v\n", buildID, err)
			continue
		}
		deletedCount++
	}

	fmt.Printf("Cleanup complete: deleted %d builds\n", deletedCount)
	return nil
}

// DeleteBuild removes a specific build and all its artifacts
func (am *ArtifactManager) DeleteBuild(buildID string) error {
	buildPath := am.GetArtifactPath(buildID)

	if _, err := os.Stat(buildPath); os.IsNotExist(err) {
		return fmt.Errorf("build %s does not exist", buildID)
	}

	if err := os.RemoveAll(buildPath); err != nil {
		return fmt.Errorf("failed to remove build directory: %w", err)
	}

	fmt.Printf("Deleted build: %s\n", buildID)
	return nil
}

// GetTotalSize calculates the total size of all artifacts
func (am *ArtifactManager) GetTotalSize() (int64, error) {
	var totalSize int64

	builds, err := am.ListBuilds()
	if err != nil {
		return 0, err
	}

	for _, build := range builds {
		for _, size := range build.ArtifactSizes {
			totalSize += size
		}
	}

	return totalSize, nil
}

// FormatSize formats a size in bytes to human-readable format
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
