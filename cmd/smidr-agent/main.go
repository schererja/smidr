package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/schererja/smidr/internal/agent/metrics"
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
	var lastLatency int64
	for {
		ctx := context.Background()
		start := time.Now()
		_, err := client.Heartbeat(ctx, &pb.HeartbeatRequest{
			AgentId:   &pb.AgentID{AgentId: agentID},
			EventTime: timestamppb.New(time.Now()),
			Status:    pb.AgentStatus_AGENT_STATUS_IDLE,
			Metrics:   getMetrics(lastLatency),
			Version:   "v1.0.0",
		})
		lastLatency = time.Since(start).Milliseconds()
		if err != nil {
			log.Fatalf("failed to send heartbeat: %v", err)
			lastLatency = -1
		} else {
			fmt.Printf("Heartbeat latency: %d ms\n", lastLatency)
			fmt.Printf("Heartbeat sent successfully\n")
		}
		time.Sleep(10 * time.Second)
	}
}

func getMetrics(lastLatency int64) *pb.Metrics {
	return &pb.Metrics{
		Metrics: []*pb.Metric{
			{
				Name:      "cpu_usage",
				Value:     metrics.GetCpuUsage(),
				Timestamp: timestamppb.New(time.Now()),
			},
			{
				Name:      "memory",
				Value:     metrics.GetMemoryUsage(),
				Timestamp: timestamppb.New(time.Now()),
			},
			{
				Name:      "network_latency",
				Value:     metrics.GetNetworkLatency(lastLatency),
				Timestamp: timestamppb.New(time.Now()),
			},
			{
				Name:      "disk_usage",
				Value:     metrics.GetDiskUsage(),
				Timestamp: timestamppb.New(time.Now()),
			},
		},
	}
}
