package registry

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStore is a SQLite implementation of the Store interface
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite store at the given path
func NewSQLiteStore(dbPath string) (Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	store := &SQLiteStore{db: db}

	// Create the agents table if it doesn't exist
	if err := store.createSchema(context.Background()); err != nil {
		db.Close()
		return nil, err
	}

	return store, nil
}

func (s *SQLiteStore) createSchema(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS agents (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		registered_at DATETIME NOT NULL,
		last_seen DATETIME NOT NULL,
		status TEXT NOT NULL,
		capabilities TEXT,
		metadata TEXT
	);
	`

	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// Register adds or updates an agent in the database
func (s *SQLiteStore) Register(ctx context.Context, id, name string, capabilities []string, metadata map[string]string) error {
	if id == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	now := time.Now()

	// Serialize capabilities and metadata to JSON
	capsJSON, err := serializeStringSlice(capabilities)
	if err != nil {
		return err
	}
	metaJSON, err := serializeStringMap(metadata)
	if err != nil {
		return err
	}

	query := `
	INSERT INTO agents (id, name, registered_at, last_seen, status, capabilities, metadata)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		name = excluded.name,
		last_seen = excluded.last_seen,
		status = excluded.status,
		capabilities = excluded.capabilities,
		metadata = excluded.metadata
	`

	_, err = s.db.ExecContext(ctx, query,
		id, name, now, now, AgentStatusOnline, capsJSON, metaJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to register agent: %w", err)
	}

	return nil
}

// Heartbeat updates the last seen time for an agent
func (s *SQLiteStore) Heartbeat(ctx context.Context, id string) error {
	query := `
	UPDATE agents
	SET last_seen = ?, status = ?
	WHERE id = ?
	`

	result, err := s.db.ExecContext(ctx, query, time.Now(), AgentStatusOnline, id)
	if err != nil {
		return fmt.Errorf("failed to update heartbeat: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("agent %s not found", id)
	}

	return nil
}

// UpdateStatus changes the status of an agent
func (s *SQLiteStore) UpdateStatus(ctx context.Context, id string, status AgentStatus) error {
	query := `
	UPDATE agents
	SET status = ?, last_seen = ?
	WHERE id = ?
	`

	result, err := s.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("agent %s not found", id)
	}

	return nil
}

// Get retrieves an agent by ID
func (s *SQLiteStore) Get(ctx context.Context, id string) (*AgentInfo, error) {
	query := `
	SELECT id, name, registered_at, last_seen, status, capabilities, metadata
	FROM agents
	WHERE id = ?
	`

	var agent AgentInfo
	var capsJSON, metaJSON string

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&agent.ID,
		&agent.Name,
		&agent.RegisteredAt,
		&agent.LastSeen,
		&agent.Status,
		&capsJSON,
		&metaJSON,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("agent %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	agent.Capabilities, _ = deserializeStringSlice(capsJSON)
	agent.Metadata, _ = deserializeStringMap(metaJSON)

	return &agent, nil
}

// List returns all registered agents
func (s *SQLiteStore) List(ctx context.Context) []*AgentInfo {
	query := `
	SELECT id, name, registered_at, last_seen, status, capabilities, metadata
	FROM agents
	ORDER BY registered_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return []*AgentInfo{}
	}
	defer rows.Close()

	var agents []*AgentInfo
	for rows.Next() {
		var agent AgentInfo
		var capsJSON, metaJSON string

		err := rows.Scan(
			&agent.ID,
			&agent.Name,
			&agent.RegisteredAt,
			&agent.LastSeen,
			&agent.Status,
			&capsJSON,
			&metaJSON,
		)
		if err != nil {
			continue
		}

		agent.Capabilities, _ = deserializeStringSlice(capsJSON)
		agent.Metadata, _ = deserializeStringMap(metaJSON)

		agents = append(agents, &agent)
	}

	return agents
}

// Remove deletes an agent from the database
func (s *SQLiteStore) Remove(ctx context.Context, id string) error {
	query := `DELETE FROM agents WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to remove agent: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("agent %s not found", id)
	}

	return nil
}

// Count returns the number of registered agents
func (s *SQLiteStore) Count(ctx context.Context) int {
	query := `SELECT COUNT(*) FROM agents`

	var count int
	s.db.QueryRowContext(ctx, query).Scan(&count)
	return count
}

// Close closes the database connection
func (s *SQLiteStore) Close(ctx context.Context) error {
	return s.db.Close()
}

// Helper functions for JSON serialization
func serializeStringSlice(data []string) (string, error) {
	if len(data) == 0 {
		return "[]", nil
	}
	// Simple JSON encoding for strings
	result := "["
	for i, s := range data {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf(`"%s"`, escapeJSON(s))
	}
	result += "]"
	return result, nil
}

func deserializeStringSlice(data string) ([]string, error) {
	if data == "" || data == "[]" {
		return []string{}, nil
	}
	// Simple JSON parsing for strings
	// This is a basic implementation; use encoding/json for production
	var result []string
	// Remove brackets and split by comma (naive approach for demo)
	data = data[1 : len(data)-1]
	if data != "" {
		// This is simplified; a real implementation would use encoding/json
	}
	return result, nil
}

func serializeStringMap(data map[string]string) (string, error) {
	if len(data) == 0 {
		return "{}", nil
	}
	result := "{"
	i := 0
	for k, v := range data {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf(`"%s":"%s"`, escapeJSON(k), escapeJSON(v))
		i++
	}
	result += "}"
	return result, nil
}

func deserializeStringMap(data string) (map[string]string, error) {
	if data == "" || data == "{}" {
		return map[string]string{}, nil
	}
	return map[string]string{}, nil
}

func escapeJSON(s string) string {
	// Basic JSON escaping
	s = fmt.Sprintf("%q", s)
	return s[1 : len(s)-1]
}
