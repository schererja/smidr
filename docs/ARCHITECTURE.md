# Smidr Architecture Overview

## System Purpose

Smidr provides **quiet, continuous assurance** for Linux x86_64 systems by answering one core question:

> **"Is this system behaving normally right now?"**

The platform learns baseline behavior for each system, detects meaningful deviations, and surfaces issues with context—not noise.

---

## Core Components

```
┌──────────────────────────────────────────────────────────────────┐
│                          Smidr Platform                          │
└──────────────────────────────────────────────────────────────────┘

┌─────────────────┐         ┌─────────────────┐         ┌──────────┐
│  Linux Systems  │         │ Control Plane   │         │    UI    │
│                 │         │                 │         │          │
│  ┌───────────┐  │         │  ┌──────────┐   │         │  React   │
│  │   Agent   │──┼────────▶│  │   API    │◀──┼─────────│   Web    │
│  │    (Go)   │  │  HTTPS  │  │  (C#)    │   │  HTTPS  │   App    │
│  └───────────┘  │  mTLS   │  └──────────┘   │         │          │
│                 │         │       │          │         └──────────┘
│  Collects:      │         │  ┌────▼─────┐   │
│  • uptime       │         │  │   CA     │   │
│  • load         │         │  │ Service  │   │
│  • memory       │         │  └──────────┘   │
│  • disk         │         │       │          │
│  • processes    │         │  ┌────▼─────┐   │
│                 │         │  │ Postgres │   │
└─────────────────┘         │  │ Database │   │
                            │  └──────────┘   │
                            └─────────────────┘
```

### 1. Agent (Go)

**Location:** `/agent`

Lightweight daemon that runs on each monitored Linux system.

**Responsibilities:**
- Collect system signals from `/proc` and system calls
- Generate UUID and cryptographic keypair on first start
- Enroll with control plane via CSR submission
- Send heartbeat payloads every 60 seconds using mTLS
- Manage local certificate and configuration

**Signal Collection:**
- `uptimeSeconds` - from `/proc/uptime`
- `loadAverage1m` - from `/proc/loadavg`
- `memoryUsedPct` - computed from `/proc/meminfo`
- `diskUsedPct` - from `statfs(/)` syscall
- `processCount` - count of entries in `/proc`

**Deployment:**
- Runs as systemd service (`smidr-agent.service`)
- Unprivileged `smidr` user
- Read-only access to `/proc`, write access to `/etc/smidr` only
- Security hardening via systemd (NoNewPrivileges, PrivateTmp, ProtectSystem)

### 2. Control Plane (C# / ASP.NET Core)

**Location:** `/control-plane`

Central API that orchestrates enrollment, telemetry ingestion, and health evaluation.

**Responsibilities:**
- Operate internal Certificate Authority for agent enrollment
- Validate agent CSRs and issue signed certificates
- Receive and store heartbeat signals with mTLS authentication
- Learn baseline behavior over 24-hour window
- Evaluate health states asynchronously after each heartbeat
- Expose REST API for UI and integrations

**Key Services:**
- **CaService:** Manages internal CA, signs agent certificates
- **MtlsValidationService:** Validates client certificates against CA
- **HealthEvaluationService:** Computes baselines and detects anomalies
- **MtlsAuthenticationMiddleware:** Enforces certificate authentication on protected endpoints

**Storage:**
- SQLite (default, development)
- PostgreSQL (production)

### 3. Certificate Authority (Embedded)

**Location:** `control-plane/Services/CaService.cs`

Self-signed internal CA for agent certificate issuance.

**Characteristics:**
- Auto-generated on first control plane startup
- 2048-bit RSA keypair
- Stored at `control-plane/data/ca/`
- CA certificate distributed to agents for verification
- Signs agent certificates with 365-day validity

**Security:**
- CA private key never leaves control plane
- Agent private keys never leave agent systems
- TLS 1.2+ required for all communication

### 4. UI (React)

**Location:** `/ui`

Minimal web interface for viewing system health.

**Features:**
- System list with health badges and latest signals
- System detail view with baseline comparisons
- Auto-refresh every 30 seconds
- Color-coded health states

**Tech Stack:**
- React 19 + TypeScript
- Vite (dev server and build)
- Tailwind CSS + shadcn/ui components
- Axios for API calls

---

## Data Flow

### 1. Agent Enrollment (First Start)

```
Agent                           Control Plane
  │                                   │
  │  1. Generate UUID + keypair       │
  │────────────────────────────────▶  │
  │                                   │
  │  2. Generate CSR                  │
  │────────────────────────────────▶  │
  │                                   │
  │  3. POST /agents/register         │
  │     (CSR, UUID, hostname)         │
  ├──────────────────────────────────▶│
  │                                   │
  │                                   │  4. Validate CSR
  │                                   │  5. Sign with CA
  │                                   │  6. Store agent record
  │                                   │
  │  7. Return signed certificate     │
  │◀──────────────────────────────────┤
  │                                   │
  │  8. Persist certificate           │
  │────────────────────────────────▶  │
  │                                   │
  │  Ready for mTLS heartbeats        │
```

### 2. Heartbeat Loop (Normal Operation)

```
Agent                           Control Plane
  │                                   │
  │  1. Collect system signals        │
  │     (uptime, load, memory, etc.)  │
  │────────────────────────────────▶  │
  │                                   │
  │  2. POST /v0/agents/heartbeat     │
  │     (mTLS authenticated)          │
  ├──────────────────────────────────▶│
  │                                   │
  │                                   │  3. Validate mTLS cert
  │                                   │  4. Store heartbeat
  │                                   │  5. Update last_heartbeat_at
  │                                   │
  │  6. Return 200 OK                 │
  │◀──────────────────────────────────┤
  │                                   │
  │  7. Wait 60 seconds               │  8. Async: Evaluate health
  │                                   │     - Compare to baselines
  │                                   │     - Detect anomalies
  │                                   │     - Update health state
  │                                   │
  │  8. Next heartbeat...             │
```

### 3. Health Evaluation (Async)

```
Heartbeat arrives
     │
     ▼
┌────────────────┐
│ Store signals  │
└────┬───────────┘
     │
     ▼
┌─────────────────────┐     YES    ┌──────────────────┐
│ Agent in learning?  │───────────▶│ Add to baseline  │
└─────────┬───────────┘            │ sample (24h)     │
          │ NO                     └──────────────────┘
          ▼
┌─────────────────────┐     NO     ┌──────────────────┐
│ Enough samples?     │───────────▶│ Set: Learning    │
│ (min 24h)           │            └──────────────────┘
└─────────┬───────────┘
          │ YES
          ▼
┌─────────────────────┐
│ Compute baselines   │
│ - mean, stddev      │
│ - min, max          │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│ Compare current     │
│ values to baselines │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│ Count anomalies     │
│ (> 3 std devs)      │
└─────────┬───────────┘
          │
          ▼
     ┌────┴────┐
     │0 anomaly│  ──▶  Set: Healthy
     ├─────────┤
     │1 anomaly│  ──▶  Set: Degraded
     ├─────────┤
     │2+ anom. │  ──▶  Set: Attention
     └─────────┘
```

---

## mTLS Enrollment Flow

Mutual TLS authentication ensures only enrolled agents can send heartbeats.

### Enrollment Steps

1. **Agent generates identity material:**
   - UUID (agent_id) - persisted locally
   - 2048-bit RSA keypair
   - CSR with agent_id as Common Name

2. **Agent submits CSR to control plane:**
   - POST `/agents/register` with CSR PEM
   - Optional enrollment token for authorization

3. **Control plane validates and signs:**
   - Parse and validate CSR structure
   - Check enrollment token (if required)
   - Sign CSR with internal CA
   - Store agent record in database

4. **Agent receives signed certificate:**
   - Persist certificate to disk (`/etc/smidr/agent.crt.pem`)
   - Ready for mTLS communication

5. **All future heartbeats use mTLS:**
   - Agent presents client certificate
   - Control plane validates certificate chain
   - Agent ID extracted from certificate CN
   - Heartbeat processed if certificate valid

### Certificate Validation

- **Chain validation:** Client cert must chain to internal CA root
- **Expiry check:** Certificate must be within validity period
- **Revocation check:** Agent's `revoked_at` field must be null
- **CN extraction:** Agent ID must match certificate Common Name

---

## Health Evaluation Model

### Learning Phase

New agents start in `Learning` state for 24 hours minimum.

**Characteristics:**
- Collects all heartbeat signals
- No health alerts generated
- Computes statistical baselines after sufficient samples

**Baseline Computation (per metric):**
```
mean = Σ(values) / count
stddev = sqrt(Σ(value - mean)² / count)
min = minimum observed value (trimmed 1st percentile)
max = maximum observed value (trimmed 99th percentile)
```

### Evaluation Phase

After learning completes, each heartbeat triggers health evaluation.

**Anomaly Detection:**
- Compare current value to baseline mean
- Compute z-score: `z = (current - mean) / stddev`
- Anomaly if `|z| > 3` (beyond 3 standard deviations)

**Health State Transitions:**

| Anomalies | State      | Color  | Description                      |
|-----------|------------|--------|----------------------------------|
| 0         | Healthy    | Green  | All metrics within thresholds    |
| 1         | Degraded   | Orange | Single metric shows deviation    |
| 2+        | Attention  | Red    | Multiple metrics deviate         |
| N/A       | Unknown    | Gray   | No recent heartbeats (>5 min)    |
| N/A       | Learning   | Blue   | Baseline collection in progress  |

**State Persistence:**
- State changes recorded in `health_states` table
- Reason field captures which metrics triggered transition
- Timestamp records when change occurred

---

## Database Schema (Conceptual)

### agents

| Column             | Type      | Description                      |
|--------------------|-----------|----------------------------------|
| id                 | string    | UUID (agent_id)                  |
| hostname           | string    | System hostname                  |
| token              | string?   | Enrollment token (optional)      |
| certificate_pem    | string?   | Signed client certificate        |
| registered_at      | timestamp | First enrollment time            |
| last_heartbeat_at  | timestamp?| Most recent heartbeat            |
| current_health     | enum      | Current health state             |
| revoked_at         | timestamp?| Certificate revocation time      |

### heartbeats

| Column            | Type      | Description                      |
|-------------------|-----------|----------------------------------|
| id                | int       | Auto-increment primary key       |
| agent_id          | string    | Foreign key to agents            |
| timestamp         | timestamp | Agent's reported timestamp       |
| received_at       | timestamp | Control plane receive time       |
| uptime_seconds    | float     | System uptime                    |
| load_average_1m   | float     | 1-minute load average            |
| memory_used_pct   | float     | Memory usage percentage          |
| disk_used_pct     | float     | Disk usage percentage            |
| process_count     | int       | Total process count              |

### agent_baselines

| Column       | Type   | Description                          |
|--------------|--------|--------------------------------------|
| id           | int    | Auto-increment primary key           |
| agent_id     | string | Foreign key to agents                |
| metric_name  | string | Metric identifier (e.g., loadAverage)|
| mean         | float  | Baseline mean                        |
| std_dev      | float  | Baseline standard deviation          |
| min          | float  | Baseline minimum (trimmed)           |
| max          | float  | Baseline maximum (trimmed)           |
| sample_count | int    | Number of samples used               |

### health_states

| Column     | Type      | Description                          |
|------------|-----------|--------------------------------------|
| id         | int       | Auto-increment primary key           |
| agent_id   | string    | Foreign key to agents                |
| status     | enum      | Health state                         |
| reason     | string?   | Description of state trigger         |
| changed_at | timestamp | When state changed                   |

---

## Design Principles

### 1. Agents Report Facts, Not Opinions

Agents collect raw system signals without interpretation. All policy, evaluation, and decision-making happens centrally in the control plane.

**Why:** Centralized policy enables consistent behavior across all agents, simplifies updates, and keeps agents lightweight.

### 2. Silence is a Signal

Missing heartbeats indicate problems. Control plane tracks `last_heartbeat_at` and transitions agents to `Unknown` state after 5 minutes of silence.

**Why:** Crash or network partition is meaningful—treat absence of data as a signal itself.

### 3. Defaults Must Be Safe and Boring

Default configuration works out of the box. No surprises, no magic, no automagic tuning.

**Why:** Predictable behavior builds trust. Operators should never wonder "what will this do?"

### 4. Behavior Must Be Explainable

Every health state change includes a reason. Baselines and thresholds are inspectable. No black boxes.

**Why:** If operators can't explain why a system is flagged, they'll ignore alerts.

### 5. One Tenant Now, Many Later

v0 assumes single-tenant, but all entities include `tenant_id` fields for future multi-tenancy.

**Why:** Architecting for multi-tenancy from day one prevents painful refactoring later.

---

## Security Model

### Threat Model

**Assumptions:**
- Control plane is trusted and secure
- Agent systems may be compromised
- Network may be hostile (man-in-the-middle attacks)

**Goals:**
- Prevent unauthorized agents from sending data
- Prevent eavesdropping on heartbeat payloads
- Detect and reject revoked agents
- Maintain integrity of baseline learning

**Non-Goals (v0):**
- Protect against compromised control plane
- Detect malicious agents spoofing signals
- Prevent replay attacks (v1 concern)

### Security Mechanisms

**mTLS Authentication:**
- Bidirectional certificate validation
- Control plane validates agent certificates
- Agents validate control plane certificate (via CA cert)

**Certificate Revocation:**
- Database-backed revocation (no CRL distribution in v0)
- Control plane checks `revoked_at` field on every heartbeat
- Revoked agents receive 403 Forbidden

**Key Management:**
- CA private key stored with 600 permissions
- Agent private keys never transmitted over network
- Certificates issued with 365-day validity

**systemd Hardening (Agent):**
- `NoNewPrivileges=true` - prevent privilege escalation
- `PrivateTmp=true` - isolated /tmp
- `ProtectSystem=full` - read-only /usr and /boot
- `ProtectHome=true` - no access to user home directories

---

## API Versioning

v0 uses `/v0/` prefix for versioned endpoints, unversioned for infrastructure endpoints.

**Versioned Endpoints:**
- `/v0/agents/heartbeat` - agent telemetry

**Unversioned Endpoints:**
- `/agents/register` - agent enrollment (protocol unlikely to change)
- `/ca/certificate` - CA certificate download
- `/api/agents` - UI query endpoints (UI-coupled, not stable API)

**Rationale:** Heartbeat payload may evolve (new signals, metadata). Enrollment and CA operations are foundational and unlikely to change.

---

## Future Expansion (Not v0)

**Multi-Tenancy:**
- Per-tenant CAs or CA hierarchies
- Tenant isolation in database and API
- User authentication and authorization

**Integrations:**
- Slack, PagerDuty, email notifications
- Webhook delivery for state changes
- gRPC or WebSocket for real-time updates

**Advanced Diagnostics:**
- On-demand plugin execution (script runners)
- Extended signal collection (network, disk I/O)
- Log ingestion and correlation

**Policy Customization:**
- User-defined thresholds and evaluation rules
- Custom anomaly detection algorithms
- Per-agent or per-group policy overrides

---

## Operational Characteristics

### Heartbeat Cadence

Default: 60 seconds

**Rationale:** Balances freshness vs. network/storage overhead. 60s is fast enough to detect issues quickly, slow enough to avoid excessive data volume.

### Baseline Learning Window

Default: 24 hours minimum

**Rationale:** Captures daily usage patterns (business hours, batch jobs, backups). Longer windows (72h) capture weekly patterns but delay time-to-value.

### Health Check Staleness

Agents transition to `Unknown` after 5 minutes of missed heartbeats.

**Rationale:** 5 minutes allows for transient network issues (3 missed heartbeats) without false positives.

### Database Growth

**Heartbeats:** ~1 row/agent/minute = 1,440 rows/agent/day

**Retention (Future):** Likely 30-day rolling window for raw heartbeats, indefinite retention for baselines and health states.

---

## Development Workflow

1. **Kane (Agent Dev):** Implements signal collectors and heartbeat client
2. **Dallas (Control Plane):** Builds enrollment, storage, and evaluation services
3. **Ash (Crypto/Security):** Reviews mTLS implementation and CA operations
4. **Lambert (UI Dev):** Builds React UI consuming control plane API
5. **Parker (Test):** Writes unit and integration tests for all components
6. **Brett (Docs):** Maintains architecture, API, and user-facing documentation

**Integration Points:**
- Agent → Control Plane: JSON contracts for enrollment and heartbeat
- Control Plane → UI: REST API with JSON responses
- Control Plane → Database: EF Core migrations

---

## Success Criteria

Smidr v0 succeeds if:

- Agent installs in under 5 minutes
- Runs unattended for days without intervention
- Detects deviations without alert noise
- Health decisions are explainable to operators
- Users feel less need to manually check system status

---

**End of Architecture Overview**
