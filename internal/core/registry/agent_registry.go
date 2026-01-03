package registry

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AgentInfo holds information about a registered agent
type AgentInfo struct {
	ID           string
	Name         string
	RegisteredAt time.Time
	LastSeen     time.Time
	Status       AgentStatus
	Capabilities []string
	Metadata     map[string]string
}

// AgentStatus represents the current state of an agent
type AgentStatus string

const (
	AgentStatusOnline  AgentStatus = "online"
	AgentStatusOffline AgentStatus = "offline"
	AgentStatusIdle    AgentStatus = "idle"
	AgentStatusBusy    AgentStatus = "busy"
)

// MemoryStore is an in-memory implementation of the Store interface
type MemoryStore struct {
	mu     sync.RWMutex
	agents map[string]*AgentInfo
}

// NewMemoryStore creates a new in-memory agent store
func NewMemoryStore() Store {
	return &MemoryStore{
		agents: make(map[string]*AgentInfo),
	}
}

// Register adds or updates an agent in the store
func (m *MemoryStore) Register(ctx context.Context, id, name string, capabilities []string, metadata map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if id == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	now := time.Now()
	if existing, ok := m.agents[id]; ok {
		// Update existing agent
		existing.Name = name
		existing.LastSeen = now
		existing.Status = AgentStatusOnline
		existing.Capabilities = capabilities
		existing.Metadata = metadata
	} else {
		// Register new agent
		m.agents[id] = &AgentInfo{
			ID:           id,
			Name:         name,
			RegisteredAt: now,
			LastSeen:     now,
			Status:       AgentStatusOnline,
			Capabilities: capabilities,
			Metadata:     metadata,
		}
	}

	return nil
}

// Heartbeat updates the last seen time for an agent
func (m *MemoryStore) Heartbeat(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	agent, ok := m.agents[id]
	if !ok {
		return fmt.Errorf("agent %s not found", id)
	}

	agent.LastSeen = time.Now()
	agent.Status = AgentStatusOnline
	return nil
}

// UpdateStatus changes the status of an agent
func (m *MemoryStore) UpdateStatus(ctx context.Context, id string, status AgentStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	agent, ok := m.agents[id]
	if !ok {
		return fmt.Errorf("agent %s not found", id)
	}

	agent.Status = status
	agent.LastSeen = time.Now()
	return nil
}

// Get retrieves an agent by ID
func (m *MemoryStore) Get(ctx context.Context, id string) (*AgentInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agent, ok := m.agents[id]
	if !ok {
		return nil, fmt.Errorf("agent %s not found", id)
	}

	// Return a copy to avoid external modification
	agentCopy := *agent
	return &agentCopy, nil
}

// List returns all registered agents
func (m *MemoryStore) List(ctx context.Context) []*AgentInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agents := make([]*AgentInfo, 0, len(m.agents))
	for _, agent := range m.agents {
		agentCopy := *agent
		agents = append(agents, &agentCopy)
	}

	return agents
}

// Remove deletes an agent from the store
func (m *MemoryStore) Remove(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.agents[id]; !ok {
		return fmt.Errorf("agent %s not found", id)
	}

	delete(m.agents, id)
	return nil
}

// Count returns the number of registered agents
func (m *MemoryStore) Count(ctx context.Context) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.agents)
}

// Close closes any resources held by the store
func (m *MemoryStore) Close(ctx context.Context) error {
	// No resources to close for in-memory store
	return nil
}
