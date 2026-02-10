package controlplane

import (
	"context"
	"net"
	"testing"
	"time"

	pb "github.com/schererja/smidr/pkg/agent/v1"
	"github.com/schererja/smidr/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func dialer(lis *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, s string) (net.Conn, error) {
		return lis.Dial()
	}
}

func TestConnectRegistersAgentCompatibilities(t *testing.T) {
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	cps := &controlPlaneServer{
		grpcServer: srv,
		listener:   lis,
		logger:     logger.NewLogger(),
		agents:     make(map[string]*AgentInfo),
	}
	pb.RegisterAgentServiceServer(srv, cps)

	go func() {
		_ = srv.Serve(lis)
	}()
	defer srv.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(dialer(lis)), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewAgentServiceClient(conn)
	stream, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("failed to open stream: %v", err)
	}

	agentID := "agent-test"
	compat := []string{"shell", "docker"}

	err = stream.Send(&pb.AgentMessage{
		AgentId:   agentID,
		Timestamp: time.Now().UnixMilli(),
		Payload: &pb.AgentMessage_Register{
			Register: &pb.Register{
				Compatibilities: compat,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to send registration: %v", err)
	}

	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		cps.agentsMu.RLock()
		info := cps.agents[agentID]
		cps.agentsMu.RUnlock()
		if info != nil {
			if len(info.Compatibilities) != len(compat) {
				t.Fatalf("expected compatibilities %v, got %v", compat, info.Compatibilities)
			}
			for i := range compat {
				if info.Compatibilities[i] != compat[i] {
					t.Fatalf("expected compatibilities %v, got %v", compat, info.Compatibilities)
				}
			}
			_ = stream.CloseSend()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	_ = stream.CloseSend()
	cps.agentsMu.RLock()
	_, ok := cps.agents[agentID]
	cps.agentsMu.RUnlock()
	if !ok {
		t.Fatalf("expected agent to be registered")
	}
}
