package runtime

import (
	"testing"
	"time"

	"github.com/intrik8-labs/smidr/internal/agent/config"
	"github.com/intrik8-labs/smidr/internal/logging"
)

// TestNewRuntime tests runtime creation
func TestNewRuntime(t *testing.T) {
	logging.Init(logging.Config{
		Level: logging.LevelInfo,
		JSON:  true,
	})
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

	if r == nil {
		t.Fatal("expected runtime to be created")
	}

	if r.cfg.AgentConfig.ID != "test-agent" {
		t.Errorf("expected agent ID to be test-agent, got: %s", r.cfg.AgentConfig.ID)
	}

	if r.heartbeat == nil {
		t.Error("expected heartbeat to be initialized")
	}

	if r.poller == nil {
		t.Error("expected poller to be initialized")
	}

	if r.plugins == nil {
		t.Error("expected plugins map to be initialized")
	}
}

// TestRuntimePlugins tests that plugins are properly initialized
func TestRuntimePlugins(t *testing.T) {
	logging.Init(logging.Config{
		Level: logging.LevelInfo,
		JSON:  true,
	})
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

	if len(r.plugins) == 0 {
		t.Error("expected plugins to be registered")
	}

	// Check for expected plugins
	expectedPlugins := []string{"command", "yocto"}
	for _, pluginName := range expectedPlugins {
		if _, exists := r.plugins[pluginName]; !exists {
			t.Errorf("expected plugin %s to be registered", pluginName)
		}
	}
}

// TestRuntimeStartStop tests starting and stopping the runtime
func TestRuntimeStartStop(t *testing.T) {
	logging.Init(logging.Config{Level: logging.LevelInfo, JSON: true})
	cfg := &config.Config{
		AgentConfig: config.AgentConfig{
			ID:                "test-agent",
			HeartBeatInterval: 1, // Short interval for testing
			Demo:              true,
		},
		ControlPlane: config.ControlPlaneConfig{
			URI: "http://localhost:8080",
		},
	}

	r := New(cfg, logging.Default())

	// Start the runtime
	err := r.Start()
	if err != nil {
		t.Fatalf("failed to start runtime: %v", err)
	}

	// Let it run for a bit
	time.Sleep(2 * time.Second)

	// Stop the runtime
	r.Stop()
}

// TestRuntimeContextCancellation tests that runtime respects context cancellation
func TestRuntimeContextCancellation(t *testing.T) {
	logging.Init(logging.Config{Level: logging.LevelInfo, JSON: true})
	cfg := &config.Config{
		AgentConfig: config.AgentConfig{
			ID:                "test-agent",
			HeartBeatInterval: 10,
			Demo:              true,
		},
		ControlPlane: config.ControlPlaneConfig{
			URI: "http://localhost:8080",
		},
	}

	r := New(cfg, logging.Default())

	err := r.Start()
	if err != nil {
		t.Fatalf("failed to start runtime: %v", err)
	}

	// Immediately stop
	r.Stop()

	// Should complete without hanging
}

// TestRuntimeActiveJobs tests job tracking
func TestRuntimeActiveJobs(t *testing.T) {
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

	// Initially should have 0 active jobs
	if r.activeJobs != 0 {
		t.Errorf("expected 0 active jobs, got: %d", r.activeJobs)
	}
}

// TestRuntimeDemoMode tests runtime in demo mode
func TestRuntimeDemoMode(t *testing.T) {
	logging.Init(logging.Config{Level: logging.LevelInfo, JSON: true})
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

	// Start in demo mode
	err := r.Start()
	if err != nil {
		t.Fatalf("failed to start runtime in demo mode: %v", err)
	}

	// Run for a short time
	time.Sleep(2 * time.Second)

	// Stop
	r.Stop()
}

// TestRuntimeConfig tests runtime configuration
func TestRuntimeConfig(t *testing.T) {
	logging.Init(logging.Config{Level: logging.LevelInfo, JSON: true})

	tests := []struct {
		name             string
		cfg              *config.Config
		expectedID       string
		expectedURI      string
		expectedDemo     bool
		expectedInterval int
	}{
		{
			name: "production config",
			cfg: &config.Config{
				AgentConfig: config.AgentConfig{
					ID:                "prod-agent",
					HeartBeatInterval: 30,
					Demo:              false,
				},
				ControlPlane: config.ControlPlaneConfig{
					URI: "https://prod.example.com",
				},
			},
			expectedID:       "prod-agent",
			expectedURI:      "https://prod.example.com",
			expectedDemo:     false,
			expectedInterval: 30,
		},
		{
			name: "demo config",
			cfg: &config.Config{
				AgentConfig: config.AgentConfig{
					ID:                "demo-agent",
					HeartBeatInterval: 5,
					Demo:              true,
				},
				ControlPlane: config.ControlPlaneConfig{
					URI: "http://localhost:8080",
				},
			},
			expectedID:       "demo-agent",
			expectedURI:      "http://localhost:8080",
			expectedDemo:     true,
			expectedInterval: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.cfg, logging.Default())

			if r.cfg.AgentConfig.ID != tt.expectedID {
				t.Errorf("expected agent ID %s, got: %s", tt.expectedID, r.cfg.AgentConfig.ID)
			}

			if r.cfg.ControlPlane.URI != tt.expectedURI {
				t.Errorf("expected URI %s, got: %s", tt.expectedURI, r.cfg.ControlPlane.URI)
			}

			if r.cfg.AgentConfig.Demo != tt.expectedDemo {
				t.Errorf("expected demo mode %v, got: %v", tt.expectedDemo, r.cfg.AgentConfig.Demo)
			}

			if r.cfg.AgentConfig.HeartBeatInterval != tt.expectedInterval {
				t.Errorf("expected heartbeat interval %d, got: %d", tt.expectedInterval, r.cfg.AgentConfig.HeartBeatInterval)
			}
		})
	}
}
