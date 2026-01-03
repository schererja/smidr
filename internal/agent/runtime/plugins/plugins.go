package plugins

import (
	"context"

	"github.com/intrik8-labs/smidr/internal/agent/job"
)

// Plugin defines the interface for job execution plugins
type Plugin[Req any, Res any] interface {
	Execute(ctx context.Context, j job.Job, req Req) (*job.JobResponse, error)
}

// YoctoBuildRequest defines the request payload for Yocto builds
type YoctoBuildRequest struct {
	// Config delivery methods (only one should be provided)
	ConfigURL        string `json:"config_url,omitempty"`        // URL to fetch config from (e.g., GitHub raw URL)
	ConfigCompressed string `json:"config_compressed,omitempty"` // Base64-encoded gzipped config
	ConfigRaw        string `json:"config,omitempty"`            // Raw YAML config as string

	// Build options
	CacheEnabled bool              `json:"cache_enabled"`            // Enable shared state cache
	CleanBuild   bool              `json:"clean_build,omitempty"`    // Force clean build (no cache)
	ExtraEnvVars map[string]string `json:"extra_env_vars,omitempty"` // Additional environment variables
}

// BuildRequest defines the request payload for generic builds
type BuildRequest struct {
	Command      string            `json:"command"`                 // Build command to execute
	WorkDir      string            `json:"workdir,omitempty"`       // Working directory (default: current dir)
	EnvVars      map[string]string `json:"env_vars,omitempty"`      // Environment variables
	Artifacts    []string          `json:"artifacts,omitempty"`     // Artifact patterns to collect
	CacheEnabled bool              `json:"cache_enabled,omitempty"` // Enable build cache
}

// MonitorRequest defines the request payload for monitoring jobs
type MonitorRequest struct {
	Target   string `json:"target"`             // Target to monitor (e.g., "system", "service:nginx")
	Duration string `json:"duration,omitempty"` // Duration to monitor (e.g., "30s", "5m")
	Interval string `json:"interval,omitempty"` // Check interval (e.g., "5s")
}

// CommandRequest defines the request payload for shell command execution
type CommandRequest struct {
	Command       string            `json:"command"`                  // Command to execute
	Args          []string          `json:"args,omitempty"`           // Command arguments
	WorkDir       string            `json:"workdir,omitempty"`        // Working directory (default: current dir)
	EnvVars       map[string]string `json:"env_vars,omitempty"`       // Environment variables
	Shell         string            `json:"shell,omitempty"`          // Shell to use (default: /bin/sh)
	Timeout       int               `json:"timeout,omitempty"`        // Timeout in seconds
	CaptureOutput bool              `json:"capture_output,omitempty"` // Capture stdout/stderr
}
