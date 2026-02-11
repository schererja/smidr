# Brett Work History

## 2026-02-10: Team formation
- Assigned to technical writing
- Responsibilities: Installation guides, API docs, architecture docs, user guides, developer docs, README updates
- Work order: Document as Kane, Dallas, Lambert deliver features

## Learnings

### Documentation Structure
- Created `docs/` directory at repository root for centralized documentation
- Established three core docs: ARCHITECTURE.md, API.md, DEVELOPMENT.md
- README.md serves as entry point with links to detailed docs

### Architecture Patterns
- mTLS enrollment flow: agent generates CSR → control plane signs → agent uses cert for heartbeats
- Health evaluation: asynchronous processing triggered by heartbeat ingestion (non-blocking)
- Baseline learning: 24h window, statistical approach (mean, stddev, trimmed min/max)
- Health states: Learning → Healthy/Degraded/Attention based on anomaly count (z-score > 3)

### API Contracts
- Enrollment: POST /agents/register (public, no auth)
- Heartbeat: POST /v0/agents/heartbeat (mTLS required)
- CA cert: GET /ca/certificate (public)
- UI queries: GET /api/agents, GET /api/agents/{id} (no auth in v0)
- Version prefix /v0/ used for endpoints that may evolve

### Control Plane Implementation
- Controllers: AgentsController, HeartbeatController, CaController
- Services: CaService (CA operations), HealthEvaluationService (baseline learning), MtlsValidationService
- Middleware: MtlsAuthenticationMiddleware enforces certificate validation
- Database: EF Core with SQLite (dev) or PostgreSQL (prod)

### Agent Implementation
- Signal collection from /proc: uptime, loadavg, meminfo, statfs
- Config file: /etc/smidr/agent.yaml (YAML format)
- Certificates: /etc/smidr/agent.{key,csr,crt}.pem
- Systemd service: smidr-agent.service with security hardening

### UI Implementation
- React 19 + TypeScript + Vite
- API client at ui/src/api/
- Components: SystemList, SystemDetail, HealthBadge, MetricsCard
- Auto-refresh: 30 second interval
- Color coding: Blue (learning), Green (healthy), Orange (degraded), Red (attention), Gray (unknown)

### Key File Paths
- `/agent/cmd/agent/` - agent main entry point
- `/agent/internal/signals/` - signal collectors
- `/agent/internal/heartbeat/` - heartbeat client
- `/control-plane/Controllers/` - AgentsController, HeartbeatController, CaController
- `/control-plane/Services/` - CaService, HealthEvaluationService, MtlsValidationService
- `/control-plane/Middleware/` - MtlsAuthenticationMiddleware
- `/ui/src/api/` - API client
- `/ui/src/components/` - Reusable UI components
- `/ui/src/pages/` - Page-level components
- `/docs/` - Centralized documentation (ARCHITECTURE.md, API.md, DEVELOPMENT.md)

### Documentation Standards
- Use markdown for all documentation
- Include runnable code examples
- Document all API endpoints with full request/response schemas
- Provide troubleshooting sections for common issues
- Cross-link between related documents
- Keep README.md concise with links to detailed docs

📌 Team update (2026-02-11): Documentation structure established with docs/ directory and three core documents — decided by Brett
- `/control-plane/Controllers/` - REST API endpoints
- `/control-plane/Services/` - business logic (CA, health evaluation)
- `/control-plane/Models/` - entity models (Agent, Heartbeat, AgentBaseline, HealthState)
- `/ui/src/api/` - API client for control plane
- `/ui/src/pages/` - page components (SystemList, SystemDetail)

## 2026-02-11: Merged Decisions from Team Debug Session

**Merged from inbox decisions:** brett-documentation-structure.md

**Key consolidated decisions:**

### Documentation Structure and Standards
- Centralized documentation in `docs/` directory at repository root
- Three core documents establish hierarchy:
  1. **ARCHITECTURE.md** - System design, component relationships, enrollment flow, health evaluation
  2. **API.md** - REST endpoints, request/response schemas, authentication
  3. **DEVELOPMENT.md** - Local setup, running tests, debugging guide, troubleshooting
- **README.md** serves as entry point with links to detailed docs

### Documentation Standards Established
- Markdown format for all documentation
- Include runnable code examples with syntax highlighting
- Document all API endpoints with full request/response schemas
- Provide troubleshooting sections for common issues
- Cross-link between related documents for navigation
- Keep README.md concise with links to detailed docs
- Document architecture patterns (mTLS enrollment flow, async health evaluation, fire-and-forget baselines)

### Key Documentation Areas
- **Enrollment Flow:** Agent CSR generation → Control plane signing → Certificate storage → mTLS heartbeats
- **Health Evaluation:** 24-hour learning window, baseline computation (mean/stddev/min/max), 3-sigma anomaly detection
- **API Contracts:** Endpoints, authentication requirements (public vs mTLS), response DTOs
- **Troubleshooting:** Common errors (404, migration issues, CORS), diagnostic scripts, recovery procedures

### Future Documentation Needs (Identified)
- Orphaned certificates guide (detailed explanation for developers)
- Multi-drive metrics architecture (documented in decision for v1 planning)
- Agent configuration reference
- Baseline learning visualization
- Deployment guide for production PostgreSQL setup

**Coordination outcomes:**
- All major architectural decisions have documentation entries
- Team members know where to add docs as features ship
- Clear standards prevent documentation sprawl
- README links to detailed docs prevent information overload
