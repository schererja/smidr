# Smidr Platform

Smidr is a multi-tenant, cross-platform orchestration platform for building, deploying, and managing binaries, Yocto images, Debian live builds, and other system tasks. It is designed to simplify embedded system workflows, help MSPs manage client systems, and provide reproducible, auditable artifact pipelines.

---

## Table of Contents

- [Smidr Platform](#smidr-platform)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [Key Features](#key-features)
  - [Architecture](#architecture)
  - [Components](#components)
    - [1. Agent](#1-agent)
    - [2. Control Plane](#2-control-plane)
    - [3. Job Controller](#3-job-controller)
    - [4. Artifact Store](#4-artifact-store)
    - [5. Plugins](#5-plugins)
    - [6. Frontend](#6-frontend)
  - [Communication \& Protocols](#communication--protocols)
  - [Plugin System](#plugin-system)
  - [Deployment](#deployment)
  - [Getting Started](#getting-started)
  - [Contributing](#contributing)
  - [License](#license)

---

## Overview

Smidr provides a **secure and extensible platform** for automating build and monitoring workflows across heterogeneous systems. It consists of:

- **Agents**: Installed on Windows, Linux, or macOS machines to execute jobs.
- **Control Plane**: Central orchestrator enforcing policies, managing multi-tenant operations, and coordinating jobs.
- **Job Controller**: Assigns jobs to agents and tracks lifecycle states.
- **Artifact Store**: Immutable storage with full lineage tracking.
- **Plugins**: Extend the platform for builds, monitoring, and custom workflows.
- **Frontend**: TypeScript/React dashboards for monitoring and management.

The platform is designed for **reproducibility, security, and scalability**.

---

## Key Features

- Multi-tenant architecture with strict isolation
- Deterministic job execution with policy enforcement
- Plugin extensibility via gRPC for cross-platform compatibility
- Artifact lineage tracking with promotion and retention policies
- Streaming logs and progress reporting
- Observability with Prometheus/Grafana
- Enterprise-ready stack with strong typing and maintainability

---

## Architecture

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

---

## Components

### 1. Agent

- Executes jobs safely and reproducibly
- Reports logs and artifacts
- Runs on Linux, Windows, macOS
- Lightweight, statically compiled Go binary

### 2. Control Plane

- Multi-tenant orchestrator
- Policy enforcement and job scheduling
- Tracks agent health, jobs, and artifacts
- Provides REST/GraphQL API for frontend

### 3. Job Controller

- Handles job assignments and lifecycle transitions
- Applies retry and timeout policies
- Integrates tightly with Control Plane and Agent

### 4. Artifact Store

- Immutable object storage (S3-compatible MinIO)
- Relational DB (PostgreSQL) for metadata and lineage
- Retention and promotion policies

### 5. Plugins

- Extend functionality (builds, monitoring)
- Isolated gRPC-based execution
- Versioned and tenant-scoped

### 6. Frontend

- Dashboard and monitoring UI
- Job submission and management
- Tenant/project visibility and reporting

---

## Communication & Protocols

| Source | Destination | Protocol | Notes |
|--------|------------|---------|-------|
| Agent | Control Plane | gRPC over TLS | Bi-directional job control, streaming logs |
| Frontend | Control Plane | REST / GraphQL | Dashboards, job submissions |
| Control Plane | Artifact Store | S3 API / DB | Artifact uploads and metadata |
| Plugins | Agent | gRPC | Isolated execution, logging, artifact reporting |
| Internal services | Each other | gRPC / HTTP | Typed service contracts |

---

## Plugin System

- Plugins are **external processes** communicating via gRPC.
- Inputs: job context, parameters, scoped credentials
- Outputs: logs, status, artifact references
- Optional: native Go plugin loading for Linux-only agents
- Ensures **cross-platform safety, isolation, and reproducibility**

---

## Deployment

- Agents: distributed binaries per platform
- Control Plane & Job Controller: Dockerized, scalable via Kubernetes
- Artifact Store: MinIO cluster + PostgreSQL DB
- Observability: Prometheus + Grafana
- Plugins: containerized or process-isolated execution

---

## Getting Started

1. Clone the repository
2. Build the Control Plane and Agent binaries
3. Start PostgreSQL and MinIO for artifact storage
4. Launch Control Plane and Job Controller services
5. Register Agents
6. Use Frontend to create tenants, projects, and submit jobs

> Detailed installation instructions and environment setup are provided in the `docs/` folder.

---

## Contributing

We welcome contributions from the community:

- Follow coding conventions (C# for Control Plane, Go for Agent, TypeScript/React for frontend)
- Write tests for new features or plugins
- Document architecture changes in Markdown files

---

## License

[MIT License](LICENSE)

---

**Notes:**
This README provides a high-level overview suitable for developers, MSPs, and embedded systems teams. For detailed implementation, see the Markdown documentation in the `docs/` folder covering:

- Architecture Overview
- Job Lifecycle Specification
- Job Controller Responsibilities
- Artifact Lineage Model
- Tenant & Policy Model
- Agent Responsibility Contract
- Control Plane Responsibility Contract
- Plugin Interface Specification
- Implementation Blueprint
