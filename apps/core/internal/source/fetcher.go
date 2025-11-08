package source

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/schererja/smidr/internal/config"
)

// Fetcher is responsible for fetching source code from various repositories.
type Fetcher struct {
	layersDir    string
	downloadsDir string
	logger       Logger
	mu           sync.Mutex
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
func NewFetcher(layersDir string, downloadsDir string, logger Logger) *Fetcher {
	return &Fetcher{
		layersDir:    layersDir,
		downloadsDir: downloadsDir,
		logger:       logger,
	}
}

// ----------------------------------------------------------------
// Public methods
// ----------------------------------------------------------------
// FetchLayers fetches all required layers for a configuration
func (f *Fetcher) FetchLayers(cfg *config.Config) ([]FetchResult, error) {
	if err := os.MkdirAll(f.layersDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create layers directory: %w", err)
	}

	// Collect unique source URLs and their config.Layer (first occurrence wins)
	sourceMap := make(map[string]config.Layer)
	for _, layer := range cfg.Layers {
		if layer.Git != "" {
			if _, exists := sourceMap[layer.Git]; !exists {
				// If no branch is set, use yocto_series as default branch
				if layer.Branch == "" && cfg.YoctoSeries != "" {
					layer.Branch = cfg.YoctoSeries
				}
				sourceMap[layer.Git] = layer
			}
		}
	}

	var results []FetchResult
	var wg sync.WaitGroup
	resultsChan := make(chan FetchResult, 32)

	// Fetch each unique repo into layersDir
	for _, layer := range sourceMap {
		wg.Add(1)
		go func(l config.Layer) {
			defer wg.Done()
			result := f.fetchGitLayerTo(l, f.layersDir)
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
	entries, err := os.ReadDir(f.layersDir)
	if err != nil {
		return err
	}
	now := time.Now()
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		repoPath := filepath.Join(f.layersDir, entry.Name())
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

	if err := os.RemoveAll(f.layersDir); err != nil {
		return fmt.Errorf("failed to clean cache: %w", err)
	}

	return os.MkdirAll(f.layersDir, 0755)
}

// GetCacheSize returns the total size of cached sources in bytes
func (f *Fetcher) GetCacheSize() (int64, error) {
	var size int64

	err := filepath.Walk(f.layersDir, func(path string, info os.FileInfo, err error) error {
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

	return f.fetchGitLayerTo(layer, f.layersDir)
}

// fetchGitLayer clones or updates a git repository
// fetchGitLayer is now deprecated; use fetchGitLayerTo with layersDir instead.

// fetchGitLayerTo clones or updates a git repository into a specified baseDir
func (f *Fetcher) fetchGitLayerTo(layer config.Layer, baseDir string) FetchResult {
	layerPath := filepath.Join(baseDir, layer.Name)

	// Acquire per-repo lock to avoid concurrent clones/updates across processes
	lockFile := filepath.Join(baseDir, layer.Name+".lock")
	locked, lockErr := acquireLock(lockFile, 10*time.Second)
	if lockErr != nil {
		return FetchResult{LayerName: layer.Name, Path: layerPath, Success: false, Error: fmt.Errorf("failed to acquire lock: %w", lockErr)}
	}
	defer func() {
		if locked {
			_ = releaseLock(lockFile)
		}
	}()

	// Check if already exists
	if f.isGitRepository(layerPath) {
		f.logger.Debug("Layer %s already exists at %s, checking status...", layer.Name, layerPath)
		if err := f.updateGitRepository(layerPath, layer.Branch); err != nil {
			f.logger.Debug("Failed to update %s: %v", layer.Name, err)
		}
		return FetchResult{LayerName: layer.Name, Path: layerPath, Success: true, Cached: true}
	}

	// Clone the repository
	branch := layer.Branch // Restore original branch handling

	cleanURL := strings.TrimSuffix(layer.Git, ".git")
	f.logger.Info("Cloning layer %s from %s (branch: %s) into %s", layer.Name, cleanURL, branch, baseDir)
	useShallow := !strings.Contains(layer.Git, "git.toradex.com")

	var cmd *exec.Cmd
	var stderr strings.Builder
	tryBranch := branch // Restore original branch handling
	cloneSucceeded := false
	cloneErr := error(nil)
	// Try initial branch
	if useShallow {
		cmd = exec.Command("git", "clone", "--branch", tryBranch, "--depth", "1", layer.Git, layerPath)
	} else {
		cmd = exec.Command("git", "clone", "--branch", tryBranch, layer.Git, layerPath)
	}
	cmd.Stderr = &stderr
	if err := cmd.Run(); err == nil {
		cloneSucceeded = true
	} else {
		// If failed, try to find a branch with the correct prefix
		errorMsg := stderr.String()
		if errorMsg == "" {
			errorMsg = err.Error()
		}
		f.logger.Debug("Git clone failed for %s: %s", layer.Name, errorMsg)
		// Always try prefix match if the branch matches yocto_series (even if set by default logic)
		// (Assume yocto_series is used as default branch if not set by user)
		// Use the value of 'branch' as the prefix for matching
		if branch != "master" && branch != "main" && branch != "" {
			// List remote branches
			lsRemoteCmd := exec.Command("git", "ls-remote", "--heads", layer.Git)
			var lsOut strings.Builder
			lsRemoteCmd.Stdout = &lsOut
			_ = lsRemoteCmd.Run()
			lines := strings.Split(lsOut.String(), "\n")
			var bestMatch string
			for _, line := range lines {
				parts := strings.Fields(line)
				if len(parts) == 2 {
					ref := parts[1]
					if strings.HasPrefix(ref, "refs/heads/"+branch+"-") {
						bestMatch = strings.TrimPrefix(ref, "refs/heads/")
						break // take first match
					}
				}
			}
			if bestMatch != "" {
				f.logger.Info("Retrying clone for %s with branch %s (prefix match)", layer.Name, bestMatch)
				if useShallow {
					cmd = exec.Command("git", "clone", "--branch", bestMatch, "--depth", "1", layer.Git, layerPath)
				} else {
					cmd = exec.Command("git", "clone", "--branch", bestMatch, layer.Git, layerPath)
				}
				stderr.Reset()
				cmd.Stderr = &stderr
				if err := cmd.Run(); err == nil {
					cloneSucceeded = true
				} else {
					errorMsg = stderr.String()
					if errorMsg == "" {
						errorMsg = err.Error()
					}
					cloneErr = fmt.Errorf("git clone failed (prefix match): %s", errorMsg)
				}
			} else {
				cloneErr = fmt.Errorf("git clone failed: %s", errorMsg)
			}
		} else {
			cloneErr = fmt.Errorf("git clone failed: %s", errorMsg)
		}
	}

	if !cloneSucceeded {
		return FetchResult{LayerName: layer.Name, Path: layerPath, Success: false, Error: cloneErr}
	}
	return FetchResult{LayerName: layer.Name, Path: layerPath, Success: true, Cached: false}
}

// isGitRepository checks if a directory is a git repository
func (f *Fetcher) isGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// acquireLock attempts to create a lockfile atomically. It will retry until timeout.
func acquireLock(path string, timeout time.Duration) (bool, error) {
	// Ensure parent directory exists so OpenFile with O_CREATE doesn't fail with ENOENT
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return false, fmt.Errorf("failed to create lock directory: %w", err)
		}
	}

	deadline := time.Now().Add(timeout)
	for {
		fd, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err == nil {
			// Write PID and timestamp for diagnostics
			_, _ = fmt.Fprintf(fd, "pid:%d\nstarted:%s\n", os.Getpid(), time.Now().Format(time.RFC3339))
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
	if branch != "" { // Restore original branch handling
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
	// Primary repository URLs
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

// getAlternativeRepository returns alternative/mirror URLs for layers that might have connectivity issues
func getAlternativeRepository(layerName string) []string {
	alternatives := map[string][]string{
		"meta-toradex-nxp": {
			"https://git.toradex.com/meta-toradex-nxp.git/", // with trailing slash
			"https://git.toradex.com/meta-toradex-nxp",      // without .git extension
		},
		"meta-toradex-bsp-common": {
			"https://git.toradex.com/meta-toradex-bsp-common.git/", // with trailing slash
			"https://git.toradex.com/meta-toradex-bsp-common",      // without .git extension
		},
	}

	return alternatives[layerName]
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
