package db

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) (*DB, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return db, cleanup
}

func TestOpen(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	if db == nil {
		t.Fatal("expected non-nil database")
	}
}

func TestCreateAndGetBuild(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	build := &Build{
		ID:             "test-build-123",
		Customer:       "acme",
		ProjectName:    "test-project",
		TargetImage:    "core-image-minimal",
		Machine:        "qemux86-64",
		Status:         StatusQueued,
		BuildDir:       "/tmp/builds/test-build-123",
		DeployDir:      "/tmp/builds/test-build-123/deploy",
		LogFilePlain:   "/tmp/builds/test-build-123/build-log.txt",
		LogFileJSONL:   "/tmp/builds/test-build-123/build-log.jsonl",
		ConfigFile:     "smidr.yaml",
		ConfigSnapshot: `{"name":"test"}`,
		User:           "testuser",
		Host:           "testhost",
		CreatedAt:      time.Now(),
	}

	err := db.CreateBuild(build)
	if err != nil {
		t.Fatalf("failed to create build: %v", err)
	}

	retrieved, err := db.GetBuild("test-build-123")
	if err != nil {
		t.Fatalf("failed to get build: %v", err)
	}

	if retrieved.ID != build.ID {
		t.Errorf("expected ID %s, got %s", build.ID, retrieved.ID)
	}
	if retrieved.Customer != build.Customer {
		t.Errorf("expected customer %s, got %s", build.Customer, retrieved.Customer)
	}
	if retrieved.Status != StatusQueued {
		t.Errorf("expected status %s, got %s", StatusQueued, retrieved.Status)
	}
}

func TestUpdateBuildStatus(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	build := &Build{
		ID:          "test-build-456",
		Customer:    "acme",
		ProjectName: "test",
		TargetImage: "core-image-minimal",
		Machine:     "qemux86-64",
		Status:      StatusQueued,
		BuildDir:    "/tmp/test",
		DeployDir:   "/tmp/test/deploy",
		User:        "testuser",
		Host:        "testhost",
		CreatedAt:   time.Now(),
	}

	db.CreateBuild(build)

	err := db.UpdateBuildStatus("test-build-456", StatusRunning, "")
	if err != nil {
		t.Fatalf("failed to update status: %v", err)
	}

	retrieved, _ := db.GetBuild("test-build-456")
	if retrieved.Status != StatusRunning {
		t.Errorf("expected status %s, got %s", StatusRunning, retrieved.Status)
	}
}

func TestStartBuild(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	build := &Build{
		ID:          "test-build-789",
		Customer:    "acme",
		ProjectName: "test",
		TargetImage: "core-image-minimal",
		Machine:     "qemux86-64",
		Status:      StatusQueued,
		BuildDir:    "/tmp/test",
		DeployDir:   "/tmp/test/deploy",
		User:        "testuser",
		Host:        "testhost",
		CreatedAt:   time.Now(),
	}

	db.CreateBuild(build)

	err := db.StartBuild("test-build-789")
	if err != nil {
		t.Fatalf("failed to start build: %v", err)
	}

	retrieved, _ := db.GetBuild("test-build-789")
	if retrieved.Status != StatusRunning {
		t.Errorf("expected status %s, got %s", StatusRunning, retrieved.Status)
	}
	if retrieved.StartedAt == nil {
		t.Error("expected started_at to be set")
	}
}

func TestCompleteBuild(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	build := &Build{
		ID:          "test-build-complete",
		Customer:    "acme",
		ProjectName: "test",
		TargetImage: "core-image-minimal",
		Machine:     "qemux86-64",
		Status:      StatusRunning,
		BuildDir:    "/tmp/test",
		DeployDir:   "/tmp/test/deploy",
		User:        "testuser",
		Host:        "testhost",
		CreatedAt:   time.Now(),
	}

	db.CreateBuild(build)

	duration := 120 * time.Second
	err := db.CompleteBuild("test-build-complete", StatusCompleted, 0, duration, "")
	if err != nil {
		t.Fatalf("failed to complete build: %v", err)
	}

	retrieved, _ := db.GetBuild("test-build-complete")
	if retrieved.Status != StatusCompleted {
		t.Errorf("expected status %s, got %s", StatusCompleted, retrieved.Status)
	}
	if retrieved.ExitCode == nil || *retrieved.ExitCode != 0 {
		t.Error("expected exit code 0")
	}
	if retrieved.DurationSeconds == nil || *retrieved.DurationSeconds != 120 {
		t.Errorf("expected duration 120s, got %v", retrieved.DurationSeconds)
	}
}

func TestListBuilds(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create multiple builds
	builds := []*Build{
		{
			ID: "build-1", Customer: "acme", ProjectName: "p1", TargetImage: "img", Machine: "m",
			Status: StatusCompleted, BuildDir: "/tmp/1", DeployDir: "/tmp/1/d",
			User: "u", Host: "h", CreatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			ID: "build-2", Customer: "acme", ProjectName: "p2", TargetImage: "img", Machine: "m",
			Status: StatusFailed, BuildDir: "/tmp/2", DeployDir: "/tmp/2/d",
			User: "u", Host: "h", CreatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			ID: "build-3", Customer: "globex", ProjectName: "p3", TargetImage: "img", Machine: "m",
			Status: StatusCompleted, BuildDir: "/tmp/3", DeployDir: "/tmp/3/d",
			User: "u", Host: "h", CreatedAt: time.Now(),
		},
	}

	for _, b := range builds {
		db.CreateBuild(b)
	}

	// List all builds
	all, err := db.ListBuilds("", false, 0)
	if err != nil {
		t.Fatalf("failed to list builds: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 builds, got %d", len(all))
	}

	// List by customer
	acmeBuilds, err := db.ListBuilds("acme", false, 0)
	if err != nil {
		t.Fatalf("failed to list acme builds: %v", err)
	}
	if len(acmeBuilds) != 2 {
		t.Errorf("expected 2 acme builds, got %d", len(acmeBuilds))
	}

	// List with limit
	limited, err := db.ListBuilds("", false, 1)
	if err != nil {
		t.Fatalf("failed to list limited builds: %v", err)
	}
	if len(limited) != 1 {
		t.Errorf("expected 1 build, got %d", len(limited))
	}
	// Should be most recent (build-3)
	if limited[0].ID != "build-3" {
		t.Errorf("expected build-3, got %s", limited[0].ID)
	}
}

func TestSoftDelete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	build := &Build{
		ID: "build-to-delete", Customer: "acme", ProjectName: "p", TargetImage: "img", Machine: "m",
		Status: StatusCompleted, BuildDir: "/tmp/del", DeployDir: "/tmp/del/d",
		User: "u", Host: "h", CreatedAt: time.Now(),
	}

	db.CreateBuild(build)

	// Soft delete
	err := db.SoftDeleteBuild("build-to-delete")
	if err != nil {
		t.Fatalf("failed to soft delete: %v", err)
	}

	// Should not appear in default list
	builds, _ := db.ListBuilds("", false, 0)
	if len(builds) != 0 {
		t.Errorf("expected 0 builds (deleted excluded), got %d", len(builds))
	}

	// Should appear when including deleted
	buildsWithDeleted, _ := db.ListBuilds("", true, 0)
	if len(buildsWithDeleted) != 1 {
		t.Errorf("expected 1 build (including deleted), got %d", len(buildsWithDeleted))
	}

	retrieved, _ := db.GetBuild("build-to-delete")
	if !retrieved.Deleted {
		t.Error("expected build to be marked as deleted")
	}
	if retrieved.DeletedAt == nil {
		t.Error("expected deleted_at to be set")
	}
}

func TestHardDelete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	build := &Build{
		ID: "build-hard-delete", Customer: "acme", ProjectName: "p", TargetImage: "img", Machine: "m",
		Status: StatusCompleted, BuildDir: "/tmp/hard", DeployDir: "/tmp/hard/d",
		User: "u", Host: "h", CreatedAt: time.Now(),
	}

	db.CreateBuild(build)

	// Hard delete
	err := db.HardDeleteBuild("build-hard-delete")
	if err != nil {
		t.Fatalf("failed to hard delete: %v", err)
	}

	// Should not exist at all
	_, err = db.GetBuild("build-hard-delete")
	if err == nil {
		t.Error("expected error when getting deleted build")
	}
}

func TestArtifacts(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	build := &Build{
		ID: "build-with-artifacts", Customer: "acme", ProjectName: "p", TargetImage: "img", Machine: "m",
		Status: StatusCompleted, BuildDir: "/tmp/art", DeployDir: "/tmp/art/d",
		User: "u", Host: "h", CreatedAt: time.Now(),
	}

	db.CreateBuild(build)

	// Add artifacts
	artifacts := []*BuildArtifact{
		{
			BuildID:      "build-with-artifacts",
			ArtifactPath: "images/core-image-minimal.wic",
			ArtifactType: "wic",
			SizeBytes:    1024000,
			Checksum:     "abc123",
			CreatedAt:    time.Now(),
		},
		{
			BuildID:      "build-with-artifacts",
			ArtifactPath: "images/manifest.txt",
			ArtifactType: "manifest",
			SizeBytes:    2048,
			Checksum:     "def456",
			CreatedAt:    time.Now(),
		},
	}

	for _, a := range artifacts {
		err := db.AddArtifact(a)
		if err != nil {
			t.Fatalf("failed to add artifact: %v", err)
		}
	}

	// List artifacts
	retrieved, err := db.ListArtifacts("build-with-artifacts")
	if err != nil {
		t.Fatalf("failed to list artifacts: %v", err)
	}

	if len(retrieved) != 2 {
		t.Errorf("expected 2 artifacts, got %d", len(retrieved))
	}
}

func TestListStaleBuilds(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a running build (stale)
	stale := &Build{
		ID: "stale-build", Customer: "acme", ProjectName: "p", TargetImage: "img", Machine: "m",
		Status: StatusRunning, BuildDir: "/tmp/stale", DeployDir: "/tmp/stale/d",
		User: "u", Host: "h", CreatedAt: time.Now().Add(-1 * time.Hour),
	}
	now := time.Now()
	stale.StartedAt = &now
	db.CreateBuild(stale)

	// Create a completed build (not stale)
	completed := &Build{
		ID: "completed-build", Customer: "acme", ProjectName: "p", TargetImage: "img", Machine: "m",
		Status: StatusCompleted, BuildDir: "/tmp/done", DeployDir: "/tmp/done/d",
		User: "u", Host: "h", CreatedAt: time.Now().Add(-2 * time.Hour),
	}
	db.CreateBuild(completed)

	// List stale builds
	staleBuilds, err := db.ListStaleBuilds()
	if err != nil {
		t.Fatalf("failed to list stale builds: %v", err)
	}

	if len(staleBuilds) != 1 {
		t.Errorf("expected 1 stale build, got %d", len(staleBuilds))
	}
	if len(staleBuilds) > 0 && staleBuilds[0].ID != "stale-build" {
		t.Errorf("expected stale-build, got %s", staleBuilds[0].ID)
	}
}
