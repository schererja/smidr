package daemon

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"github.com/schererja/smidr/internal/artifacts"
	buildpkg "github.com/schererja/smidr/internal/build"
	"github.com/schererja/smidr/internal/config"
	"github.com/schererja/smidr/internal/db"
	"github.com/schererja/smidr/pkg/logger"
	v1 "github.com/schererja/smidr/pkg/smidr-sdk/v1"
)

// Server implements the Smidr gRPC service
type Server struct {
	v1.UnimplementedArtifactServiceServer
	v1.UnimplementedBuildServiceServer
	v1.UnimplementedLogServiceServer

	address        string
	grpcServer     *grpc.Server
	builds         map[string]*BuildInfo
	buildsMutex    sync.RWMutex
	artifactMgr    *artifacts.ArtifactManager
	buildSemaphore chan struct{}            // global semaphore to limit concurrent builds based on available resources
	customerQueues map[string]chan struct{} // per-customer queues for serializing builds within the same customer
	queuesMutex    sync.RWMutex             // protects customerQueues map
	logger         *logger.Logger           // structured logger
	database       *db.DB                   // optional database for build persistence
}

// BuildInfo holds information about an active or completed build
type BuildInfo struct {
	ID             string
	Target         string
	State          v1.BuildState
	ExitCode       int32
	ErrorMsg       string
	StartedAt      time.Time
	CompletedAt    time.Time
	ConfigPath     string
	Config         *config.Config
	LogBuffer      []*v1.LogEntry
	LogMutex       sync.RWMutex
	LogSubscribers map[chan *v1.LogEntry]bool
	cancel         context.CancelFunc
	ArtifactPaths  []string
}

// LogWriter implements bitbake.BuildLogWriter for streaming logs
type LogWriter struct {
	buildInfo   *BuildInfo
	buildLogger *logger.Logger // structured logger with build context
}

func (lw *LogWriter) WriteLog(stream, content string) {
	logLine := &v1.LogEntry{
		TimestampUnixSeconds: time.Now().UnixNano(),
		Stream:               stream,
		Message:              content,
	}

	lw.buildInfo.LogMutex.Lock()
	defer lw.buildInfo.LogMutex.Unlock()

	// Add to buffer
	lw.buildInfo.LogBuffer = append(lw.buildInfo.LogBuffer, logLine)

	// Send to all subscribers
	for ch := range lw.buildInfo.LogSubscribers {
		select {
		case ch <- logLine:
		default:
			// Don't block if subscriber is slow
		}
	}

	// Also log to structured logger if available
	if lw.buildLogger != nil {
		if stream == "stderr" {
			lw.buildLogger.Warn(content, slog.String("stream", stream))
		} else {
			lw.buildLogger.Info(content, slog.String("stream", stream))
		}
	}
}

// NewServer creates a new daemon server
func NewServer(address string, log *logger.Logger, database *db.DB) *Server {
	// Initialize artifact manager with default location
	artifactMgr, err := artifacts.NewArtifactManager("")
	if err != nil {
		// Fall back to a temporary directory if we can't create default location
		log.Warn("Failed to initialize artifact manager", slog.String("error", err.Error()))
		artifactMgr = nil
	}

	return &Server{
		address:        address,
		builds:         make(map[string]*BuildInfo),
		artifactMgr:    artifactMgr,
		customerQueues: make(map[string]chan struct{}),
		logger:         log,
		database:       database,
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	// Attempt recovery of stale builds if DB is available
	if s.database != nil {
		if err := s.recoverStaleBuilds(); err != nil {
			s.logger.Warn("Stale build recovery encountered errors", slog.String("error", err.Error()))
		}
	}

	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.address, err)
	}

	s.grpcServer = grpc.NewServer()
	v1.RegisterArtifactServiceServer(s.grpcServer, s)
	v1.RegisterBuildServiceServer(s.grpcServer, s)
	v1.RegisterLogServiceServer(s.grpcServer, s)
	s.logger.Info("Smidr daemon listening", slog.String("address", s.address))

	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// getOrCreateCustomerQueue returns a semaphore channel for the given customer.
// Each customer gets their own queue to allow parallel builds across customers
// while serializing builds within the same customer to prevent BitBake deadlocks.
func (s *Server) getOrCreateCustomerQueue(customer string) chan struct{} {
	s.queuesMutex.Lock()
	defer s.queuesMutex.Unlock()

	if queue, exists := s.customerQueues[customer]; exists {
		return queue
	}

	// Create new queue with capacity 1 (one build at a time per customer)
	queue := make(chan struct{}, 1)
	queue <- struct{}{} // initialize with one token
	s.customerQueues[customer] = queue
	return queue
}

// generateShortID creates a short unique identifier (first 8 chars of UUID)
func generateShortID() string {
	return uuid.New().String()[:8]
}

// recoverStaleBuilds marks any previously running builds as failed so users can rebuild
func (s *Server) recoverStaleBuilds() error {
	builds, err := s.database.ListStaleBuilds()
	if err != nil {
		return fmt.Errorf("failed to list stale builds: %w", err)
	}
	if len(builds) == 0 {
		s.logger.Info("No stale builds detected")
		return nil
	}

	s.logger.Warn("Recovering stale builds", slog.Int("count", len(builds)))
	for _, b := range builds {
		// Compute a best-effort duration
		var dur time.Duration
		if b.StartedAt != nil {
			dur = time.Since(*b.StartedAt)
		} else {
			dur = time.Since(b.CreatedAt)
		}
		msg := "daemon restarted: marking stale build as failed (re-run to rebuild)"
		if err := s.database.CompleteBuild(b.ID, db.StatusFailed, 1, dur, msg); err != nil {
			s.logger.Warn("Failed to mark stale build as failed", slog.String("buildID", b.ID), slog.String("error", err.Error()))
			continue
		}
		s.logger.Info("Marked stale build as failed", slog.String("buildID", b.ID), slog.String("customer", b.Customer), slog.String("target", b.TargetImage))
	}
	return nil
}

// Stop gracefully stops the gRPC server and cancels all running builds
func (s *Server) Stop() {
	s.logger.Info("Stopping daemon...")

	// Cancel all running builds
	s.buildsMutex.Lock()
	s.logger.Info("Cancelling active builds", slog.Int("count", len(s.builds)))
	for buildID, build := range s.builds {
		if build.State == v1.BuildState_BUILD_STATE_BUILDING ||
			build.State == v1.BuildState_BUILD_STATE_PREPARING ||
			build.State == v1.BuildState_BUILD_STATE_QUEUED {
			s.logger.Info("Cancelling build",
				slog.String("buildID", buildID),
				slog.String("state", build.State.String()))
			if build.cancel != nil {
				build.cancel()
			}
		}
	}
	s.buildsMutex.Unlock()

	// Give builds a moment to clean up
	time.Sleep(2 * time.Second)

	// Stop the gRPC server
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}

	s.logger.Info("Daemon stopped")
}

// StartBuild handles build start requests
func (s *Server) StartBuild(ctx context.Context, req *v1.StartBuildRequest) (*v1.BuildStatusResponse, error) {
	if req.Config == "" {
		return nil, fmt.Errorf("config is required (path or inline YAML/JSON content)")
	}

	// Load configuration: treat req.Config as path if it exists; otherwise as inline content
	var (
		cfg             *config.Config
		err             error
		configPathLabel string
	)
	if st, statErr := os.Stat(req.Config); statErr == nil && !st.IsDir() {
		cfg, err = config.Load(req.Config)
		configPathLabel = req.Config
	} else {
		cfg, err = config.LoadFromBytes([]byte(req.Config))
		configPathLabel = "<inline>"
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// If target wasn't provided, default to config's build.image for reproducibility and DB persistence
	if req.Target == "" && cfg != nil && cfg.Build.Image != "" {
		req.Target = cfg.Build.Image
	}

	// Generate a unique build ID using customer name (or config name fallback) and short UUID suffix
	buildIDPrefix := req.Customer
	if buildIDPrefix == "" {
		buildIDPrefix = cfg.Name
	}
	buildID := fmt.Sprintf("%s-%s", buildIDPrefix, generateShortID())

	buildCtx, cancel := context.WithCancel(context.Background())

	s.buildsMutex.Lock()
	buildInfo := &BuildInfo{
		ID:             buildID,
		Target:         req.Target,
		State:          v1.BuildState_BUILD_STATE_QUEUED,
		StartedAt:      time.Now(),
		ConfigPath:     configPathLabel,
		Config:         cfg,
		LogBuffer:      make([]*v1.LogEntry, 0),
		LogSubscribers: make(map[chan *v1.LogEntry]bool),
		cancel:         cancel,
	}
	s.builds[buildID] = buildInfo
	s.buildsMutex.Unlock()

	// Start the build in a goroutine
	go s.executeBuild(buildCtx, buildInfo, req)
	buildIdentifier := v1.BuildIdentifier{BuildId: buildID}
	return &v1.BuildStatusResponse{
		BuildIdentifier: &buildIdentifier,
		State:           buildInfo.State,
		Customer:        req.Customer,
		Target:          req.Target,
		Timestamps: &v1.TimeStampRange{
			StartTimeUnixSeconds: buildInfo.StartedAt.Unix(),
		},
		ConfigPath: buildInfo.ConfigPath,
	}, nil
}

// executeBuild runs the actual build process
func (s *Server) executeBuild(ctx context.Context, buildInfo *BuildInfo, req *v1.StartBuildRequest) {
	// Determine the customer key for queuing (use customer name or config name as fallback)
	customerKey := req.Customer
	if customerKey == "" {
		customerKey = buildInfo.Config.Name
	}
	if customerKey == "" {
		customerKey = "default"
	}

	// Create a build-specific logger with context
	buildLogger := s.logger.With(
		slog.String("buildID", buildInfo.ID),
		slog.String("customer", customerKey),
		slog.String("target", req.Target),
	)

	// Create log writer with structured logger
	logWriter := &LogWriter{
		buildInfo:   buildInfo,
		buildLogger: buildLogger,
	}

	// Get or create a queue for this customer
	queue := s.getOrCreateCustomerQueue(customerKey)

	// Acquire the customer's build queue token to serialize builds per customer
	logWriter.WriteLog("stdout", fmt.Sprintf("⏳ Waiting for build slot for customer '%s'...", customerKey))
	select {
	case <-queue:
		// Got the token, proceed with build
		defer func() {
			queue <- struct{}{} // return token when done
		}()
		logWriter.WriteLog("stdout", "Build slot acquired, proceeding...")
	case <-ctx.Done():
		s.failBuild(buildInfo.ID, "Build cancelled while waiting for queue slot")
		return
	}

	// Update state to preparing
	s.updateBuildState(buildInfo.ID, v1.BuildState_BUILD_STATE_PREPARING)

	logWriter.WriteLog("stdout", "Starting build process...")
	logWriter.WriteLog("stdout", fmt.Sprintf("Config: %s", buildInfo.ConfigPath))
	logWriter.WriteLog("stdout", fmt.Sprintf("Target: %s", req.Target))

	// Build options for runner
	opts := buildpkg.BuildOptions{
		BuildID:    buildInfo.ID,
		Target:     req.Target,
		Customer:   req.Customer,
		ForceClean: req.ForceClean,
		ForceImage: req.ForceImageRebuild,
		ConfigPath: buildInfo.ConfigPath,
	}

	// Bridge for runner logs -> gRPC stream subscribers
	sink := &runnerLogSink{lw: logWriter}

	// Mark as BUILDING
	s.updateBuildState(buildInfo.ID, v1.BuildState_BUILD_STATE_BUILDING)

	// Use runner to execute build with DB persistence if available
	runner := buildpkg.NewRunner(logWriter.buildLogger, s.database)
	result, err := runner.Run(ctx, buildInfo.Config, opts, sink)
	if err != nil {
		// Failed
		buildInfo.ExitCode = 1
		buildInfo.ErrorMsg = err.Error()
		buildInfo.CompletedAt = time.Now()
		s.updateBuildState(buildInfo.ID, v1.BuildState_BUILD_STATE_FAILED)
		logWriter.WriteLog("stderr", fmt.Sprintf("Build failed: %v", err))
		return
	}

	// Success or non-zero exit
	buildInfo.ExitCode = int32(result.ExitCode)
	buildInfo.CompletedAt = time.Now()
	if result.Success {
		// Extract artifacts if artifact manager is available and build succeeded
		if s.artifactMgr != nil {
			s.updateBuildState(buildInfo.ID, v1.BuildState_BUILD_STATE_EXTRACTING_ARTIFACTS)
			logWriter.WriteLog("stdout", "Extracting artifacts...")

			if err := s.extractArtifacts(ctx, buildInfo, result, logWriter); err != nil {
				logWriter.WriteLog("stderr", fmt.Sprintf("Failed to extract artifacts: %v", err))
				// Don't fail the build just because artifact extraction failed
			} else {
				logWriter.WriteLog("stdout", "Artifacts extracted successfully")
			}
		}

		s.updateBuildState(buildInfo.ID, v1.BuildState_BUILD_STATE_COMPLETED)
		logWriter.WriteLog("stdout", fmt.Sprintf("Build completed in %v", result.Duration))
	} else {
		s.updateBuildState(buildInfo.ID, v1.BuildState_BUILD_STATE_FAILED)
		logWriter.WriteLog("stderr", fmt.Sprintf("Build finished with errors in %v (exit=%d)", result.Duration, result.ExitCode))
	}
}

// runnerLogSink adapts daemon LogWriter to the Runner LogSink interface
type runnerLogSink struct{ lw *LogWriter }

func (s *runnerLogSink) Write(stream string, line string) {
	if s == nil || s.lw == nil {
		return
	}
	s.lw.WriteLog(stream, line)
}

// extractArtifacts extracts build artifacts from the build result
func (s *Server) extractArtifacts(ctx context.Context, buildInfo *BuildInfo, result *buildpkg.BuildResult, logWriter *LogWriter) error {
	// Get current user for metadata
	user := "unknown"
	if u := os.Getenv("USER"); u != "" {
		user = u
	}

	// Create build metadata for artifact manager
	metadata := artifacts.BuildMetadata{
		BuildID:       buildInfo.ID,
		ProjectName:   buildInfo.Target, // Use target as project name
		User:          user,
		Timestamp:     buildInfo.StartedAt,
		ConfigUsed:    map[string]string{"target": buildInfo.Target},
		BuildDuration: time.Since(buildInfo.StartedAt),
		TargetImage:   buildInfo.Target,
		ArtifactSizes: make(map[string]int64),
		Status:        "success",
	}

	// We need to create a temporary container to extract artifacts from the deploy directory
	// Since the original build container is cleaned up, we'll copy artifacts directly from host filesystem

	// For now, let's extract artifacts directly from the build result's deploy directory
	if result.DeployDir == "" {
		return fmt.Errorf("no deploy directory specified in build result")
	}

	// Check if deploy directory exists and has contents
	deployPath := result.DeployDir
	if _, err := os.Stat(deployPath); os.IsNotExist(err) {
		return fmt.Errorf("deploy directory does not exist: %s", deployPath)
	}

	// Create artifact storage directory
	artifactPath := s.artifactMgr.GetArtifactPath(buildInfo.ID)
	if err := os.MkdirAll(artifactPath, 0755); err != nil {
		return fmt.Errorf("failed to create artifact directory: %w", err)
	}

	// Copy deploy directory contents to artifact storage
	logWriter.WriteLog("stdout", fmt.Sprintf("Copying artifacts from %s to %s", deployPath, artifactPath))

	if err := s.copyDirectory(deployPath, filepath.Join(artifactPath, "deploy"), &metadata); err != nil {
		return fmt.Errorf("failed to copy deploy directory: %w", err)
	}

	// Save metadata
	if err := s.artifactMgr.SaveMetadata(metadata); err != nil {
		return fmt.Errorf("failed to save artifact metadata: %w", err)
	}

	// Store artifact paths in BuildInfo for later retrieval
	artifacts, err := s.artifactMgr.ListArtifacts(buildInfo.ID)
	if err != nil {
		logWriter.WriteLog("stderr", fmt.Sprintf("Failed to list artifacts: %v", err))
	} else {
		buildInfo.ArtifactPaths = artifacts
	}

	return nil
}

// copyDirectory recursively copies a directory and calculates file sizes
func (s *Server) copyDirectory(src, dst string, metadata *artifacts.BuildMetadata) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path from source
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		// Preserve symlinks to avoid dereferencing directory links (e.g., deploy/licenses/*-<machine>)
		if info.Mode()&os.ModeSymlink != 0 {
			// Read the symlink target and recreate the symlink at destination
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			// Ensure destination parent exists
			if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
				return err
			}
			// Best-effort remove an existing file/dir at dstPath before creating the symlink
			_ = os.RemoveAll(dstPath)
			if err := os.Symlink(target, dstPath); err != nil {
				return err
			}
			// Record size as 0 for symlinks in metadata (content tracked at target path if also copied)
			if metadata != nil {
				artifactRelPath := filepath.Join("deploy", relPath)
				metadata.ArtifactSizes[artifactRelPath] = 0
			}
			return nil
		}

		if info.IsDir() {
			// Create directory
			return os.MkdirAll(dstPath, info.Mode())
		} else {
			// Copy file
			if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
				return err
			}

			srcFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			dstFile, err := os.Create(dstPath)
			if err != nil {
				return err
			}
			defer dstFile.Close()

			_, err = io.Copy(dstFile, srcFile)
			if err != nil {
				return err
			}

			// Calculate checksum and store size
			srcFile.Seek(0, 0)
			hash := sha256.New()
			io.Copy(hash, srcFile)

			// Store in metadata
			artifactRelPath := filepath.Join("deploy", relPath)
			metadata.ArtifactSizes[artifactRelPath] = info.Size()
		}

		return nil
	})
}

// updateBuildState updates the state of a build
func (s *Server) updateBuildState(buildID string, state v1.BuildState) {
	s.buildsMutex.Lock()
	defer s.buildsMutex.Unlock()

	if build, exists := s.builds[buildID]; exists {
		build.State = state
	}
}

// failBuild marks a build as failed
func (s *Server) failBuild(buildID string, errorMsg string) {
	s.buildsMutex.Lock()
	defer s.buildsMutex.Unlock()

	if build, exists := s.builds[buildID]; exists {
		build.State = v1.BuildState_BUILD_STATE_FAILED
		build.ErrorMsg = errorMsg
		build.CompletedAt = time.Now()
		build.ExitCode = 1

		logWriter := &LogWriter{buildInfo: build}
		logWriter.WriteLog("stderr", fmt.Sprintf("Build failed: %s", errorMsg))
	}
}

// GetBuildStatus retrieves the status of a build
func (s *Server) GetBuildStatus(ctx context.Context, req *v1.BuildStatusRequest) (*v1.BuildStatusResponse, error) {
	s.buildsMutex.RLock()
	build, exists := s.builds[req.BuildIdentifier.BuildId]
	s.buildsMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("build %s not found", req.BuildIdentifier.BuildId)
	}

	status := &v1.BuildStatusResponse{
		BuildIdentifier: &v1.BuildIdentifier{BuildId: build.ID},
		Target:          build.Target,
		State:           build.State,
		ExitCode:        build.ExitCode,
		Timestamps: &v1.TimeStampRange{
			StartTimeUnixSeconds: build.StartedAt.Unix(),
		},
		ConfigPath: build.ConfigPath,
	}

	if build.ErrorMsg != "" {
		status.ErrorMessage = build.ErrorMsg
	}

	if !build.CompletedAt.IsZero() {
		status.Timestamps.EndTimeUnixSeconds = build.CompletedAt.Unix()
	}

	return status, nil
}

// StreamLogs streams build logs to the client
func (s *Server) StreamLogs(req *v1.StreamBuildLogsRequest, stream v1.LogService_StreamBuildLogsServer) error {
	s.buildsMutex.RLock()
	build, exists := s.builds[req.BuildIdentifier.BuildId]
	s.buildsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("build %s not found", req.BuildIdentifier.BuildId)
	}

	// Send existing logs
	build.LogMutex.RLock()
	for _, logLine := range build.LogBuffer {
		if err := stream.Send(logLine); err != nil {
			build.LogMutex.RUnlock()
			return err
		}
	}
	build.LogMutex.RUnlock()

	// If follow=true, continue streaming new logs
	if req.Follow {
		// Create a channel for this subscriber
		logChan := make(chan *v1.LogEntry, 100)

		build.LogMutex.Lock()
		build.LogSubscribers[logChan] = true
		build.LogMutex.Unlock()

		// Clean up subscriber when done
		defer func() {
			build.LogMutex.Lock()
			delete(build.LogSubscribers, logChan)
			build.LogMutex.Unlock()
			close(logChan)
		}()

		// Stream logs until build completes or client disconnects
		for {
			select {
			case logLine := <-logChan:
				if err := stream.Send(logLine); err != nil {
					return err
				}
			case <-stream.Context().Done():
				return nil
			}

			// Check if build is complete
			build.LogMutex.RLock()
			isComplete := build.State == v1.BuildState_BUILD_STATE_COMPLETED ||
				build.State == v1.BuildState_BUILD_STATE_FAILED ||
				build.State == v1.BuildState_BUILD_STATE_CANCELLED
			build.LogMutex.RUnlock()

			if isComplete {
				// Drain any remaining logs
				for {
					select {
					case logLine := <-logChan:
						if err := stream.Send(logLine); err != nil {
							return err
						}
					default:
						return nil
					}
				}
			}
		}
	}

	return nil
}

// ListArtifacts lists all artifacts from a completed build
func (s *Server) ListArtifacts(ctx context.Context, req *v1.ListArtifactsRequest) (*v1.ListArtifactsResponse, error) {
	s.buildsMutex.RLock()
	build, exists := s.builds[req.BuildIdentifier.BuildId]
	s.buildsMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("build %s not found", req.BuildIdentifier.BuildId)
	}

	if build.State != v1.BuildState_BUILD_STATE_COMPLETED {
		return nil, fmt.Errorf("build %s is not completed", req.BuildIdentifier.BuildId)
	}

	// Return empty list if artifact manager is not available
	if s.artifactMgr == nil {
		return &v1.ListArtifactsResponse{
			BuildIdentifier: req.BuildIdentifier,
			Artifacts:       []*v1.ArtifactSummary{},
		}, nil
	}

	// Get artifacts from artifact manager
	artifactFiles, err := s.artifactMgr.ListArtifacts(req.BuildIdentifier.BuildId)
	if err != nil {
		return nil, fmt.Errorf("failed to list artifacts: %w", err)
	}

	// Load metadata to get file sizes
	metadata, err := s.artifactMgr.LoadMetadata(req.BuildIdentifier.BuildId)
	if err != nil {
		// If we can't load metadata, just return files without sizes
		metadata = &artifacts.BuildMetadata{ArtifactSizes: make(map[string]int64)}
	}

	// Convert to protobuf artifacts
	var protoArtifacts []*v1.ArtifactSummary
	for _, artifactFile := range artifactFiles {
		artifact := &v1.ArtifactSummary{
			Name:        filepath.Base(artifactFile),
			DownloadUrl: artifactFile,
			SizeBytes:   metadata.ArtifactSizes[artifactFile],
		}

		// Calculate checksum if file exists
		fullPath := filepath.Join(s.artifactMgr.GetArtifactPath(req.BuildIdentifier.BuildId), artifactFile)
		if checksum, err := s.calculateChecksum(fullPath); err == nil {
			artifact.Checksum = checksum
		}

		protoArtifacts = append(protoArtifacts, artifact)
	}

	return &v1.ListArtifactsResponse{
		BuildIdentifier: req.BuildIdentifier,
		Artifacts:       protoArtifacts,
	}, nil
}

// calculateChecksum calculates SHA256 checksum of a file
func (s *Server) calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// CancelBuild cancels a running build
func (s *Server) CancelBuild(ctx context.Context, req *v1.CancelBuildRequest) (*v1.CancelBuildResponse, error) {
	s.buildsMutex.Lock()
	defer s.buildsMutex.Unlock()

	build, exists := s.builds[req.BuildIdentifier.BuildId]
	if !exists {
		return &v1.CancelBuildResponse{
			Success: false,
			Message: fmt.Sprintf("build %s not found", req.BuildIdentifier.BuildId),
		}, nil
	}

	if build.State != v1.BuildState_BUILD_STATE_BUILDING {
		return &v1.CancelBuildResponse{
			Success: false,
			Message: fmt.Sprintf("build %s is not in a cancellable state", req.BuildIdentifier.BuildId),
		}, nil
	}

	// TODO: Actually cancel the build
	if build.cancel != nil {
		build.cancel()
	}

	build.State = v1.BuildState_BUILD_STATE_CANCELLED
	build.CompletedAt = time.Now()

	return &v1.CancelBuildResponse{
		Success: true,
		Message: "Build cancelled successfully",
	}, nil
}

// ListBuilds lists all builds
func (s *Server) ListBuilds(ctx context.Context, req *v1.ListBuildsRequest) (*v1.ListBuildsResponse, error) {
	// If a database is available, return persisted builds (includes historical and running)
	if s.database != nil {
		// Fetch without limit to allow filtering by state, then apply limit
		dbBuilds, err := s.database.ListBuilds("", false, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to list builds from database: %w", err)
		}

		// Helper to map db.BuildStatus -> proto BuildState
		toProtoState := func(st db.BuildStatus) v1.BuildState {
			switch st {
			case db.StatusQueued:
				return v1.BuildState_BUILD_STATE_QUEUED
			case db.StatusRunning:
				return v1.BuildState_BUILD_STATE_BUILDING
			case db.StatusCompleted:
				return v1.BuildState_BUILD_STATE_COMPLETED
			case db.StatusFailed:
				return v1.BuildState_BUILD_STATE_FAILED
			case db.StatusCancelled:
				return v1.BuildState_BUILD_STATE_CANCELLED
			default:
				return v1.BuildState_BUILD_STATE_UNSPECIFIED
			}
		}

		// Optional state filter set from request
		filterByState := func(state v1.BuildState) bool {
			if len(req.StateFilter) == 0 {
				return true
			}
			for _, s := range req.StateFilter {
				if s == state {
					return true
				}
			}
			return false
		}

		builds := make([]*v1.BuildDetails, 0, len(dbBuilds))
		for _, b := range dbBuilds {
			protoState := toProtoState(b.Status)
			if !filterByState(protoState) {
				continue
			}

			bd := &v1.BuildDetails{
<<<<<<< HEAD:internal/daemon/server.go
				Id:           b.ID,
				TargetImage:  b.TargetImage,
				Status:       protoState,
				ConfigFile:   b.ConfigFile,
				Customer:     b.Customer,
				ProjectName:  b.ProjectName,
				Machine:      b.Machine,
				BuildDir:     b.BuildDir,
				DeployDir:    b.DeployDir,
				User:         b.User,
				Host:         b.Host,
				ErrorMessage: b.ErrorMessage,
				Deleted:      b.Deleted,
=======
				BuildIdentifier:   &v1.BuildIdentifier{BuildId: b.ID},
				TargetImage:       b.TargetImage,
				BuildState:        protoState,
				ConfigFile:        b.ConfigFile,
				Customer:          b.Customer,
				ProjectName:       b.ProjectName,
				Machine:           b.Machine,
				BuildDirectory:    b.BuildDir,
				DownloadDirectory: b.DeployDir,
				User:              b.User,
				Host:              b.Host,
				ErrorMessage:      b.ErrorMessage,
				Deleted:           b.Deleted,
>>>>>>> e73b5cf168c6534cab24d9771eef5581b620991b:apps/daemon/internal/daemon/server.go
			}
			if b.ExitCode != nil {
				bd.ExitCode = int32(*b.ExitCode)
			}
			if !b.CreatedAt.IsZero() {
				bd.CreatedAt = b.CreatedAt.Unix()
			}
			if b.StartedAt != nil {
<<<<<<< HEAD:internal/daemon/server.go
				bd.StartedAt = b.StartedAt.Unix()
			}
			if b.CompletedAt != nil {
				bd.CompletedAt = b.CompletedAt.Unix()
=======
				bd.Timestamps.StartTimeUnixSeconds = b.StartedAt.Unix()
			}
			if b.CompletedAt != nil {
				bd.Timestamps.EndTimeUnixSeconds = b.CompletedAt.Unix()
>>>>>>> e73b5cf168c6534cab24d9771eef5581b620991b:apps/daemon/internal/daemon/server.go
				if b.StartedAt != nil {
					bd.DurationSeconds = int32(b.CompletedAt.Sub(*b.StartedAt).Seconds())
				}
			}
			if b.DeletedAt != nil {
				bd.DeletedAt = b.DeletedAt.Unix()
			}

			builds = append(builds, bd)

			// Apply limit if requested
			if req.PageSize > 0 && len(builds) >= int(req.PageSize) {
				break
			}
		}

		return &v1.ListBuildsResponse{Builds: builds}, nil
	}

	// Fallback to in-memory builds when DB is not enabled
	s.buildsMutex.RLock()
	defer s.buildsMutex.RUnlock()

	builds := make([]*v1.BuildDetails, 0)
	for _, build := range s.builds {
		// Filter by state if requested
		if len(req.StateFilter) > 0 {
			matched := false
			for _, state := range req.StateFilter {
				if build.State == state {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		details := &v1.BuildDetails{
			BuildIdentifier: &v1.BuildIdentifier{BuildId: build.ID},
			TargetImage:     build.Target,
			BuildState:      build.State,
			ExitCode:        build.ExitCode,
			ConfigFile:      build.ConfigPath,
			Timestamps:      &v1.TimeStampRange{},
		}

		if !build.StartedAt.IsZero() {
			details.Timestamps.StartTimeUnixSeconds = build.StartedAt.Unix()
		}

		if build.ErrorMsg != "" {
			details.ErrorMessage = build.ErrorMsg
		}

		if !build.CompletedAt.IsZero() {
			details.Timestamps.EndTimeUnixSeconds = build.CompletedAt.Unix()
			if !build.StartedAt.IsZero() {
				details.DurationSeconds = int32(build.CompletedAt.Sub(build.StartedAt).Seconds())
			}
		}

		builds = append(builds, details)

		// Limit results if requested
		if req.PageSize > 0 && len(builds) >= int(req.PageSize) {
			break
		}
	}

	return &v1.ListBuildsResponse{Builds: builds}, nil
}
