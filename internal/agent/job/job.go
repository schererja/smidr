package job

import (
	"encoding/json"
	"time"
)

// Job represents a job assigned to the agent
type Job struct {
	ID              string                 `json:"id"`
	TenantID        string                 `json:"tenant_id"`
	ProjectID       string                 `json:"project_id"`
	PluginType      string                 `json:"plugin_type"`
	PluginVersion   string                 `json:"plugin_version"`
	Payload         map[string]interface{} `json:"payload"`
	ResourceLimits  ResourceLimits         `json:"resource_limits"`
	TimeoutSeconds  int                    `json:"timeout_seconds"`
	SubmissionTime  time.Time              `json:"submission_time"`
	SubmittedByUser string                 `json:"submitted_by_user"`
}

// UnmarshalRequest unmarshals the job payload into a typed request struct
func (j *Job) UnmarshalRequest(req interface{}) error {
	if j.Payload == nil {
		return nil
	}
	// Marshal to JSON and back to get proper type conversion
	data, err := json.Marshal(j.Payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, req)
}

// ResourceLimits defines maximum resources for a job
type ResourceLimits struct {
	MaxCPU    float64 `json:"max_cpu"`
	MaxMemory int64   `json:"max_memory_mb"`
	MaxDisk   int64   `json:"max_disk_mb"`
}

// JobResponse represents the response from a job execution
type JobResponse struct {
	Status       string            `json:"status"`        // completed, failed, cancelled
	Message      string            `json:"message"`       // Human-readable message
	Artifacts    []string          `json:"artifacts"`     // Paths or URLs to generated artifacts
	Metadata     map[string]string `json:"metadata"`      // Additional metadata (e.g., build time, commit hash)
	ErrorMessage string            `json:"error_message"` // Error details if failed
}
