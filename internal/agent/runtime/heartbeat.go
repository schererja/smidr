package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/intrik8-labs/smidr/internal/logging"
)

// HeartbeatRequest represents the agent status sent to control plane
type HeartbeatRequest struct {
	AgentID   string        `json:"agent_id"`
	Status    AgentStatus   `json:"status"`
	Resources AgentResources `json:"resources"`
	Timestamp time.Time     `json:"timestamp"`
}

// AgentStatus represents current agent state
type AgentStatus struct {
	State      string `json:"state"`       // "idle", "busy", "draining"
	ActiveJobs int    `json:"active_jobs"`
	Uptime     int64  `json:"uptime_seconds"`
}

// AgentResources represents available resources
type AgentResources struct {
	AvailableCPU    float64 `json:"available_cpu"`
	AvailableMemory int64   `json:"available_memory_mb"`
	DiskSpace       int64   `json:"disk_space_mb"`
}

// HeartbeatResponse represents control plane acknowledgment
type HeartbeatResponse struct {
	Acknowledged bool      `json:"acknowledged"`
	Commands     []string  `json:"commands,omitempty"` // Future: control plane commands
	Timestamp    time.Time `json:"timestamp"`
}

// Heartbeat manages periodic health signals to the control plane
type Heartbeat struct {
	intervalSeconds int
	startTime       time.Time
	cfg             *HeartbeatConfig
}

// HeartbeatConfig holds heartbeat configuration
type HeartbeatConfig struct {
	AgentID         string
	ControlPlaneURI string
}

// NewHeartbeat creates a new heartbeat manager
func NewHeartbeat(interval int, cfg *HeartbeatConfig) *Heartbeat {
	return &Heartbeat{
		intervalSeconds: interval,
		startTime:       time.Now(),
		cfg:             cfg,
	}
}

// Send sends a single heartbeat to the control plane
func (hb *Heartbeat) Send(ctx context.Context, activeJobs int) error {
	log := logging.FromContext(ctx)

	// Build heartbeat request
	req := HeartbeatRequest{
		AgentID: hb.cfg.AgentID,
		Status: AgentStatus{
			State:      hb.determineState(activeJobs),
			ActiveJobs: activeJobs,
			Uptime:     int64(time.Since(hb.startTime).Seconds()),
		},
		Resources: AgentResources{
			AvailableCPU:    0.8,    // TODO: Get real CPU availability
			AvailableMemory: 4096,   // TODO: Get real memory stats
			DiskSpace:       102400, // TODO: Get real disk space
		},
		Timestamp: time.Now().UTC(),
	}

	log.DebugContext(ctx, "sending heartbeat",
		logging.String("state", req.Status.State),
		logging.Int("active_jobs", req.Status.ActiveJobs),
		logging.Int64("uptime", req.Status.Uptime),
	)

	// TODO: Replace with real HTTP call to control plane
	// Example:
	// resp, err := r.httpClient.Post(
	//     hb.cfg.ControlPlaneURI + "/v1/agents/heartbeat",
	//     "application/json",
	//     marshal(req),
	// )

	// Simulate heartbeat (remove when control plane is ready)
	resp, err := hb.simulateHeartbeat(ctx, req)
	if err != nil {
		return fmt.Errorf("heartbeat failed: %w", err)
	}

	if !resp.Acknowledged {
		log.WarnContext(ctx, "heartbeat not acknowledged by control plane")
	} else {
		log.InfoContext(ctx, "heartbeat sent",
			logging.String("state", req.Status.State),
			logging.Int("active_jobs", req.Status.ActiveJobs),
			logging.Int64("uptime", req.Status.Uptime),
		)
	}

	// Process any commands from control plane
	if len(resp.Commands) > 0 {
		log.InfoContext(ctx, "received commands from control plane",
			logging.Int("command_count", len(resp.Commands)),
		)
		// TODO: Process commands
	}

	return nil
}

// determineState returns the current agent state based on workload
func (hb *Heartbeat) determineState(activeJobs int) string {
	if activeJobs == 0 {
		return "idle"
	}
	return "busy"
}

// simulateHeartbeat mocks the control plane response
// TODO: Remove this when control plane API is ready
func (hb *Heartbeat) simulateHeartbeat(ctx context.Context, req HeartbeatRequest) (*HeartbeatResponse, error) {
	log := logging.FromContext(ctx)

	log.DebugContext(ctx, "simulating control plane heartbeat acknowledgment",
		logging.String("state", req.Status.State),
		logging.Float64("cpu", req.Resources.AvailableCPU),
	)

	// Simulate network delay
	time.Sleep(50 * time.Millisecond)

	return &HeartbeatResponse{
		Acknowledged: true,
		Commands:     []string{}, // No commands in simulation
		Timestamp:    time.Now().UTC(),
	}, nil
}

