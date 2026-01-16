package db

import (
	"database/sql"
	"embed"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaFS embed.FS

// DB wraps the SQLite database connection
type DB struct {
	conn *sql.DB
	path string
}

// BuildStatus represents the current state of a build
type BuildStatus string

const (
	StatusQueued    BuildStatus = "queued"
	StatusRunning   BuildStatus = "running"
	StatusCompleted BuildStatus = "completed"
	StatusFailed    BuildStatus = "failed"
	StatusCancelled BuildStatus = "cancelled"
)

// Build represents a persisted build record
type Build struct {
	ID              string
	Customer        string
	ProjectName     string
	TargetImage     string
	Machine         string
	Status          BuildStatus
	ExitCode        *int
	BuildDir        string
	DeployDir       string
	LogFilePlain    string
	LogFileJSONL    string
	ConfigFile      string
	ConfigSnapshot  string
	User            string
	Host            string
	CreatedAt       time.Time
	StartedAt       *time.Time
	CompletedAt     *time.Time
	DurationSeconds *int
	Deleted         bool
	DeletedAt       *time.Time
	ErrorMessage    string
}

// BuildArtifact represents a file produced by a build
type BuildArtifact struct {
	ID           int64
	BuildID      string
	ArtifactPath string
	ArtifactType string
	SizeBytes    int64
	Checksum     string
	CreatedAt    time.Time
}

// Open opens or creates the SQLite database at the given path
func Open(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys and WAL mode for better concurrency
	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}
	if _, err := conn.Exec("PRAGMA journal_mode = WAL"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	db := &DB{
		conn: conn,
		path: dbPath,
	}

	// Run migrations
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// migrate applies the schema to the database
func (db *DB) migrate() error {
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema: %w", err)
	}

	if _, err := db.conn.Exec(string(schema)); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// Path returns the underlying database file path
func (db *DB) Path() string {
	return db.path
}

// CreateBuild inserts a new build record
func (db *DB) CreateBuild(build *Build) error {
	query := `
		INSERT INTO builds (
			id, customer, project_name, target_image, machine, status,
			build_dir, deploy_dir, log_file_plain, log_file_jsonl,
			config_file, config_snapshot, user, host, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.conn.Exec(query,
		build.ID, build.Customer, build.ProjectName, build.TargetImage, build.Machine, build.Status,
		build.BuildDir, build.DeployDir, build.LogFilePlain, build.LogFileJSONL,
		build.ConfigFile, build.ConfigSnapshot, build.User, build.Host, build.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create build: %w", err)
	}
	return nil
}

// UpdateBuildStatus updates the status of a build
func (db *DB) UpdateBuildStatus(buildID string, status BuildStatus, errorMsg string) error {
	query := `UPDATE builds SET status = ?, error_message = ? WHERE id = ?`
	_, err := db.conn.Exec(query, status, errorMsg, buildID)
	if err != nil {
		return fmt.Errorf("failed to update build status: %w", err)
	}
	return nil
}

// StartBuild marks a build as running
func (db *DB) StartBuild(buildID string) error {
	query := `UPDATE builds SET status = ?, started_at = ? WHERE id = ?`
	_, err := db.conn.Exec(query, StatusRunning, time.Now(), buildID)
	if err != nil {
		return fmt.Errorf("failed to start build: %w", err)
	}
	return nil
}

// CompleteBuild marks a build as completed or failed
func (db *DB) CompleteBuild(buildID string, status BuildStatus, exitCode int, duration time.Duration, errorMsg string) error {
	query := `
		UPDATE builds
		SET status = ?, exit_code = ?, completed_at = ?, duration_seconds = ?, error_message = ?
		WHERE id = ?
	`
	_, err := db.conn.Exec(query, status, exitCode, time.Now(), int(duration.Seconds()), errorMsg, buildID)
	if err != nil {
		return fmt.Errorf("failed to complete build: %w", err)
	}
	return nil
}

// GetBuild retrieves a build by ID
func (db *DB) GetBuild(buildID string) (*Build, error) {
	query := `
		SELECT id, customer, project_name, target_image, machine, status, exit_code,
			build_dir, deploy_dir, log_file_plain, log_file_jsonl,
			config_file, config_snapshot, user, host,
			created_at, started_at, completed_at, duration_seconds,
			deleted, deleted_at, error_message
		FROM builds WHERE id = ?
	`
	build := &Build{}
	var errorMessage sql.NullString
	var configSnapshot sql.NullString
	err := db.conn.QueryRow(query, buildID).Scan(
		&build.ID, &build.Customer, &build.ProjectName, &build.TargetImage, &build.Machine,
		&build.Status, &build.ExitCode, &build.BuildDir, &build.DeployDir,
		&build.LogFilePlain, &build.LogFileJSONL, &build.ConfigFile, &configSnapshot,
		&build.User, &build.Host, &build.CreatedAt, &build.StartedAt, &build.CompletedAt,
		&build.DurationSeconds, &build.Deleted, &build.DeletedAt, &errorMessage,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("build not found: %s", buildID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get build: %w", err)
	}
	if errorMessage.Valid {
		build.ErrorMessage = errorMessage.String
	}
	if configSnapshot.Valid {
		build.ConfigSnapshot = configSnapshot.String
	}
	return build, nil
}

// ListBuilds retrieves builds with optional filters
func (db *DB) ListBuilds(customer string, includeDeleted bool, limit int) ([]*Build, error) {
	query := `
		SELECT id, customer, project_name, target_image, machine, status, exit_code,
			build_dir, deploy_dir, log_file_plain, log_file_jsonl,
			config_file, user, host,
			created_at, started_at, completed_at, duration_seconds,
			deleted, deleted_at, error_message
		FROM builds
		WHERE 1=1
	`
	args := []interface{}{}

	if customer != "" {
		query += " AND customer = ?"
		args = append(args, customer)
	}
	if !includeDeleted {
		query += " AND deleted = 0"
	}

	query += " ORDER BY created_at DESC"
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list builds: %w", err)
	}
	defer rows.Close()

	builds := []*Build{}
	for rows.Next() {
		build := &Build{}
		var errorMessage sql.NullString
		err := rows.Scan(
			&build.ID, &build.Customer, &build.ProjectName, &build.TargetImage, &build.Machine,
			&build.Status, &build.ExitCode, &build.BuildDir, &build.DeployDir,
			&build.LogFilePlain, &build.LogFileJSONL, &build.ConfigFile,
			&build.User, &build.Host, &build.CreatedAt, &build.StartedAt, &build.CompletedAt,
			&build.DurationSeconds, &build.Deleted, &build.DeletedAt, &errorMessage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan build: %w", err)
		}
		if errorMessage.Valid {
			build.ErrorMessage = errorMessage.String
		}
		builds = append(builds, build)
	}

	return builds, nil
}

// SoftDeleteBuild marks a build as deleted
func (db *DB) SoftDeleteBuild(buildID string) error {
	query := `UPDATE builds SET deleted = 1, deleted_at = ? WHERE id = ?`
	_, err := db.conn.Exec(query, time.Now(), buildID)
	if err != nil {
		return fmt.Errorf("failed to soft delete build: %w", err)
	}
	return nil
}

// HardDeleteBuild permanently removes a build and its artifacts
func (db *DB) HardDeleteBuild(buildID string) error {
	query := `DELETE FROM builds WHERE id = ?`
	_, err := db.conn.Exec(query, buildID)
	if err != nil {
		return fmt.Errorf("failed to hard delete build: %w", err)
	}
	return nil
}

// ListStaleBuilds retrieves builds that were running when daemon stopped
func (db *DB) ListStaleBuilds() ([]*Build, error) {
	query := `
		SELECT id, customer, project_name, target_image, machine, status,
			build_dir, deploy_dir, created_at, started_at
		FROM stale_builds
	`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list stale builds: %w", err)
	}
	defer rows.Close()

	builds := []*Build{}
	for rows.Next() {
		build := &Build{}
		err := rows.Scan(
			&build.ID, &build.Customer, &build.ProjectName, &build.TargetImage, &build.Machine,
			&build.Status, &build.BuildDir, &build.DeployDir, &build.CreatedAt, &build.StartedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stale build: %w", err)
		}
		builds = append(builds, build)
	}

	return builds, nil
}

// AddArtifact records a build artifact
func (db *DB) AddArtifact(artifact *BuildArtifact) error {
	query := `
		INSERT INTO build_artifacts (build_id, artifact_path, artifact_type, size_bytes, checksum, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := db.conn.Exec(query,
		artifact.BuildID, artifact.ArtifactPath, artifact.ArtifactType,
		artifact.SizeBytes, artifact.Checksum, artifact.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to add artifact: %w", err)
	}
	id, _ := result.LastInsertId()
	artifact.ID = id
	return nil
}

// ListArtifacts retrieves all artifacts for a build
func (db *DB) ListArtifacts(buildID string) ([]*BuildArtifact, error) {
	query := `
		SELECT id, build_id, artifact_path, artifact_type, size_bytes, checksum, created_at
		FROM build_artifacts WHERE build_id = ?
		ORDER BY created_at DESC
	`
	rows, err := db.conn.Query(query, buildID)
	if err != nil {
		return nil, fmt.Errorf("failed to list artifacts: %w", err)
	}
	defer rows.Close()

	artifacts := []*BuildArtifact{}
	for rows.Next() {
		artifact := &BuildArtifact{}
		err := rows.Scan(
			&artifact.ID, &artifact.BuildID, &artifact.ArtifactPath, &artifact.ArtifactType,
			&artifact.SizeBytes, &artifact.Checksum, &artifact.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan artifact: %w", err)
		}
		artifacts = append(artifacts, artifact)
	}

	return artifacts, nil
}
