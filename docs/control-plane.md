# Control Plane Responsibility Contract

## Purpose

Defines the responsibilities of the centralized Control Plane as the authoritative governance and orchestration layer of the platform. Ensures consistent policy enforcement, multi-tenant isolation, and auditability.

---

## Responsibilities

### Identity & Access Management

- Authenticate users, agents, and plugins
- Enforce role-based access control (RBAC)
- Manage scoped credentials per tenant/project

### Policy Enforcement

- Evaluate execution, agent, artifact, and security policies
- Merge tenant and project-level rules deterministically
- Issue immutable job contracts

### Job Lifecycle Orchestration

- Track job states (Submitted → Completed)
- Assign jobs to eligible agents
- Handle retries, cancellations, and timeouts according to policy

### Agent Coordination

- Maintain agent registry
- Monitor agent health, status, and capabilities
- Notify Job Controller of agent availability or failures

### Artifact Management

- Maintain metadata and lineage DAG
- Track promotions, retention, and access controls
- Integrate with Artifact Store for immutable storage

### Audit & Observability

- Log all policy decisions, job state transitions, and agent interactions
- Provide dashboards and reporting for tenants and admins

---

## Non-Responsibilities

- Direct job execution (handled by agents)
- Host-level operations or sandboxing
- Storing artifacts permanently (Artifact Store responsibility)
- Interpreting or enforcing plugin-specific logic outside policies

---

## Control Plane Metadata

- Tenant and project definitions
- Policy definitions per tenant/project
- Job contracts and state records
- Agent registry and health status
- Artifact metadata and lineage information

---

## User Stories

### Platform Administrator

- I want consistent policy enforcement across all tenants.
- I want full observability of agent and job states.

### MSP Operator

- I want strict isolation between tenants and projects.
- I want per-customer audit logs and reporting.

### Embedded Systems Lead

- I want governance without micromanagement.
- I want reproducible builds and artifact tracking across environments.

---

## Notes

- The Control Plane is the authoritative layer: all agents, jobs, artifacts, and policies ultimately defer to it.
- It ensures deterministic, auditable, and secure execution across multi-tenant deployments.
- Any changes in policy or job orchestration are logged for compliance and traceability.
