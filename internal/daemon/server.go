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

	v1 "github.com/schererja/smidr/api/proto"
	"github.com/schererja/smidr/internal/artifacts"
	buildpkg "github.com/schererja/smidr/internal/build"
	"github.com/schererja/smidr/internal/config"
	"github.com/schererja/smidr/pkg/logger"
)

// Server implements the Smidr gRPC service
type Server struct {
	v1.UnimplementedSmidrServer

	address        string
	grpcServer     *grpc.Server
	builds         map[string]*BuildInfo
	buildsMutex    sync.RWMutex
	artifactMgr    *artifacts.ArtifactManager
	buildSemaphore chan struct{}            // global semaphore to limit concurrent builds based on available resources
	customerQueues map[string]chan struct{} // per-customer queues for serializing builds within the same customer
	queuesMutex    sync.RWMutex             // protects customerQueues map
	logger         *logger.Logger           // structured logger
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
	LogBuffer      []*v1.LogLine
	LogMutex       sync.RWMutex
	LogSubscribers map[chan *v1.LogLine]bool
	cancel         context.CancelFunc
	ArtifactPaths  []string
}

// LogWriter implements bitbake.BuildLogWriter for streaming logs
type LogWriter struct {
	buildInfo   *BuildInfo
	buildLogger *logger.Logger // structured logger with build context
}

func (lw *LogWriter) WriteLog(stream, content string) {
	logLine := &v1.LogLine{
		Timestamp: time.Now().UnixNano(),
		Stream:    stream,
		Content:   content,
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
func NewServer(address string, log *logger.Logger) *Server {
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
	}
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

// Start starts the gRPC server
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.address, err)
	}

	s.grpcServer = grpc.NewServer()
	v1.RegisterSmidrServer(s.grpcServer, s)

	s.logger.Info("🚀 Smidr daemon listening", slog.String("address", s.address))

	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// Stop gracefully stops the gRPC server and cancels all running builds
func (s *Server) Stop() {
	s.logger.Info("🛑 Stopping daemon...")

	// Cancel all running builds
	s.buildsMutex.Lock()
	s.logger.Info("📋 Cancelling active builds", slog.Int("count", len(s.builds)))
	for buildID, build := range s.builds {
		if build.State == v1.BuildState_BUILD_STATE_BUILDING ||
			build.State == v1.BuildState_BUILD_STATE_PREPARING ||
			build.State == v1.BuildState_BUILD_STATE_QUEUED {
			s.logger.Info("🚫 Cancelling build",
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

	s.logger.Info("✅ Daemon stopped")
}

// StartBuild handles build start requests
func (s *Server) StartBuild(ctx context.Context, req *v1.StartBuildRequest) (*v1.BuildStatus, error) {
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
		LogBuffer:      make([]*v1.LogLine, 0),
		LogSubscribers: make(map[chan *v1.LogLine]bool),
		cancel:         cancel,
	}
	s.builds[buildID] = buildInfo
	s.buildsMutex.Unlock()

	// Start the build in a goroutine
	go s.executeBuild(buildCtx, buildInfo, req)

	return &v1.BuildStatus{
		BuildId:    buildID,
		Target:     req.Target,
		State:      v1.BuildState_BUILD_STATE_QUEUED,
		StartedAt:  buildInfo.StartedAt.Unix(),
		ConfigPath: configPathLabel,
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
		logWriter.WriteLog("stdout", "✅ Build slot acquired, proceeding...")
	case <-ctx.Done():
		s.failBuild(buildInfo.ID, "Build cancelled while waiting for queue slot")
		return
	}

	// Update state to preparing
	s.updateBuildState(buildInfo.ID, v1.BuildState_BUILD_STATE_PREPARING)

	logWriter.WriteLog("stdout", "🚀 Starting build process...")
	logWriter.WriteLog("stdout", fmt.Sprintf("📝 Config: %s", buildInfo.ConfigPath))
	logWriter.WriteLog("stdout", fmt.Sprintf("🎯 Target: %s", req.Target))

	// Build options for runner
	opts := buildpkg.BuildOptions{
		BuildID:    buildInfo.ID,
		Target:     req.Target,
		Customer:   req.Customer,
		ForceClean: req.ForceClean,
		ForceImage: req.ForceImage,
	}

	// Bridge for runner logs -> gRPC stream subscribers
	sink := &runnerLogSink{lw: logWriter}

	// Mark as BUILDING
	s.updateBuildState(buildInfo.ID, v1.BuildState_BUILD_STATE_BUILDING)

	// Use runner to execute build
	runner := buildpkg.NewRunner()
	result, err := runner.Run(ctx, buildInfo.Config, opts, sink)
	if err != nil {
		// Failed
		buildInfo.ExitCode = 1
		buildInfo.ErrorMsg = err.Error()
		buildInfo.CompletedAt = time.Now()
		s.updateBuildState(buildInfo.ID, v1.BuildState_BUILD_STATE_FAILED)
		logWriter.WriteLog("stderr", fmt.Sprintf("❌ Build failed: %v", err))
		return
	}

	// Success or non-zero exit
	buildInfo.ExitCode = int32(result.ExitCode)
	buildInfo.CompletedAt = time.Now()
	if result.Success {
		// Extract artifacts if artifact manager is available and build succeeded
		if s.artifactMgr != nil {
			s.updateBuildState(buildInfo.ID, v1.BuildState_BUILD_STATE_EXTRACTING_ARTIFACTS)
			logWriter.WriteLog("stdout", "📦 Extracting artifacts...")

			if err := s.extractArtifacts(ctx, buildInfo, result, logWriter); err != nil {
				logWriter.WriteLog("stderr", fmt.Sprintf("⚠️ Failed to extract artifacts: %v", err))
				// Don't fail the build just because artifact extraction failed
			} else {
				logWriter.WriteLog("stdout", "✅ Artifacts extracted successfully")
			}
		}

		s.updateBuildState(buildInfo.ID, v1.BuildState_BUILD_STATE_COMPLETED)
		logWriter.WriteLog("stdout", fmt.Sprintf("✅ Build completed in %v", result.Duration))
	} else {
		s.updateBuildState(buildInfo.ID, v1.BuildState_BUILD_STATE_FAILED)
		logWriter.WriteLog("stderr", fmt.Sprintf("❌ Build finished with errors in %v (exit=%d)", result.Duration, result.ExitCode))
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
	logWriter.WriteLog("stdout", fmt.Sprintf("📁 Copying artifacts from %s to %s", deployPath, artifactPath))

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
		logWriter.WriteLog("stderr", fmt.Sprintf("❌ Build failed: %s", errorMsg))
	}
}

// GetBuildStatus retrieves the status of a build
func (s *Server) GetBuildStatus(ctx context.Context, req *v1.BuildStatusRequest) (*v1.BuildStatus, error) {
	s.buildsMutex.RLock()
	build, exists := s.builds[req.BuildId]
	s.buildsMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("build %s not found", req.BuildId)
	}

	status := &v1.BuildStatus{
		BuildId:    build.ID,
		Target:     build.Target,
		State:      build.State,
		ExitCode:   build.ExitCode,
		StartedAt:  build.StartedAt.Unix(),
		ConfigPath: build.ConfigPath,
	}

	if build.ErrorMsg != "" {
		status.ErrorMessage = build.ErrorMsg
	}

	if !build.CompletedAt.IsZero() {
		status.CompletedAt = build.CompletedAt.Unix()
	}

	return status, nil
}

// StreamLogs streams build logs to the client
func (s *Server) StreamLogs(req *v1.StreamLogsRequest, stream v1.Smidr_StreamLogsServer) error {
	s.buildsMutex.RLock()
	build, exists := s.builds[req.BuildId]
	s.buildsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("build %s not found", req.BuildId)
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
		logChan := make(chan *v1.LogLine, 100)

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
func (s *Server) ListArtifacts(ctx context.Context, req *v1.ListArtifactsRequest) (*v1.ArtifactsList, error) {
	s.buildsMutex.RLock()
	build, exists := s.builds[req.BuildId]
	s.buildsMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("build %s not found", req.BuildId)
	}

	if build.State != v1.BuildState_BUILD_STATE_COMPLETED {
		return nil, fmt.Errorf("build %s is not completed", req.BuildId)
	}

	// Return empty list if artifact manager is not available
	if s.artifactMgr == nil {
		return &v1.ArtifactsList{
			BuildId:   req.BuildId,
			Artifacts: []*v1.Artifact{},
		}, nil
	}

	// Get artifacts from artifact manager
	artifactFiles, err := s.artifactMgr.ListArtifacts(req.BuildId)
	if err != nil {
		return nil, fmt.Errorf("failed to list artifacts: %w", err)
	}

	// Load metadata to get file sizes
	metadata, err := s.artifactMgr.LoadMetadata(req.BuildId)
	if err != nil {
		// If we can't load metadata, just return files without sizes
		metadata = &artifacts.BuildMetadata{ArtifactSizes: make(map[string]int64)}
	}

	// Convert to protobuf artifacts
	var protoArtifacts []*v1.Artifact
	for _, artifactFile := range artifactFiles {
		artifact := &v1.Artifact{
			Name: filepath.Base(artifactFile),
			Path: artifactFile,
			Size: metadata.ArtifactSizes[artifactFile],
		}

		// Calculate checksum if file exists
		fullPath := filepath.Join(s.artifactMgr.GetArtifactPath(req.BuildId), artifactFile)
		if checksum, err := s.calculateChecksum(fullPath); err == nil {
			artifact.Checksum = checksum
		}

		protoArtifacts = append(protoArtifacts, artifact)
	}

	return &v1.ArtifactsList{
		BuildId:   req.BuildId,
		Artifacts: protoArtifacts,
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
func (s *Server) CancelBuild(ctx context.Context, req *v1.CancelBuildRequest) (*v1.CancelResult, error) {
	s.buildsMutex.Lock()
	defer s.buildsMutex.Unlock()

	build, exists := s.builds[req.BuildId]
	if !exists {
		return &v1.CancelResult{
			Success: false,
			Message: fmt.Sprintf("build %s not found", req.BuildId),
		}, nil
	}

	if build.State != v1.BuildState_BUILD_STATE_BUILDING {
		return &v1.CancelResult{
			Success: false,
			Message: fmt.Sprintf("build %s is not in a cancellable state", req.BuildId),
		}, nil
	}

	// TODO: Actually cancel the build
	if build.cancel != nil {
		build.cancel()
	}

	build.State = v1.BuildState_BUILD_STATE_CANCELLED
	build.CompletedAt = time.Now()

	return &v1.CancelResult{
		Success: true,
		Message: "Build cancelled successfully",
	}, nil
}

// ListBuilds lists all builds
func (s *Server) ListBuilds(ctx context.Context, req *v1.ListBuildsRequest) (*v1.BuildsList, error) {
	s.buildsMutex.RLock()
	defer s.buildsMutex.RUnlock()

	builds := make([]*v1.BuildStatus, 0)
	for _, build := range s.builds {
		// Filter by state if requested
		if len(req.States) > 0 {
			matched := false
			for _, state := range req.States {
				if build.State == state {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		status := &v1.BuildStatus{
			BuildId:    build.ID,
			Target:     build.Target,
			State:      build.State,
			ExitCode:   build.ExitCode,
			StartedAt:  build.StartedAt.Unix(),
			ConfigPath: build.ConfigPath,
		}

		if build.ErrorMsg != "" {
			status.ErrorMessage = build.ErrorMsg
		}

		if !build.CompletedAt.IsZero() {
			status.CompletedAt = build.CompletedAt.Unix()
		}

		builds = append(builds, status)

		// Limit results if requested
		if req.Limit > 0 && len(builds) >= int(req.Limit) {
			break
		}
	}

	return &v1.BuildsList{
		Builds: builds,
	}, nil
}
