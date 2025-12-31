# Agent Responsibility Contract

## Purpose

Defines the responsibilities and design constraints of the Agent as a reliable, offline-capable edge executor. Agents must operate independently of control plane availability and ensure that job execution and logging are never blocked by external dependencies.

**Design Principle**: Agent-first. Treat the control plane as an external, unreliable dependency.

---

## Technology Stack

- **Runtime**: Single statically-compiled Go binary (no separate daemon/ctl)
- **Execution**: Local process isolation (Docker or native)
- **Logging**: In-memory ring buffer + append-only file (local durability)
- **Communication**: HTTP/REST + HMAC authentication
- **Identifiers**: job_id (logical unit of work), build_id (specific execution)

---

## Responsibilities

### Job Execution

- Accept jobs from Control Plane via pull-based REST API (`GET /api/agents/{agentId}/jobs/claim`)
- Execute jobs in isolated environments (Docker containers or native process sandboxing)
- Generate `job_id` (provided by Control Plane) and `build_id` (locally generated UUID)
- Enforce resource limits, timeouts, and execution policies
- Report job progress and completion via REST API

### Logging (Critical Path)

**Logging must never block or fail job execution.**

1. **Local Logging** (immediate, non-blocking):
   - Write all output to in-memory ring buffer (configurable size, default 10MB)
   - Write to append-only local file for durability (e.g., `/var/smidr/logs/{job_id}-{build_id}.log`)
   - Timestamp and sequence each log entry locally

2. **Log Structure** (per entry):
   - `job_id`: Identifies the logical job
   - `build_id`: Identifies this specific execution
   - `timestamp`: ISO 8601 UTC timestamp
   - `level`: DEBUG, INFO, WARNING, ERROR
   - `sequence`: Monotonic counter (per job_id/build_id) for ordering
   - `source`: "agent", "executor", "plugin", or custom
   - `message`: Log text or structured payload

3. **Batched Log Push** (eventual consistency):
   - Every N seconds (default 5s) or every M log entries (default 1000), batch logs and POST to Control Plane
   - Endpoint: `POST /api/jobs/{jobId}/logs/batch`
   - If Control Plane is unreachable: buffer locally, exponential backoff retry
   - On success: truncate local log file; preserve in-memory ring buffer

4. **Real-time Log Access** (optional, on-demand):
   - Expose local HTTP endpoint for log streaming (Server-Sent Events or HTTP streaming)
   - Frontend or CLI can request `GET /api/logs/{jobId}/{buildId}?follow=true` to stream live logs from agent
   - If agent is offline or unreachable: fall back to persisted logs in Control Plane

### Offline Operation

- Agent must operate and execute jobs **even if Control Plane is unreachable**
- Cache job definitions locally after claiming
- Write logs to local append-only file immediately
- Buffer job completion reports and metrics locally
- Retry pushing logs and completion status with exponential backoff when Control Plane becomes available
- Never fail a job due to Control Plane unavailability

### Status Reporting

- Send periodic heartbeats to Control Plane (`POST /api/agents/{agentId}/heartbeat`)
- Include current state (Idle, Busy, Offline), job count, system metrics (CPU, memory, disk)
- Report job completion with exit code, error message, artifacts uploaded (`POST /api/jobs/{jobId}/complete`)
- If Control Plane unreachable: continue executing jobs, queue completion reports locally

### Artifact Handling

- Execute build systems that produce artifacts (Yocto, live-build, Docker, etc.)
- Upload artifacts to S3 directly (agent obtains presigned URL from Control Plane during job claim or retries with local cache)
- Report artifact metadata to Control Plane (name, type, size, checksum)
- If artifact upload fails: retry locally, buffer metadata, resume on Control Plane availability

### Dependency Management

- Single Go binary with minimal external dependencies
- Plugins run as separate processes (gRPC-based for isolation)
- Plugins inherit agent's offline-first logging model

---

## Non-Responsibilities

- **Control Plane logic**: Scheduling, policy evaluation, tenant isolation
- **Persistent storage**: Control Plane owns all metadata and state
- **Cross-agent coordination**: No peer-to-peer communication
- **Real-time updates**: Eventual consistency is acceptable for logs and status

---

## Offline & Eventual Consistency Guarantees

### Agent Goes Offline

1. Job execution continues normally
2. Logs are written to local file
3. Job completion is buffered locally
4. Artifacts are cached locally (or upload is retried on reconnection)

### Agent Comes Back Online

1. Resume pending job completion reports
2. Resume buffered logs in batches
3. Resume artifact uploads
4. Control Plane deduplicates using `sequence` numbers in logs and idempotent job completion

### Control Plane Down (Agent Online)

1. Agent continues executing queued jobs
2. Logs are buffered locally
3. Agents retry Control Plane endpoints with exponential backoff (up to 24h default)
4. No job execution is blocked or delayed

### Control Plane Back Online

1. Agents flush buffered logs, completions, and metrics
2. Control Plane accepts and deduplicates based on `job_id`, `build_id`, and `sequence`
3. State is eventually consistent within minutes

---

## Identifiers & Contracts

**job_id**: Assigned by Control Plane during job claim
- Format: UUID
- Used across: job definition, all log entries, status reports, artifact references
- Never reused during agent's lifetime

**build_id**: Generated by agent for each execution
- Format: UUID (locally generated)
- Allows agent to retry a failed job with a new build_id
- Used in logs, artifacts, status reports
- Enables audit trail: same job_id, different build_ids = retries

**Consistency**: All APIs (logs, status, artifacts) reference `job_id` and `build_id` together

---

## API Contract with Control Plane

### Pull-Based Job Claiming

```
POST /api/agents/{agentId}/jobs/claim
Headers: X-Tenant-Id, X-Agent-Id, X-Timestamp, X-Signature
Response: { jobs: [{ job_id, job_type, parameters, deadline_utc, ... }] }
```

### Heartbeat

```
POST /api/agents/{agentId}/heartbeat
Body: { state: "Idle|Busy|Offline", current_job_count, metrics: { cpu, memory, disk } }
Response: { alive: true, jobs_to_cancel?: [...] }
```

### Log Batch Upload

```
POST /api/jobs/{jobId}/logs/batch
Body: { job_id, build_id, logs: [{ timestamp, level, sequence, message, ... }] }
Response: { acknowledged: true }
```

### Job Completion

```
POST /api/jobs/{jobId}/complete
Body: { job_id, build_id, final_state: "Succeeded|Failed|Cancelled|TimedOut", exit_code, error_message, artifacts: [...] }
Response: { acknowledged: true }
```

---

## Design Constraints (Immutable)

1. **Agent must function offline**: No hard dependency on Control Plane for execution
2. **Single executable**: No daemon/ctl split; one binary per platform
3. **Logging durability**: All logs persisted locally before and after upload
4. **Logging non-blocking**: Job execution never waits for log upload or Control Plane availability
5. **Eventual consistency**: Logs, status, and artifacts are eventually consistent, not real-time
6. **Append-only logs**: No modifications, only additions; sequence numbers for ordering
7. **HMAC authentication**: No PKI; secrets managed in AWS Secrets Manager
8. **Idempotent APIs**: All Control Plane endpoints idempotent to support retries

---

## Deployment & Configuration

### Agent Binary

- Built per platform: `smidr-agent-linux-amd64`, `smidr-agent-darwin-arm64`, `smidr-agent-windows-amd64.exe`
- Statically linked; no runtime dependencies except Docker (optional)
- Single config file: `smidr.yaml` or environment variables

### Configuration Example

```yaml
agent:
  id: <agent-uuid>  # Registered on first run
  name: "builder-01"
  version: "1.0.0"

control_plane:
  url: "https://api.smidr.example.com"
  tenant_id: "<tenant-uuid>"
  hmac_secret: "<base64-encoded-secret>"  # from AWS Secrets Manager

logging:
  local_dir: "/var/smidr/logs"
  ring_buffer_mb: 10
  batch_interval_sec: 5
  batch_size_entries: 1000
  retry_max_hours: 24

execution:
  runtime: "docker|native"
  max_concurrent_jobs: 4
  timeout_sec: 3600

plugins:
  enabled: true
  dir: "/opt/smidr/plugins"
```

---

## Future Extensions

Without breaking existing contracts:

- Support agent-local job definitions (e.g., cron jobs)
- Add agent-to-agent artifact transfer (mesh networking)
- Support agent clustering and load balancing
- Add advanced scheduling policies (affinity, anti-affinity)
- Extend logging with structured fields and custom levels

---

## Related

- [Control Plane Responsibility Contract](control-plane.md)
- [Job Lifecycle Specification](job-lifecycle.md)
- [Logging Architecture](logging.md) (to be created)
- [Offline Design Patterns](offline-patterns.md) (to be created)
