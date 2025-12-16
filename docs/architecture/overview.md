# Smidr Daemon – Architecture Overview

## 1. Purpose

Smidr Daemon is a lightweight agent responsible for managing and executing tasks on behalf of the Smidr system. It serves as the local orchestrator that interfaces with the Smidr API to receive instructions, execute workloads, and report state transitions back to the central system.

This solves the need for a distributed execution environment that can operate close to the resources being managed, providing low-latency task execution and efficient resource utilization.

This is not handled by the API because the API is designed to be a centralized service that manages global state and user interactions, while the daemon is focused on local execution and resource management. The daemon typically runs on the local machine or node where tasks need to be executed, which may include edge devices or servers in a distributed network.

The Smidr API is the authoritative control plane. The daemon acts as a managed execution agent that initiates outbound connections to the API for receiving commands and reporting status. This design ensures that the daemon can operate behind firewalls and NATs without requiring complex inbound connectivity.

---

## 2. Non-Goals

There are things that the Smidr Daemon is explicitly not responsible for, including but not limited to:

* User interface rendering or management
* Long-term data storage or database management
* Complex business logic processing
* Cross-node orchestration or clustering

---

## 3. High-Level Responsibilities

This section outlines the primary functions of the Smidr Daemon. Those include:

* Orchestrating local system resources to execute tasks
* Managing container runtimes and other execution environments
* Executing and monitoring tasks as instructed by the Smidr API
* Establishing secure communication channels with the Smidr API for receiving commands and reporting state transitions
* Reporting task state transitions and daemon health back to the Smidr API
* Handling configuration and runtime parameters for task execution

---

## 4. Position Within the Smidr System

The Smidr Daemon operates as a local agent that interfaces with the Smidr API. It is responsible for executing tasks on the local machine or node, while the Smidr API serves as the central management point for user interactions and global state.

Mental model: the API is the control plane; the daemon is the execution plane. The daemon initiates outbound, secure connections to the API; the API does not assume inbound network access to the daemon.

The daemon communicates with the API to receive task instructions and report back execution status, ensuring that tasks are carried out efficiently and reliably on the local system. It acts as the bridge between the centralized control of the Smidr API and the distributed execution environment on individual nodes.

### Interaction Patterns

* **Task Reception**: The daemon maintains an outbound, persistent, secure connection (e.g., gRPC over TLS) to the API, listening for new task assignments
* **State Transition Reporting**: Daemon reports state transitions for tasks and itself to the API via outbound channels; no inbound ports are required on the daemon
* **Resource Registration**: On startup, the daemon registers its capabilities and available resources with the API
* **Event Streaming**: Real-time event notifications are streamed outbound to the API for critical state changes

### Data Flow

1. User submits a task through the Smidr API
2. API assigns the task to an appropriate daemon based on resource requirements and availability
3. Daemon receives task specification and validates prerequisites
4. Daemon executes the task using the appropriate provider
5. Execution progress and results are streamed back to the API
6. API makes results available to the user

### Multi-Daemon Coordination

When multiple daemons are deployed, the API serves as the coordination point. Daemons operate independently and do not communicate directly with each other. The API is responsible for:

* Load balancing tasks across available daemons
* Ensuring tasks are assigned to daemons with appropriate capabilities
* Handling failover if a daemon becomes unavailable

---

## 5. API Contract & Interface Definition

The daemon interacts with the Smidr API through API-hosted services. The daemon acts as a client and initiates outbound connections. Communications may use gRPC (primary, with optional bidirectional streaming) and/or REST where appropriate. This section is intentionally non-normative; exact service names, message shapes, and streaming patterns will be defined and versioned in the API repository.

### Service Interaction Model

* **API-hosted services**: The API exposes services for task assignment, status ingestion, metrics, and capability registration
* **Daemon as client**: The daemon connects outward to consume task assignments and report state transitions and metrics
* **Streaming patterns**: Bidirectional or server-streaming channels may be used for task assignments and state reports depending on QoS and latency needs
* **Protocols**: gRPC over TLS is preferred; REST endpoints may be provided for compatibility and tooling

### Responsibilities (non-normative)

* Register daemon identity, capabilities, and available resources
* Receive task assignments and acknowledgements
* Fetch task specifications and required artifacts
* Report task state transitions (accepted, running, completed, failed, canceled)
* Stream execution progress, logs, and result metadata
* Report daemon health and runtime metrics
* Send heartbeats and participate in liveness checks
* Apply retry/backoff policies and idempotent reporting

### Task State Machine

Tasks follow a well-defined state machine throughout their lifecycle. The daemon is responsible for correctly capturing and reporting state transitions to the API. The minimal set of states includes:

* `PENDING`: Task received and queued, awaiting execution resources
* `ACCEPTED`: Daemon has claimed the task and begun setup
* `RUNNING`: Task is actively executing
* `COMPLETED`: Task finished successfully
* `FAILED`: Task encountered a fatal error during execution
* `CANCELED`: Task was canceled by the API or user
* `TIMED_OUT`: Task exceeded its deadline
* `REJECTED`: Daemon rejected the task before execution (e.g., unsupported type)
* `ORPHANED`: Task state is unknown due to daemon restart or crash

Exact state definitions, including any intermediate or provider-specific states, will be formalized in the API specification.

### Illustrative examples (non-normative)

The following snippets illustrate potential shapes for messages. They are examples only and are not binding.

```protobuf
enum TaskType {
  TASK_TYPE_UNSPECIFIED = 0;
  CONTAINER_BUILD = 1;
  SCRIPT_EXECUTION = 2;
  IMAGE_PUSH = 3;
  HEALTH_CHECK = 4;
}

message TaskInstruction {
  string task_id = 1;
  TaskType task_type = 2;
  map<string, string> parameters = 3;
  ResourceRequirements resources = 4;
  int32 priority = 5;
  google.protobuf.Timestamp deadline = 6;
}
enum TaskState {
  TASK_STATE_UNSPECIFIED = 0;
  PENDING = 1;
  ACCEPTED = 2;
  RUNNING = 3;
  COMPLETED = 4;
  FAILED = 5;
  CANCELED = 6;
  REJECTED = 7;
  TIMED_OUT = 8;
  ORPHANED = 9;
}

message TaskStatusReport {
  string task_id = 1;
  TaskState state = 2;
  string message = 3;
  float progress_percent = 4;
  google.protobuf.Timestamp timestamp = 5;
}
```

Notes:

* Task types are represented as an enum to avoid stringly-typed contracts.
* Exact enums, fields, and services will be finalized alongside API implementation.

---

## 6. System Architecture

[Placeholder for architecture diagram]

The Smidr Daemon consists of the following major components and their interactions:

* **API Communication Layer**: Maintains bidirectional gRPC connections with the Smidr API
* **Task Manager**: Receives task instructions and coordinates execution
* **Provider System**: Abstracts runtime-specific implementations (Docker, Podman, etc.)
* **Execution Engine**: Manages task lifecycle and monitors execution
* **Scheduler**: Queues and prioritizes tasks based on resources
* **State Manager**: Persists task and system state
* **Health Monitor**: Tracks daemon and task health
* **Metrics Collector**: Gathers performance metrics

These components work together to provide a robust, scalable task execution environment.

## 7. Core Subsystems

### 7.1 Provider System

The provider system abstracts the underlying execution environments. This allows the daemon to support multiple runtimes (e.g., Docker, Podman, native processes) without changing its core logic.  This subsystem is responsible for:

* Detecting available runtimes on the host system
* Providing a uniform interface for task execution regardless of the underlying runtime
* Managing runtime-specific configurations and optimizations

---

### 7.2 Runtime and Execution Engine

The execution engine is responsible for the actual execution of tasks while this subsystem manages the lifecycle of these tasks.  The execution engine handles:

* Starting and stopping tasks
* Monitoring task health and status
* Reporting task results back to the main daemon logic

---

### 7.3 Communication Layer

Communication for Smidr Daemon is handled through a dedicated communication layer. This will be using gRPC as the main form of communication, but other protocols may be supported in the future. This subsystem is responsible for:

* Establishing secure connections to the Smidr API
* Sending and receiving messages (task instructions, status updates)
* Handling retries and error conditions in communication
* Supporting multiple communication protocols if needed
* Ensuring message integrity and confidentiality
* Managing connection lifecycle (reconnects, timeouts)
* Logging communication events for diagnostics
* Supporting asynchronous communication patterns
* Facilitating local inter-process communication if required
* Implementing protocol versioning and compatibility checks
* Providing hooks for custom communication handlers or plugins
* Monitoring communication performance and latency

Authentication and authorization mechanisms to ensure secure access to the Smidr API will also be part of this subsystem. Along with encryption of data in transit to protect sensitive information.  There will be an expected request/response pattern for task instructions and status updates.  This will follow a publish/subscribe model for certain event notifications.

---

### 7.4 Scheduling and Background Work

The daemon includes a scheduling subsystem to manage background tasks and long-running jobs.  This subsystem is responsible for:

* Scheduling tasks based on priority and resource availability
* Managing task queues
* Allocating system resources for task execution
* Retrying failed tasks based on configurable policies

---

### 7.5 State Management

Tasks and system state are tracked by the daemon to ensure accurate execution and reporting.  This subsystem is responsible for:

* Maintaining the current state of tasks (running, completed, failed, etc.)
* Tracking system resource usage
* Persisting state information to survive restarts
* Providing state information to other subsystems as needed

---

## 8. Configuration Model

The daemon is configurated through a combination of configuration files, environment variables, and API-driven settings. This will allow for flexible deployment and management of the daemon across different environments.  The configuration model includes:

* Support for multiple configuration sources with precedence rules
* Runtime-specific configuration options for different execution environments
* Secure handling of sensitive information such as API keys and credentials through secrets management mechanisms
* Dynamic reloading of configuration without requiring a restart
* Validation of configuration parameters to prevent misconfiguration

---

## 9. Security Model

For security, the daemon implements several controls to ensure safe operation. This will be critical given its role in executing potentially sensitive tasks.  The following security aspects will be covered:

* Local vs remote trust
* API authentication
* Privileged operations
* OS-level permissions
* Secure storage of sensitive data
* Network security measures (e.g., TLS encryption)
* Regular security audits and updates

---

## 10. Deployment and Lifecycle

Smidr Daemon deployment and lifecycle management is designed to be straightforward and robust. It will be packaged for easy installation across supported operating systems.  It should be able to run as a background service/daemon, managed by standard service managers.  The lifecycle management includes:

* Installation procedures for different platforms - package managers, binaries, etc.
* Starting, stopping, and restarting the daemon
* Monitoring and automatic recovery from failures
* Upgrade strategies to ensure minimal downtime and data integrity

We will have github releases for versioned binaries.  There will be support for containerized deployments using Docker or similar technologies.  The daemon will also support auto-updates to ensure it remains current with the latest features and security patches.  Github actions can be used to build and publish releases automatically.

---

## 11. Observability

Observability is a key aspect of the Smidr Daemon, enabling operators to monitor its health and performance.  The daemon will expose various diagnostics through logging, metrics, and health checks.  This includes:

* Detailed logging of operations, errors, and significant events
* Metrics collection for performance monitoring and alerting
* Health check endpoints to verify the daemon's operational status
* Integration with external monitoring systems (e.g., Prometheus, Grafana)
* Support for log aggregation and analysis tools
* Tracing capabilities to follow task execution flows

---

## 12. Future Considerations

There are many nice to have features that can be considered for future versions of the Smidr Daemon. Included in that are plugins to extend functionality, enhanced security features, and improved performance optimizations. Multi-node coordination could be another area of future development to allow for more complex distributed task execution scenarios. Remote ex Pecution capabilities could also be explored to enable tasks to be run on other nodes or systems as needed, this however may introduce additional security and complexity considerations. And sandboxing and isolation improvements to enhance security and reliability of task execution.
