# Smidr Platform: Implementation Blueprint

## Purpose

This document outlines the recommended technology stack, frameworks, languages, storage, and communication protocols for the Smidr platform. It aligns with the architectural design, multi-tenant considerations, and plugin extensibility.

---

## 1. Edge Agent (Execution Node)

**Responsibilities:** Execute jobs, enforce job policies, report logs and artifacts.

| Attribute | Recommendation | Notes |
|-----------|----------------|-------|
| Language | Go | Cross-platform static binaries, minimal runtime, strong concurrency |
| Sandbox / Isolation | OS-level containers / job objects / chroot | Prevents cross-job or cross-tenant contamination |
| Communication | gRPC over TLS | Bi-directional streaming for logs, status, and job control |
| Logging | Structured JSON | Easy ingestion by Control Plane |
| Packaging | Single binary per platform | Simplifies deployment |
| Optional | Rust | Safer memory management, but heavier build |

**Notes:** Agent must be lightweight, idempotent, and capable of self-registering with Control Plane.

---

## 2. Control Plane

**Responsibilities:** Central orchestration, policy enforcement, multi-tenant governance, API gateway.

| Attribute | Recommendation | Notes |
|-----------|----------------|-------|
| Language | C# (.NET 8, ASP.NET Core) | Strong typing, async concurrency, enterprise-ready |
| Web API | REST + GraphQL | REST for CRUD, GraphQL for dashboards |
| Internal Communication | gRPC | Typed contracts with Job Controller and agents |
| Database | PostgreSQL | Stores job metadata, tenants/projects, artifact lineage, policies |
| Policy Engine | Open Policy Agent (OPA) | Deterministic, versionable, reusable across tenants |
| Authentication | OAuth2 / JWT | Scoped tokens for agents and frontend |
| Observability | Prometheus + Grafana | Metrics, health monitoring, job metrics |
| Deployment | Docker → Kubernetes | Containerized scaling |

---

## 3. Job Controller

**Responsibilities:** Assign jobs to agents, enforce retries, track lifecycle.

| Attribute | Recommendation | Notes |
|-----------|----------------|-------|
| Language | C# (within Control Plane) | Keeps stack consistent; may be isolated later |
| Communication | gRPC | Typed contracts with agents, policy engine, artifact store |
| Retry & Scheduling | Configurable policies in PostgreSQL | Deterministic, per tenant/project |
| Optional | Go microservice | For high-concurrency scenarios |

---

## 4. Artifact Store & Lineage

**Responsibilities:** Immutable artifact storage, lineage DAG, retention/promotion policies.

| Attribute | Recommendation | Notes |
|-----------|----------------|-------|
| Object Storage | MinIO (S3-compatible) | Multi-tenant safe, scalable |
| Metadata DB | PostgreSQL | Tracks lineage DAG, artifact metadata, retention rules |
| Lineage Representation | PostgreSQL adjacency list / DAG table | Optional Neo4j if complex graph queries |
| Access Control | Enforced via Control Plane | Prevent cross-tenant access |

---

## 5. Policy Engine

**Responsibilities:** Evaluate execution, agent, artifact, and security policies.

| Attribute | Recommendation | Notes |
|-----------|----------------|-------|
| Engine | OPA | Versioned, deterministic, decoupled from business logic |
| Policy Format | Rego | Declarative, auditable, JSON-compatible |
| Integration | Control Plane & Job Controller | Decisions returned as immutable job contracts |

---

## 6. Plugin Interface

**Responsibilities:** Extend platform with builds or monitoring tools.

| Attribute | Recommendation | Notes |
|-----------|----------------|-------|
| Execution | Separate process or container | Ensures cross-platform support, isolation, safe execution |
| Language | Go, Python, Bash, or compiled binaries | Depends on task |
| Communication | gRPC | Typed, streaming logs/status/artifacts |
| Versioning | Immutable per tenant/project | Ensures reproducible builds |
| Input/Output | Structured JSON, job context, artifact refs | Standardized interface |

**Notes:** Go native plugins are optional Linux optimization; primary model is external gRPC-based plugins for safety and portability.

---

## 7. Frontend

**Responsibilities:** Job submission, monitoring, artifact promotion, tenant dashboards.

| Attribute | Recommendation | Notes |
|-----------|----------------|-------|
| Language | TypeScript | Modern, strongly typed |
| Framework | React + Next.js | Modular UI, SSR support |
| Communication | REST / GraphQL | Dashboard flexibility, CRUD operations |
| Authentication | JWT / OAuth2 | Scoped by tenant/project |

---

## 8. Communication Overview

| Source | Destination | Protocol | Notes |
|--------|------------|---------|-------|
| Agent | Control Plane | gRPC over TLS | Bi-directional logs, job commands |
| Control Plane | Agent | gRPC | Job dispatch, policy contracts |
| Frontend | Control Plane | REST / GraphQL | Job submission, monitoring |
| Control Plane | Artifact Store | S3 API / DB | Artifact upload, metadata |
| Plugins | Agent | gRPC | Isolated execution reporting |
| Internal Services | Each other | gRPC / HTTP | Typed service contracts |

---

## 9. Deployment & Scaling

- **Agents:** Distributed, cross-platform binaries. Auto-register with Control Plane.
- **Control Plane & Job Controller:** Containerized services; horizontal scaling with Kubernetes.
- **Artifact Store:** MinIO cluster, metadata DB replicated for high availability.
- **Policy Engine:** Sidecar or embedded library; versioned policies.
- **Monitoring:** Prometheus/Grafana; log aggregation via ELK or Loki.

---

## Summary of Recommended Stack

| Component | Language / Framework | Storage | Protocol / API |
|-----------|-------------------|--------|----------------|
| Agent | Go | Local / Artifact Store | gRPC over TLS |
| Control Plane | C# ASP.NET Core | PostgreSQL | REST / GraphQL + gRPC |
| Job Controller | C# | PostgreSQL | gRPC |
| Artifact Store | MinIO + PostgreSQL | Object + relational | S3 API + DB |
| Policy Engine | OPA (Rego) | N/A | JSON over gRPC/HTTP |
| Plugin | Go/Python/Bash | Agent-local / Artifact Store | gRPC |
| Frontend | React + TypeScript | N/A | REST / GraphQL |

---

## Notes

- Strong typing end-to-end with gRPC, Protobuf, JSON schemas, and relational DB schemas.
- Sandbox and isolation for plugins and jobs.
- Multi-tenant support enforced at Control Plane level.
- Auditability and reproducibility preserved across all components.
