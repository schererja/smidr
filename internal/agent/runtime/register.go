package runtime

import (
	"context"
	"fmt"
	"time"

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
func (r *Runtime) Register(ctx context.Context) error {
	log := logging.FromContext(ctx)

	log.InfoContext(ctx, "registering agent with control plane",
		logging.String("control_plane_uri", r.cfg.ControlPlane.URI),
		logging.String("agent_id", r.cfg.AgentConfig.ID),
	)

	// Build registration request
	req := RegistrationRequest{
		AgentID: r.cfg.AgentConfig.ID,
		Capabilities: AgentCapabilities{
			Architecture: "amd64", // TODO: Detect actual architecture
			OS:           "linux", // TODO: Detect actual OS
			MaxJobs:      1,       // TODO: Make configurable
			Features: []string{
				"docker",
				"build",
			},
		},
		Timestamp: time.Now().UTC(),
	}

	// TODO: Replace with real HTTP call to control plane
	// Example:
	// resp, err := r.httpClient.Post(
	//     r.cfg.ControlPlane.URI + "/v1/agents/register",
	//     "application/json",
	//     marshal(req),
	// )

	// Simulate registration (remove when control plane is ready)
	resp := r.simulateRegistration(ctx, req)

	if !resp.Registered {
		return fmt.Errorf("registration failed: %s", resp.Message)
	}

	log.InfoContext(ctx, "agent registered successfully",
		logging.String("agent_id", resp.AgentID),
		logging.String("message", resp.Message),
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
