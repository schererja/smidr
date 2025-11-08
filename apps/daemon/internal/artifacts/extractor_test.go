package artifacts

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListArtifactsAndCopyArtifact(t *testing.T) {
	tmpDir := t.TempDir()
	am, err := NewArtifactManager(tmpDir)
	if err != nil {
		t.Fatalf("failed to create ArtifactManager: %v", err)
	}
	buildID := "build1"
	buildPath := filepath.Join(tmpDir, buildID)
	os.MkdirAll(filepath.Join(buildPath, "subdir"), 0755)
	file1 := filepath.Join(buildPath, "file1.txt")
	file2 := filepath.Join(buildPath, "subdir", "file2.txt")
	os.WriteFile(file1, []byte("hello"), 0644)
	os.WriteFile(file2, []byte("world"), 0644)
	// Should list both files
	files, err := am.ListArtifacts(buildID)
	if err != nil {
		t.Fatalf("ListArtifacts failed: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 artifacts, got %d", len(files))
	}
	// Should copy file1
	dest := filepath.Join(tmpDir, "copied.txt")
	if err := am.CopyArtifact(buildID, "file1.txt", dest); err != nil {
		t.Errorf("CopyArtifact failed: %v", err)
	}
	if data, err := os.ReadFile(dest); err != nil || string(data) != "hello" {
		t.Errorf("copied file content mismatch: %v, %s", err, string(data))
	}
	// Should error if artifact does not exist
	if err := am.CopyArtifact(buildID, "nope.txt", dest); err == nil {
		t.Errorf("expected error for missing artifact, got nil")
	}
}
