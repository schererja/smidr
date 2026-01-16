# Tenant & Policy Model

## Purpose

Defines organizational isolation, governance, and policy enforcement across tenants, projects, users, agents, and jobs. This ensures multi-tenant safety, reproducibility, and operational compliance.

---

## Core Concepts

### Tenant

- Top-level organizational boundary
- Represents a company, customer, or department
- Defines access control, policies, and resource limits

### Project / Workspace

- Logical subdivision within a tenant
- Allows grouping of jobs, agents, and artifacts
- Policies can be applied at the project level

### User

- Human identity with roles and permissions
- Roles include:
  - Admin
  - Operator
  - Engineer
- Role-based access controls enforce allowed operations

### Agent

- Execution node tied to tenant/project
- Agents may have capabilities restricting which jobs they can execute

---

## Policy Categories

### Execution Policies

- Job type restrictions
- Resource limits (CPU, memory, disk)
- Network and environment constraints

### Agent Policies

- Allowed job types per agent
- Maximum concurrent jobs
- Platform/hardware constraints

### Artifact Policies

- Retention duration
- Promotion rules (dev → QA → prod)
- Signing and verification requirements

### Security Policies

- MFA enforcement
- Approval workflows for sensitive operations
- Scoped secrets access

---

## Policy Evaluation Flow

1. Job request submitted by user
2. Tenant-level policies evaluated
3. Project-level policies merged
4. Agent capability matching
5. Final policy decision issued
6. Immutable job contract created with enforced constraints

---

## User Stories

### Tenant Admin

- I want to enforce global execution limits per tenant.
- I want retention rules and promotion policies applied automatically.

### MSP Operator

- I want separate projects per customer with strict isolation.
- I want to restrict allowed job types based on customer agreements.

### Build Engineer

- I want reproducible builds that respect project policies.
- I want automatic artifact promotion following defined rules.

### Auditor

- I want immutable records of all job executions and policy decisions.
- I want full visibility of policy enforcement and artifact handling.

---

## Notes

- This model underpins multi-tenant governance and safety.
- All other components (agents, job controller, plugins, artifact store) rely on this model for policy decisions.
- Policy enforcement is deterministic and auditable.
