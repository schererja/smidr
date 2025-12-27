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

- **Agents**: Installed on Windows, Linux, or macOS machines to execute jobs. Written in Go for portability.
- **Control Plane**: Serverless orchestrator (TypeScript/Lambda) enforcing policies, managing multi-tenant operations, and coordinating jobs via REST APIs.
- **Job Controller**: Serverless Lambda functions assign jobs to agents via SQS and track lifecycle states.
- **Artifact Store**: Immutable storage with full lineage tracking (S3/DynamoDB).
- **Plugins**: Extend agents for builds, monitoring, and custom workflows.
- **Frontend**: TypeScript/React dashboards for monitoring and management.

The platform is designed for **reproducibility, security, and scalability**.

---

## Key Features

- Multi-tenant architecture with strict isolation and HMAC authentication
- Deterministic job execution with policy enforcement
- Plugin extensibility via gRPC for cross-platform compatibility
- Artifact lineage tracking with promotion and retention policies
- Real-time log streaming and progress reporting
- Serverless infrastructure for cost efficiency and auto-scaling
- Enterprise-ready stack with strong typing and maintainability

---

## Architecture

```text
       +------------------------+
       |      Frontend          |
       |  (React + TypeScript)  |
       +------------------------+
                 |
             REST / HTTPS
                 |
   +-------------------------------+
   |      AWS API Gateway         |
   +-------------------------------+
                 |
   +-------------------------------+
   |  Control Plane (Lambda)      |
   |   (TypeScript/Node.js)       |
   +-------------------------------+
         |                  |
       SQS            DynamoDB
         |                  |
   +-------------------------------+
   |           Agent               |
   |       (Go binary)            |
   +-------------------------------+
   |   Plugins (gRPC processes)   |
   +-------------------------------+
                 |
      Artifact Store (S3/DynamoDB)
```

---

## Components

### 1. Agent

- Executes jobs safely and reproducibly
- Reports logs and artifacts
- Runs on Linux, Windows, macOS
- Lightweight, statically compiled Go binary

### 2. Control Plane

- Multi-tenant orchestrator running on AWS Lambda
- Policy enforcement and job scheduling via REST APIs
- Tracks agent health, jobs, and artifacts using DynamoDB
- Scales automatically with workload; pay-per-request pricing
- HMAC authentication for agent-to-control-plane communication

### 3. Job Controller

- Handles job assignments and lifecycle transitions
- Applies retry and timeout policies
- Integrates tightly with Control Plane and Agent

### 4. Artifact Store

- Immutable object storage (AWS S3)
- NoSQL metadata and lineage tracking (DynamoDB)
- Retention and promotion policies enforced by Control Plane

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
| Agent | Control Plane | REST + HMAC | Job claims, heartbeat, completion reports |
| Frontend | Control Plane | REST + HTTPS | Dashboards, job submissions |
| Control Plane | Artifact Store | S3 API / DynamoDB | Artifact uploads and metadata |
| Control Plane | Job Queue | SQS | Job dispatch and agent pull-based claims |
| Plugins | Agent | gRPC | Isolated execution, logging, artifact reporting |
| Internal services | Each other | Lambda / SQS | Event-driven serverless architecture |

---

## Plugin System

- Plugins are **external processes** communicating via gRPC.
- Inputs: job context, parameters, scoped credentials
- Outputs: logs, status, artifact references
- Optional: native Go plugin loading for Linux-only agents
- Ensures **cross-platform safety, isolation, and reproducibility**

---

## Deployment

- Agents: distributed binaries per platform (Windows, Linux, macOS)
- Control Plane: Serverless on AWS Lambda with API Gateway
- Job Queue: AWS SQS for job distribution
- Artifact Store: AWS S3 with DynamoDB metadata
- Observability: CloudWatch Logs & Metrics, optional X-Ray tracing
- Plugins: containerized or process-isolated execution on agent hosts

---

## Getting Started

1. Clone the repository
2. Build the Agent binaries for your platforms
3. Set up AWS infrastructure (S3, DynamoDB, SQS, Lambda, API Gateway) via SAM template
4. Deploy Control Plane Lambda functions
5. Configure tenant HMAC secrets in AWS Secrets Manager
6. Register Agents with Control Plane
7. Use Frontend to create tenants, projects, and submit jobs

> Detailed installation instructions and AWS deployment guide are provided in the `docs/` folder.

---

## Contributing

We welcome contributions from the community:

- Follow coding conventions (TypeScript for Control Plane/Lambda, Go for Agent, TypeScript/React for Frontend)
- Write tests for new features or plugins
- Document architecture changes in Markdown files
- Follow serverless best practices (Lambda optimization, DynamoDB design, SQS batching)

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
