package server

import (
	"context"
	"fmt"

	pb "github.com/schererja/smidr/pkg/smidr-sdk/v1"
)

type Server struct {
	pb.UnimplementedAgentServiceServer
}

func NewServer() *Server {
	return &Server{}
}
func (s *Server) RegisterAgent(ctx context.Context, req *pb.RegisterAgentRequest) (*pb.RegisterAgentResponse, error) {
	fmt.Println("RegisterAgent called")
	return &pb.RegisterAgentResponse{
		AgentId: req.AgentId,
		Success: true,
	}, nil
}

func (s *Server) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	fmt.Println("Heartbeat found: ", req)
	return &pb.HeartbeatResponse{}, nil
}
