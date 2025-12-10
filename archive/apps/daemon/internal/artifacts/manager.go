package artifacts

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

// BuildMetadata contains information about a build
type BuildMetadata struct {
	BuildID       string            `json:"build_id"`
	ProjectName   string            `json:"project_name"`
	User          string            `json:"user"`
	Timestamp     time.Time         `json:"timestamp"`
	ConfigUsed    map[string]string `json:"config_used"`
	BuildDuration time.Duration     `json:"build_duration"`
	TargetImage   string            `json:"target_image"`
	ArtifactSizes map[string]int64  `json:"artifact_sizes"`
	Status        string            `json:"status"` // "success", "failed", "in-progress"
}

// ArtifactManager handles artifact storage and management
type ArtifactManager struct {
	baseDir string // Base directory for all artifacts (e.g., ~/.smidr/artifacts)
}

// NewArtifactManager creates a new artifact manager
func NewArtifactManager(baseDir string) (*ArtifactManager, error) {
	if baseDir == "" {
		// Default to ~/.smidr/artifacts
		usr, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("failed to get current user: %w", err)
		}
		baseDir = filepath.Join(usr.HomeDir, ".smidr", "artifacts")
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create artifacts directory %s: %w", baseDir, err)
	}

	return &ArtifactManager{
		baseDir: baseDir,
	}, nil
}

// GetArtifactPath returns the path for storing artifacts for a specific build
func (am *ArtifactManager) GetArtifactPath(buildID string) string {
	return filepath.Join(am.baseDir, buildID)
}

// GetMetadataPath returns the path for storing build metadata
func (am *ArtifactManager) GetMetadataPath(buildID string) string {
	return filepath.Join(am.GetArtifactPath(buildID), "build-metadata.json")
}

// SaveMetadata saves build metadata to disk
func (am *ArtifactManager) SaveMetadata(metadata BuildMetadata) error {
	metadataPath := am.GetMetadataPath(metadata.BuildID)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(metadataPath), 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	// Write metadata as JSON
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// LoadMetadata loads build metadata from disk
func (am *ArtifactManager) LoadMetadata(buildID string) (*BuildMetadata, error) {
	metadataPath := am.GetMetadataPath(buildID)

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata BuildMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &metadata, nil
}

// ListBuilds returns a list of all builds in the artifact store
func (am *ArtifactManager) ListBuilds() ([]BuildMetadata, error) {
	var builds []BuildMetadata

	// Recursively walk the artifact base directory to find all build-metadata.json files
	err := filepath.Walk(am.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip unreadable files/dirs
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if info.Name() == "build-metadata.json" {
			buildDir := filepath.Dir(path)
			buildID := filepath.Base(buildDir)
			metadata, err := am.LoadMetadata(buildID)
			if err != nil {
				// Skip builds with invalid metadata
				return nil
			}
			builds = append(builds, *metadata)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk artifacts directory: %w", err)
	}
	return builds, nil
}

// GenerateBuildID generates a unique build ID
func GenerateBuildID(projectName, user string) string {
	timestamp := time.Now().Format("20060102-150405.000")
	return fmt.Sprintf("%s-%s-%s", user, projectName, timestamp)
}

// ContainerExtractor interface for extracting files from containers
type ContainerExtractor interface {
	CopyFromContainer(ctx context.Context, containerID, containerPath, hostPath string) error
}
