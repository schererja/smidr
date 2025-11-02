package build

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/schererja/smidr/internal/bitbake"
	"github.com/schererja/smidr/internal/config"
	smidrcontainer "github.com/schererja/smidr/internal/container"
	"github.com/schererja/smidr/internal/container/docker"
	"github.com/schererja/smidr/internal/source"
	"github.com/schererja/smidr/pkg/logger"
)

// LogSink is a minimal interface for streaming logs
// stream should be "stdout" or "stderr"
type LogSink interface {
	Write(stream string, line string)
}

// BuildOptions captures caller-provided options
type BuildOptions struct {
	BuildID    string
	Target     string
	Customer   string
	ForceClean bool
	ForceImage bool
}

// BuildResult summarizes the build execution
type BuildResult struct {
	Success   bool
	ExitCode  int
	Duration  time.Duration
	BuildDir  string
	TmpDir    string
	DeployDir string
}

// Runner executes the Yocto build pipeline
type Runner struct {
	logger *logger.Logger
}

// NewRunner creates a new build Runner
func NewRunner(logger *logger.Logger) *Runner {
	return &Runner{logger: logger}
}

// Run orchestrates directory setup, layer fetch, container start, bitbake execution, and cleanup
func (r *Runner) Run(ctx context.Context, cfg *config.Config, opts BuildOptions, log LogSink) (*BuildResult, error) {
	start := time.Now()

	// Expand and prepare directories
	expand := func(p string) string {
		if p == "" {
			return p
		}
		if strings.HasPrefix(p, "~") {
			h, _ := os.UserHomeDir()
			p = filepath.Join(h, strings.TrimPrefix(p, "~"))
		}
		if !filepath.IsAbs(p) {
			cwd, _ := os.Getwd()
			p = filepath.Join(cwd, p)
		}
		return p
	}

	// Determine working directory (not strictly needed since we mount cfg.Directories.Build as workspace)
	// Basic defaults similar to CLI
	// IMPORTANT: Use BuildID to ensure each build has isolated directories
	if cfg.Directories.Build == "" {
		h, _ := os.UserHomeDir()
		if opts.BuildID != "" {
			// Use the unique build ID to prevent concurrent build collisions
			cfg.Directories.Build = filepath.Join(h, ".smidr", "builds", opts.BuildID)
		} else if opts.Customer != "" {
			// Fallback: use customer name + UUID for uniqueness
			cfg.Directories.Build = filepath.Join(h, ".smidr", "builds", fmt.Sprintf("build-%s-%s", opts.Customer, uuid.New().String()[:8]))
		} else {
			cfg.Directories.Build = filepath.Join(h, ".smidr", "builds", fmt.Sprintf("build-%s", uuid.New().String()))
		}
	}

	cfg.Directories.Build = expand(cfg.Directories.Build)
	if cfg.Directories.Tmp == "" {
		cfg.Directories.Tmp = filepath.Join(cfg.Directories.Build, "tmp")
	}
	if cfg.Directories.Deploy == "" {
		cfg.Directories.Deploy = filepath.Join(cfg.Directories.Build, "deploy")
	}

	cfg.Directories.Layers = expand(cfg.Directories.Layers)
	cfg.Directories.Source = expand(cfg.Directories.Source)
	cfg.Directories.SState = expand(cfg.Directories.SState)
	cfg.Directories.Downloads = expand(cfg.Directories.Downloads)
	cfg.Directories.Tmp = expand(cfg.Directories.Tmp)
	cfg.Directories.Deploy = expand(cfg.Directories.Deploy)

	// If force clean requested, wipe build directory for a full rebuild
	if opts.ForceClean && strings.TrimSpace(cfg.Directories.Build) != "" {
		_ = os.RemoveAll(cfg.Directories.Build)
	}

	// Ensure tmp and deploy exist and are permissive for container user
	_ = os.MkdirAll(cfg.Directories.Tmp, 0o777)
	_ = os.MkdirAll(cfg.Directories.Deploy, 0o777)

	// Fetch layers
	log.Write("stdout", "📦 Fetching layers...")
	fetcher := source.NewFetcher(cfg.Directories.Layers, cfg.Directories.Downloads, source.NewConsoleLogger(io.Discard, false))
	if _, err := fetcher.FetchLayers(cfg); err != nil {
		return &BuildResult{Success: false, Duration: time.Since(start), BuildDir: cfg.Directories.Build, TmpDir: cfg.Directories.Tmp, DeployDir: cfg.Directories.Deploy}, fmt.Errorf("failed to fetch layers: %w", err)
	}

	// Prepare container config and manager (mirror CLI behavior)
	// Determine container image
	imageToUse := cfg.Container.BaseImage
	if strings.TrimSpace(imageToUse) == "" {
		imageToUse = "crops/yocto:ubuntu-22.04-builder"
	}

	// Build layer mount list from cfg.Layers, mounting parent folder for sublayers
	var layerDirs []string
	var layerNames []string
	mountedParents := make(map[string]bool)
	for _, l := range cfg.Layers {
		var layerPath string
		if l.Path != "" {
			if !strings.HasPrefix(l.Path, "/") && !strings.HasPrefix(l.Path, "~") {
				layerPath = filepath.Join(cfg.Directories.Layers, l.Path)
			} else {
				layerPath = l.Path
			}
		} else {
			layerPath = filepath.Join(cfg.Directories.Layers, l.Name)
		}
		// Determine mount parent and name
		var mountPath string
		var mountName string
		if l.Path != "" && strings.Contains(l.Path, "/") {
			parent := strings.Split(l.Path, "/")[0]
			mountPath = filepath.Join(cfg.Directories.Layers, parent)
			mountName = parent
		} else {
			mountPath = layerPath
			mountName = l.Name
		}
		if !mountedParents[mountPath] {
			// Ensure it exists for bind mount
			_ = os.MkdirAll(expand(mountPath), 0o755)
			layerDirs = append(layerDirs, expand(mountPath))
			layerNames = append(layerNames, mountName)
			mountedParents[mountPath] = true
		}
	}

	// Container environment - only pass specific variables, not all host env vars
	// IMPORTANT: Do NOT export host TMPDIR into the container. BitBake interprets
	// TMPDIR and will then try to use that path inside the container, which can
	// point to a non-existent/unwritable host path (e.g. /home/<host>/.smidr/tmp)
	// and stall the build. Let oe-init-build-env set TMPDIR to "tmp" under the
	// per-build workspace instead.
	env := []string{
		"HOME=/home/builder",
		"USER=builder",
	}
	// Optionally pass through specific needed vars (e.g., proxy settings)
	for _, key := range []string{"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY", "http_proxy", "https_proxy", "no_proxy"} {
		if val := os.Getenv(key); val != "" {
			env = append(env, fmt.Sprintf("%s=%s", key, val))
		}
	}

	// Assign a unique container name based on build ID to prevent collisions
	containerName := ""
	containerWorkspace := "/home/builder/build" // default workspace path inside container
	if opts.BuildID != "" {
		containerName = fmt.Sprintf("smidr-build-%s", opts.BuildID)
		// CRITICAL: Use unique workspace path inside container to prevent BitBake server collisions
		// Each build gets its own workspace so BitBake lock files don't conflict
		containerWorkspace = fmt.Sprintf("/home/builder/build-%s", opts.BuildID)
	}

	containerCfg := smidrcontainer.ContainerConfig{
		Image:                imageToUse,
		Name:                 containerName, // unique per build to prevent Docker name conflicts
		Env:                  env,
		Cmd:                  []string{"echo 'Container ready' && sleep 86400"},
		DownloadsDir:         cfg.Directories.Downloads,
		SstateCacheDir:       cfg.Directories.SState,
		BuildDir:             cfg.Directories.Build,
		WorkspaceDir:         cfg.Directories.Build,
		WorkspaceMountTarget: containerWorkspace, // unique path inside container prevents BitBake collisions
		LayerDirs:            layerDirs,
		LayerNames:           layerNames,
		MemoryLimit:          cfg.Container.Memory,
		CPUCount:             cfg.Container.CPUCount,
		// Keep host tmp mounted if configured by user for auxiliary tooling,
		// but it will NOT be used by BitBake unless explicitly set in local.conf.
		TmpDir: cfg.Directories.Tmp,
	}

	dm, err := docker.NewDockerManager()
	if err != nil {
		return &BuildResult{Success: false, Duration: time.Since(start)}, fmt.Errorf("failed to create docker manager: %w", err)
	}

	// Image pull/check
	if !dm.ImageExists(ctx, containerCfg.Image) {
		log.Write("stdout", fmt.Sprintf("🐳 Pulling image %s...", containerCfg.Image))
		if err := dm.PullImage(ctx, containerCfg.Image); err != nil {
			return &BuildResult{Success: false, Duration: time.Since(start)}, fmt.Errorf("failed to pull image: %w", err)
		}
	}

	// Create container
	containerID, err := dm.CreateContainer(ctx, containerCfg)
	if err != nil {
		return &BuildResult{Success: false, Duration: time.Since(start)}, fmt.Errorf("failed to create container: %w", err)
	}
	defer func() {
		_ = dm.StopContainer(context.Background(), containerID, 2*time.Second)
		_ = dm.RemoveContainer(context.Background(), containerID, true)
	}()

	// Start container
	if err := dm.StartContainer(ctx, containerID); err != nil {
		return &BuildResult{Success: false, Duration: time.Since(start)}, fmt.Errorf("failed to start container: %w", err)
	}

	// Run bitbake
	// Pass the container's workspace path (not host path) so BitBake runs in the right directory
	executor := bitbake.NewBuildExecutor(cfg, dm, containerID, containerWorkspace, r.logger)
	executor.SetForceImage(opts.ForceImage)

	// Set build prefix for log identification (e.g., "[customer/build-abc123]")
	if opts.Customer != "" && opts.BuildID != "" {
		executor.SetBuildPrefix(fmt.Sprintf("[%s/%s]", opts.Customer, opts.BuildID[:8]))
	} else if opts.BuildID != "" {
		executor.SetBuildPrefix(fmt.Sprintf("[%s]", opts.BuildID[:8]))
	}

	// Log adapter to forward bitbake output into LogSink
	// Only forward important messages (errors, warnings, summaries) to reduce noise
	bbLog := &bitbake.BuildLogWriter{
		PlainWriter: logWriterFunc(func(p []byte) (int, error) {
			// Split and forward lines to provided sink
			for _, line := range strings.Split(string(p), "\n") {
				trimmed := strings.TrimSpace(line)
				if trimmed == "" {
					continue
				}
				// Only log important lines: errors, warnings, summaries, and key status updates
				if shouldLogLine(trimmed) {
					log.Write("stdout", line)
				}
			}
			return len(p), nil
		}),
		JSONLWriter: nil,
	}

	result, err := executor.ExecuteBuild(ctx, bbLog)
	exitCode := 0
	if result != nil {
		exitCode = result.ExitCode
	}
	br := &BuildResult{Success: err == nil && result != nil && result.Success, ExitCode: exitCode, Duration: time.Since(start), BuildDir: cfg.Directories.Build, TmpDir: cfg.Directories.Tmp, DeployDir: cfg.Directories.Deploy}
	if err != nil {
		return br, err
	}
	return br, nil
}

type logWriterFunc func(p []byte) (n int, err error)

func (f logWriterFunc) Write(p []byte) (n int, err error) { return f(p) }

// shouldLogLine determines if a BitBake log line should be forwarded to the daemon
// Only logs errors, warnings, summaries, and key status updates to reduce noise
func shouldLogLine(line string) bool {
	// Always log errors and warnings
	if strings.Contains(line, "ERROR") ||
		strings.Contains(line, "FAILED") ||
		strings.Contains(line, "WARNING") ||
		strings.Contains(line, "WARN") {
		return true
	}

	// Log summary and completion messages
	if strings.HasPrefix(line, "Summary:") ||
		strings.HasPrefix(line, "NOTE: Tasks Summary:") ||
		strings.Contains(line, "Build completed") ||
		strings.Contains(line, "succeeded.") {
		return true
	}

	// Log key status markers (emoji-prefixed messages from our scripts)
	if strings.HasPrefix(line, "🚀") ||
		strings.HasPrefix(line, "✅") ||
		strings.HasPrefix(line, "❌") ||
		strings.HasPrefix(line, "📦") ||
		strings.HasPrefix(line, "⬇️") ||
		strings.HasPrefix(line, "📺") ||
		strings.HasPrefix(line, "===") {
		return true
	}

	// Log Initialising and Sstate summary (high-level progress)
	if strings.HasPrefix(line, "Initialising tasks") ||
		strings.HasPrefix(line, "Sstate summary:") ||
		strings.HasPrefix(line, "NOTE: Executing Tasks") {
		return true
	}

	// Skip verbose task-by-task "NOTE: Running task N of M" and "NOTE: recipe X: task Y: Started/Succeeded"
	// These create hundreds of lines of noise
	if strings.HasPrefix(line, "NOTE: Running task") ||
		strings.HasPrefix(line, "NOTE: recipe") ||
		strings.HasPrefix(line, "NOTE: Reconnecting") ||
		strings.HasPrefix(line, "NOTE: No reply") ||
		strings.HasPrefix(line, "NOTE: Retrying") {
		return false
	}

	// Skip shell command echoes and environment setup noise
	if strings.HasPrefix(line, "+") ||
		strings.HasPrefix(line, "++") ||
		strings.HasPrefix(line, "+++") {
		return false
	}

	// Skip layer metadata and config dumps
	if strings.Contains(line, "meta-") && strings.Contains(line, "=") {
		return false
	}

	// Skip empty bitbake status messages
	if strings.HasPrefix(line, "Bitbake still alive") {
		return false
	}

	// Default: skip (most BitBake output is verbose noise)
	return false
}
