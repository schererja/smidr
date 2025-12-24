# Agent Responsibility Contract

## Purpose

Defines what the agent is responsible for and what it must never do. Ensures safe, predictable execution in a multi-tenant environment and clear boundaries between agent, control plane, and plugins.

---

## Responsibilities

### Authentication & Security

- Securely authenticate to the Control Plane
- Use scoped credentials and encryption for communication
- Apply sandboxing to prevent cross-tenant access

### Job Execution

- Pull and execute assigned jobs in isolated environments
- Enforce job-level policy constraints received from the Control Plane
- Report status, progress, logs, and artifacts back to Control Plane

### Reliability

- Fail safely and idempotently
- Retry transient operations according to job policy
- Handle cancellations or timeouts cleanly

---

## Non-Responsibilities

- Scheduling jobs or retry policies (handled by Job Controller)
- Interpreting or modifying policies (Control Plane responsibility)
- Cross-tenant operations
- Modifying job definitions or artifacts outside the assigned scope
- Storing artifacts permanently (Artifact Store responsibility)

---

## Agent Metadata

- Agent ID
- Tenant/Project association
- Supported platforms / capabilities
- Registration timestamp
- Last heartbeat / status

---

## User Stories

### Embedded Engineer

- I want consistent, reproducible builds across agent restarts.
- I want safe failure handling to prevent job corruption.

### MSP Operator

- I want agents to enforce tenant isolation automatically.
- I want the ability to revoke agents without impacting other tenants.

### Platform Administrator

- I want agents to report logs and metrics reliably.
- I want clear boundaries of responsibilities to enforce system integrity.

---

## Notes

- The Agent Responsibility Contract ensures deterministic and safe job execution.
- It defines clear separation of concerns between agent, job controller, and control plane.
- Agents must be replaceable and observable without disrupting ongoing jobs.
