package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/intrik8-labs/smidr/internal/agent/client"
	"github.com/intrik8-labs/smidr/internal/agent/job"
	"github.com/intrik8-labs/smidr/internal/logging"
)

// ResourceLimits defines maximum resources for a job
type ResourceLimits struct {
	MaxCPU    float64 `json:"max_cpu"`
	MaxMemory int64   `json:"max_memory_mb"`
	MaxDisk   int64   `json:"max_disk_mb"`
}

// PollRequest represents the agent's poll for new jobs
type PollRequest struct {
	AgentID           string            `json:"agent_id"`
	AvailableCapacity int               `json:"available_capacity"`
	SupportedPlugins  []string          `json:"supported_plugins"`
	AgentCapabilities AgentCapabilities `json:"agent_capabilities"`
}

// PollResponse contains jobs available for this agent
type PollResponse struct {
	Jobs      []job.Job `json:"jobs"`
	Timestamp time.Time `json:"timestamp"`
}

// Poller manages job polling from the control plane
type Poller struct {
	agentID             string
	controlPlaneURI     string
	pollIntervalSeconds int
	maxConcurrentJobs   int
	supportedPlugins    []string
	jobQueue            []job.Job // Simulated job queue
	demoMode            bool
	client              *client.Client
}

// NewPoller creates a new job poller
func NewPoller(agentID, controlPlaneURI string, pollInterval, maxJobs int, supportedPlugins []string, demoMode bool) *Poller {
	return &Poller{
		agentID:             agentID,
		controlPlaneURI:     controlPlaneURI,
		pollIntervalSeconds: pollInterval,
		maxConcurrentJobs:   maxJobs,
		supportedPlugins:    supportedPlugins,
		jobQueue:            []job.Job{}, // Start with empty queue
		demoMode:            demoMode,
		client:              client.NewClient(controlPlaneURI),
	}
}

// Poll requests available jobs from the control plane
func (p *Poller) Poll(ctx context.Context, activeJobs int) ([]job.Job, error) {
	log := logging.FromContext(ctx)

	// Calculate available capacity
	availableCapacity := p.maxConcurrentJobs - activeJobs
	if availableCapacity <= 0 {
		log.DebugContext(ctx, "no available job slots",
			logging.Int("active_jobs", activeJobs),
			logging.Int("max_concurrent", p.maxConcurrentJobs),
		)
		return []job.Job{}, nil
	}

	req := PollRequest{
		AgentID:           p.agentID,
		AvailableCapacity: availableCapacity,
		SupportedPlugins:  p.supportedPlugins,
		AgentCapabilities: AgentCapabilities{
			Architecture: "amd64",
			OS:           "linux",
			MaxJobs:      p.maxConcurrentJobs,
			Features: []string{
				"docker",
				"build",
			},
		},
	}

	log.DebugContext(ctx, "polling for jobs",
		logging.Int("available_capacity", availableCapacity),
		logging.Int("supported_plugins", len(req.SupportedPlugins)),
		logging.Bool("demo_mode", p.demoMode),
	)

	if p.demoMode {
		// Simulate poll in demo mode
		resp, err := p.simulatePoll(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("poll failed: %w", err)
		}

		if len(resp.Jobs) > 0 {
			log.InfoContext(ctx, "received jobs from control plane (demo mode)",
				logging.Int("job_count", len(resp.Jobs)),
			)
		}

		return resp.Jobs, nil
	}

	// Real poll from control plane
	jobs, err := p.client.PollJobs(ctx, p.agentID)
	if err != nil {
		return nil, fmt.Errorf("poll failed: %w", err)
	}

	if len(jobs) > 0 {
		log.InfoContext(ctx, "received jobs from control plane",
			logging.Int("job_count", len(jobs)),
		)
	}

	// Convert interface{} jobs to job.Job (placeholder)
	// TODO: Implement proper unmarshaling of job objects
	return []job.Job{}, nil
}

// simulatePoll mocks the control plane response
// TODO: Remove this when control plane API is ready
func (p *Poller) simulatePoll(ctx context.Context, req PollRequest) (*PollResponse, error) {
	log := logging.FromContext(ctx)

	log.DebugContext(ctx, "simulating control plane poll",
		logging.Int("available_capacity", req.AvailableCapacity),
		logging.String("architecture", req.AgentCapabilities.Architecture),
	)

	// Simulate network delay
	time.Sleep(100 * time.Millisecond)

	// Occasionally add a new job to simulate jobs coming in
	if len(p.jobQueue) < 5 { // Limit queue size for simulation
		// Simple random: add if a random number is even
		if time.Now().UnixNano()%2 == 0 {
			pluginType := "yocto" // Force yocto for testing
			payload := map[string]interface{}{
				"config": `# Smidr Build Configuration
name: my-embedded-project
base:
  provider: toradex
  machine: verdin-imx8mp
build:
  image: core-image-weston
`,
			}
			newJob := job.Job{
				ID:              fmt.Sprintf("job-%d", time.Now().UnixNano()),
				TenantID:        "tenant-abc",
				ProjectID:       "project-xyz",
				PluginType:      pluginType,
				PluginVersion:   "1.0.0",
				Payload:         payload,
				ResourceLimits:  job.ResourceLimits{MaxCPU: 2.0, MaxMemory: 1024, MaxDisk: 2048},
				TimeoutSeconds:  30, // Shorter for simulation
				SubmissionTime:  time.Now().UTC(),
				SubmittedByUser: "user@example.com",
			}
			p.jobQueue = append(p.jobQueue, newJob)
			log.DebugContext(ctx, "added simulated job to queue", logging.String("job_id", newJob.ID), logging.String("plugin_type", pluginType))
		}
	}

	// Return up to availableCapacity jobs from the queue (without removing)
	jobsToReturn := []job.Job{}
	for i := 0; i < req.AvailableCapacity && i < len(p.jobQueue); i++ {
		jobsToReturn = append(jobsToReturn, p.jobQueue[i])
	}

	return &PollResponse{
		Jobs:      jobsToReturn,
		Timestamp: time.Now().UTC(),
	}, nil
}

// ClaimJob informs the control plane that this agent is executing a job
func (p *Poller) ClaimJob(ctx context.Context, jobID string) error {
	log := logging.FromContext(ctx)

	log.InfoContext(ctx, "claiming job",
		logging.String("job_id", jobID),
	)

	// Remove job from queue
	for i, job := range p.jobQueue {
		if job.ID == jobID {
			p.jobQueue = append(p.jobQueue[:i], p.jobQueue[i+1:]...)
			break
		}
	}

	// TODO: Send claim to control plane
	// POST /v1/jobs/{jobID}/claim with agent_id

	// Simulate claim acknowledgment
	time.Sleep(50 * time.Millisecond)

	log.InfoContext(ctx, "job claimed",
		logging.String("job_id", jobID),
	)

	return nil
}

// ReportJobCompletion informs the control plane that a job finished
func (p *Poller) ReportJobCompletion(ctx context.Context, jobID string, status string, logs string, artifacts []string) error {
	log := logging.FromContext(ctx)

	log.InfoContext(ctx, "reporting job completion",
		logging.String("job_id", jobID),
		logging.String("status", status),
		logging.Int("artifact_count", len(artifacts)),
	)

	// TODO: Send completion to control plane
	// POST /v1/jobs/{jobID}/complete with status, logs, artifacts

	// Simulate completion acknowledgment
	time.Sleep(100 * time.Millisecond)

	log.InfoContext(ctx, "job completion reported",
		logging.String("job_id", jobID),
		logging.String("status", status),
	)

	return nil
}
