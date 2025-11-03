package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/schererja/smidr/internal/artifacts"
	"github.com/spf13/cobra"
)

func newTestCmd() *cobra.Command {
	return &cobra.Command{Use: "test"}
}

func TestRunArtifactsList_Basic(t *testing.T) {
	cmd := newTestCmd()
	err := runArtifactsList(cmd)
	if err != nil {
		t.Logf("runArtifactsList returned error: %v", err)
	}
}

func TestRunArtifactsCopy_Basic(t *testing.T) {
	cmd := newTestCmd()
	err := runArtifactsCopy(cmd, "fake-build", "/tmp/dest")
	if err != nil {
		t.Logf("runArtifactsCopy returned error: %v", err)
	}
}

func TestRunArtifactsClean_Basic(t *testing.T) {
	cmd := newTestCmd()
	err := runArtifactsClean(cmd)
	if err != nil {
		t.Logf("runArtifactsClean returned error: %v", err)
	}
}

func TestRunArtifactsShow_Basic(t *testing.T) {
	cmd := newTestCmd()
	err := runArtifactsShow(cmd, "fake-build")
	if err != nil {
		t.Logf("runArtifactsShow returned error: %v", err)
	}
}

func TestRunArtifactsList_WithActualArtifacts(t *testing.T) {
	// Create temporary artifact directory
	tmpDir := t.TempDir()

	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer func() { os.Setenv("HOME", originalHome) }()
	os.Setenv("HOME", tmpDir)

	// Create artifact manager and save some metadata
	am, err := artifacts.NewArtifactManager("")
	if err != nil {
		t.Fatalf("failed to create artifact manager: %v", err)
	}

	// Create test build metadata
	metadata := artifacts.BuildMetadata{
		BuildID:       "test-build-123",
		ProjectName:   "test-project",
		User:          "testuser",
		Timestamp:     time.Now(),
		BuildDuration: 10 * time.Second,
		Status:        "success",
		TargetImage:   "core-image-minimal",
	}

	if err := am.SaveMetadata(metadata); err != nil {
		t.Fatalf("failed to save metadata: %v", err)
	}

	// Now run the list command
	cmd := newTestCmd()
	err = runArtifactsList(cmd)
	if err != nil {
		t.Errorf("runArtifactsList failed: %v", err)
	}
}

func TestRunArtifactsShow_WithActualArtifacts(t *testing.T) {
	// Create temporary artifact directory
	tmpDir := t.TempDir()

	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer func() { os.Setenv("HOME", originalHome) }()
	os.Setenv("HOME", tmpDir)

	// Create artifact manager and save some metadata
	am, err := artifacts.NewArtifactManager("")
	if err != nil {
		t.Fatalf("failed to create artifact manager: %v", err)
	}

	// Create test build metadata
	buildID := "test-build-456"
	metadata := artifacts.BuildMetadata{
		BuildID:       buildID,
		ProjectName:   "test-project",
		User:          "testuser",
		Timestamp:     time.Now(),
		BuildDuration: 15 * time.Second,
		Status:        "success",
		TargetImage:   "core-image-full-cmdline",
		ArtifactSizes: map[string]int64{"test.img": 1024},
	}

	if err := am.SaveMetadata(metadata); err != nil {
		t.Fatalf("failed to save metadata: %v", err)
	}

	// Create a test artifact file
	artifactPath := am.GetArtifactPath(buildID)
	os.MkdirAll(artifactPath, 0755)
	os.WriteFile(filepath.Join(artifactPath, "test.img"), []byte("test data"), 0644)

	// Now run the show command
	cmd := newTestCmd()
	err = runArtifactsShow(cmd, buildID)
	if err != nil {
		t.Errorf("runArtifactsShow failed: %v", err)
	}
}

func TestRunArtifactsCopy_WithActualArtifacts(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	destDir := t.TempDir()

	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer func() { os.Setenv("HOME", originalHome) }()
	os.Setenv("HOME", tmpDir)

	// Create artifact manager and save some metadata
	am, err := artifacts.NewArtifactManager("")
	if err != nil {
		t.Fatalf("failed to create artifact manager: %v", err)
	}

	// Create test build metadata
	buildID := "test-build-789"
	metadata := artifacts.BuildMetadata{
		BuildID:     buildID,
		ProjectName: "test-project",
		User:        "testuser",
		Timestamp:   time.Now(),
		Status:      "success",
	}

	if err := am.SaveMetadata(metadata); err != nil {
		t.Fatalf("failed to save metadata: %v", err)
	}

	// Create test artifact files
	artifactPath := am.GetArtifactPath(buildID)
	os.MkdirAll(artifactPath, 0755)
	os.WriteFile(filepath.Join(artifactPath, "image.wic"), []byte("disk image"), 0644)
	os.WriteFile(filepath.Join(artifactPath, "manifest.txt"), []byte("manifest"), 0644)

	// Now run the copy command
	cmd := newTestCmd()
	err = runArtifactsCopy(cmd, buildID, destDir)
	if err != nil {
		t.Errorf("runArtifactsCopy failed: %v", err)
	}

	// Verify files were copied (note: they're copied to a subdirectory)
	actualDestDir := filepath.Join(destDir, "smidr-artifacts-"+buildID)
	copiedFile := filepath.Join(actualDestDir, "image.wic")
	if _, err := os.Stat(copiedFile); os.IsNotExist(err) {
		t.Errorf("expected file to be copied to %s", copiedFile)
	}
}

func TestRunArtifactsClean_WithOldArtifacts(t *testing.T) {
	// Create temporary artifact directory
	tmpDir := t.TempDir()

	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer func() { os.Setenv("HOME", originalHome) }()
	os.Setenv("HOME", tmpDir)

	// Create artifact manager and save old metadata
	am, err := artifacts.NewArtifactManager("")
	if err != nil {
		t.Fatalf("failed to create artifact manager: %v", err)
	}

	// Create old build metadata (100 days ago)
	oldTime := time.Now().AddDate(0, 0, -100)
	metadata := artifacts.BuildMetadata{
		BuildID:     "old-build",
		ProjectName: "test-project",
		User:        "testuser",
		Timestamp:   oldTime,
		Status:      "success",
	}

	if err := am.SaveMetadata(metadata); err != nil {
		t.Fatalf("failed to save metadata: %v", err)
	}

	// Create artifact file
	artifactPath := am.GetArtifactPath("old-build")
	os.MkdirAll(artifactPath, 0755)
	os.WriteFile(filepath.Join(artifactPath, "old.img"), []byte("old data"), 0644)

	// Run clean command (default is 30 days)
	cmd := newTestCmd()
	cmd.Flags().Int("days", 30, "")
	err = runArtifactsClean(cmd)
	if err != nil {
		t.Errorf("runArtifactsClean failed: %v", err)
	}
}

func TestArtifacts_WithCustomerScope(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer func() { os.Setenv("HOME", originalHome) }()
	os.Setenv("HOME", tmpDir)

	// Create artifact-acme scoped build
	acmeBase := filepath.Join(tmpDir, ".smidr", "artifacts", "artifact-acme")
	if err := os.MkdirAll(acmeBase, 0o755); err != nil {
		t.Fatalf("mkdirs: %v", err)
	}
	am, err := artifacts.NewArtifactManager(filepath.Join(acmeBase))
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	buildID := "acme-build-1"
	meta := artifacts.BuildMetadata{
		BuildID:       buildID,
		ProjectName:   "proj",
		User:          "u",
		Timestamp:     time.Now(),
		BuildDuration: 1 * time.Second,
		Status:        "success",
	}
	if err := am.SaveMetadata(meta); err != nil {
		t.Fatalf("save meta: %v", err)
	}
	// create a dummy file
	p := am.GetArtifactPath(buildID)
	_ = os.MkdirAll(p, 0o755)
	_ = os.WriteFile(filepath.Join(p, "art.bin"), []byte("x"), 0o644)

	// List with --customer
	listCmd := newTestCmd()
	listCmd.Flags().String("customer", "acme", "")
	if err := runArtifactsList(listCmd); err != nil {
		t.Fatalf("runArtifactsList(--customer) failed: %v", err)
	}

	// Show with --customer
	showCmd := newTestCmd()
	showCmd.Flags().String("customer", "acme", "")
	if err := runArtifactsShow(showCmd, buildID); err != nil {
		t.Fatalf("runArtifactsShow(--customer) failed: %v", err)
	}

	// Copy with --customer
	dest := t.TempDir()
	copyCmd := newTestCmd()
	copyCmd.Flags().String("customer", "acme", "")
	if err := runArtifactsCopy(copyCmd, buildID, dest); err != nil {
		t.Fatalf("runArtifactsCopy(--customer) failed: %v", err)
	}
	// verify
	copied := filepath.Join(dest, "smidr-artifacts-"+buildID, "art.bin")
	if _, err := os.Stat(copied); err != nil {
		t.Fatalf("expected copied file: %v", err)
	}
}

func TestRunArtifactsClean_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer func() { os.Setenv("HOME", originalHome) }()
	os.Setenv("HOME", tmpDir)

	am, err := artifacts.NewArtifactManager("")
	if err != nil {
		t.Fatalf("failed to create artifact manager: %v", err)
	}

	// Create a few builds, one old and two recent
	oldTime := time.Now().AddDate(0, 0, -100)
	builds := []artifacts.BuildMetadata{
		{BuildID: "old", ProjectName: "p", User: "u", Timestamp: oldTime, Status: "success"},
		{BuildID: "keep1", ProjectName: "p", User: "u", Timestamp: time.Now(), Status: "success"},
		{BuildID: "keep2", ProjectName: "p", User: "u", Timestamp: time.Now(), Status: "success"},
	}
	for _, m := range builds {
		if err := am.SaveMetadata(m); err != nil {
			t.Fatalf("save metadata: %v", err)
		}
		// create a tiny artifact file to avoid empty dirs
		p := am.GetArtifactPath(m.BuildID)
		_ = os.MkdirAll(p, 0o755)
		_ = os.WriteFile(filepath.Join(p, "a.txt"), []byte("x"), 0o644)
	}

	cmd := newTestCmd()
	cmd.Flags().Int("keep", 2, "")
	cmd.Flags().Int("days", 30, "")
	cmd.Flags().Bool("dry-run", true, "")
	if err := runArtifactsClean(cmd); err != nil {
		t.Fatalf("runArtifactsClean dry-run failed: %v", err)
	}
}
