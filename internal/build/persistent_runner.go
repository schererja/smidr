package build

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/schererja/smidr/internal/config"
	"github.com/schererja/smidr/internal/db"
	"github.com/schererja/smidr/pkg/logger"
)

// PersistentRunner wraps Runner with database persistence
type PersistentRunner struct {
	runner *Runner
	db     *db.DB
	logger *logger.Logger
}

// NewPersistentRunner creates a new database-backed runner
func NewPersistentRunner(database *db.DB, logger *logger.Logger) *PersistentRunner {
	return &PersistentRunner{
		runner: NewRunner(logger),
		db:     database,
		logger: logger,
	}
}

// Run executes a build with full database persistence
func (pr *PersistentRunner) Run(ctx context.Context, cfg *config.Config, opts BuildOptions, log LogSink) (*BuildResult, error) {
	start := time.Now()

	// Generate build ID if not provided
	if opts.BuildID == "" {
		opts.BuildID = generateBuildID(opts.Customer)
	}

	// Serialize config for snapshot
	configSnapshot, err := json.Marshal(cfg)
	if err != nil {
		pr.logger.Warn("failed to serialize config snapshot", slog.String("error", err.Error()))
		configSnapshot = []byte("{}")
	}

	// Get hostname and user
	hostname, _ := os.Hostname()
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME") // Windows
	}
	if username == "" {
		username = "unknown"
	}

	// Determine paths (expand similar to runner)
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

	// Create initial build record (QUEUED state)
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
		ConfigFile:     "", // Could be set if we know the config file path
		ConfigSnapshot: string(configSnapshot),
		User:           username,
		Host:           hostname,
		CreatedAt:      start,
	}

	if err := pr.db.CreateBuild(build); err != nil {
		pr.logger.Error("failed to create build record", err)
		// Continue anyway - persistence failure shouldn't block build
	}

	// Mark build as RUNNING when execution starts
	defer func() {
		// Ensure build state is updated even if panic occurs
		if r := recover(); r != nil {
			pr.db.CompleteBuild(opts.BuildID, db.StatusFailed, 1, time.Since(start), fmt.Sprintf("panic: %v", r))
			panic(r) // re-panic after recording
		}
	}()

	if err := pr.db.StartBuild(opts.BuildID); err != nil {
		pr.logger.Error("failed to mark build as running", err)
	}

	// Execute the actual build
	result, buildErr := pr.runner.Run(ctx, cfg, opts, log)

	// Update build completion status
	duration := time.Since(start)
	var status db.BuildStatus
	var errorMsg string

	if buildErr != nil {
		status = db.StatusFailed
		errorMsg = buildErr.Error()
	} else if result != nil && result.Success {
		status = db.StatusCompleted
	} else {
		status = db.StatusFailed
		if result != nil {
			errorMsg = fmt.Sprintf("build failed with exit code %d", result.ExitCode)
		}
	}

	exitCode := 0
	if result != nil {
		exitCode = result.ExitCode
	}

	if err := pr.db.CompleteBuild(opts.BuildID, status, exitCode, duration, errorMsg); err != nil {
		pr.logger.Error("failed to update build completion", err)
	}

	// TODO: Scan and record artifacts if build succeeded
	if status == db.StatusCompleted && result != nil {
		pr.recordArtifacts(opts.BuildID, result.DeployDir)
	}

	return result, buildErr
}

// recordArtifacts scans the deploy directory and records artifacts in the database
func (pr *PersistentRunner) recordArtifacts(buildID string, deployDir string) {
	if deployDir == "" {
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
			Checksum:     "", // TODO: compute SHA256 if needed
			CreatedAt:    time.Now(),
		}

		if err := pr.db.AddArtifact(artifact); err != nil {
			pr.logger.Warn("failed to record artifact", slog.String("path", relPath), slog.String("error", err.Error()))
		}

		return nil
	})

	if err != nil {
		pr.logger.Error("failed to scan artifacts", err)
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
