package main

import (
	"log"
	"net"

	"github.com/schererja/smidr/internal/config"
	"github.com/schererja/smidr/internal/server"
	pb "github.com/schererja/smidr/pkg/smidr-sdk/v1"
	"google.golang.org/grpc"
)

func main() {
	StartGrpcServer("config.yaml")

}
func StartGrpcServer(configPath string) {
	cfg, err := config.Load(configPath)
	if err != nil {
		// Handle error appropriately, e.g., log and exit
	}
	_ = cfg
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen on port 50051: %v", err)
	}

	s := grpc.NewServer()
	smidrServer := server.NewServer()
	pb.RegisterAgentServiceServer(s, smidrServer)
	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
