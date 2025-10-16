package artifacts

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// MockContainerExtractor for testing
type MockContainerExtractor struct {
	extractedPaths map[string]string // containerPath -> hostPath mapping
}

func (m *MockContainerExtractor) CopyFromContainer(ctx context.Context, containerID, containerPath, hostPath string) error {
	m.extractedPaths[containerPath] = hostPath

	// Create a dummy file to simulate extraction
	if err := os.MkdirAll(hostPath, 0755); err != nil {
		return err
	}

	dummyFile := filepath.Join(hostPath, "dummy-artifact.img")
	return os.WriteFile(dummyFile, []byte("dummy artifact content"), 0644)
}

func TestArtifactManager_NewArtifactManager(t *testing.T) {
	tmpDir := t.TempDir()

	am, err := NewArtifactManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create artifact manager: %v", err)
	}

	if am.baseDir != tmpDir {
		t.Errorf("Expected base dir %s, got %s", tmpDir, am.baseDir)
	}

	// Verify directory was created
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Errorf("Artifact directory was not created")
	}
}

func TestArtifactManager_SaveAndLoadMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	am, err := NewArtifactManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create artifact manager: %v", err)
	}

	metadata := BuildMetadata{
		BuildID:       "test-user-myproject-20231014-120000",
		ProjectName:   "myproject",
		User:          "test-user",
		Timestamp:     time.Now(),
		ConfigUsed:    map[string]string{"machine": "qemux86-64"},
		BuildDuration: 30 * time.Minute,
		TargetImage:   "core-image-minimal",
		Status:        "success",
	}

	// Save metadata
	if err := am.SaveMetadata(metadata); err != nil {
		t.Fatalf("Failed to save metadata: %v", err)
	}

	// Load metadata
	loadedMetadata, err := am.LoadMetadata(metadata.BuildID)
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	// Verify loaded metadata
	if loadedMetadata.BuildID != metadata.BuildID {
		t.Errorf("Expected BuildID %s, got %s", metadata.BuildID, loadedMetadata.BuildID)
	}
	if loadedMetadata.ProjectName != metadata.ProjectName {
		t.Errorf("Expected ProjectName %s, got %s", metadata.ProjectName, loadedMetadata.ProjectName)
	}
	if loadedMetadata.Status != metadata.Status {
		t.Errorf("Expected Status %s, got %s", metadata.Status, loadedMetadata.Status)
	}
}

func TestArtifactManager_ExtractArtifacts(t *testing.T) {
	tmpDir := t.TempDir()
	am, err := NewArtifactManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create artifact manager: %v", err)
	}

	mockExtractor := &MockContainerExtractor{
		extractedPaths: make(map[string]string),
	}

	metadata := BuildMetadata{
		BuildID:     "test-extraction-20231014-120000",
		ProjectName: "test-project",
		User:        "test-user",
		Timestamp:   time.Now(),
		Status:      "success",
	}

	ctx := context.Background()
	containerID := "test-container"

	// Extract artifacts
	if err := am.ExtractArtifacts(ctx, mockExtractor, containerID, metadata); err != nil {
		t.Fatalf("Failed to extract artifacts: %v", err)
	}

	// Verify extraction was called
	expectedPath := "/home/builder/work/deploy"
	if _, exists := mockExtractor.extractedPaths[expectedPath]; !exists {
		t.Errorf("Expected container path %s to be extracted", expectedPath)
	}

	// Verify artifact file exists
	artifactPath := am.GetArtifactPath(metadata.BuildID)
	dummyFile := filepath.Join(artifactPath, "deploy", "dummy-artifact.img")
	if _, err := os.Stat(dummyFile); os.IsNotExist(err) {
		t.Errorf("Expected artifact file %s to exist", dummyFile)
	}

	// Verify metadata was saved
	loadedMetadata, err := am.LoadMetadata(metadata.BuildID)
	if err != nil {
		t.Fatalf("Failed to load saved metadata: %v", err)
	}

	if len(loadedMetadata.ArtifactSizes) == 0 {
		t.Errorf("Expected artifact sizes to be calculated")
	}
}

func TestGenerateBuildID(t *testing.T) {
	projectName := "my-project"
	user := "testuser"

	buildID := GenerateBuildID(projectName, user)

	// Should contain user, project name, and timestamp
	if !contains(buildID, user) {
		t.Errorf("Build ID should contain user: %s", buildID)
	}
	if !contains(buildID, projectName) {
		t.Errorf("Build ID should contain project name: %s", buildID)
	}

	// Small delay to ensure unique timestamp
	time.Sleep(1 * time.Millisecond)

	// Should be unique (generate two and compare)
	buildID2 := GenerateBuildID(projectName, user)
	if buildID == buildID2 {
		t.Errorf("Build IDs should be unique, got same ID twice: %s", buildID)
	}
}

func TestFormatSize(t *testing.T) {
	testCases := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tc := range testCases {
		result := FormatSize(tc.bytes)
		if result != tc.expected {
			t.Errorf("FormatSize(%d) = %s, expected %s", tc.bytes, result, tc.expected)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsAt(s, substr))))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
