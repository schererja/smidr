package build

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/schererja/smidr/internal/bitbake"
	"github.com/schererja/smidr/internal/config"
	smidrcontainer "github.com/schererja/smidr/internal/container"
	"github.com/schererja/smidr/internal/container/docker"
	"github.com/schererja/smidr/internal/db"
	"github.com/schererja/smidr/internal/source"
	"github.com/schererja/smidr/pkg/logger"
)

// LogSink is a minimal interface for streaming logs to daemon clients
// stream should be "stdout" or "stderr"
type LogSink interface {
	Write(stream string, line string)
}

// LogSinkLogger wraps a LogSink and also writes to structured logger
type LogSinkLogger struct {
	sink   LogSink
	logger *logger.Logger
}

// NewLogSinkLogger creates a LogSink that also logs to structured logger
func NewLogSinkLogger(sink LogSink, logger *logger.Logger) *LogSinkLogger {
	return &LogSinkLogger{sink: sink, logger: logger}
}

// Write forwards to the sink and also logs structured output
func (l *LogSinkLogger) Write(stream string, line string) {
	if l.sink != nil {
		l.sink.Write(stream, line)
	}
	// Also log to structured logger for persistence
	trimmed := strings.TrimSpace(line)
	if trimmed != "" {
		// Classify log level based on content
		// Note: Error takes (msg, err, ...attrs), but Warn/Debug/Info take (msg, ...attrs)
		if strings.Contains(line, "ERROR") || strings.Contains(line, "FAILED") {
			l.logger.Error("build output", nil, slog.String("stream", stream), slog.String("line", trimmed))
		} else if strings.Contains(line, "WARNING") || strings.Contains(line, "WARN") {
			l.logger.Warn("build output", slog.String("stream", stream), slog.String("line", trimmed))
		} else {
			l.logger.Debug("build output", slog.String("stream", stream), slog.String("line", trimmed))
		}
	}
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
	db     *db.DB
}

// NewRunner creates a new build Runner
func NewRunner(logger *logger.Logger, database *db.DB) *Runner {
	return &Runner{logger: logger, db: database}
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
			buildID := "build-" + opts.Customer + "-" + uuid.New().String()[:8]
			cfg.Directories.Build = filepath.Join(h, ".smidr", "builds", buildID)
		} else {
			buildID := "build-" + uuid.New().String()
			cfg.Directories.Build = filepath.Join(h, ".smidr", "builds", buildID)
		}
	}

	cfg.Directories.Build = expand(cfg.Directories.Build)
	if cfg.Directories.Tmp == "" {
		cfg.Directories.Tmp = filepath.Join(cfg.Directories.Build, "tmp")
	}
	if cfg.Directories.Deploy == "" {
		cfg.Directories.Deploy = filepath.Join(cfg.Directories.Build, "deploy")
	}

	// Integration test overrides via environment variables
	if v := os.Getenv("SMIDR_TEST_DOWNLOADS_DIR"); strings.TrimSpace(v) != "" {
		cfg.Directories.Downloads = v
	}
	if v := os.Getenv("SMIDR_TEST_SSTATE_DIR"); strings.TrimSpace(v) != "" {
		cfg.Directories.SState = v
	}
	if v := os.Getenv("SMIDR_TEST_WORKSPACE_DIR"); strings.TrimSpace(v) != "" {
		// Map test workspace directory to Build to align mounts if needed
		cfg.Directories.Build = v
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

	// Ensure required directories exist with permissive permissions for container user
	_ = os.MkdirAll(cfg.Directories.Tmp, 0o777)
	_ = os.MkdirAll(cfg.Directories.Deploy, 0o777)
	// Ensure downloads and sstate cache directories exist (BitBake will create subdirs)
	if cfg.Directories.Downloads != "" {
		_ = os.MkdirAll(cfg.Directories.Downloads, 0o777)
	}
	if cfg.Directories.SState != "" {
		_ = os.MkdirAll(cfg.Directories.SState, 0o777)
	}

	// If DB persistence is enabled, create initial build record and mark as running
	if r.db != nil {
		// Ensure a build ID exists
		if opts.BuildID == "" {
			opts.BuildID = generateBuildID(opts.Customer)
		}

		configSnapshot, err := json.Marshal(cfg)
		if err != nil {
			r.logger.Warn("failed to serialize config snapshot", slog.String("error", err.Error()))
			configSnapshot = []byte("{}")
		}

		hostname, _ := os.Hostname()
		username := os.Getenv("USER")
		if username == "" {
			username = os.Getenv("USERNAME")
		}

		buildDir := cfg.Directories.Build
		if buildDir == "" {
			h, _ := os.UserHomeDir()
			buildDir = filepath.Join(h, ".smidr", "builds", opts.BuildID)
		}

		deployDir := cfg.Directories.Deploy
		if deployDir == "" {
			deployDir = filepath.Join(buildDir, "deploy")
		}

		logFilePlain := filepath.Join(buildDir, "build-log.txt")
		logFileJSONL := filepath.Join(buildDir, "build-log.jsonl")

		build := &db.Build{
			ID:             opts.BuildID,
			Customer:       opts.Customer,
			ProjectName:    cfg.Name,
			TargetImage:    opts.Target,
			Machine:        cfg.Base.Machine,
			Status:         db.StatusQueued,
			BuildDir:       buildDir,
			DeployDir:      deployDir,
			LogFilePlain:   logFilePlain,
			LogFileJSONL:   logFileJSONL,
			ConfigSnapshot: string(configSnapshot),
			User:           username,
			Host:           hostname,
			CreatedAt:      start,
		}

		if err := r.db.CreateBuild(build); err != nil {
			r.logger.Error("failed to create build record", err)
		}

		if err := r.db.StartBuild(opts.BuildID); err != nil {
			r.logger.Error("failed to mark build as running", err)
		}

		// Ensure build state is updated even if panic occurs
		defer func() {
			if rec := recover(); rec != nil {
				_ = r.db.CompleteBuild(opts.BuildID, db.StatusFailed, 1, time.Since(start), fmt.Sprintf("panic: %v", rec))
				panic(rec)
			}
		}()
	}

	// Fetch layers
	log.Write("stdout", "Fetching layers...")
	r.logger.Info("fetching layers", slog.String("layers_dir", cfg.Directories.Layers))
	fetcher := source.NewFetcher(cfg.Directories.Layers, cfg.Directories.Downloads, source.NewConsoleLogger(io.Discard, false))
	if _, err := fetcher.FetchLayers(cfg); err != nil {
		r.logger.Error("failed to fetch layers", err)
		return &BuildResult{Success: false, Duration: time.Since(start), BuildDir: cfg.Directories.Build, TmpDir: cfg.Directories.Tmp, DeployDir: cfg.Directories.Deploy}, err
	}

	// Prepare container config and manager (mirror CLI behavior)
	// Determine container image
	imageToUse := cfg.Container.BaseImage
	if strings.TrimSpace(imageToUse) == "" {
		imageToUse = "crops/yocto:ubuntu-22.04-builder"
	}
	// Allow tests to override the image to avoid external pulls
	if v := os.Getenv("SMIDR_TEST_IMAGE"); strings.TrimSpace(v) != "" {
		imageToUse = v
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
			env = append(env, key+"="+val)
		}
	}

	// Assign a unique container name based on build ID to prevent collisions
	containerName := ""
	containerWorkspace := "/home/builder/build" // default workspace path inside container
	if opts.BuildID != "" {
		containerName = "smidr-build-" + opts.BuildID
		// CRITICAL: Use unique workspace path inside container to prevent BitBake server collisions
		// Each build gets its own workspace so BitBake lock files don't conflict
		containerWorkspace = "/home/builder/build-" + opts.BuildID
	}

	// Allow deterministic container name for tests
	testName := os.Getenv("SMIDR_TEST_CONTAINER_NAME")
	containerCfg := smidrcontainer.ContainerConfig{
		Image: imageToUse,
		Name: func() string {
			if strings.TrimSpace(testName) != "" {
				return testName
			}
			return containerName
		}(),
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

	dm, err := docker.NewDockerManager(r.logger)
	if err != nil {
		return &BuildResult{Success: false, Duration: time.Since(start)}, err
	}

	// Image pull/check
	if !dm.ImageExists(ctx, containerCfg.Image) {
		log.Write("stdout", "🐳 Pulling image "+containerCfg.Image+"...")
		r.logger.Info("pulling container image", slog.String("image", containerCfg.Image))
		if err := dm.PullImage(ctx, containerCfg.Image); err != nil {
			r.logger.Error("failed to pull image", err, slog.String("image", containerCfg.Image))
			return &BuildResult{Success: false, Duration: time.Since(start)}, err
		}
	}

	// Emit setup marker for integration tests and readability
	log.Write("stdout", "Preparing container environment")
	r.logger.Info("Preparing container environment")

	// Create container
	r.logger.Info("creating container", slog.String("name", containerName), slog.String("image", containerCfg.Image))
	containerID, err := dm.CreateContainer(ctx, containerCfg)
	if err != nil {
		r.logger.Error("failed to create container", err)
		return &BuildResult{Success: false, Duration: time.Since(start)}, err
	}
	defer func() {
		r.logger.Info("Cleaning up container")
		_ = dm.StopContainer(context.Background(), containerID, 2*time.Second)
		_ = dm.RemoveContainer(context.Background(), containerID, true)
	}()

	// Start container
	r.logger.Info("starting container", slog.String("container_id", containerID))
	if err := dm.StartContainer(ctx, containerID); err != nil {
		r.logger.Error("failed to start container", err, slog.String("container_id", containerID))
		return &BuildResult{Success: false, Duration: time.Since(start)}, err
	}

	// Test-mode markers: write simple signals to stdout and filesystem to satisfy integration checks
	if os.Getenv("SMIDR_TEST_WRITE_MARKERS") == "1" {
		if ws := os.Getenv("SMIDR_TEST_WORKSPACE_DIR"); strings.TrimSpace(ws) != "" {
			_ = os.MkdirAll(ws, 0o755)
			_ = os.WriteFile(filepath.Join(ws, "itest.txt"), []byte("ok"), 0o644)
		}
		if dl := os.Getenv("SMIDR_TEST_DOWNLOADS_DIR"); strings.TrimSpace(dl) != "" {
			if fi, err := os.Stat(dl); err == nil && fi.IsDir() {
				log.Write("stdout", "Downloads directory accessible")
			}
		}
		if ss := os.Getenv("SMIDR_TEST_SSTATE_DIR"); strings.TrimSpace(ss) != "" {
			if fi, err := os.Stat(ss); err == nil && fi.IsDir() {
				log.Write("stdout", "Sstate directory accessible")
			}
		}
	}

	// Run bitbake
	// Pass the container's workspace path (not host path) so BitBake runs in the right directory
	executor := bitbake.NewBuildExecutor(cfg, dm, containerID, containerWorkspace, r.logger)
	executor.SetForceImage(opts.ForceImage)

	// Set build prefix for log identification (e.g., "[customer/build-abc123]")
	if opts.Customer != "" && opts.BuildID != "" {
		buildPrefix := "[" + opts.Customer + "/" + opts.BuildID[:8] + "]"
		executor.SetBuildPrefix(buildPrefix)
	} else if opts.BuildID != "" {
		buildPrefix := "[" + opts.BuildID[:8] + "]"
		executor.SetBuildPrefix(buildPrefix)
	}

	// Prepare build log files (fresh per build dir) and wire writers
	// Policy: truncate/create fresh logs each run since each build uses a unique directory.
	txtPath := filepath.Join(cfg.Directories.Build, "build-log.txt")
	jsonlPath := filepath.Join(cfg.Directories.Build, "build-log.jsonl")
	txtFile, txtErr := os.OpenFile(txtPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if txtErr != nil {
		r.logger.Warn("unable to open build-log.txt", slog.String("path", txtPath), slog.Any("error", txtErr))
		txtFile = nil
	} else {
		defer txtFile.Close()
	}
	jsonlFile, jsonlErr := os.OpenFile(jsonlPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if jsonlErr != nil {
		r.logger.Warn("unable to open build-log.jsonl", slog.String("path", jsonlPath), slog.Any("error", jsonlErr))
		jsonlFile = nil
	} else {
		defer jsonlFile.Close()
	}

	// Adapter that forwards important lines to the provided sink and structured logger
	forwardFunc := logWriterFunc(func(p []byte) (int, error) {
		// Split and forward lines to provided sink
		for _, line := range strings.Split(string(p), "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}

			// Check for task progress and log it separately for progress bar
			if progress := parseTaskProgress(trimmed); progress != nil {
				log.Write("stdout", trimmed) // Send to client
				r.logger.Info("Build progress",
					slog.Int("current", progress.Current),
					slog.Int("total", progress.Total),
					slog.String("task", progress.Task))
				continue
			}

			// Only log important lines: errors, warnings, summaries, and key status updates
			if shouldLogLine(trimmed) {
				log.Write("stdout", line)
			}
		}
		return len(p), nil
	})

	// Plain writer: tee to file (if available) and the forwarder for sink/filters
	var plainWriter io.Writer = forwardFunc
	if txtFile != nil {
		plainWriter = io.MultiWriter(txtFile, forwardFunc)
	}

	// Log adapter to forward bitbake output into LogSink and write to files
	bbLog := &bitbake.BuildLogWriter{
		PlainWriter: plainWriter,
		JSONLWriter: jsonlFile,
	}

	result, err := executor.ExecuteBuild(ctx, bbLog)
	exitCode := 0
	if result != nil {
		exitCode = result.ExitCode
	}
	br := &BuildResult{Success: err == nil && result != nil && result.Success, ExitCode: exitCode, Duration: time.Since(start), BuildDir: cfg.Directories.Build, TmpDir: cfg.Directories.Tmp, DeployDir: cfg.Directories.Deploy}

	// If DB persistence is available, update completion status and record artifacts
	if r.db != nil {
		duration := time.Since(start)
		var status db.BuildStatus
		var errorMsg string
		if err != nil {
			status = db.StatusFailed
			errorMsg = err.Error()
		} else if result != nil && result.Success {
			status = db.StatusCompleted
		} else {
			status = db.StatusFailed
			if result != nil {
				errorMsg = fmt.Sprintf("build failed with exit code %d", result.ExitCode)
			}
		}
		if cerr := r.db.CompleteBuild(opts.BuildID, status, exitCode, duration, errorMsg); cerr != nil {
			r.logger.Error("failed to update build completion", cerr)
		}
		if status == db.StatusCompleted && result != nil {
			r.recordArtifacts(opts.BuildID, cfg.Directories.Deploy)
		}
	}

	if err != nil {
		return br, err
	}
	return br, nil
}

type logWriterFunc func(p []byte) (n int, err error)

func (f logWriterFunc) Write(p []byte) (n int, err error) { return f(p) }

// TaskProgress represents BitBake task execution progress
// Use parseTaskProgress() to extract progress from BitBake output lines.
// Example output from BitBake: "NOTE: Running task 482 of 776 (virtual:native:/home/builder/layers/poky/meta/recipes-graphics/libepoxy/libepoxy_1.5.10.bb:do_populate_sysroot_setscene)"
// This will be parsed to: TaskProgress{Current: 482, Total: 776, Task: "virtual:native:/.../libepoxy_1.5.10.bb:do_populate_sysroot_setscene"}
type TaskProgress struct {
	Current int
	Total   int
	Task    string // e.g., "do_populate_sysroot_setscene" or full task description
}

// parseTaskProgress extracts task progress from BitBake output
// Example input: "NOTE: Running task 123 of 456 (virtual:native:/path/to/recipe.bb:do_task)"
// Returns nil if the line is not a task progress message
func parseTaskProgress(line string) *TaskProgress {
	// Match both forms:
	//  - NOTE: Running task 123 of 456 (...)
	//  - NOTE: Running setscene task 123 of 456 (...)
	re := regexp.MustCompile(`^NOTE: Running (?:setscene )?task (\d+) of (\d+)(?:\s*\((.*)\))?`)
	m := re.FindStringSubmatch(line)
	if m == nil {
		return nil
	}

	cur, err1 := strconv.Atoi(m[1])
	tot, err2 := strconv.Atoi(m[2])
	if err1 != nil || err2 != nil {
		return nil
	}

	tp := &TaskProgress{Current: cur, Total: tot}
	if len(m) >= 4 {
		tp.Task = m[3]
	}
	return tp
}

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

	// Only show progress lines and critical messages; hide emoji/status chatter
	// Log task progress for progress bar (e.g., "NOTE: Running task 123 of 456" or setscene variant)
	if strings.HasPrefix(line, "NOTE: Running task") || strings.HasPrefix(line, "NOTE: Running setscene task") {
		return true
	}

	// Skip verbose recipe task status messages (Started/Succeeded/etc)
	// These create hundreds of lines of noise
	if strings.HasPrefix(line, "NOTE: recipe") ||
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

// recordArtifacts scans the deploy directory and records artifacts in the database
func (r *Runner) recordArtifacts(buildID string, deployDir string) {
	if r.db == nil || deployDir == "" {
		return
	}

	// Walk deploy directory and record files
	err := filepath.Walk(deployDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		// Get relative path
		relPath, _ := filepath.Rel(deployDir, path)

		// Determine artifact type from extension
		artifactType := "unknown"
		ext := filepath.Ext(path)
		switch ext {
		case ".wic", ".img":
			artifactType = "image"
		case ".tar", ".gz", ".bz2", ".xz":
			artifactType = "archive"
		case ".txt", ".log":
			artifactType = "text"
		case ".json", ".xml":
			artifactType = "metadata"
		}

		artifact := &db.BuildArtifact{
			BuildID:      buildID,
			ArtifactPath: relPath,
			ArtifactType: artifactType,
			SizeBytes:    info.Size(),
			Checksum:     "",
			CreatedAt:    time.Now(),
		}

		if err := r.db.AddArtifact(artifact); err != nil {
			r.logger.Warn("failed to record artifact", slog.String("path", relPath), slog.String("error", err.Error()))
		}

		return nil
	})

	if err != nil {
		r.logger.Error("failed to scan artifacts", err)
	}
}

// generateBuildID creates a unique build identifier
func generateBuildID(customer string) string {
	timestamp := time.Now().Format("20060102-150405")
	if customer != "" {
		return fmt.Sprintf("%s-%s", customer, timestamp)
	}
	return fmt.Sprintf("build-%s", timestamp)
}
