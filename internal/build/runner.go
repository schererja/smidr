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
)

// LogSink is a minimal interface for streaming logs
// stream should be "stdout" or "stderr"
type LogSink interface {
	Write(stream string, line string)
}

// BuildOptions captures caller-provided options
type BuildOptions struct {
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
type Runner struct{}

// NewRunner creates a new build Runner
func NewRunner() *Runner { return &Runner{} }

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
	if cfg.Directories.Build == "" {
		h, _ := os.UserHomeDir()
		if opts.Customer != "" {
			cfg.Directories.Build = filepath.Join(h, ".smidr", "builds", fmt.Sprintf("build-%s", opts.Customer))
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

	// Container environment
	env := os.Environ()
	env = append(env, fmt.Sprintf("TMPDIR=%s", cfg.Directories.Tmp))

	containerCfg := smidrcontainer.ContainerConfig{
		Image:          imageToUse,
		Name:           "", // let Docker assign unless tests override later
		Env:            env,
		Cmd:            []string{"echo 'Container ready' && sleep 86400"},
		DownloadsDir:   cfg.Directories.Downloads,
		SstateCacheDir: cfg.Directories.SState,
		BuildDir:       cfg.Directories.Build,
		WorkspaceDir:   cfg.Directories.Build,
		LayerDirs:      layerDirs,
		LayerNames:     layerNames,
		MemoryLimit:    cfg.Container.Memory,
		CPUCount:       cfg.Container.CPUCount,
		TmpDir:         cfg.Directories.Tmp,
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
	executor := bitbake.NewBuildExecutor(cfg, dm, containerID, containerCfg.WorkspaceDir)
	executor.SetForceImage(opts.ForceImage)

	// Log adapter to forward bitbake output into LogSink
	bbLog := &bitbake.BuildLogWriter{
		PlainWriter: logWriterFunc(func(p []byte) (int, error) {
			// Split and forward lines to provided sink
			for _, line := range strings.Split(string(p), "\n") {
				if strings.TrimSpace(line) != "" {
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
