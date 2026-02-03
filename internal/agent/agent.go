package agent

import (
	"context"
	"log"
	"time"

	config "github.com/schererja/smidr/internal/config/agent"
	pb "github.com/schererja/smidr/pkg/agent/v1"
	"google.golang.org/grpc"
)

type Agent struct {
	// Agent fields
	agentConfig config.AgentConfig
}

func NewAgent(config *config.AgentConfig) *Agent {
	return &Agent{
		agentConfig: *config,
	}
}
func (a *Agent) Start() error {

	// Start agent logic
	userConn, err := grpc.Dial(a.agentConfig.ControlPlaneAddress, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer userConn.Close()

	client := pb.NewAgentServiceClient(userConn)
	stream, err := client.Connect(context.Background())
	if err != nil {
		log.Fatalf("Failed to connect stream: %v", err)
	}

	agentID := "agent" + time.Now().Format("20060102150405")

	// Send heartbeats
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			err := stream.Send(&pb.AgentMessage{
				AgentId:   agentID,
				Timestamp: time.Now().UnixMilli(),
				Payload: &pb.AgentMessage_Heartbeat{
					Heartbeat: &pb.Heartbeat{
						Status: "online",
					},
				},
			})
			if err != nil {
				log.Printf("Failed to send heartbeat: %v", err)
				return
			}
			log.Printf("Sent heartbeat from %s", agentID)
		}
	}()

	// Receive and execute tasks
	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Printf("Error receiving: %v", err)
			break
		}

		switch payload := msg.Payload.(type) {
		case *pb.ControlPlaneMessage_Task:
			task := payload.Task
			log.Printf("Received task: %s", task.TaskId)

		}
	}
	return nil

}
func (a *Agent) Connect(stream pb.AgentService_ConnectClient) error {
	// Connection logic
	return nil
}
func (a *Agent) Stop() error {
	// Stop agent logic
	return nil
}
