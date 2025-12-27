# Architecture Overview

## Purpose

This document provides a high-level view of the platform, including the interaction between agents, the control plane, users, jobs, and artifacts. It serves as the guiding blueprint for all subsequent design and development efforts.

The platform is designed to support:

- Embedded systems and custom Linux/Yocto builds
- MSP-style multi-tenant operations
- Reproducible, auditable job execution
- Extensible plugins without modifying core services

---

## Core Components

### Agent (Edge Execution Node)

- Installed on Windows, Linux, or macOS hosts
- Executes jobs in isolated environments
- Reports logs, progress, and artifacts to the Control Plane
- Enforces job-level policies received from Control Plane

### Control Plane API

- Central authority for policy enforcement (runs on AWS Lambda)
- Orchestrates job lifecycle via REST + HMAC authentication
- Maintains artifact lineage and audit logs (DynamoDB)
- Provides multi-tenant separation via tenant context

### Job Controller

- Serverless Lambda function matches jobs to capable agents
- Tracks job states and retries
- Distributes jobs via SQS queue; agents pull jobs on demand
- Enforces job-specific constraints and policies

### Artifact Store

- Stores job outputs and logs (AWS S3)
- Maintains immutable references and lineage tracking (DynamoDB)
- Supports promotion, retention, and signing rules enforced by Control Plane

### Policy Engine

- Evaluates execution, agent, artifact, and security policies
- Applies tenant/project-scoped rules
- Returns deterministic decisions to Job Controller

### Web UI / Frontend

- Written in TypeScript and React
- Provides job submission, monitoring, and management
- Supports tenant and project dashboards
- Provides plugin configuration and monitoring insights
- Communicates with Control Plane via REST + HTTPS

---

## Data Flows

### Job Submission Flow

1. User submits a job via the frontend or REST API.
2. Control Plane (Lambda) validates permissions and evaluates policies.
3. Job is placed on SQS queue for agent pickup.
4. Agent pulls the job, executes it, and reports progress.
5. Artifacts and logs are uploaded to S3 and registered in DynamoDB.
6. Completion status is communicated back to the user and logged for audit.

### Artifact Lineage

- All outputs are tagged with job ID, agent ID, and plugin version.
- Lineage ensures reproducibility and traceability.

---

## Design Principles

- **Secure by Default:** Multi-tenant isolation, least privilege, sandboxed agents.
- **Deterministic Execution:** Jobs and plugins execute in reproducible environments.
- **Extensible:** Plugins allow new build and monitoring capabilities without core changes.
- **Auditable:** Every state transition, artifact, and policy decision is logged.
- **Policy-Driven:** Policies define allowed operations, agent capabilities, artifact handling, and security requirements.

---

## Project Structure

```text
smidr/
├── LICENSE
├── README.md
├── docs/
│   ├── architecture.md
│   ├── job_lifecycle.md
│   ├── agent_contract.md
│   ├── control_plane_contract.md
│   ├── artifact_lineage.md
│   ├── tenant_policy.md
│   └── blueprint.md
├── agents/
│   ├── go/                 # Go-based cross-platform agent
│   │   ├── cmd/            # CLI entry points for agent
│   │   ├── internal/       # Core agent logic
│   │   └── plugins/        # gRPC plugin interface helpers
│   └── rust/               # Optional Rust implementation for performance
├── control-plane/
│   ├── src/
│   │   ├── api/            # REST / GraphQL endpoints
│   │   ├── job-controller/ # Job scheduling and lifecycle
│   │   ├── policies/       # Policy enforcement engine (OPA integration)
│   │   ├── services/       # Business logic services
│   │   └── models/         # DB models, schemas
│   └── tests/
├── frontend/
│   ├── nextjs/             # Next.js + React frontend
│   │   ├── components/
│   │   ├── pages/
│   │   ├── hooks/
│   │   └── styles/
├── plugins/                # External plugins for builds or monitoring
│   ├── go/
│   ├── python/
│   └── bash/
├── artifacts/              # Local dev environment artifact store
├── proto/                  # Shared protobuf definitions (gRPC)
├── deployments/
│   ├── docker/             # Docker Compose configs, Dockerfiles
│   ├── k8s/                # Kubernetes manifests
│   └── helm/               # Helm charts for production deployment
├── scripts/                # Build, test, release scripts
└── tools/                  # Dev tools, linters, CI/CD helpers

```

---

## Conceptual Diagram

```text
                +-------------------+
                |   Web Frontend    |
                +---------+---------+
                          |
                          v
                +-------------------+
                |   Control Plane   |
                | (Policy Engine &  |
                |  Job Controller)  |
                +---------+---------+
                          |
         -----------------+-----------------
         |                                 |
         v                                 v
   +-------------+                   +-------------+
   |   Agent 1   |                   |   Agent N   |
   | (Linux/mac) |                   | (Windows)   |
   +------+------+                   +------+------+
          |                                 |
          v                                 v
     +---------+                       +---------+
     | Artifact|                       | Artifact|
     +---------+                       +---------+
```

## User Stories

### Embedded Systems Engineer

- I want builds for multiple hardware targets without modifying the core platform.

- I want deterministic builds to ensure reproducibility.

### MSP Operator

- I want to manage multiple customer tenants with strict isolation.

- I want to enforce resource and job type limits per tenant.

### Platform Administrator

- I want complete audit logs of job execution and artifact creation.

- I want to manage plugin versions and agent registrations safely.
