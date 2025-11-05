-- Build state persistence schema for smidr
-- Tracks all builds with soft-delete support and state recovery

CREATE TABLE IF NOT EXISTS builds (
    id TEXT PRIMARY KEY,                    -- UUID build ID
    customer TEXT NOT NULL,                 -- Customer/tenant identifier
    project_name TEXT NOT NULL,             -- Project name from config
    target_image TEXT NOT NULL,             -- Image being built (e.g., core-image-minimal)
    machine TEXT NOT NULL,                  -- Target machine (e.g., qemux86-64)
    status TEXT NOT NULL,                   -- queued, running, completed, failed, cancelled
    exit_code INTEGER,                      -- BitBake exit code (NULL if not completed)

    -- Directories
    build_dir TEXT NOT NULL,                -- Absolute path to build workspace
    deploy_dir TEXT NOT NULL,               -- Deploy artifacts location
    log_file_plain TEXT,                    -- Path to plain text log file
    log_file_jsonl TEXT,                    -- Path to JSONL log file

    -- Metadata
    config_file TEXT,                       -- Path to smidr.yaml used
    config_snapshot TEXT,                   -- JSON snapshot of config at build time
    user TEXT,                              -- Username who initiated build
    host TEXT,                              -- Hostname where build ran

    -- Timing
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at DATETIME,                    -- When build execution started
    completed_at DATETIME,                  -- When build finished (success or failure)
    duration_seconds INTEGER,               -- Total build duration

    -- Soft delete
    deleted BOOLEAN NOT NULL DEFAULT 0,
    deleted_at DATETIME,

    -- Error tracking
    error_message TEXT                      -- Human-readable error if failed
);

-- Indexes for builds table
CREATE INDEX IF NOT EXISTS idx_builds_customer ON builds(customer);
CREATE INDEX IF NOT EXISTS idx_builds_status ON builds(status);
CREATE INDEX IF NOT EXISTS idx_builds_created_at ON builds(created_at);
CREATE INDEX IF NOT EXISTS idx_builds_deleted ON builds(deleted);
CREATE TABLE IF NOT EXISTS build_artifacts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    build_id TEXT NOT NULL,                 -- FK to builds.id
    artifact_path TEXT NOT NULL,            -- Relative path within deploy_dir
    artifact_type TEXT,                     -- wic, tar.gz, manifest, etc.
    size_bytes INTEGER,
    checksum TEXT,                          -- SHA256 or similar
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (build_id) REFERENCES builds(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_artifacts_build_id ON build_artifacts(build_id);CREATE TABLE IF NOT EXISTS build_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    build_id TEXT NOT NULL,                 -- FK to builds.id
    metric_name TEXT NOT NULL,              -- e.g., "tasks_completed", "cache_hit_rate"
    metric_value REAL NOT NULL,
    recorded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (build_id) REFERENCES builds(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_metrics_build_id ON build_metrics(build_id);-- View for active (non-deleted) builds
CREATE VIEW IF NOT EXISTS active_builds AS
SELECT * FROM builds WHERE deleted = 0;

-- View for builds that need recovery (were running when daemon stopped)
CREATE VIEW IF NOT EXISTS stale_builds AS
SELECT * FROM builds
WHERE status IN ('queued', 'running')
  AND deleted = 0
  AND datetime(created_at, '+24 hours') > datetime('now');
