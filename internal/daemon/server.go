package daemon

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"

	v1 "github.com/schererja/smidr/api/proto"
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
	ID            string
	Target        string
	State         v1.BuildState
	ExitCode      int32
	ErrorMsg      string
	StartedAt     time.Time
	CompletedAt   time.Time
	ConfigPath    string
	Config        *config.Config
	LogBuffer     []v1.LogLine
	LogMutex      sync.RWMutex
	LogSubscribers map[chan *v1.LogLine]bool
	cancel        context.CancelFunc
	ArtifactPaths []string
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
	lw.buildInfo.LogBuffer = append(lw.buildInfo.LogBuffer, *logLine)

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
	// Load configuration
	cfg, err := config.Load(req.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Generate a unique build ID
	buildID := fmt.Sprintf("build-%d", time.Now().UnixNano())

	buildCtx, cancel := context.WithCancel(context.Background())

	s.buildsMutex.Lock()
	buildInfo := &BuildInfo{
		ID:             buildID,
		Target:         req.Target,
		State:          v1.BuildState_BUILD_STATE_QUEUED,
		StartedAt:      time.Now(),
		ConfigPath:     req.Config,
		Config:         cfg,
		LogBuffer:      make([]v1.LogLine, 0),
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
		ConfigPath: req.Config,
	}, nil
}

// executeBuild runs the actual build process
func (s *Server) executeBuild(ctx context.Context, buildInfo *BuildInfo, req *v1.StartBuildRequest) {
	logWriter := &LogWriter{buildInfo: buildInfo}

	// Update state to preparing
	s.updateBuildState(buildInfo.ID, v1.BuildState_BUILD_STATE_PREPARING)

	logWriter.WriteLog("stdout", "🚀 Starting build process...")
	logWriter.WriteLog("stdout", fmt.Sprintf("📝 Config: %s", req.Config))
	logWriter.WriteLog("stdout", fmt.Sprintf("🎯 Target: %s", req.Target))

	// TODO: Implement full build workflow similar to CLI
	// For now, just mark as completed
	time.Sleep(2 * time.Second) // Simulate some work

	s.buildsMutex.Lock()
	buildInfo.State = v1.BuildState_BUILD_STATE_COMPLETED
	buildInfo.CompletedAt = time.Now()
	buildInfo.ExitCode = 0
	s.buildsMutex.Unlock()

	logWriter.WriteLog("stdout", "✅ Build completed successfully (stub implementation)!")
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
		if err := stream.Send(&logLine); err != nil {
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
