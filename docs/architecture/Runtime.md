# Runtime

## 1. Purpose

This document defines the runtime execution model for the Smidr Daemon.

The runtime layer is responsible for *how* a task is executed on a host once it has been accepted by the daemon and assigned to a provider. It bridges the gap between abstract task intent and concrete operating system or container-level execution.

The goals of the runtime model are to:

* Provide a consistent execution environment across providers
* Enforce isolation and safety boundaries
* Enable predictable lifecycle management and cleanup
* Support multiple execution backends without leaking runtime details upward

This document is normative unless otherwise stated.

---

## 2. Architectural Placement

The runtime layer sits below task scheduling and above provider-specific implementations.

Layering (top to bottom):

1. Smidr API (control plane)
2. Smidr Daemon (execution plane)
3. Scheduler / Task Manager
4. **Runtime Layer**
5. Provider Implementations
6. Host OS / Container Runtime

The runtime layer is internal to the daemon and is not exposed directly to the API.

---

## 3. Runtime Responsibilities

The runtime layer is responsible for:

* Creating execution contexts for tasks
* Applying isolation, permissions, and resource limits
* Managing process or container lifecycles
* Collecting execution signals (exit codes, signals, crashes)
* Ensuring deterministic cleanup of resources

The runtime layer is **not** responsible for:

* Deciding which tasks to run (scheduling)
* Persisting authoritative task state
* Communicating directly with the Smidr API
* Implementing provider-specific logic

---

## 4. Execution Context

Each task execution occurs within an *execution context*.

An execution context encapsulates:

* Task identity
* Assigned provider
* Runtime configuration (resources, limits, environment)
* Isolation boundaries
* Execution handles (process IDs, container IDs)

Execution contexts are created by the runtime layer and passed to providers for execution.

Contexts are immutable after creation, except for runtime-observed state (e.g., timestamps, exit codes).

---

## 5. Isolation Model

Isolation is applied by the runtime layer in cooperation with the provider.

Depending on provider capabilities, isolation may include:

* Process isolation
* Container namespaces
* cgroups or equivalent resource controls
* Filesystem sandboxing
* Network restrictions

The runtime layer defines *what* isolation guarantees must exist. Providers define *how* they are implemented.

Isolation must default to the principle of least privilege.

---

## 6. Resource Management

The runtime layer enforces resource constraints declared in the task specification.

Resources may include:

* CPU
* Memory
* Disk I/O
* Network access
* Specialized resources (e.g., GPU)

The runtime layer is responsible for:

* Translating abstract resource requests into runtime-enforceable limits
* Monitoring resource consumption during execution
* Emitting signals when limits are exceeded

The API determines policy consequences of resource violations.

---

## 7. Lifecycle Management

The runtime layer manages the execution lifecycle once a task enters the RUNNING state.

Lifecycle phases:

1. Context creation
2. Environment preparation
3. Execution start
4. Active monitoring
5. Graceful termination or forced stop
6. Cleanup and teardown

The runtime layer must guarantee that cleanup occurs even in the presence of failures or daemon restarts.

---

## 8. Signals and Exit Semantics

The runtime layer normalizes execution outcomes into daemon-understood signals.

Examples include:

* Normal completion
* Non-zero exit codes
* Termination signals
* Timeouts
* Resource exhaustion

Providers report raw signals. The runtime layer translates them into standardized execution outcomes.

---

## 9. Idempotency and Recovery

Runtime operations must tolerate retries and partial execution.

Principles:

* Execution start must be idempotent where possible
* Orphaned execution units must be detected and reconciled
* Cleanup operations must be safe to re-run

On daemon restart, the runtime layer attempts to reconcile observed execution state with persisted task state.

---

## 10. Observability

The runtime layer emits execution-level telemetry, including:

* Start and stop timestamps
* Resource usage samples
* Exit reasons and codes
* Runtime errors

Telemetry is consumed by the daemon and forwarded to the API through the communication layer.

---

## 11. Security Considerations

The runtime layer is a security boundary.

Key considerations:

* Minimizing privilege escalation risk
* Preventing task escape from isolation boundaries
* Protecting host resources from abuse
* Securing secrets and environment variables

Security posture must favor safety over performance.

---

## 12. Future Extensions

Potential future enhancements include:

* Stronger sandboxing (VM-based execution)
* Remote or delegated execution contexts
* Runtime policy enforcement engines
* Execution replay or debugging modes

Extensions must preserve the existing runtime contract.

---

## 13. Summary

The runtime layer provides a stable, secure, and predictable foundation for task execution within Smidr. By separating runtime concerns from providers and scheduling logic, Smidr enables extensibility while maintaining strong guarantees around isolation, cleanup, and observability.
