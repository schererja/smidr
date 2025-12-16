# Task Lifecycle

## 1. Purpose

This document defines the authoritative lifecycle for tasks within the Smidr system.

A *task* represents a unit of work requested by the Smidr API and executed by a Smidr Daemon. The task lifecycle establishes a shared contract between the API (control plane), the daemon (execution plane), and providers (execution backends).

The goals of this lifecycle definition are to:

* Provide a single, consistent task state model across the system
* Clearly define ownership of state transitions
* Enable reliable retries, recovery, and observability
* Prevent state divergence between API and daemon

This document is normative unless otherwise stated.

---

## 2. Architectural Placement

This document belongs in the **architecture** documentation domain.

Rationale:

* The task lifecycle is a cross-cutting concern spanning API, daemon, and providers
* It defines system-wide invariants, not implementation details
* It must remain stable even as languages, frameworks, or transports change

Neither the API nor the daemon owns the lifecycle definition; they both implement it.

---

## 3. Task Ownership Model

A task exists in two conceptual forms:

* **Authoritative task record** (API)
* **Execution instance** (Daemon)

Ownership rules:

* The Smidr API is the *source of truth* for task existence and final outcome
* The Smidr Daemon is the *source of truth* for in-progress execution state
* Providers do not own task state; they emit execution signals

State transitions are proposed by the daemon and accepted, persisted, and exposed by the API.

---

## 4. Task State Enumeration

The system defines a minimal, explicit set of task states. These states are shared across API, daemon, and providers.

```text
CREATED
ASSIGNED
ACCEPTED
RUNNING
COMPLETED
FAILED
CANCELED
REJECTED
TIMED_OUT
ORPHANED
```

### State Definitions

* **CREATED** – Task has been created by the API but not yet assigned
* **ASSIGNED** – Task has been assigned to a specific daemon
* **ACCEPTED** – Daemon has acknowledged the task and committed to execution
* **RUNNING** – Task execution has started
* **COMPLETED** – Task finished successfully
* **FAILED** – Task terminated with an error
* **CANCELED** – Task was canceled before completion
* **REJECTED** – Task was rejected by the daemon (transient failure)
* **TIMED_OUT** – Task exceeded its allowed execution time
* **ORPHANED** – Task lost its daemon assignment (e.g., daemon crash)

States are intentionally coarse-grained. Additional detail is conveyed through metadata, events, and logs.

---

## 5. Legal State Transitions

Only the following transitions are permitted:

```text
CREATED   -> ASSIGNED
ASSIGNED  -> ACCEPTED
ACCEPTED  -> RUNNING
RUNNING   -> COMPLETED
RUNNING   -> FAILED
ASSIGNED  -> CANCELED
ACCEPTED  -> CANCELED
RUNNING   -> CANCELED
RUNNING   -> TIMED_OUT
ASSIGNED  -> REJECTED
ACCEPTED  -> REJECTED
RUNNING   -> ORPHANED
```

Transitions not listed above are invalid and must be rejected by the API. This is a work in progress and may be extended in future versions.

---

## 6. Component Responsibilities

### API Responsibilities

* Create tasks and assign them to daemons
* Persist authoritative task state
* Validate state transitions
* Expose task state to users and external systems
* Decide retry, reschedule, or terminal behavior

### Daemon Responsibilities

* Accept or reject assigned tasks
* Execute tasks according to the lifecycle
* Propose valid state transitions
* Report progress, logs, and metadata
* Recover execution state after restarts

### Provider Responsibilities

* Execute runtime-specific actions
* Emit execution signals (started, exited, errored)
* Never mutate task state directly

---

## 7. Failure Semantics

Failures are classified as:

* **Execution failures** (runtime errors, non-zero exit codes)
* **Infrastructure failures** (resource exhaustion, provider unavailability)
* **Daemon failures** (crash, restart, loss of connectivity)

The daemon reports failures with sufficient context. The API determines whether a failure is terminal or retryable.

---

## 8. Idempotency and Recovery

The task lifecycle is designed to tolerate:

* Daemon restarts
* Network interruptions
* Duplicate messages

Key principles:

* State transitions must be idempotent
* ACCEPTED implies execution intent but not completion
* RUNNING may be re-reported after recovery
* Final states (COMPLETED, FAILED, CANCELED) are immutable

---

## 9. Observability and State

Every state transition must produce:

* A timestamped event
* A structured reason or cause
* Optional metadata for diagnostics

State transitions are the backbone of system observability and auditability.

---

## 10. Future Extensions

Potential future enhancements include:

* Explicit retry or backoff states
* Paused or suspended execution
* Task dependencies and DAGs
* Partial completion semantics

Any extension must preserve backward compatibility with existing states.

---

## 11. Summary

The task lifecycle defines the contract that binds the Smidr system together. By strictly separating ownership, enforcing valid transitions, and centralizing authority, Smidr achieves predictable, debuggable, and resilient task execution.
