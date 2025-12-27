# Control Plane Responsibility Contract

## Purpose

Defines the responsibilities of the serverless Control Plane (TypeScript/Lambda on AWS) as the authoritative governance and orchestration layer of the platform. Ensures consistent policy enforcement, multi-tenant isolation, and auditability while keeping operational costs low.

---

## Technology Stack

- **Runtime**: AWS Lambda (Node.js 20+)
- **API**: HTTP API Gateway (not REST API to reduce cost)
- **Storage**: DynamoDB on-demand (tenants, agents, jobs, artifacts metadata)
- **Queue**: SQS (job distribution)
- **Authentication**: HMAC-SHA256 (agent requests), API Keys (frontend)
- **Observability**: CloudWatch Logs, CloudWatch Metrics, optional X-Ray

---

## Responsibilities

### Identity & Access Management

- Authenticate agents via HMAC signatures (tenant ID + base64 secret)
- Authenticate users/frontend via API keys or AWS IAM
- Enforce role-based access control (RBAC)
- Store credentials securely (AWS Secrets Manager for HMAC secrets, IAM for users)

### Policy Enforcement

- Evaluate execution, agent, artifact, and security policies
- Merge tenant and project-level rules deterministically
- Issue immutable job contracts

### Job Lifecycle Orchestration

- Track job states (Submitted → Completed) in DynamoDB
- Place jobs on SQS queue for agent pickup (pull model)
- Assign jobs to eligible agents based on capabilities
- Handle retries, cancellations, and timeouts according to policy

### Agent Coordination

- Maintain agent registry in DynamoDB (tenant_id + agent_id)
- Monitor agent health, status, and capabilities via heartbeat REST endpoint
- Track agent state (Idle, Busy, Offline, Decommissioned)
- Notify agents of policy updates or job cancellations

### Artifact Management

- Maintain metadata and lineage DAG in DynamoDB
- Track promotions, retention, and access controls
- Integrate with S3 for immutable object storage
- Enforce artifact policies (stage promotion, expiration, signing)

### Audit & Observability

- Log all policy decisions, job state transitions, and agent interactions to CloudWatch
- Provide metric data (job counts, execution times, error rates)
- Enable tracing via CloudWatch Logs and optional X-Ray
- Support tenant-scoped audit log retrieval via REST API

---

## Non-Responsibilities

- Direct job execution (handled by agents)
- Host-level operations or sandboxing
- Storing artifacts permanently (Artifact Store responsibility)
- Interpreting or enforcing plugin-specific logic outside policies

---

## Control Plane Metadata

All metadata stored in DynamoDB on-demand:
- Tenant and project definitions (tenant_id, project_id, policies, quotas)
- Policy definitions per tenant/project (execution, agent, artifact, security)
- Job records and state history (job_id, state transitions, assigned agent, result)
- Agent registry and health status (agent_id, capabilities, state, last heartbeat)
- Artifact metadata and lineage information (artifact_id, parent/child, stage, expiration)

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
- Any changes in policy or job orchestration are logged to CloudWatch for compliance and traceability.
- Serverless design (Lambda + DynamoDB on-demand) minimizes operational overhead and scales automatically with workload.
- HMAC authentication avoids managing PKI; secrets are rotated in AWS Secrets Manager.
