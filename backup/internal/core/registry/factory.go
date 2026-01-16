package registry

import (
	"context"
	"fmt"
)

// Config holds store configuration options
type Config struct {
	Type       string // "memory", "sqlite"
	SQLitePath string // Path for SQLite database
}

// NewStore creates a new Store based on configuration
func NewStore(ctx context.Context, cfg Config) (Store, error) {
	switch cfg.Type {
	case "memory":
		return NewMemoryStore(), nil
	case "sqlite":
		if cfg.SQLitePath == "" {
			cfg.SQLitePath = "smidr.db"
		}
		return NewSQLiteStore(cfg.SQLitePath)
	default:
		return nil, fmt.Errorf("unknown store type: %s", cfg.Type)
	}
}
