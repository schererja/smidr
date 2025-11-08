package artifacts

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// ExtractArtifacts extracts build artifacts from a container to the artifact storage
func (am *ArtifactManager) ExtractArtifacts(ctx context.Context, extractor ContainerExtractor, containerID string, metadata BuildMetadata) error {
	artifactPath := am.GetArtifactPath(metadata.BuildID)

	// Create artifact directory
	if err := os.MkdirAll(artifactPath, 0755); err != nil {
		return fmt.Errorf("failed to create artifact directory: %w", err)
	}

	// Define container paths to extract
	containerPaths := []string{
		"/home/builder/build/deploy/images/qemux86-64", // Actual deploy images directory
	}

	// Extract artifacts from each container path
	for _, containerPath := range containerPaths {
		// Determine the destination based on container path
		destName := "deploy-images-qemux86-64"
		destPath := filepath.Join(artifactPath, destName)

		// Extract from container
		if err := extractor.CopyFromContainer(ctx, containerID, containerPath, destPath); err != nil {
			// Log warning but don't fail the entire extraction if one path fails
			fmt.Printf("[WARNING] Failed to extract %s: %v\n", containerPath, err)
			continue
		}

		fmt.Printf("✓ Extracted artifacts from %s to %s\n", containerPath, destPath)
	}

	// Calculate artifact sizes
	metadata.ArtifactSizes = make(map[string]int64)
	if err := am.calculateArtifactSizes(artifactPath, &metadata); err != nil {
		fmt.Printf("[WARNING] Failed to calculate artifact sizes: %v\n", err)
	}

	// Save metadata
	if err := am.SaveMetadata(metadata); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}

// calculateArtifactSizes recursively calculates sizes of extracted artifacts
func (am *ArtifactManager) calculateArtifactSizes(artifactPath string, metadata *BuildMetadata) error {
	return filepath.Walk(artifactPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relPath, err := filepath.Rel(artifactPath, path)
			if err != nil {
				return err
			}

			// Skip the metadata file itself
			if relPath == "build-metadata.json" {
				return nil
			}

			metadata.ArtifactSizes[relPath] = info.Size()
		}

		return nil
	})
}

// ListArtifacts returns a list of artifact files for a specific build
func (am *ArtifactManager) ListArtifacts(buildID string) ([]string, error) {
	artifactPath := am.GetArtifactPath(buildID)

	var artifacts []string
	err := filepath.Walk(artifactPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(artifactPath, path)
		if err != nil {
			return err
		}
		// Skip the metadata file (at any depth)
		if info.Name() == "build-metadata.json" {
			return nil
		}
		artifacts = append(artifacts, relPath)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list artifacts: %w", err)
	}
	return artifacts, nil
}

// CopyArtifact copies a specific artifact to a destination path
func (am *ArtifactManager) CopyArtifact(buildID, artifactName, destPath string) error {
	sourcePath := filepath.Join(am.GetArtifactPath(buildID), artifactName)

	// Check if source exists
	if _, err := os.Stat(sourcePath); err != nil {
		return fmt.Errorf("artifact %s not found in build %s: %w", artifactName, buildID, err)
	}

	// Create destination directory if needed
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Copy file
	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source artifact: %w", err)
	}

	if err := os.WriteFile(destPath, sourceData, 0644); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}
