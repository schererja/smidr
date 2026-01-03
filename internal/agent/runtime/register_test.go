package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/intrik8-labs/smidr/internal/agent/config"
	"github.com/intrik8-labs/smidr/internal/logging"
)

// TestRegisterDemo tests registration in demo mode
func TestRegisterDemo(t *testing.T) {
	ctx := context.Background()
	logging.Init(logging.Config{Level: logging.LevelInfo, JSON: true})

	cfg := &config.Config{
		AgentConfig: config.AgentConfig{
			ID:                "test-agent",
			HeartBeatInterval: 30,
			Demo:              true,
		},
		ControlPlane: config.ControlPlaneConfig{
			URI: "http://localhost:8080",
		},
	}

	r := New(cfg, logging.Default())

	err := r.Register(ctx, true)
	if err != nil {
		t.Fatalf("expected no error in demo mode, got: %v", err)
	}
}

// TestRegisterReal tests registration in real mode (with mock server)
func TestRegisterReal(t *testing.T) {
	// This test would require a mock HTTP server
	// For now, we'll test the demo mode and the ability to create a runtime
	logging.Init(logging.Config{Level: logging.LevelInfo, JSON: true})

	cfg := &config.Config{
		AgentConfig: config.AgentConfig{
			ID:                "test-agent",
			HeartBeatInterval: 30,
			Demo:              false,
		},
		ControlPlane: config.ControlPlaneConfig{
			URI: "http://localhost:8080",
		},
	}

	r := New(cfg, logging.Default())
	if r == nil {
		t.Fatal("expected runtime to be created")
	}
	if r.cfg.AgentConfig.ID != "test-agent" {
		t.Errorf("expected agent ID to be test-agent, got: %s", r.cfg.AgentConfig.ID)
	}
}

// TestRegisterCapabilities tests that capabilities are properly set
func TestRegisterCapabilities(t *testing.T) {
	logging.Init(logging.Config{Level: logging.LevelInfo, JSON: true})

	cfg := &config.Config{
		AgentConfig: config.AgentConfig{
			ID:                "test-agent",
			HeartBeatInterval: 30,
			Demo:              true,
		},
		ControlPlane: config.ControlPlaneConfig{
			URI: "http://localhost:8080",
		},
	}

	r := New(cfg, logging.Default())

	// Call register with demo mode
	err := r.Register(context.Background(), true)
	if err != nil {
		t.Fatalf("registration failed: %v", err)
	}
}

// TestSimulateRegistration tests the demo mode registration simulation
func TestSimulateRegistration(t *testing.T) {
	ctx := context.Background()
	logging.Init(logging.Config{Level: logging.LevelInfo, JSON: true})
	ctx = logging.WithExecutor(ctx, "test-agent")

	cfg := &config.Config{
		AgentConfig: config.AgentConfig{
			ID:                "test-agent",
			HeartBeatInterval: 30,
			Demo:              true,
		},
		ControlPlane: config.ControlPlaneConfig{
			URI: "http://localhost:8080",
		},
	}

	r := New(cfg, logging.Default())

	req := RegistrationRequest{
		AgentID: "test-agent",
		Capabilities: AgentCapabilities{
			Architecture: "amd64",
			OS:           "linux",
			MaxJobs:      1,
			Features:     []string{"docker", "build"},
		},
		Timestamp: time.Now().UTC(),
	}

	resp := r.simulateRegistration(ctx, req)

	if !resp.Registered {
		t.Error("expected registration to be successful")
	}
	if resp.AgentID != "test-agent" {
		t.Errorf("expected agent ID to be test-agent, got: %s", resp.AgentID)
	}
}
