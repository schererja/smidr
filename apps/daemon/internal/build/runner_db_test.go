package build

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/schererja/smidr/internal/config"
	"github.com/schererja/smidr/internal/db"
	"github.com/schererja/smidr/pkg/logger"
)

// TestRunnerWithDatabase verifies that Runner creates and updates DB records when a DB is provided
func TestRunnerWithDatabase(t *testing.T) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Initialize test database
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	defer database.Close()

	// Create logger
	log := logger.NewLogger()

	// Create runner with database
	runner := NewRunner(log, database)
	if runner.db == nil {
		t.Fatal("expected runner to have DB set")
	}

	// Create minimal config for a smoke test (will skip actual BitBake)
	// Set test environment variables to trigger smoke mode
	os.Setenv("SMIDR_TEST_WRITE_MARKERS", "1")
	os.Setenv("SMIDR_TEST_ENTRYPOINT", "sh,-c,sleep 1")
	defer os.Unsetenv("SMIDR_TEST_WRITE_MARKERS")
	defer os.Unsetenv("SMIDR_TEST_ENTRYPOINT")

	cfg := &config.Config{
		Name:        "test-project",
		YoctoSeries: "scarthgap",
		Base: config.BaseConfig{
			Provider: "poky",
			Version:  "scarthgap",
			Machine:  "qemux86-64",
		},
		Build: config.BuildConfig{
			Image:           "core-image-minimal",
			Machine:         "qemux86-64",
			ParallelMake:    2,
			BBNumberThreads: 2,
		},
		Directories: config.DirectoryConfig{
			Build:     filepath.Join(tmpDir, "build"),
			Layers:    filepath.Join(tmpDir, "layers"),
			Source:    filepath.Join(tmpDir, "source"),
			Downloads: filepath.Join(tmpDir, "downloads"),
			SState:    filepath.Join(tmpDir, "sstate"),
		},
		Layers: []config.Layer{
			{Name: "poky", Git: "https://git.yoctoproject.org/poky", Branch: "scarthgap"},
		},
	}

	// Create required directories
	os.MkdirAll(cfg.Directories.Layers, 0755)
	os.MkdirAll(filepath.Join(cfg.Directories.Layers, "poky"), 0755)

	opts := BuildOptions{
		BuildID:  "test-build-123",
		Target:   "core-image-minimal",
		Customer: "test-customer",
	}

	// Mock log sink
	sink := &mockLogSink{lines: []string{}}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Run the build (will be a smoke test due to env vars)
	// We expect it to create a DB record, mark as running, then complete
	// Note: This will fail at layer fetch since we don't have real layers, but DB ops should still work
	_, _ = runner.Run(ctx, cfg, opts, sink)

	// Verify that a build record was created in the database
	build, err := database.GetBuild(opts.BuildID)
	if err != nil {
		t.Fatalf("failed to retrieve build from database: %v", err)
	}

	if build == nil {
		t.Fatal("expected build record to exist in database")
	}

	if build.ID != opts.BuildID {
		t.Errorf("expected build ID %s, got %s", opts.BuildID, build.ID)
	}

	if build.Customer != opts.Customer {
		t.Errorf("expected customer %s, got %s", opts.Customer, build.Customer)
	}

	if build.ProjectName != cfg.Name {
		t.Errorf("expected project name %s, got %s", cfg.Name, build.ProjectName)
	}

	// Verify status is no longer QUEUED (should be RUNNING or COMPLETED/FAILED)
	if build.Status == db.StatusQueued {
		t.Error("expected build status to have progressed beyond QUEUED")
	}

	t.Logf("Build record created successfully: ID=%s, Status=%s, ExitCode=%d",
		build.ID, build.Status, build.ExitCode)
}

// TestRunnerWithoutDatabase verifies that Runner works normally when DB is nil
func TestRunnerWithoutDatabase(t *testing.T) {
	log := logger.NewLogger()

	// Create runner without database
	runner := NewRunner(log, nil)
	if runner.db != nil {
		t.Fatal("expected runner to have nil DB")
	}

	// Test environment setup (smoke mode)
	os.Setenv("SMIDR_TEST_WRITE_MARKERS", "1")
	os.Setenv("SMIDR_TEST_ENTRYPOINT", "sh,-c,sleep 1")
	defer os.Unsetenv("SMIDR_TEST_WRITE_MARKERS")
	defer os.Unsetenv("SMIDR_TEST_ENTRYPOINT")

	tmpDir := t.TempDir()
	cfg := &config.Config{
		Name:        "test-project",
		YoctoSeries: "scarthgap",
		Base: config.BaseConfig{
			Provider: "poky",
			Version:  "scarthgap",
			Machine:  "qemux86-64",
		},
		Build: config.BuildConfig{
			Image:           "core-image-minimal",
			Machine:         "qemux86-64",
			ParallelMake:    2,
			BBNumberThreads: 2,
		},
		Directories: config.DirectoryConfig{
			Build:     filepath.Join(tmpDir, "build"),
			Layers:    filepath.Join(tmpDir, "layers"),
			Source:    filepath.Join(tmpDir, "source"),
			Downloads: filepath.Join(tmpDir, "downloads"),
			SState:    filepath.Join(tmpDir, "sstate"),
		},
		Layers: []config.Layer{
			{Name: "poky", Git: "https://git.yoctoproject.org/poky", Branch: "scarthgap"},
		},
	}

	os.MkdirAll(cfg.Directories.Layers, 0755)
	os.MkdirAll(filepath.Join(cfg.Directories.Layers, "poky"), 0755)

	opts := BuildOptions{
		BuildID:  "test-build-no-db",
		Target:   "core-image-minimal",
		Customer: "test-customer",
	}

	sink := &mockLogSink{lines: []string{}}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Run should work normally without DB
	_, _ = runner.Run(ctx, cfg, opts, sink)

	// No database assertions - just verify it didn't crash
	t.Log("Runner completed without database (expected behavior)")
}

// mockLogSink implements LogSink for testing
type mockLogSink struct {
	lines []string
}

func (m *mockLogSink) Write(stream string, line string) {
	m.lines = append(m.lines, line)
}
