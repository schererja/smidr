# Implementation Roadmap

## Overview

This document outlines the phased implementation plan for Yggdrasil, from v0.5 (MVP) through v2.0+. Each phase builds upon the previous, delivering incremental value while maintaining stability.

---

## Version Strategy

| Version | Focus                          | Timeline    | Status   |
| ------- | ------------------------------ | ----------- | -------- |
| v0.5    | MVP - Core Foundation          | 14-16 weeks | Planning |
| v1.0    | Billing & Task Execution       | +8 weeks    | Planned  |
| v1.5    | Plugin SDK & Advanced Features | +12 weeks   | Planned  |
| v2.0    | Enterprise & Scale             | +16 weeks   | Future   |

**Key Principle**: Ship early, iterate fast. v0.5 should be functional enough for internal use (dogfooding).

---

## v0.5: MVP - Core Foundation

### Goal

Build the foundational platform with essential MSP features: ticketing, time tracking, client management, and basic agent monitoring.

### Duration: 14-16 Weeks

### Features

#### 1. Multi-Tenant Infrastructure (Week 1-2)

- [ ] PostgreSQL setup with TimescaleDB extension
- [ ] NATS message queue setup
- [ ] Database schema implementation (golang-migrate migrations)
- [ ] Row-level security (RLS) for multi-tenancy
- [ ] Seed data for development
- [ ] sqlc configuration and initial query generation

**Deliverables**:

- Docker Compose environment for local development
- Initial database migrations
- Multi-tenant test data
- sqlc generated code for initial tables

---

#### 2. Authentication & Authorization (Week 2-3)

- [ ] User model and registration
- [ ] JWT-based authentication (access + refresh tokens)
- [ ] Password hashing with bcrypt
- [ ] Role-based access control (RBAC)
  - MSP Admin, MSP Tech, MSP Manager, Client User
- [ ] Permission middleware for API endpoints
- [ ] Login/logout REST endpoints

**Deliverables**:

- `/auth/login`, `/auth/refresh`, `/auth/logout` endpoints
- User management endpoints (CRUD)
- Authentication middleware using Chi

---

#### 3. Core Data Models & Services (Week 3-5)

**Models** (see [Data Model](./02-data-model.md)):

- [ ] MSPTenant
- [ ] SubOrg
- [ ] Client
- [ ] System
- [ ] User
- [ ] Ticket
- [ ] TicketComment
- [ ] TimeEntry
- [ ] Agent
- [ ] MetricData (TimescaleDB hypertable)

**Services** (service layer pattern):

- [ ] ClientService (CRUD, filtering, search)
- [ ] TicketingService (CRUD, status transitions, assignments)
- [ ] TimeTrackingService (CRUD, billable/non-billable)
- [ ] AgentService (registration, approval, heartbeat)
- [ ] MetricsService (ingestion, querying)
- [ ] UserService (CRUD, role management)

**Deliverables**:

- Complete service layer implementation
- Unit tests for all services (80%+ coverage)
- sqlc queries for all data operations
- Repository pattern with type-safe SQL

---

#### 4. REST API Implementation (Week 5-6)

**Endpoints** (see [API Contract](./03-api-contract.md)):

- [ ] `/tickets` - List, create, update, get, add comments
- [ ] `/time-entries` - List, create, update, delete
- [ ] `/clients` - List, create, update, get
- [ ] `/systems` - List, get, update
- [ ] `/systems/{id}/metrics` - Query metrics
- [ ] `/agents` - List, approve, revoke
- [ ] `/users` - List, create, update (admin only)
- [ ] `/sub-orgs` - List, create

**Features**:

- Pagination for all list endpoints
- Filtering and sorting
- OpenAPI/Swagger docs (auto-generated)
- Error handling and validation (Go structs + tags)

**Deliverables**:

- Go/Chi application with all v0.5 endpoints
- Integration tests for API (Go testing + httptest)
- API documentation via OpenAPI spec

---

#### 5. gRPC API for Agents (Week 6-7)

**Protobuf Definitions**:

- [ ] `Register` - Agent registration
- [ ] `Heartbeat` - Keep-alive
- [ ] `SendMetrics` - Metrics ingestion

**Implementation**:

- [ ] gRPC server in Go (google.golang.org/grpc)
- [ ] mTLS authentication
- [ ] Tenant extraction from certificate
- [ ] Rate limiting for metrics ingestion

**Deliverables**:

- gRPC service running on port 50051
- Proto file in `packages/proto/agent/v1/agent.proto`
- Agent authentication via mTLS

---

#### 6. Agent Implementation (Week 7-9)

**Core Agent**:

- [ ] Go binary with viper configuration
- [ ] gRPC client for control plane communication
- [ ] mTLS certificate generation and authentication
- [ ] Heartbeat ticker (every 60s)
- [ ] Structured logging with slog

**Plugin System**:

- [ ] Plugin interface (sdk.Plugin, sdk.MetricsPlugin)
- [ ] Plugin registry
- [ ] Plugin initialization and lifecycle
- [ ] Metrics collection scheduler

**System Metrics Plugin**:

- [ ] CPU usage (gopsutil)
- [ ] Memory usage
- [ ] Disk usage (per partition)
- [ ] Network I/O (basic)

**Deliverables**:

- Agent binary for Linux, Windows, macOS (goreleaser)
- System metrics plugin collecting basic metrics
- Metrics sent to control plane every 60s

---

#### 7. Frontend - Core UI (Week 8-10)

**Setup**:

- [ ] React + TypeScript + Vite project
- [ ] TailwindCSS + shadcn/ui components
- [ ] React Router v6 setup
- [ ] TanStack Query for data fetching
- [ ] React Hook Form + Zod for forms

**Pages**:

- [ ] Login page
- [ ] Dashboard (ticket stats, agent status overview)
- [ ] Tickets list (with filtering, sorting, pagination)
- [ ] Ticket detail (with comments and time entries)
- [ ] Ticket create/edit form
- [ ] Clients list
- [ ] Client detail (with systems)
- [ ] Systems list
- [ ] System detail (with basic metrics chart)
- [ ] Time entries list
- [ ] Agents list (with approve/revoke)
- [ ] Users management (admin only)

**Components**:

- [ ] Navbar with user menu
- [ ] Sidebar navigation
- [ ] Data tables (TanStack Table)
- [ ] Form components (shadcn/ui)
- [ ] Status badges
- [ ] Loading states

**Deliverables**:

- Functional web application
- Responsive design (mobile-friendly)
- Authentication flow (login, logout)
- All v0.5 features accessible via UI

---

#### 8. Testing & Documentation (Week 10-11)

**Testing**:

- [ ] Unit tests for services (Go testing)
- [ ] Integration tests for API (Go testing + httptest)
- [ ] Agent plugin tests (Go testing)
- [ ] Frontend component tests (Vitest + Testing Library)
- [ ] Manual E2E testing (scripted test cases)

**Documentation**:

- [ ] README with setup instructions
- [ ] API documentation (OpenAPI spec)
- [ ] Agent installation guide
- [ ] User guide (basic usage)
- [ ] Architecture docs (this folder!)

**Deliverables**:

- 80%+ test coverage for backend
- 70%+ test coverage for frontend
- Comprehensive documentation

---

#### 9. Deployment & Polish (Week 11-12)

**Deployment**:

- [ ] Docker images for control plane and web
- [ ] Docker Compose for production-like environment
- [ ] Environment variable configuration
- [ ] Database backup strategy

**Polish**:

- [ ] Error handling improvements
- [ ] Loading states and UX polish
- [ ] Performance optimization
- [ ] Security audit (basic)
- [ ] Linting and formatting (golangci-lint, eslint, prettier)

**Deliverables**:

- Docker Compose stack that "just works"
- Agent binaries for all platforms
- v0.5 ready for internal dogfooding

---

### v0.5 Acceptance Criteria

**Core Functionality**:

- ✅ Create MSP tenant and users
- ✅ Add clients and systems
- ✅ Install and approve agents on systems
- ✅ Agents collect and send metrics
- ✅ Create tickets and assign to techs
- ✅ Track time on tickets (billable/non-billable)
- ✅ View system metrics in UI
- ✅ Multi-tenant isolation enforced

**Non-Functional**:

- ✅ API response time <200ms (p95)
- ✅ Support 100 concurrent agents
- ✅ 80%+ test coverage
- ✅ Zero security vulnerabilities (Snyk/Dependabot)

---

## v1.0: Billing & Task Execution

### Version 1.0 Goal

Add invoicing/billing capabilities and agent task execution (script running, Docker commands).

### Duration: +8 Weeks (Weeks 13-20)

### Version 1.0 Features

#### 1. Contracts & SLAs (Week 13-14)

- [ ] Contract model (fixed, hourly, hybrid, prepaid hours)
- [ ] SLA definitions (response time, resolution time)
- [ ] Contract-client association
- [ ] Contract renewal logic
- [ ] UI for contract management

#### 2. Invoicing & Billing (Week 14-16)

- [ ] Invoice model and line items
- [ ] Invoice generation from time entries
- [ ] Invoice status workflow (draft, sent, paid, overdue)
- [ ] PDF invoice generation
- [ ] Invoice email sending
- [ ] Payment status tracking
- [ ] UI for invoice management

#### 3. Agent Task Execution (Week 16-18)

- [ ] Task model and NATS queue integration
- [ ] Script execution plugin (bash, PowerShell)
- [ ] Docker plugin (run containers, execute commands)
- [ ] Task result reporting
- [ ] Task history and logging
- [ ] UI for task creation and monitoring

#### 4. Email-to-Ticket (Week 18-19)

- [ ] Email ingestion (IMAP/webhook)
- [ ] Basic email parsing (sender, subject, body)
- [ ] Ticket creation from email
- [ ] UI for email-to-ticket configuration

#### 5. Client Portal (Week 19-20)

- [ ] Client user authentication
- [ ] Read-only ticket viewing
- [ ] Invoice viewing and download
- [ ] System status viewing
- [ ] Ticket creation by clients

#### 6. OAuth/SSO (Week 20)

- [ ] OAuth2 provider integration (Google, Microsoft)
- [ ] SSO configuration per tenant
- [ ] User provisioning from SSO

**v1.0 Deliverable**: Production-ready MSP platform with billing and task automation.

---

## v1.5: Plugin SDK & Advanced Features

### Version 1.5 Goal

Open platform for extensibility, advanced monitoring, and workflow automation.

### Duration: +12 Weeks (Weeks 21-32)

### Version 1.5 Features

#### 1. Dynamic Plugin Loading (Week 21-23)

- [ ] Migrate agent plugins to Hashicorp go-plugin
- [ ] Plugin distribution mechanism
- [ ] Plugin versioning and updates
- [ ] External plugin support (with security scanning)

#### 2. Control Plane Plugin SDK (Week 23-25)

- [ ] Go plugin interface (using go-plugin)
- [ ] Plugin loader and lifecycle management
- [ ] Integration plugin examples (Slack, PagerDuty)
- [ ] Workflow automation plugins

#### 3. Plugin Marketplace (Week 25-27)

- [ ] Plugin registry and discovery
- [ ] Plugin installation via CLI
- [ ] Plugin ratings and reviews
- [ ] Code signing and verification

#### 4. Advanced Metrics & Alerting (Week 27-29)

- [ ] Custom metric definitions
- [ ] Alert rules and thresholds
- [ ] Alert routing (email, Slack, PagerDuty)
- [ ] Dashboards with charts (Tremor or Recharts)
- [ ] Metric retention policies

#### 5. Reporting & Analytics (Week 29-30)

- [ ] Ticket statistics and trends
- [ ] Time tracking reports (per user, per client)
- [ ] SLA compliance reports
- [ ] Revenue reports
- [ ] Exportable reports (PDF, CSV)

#### 6. Workflow Customization (Week 30-32)

- [ ] Custom ticket workflows per tenant
- [ ] Workflow builder UI (drag-and-drop?)
- [ ] Approval workflows for time entries
- [ ] Automation triggers (ticket created → assign → notify)

**v1.5 Deliverable**: Extensible platform with marketplace, advanced monitoring, and custom workflows.

---

## v2.0: Enterprise & Scale

### Version 2.0 Goal

Enterprise features, high availability, and massive scale.

### Duration: +16 Weeks (Weeks 33-48)

### Version 2.0 Features

#### 1. Database-per-Tenant Option (Week 33-35)

- [ ] Connection routing layer
- [ ] Tenant provisioning workflow
- [ ] Independent backups per tenant
- [ ] Migration from shared to dedicated DB

#### 2. High Availability (Week 35-37)

- [ ] Kubernetes deployment templates
- [ ] PostgreSQL clustering (Patroni or managed DB)
- [ ] NATS clustering
- [ ] Control plane horizontal scaling
- [ ] Load balancing and health checks

#### 3. Advanced Security (Week 37-39)

- [ ] Audit logging for all mutations
- [ ] SOC 2 compliance preparation
- [ ] RBAC enhancements (custom roles)
- [ ] IP whitelisting
- [ ] Two-factor authentication (2FA)

#### 4. AI/ML Features (Week 39-42)

- [ ] Email-to-ticket with AI parsing (OpenAI/Claude)
- [ ] Ticket auto-categorization
- [ ] Predictive maintenance (anomaly detection)
- [ ] Chatbot for ticket creation
- [ ] Smart ticket routing (AI-based assignment)

#### 5. Mobile App (Week 42-46)

- [ ] React Native app for iOS/Android
- [ ] Ticket management on mobile
- [ ] Time tracking on mobile
- [ ] Push notifications
- [ ] Agent status monitoring

#### 6. Advanced Integrations (Week 46-48)

- [ ] Accounting software (QuickBooks, Xero)
- [ ] CRM systems (Salesforce, HubSpot)
- [ ] Monitoring tools (Datadog, New Relic)
- [ ] Zapier integration
- [ ] Webhook system for custom integrations

**v2.0 Deliverable**: Enterprise-grade MSP platform with AI, mobile, and extensive integrations.

---

## Development Workflow

### Sprint Structure

- **Sprint Length**: 2 weeks
- **Sprint Planning**: Define tasks, estimate effort
- **Daily Progress**: Update task status
- **Sprint Review**: Demo completed features
- **Sprint Retro**: Lessons learned, process improvements

### Git Workflow

```bash
main (production)
  ↓
develop (integration branch)
  ↓
feature/ticket-system
feature/agent-metrics
feature/ui-dashboard
```

**Branch Naming**:

- `feature/` - New features
- `fix/` - Bug fixes
- `refactor/` - Code improvements
- `docs/` - Documentation

**Commit Messages**:

```bash
feat(tickets): add ticket filtering by status
fix(agent): handle heartbeat timeout gracefully
docs(api): update ticket endpoints documentation
```

### Release Process

1. Feature complete on `develop`
2. Create release branch `release/v0.5.0`
3. QA testing and bug fixes on release branch
4. Merge to `main` and tag `v0.5.0`
5. Deploy to production
6. Merge back to `develop`

---

## Risk Mitigation

### Technical Risks

| Risk                            | Probability | Impact   | Mitigation                                  |
| ------------------------------- | ----------- | -------- | ------------------------------------------- |
| Agent compatibility issues      | Medium      | High     | Test on all OS versions early               |
| Database performance at scale   | Low         | High     | Benchmark early, optimize indexes           |
| Multi-tenancy leaks             | Low         | Critical | Comprehensive RLS tests, security audit     |
| Plugin security vulnerabilities | Medium      | High     | Code signing, sandboxing, security scanning |
| Third-party API rate limits     | Medium      | Medium   | Implement retries, caching, backoff         |

### Schedule Risks

| Risk                          | Probability | Impact | Mitigation                                 |
| ----------------------------- | ----------- | ------ | ------------------------------------------ |
| Feature creep                 | High        | High   | Strict v0.5 scope, defer non-essentials    |
| Underestimation of effort     | Medium      | Medium | Buffer time (8-12 weeks = realistic 12)    |
| Solo developer bandwidth      | High        | High   | Prioritize ruthlessly, cut scope if needed |
| External dependencies delayed | Low         | Low    | Use stable, mature libraries               |

---

## Success Metrics

### v0.5 Success Metrics

**Adoption** (internal):

- ✅ 1 MSP tenant (your own) using it daily
- ✅ 10+ systems with agents installed
- ✅ 50+ tickets created and managed
- ✅ 100+ time entries tracked

**Technical**:

- ✅ 99% uptime during dogfooding period
- ✅ API response time <200ms (p95)
- ✅ Zero data loss incidents
- ✅ Zero cross-tenant data leaks

**Quality**:

- ✅ 80%+ test coverage
- ✅ Zero critical bugs in production
- ✅ All features documented

- ✅ 80%+ test coverage
- ✅ Zero critical bugs in production
- ✅ All features documented

### v1.0+ Metrics

- **v1.0**: 3+ paying MSP customers, $5k+ MRR
- **v1.5**: 10+ MSP customers, plugin marketplace launched, $20k+ MRR
- **v2.0**: 50+ MSP customers, enterprise deals, $100k+ MRR

---

## Next Steps

1. **Review** all architecture documents
2. **Set up** monorepo structure (see [Project Structure](./06-project-structure.md))
3. **Begin** Week 1 tasks (database setup, Docker Compose)
4. **Track** progress with project management tool (GitHub Projects, Linear, etc.)
5. **Ship** v0.5 in 12 weeks!

---

## Resources

- [System Architecture](./system-architecture.md)
- [Data Model](./02-data-model.md)
- [API Contract](./03-api-contract.md)
- [Plugin Architecture](./04-plugin-architecture.md)
- [Project Structure](./06-project-structure.md)
- [Database Migrations](../backend/database-migrations.md)
- [Testing Strategy](./08-testing-strategy.md)
- [CI/CD Pipeline](./09-ci-cd-pipeline.md)
