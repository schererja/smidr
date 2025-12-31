package job

import "time"

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

// ResourceLimits defines maximum resources for a job
type ResourceLimits struct {
	MaxCPU    float64 `json:"max_cpu"`
	MaxMemory int64   `json:"max_memory_mb"`
	MaxDisk   int64   `json:"max_disk_mb"`
}
