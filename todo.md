# Smidr Initial Development TODO List

This TODO list outlines the first set of tasks to get the **Smidr platform** up and running with foundational infrastructure. Mark tasks as complete as you progress.

---

## 1. Repository Setup

- [X] Initialize monorepo with folder structure
- [X] Add LICENSE (BUSL 1.1)
- [X] Add README with basic description
- [X] Add `docs/` folder with existing documentation
- [X] Setup `.gitignore` and common linter/prettier configs

---

## 2. Protobuf / gRPC Contracts

- [ ] Define `proto/` folder structure
- [ ] Create Job, Agent, and Plugin gRPC messages
- [ ] Generate Go client for agent
- [ ] Generate C# server stubs for control plane
- [ ] Verify round-trip serialization for test job

---

## 3. Control Plane Core

- [ ] Initialize C# ASP.NET Core project
- [ ] Implement Job Controller service:
  - [ ] Job submission
  - [ ] Job lifecycle management
  - [ ] Retry & timeout logic
- [ ] Implement Policy Enforcement module (start with simple hardcoded policies)
- [ ] Setup database models (PostgreSQL) for jobs, tenants, artifacts
- [ ] Add unit tests for Job Controller and models

---

## 4. Agent Core

- [ ] Initialize Go agent project
- [ ] Implement agent registration with Control Plane
- [ ] Implement job execution stub (logging + status reporting)
- [ ] Add gRPC client to talk to Control Plane
- [ ] Implement artifact upload stub
- [ ] Add unit tests for agent logic

---

## 5. Plugin Infrastructure

- [ ] Define Go plugin interface (gRPC)
- [ ] Create simple sample plugin (e.g., "Hello World" job)
- [ ] Integrate plugin execution into agent workflow
- [ ] Test plugin registration and execution via local agent

---

## 6. Development Environment

- [ ] Setup Docker Compose environment:
  - [ ] PostgreSQL
  - [ ] MinIO / artifact storage
  - [ ] Control Plane service
  - [ ] Local agent
- [ ] Verify end-to-end local job submission and execution
- [ ] Document setup instructions in `docs/developer_onboarding.md`

---

## 7. Frontend Skeleton (Optional First Step)

- [ ] Initialize Next.js + TypeScript project
- [ ] Connect to Control Plane API (stub endpoints)
- [ ] Build simple job submission page (mocked data)
- [ ] Verify connection to Control Plane API

---

## 8. Observability (Optional Early Step)

- [ ] Setup logging structure for agent and control plane
- [ ] Add Prometheus metrics stub in Control Plane
- [ ] Add log streaming stub from Agent

---

## 9. Testing & CI/CD

- [ ] Setup test framework for C# and Go projects
- [ ] Write integration test for Job Controller + Agent + Plugin
- [ ] Add basic GitHub Actions workflow:
  - [ ] Build
  - [ ] Run tests
  - [ ] Lint

---

## 10. Documentation Maintenance

- [ ] Keep `docs/developer_onboarding.md` up-to-date
- [ ] Add diagrams and screenshots of workflow as implemented
- [ ] Document plugin interface with example

---

**Notes:**

- Focus first on the **core agent ↔ control plane communication** and job lifecycle.
- Keep early jobs and plugins simple; the goal is to **prove the architecture works**.
- Once foundational layers are solid, expand to **multi-tenant policies**, **artifact lineage**, and **frontend**.

### Loop Diagram

```plaintext
startup
 ├─ load config
 ├─ detect capabilities
 ├─ register
 └─ loop
     ├─ heartbeat
     ├─ poll for job
     ├─ execute job
     ├─ stream logs
     └─ report result
```
