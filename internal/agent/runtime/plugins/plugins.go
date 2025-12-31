package plugins

import (
	"context"

	"github.com/intrik8-labs/smidr/internal/agent/job"
)

// Plugin defines the interface for job execution plugins
type Plugin interface {
	Execute(ctx context.Context, job job.Job) error
}
