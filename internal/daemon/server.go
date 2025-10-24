package daemon

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	v1 "github.com/schererja/smidr/api/proto"
	buildpkg "github.com/schererja/smidr/internal/build"
	"github.com/schererja/smidr/internal/config"
)

// Server implements the Smidr gRPC service
type Server struct {
	v1.UnimplementedSmidrServer

	address     string
	grpcServer  *grpc.Server
	builds      map[string]*BuildInfo
	buildsMutex sync.RWMutex
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
	buildInfo *BuildInfo
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
}

// NewServer creates a new daemon server
func NewServer(address string) *Server {
	return &Server{
		address: address,
		builds:  make(map[string]*BuildInfo),
	}
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

	fmt.Printf("🚀 Smidr daemon listening on %s\n", s.address)

	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	if s.grpcServer != nil {
		fmt.Println("🛑 Stopping daemon...")
		s.grpcServer.GracefulStop()
	}
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
	logWriter := &LogWriter{buildInfo: buildInfo}

	// Update state to preparing
	s.updateBuildState(buildInfo.ID, v1.BuildState_BUILD_STATE_PREPARING)

	logWriter.WriteLog("stdout", "🚀 Starting build process...")
	logWriter.WriteLog("stdout", fmt.Sprintf("📝 Config: %s", buildInfo.ConfigPath))
	logWriter.WriteLog("stdout", fmt.Sprintf("🎯 Target: %s", req.Target))

	// Build options for runner
	opts := buildpkg.BuildOptions{
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

	// TODO: Actually list artifacts from the build
	return &v1.ArtifactsList{
		BuildId:   req.BuildId,
		Artifacts: []*v1.Artifact{},
	}, nil
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
