package controlplane

import (
	"fmt"
	"io"
	"log/slog"
	"net"

	config "github.com/schererja/smidr/internal/config/controlplane"
	pb "github.com/schererja/smidr/pkg/agent/v1"
	"github.com/schererja/smidr/pkg/logger"
	"google.golang.org/grpc"
)

type controlPlaneServer struct {
	grpcServer *grpc.Server
	listener   net.Listener
	address    string
	port       string
	logger     *logger.Logger
	pb.UnimplementedAgentServiceServer
}

func NewControlPlaneServer(serverConfig *config.ServerConfig, logger *logger.Logger) (*controlPlaneServer, error) {
	lis, err := net.Listen("tcp", net.JoinHostPort(serverConfig.Address, serverConfig.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s:%s: %w", serverConfig.Address, serverConfig.Port, err)
	}

	grpcServer := grpc.NewServer()

	// Register the service
	pb.RegisterAgentServiceServer(grpcServer, &controlPlaneServer{
		logger: logger,
	})

	cps := &controlPlaneServer{
		grpcServer: grpcServer,
		listener:   lis,
		address:    serverConfig.Address,
		port:       serverConfig.Port,
		logger:     logger,
	}
	return cps, nil
}

func (s *controlPlaneServer) Start() error {
	if s.listener == nil {
		return fmt.Errorf("listener not initialized")
	}
	s.logger.Info("Control plane server listening",
		slog.String("address", s.address),
		slog.String("port", s.port))
	return s.grpcServer.Serve(s.listener)
}
func (s *controlPlaneServer) Connect(stream pb.AgentService_ConnectServer) error {

	var agentID string

	// Receive messages from agent
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			s.logger.Info("Agent disconnected", slog.String("agent_id", agentID))
			return nil
		}
		if err != nil {
			s.logger.Error("Error receiving from agent", err)
			return err
		}

		// Handle different message types
		switch payload := msg.Payload.(type) {
		case *pb.AgentMessage_Heartbeat:
			agentID = msg.AgentId
			s.logger.Debug("Heartbeat received",
				slog.String("agent_id", agentID),
				slog.String("status", payload.Heartbeat.Status))

		case *pb.AgentMessage_TaskResult:
			s.logger.Info("Task result received",
				slog.String("agent_id", agentID),
				slog.String("task_id", payload.TaskResult.TaskId))
		}
	}
}

// Stop stops the gRPC server
func (cps *controlPlaneServer) Stop() {
	cps.logger.Info("Stopping Control Plane...")

	if cps.grpcServer != nil {
		cps.grpcServer.Stop()
	}

	cps.logger.Info("Control Plane stopped")
}
