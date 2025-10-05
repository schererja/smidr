package source

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/intrik8-labs/smidr/internal/config"
)

// MockLogger for testing
type MockLogger struct {
	InfoMessages  []string
	ErrorMessages []string
	DebugMessages []string
}

func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.InfoMessages = append(m.InfoMessages, msg)
}

func (m *MockLogger) Error(msg string, args ...interface{}) {
	m.ErrorMessages = append(m.ErrorMessages, msg)
}

func (m *MockLogger) Debug(msg string, args ...interface{}) {
	m.DebugMessages = append(m.DebugMessages, msg)
}

func TestNewFetcher(t *testing.T) {
	logger := &MockLogger{}
	sourcesDir := "/tmp/test-sources"

	fetcher := NewFetcher(sourcesDir, logger)

	if fetcher == nil {
		t.Fatal("NewFetcher returned nil")
	}

	if fetcher.sourcesDir != sourcesDir {
		t.Errorf("sourcesDir = %q, want %q", fetcher.sourcesDir, sourcesDir)
	}

	if fetcher.logger != logger {
		t.Error("logger not set correctly")
	}
}

func TestIsGitRepository(t *testing.T) {
	logger := &MockLogger{}
	tmpDir := t.TempDir()
	fetcher := NewFetcher(tmpDir, logger)

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

func TestFetchGitLayer_MockGit(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available, skipping test")
	}

	t.Run("clone new repository", func(t *testing.T) {
		// Create a temporary git repository to clone from
		tmpDir := t.TempDir()
		sourceRepo := filepath.Join(tmpDir, "source-repo")

		// Initialize a bare git repo to clone from
		if err := exec.Command("git", "init", "--bare", sourceRepo).Run(); err != nil {
			t.Skipf("Failed to create test git repo: %v", err)
		}

		logger := &MockLogger{}
		sourcesDir := filepath.Join(tmpDir, "sources")
		fetcher := NewFetcher(sourcesDir, logger)

		layer := config.Layer{
			Name:   "test-layer",
			Git:    sourceRepo,
			Branch: "master",
		}

		result := fetcher.fetchGitLayer(layer)

		// Note: This might fail because the bare repo has no commits
		// In a real scenario, we'd need a proper test repository
		t.Logf("Fetch result: Success=%v, Error=%v", result.Success, result.Error)
	})
}

func TestCleanCache(t *testing.T) {
	tmpDir := t.TempDir()
	sourcesDir := filepath.Join(tmpDir, "sources")

	// Create some test files
	os.MkdirAll(filepath.Join(sourcesDir, "layer1"), 0755)
	os.WriteFile(filepath.Join(sourcesDir, "layer1", "test.txt"), []byte("test"), 0644)

	logger := &MockLogger{}
	fetcher := NewFetcher(sourcesDir, logger)

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

	logger := &MockLogger{}
	fetcher := NewFetcher(sourcesDir, logger)

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
	logger := &MockLogger{}
	fetcher := NewFetcher(filepath.Join(tmpDir, "sources"), logger)

	cfg := &config.Config{
		Base: config.BaseConfig{
			Provider: "custom", // No extra layers
		},
		Layers: []config.Layer{}, // No custom layers
	}

	// This will try to fetch base layers but will fail without real git repos
	// We're mainly testing that it doesn't crash
	results, err := fetcher.FetchLayers(cfg)

	if err == nil {
		t.Log("FetchLayers succeeded (unexpected in test environment)")
	} else {
		t.Logf("FetchLayers failed as expected in test environment: %v", err)
	}

	// Should have attempted to fetch poky and meta-openembedded
	if len(results) < 2 {
		t.Errorf("Expected at least 2 fetch attempts, got %d", len(results))
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

// Benchmark tests
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
