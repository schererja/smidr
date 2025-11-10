package source

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/schererja/smidr/internal/config"
	"github.com/schererja/smidr/pkg/logger"
)

// ...existing code...
func TestFetcher_fetchBaseLayer(t *testing.T) {
	tmpDir := t.TempDir()
	logger := logger.NewLogger()
	fetcher := NewFetcher(tmpDir, tmpDir, logger)
	cfg := &config.Config{Base: config.BaseConfig{Version: "6.0.0"}}

	t.Run("unknown layer returns error", func(t *testing.T) {
		result := fetcher.fetchBaseLayer("not-a-real-layer", cfg)
		if result.Success {
			t.Errorf("Expected failure for unknown layer, got success")
		}
		if result.Error == nil {
			t.Errorf("Expected error for unknown layer, got nil")
		}
	})

	t.Run("known layer returns result", func(t *testing.T) {
		result := fetcher.fetchBaseLayer("poky", cfg)
		if result.LayerName != "poky" {
			t.Errorf("Expected LayerName 'poky', got %q", result.LayerName)
		}
		// Success may depend on git availability, but should not error for known layer
		if result.Error != nil && !strings.Contains(result.Error.Error(), "git") {
			t.Errorf("Unexpected error for known layer: %v", result.Error)
		}
	})
}

func TestFetcher_fetchLayer(t *testing.T) {
	tmpDir := t.TempDir()
	logger := logger.NewLogger()
	fetcher := NewFetcher(tmpDir, tmpDir, logger)
	cfg := &config.Config{}

	t.Run("unknown layer returns error", func(t *testing.T) {
		result := fetcher.fetchLayer("not-a-real-layer", cfg)
		if result.Success {
			t.Errorf("Expected failure for unknown layer, got success")
		}
		if result.Error == nil {
			t.Errorf("Expected error for unknown layer, got nil")
		}
	})

	t.Run("known layer returns valid result", func(t *testing.T) {
		result := fetcher.fetchLayer("poky", cfg)
		if result.LayerName != "poky" {
			t.Errorf("Expected LayerName 'poky', got %q", result.LayerName)
		}
		if result.Error != nil {
			t.Errorf("Unexpected error for known layer: %v", result.Error)
		}
	})
}

func TestNewFetcher(t *testing.T) {
	logger := logger.NewLogger()

	layersDir := "/tmp/test-layers"
	downloadsDir := "/tmp/test-downloads"

	fetcher := NewFetcher(layersDir, downloadsDir, logger)

	if fetcher == nil {
		t.Fatal("NewFetcher returned nil")
	}

	if fetcher.logger != logger {
		t.Error("logger not set correctly")
	}
}

func TestIsGitRepository(t *testing.T) {
	logger := logger.NewLogger()
	tmpDir := t.TempDir()
	fetcher := NewFetcher(tmpDir, tmpDir, logger)

	t.Run("directory with .git", func(t *testing.T) {
		gitDir := filepath.Join(tmpDir, "repo-with-git")
		os.MkdirAll(filepath.Join(gitDir, ".git"), 0755)

		if !fetcher.isGitRepository(gitDir) {
			t.Error("Should detect .git directory")
		}
	})

	t.Run("directory without .git", func(t *testing.T) {
		noGitDir := filepath.Join(tmpDir, "repo-without-git")
		os.MkdirAll(noGitDir, 0755)

		if fetcher.isGitRepository(noGitDir) {
			t.Error("Should not detect git in directory without .git")
		}
	})

	t.Run("non-existent directory", func(t *testing.T) {
		if fetcher.isGitRepository("/nonexistent/path") {
			t.Error("Should return false for non-existent directory")
		}
	})
}

func TestGetRequiredBaseLayers(t *testing.T) {
	tests := []struct {
		provider      string
		expectedCount int
		mustInclude   []string
	}{
		{
			provider:      "toradex",
			expectedCount: 4,
			mustInclude:   []string{"poky", "meta-openembedded", "meta-toradex-nxp", "meta-toradex-bsp-common"},
		},
		{
			provider:      "raspberrypi",
			expectedCount: 3,
			mustInclude:   []string{"poky", "meta-openembedded", "meta-raspberrypi"},
		},
		{
			provider:      "nvidia",
			expectedCount: 3,
			mustInclude:   []string{"poky", "meta-openembedded", "meta-tegra"},
		},
		{
			provider:      "custom",
			expectedCount: 2,
			mustInclude:   []string{"poky", "meta-openembedded"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			layers := getRequiredBaseLayers(tt.provider)

			if len(layers) != tt.expectedCount {
				t.Errorf("Expected %d layers, got %d: %v", tt.expectedCount, len(layers), layers)
			}

			for _, required := range tt.mustInclude {
				found := false
				for _, layer := range layers {
					if layer == required {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Required layer %q not found in %v", required, layers)
				}
			}
		})
	}
}

func TestGetBaseLayerRepository(t *testing.T) {
	tests := []struct {
		layerName string
		wantRepo  bool // Should return a non-empty repo
	}{
		{"poky", true},
		{"meta-openembedded", true},
		{"meta-toradex-nxp", true},
		{"meta-toradex-bsp-common", true},
		{"meta-raspberrypi", true},
		{"meta-tegra", true},
		{"unknown-layer", false},
	}

	for _, tt := range tests {
		t.Run(tt.layerName, func(t *testing.T) {
			repo := getBaseLayerRepository(tt.layerName)

			if tt.wantRepo && repo == "" {
				t.Errorf("Expected non-empty repo for %q", tt.layerName)
			}

			if !tt.wantRepo && repo != "" {
				t.Errorf("Expected empty repo for unknown layer %q, got %q", tt.layerName, repo)
			}

			// Verify URLs are valid format
			if repo != "" {
				if !strings.HasPrefix(repo, "http") && !strings.HasPrefix(repo, "git://") {
					t.Errorf("Invalid repo URL format: %q", repo)
				}
			}
		})
	}
}

func TestGetBranchForLayer(t *testing.T) {
	tests := []struct {
		name      string
		layerName string
		version   string
		expected  string
	}{
		{
			name:      "version 6.0.0 maps to kirkstone",
			layerName: "poky",
			version:   "6.0.0",
			expected:  "kirkstone",
		},
		{
			name:      "version 5.0.0 maps to dunfell",
			layerName: "poky",
			version:   "5.0.0",
			expected:  "dunfell",
		},
		{
			name:      "toradex layer has special branch format",
			layerName: "meta-toradex-nxp",
			version:   "",
			expected:  "kirkstone-6.x.y",
		},
		{
			name:      "unknown layer defaults to master",
			layerName: "meta-unknown",
			version:   "",
			expected:  "master",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			branch := getBranchForLayer(tt.layerName, tt.version)
			if branch != tt.expected {
				t.Errorf("getBranchForLayer(%q, %q) = %q, want %q",
					tt.layerName, tt.version, branch, tt.expected)
			}
		})
	}
}

func TestGetAlternativeRepository(t *testing.T) {
	cases := []struct {
		layer   string
		wantLen int
	}{
		{"meta-toradex-nxp", 2},
		{"meta-toradex-bsp-common", 2},
		{"poky", 0},
		{"meta-openembedded", 0},
		{"meta-raspberrypi", 0},
		{"unknown-layer", 0},
	}

	for _, c := range cases {
		t.Run(c.layer, func(t *testing.T) {
			alts := getAlternativeRepository(c.layer)
			if len(alts) != c.wantLen {
				t.Fatalf("getAlternativeRepository(%q) len=%d, want %d (%v)", c.layer, len(alts), c.wantLen, alts)
			}
			// basic sanity: when present, ensure entries look like URLs
			for _, a := range alts {
				if !strings.HasPrefix(a, "http") {
					t.Errorf("alternative %q does not look like a URL", a)
				}
			}
		})
	}
}

func TestFetchGitLayer_MockGit(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available, skipping test")
	}

	t.Run("clone new repository", func(t *testing.T) {
		// Create a temporary git repository to clone from
		tmpDir := t.TempDir()
		sourceRepo := filepath.Join(tmpDir, "source-repo")

		// Create a proper git repository with commits
		if err := os.MkdirAll(sourceRepo, 0755); err != nil {
			t.Skipf("Failed to create repo dir: %v", err)
		}

		// Initialize the repo with main branch
		if err := exec.Command("git", "init", "-b", "main", sourceRepo).Run(); err != nil {
			t.Skipf("Failed to init git repo: %v", err)
		}

		// Configure git user for the test
		exec.Command("git", "-C", sourceRepo, "config", "user.email", "test@example.com").Run()
		exec.Command("git", "-C", sourceRepo, "config", "user.name", "Test User").Run()

		// Create a test file and commit it
		testFile := filepath.Join(sourceRepo, "test.txt")
		if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
			t.Skipf("Failed to create test file: %v", err)
		}

		if err := exec.Command("git", "-C", sourceRepo, "add", "test.txt").Run(); err != nil {
			t.Skipf("Failed to add file: %v", err)
		}

		if err := exec.Command("git", "-C", sourceRepo, "commit", "-m", "Initial commit").Run(); err != nil {
			t.Skipf("Failed to commit: %v", err)
		}

		logger := logger.NewLogger()
		sourcesDir := filepath.Join(tmpDir, "sources")
		fetcher := NewFetcher(sourcesDir, sourcesDir, logger)

		layer := config.Layer{
			Name:   "test-layer",
			Git:    sourceRepo,
			Branch: "main", // Use modern git default branch name
		}

		// Test using the current FetchLayers API
		cfg := &config.Config{
			Layers: []config.Layer{layer},
		}
		results, err := fetcher.FetchLayers(cfg)
		if err != nil {
			t.Errorf("FetchLayers failed: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}

		if len(results) > 0 && !results[0].Success {
			t.Errorf("Expected successful fetch, got error: %v", results[0].Error)
		}

		t.Logf("Fetch results: %+v", results)
	})
}

func TestCleanCache(t *testing.T) {
	tmpDir := t.TempDir()
	sourcesDir := filepath.Join(tmpDir, "sources")

	// Create some test files
	os.MkdirAll(filepath.Join(sourcesDir, "layer1"), 0755)
	os.WriteFile(filepath.Join(sourcesDir, "layer1", "test.txt"), []byte("test"), 0644)

	logger := logger.NewLogger()
	fetcher := NewFetcher(sourcesDir, sourcesDir, logger)

	err := fetcher.CleanCache()
	if err != nil {
		t.Fatalf("CleanCache failed: %v", err)
	}

	// Verify directory was cleaned
	entries, err := os.ReadDir(sourcesDir)
	if err != nil {
		t.Fatalf("Failed to read sources directory: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("Expected empty directory after CleanCache, got %d entries", len(entries))
	}

	// Verify directory still exists
	if _, err := os.Stat(sourcesDir); os.IsNotExist(err) {
		t.Error("Sources directory should still exist after CleanCache")
	}
}

func TestGetCacheSize(t *testing.T) {
	tmpDir := t.TempDir()
	sourcesDir := filepath.Join(tmpDir, "sources")

	// Create test files with known sizes
	os.MkdirAll(filepath.Join(sourcesDir, "layer1"), 0755)

	testData := []byte("test data with some content")
	file1 := filepath.Join(sourcesDir, "layer1", "file1.txt")
	file2 := filepath.Join(sourcesDir, "layer1", "file2.txt")

	os.WriteFile(file1, testData, 0644)
	os.WriteFile(file2, testData, 0644)

	expectedSize := int64(len(testData) * 2)

	logger := logger.NewLogger()
	fetcher := NewFetcher(sourcesDir, sourcesDir, logger)

	size, err := fetcher.GetCacheSize()
	if err != nil {
		t.Fatalf("GetCacheSize failed: %v", err)
	}

	if size != expectedSize {
		t.Errorf("GetCacheSize() = %d, want %d", size, expectedSize)
	}
}

func TestFetchLayers_EmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	logger := logger.NewLogger()
	sourcesDir := filepath.Join(tmpDir, "sources")
	fetcher := NewFetcher(sourcesDir, sourcesDir, logger)

	cfg := &config.Config{
		Base: config.BaseConfig{
			Provider: "custom", // No extra layers
		},
		Layers: []config.Layer{}, // No custom layers
	}

	// With the current implementation, empty layers means no fetch operations
	results, err := fetcher.FetchLayers(cfg)

	if err != nil {
		t.Errorf("FetchLayers should not fail with empty config: %v", err)
	}

	// Should have no fetch attempts since no layers are configured
	if len(results) != 0 {
		t.Errorf("Expected 0 fetch attempts for empty config, got %d", len(results))
	}
}

func TestFetchResult(t *testing.T) {
	result := FetchResult{
		LayerName: "test-layer",
		Path:      "/path/to/layer",
		Success:   true,
		Cached:    false,
	}

	if result.LayerName != "test-layer" {
		t.Errorf("LayerName = %q, want %q", result.LayerName, "test-layer")
	}

	if !result.Success {
		t.Error("Expected Success to be true")
	}

	if result.Cached {
		t.Error("Expected Cached to be false")
	}
}

func TestEvictOldCache(t *testing.T) {
	tmpDir := t.TempDir()
	logger := logger.NewLogger()
	fetcher := NewFetcher(tmpDir, tmpDir, logger)

	// Create two fake repos: one old, one recent
	oldRepo := filepath.Join(tmpDir, "old-repo")
	newRepo := filepath.Join(tmpDir, "new-repo")
	os.MkdirAll(oldRepo, 0755)
	os.MkdirAll(newRepo, 0755)

	// Write meta: old repo last accessed 48h ago, new repo now
	oldMeta := CacheMeta{LastAccess: time.Now().Add(-48 * time.Hour)}
	newMeta := CacheMeta{LastAccess: time.Now()}
	oldMetaBytes, _ := json.Marshal(oldMeta)
	newMetaBytes, _ := json.Marshal(newMeta)
	os.WriteFile(filepath.Join(oldRepo, ".smidr_meta.json"), oldMetaBytes, 0644)
	os.WriteFile(filepath.Join(newRepo, ".smidr_meta.json"), newMetaBytes, 0644)

	// Run eviction with TTL = 24h
	err := fetcher.EvictOldCache(24 * time.Hour)
	if err != nil {
		t.Fatalf("EvictOldCache failed: %v", err)
	}

	// Old repo should be gone, new repo should remain
	if _, err := os.Stat(oldRepo); !os.IsNotExist(err) {
		t.Errorf("Old repo was not evicted")
	}
	if _, err := os.Stat(newRepo); err != nil {
		t.Errorf("New repo was incorrectly evicted: %v", err)
	}
}

func TestPerRepoLocking(t *testing.T) {
	tmpDir := t.TempDir()
	_ = logger.NewLogger()

	lockFile := filepath.Join(tmpDir, "test-layer.lock")

	// Acquire lock in background (simulate another process)
	acquired := make(chan struct{})
	go func() {
		ok, err := acquireLock(lockFile, 2*time.Second)
		if err != nil || !ok {
			t.Logf("background acquire failed: %v", err)
			close(acquired)
			return
		}
		// Signal that lock is held
		close(acquired)
		// Hold the lock briefly
		time.Sleep(500 * time.Millisecond)
		_ = releaseLock(lockFile)
	}()

	// Wait for background goroutine to acquire the lock
	<-acquired

	// Now attempt to acquire lock with a short timeout; should either wait and then succeed after release
	start := time.Now()
	ok, err := acquireLock(lockFile, 3*time.Second)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}
	if !ok {
		t.Fatalf("acquireLock returned false")
	}

	// We should have waited at least some time (background held it for 500ms)
	if elapsed < 400*time.Millisecond {
		t.Errorf("expected to wait for lock, but acquired too quickly: %v", elapsed)
	}

	// Clean up
	_ = releaseLock(lockFile)
}

func BenchmarkGetRequiredBaseLayers(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = getRequiredBaseLayers("toradex")
	}
}

func BenchmarkGetBaseLayerRepository(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = getBaseLayerRepository("poky")
	}
}

func BenchmarkGetBranchForLayer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = getBranchForLayer("poky", "6.0.0")
	}
}
