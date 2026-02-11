# Smidr - Secure, Managed Infrastructure Delivery & Response – v0 Specification

## 1. Purpose

The v0 system provides **quiet, continuous assurance** for Linux x86_64 systems by answering one core question:

> **“Is this system behaving normally right now?”**

The platform:

- Runs a lightweight agent on Linux systems
- Collects minimal system signals
- Learns baseline behavior centrally
- Detects meaningful deviations
- Surfaces issues with context, not noise

v0 prioritizes:

- correctness over completeness
- clarity over flexibility
- shipping over generality

---

## 2. Non-Goals (v0)

The following are **explicitly out of scope** for v0:

- No integrations (Slack, email, PagerDuty, etc.)
- No inbound agent connections
- No remote command execution
- No container-only abstraction
- No Windows or non-x86 architectures
- No user-configurable alert rules
- No multi-tenant UI or auth
- No dashboards or graphs
- No AI/ML beyond basic statistics

---

## 3. Target Environment

### Agents

- OS: Linux
- Architecture: x86_64
- Execution: systemd service
- Language: Go
- Network: outbound HTTPS only

### Control Plane

- Single-tenant (v0)
- Self-hosted or SaaS
- HTTPS + mutual TLS
- Centralized policy and learning

---

## 4. Core Design Principles

1. Agents report facts, not opinions
2. Policy and learning live centrally
3. Silence is a signal
4. Defaults must be safe and boring
5. Behavior must be explainable
6. One tenant now, many later

---

## 5. Agent Identity

### Canonical Identity

- Each agent has a UUID (`agent_id`)
- Generated once on first start
- Persisted locally
- Never changes

### Friendly Metadata

- Hostname
- OS / kernel version
- Agent version

UUID is authoritative.
Hostname is informational.

---

## 6. Authentication & Trust Model

### Mutual TLS (mTLS)

- Control plane generates a Certificate Authority (CA)
- Agent generates its own keypair on first start
- Agent submits CSR for enrollment
- Control plane signs and returns agent certificate
- All future communication uses mTLS

No API tokens are used.

---

## 7. Agent Lifecycle

### 7.1 First Start

1. Generate UUID
2. Generate private key
3. Generate CSR
4. Register with control plane
5. Receive signed certificate
6. Persist identity material locally

### 7.2 Normal Operation

- Agent runs indefinitely
- Collects system signals at fixed interval
- Sends heartbeat payloads
- Does not evaluate health
- Does not store historical data beyond transient buffering

### 7.3 Failure Modes

- If control plane unreachable, agent retries on next interval
- Agent never blocks system startup
- Agent failure does not affect host operation

---

## 8. Agent Data Collection (v0)

Collected metrics:

- uptime (seconds)
- load average (1m)
- memory used percentage
- disk used percentage (root filesystem)
- process count

Metrics are raw and unaggregated.

---

## 9. Communication Model

- Direction: Push-only (agent → control plane)
- Transport: HTTPS + JSON
- Authentication: mTLS
- Default cadence: 60 seconds

---

## 10. API Endpoints (v0)

### POST /v0/agents/register

Purpose: Agent enrollment
Input: CSR, agent UUID, hostname, OS metadata
Output: Signed agent certificate, CA chain

### POST /v0/agents/heartbeat

Purpose: Periodic state reporting
Authenticated via mTLS
Payload includes metrics and metadata

---

## 11. Learning & Baseline Model

### Learning Phase

- New agents start in `learning`
- Default window: 24–72 hours
- No alerts generated

### Baseline Computation

Per-metric:

- mean
- standard deviation
- trimmed min/max

### Threshold Derivation

- Warning and critical thresholds
- Derived centrally
- Explainable and inspectable

---

## 12. Health Evaluation

Health is evaluated only in the control plane.

### States

- learning
- healthy
- degraded
- attention
- unknown

### Transitions

- Threshold crossings
- Missing heartbeats
- Context captured on change

---

## 13. Context Capture (v0 Lite)

On state transition:

- Last known good state
- Current state
- Computed deltas

---

## 14. Storage Model (Conceptual)

All entities include `tenant_id` (fixed in v0).

Core entities:

- agents
- heartbeats
- baselines
- thresholds
- health events

---

## 15. UI (v0)

Minimal UI:

- System list
- System detail view

Displayed:

- Hostname
- Health state
- Last heartbeat
- Recent events
- Baseline vs current

No charts or dashboards.

---

## 16. Deployment Model

### Self-Hosted

- Single binary control plane
- Local CA generation
- Agent install script

### SaaS

- Same codebase
- Same protocol
- Same behavior

---

## 17. Future Expansion (Not v0)

- Multi-tenant support
- Per-tenant CAs
- Integrations
- Diagnostics plugins
- Event streaming
- gRPC transport
- Remote commands
- Policy customization

---

## 18. Guiding Rule

> **Agents report facts.
> Control plane decides meaning.**

---

## 19. Success Criteria

v0 is successful if:

- Agent installs in under 5 minutes
- Runs unattended for days
- Detects deviations without noise
- Health decisions are explainable
- Users feel less need to check manually
