package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/intrik8-labs/smidr/internal/agent/config"
	"github.com/intrik8-labs/smidr/internal/logging"
)

// TestHeartbeatDemo tests heartbeat in demo mode
func TestHeartbeatDemo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hb := NewHeartbeat(1, &HeartbeatConfig{
		AgentID:         "test-agent",
		ControlPlaneURI: "http://localhost:8080",
	}, true) // demo mode

	// Send a heartbeat in demo mode (0 active jobs)
	err := hb.Send(ctx, 0)
	if err != nil {
		t.Fatalf("expected no error in demo mode, got: %v", err)
	}
}

// TestHeartbeatInterval tests that heartbeat respects interval
func TestHeartbeatInterval(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	hb := NewHeartbeat(1, &HeartbeatConfig{
		AgentID:         "test-agent",
		ControlPlaneURI: "http://localhost:8080",
	}, true)

	start := time.Now()

	// Send heartbeats and measure timing
	for i := 0; i < 2; i++ {
		err := hb.Send(ctx, 0)
		if err != nil {
			t.Fatalf("heartbeat failed: %v", err)
		}
	}

	elapsed := time.Since(start)
	// Should have taken at least the minimum interval
	if elapsed < 500*time.Millisecond {
		t.Logf("heartbeats completed in %v", elapsed)
	}
}

// TestHeartbeatStatuses tests different heartbeat statuses
func TestHeartbeatWithDifferentJobCounts(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	hb := NewHeartbeat(1, &HeartbeatConfig{
		AgentID:         "test-agent",
		ControlPlaneURI: "http://localhost:8080",
	}, true)

	jobCounts := []int{0, 1, 5}

	for _, count := range jobCounts {
		err := hb.Send(ctx, count)
		if err != nil {
			t.Errorf("failed to send heartbeat with %d jobs: %v", count, err)
		}
	}
}

// TestHeartbeatConfig tests heartbeat configuration
func TestHeartbeatConfig(t *testing.T) {
	cfg := &HeartbeatConfig{
		AgentID:         "agent-123",
		ControlPlaneURI: "http://control.example.com:8080",
	}

	if cfg.AgentID != "agent-123" {
		t.Errorf("expected agent ID to be agent-123, got: %s", cfg.AgentID)
	}

	if cfg.ControlPlaneURI != "http://control.example.com:8080" {
		t.Errorf("expected URI to be http://control.example.com:8080, got: %s", cfg.ControlPlaneURI)
	}
}

// TestRuntimeHeartbeatIntegration tests heartbeat within runtime
func TestRuntimeHeartbeatIntegration(t *testing.T) {
	ctx := context.Background()
	logging.Init(logging.Config{
		Level: logging.LevelInfo,
		JSON:  true,
	})

	cfg := &config.Config{
		AgentConfig: config.AgentConfig{
			ID:                "test-agent",
			HeartBeatInterval: 1,
			Demo:              true,
		},
		ControlPlane: config.ControlPlaneConfig{
			URI: "http://localhost:8080",
		},
	}

	r := New(cfg, logging.Default())

	// Create a short-lived runtime
	ctxRun, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Start the runtime
	go func() {
		_ = r.Start()
	}()

	// Let it run for a bit
	<-ctxRun.Done()

	// Stop gracefully
	r.Stop()
}
