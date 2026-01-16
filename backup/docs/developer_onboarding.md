# Smidr Developer Onboarding Guide

Welcome to the Smidr platform! This document is your **single point of reference** to understand the architecture, repository structure, job lifecycle, plugin system, and development workflow. It is intended for new developers, contributors, and embedded systems teams looking to build, extend, or deploy Smidr.

---

## Table of Contents

- [Smidr Developer Onboarding Guide](#smidr-developer-onboarding-guide)
  - [Table of Contents](#table-of-contents)
  - [Platform Overview](#platform-overview)
  - [Architecture Overview](#architecture-overview)
  - [Repository Structure](#repository-structure)
  - [Job Lifecycle](#job-lifecycle)
  - [Agent \& Control Plane Contracts](#agent--control-plane-contracts)
  - [Tenant \& Policy Model](#tenant--policy-model)
  - [Plugin System](#plugin-system)
  - [API \& Communication](#api--communication)
  - [Artifact Management](#artifact-management)
  - [Development Environment Setup](#development-environment-setup)
  - [Testing Strategy](#testing-strategy)
  - [Deployment \& Observability](#deployment--observability)
  - [Contribution Guidelines](#contribution-guidelines)
  - [BUSL Licensing Overview](#busl-licensing-overview)
  - [Resources \& References](#resources--references)

---

## Platform Overview

Smidr is a **multi-tenant, cross-platform orchestration platform** designed for building, deploying, and managing binaries, Yocto images, Debian live builds, and other system-level tasks.

**Key Features:**

- Secure, deterministic job execution
- Multi-tenant architecture with policy enforcement
- Plugin extensibility via gRPC
- Artifact lineage tracking with promotion & retention
- Observability & monitoring integrations
- Enterprise-ready, strongly typed backend

---

## Architecture Overview

```text
       +------------------------+
       |      Frontend          |
       |  (React + TypeScript)  |
       +------------------------+
                 |
          REST / GraphQL
                 |
       +------------------------+
       |    Control Plane       |
       |  (C# ASP.NET Core)    |
       +------------------------+
       | Job Controller / Policy|
       +------------------------+
                 |
             gRPC / TLS
                 |
   +-------------------------------+
   |           Agent               |
   |       (Go binary)            |
   +-------------------------------+
   |   Plugins (gRPC processes)   |
   +-------------------------------+
                 |
      Artifact Store (MinIO/PostgreSQL)
```

**Layer Principles:**

- **Agents**: Execute jobs safely on target systems.
- **Control Plane**: Orchestrates jobs, enforces policies, tracks tenants.
- **Job Controller**: Manages job assignment, lifecycle, retries.
- **Plugins**: Extend functionality without touching agent core.
- **Frontend**: Dashboard for monitoring, job submission, tenant/project management.
- **Artifact Store**: Immutable storage with lineage metadata.

---

## Repository Structure

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

**Guiding Principles:**

- Clear separation of concerns
- Typed contracts between layers (`proto/`)
- Independent tests per layer
- Modular plugins for extensibility
- Deployment-ready Docker/K8s manifests

---

## Job Lifecycle

**States of a Job:**

- `Pending` → waiting for scheduling
- `Scheduled` → assigned to an agent
- `Running` → actively executing
- `Succeeded` → completed successfully
- `Failed` → execution failed
- `Cancelled` → terminated manually
- `Retrying` → retry policy triggered

**Responsibilities:**

- **Agent:** Executes job steps, streams logs, reports status.
- **Job Controller:** Assigns jobs, tracks lifecycle, handles retries/timeouts.
- **Control Plane:** Enforces policies, persists job and artifact metadata.

---

## Agent & Control Plane Contracts

- **Agent Contract:** Defines RPC endpoints, job execution protocol, log streaming, artifact upload.
- **Control Plane Contract:** Defines job submission API, policy enforcement API, agent registration, artifact retrieval.
- **Protobuf Definitions:** Stored in `proto/` for cross-language type safety.

---

## Tenant & Policy Model

**Tenants:** Logical separation of users, projects, and resources.

**Policies:** Rules controlling:

- Job execution privileges
- Resource quotas
- Artifact retention and promotion
- Security constraints

**User Stories Examples:**

- Tenant admin can create a new project with its own policies.
- Agents report metrics only to the tenant they belong to.
- Policy engine prevents unauthorized builds or deployments.

---

## Plugin System

- Plugins are **external processes** communicating with agents via gRPC.
- Can be implemented in Go, Python, or shell.
- Plugins extend:
  - Build pipelines (Yocto, Debian, custom binaries)
  - Monitoring and sensor data collection
  - Custom automation tasks
- Versioned per tenant and major Smidr release.

---

## API & Communication

| Source            | Destination   | Protocol       | Notes                                           |
| ----------------- | ------------- | -------------- | ----------------------------------------------- |
| Agent             | Control Plane | gRPC over TLS  | Bi-directional job control, streaming logs      |
| Frontend          | Control Plane | REST / GraphQL | Dashboards, job submissions                     |
| Plugins           | Agent         | gRPC           | Isolated execution, logging, artifact reporting |
| Internal services | Each other    | gRPC / HTTP    | Typed service contracts                         |

---

## Artifact Management

- **Immutable artifact storage** using MinIO or S3-compatible backend.
- **Metadata stored in PostgreSQL** for artifact lineage.
- **Promotion policies** allow moving artifacts between environments (dev → staging → prod).

---

## Development Environment Setup

1. Clone the repo.
2. Install Go, C#, Node.js/TypeScript.
3. Start PostgreSQL and MinIO (via Docker Compose in `deployments/docker/`).
4. Launch Control Plane and Job Controller services.
5. Register an agent locally.
6. Run frontend with `npm run dev` or `yarn dev`.
7. Submit a test job via frontend or CLI.

---

## Testing Strategy

- **Unit Tests:** Each layer independently (`agents`, `control-plane`, `plugins`).
- **Integration Tests:** Agent ↔ Control Plane ↔ Plugin workflows.
- **End-to-End Tests:** Full workflow from job submission to artifact collection.
- **CI/CD:** Automated tests triggered per commit via scripts in `scripts/`.

---

## Deployment & Observability

- **Docker Compose:** Local dev environment.
- **Kubernetes / Helm:** Production-ready scaling.
- **Metrics & Logs:** Prometheus + Grafana dashboards.
- **Plugins:** Can expose metrics via gRPC to agent, reported to control plane.

---

## Contribution Guidelines

- Use semantic versioning.
- Follow type conventions per language:
  - Go: Agents & plugins
  - C#: Control Plane & Job Controller
  - TypeScript: Frontend
- Write tests for all new features.
- Update relevant documentation in `docs/`.
- Submit PRs with review and automated test coverage.

---

## BUSL Licensing Overview

- Smidr uses **BUSL 1.1**: proprietary for 3 years per major version.
- Each new major version resets the 3-year proprietary period.
- After the Change Date, that major version is licensed under MIT.
- See `LICENSE` file for full BUSL terms.

---

## Resources & References

- `docs/architecture.md` – Detailed architecture diagrams
- `docs/job_lifecycle.md` – Full lifecycle and controller responsibilities
- `docs/tenant_policy.md` – Tenant & policy model
- `docs/blueprint.md` – Full implementation blueprint
- `proto/` – gRPC contracts for agents and plugins

---

**Welcome aboard!** Start with the Control Plane and agent local setup, then incrementally explore plugin development and frontend integration. Use this guide as your central reference for design, workflow, and best practices.
