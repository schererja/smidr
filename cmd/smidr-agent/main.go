package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/schererja/smidr/pkg/smidr-sdk/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {

	fmt.Printf("Starting smidr agent...")
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to gRPC server at localhost:50051: %v", err)
	}
	defer conn.Close()
	client := pb.NewAgentServiceClient(conn)
	agentID := "agent-123"
	ctx := context.Background()

	_, err = client.RegisterAgent(ctx, &pb.RegisterAgentRequest{
		AgentId: &pb.AgentID{AgentId: agentID},
	})
	if err != nil {
		log.Fatalf("failed to register agent: %v", err)
	}
	fmt.Printf("Agent registered successfully\n")
	go sendHeartbeat(client, agentID)

	select {}
}

func sendHeartbeat(client pb.AgentServiceClient, agentID string) {
	for {
		ctx := context.Background()
		_, err := client.Heartbeat(ctx, &pb.HeartbeatRequest{
			AgentId:   &pb.AgentID{AgentId: agentID},
			EventTime: timestamppb.New(time.Now()),
			Status:    pb.AgentStatus_AGENT_STATUS_IDLE,
			Metrics:   getMetrics(),
			Version:   "v1.0.0",
		})
		if err != nil {
			log.Fatalf("failed to send heartbeat: %v", err)
		}
		fmt.Printf("Heartbeat sent successfully\n")
		time.Sleep(10 * time.Second)
	}
}
func getCpuUsage() float64 {
	// Placeholder for actual CPU usage retrieval logic
	return 42.0
}

func getMemoryUsage() float64 {
	// Placeholder for actual Memory usage retrieval logic
	return 73.5
}

func getMetrics() *pb.Metrics {
	return &pb.Metrics{
		Metrics: []*pb.Metric{
			{
				Name:      "cpu_usage",
				Value:     getCpuUsage(),
				Timestamp: timestamppb.New(time.Now()),
			},
			{
				Name:      "memory_usage",
				Value:     getMemoryUsage(),
				Timestamp: timestamppb.New(time.Now()),
			},
			{
				Name:      "network_latency",
				Value:     getNetworkLatency(),
				Timestamp: timestamppb.New(time.Now()),
			},
			{
				Name:      "disk_usage",
				Value:     getDiskUsage(),
				Timestamp: timestamppb.New(time.Now()),
			},
		},
	}
}

func getNetworkLatency() float64 {
	// Placeholder for actual Network Latency retrieval logic
	return 15.3
}
func getDiskUsage() float64 {
	// Placeholder for actual Disk Usage retrieval logic
	return 58.2
}
