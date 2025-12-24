# Job Lifecycle Specification

## Purpose

Defines the full lifecycle of a job from submission to completion, including state transitions, guarantees, and user interactions. This ensures reproducibility, traceability, and proper policy enforcement.

---

## Job States

- **Submitted:** Job has been created and accepted by the control plane.
- **Queued:** Job is waiting for an eligible agent.
- **Dispatched:** Job has been assigned to an agent.
- **Running:** Agent has started execution.
- **Succeeded:** Job finished successfully.
- **Failed:** Job finished with an error.
- **Cancelled:** Job was cancelled by a user or system policy.
- **Timed Out:** Job exceeded the maximum allowed execution time.

---

## State Transitions

| From         | To           | Trigger / Notes |
| ------------ | ------------- | ---------------- |
| Submitted  | Queued      | Job accepted and validated by Control Plane |
| Queued     | Dispatched  | Agent assigned by Job Controller |
| Dispatched | Running     | Agent acknowledges job start |
| Running    | Succeeded   | Job completes successfully |
| Running    | Failed      | Job completes with error |
| Running    | Timed Out   | Execution exceeds configured timeout |
| Running    | Cancelled   | Cancelled by user or policy |
| Failed / Timed Out | Queued | Optional retry controlled by policy |

---

## Guarantees

- **Immutable Job Definition:** Once submitted, job configuration cannot be altered.
- **Audit Trail:** All state transitions and actions are logged.
- **Deterministic Transitions:** Transitions are predictable and reproducible across agents.
- **Policy Enforcement:** Jobs cannot bypass tenant or project policies.

---

## Job Metadata

Each job contains:

- Job ID (unique)
- Tenant ID
- Project / Workspace ID
- Agent requirements
- Plugin version / type
- Resource limits (CPU, memory, disk)
- Submission timestamp
- User ID of submitter

---

## User Stories

### Embedded Systems Engineer

- I want predictable state transitions to ensure reproducible builds.
- I want to see detailed job logs from all stages.

### MSP Operator

- I want to trace the lifecycle of every job for compliance per customer.
- I want to automatically retry failed jobs under policy constraints.

### Platform Administrator

- I want full visibility of queued, running, and completed jobs.
- I want alerts for long-running or stuck jobs.

---

## Notes

- Job lifecycle is central to multi-tenant isolation and auditability.
- All other modules (agents, policy engine, plugin interface) rely on these states to enforce contracts.
- Job metadata and logs feed into artifact lineage and compliance tracking.
