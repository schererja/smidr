package source

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/intrik8-labs/smidr/internal/config"
)


// Fetcher is responsible for fetching source code from various repositories.
type Fetcher struct {
	sourcesDir string
	logger     Logger
	mu         sync.Mutex
}

type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

type FetchResult struct {
	LayerName string
	Path      string
	Success   bool
	Error     error
	Cached    bool // Meaning already cloned
}

// NewFetcher creates a new Fetcher instance
func NewFetcher(sourcesDir string, logger Logger) *Fetcher {
	return &Fetcher{
		sourcesDir: sourcesDir,
		logger:     logger,
	}
}

// ----------------------------------------------------------------
// Public methods
// ----------------------------------------------------------------
// FetchLayers fetches all required layers for a configuration
func (f *Fetcher) FetchLayers(cfg *config.Config) ([]FetchResult, error) {
	if err := os.MkdirAll(f.sourcesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sources directory: %w", err)
	}

	// Track which layers we need to fetch (deduplicate by name)
	layersToFetch := make(map[string]config.Layer)

	// Add base layers
	baseLayerNames := getRequiredBaseLayers(cfg.Base.Provider)
	for _, layerName := range baseLayerNames {
		repo := getBaseLayerRepository(layerName)
		branch := getBranchForLayer(layerName, cfg.Base.Version)

		layersToFetch[layerName] = config.Layer{
			Name:   layerName,
			Git:    repo,
			Branch: branch,
		}
	}

	// Add custom layers from config (may override base layers)
	for _, layer := range cfg.Layers {
		if layer.Git != "" {
			// Custom layer overrides base layer if same name
			layersToFetch[layer.Name] = layer
		}
	}

	// Fetch all layers in parallel
	var results []FetchResult
	var wg sync.WaitGroup
	resultsChan := make(chan FetchResult, len(layersToFetch))

	for _, layer := range layersToFetch {
		wg.Add(1)
		go func(l config.Layer) {
			defer wg.Done()
			result := f.fetchGitLayer(l)
			resultsChan <- result
		}(layer)
	}

	// Wait for all fetches to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	for result := range resultsChan {
		results = append(results, result)
		if result.Success {
			if result.Cached {
				f.logger.Info("Layer %s already cached at %s", result.LayerName, result.Path)
				_ = writeCacheMeta(filepath.Join(result.Path, ".smidr_meta.json"))
			} else {
				f.logger.Info("Successfully fetched layer %s to %s", result.LayerName, result.Path)
				_ = writeCacheMeta(filepath.Join(result.Path, ".smidr_meta.json"))
			}
		} else {
			f.logger.Error("Failed to fetch layer %s: %v", result.LayerName, result.Error)
		}
	}
	return results, nil
}

// EvictOldCache removes cached repos not accessed within ttl
func (f *Fetcher) EvictOldCache(ttl time.Duration) error {
	entries, err := os.ReadDir(f.sourcesDir)
	if err != nil {
		return err
	}
	now := time.Now()
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		repoPath := filepath.Join(f.sourcesDir, entry.Name())
	meta, err := readCacheMeta(filepath.Join(repoPath, ".smidr_meta.json"))
		if err != nil {
			// If no meta, skip eviction (could be in use or legacy)
			continue
		}
		if now.Sub(meta.LastAccess) > ttl {
			f.logger.Info("Evicting repo %s (last accessed %s)", entry.Name(), meta.LastAccess.Format(time.RFC3339))
			_ = os.RemoveAll(repoPath)
		}
	}
	return nil
}


// CleanCache removes all cached sources
func (f *Fetcher) CleanCache() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := os.RemoveAll(f.sourcesDir); err != nil {
		return fmt.Errorf("failed to clean cache: %w", err)
	}

	return os.MkdirAll(f.sourcesDir, 0755)
}

// GetCacheSize returns the total size of cached sources in bytes
func (f *Fetcher) GetCacheSize() (int64, error) {
	var size int64

	err := filepath.Walk(f.sourcesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// ---------------------------------------------------------------
// Internal methods
// ---------------------------------------------------------------
// fetchBaseLayer fetches a standard base layer (poky, meta-openembedded, etc.)
func (f *Fetcher) fetchBaseLayer(layerName string, cfg *config.Config) FetchResult {
	repo := getBaseLayerRepository(layerName)
	if repo == "" {
		return FetchResult{
			LayerName: layerName,
			Success:   false,
			Error:     fmt.Errorf("unknown base layer: %s", layerName),
		}
	}

	// Determine branch based on version or use default
	branch := getBranchForLayer(layerName, cfg.Base.Version)

	layer := config.Layer{
		Name:   layerName,
		Git:    repo,
		Branch: branch,
	}

	return f.fetchGitLayer(layer)
}

// fetchGitLayer clones or updates a git repository
func (f *Fetcher) fetchGitLayer(layer config.Layer) FetchResult {
	layerPath := filepath.Join(f.sourcesDir, layer.Name)

	// Acquire per-repo lock to avoid concurrent clones/updates across processes
	lockFile := filepath.Join(f.sourcesDir, layer.Name+".lock")
	locked, lockErr := acquireLock(lockFile, 10*time.Second)
	if lockErr != nil {
		return FetchResult{LayerName: layer.Name, Path: layerPath, Success: false, Error: fmt.Errorf("failed to acquire lock: %w", lockErr)}
	}
	// Ensure lock is released when done
	defer func() {
		if locked {
			_ = releaseLock(lockFile)
		}
	}()

	// Check if already exists
	if f.isGitRepository(layerPath) {
		f.logger.Debug("Layer %s already exists, checking status...", layer.Name)

		// Try to update existing repository
		if err := f.updateGitRepository(layerPath, layer.Branch); err != nil {
			f.logger.Debug("Failed to update %s: %v", layer.Name, err)
			// Don't fail - just use existing repository
		}

		return FetchResult{
			LayerName: layer.Name,
			Path:      layerPath,
			Success:   true,
			Cached:    true,
		}
	}

	// Clone the repository
	branch := layer.Branch
	if branch == "" {
		branch = "master" // Default branch
	}

	// Clean up URL (remove trailing .git if present for logging)
	cleanURL := strings.TrimSuffix(layer.Git, ".git")
	f.logger.Info("Cloning layer %s from %s (branch: %s)", layer.Name, cleanURL, branch)
	// Determine if we should use shallow clone
	useShallow := !strings.Contains(layer.Git, "git.toradex.com")

	// Build git clone command
	var cmd *exec.Cmd
	if useShallow {
		cmd = exec.Command("git", "clone", "--branch", branch, "--depth", "1", layer.Git, layerPath)
	} else {
		// Toradex repos don't support shallow clones over HTTP
		cmd = exec.Command("git", "clone", "--branch", branch, layer.Git, layerPath)
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// If clone failed, provide detailed error
		errorMsg := stderr.String()
		if errorMsg == "" {
			errorMsg = err.Error()
		}

		f.logger.Debug("Git clone failed for %s: %s", layer.Name, errorMsg)

		// Try without branch specification if branch clone failed
		if strings.Contains(errorMsg, "Remote branch") || strings.Contains(errorMsg, "not found") {
			f.logger.Debug("Branch %s not found, trying default branch...", branch)
			// Cleanup failed clone attempt
			os.RemoveAll(layerPath)

			if useShallow {
				cmd = exec.Command("git", "clone", "--depth", "1", layer.Git, layerPath)
			} else {
				cmd = exec.Command("git", "clone", layer.Git, layerPath)
			}
			stderr.Reset()
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				return FetchResult{
					LayerName: layer.Name,
					Path:      layerPath,
					Success:   false,
					Error:     fmt.Errorf("git clone failed: %s", stderr.String()),
				}
			}

			// Successfully cloned with default branch
			return FetchResult{
				LayerName: layer.Name,
				Path:      layerPath,
				Success:   true,
				Cached:    false,
			}
		}

		return FetchResult{
			LayerName: layer.Name,
			Path:      layerPath,
			Success:   false,
			Error:     fmt.Errorf("git clone failed: %s", errorMsg),
		}
	}

	return FetchResult{
		LayerName: layer.Name,
		Path:      layerPath,
		Success:   true,
		Cached:    false,
	}
}

// isGitRepository checks if a directory is a git repository
func (f *Fetcher) isGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// acquireLock attempts to create a lockfile atomically. It will retry until timeout.
func acquireLock(path string, timeout time.Duration) (bool, error) {
	deadline := time.Now().Add(timeout)
	for {
		fd, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err == nil {
			// Write PID and timestamp for diagnostics
			_, _ = fd.WriteString(fmt.Sprintf("pid:%d\nstarted:%s\n", os.Getpid(), time.Now().Format(time.RFC3339)))
			_ = fd.Close()
			return true, nil
		}

		if !os.IsExist(err) {
			return false, err
		}

		if time.Now().After(deadline) {
			return false, fmt.Errorf("timed out waiting for lock %s", path)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// releaseLock removes the lockfile.
func releaseLock(path string) error {
	return os.Remove(path)
}

// updateGitRepository updates an existing git repository
func (f *Fetcher) updateGitRepository(path string, branch string) error {
	// Fetch latest changes
	fetchCmd := exec.Command("git", "-C", path, "fetch", "origin")
	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("git fetch failed: %w", err)
	}

	// Checkout the desired branch if specified
	if branch != "" {
		checkoutCmd := exec.Command("git", "-C", path, "checkout", branch)
		if err := checkoutCmd.Run(); err != nil {
			return fmt.Errorf("git checkout failed: %w", err)
		}
	}

	// Pull latest changes
	pullCmd := exec.Command("git", "-C", path, "pull", "--ff-only")
	if err := pullCmd.Run(); err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}

	return nil
}

func (f *Fetcher) fetchLayer(layerName string, cfg *config.Config) FetchResult {
	repo := getBaseLayerRepository(layerName)
	if repo == "" {
		return FetchResult{
			LayerName: layerName,
			Success:   false,
			Error:     fmt.Errorf("unknown base layer:  %s", layerName),
		}
	}
	return FetchResult{
		LayerName: layerName,
	}
}

func (f *Fetcher) getBaseLayerRepo(layerName string) string {
	return ""
}

// ---------------------------------------------------------------
// Helper functions for base layers
// ---------------------------------------------------------------
func getRequiredBaseLayers(provider string) []string {
	layers := []string{"poky", "meta-openembedded"}

	switch provider {
	case "toradex":
		layers = append(layers, "meta-toradex-nxp", "meta-toradex-bsp-common")
	case "raspberrypi":
		layers = append(layers, "meta-raspberrypi")
	case "nvidia":
		layers = append(layers, "meta-tegra")
	}

	return layers
}

func getBaseLayerRepository(layerName string) string {
	repos := map[string]string{
		"poky":                    "https://git.yoctoproject.org/poky",
		"meta-openembedded":       "https://git.openembedded.org/meta-openembedded",
		"meta-toradex-nxp":        "https://git.toradex.com/meta-toradex-nxp.git",
		"meta-toradex-bsp-common": "https://git.toradex.com/meta-toradex-bsp-common.git",
		"meta-raspberrypi":        "https://git.yoctoproject.org/meta-raspberrypi",
		"meta-tegra":              "https://github.com/OE4T/meta-tegra",
	}

	return repos[layerName]
}

func getBranchForLayer(layerName, version string) string {
	// Map version to Yocto release codenames
	versionToBranch := map[string]string{
		"6.0.0": "kirkstone",
		"5.0.0": "dunfell",
		"4.0.0": "zeus",
	}

	// If version is specified, try to map it
	if version != "" {
		if branch, ok := versionToBranch[version]; ok {
			// Special handling for Toradex layers
			if strings.Contains(layerName, "toradex") {
				return branch + "-6.x.y"
			}
			return branch
		}
	}
	// Default branches for specific layers
	defaultBranches := map[string]string{
		"poky":                    "kirkstone",
		"meta-openembedded":       "kirkstone",
		"meta-toradex-nxp":        "kirkstone-6.x.y",
		"meta-toradex-bsp-common": "kirkstone-6.x.y",
		"meta-raspberrypi":        "kirkstone",
		"meta-tegra":              "kirkstone",
	}

	if branch, ok := defaultBranches[layerName]; ok {
		return branch
	}

	return "master"
}
