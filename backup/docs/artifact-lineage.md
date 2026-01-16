# Artifact Lineage Model

## Purpose

The Artifact Lineage Model defines how all outputs from jobs are tracked, versioned, and related to inputs for reproducibility, compliance, and auditability. It ensures that any artifact can be traced back to its origin and associated job, agent, and plugin.

---

## Core Concepts

### Artifact

- Immutable output produced by a job
- Associated with:
  - Job ID
  - Agent ID
  - Plugin version
  - Tenant and project
- Stored in the Artifact Store (metadata tracked in the Control Plane)

### Lineage

- Directed Acyclic Graph (DAG) representing dependencies
- Nodes: artifacts
- Edges: input/output relationships (e.g., source files → binaries)
- Supports reproducibility and debugging

### Metadata Tracked

- Artifact ID (unique)
- Parent artifacts (if any)
- Creation timestamp
- Job and agent references
- Plugin version
- Retention and promotion policies

---

## Lineage Policies

### Retention

- Artifacts have a tenant-defined lifespan
- Expired artifacts are archived or deleted according to policy

### Promotion

- Artifacts can be promoted through environments:
  - Development → QA → Production
- Promotions follow explicit rules per tenant/project

### Access Control

- Artifact visibility is restricted by tenant/project
- Only authorized users or systems can retrieve or promote artifacts

---

## Data Flow

```text
   Job Execution
       |
       v
   +-----------+
   |   Agent   |
   +-----------+
       |
       v
   +-----------------+
   | Artifact Store   |
   +-----------------+
       |
       v
   Control Plane (metadata, lineage DAG)
```

- Agents generate artifacts

- Artifacts are stored immutably in the Artifact Store

- Metadata is recorded in the Control Plane for lineage, audit, and policy enforcement

## User Stories

### Embedded Systems Engineer

- I want to trace exactly which source files and configurations produced a specific build artifact.

- I want deterministic reproduction of artifacts in different environments.

### MSP Operator

- I want to enforce retention policies for each customer.
- I want to visualize lineage to confirm that customer builds are isolated.

### Platform Administrator

- I want an auditable DAG of all artifacts.

- I want to enforce promotion and retention policies consistently across tenants.

---

## Notes

- The Artifact Lineage Model is critical for compliance and reproducibility.

- All other modules (agents, plugins, job controller, policy engine) integrate with lineage tracking.

- Any change to artifact handling must preserve DAG integrity and immutability.
