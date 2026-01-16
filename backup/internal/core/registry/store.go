package registry

import "context"

// Store defines the interface for agent storage backends
type Store interface {
	// Register adds or updates an agent
	Register(ctx context.Context, id, name string, capabilities []string, metadata map[string]string) error

	// Heartbeat updates the last seen time for an agent
	Heartbeat(ctx context.Context, id string) error

	// UpdateStatus changes the status of an agent
	UpdateStatus(ctx context.Context, id string, status AgentStatus) error

	// Get retrieves an agent by ID
	Get(ctx context.Context, id string) (*AgentInfo, error)

	// List returns all registered agents
	List(ctx context.Context) []*AgentInfo

	// Remove deletes an agent from the registry
	Remove(ctx context.Context, id string) error

	// Count returns the number of registered agents
	Count(ctx context.Context) int

	// Close closes any resources held by the store
	Close(ctx context.Context) error
}
