# Providers

## 1. Purpose

Providers define *how* tasks are executed by the Smidr Daemon. A provider is a pluggable execution backend that translates a generic task definition into concrete runtime actions on the host system.

The provider system exists to decouple **task intent** (defined by the Smidr API) from **execution mechanics** (host-specific runtimes such as containers or native processes). This allows the daemon to support multiple execution environments without changing core scheduling, state, or communication logic.

In short:

* The API decides *what* should run
* The daemon decides *where* it can run
* The provider decides *how* it runs

---

## 2. Responsibilities

A provider is responsible for:

* Validating whether it can execute a given task type
* Preparing the execution environment
* Starting, monitoring, and stopping execution units
* Translating runtime-specific signals into daemon-level task state updates
* Exposing runtime capabilities and constraints to the daemon

A provider is **not** responsible for:

* Scheduling or prioritization
* Cross-task orchestration
* Persisting global state
* Communicating directly with the Smidr API

---

## 3. Provider Abstraction Model

Providers are implemented behind a stable interface that allows the daemon to treat all execution backends uniformly.

Conceptually, each provider implements the following lifecycle:

1. Capability detection
2. Task admission check
3. Execution preparation
4. Execution start
5. Runtime monitoring
6. Graceful termination or forced cleanup

The daemon selects a provider based on task type, declared requirements, and provider availability.

---

## 4. Provider Interface (Conceptual)

This section describes the *conceptual* interface. Exact method names and signatures are implementation-specific and language-dependent.

A provider should support operations equivalent to:

* `CanHandle(task)` – Determine whether the provider supports the task type and requirements
* `Prepare(task)` – Perform pre-execution setup (images, binaries, permissions)
* `Start(task)` – Begin execution and return a handle
* `GetStatus(handle)` – Report runtime status and health
* `Stop(handle, reason)` – Attempt graceful shutdown
* `Cleanup(handle)` – Ensure resources are released

Providers should be designed to be **idempotent** where possible, as execution may be retried after daemon restarts or transient failures.

---

## 5. Task Types and Provider Mapping

Task types are defined by the API and expressed to the daemon as enums or strongly-typed identifiers.

Providers declare which task types they support. Examples include:

* Container-based providers handling `CONTAINER_BUILD`, `IMAGE_PUSH`
* Native execution providers handling `SCRIPT_EXECUTION`
* Diagnostic providers handling `HEALTH_CHECK`

The daemon may support multiple providers for the same task type. Selection criteria may include:

* Provider availability
* Host capabilities
* Resource requirements
* Operator configuration

---

## 6. Capability Detection

On startup and periodically thereafter, providers perform capability detection. This allows the daemon to advertise accurate execution capabilities to the Smidr API.

Examples of detected capabilities:

* Presence and version of container runtimes
* OS and architecture compatibility
* Available CPU, memory, and disk constraints
* Feature flags (e.g., rootless execution, GPU support)

Capability detection must be fast, deterministic, and safe to run repeatedly.

---

## 7. Execution Isolation and Safety

Providers are responsible for enforcing isolation boundaries appropriate to their runtime. Depending on the provider, this may include:

* Container isolation (namespaces, cgroups)
* OS-level process isolation
* Filesystem sandboxing
* Network restrictions

When steps declare `inputs`, providers must ensure the materialized workdir is visible inside the execution environment (e.g., bind-mount host workdir into a container) while preserving path safety and least-privilege access.

Providers must assume that tasks may be untrusted and should apply the principle of least privilege whenever possible.

---

## 8. Error Handling and Failure Semantics

Providers translate runtime-specific failures into daemon-understood execution outcomes.

Examples include:

* Runtime startup failures
* Resource exhaustion
* Crashes or non-zero exit codes
* Forced termination

Providers do **not** decide retry policies. They report failures with sufficient context for the daemon and API to make scheduling and retry decisions.

---

## 9. Observability Integration

Providers integrate with the daemon’s observability pipeline by emitting:

* Execution logs
* Resource usage metrics
* Runtime health signals
* Structured failure reasons

Providers should avoid direct integration with external observability systems. All telemetry flows through the daemon.

---

## 10. Extensibility and Versioning

The provider system is designed to be extensible. New providers can be added without modifying existing ones, provided they conform to the provider interface.

Compatibility considerations:

* Providers may be versioned independently of the daemon
* Capability negotiation must tolerate older providers
* Breaking changes to the provider interface require coordinated upgrades

---

## 11. Example Providers (Non-Normative)

Examples of potential providers include:

* Docker Provider
* Podman Provider
* Native Process Provider
* Sandbox / VM-based Provider

These examples are illustrative only. The exact set of supported providers will evolve over time.

---

## 12. Future Considerations

Future enhancements to the provider system may include:

* Plugin-based loading of providers
* Remote execution providers
* Stronger sandboxing and policy enforcement
* Provider health scoring and automatic failover

The provider abstraction is intentionally conservative to ensure reliability, security, and long-term maintainability.
