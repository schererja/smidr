# Smidr Platform - Protocol Buffer Definitions

This directory contains the complete Protocol Buffer (protobuf) definitions for the Smidr platform, defining all gRPC services and message types for communication between the Control Plane, Agents, and other components.

## Directory Structure

```
proto/smidr/v1/
├── common/
│   ├── types.proto           # Core data types (IDs, states, capabilities, metrics)
│   ├── policy.proto          # Policy definitions (execution, agent, artifact, security)
│   ├── execution.proto       # Execution context (deprecated - use Job)
│   ├── status.proto          # Status enums (deprecated - use JobState)
│   ├── errors.proto          # Error definitions
│   └── identifiers.proto     # Identifier types (deprecated - use ID)
│
├── agent/
│   └── agent.proto           # Agent service definitions and messages
│
└── control_plane/
    ├── job.proto             # Job management services
    ├── artifact.proto        # Artifact lineage and management
    ├── tenant.proto          # Tenant, project, and user management
    └── control_plane.proto   # Control plane orchestration services
```

## Core Components

### 1. Common Types (`common/types.proto`)

Defines foundational data structures used across all services:

- **ID**: Universal identifier type for jobs, agents, artifacts, tenants, users, etc.
- **JobState**: Enum defining job lifecycle states (Submitted, Queued, Dispatched, Running, Succeeded, Failed, Cancelled, TimedOut)
- **AgentState**: Enum for agent lifecycle (Registering, Idle, Busy, Offline, Decommissioned)
- **ResourceLimits**: CPU, memory, disk, and timeout constraints
- **Platform & Architecture**: Supported platforms (Linux, Windows, macOS) and architectures (AMD64, ARM64, ARM)
- **AgentCapabilities**: Agent features, platform, architecture, and custom labels
- **AuditMetadata**: Timestamps and user tracking for audit trails
- **Metadata**: Generic key-value pairs for extensibility

### 2. Policy Definitions (`common/policy.proto`)

Defines policy evaluation and enforcement:

- **ExecutionPolicy**: Constraints on job execution (allowed types, resource limits, concurrent jobs)
- **AgentPolicy**: Restrictions on what agents can execute
- **ArtifactPolicy**: Retention, promotion, and signing rules
- **SecurityPolicy**: Authentication, MFA, approval requirements
- **PolicySet**: Complete policy bundle for a tenant or project
- **PolicyDecision**: Result of policy evaluation

### 3. Agent Service (`agent/agent.proto`)

**Services:**

- `RegisterAgent`: Register a new agent with the control plane
- `Heartbeat`: Periodic health checks and status reporting
- `GetJobAssignment`: Pull model for job assignment
- `ReportJobProgress`: Stream progress updates during execution
- `ReportJobCompletion`: Report final job status and artifacts
- `StreamJobLogs`: Real-time log streaming
- `ReportArtifact`: Register generated artifacts
- `DeregisterAgent`: Remove agent from service

**Key Messages:**

- `Agent`: Complete agent definition
- `RegisterAgentRequest/Response`: Agent registration
- `SystemMetrics`: CPU, memory, disk usage
- `JobAssignment`: Job details assigned to agent
- `StreamJobLogsRequest`: Log entry streaming

### 4. Job Management (`control_plane/job.proto`)

**Services:**

- `SubmitJob`: Submit new job for execution
- `GetJobStatus`: Query current job status
- `StreamJobLogs`: Stream logs to client
- `CancelJob`: Cancel running/queued job
- `ListJobs`: List jobs with filtering
- `GetJobDetails`: Get complete job history

**Key Messages:**

- `Job`: Complete job definition with state, metadata, results
- `JobStateTransition`: Track state changes with timestamps
- `JobLogEntry`: Log message with level and source
- `SubmitJobRequest`: Submit job with parameters and constraints

### 5. Artifact Management (`control_plane/artifact.proto`)

**Services:**

- `RegisterArtifact`: Register generated artifact
- `GetArtifact`: Retrieve artifact details
- `ListArtifacts`: List artifacts with filtering
- `GetArtifactLineage`: Get DAG of artifact dependencies
- `PromoteArtifact`: Promote to next stage
- `SignArtifact`: Apply cryptographic signature
- `DeleteArtifact`: Remove or archive artifact
- `GetArtifactDownloadUrl`: Generate presigned download URL

**Key Messages:**

- `Artifact`: Complete artifact metadata
- `ArtifactStage`: Enum for promotion stages (Dev, QA, Staging, Production, Archived)
- `ArtifactLineage`: DAG representation with nodes and edges
- `ArtifactSignature`: Cryptographic signature metadata

### 6. Tenant & Multi-Tenancy (`control_plane/tenant.proto`)

**Services:**

- **TenantService**: Create, read, update, list tenants; set policies
- **ProjectService**: Create, read, update, delete projects; manage policies
- **UserService**: Manage users, roles, and project assignments

**Key Messages:**

- `Tenant`: Top-level organizational boundary with policies and quotas
- `Project`: Logical grouping within tenant
- `User`: User identity with roles (Admin, Operator, Engineer, Viewer, ProjectAdmin)
- `TenantStatus`: Active, Suspended, Trial, Deleted
- `ProjectStatus`: Active, Archived, Deleted

### 7. Control Plane Orchestration (`control_plane/control_plane.proto`)

**Services:**

- `HealthCheck`: System health status
- `GetSystemMetrics`: Aggregate platform metrics
- `GetAgentRegistry`: Query agent status
- `EvaluatePolicy`: Evaluate policies for requests

**Key Messages:**

- `SystemMetrics`: Platform-wide statistics
- `HealthStatus`: Healthy, Degraded, Unhealthy
- `AgentInfo`: Agent summary with capabilities

## Key Design Patterns

### 1. Pull-Based Job Assignment

Agents use the pull model (`GetJobAssignment`) rather than push, allowing:

- Agents to control their own load
- Natural back-pressure
- Compatibility with agents behind firewalls

### 2. Streaming for Real-Time Data

Jobs stream logs and progress using bidirectional gRPC streams:

- Real-time visibility into execution
- Efficient bandwidth usage
- Clean backpressure handling

### 3. Immutable Job Contracts

Once submitted, jobs cannot be modified. This ensures:

- Reproducibility
- Audit trail integrity
- Clear policy enforcement

### 4. Artifact Lineage DAG

All artifacts maintain a Directed Acyclic Graph showing:

- Parent/child relationships
- Full traceability
- Reproducibility verification

### 5. Multi-Tenant Isolation

Strict separation enforced at:

- Token/credential level
- Query result filtering
- Storage and compute isolation

## Job Lifecycle State Machine

```
┌─────────────┐
│  Submitted  │
└──────┬──────┘
       │
       v
┌─────────────┐
│   Queued    │
└──────┬──────┘
       │
       v
┌──────────────┐
│ Dispatched   │
└──────┬───────┘
       │
       v
┌──────────────┐         ┌──────────────┐
│   Running    ├────────→│   Succeeded  │
└──────┬───────┘         └──────────────┘
       │
       ├────────┐
       │        │
       v        v
  ┌────────┐  ┌──────────┐
  │ Failed │  │ TimedOut │
  └────────┘  └──────────┘
       │        │
       └───┬────┘
           v
     ┌──────────────┐
     │   Queued     │ (retry)
     │   (optional) │
     └──────────────┘

Cancel at any state → Cancelled
```

## Generation & SDKs

Proto files are compiled to generate SDKs for multiple languages:

```bash
cd proto/
buf generate  # Generates Go SDKs to agents/go/pkg/smidr-sdk/v1
```

Generated SDKs include:

- Message marshalers/unmarshalers
- gRPC service stubs
- Full type definitions

## Design Principles

1. **Security First**: All communication authenticated and encrypted
2. **Deterministic**: Job execution reproducible across environments
3. **Auditable**: All decisions and state changes logged
4. **Extensible**: Custom metadata and labels for future features
5. **Observable**: Streaming logs, metrics, and state transitions
6. **Scalable**: Designed for multi-tenant, distributed execution

## Future Extensions

Planned proto extensions:

- Enhanced monitoring/observability messages
- Advanced scheduling constraints
- Plugin isolation and versioning (if plugin model returns)
- Advanced artifact promotion workflows
- Custom workflow definitions

## Related Documentation

- [Architecture Overview](../../docs/architecture.md)
- [Job Lifecycle Specification](../../docs/job-lifecycle.md)
- [Agent Responsibility Contract](../../docs/agent-responsibility.md)
- [Artifact Lineage Model](../../docs/artifact-lineage.md)
- [Tenant & Policy Model](../../docs/tenant-policy.md)
- [Control Plane Responsibilities](../../docs/control-plane.md)
- [Job Controller Responsibilities](../../docs/job-controller.md)
