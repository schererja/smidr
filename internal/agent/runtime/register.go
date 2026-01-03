package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/intrik8-labs/smidr/internal/agent/client"
	"github.com/intrik8-labs/smidr/internal/logging"
)

// RegistrationRequest represents the agent capabilities sent to control plane
type RegistrationRequest struct {
	AgentID      string            `json:"agent_id"`
	Capabilities AgentCapabilities `json:"capabilities"`
	Timestamp    time.Time         `json:"timestamp"`
}

// AgentCapabilities describes what this agent can do
type AgentCapabilities struct {
	Architecture string   `json:"architecture"`
	OS           string   `json:"os"`
	MaxJobs      int      `json:"max_jobs"`
	Features     []string `json:"features"`
}

// RegistrationResponse represents the control plane acknowledgment
type RegistrationResponse struct {
	Registered bool      `json:"registered"`
	AgentID    string    `json:"agent_id"`
	Message    string    `json:"message"`
	Timestamp  time.Time `json:"timestamp"`
}

// Register announces the agent to the control plane
func (r *Runtime) Register(ctx context.Context, demoMode bool) error {
	log := logging.FromContext(ctx)

	log.InfoContext(ctx, "registering agent with control plane",
		logging.String("control_plane_uri", r.cfg.ControlPlane.URI),
		logging.String("agent_id", r.cfg.AgentConfig.ID),
		logging.Bool("demo_mode", demoMode),
	)

	// Build registration data
	capabilities := []string{"yocto", "docker", "command"}
	metadata := map[string]string{
		"arch": "amd64",
		"os":   "linux",
	}

	if demoMode {
		// Simulate registration in demo mode
		resp := r.simulateRegistration(ctx, RegistrationRequest{
			AgentID: r.cfg.AgentConfig.ID,
			Capabilities: AgentCapabilities{
				Architecture: "amd64",
				OS:           "linux",
				MaxJobs:      1,
				Features:     []string{"docker", "build"},
			},
			Timestamp: time.Now().UTC(),
		})

		if !resp.Registered {
			return fmt.Errorf("registration failed: %s", resp.Message)
		}

		log.InfoContext(ctx, "agent registered successfully (demo mode)",
			logging.String("agent_id", resp.AgentID),
		)
		return nil
	}

	// Real registration with control plane
	c := client.NewClient(r.cfg.ControlPlane.URI)
	if err := c.RegisterAgent(ctx, r.cfg.AgentConfig.ID, "smidr-agent", capabilities, metadata); err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	log.InfoContext(ctx, "agent registered successfully",
		logging.String("agent_id", r.cfg.AgentConfig.ID),
	)

	return nil
}

// simulateRegistration mocks the control plane response
// TODO: Remove this when control plane API is ready
func (r *Runtime) simulateRegistration(ctx context.Context, req RegistrationRequest) RegistrationResponse {
	log := logging.FromContext(ctx)

	log.DebugContext(ctx, "simulating control plane registration",
		logging.String("architecture", req.Capabilities.Architecture),
		logging.String("os", req.Capabilities.OS),
		logging.Int("max_jobs", req.Capabilities.MaxJobs),
	)

	// Simulate network delay
	time.Sleep(100 * time.Millisecond)

	return RegistrationResponse{
		Registered: true,
		AgentID:    req.AgentID,
		Message:    "agent registered (simulated)",
		Timestamp:  time.Now().UTC(),
	}
}
