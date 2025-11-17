package artifacts

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultRetentionPolicy(t *testing.T) {
	policy := DefaultRetentionPolicy()
	if policy.KeepLastN != 10 {
		t.Errorf("expected KeepLastN=10, got %d", policy.KeepLastN)
	}
	if policy.MaxAge != 30*24*time.Hour {
		t.Errorf("expected MaxAge=30d, got %v", policy.MaxAge)
	}
	if policy.MaxSizeGB != 0 {
		t.Errorf("expected MaxSizeGB=0, got %d", policy.MaxSizeGB)
	}
}

func TestDeleteBuild(t *testing.T) {
	tmpDir := t.TempDir()
	am, err := NewArtifactManager(tmpDir)
	if err != nil {
		t.Fatalf("failed to create ArtifactManager: %v", err)
	}
	buildID := "test-build"
	buildPath := filepath.Join(tmpDir, buildID)
	if err := os.MkdirAll(buildPath, 0755); err != nil {
		t.Fatalf("failed to create build dir: %v", err)
	}
	// Should delete existing build
	if err := am.DeleteBuild(buildID); err != nil {
		t.Errorf("DeleteBuild failed: %v", err)
	}
	if _, err := os.Stat(buildPath); !os.IsNotExist(err) {
		t.Errorf("build dir still exists after DeleteBuild")
	}
	// Should error if build does not exist
	if err := am.DeleteBuild("nonexistent"); err == nil {
		t.Errorf("expected error for nonexistent build, got nil")
	}
}

func TestCleanupArtifacts_Empty(t *testing.T) {
	am, err := NewArtifactManager("")
	if err != nil {
		t.Fatalf("failed to create ArtifactManager: %v", err)
	}
	// Should not error if no builds
	if err := am.CleanupArtifacts(DefaultRetentionPolicy()); err != nil {
		t.Errorf("CleanupArtifacts failed on empty: %v", err)
	}
}

func TestRetentionPolicy_AgeAndCount(t *testing.T) {
	am, err := NewArtifactManager("")
	if err != nil {
		t.Fatalf("failed to create ArtifactManager: %v", err)
	}
	// Add 3 builds: one old, two recent
	old := BuildMetadata{BuildID: "old", Timestamp: time.Now().Add(-40 * 24 * time.Hour)}
	recent1 := BuildMetadata{BuildID: "recent1", Timestamp: time.Now().Add(-2 * time.Hour)}
	recent2 := BuildMetadata{BuildID: "recent2", Timestamp: time.Now().Add(-1 * time.Hour)}
	_ = am.SaveMetadata(old)
	_ = am.SaveMetadata(recent1)
	_ = am.SaveMetadata(recent2)
	policy := RetentionPolicy{KeepLastN: 2, MaxAge: 30 * 24 * time.Hour}
	if err := am.CleanupArtifacts(policy); err != nil {
		t.Errorf("CleanupArtifacts failed: %v", err)
	}
	builds, _ := am.ListBuilds()
	if len(builds) != 2 {
		t.Errorf("expected 2 builds after cleanup, got %d", len(builds))
	}
}

func TestArtifactManager_GetTotalSize(t *testing.T) {
	tmpDir := t.TempDir()
	am, err := NewArtifactManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create artifact manager: %v", err)
	}
	// No builds: should return 0, nil
	total, err := am.GetTotalSize()
	if err != nil {
		t.Errorf("Expected nil error for empty store, got %v", err)
	}
	if total != 0 {
		t.Errorf("Expected total size 0 for empty store, got %d", total)
	}
	// Add builds with artifact sizes
	md1 := BuildMetadata{
		BuildID:       "b1",
		ProjectName:   "p",
		User:          "u",
		Timestamp:     time.Now(),
		Status:        "success",
		ArtifactSizes: map[string]int64{"a1": 100, "a2": 200},
	}
	md2 := BuildMetadata{
		BuildID:       "b2",
		ProjectName:   "p",
		User:          "u",
		Timestamp:     time.Now(),
		Status:        "success",
		ArtifactSizes: map[string]int64{"a3": 300},
	}
	am.SaveMetadata(md1)
	am.SaveMetadata(md2)
	total, err = am.GetTotalSize()
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if total != 600 {
		t.Errorf("Expected total size 600, got %d", total)
	}
}
