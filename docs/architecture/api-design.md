# Control Plane API Design (v0.1.0)

## Scope

Authoritative control plane for Smidr. Owns task lifecycle, scheduling, policy, persistence, and surfacing state to users and executors.

## Responsibilities

- Accept job submissions (REST/gRPC) with validation and policy enforcement
- Persist tasks, queues, and results; expose filtered views to users
- Schedule jobs by matching requirements to executor capabilities and labels
- Enforce the task lifecycle contract and reject invalid transitions
- Stream logs/events and surface artifact metadata
- Authn/z: secure the UI and executor endpoints; support executor identity + liveness
- Versioned contracts: advertise server-supported versions; reject incompatible executors

## Interfaces (initial)

- Job submission/query: create job, fetch job, list jobs, cancel job
- Executor control: register/update capabilities, heartbeat, pull next job, report status/logs, upload artifacts metadata
- Observability: stream events/logs; health endpoint
- Admin/policy: configure queues, priorities, and capability-based routing (future)

## Data + Storage

- Jobs/queue/state: durable DB (SQL/NoSQL acceptable; choose based on earliest implementation)
- Logs/events: append-only; enable streaming and retention
- Artifacts: object storage (S3-compatible) referenced by metadata in DB
- Idempotency keys for job submission and executor reports

## Scheduling Model

- Pull-based: executors request work with capability payload and liveness context
- Matching: labels/capabilities + resource hints; priority-aware FIFO to start
- Backoff: per-executor backoff when no work or on repeated failures
- Fairness: avoid hot-spotting; consider per-queue weights (future)

## Security

- Mutual TLS or token-based auth for executors; user auth for UI/API
- Authorization: role-based for users; executor-scoped permissions for only its jobs
- Audit: log all state transitions and admin actions

## Compatibility

- All public messages versioned; executors declare supported versions on registration
- Control plane assigns jobs only when version intersection is valid

## Failure Handling

- Duplicate/late reports must be idempotent
- Missed heartbeats → mark executor stale; jobs move to ORPHANED and reassignable
- Partial uploads/logs handled via resumable streams where possible

## Open Questions

- Exact persistence tech for v0.1.0
- Priority semantics and starvation avoidance
- Multi-tenant boundaries and per-tenant quotas
