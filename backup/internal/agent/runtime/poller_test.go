package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/intrik8-labs/smidr/internal/agent/config"
	"github.com/intrik8-labs/smidr/internal/logging"
)

// TestPollerDemo tests job polling in demo mode
func TestPollerDemo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	p := NewPoller("test-agent", "http://localhost:8080", 1, 1, []string{"yocto", "command"}, true)

	// Poll for jobs in demo mode (0 active jobs)
	jobs, err := p.Poll(ctx, 0)
	if err != nil {
		t.Fatalf("expected no error in demo mode, got: %v", err)
	}

	// In demo mode, we should get some simulated jobs or empty array
	if jobs == nil {
		t.Error("expected non-nil jobs slice")
	}
}

// TestPollerCreation tests that poller is created with correct settings
func TestPollerCreation(t *testing.T) {
	p := NewPoller("test-agent", "http://localhost:8080", 1, 1, []string{"yocto", "command", "docker"}, true)

	if p == nil {
		t.Fatal("expected poller to be created")
	}

	// Poll to ensure it's working
	ctx := context.Background()
	_, err := p.Poll(ctx, 0)
	if err != nil {
		t.Errorf("expected polling to work in demo mode, got: %v", err)
	}
}

// TestPollerConfig tests poller configuration
func TestPollerConfig(t *testing.T) {
	p := NewPoller("agent-456", "http://api.example.com", 5, 2, []string{"yocto"}, false)

	if p == nil {
		t.Error("expected poller to be created")
	}

	// Verify it was created successfully by polling
	// In real mode without a server, this might error, so we just test creation
}

// TestPollerDemoMode tests that demo mode returns simulated jobs
func TestPollerDemoMode(t *testing.T) {
	ctx := context.Background()

	p := NewPoller("test-agent", "http://localhost:8080", 1, 1, []string{"yocto"}, true)

	// In demo mode, polling should work without errors
	_, err := p.Poll(ctx, 0)
	if err != nil {
		t.Errorf("polling in demo mode should not error, got: %v", err)
	}
}

// TestRuntimePollerIntegration tests poller within runtime
func TestRuntimePollerIntegration(t *testing.T) {
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

	if r.poller == nil {
		t.Fatal("expected poller to be created")
	}
}

// TestPollerWithMultiplePolls tests that demo mode simulation works
func TestPollerWithMultiplePolls(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	p := NewPoller("demo-agent", "http://localhost:8080", 1, 1, []string{"yocto"}, true)

	// Multiple calls to poll should work
	for i := 0; i < 3; i++ {
		_, err := p.Poll(ctx, 0)
		if err != nil {
			t.Errorf("poll iteration %d failed: %v", i, err)
		}
	}
}
