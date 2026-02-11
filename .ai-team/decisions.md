# Team Decisions

This file records scope, architecture, and process decisions made by the team.

Agents write decisions to `.ai-team/decisions/inbox/` and Scribe merges them here.

---

## Initial Decisions

### 2026-02-10: Project scope and tech stack
**By:** Squad (Coordinator)
**What:** Established project scope, tech stack, and work order for Smidr v0
**Why:** Team formation — agents need day-1 context

**Scope:**
- v0 system: quiet, continuous assurance for Linux x86_64 systems
- Agent (Go) → Control Plane (C#) → UI (React)
- mTLS authentication via internal CA
- Baseline learning + anomaly detection
- PostgreSQL storage

**Tech Stack:**
- Agent: Go, systemd service
- Control Plane: C# REST API (future: gRPC/WebSockets)
- Frontend: React with user registration
- Storage: PostgreSQL
- Auth: mTLS (internal CA)
- Build: GoReleaser (agent), GitHub Actions (CI/CD)

**Work Order:**
1. Kane (Agent Dev) writes first
2. Dallas (Control Plane) integrates
3. Back-and-forth as needed between agent ↔ control plane
4. Ash handles mTLS/crypto
5. Unit tests throughout (Parker)
6. Brett documents everything

---

## Agent Development

### 2026-02-10: Agent unit test infrastructure and patterns
**By:** Parker
**What:** Established Go testing infrastructure for agent package with comprehensive unit tests for signals, config, certificates, and UUID generation. Tests cover edge cases, platform differences, and security requirements. Achieved 43.6% coverage on agent internals and 50.5% on signals.

**Why:** 
- **Platform-aware testing**: Linux-specific /proc tests skip gracefully on macOS/Windows using error detection, not build tags
- **Security validation**: All tests verify 0600 permissions on sensitive files (keys, certs, configs)
- **Edge case coverage**: Tests handle missing files, invalid inputs, malformed data, empty fields
- **Idempotency**: Critical operations (key/CSR generation, config writing) tested for repeated calls
- **No mocking framework**: Used standard library testing with table-driven tests and subtests
- **Colocated tests**: `*_test.go` files live alongside implementation for discoverability

**Testing Patterns:**
- Skip platform-specific tests by checking error messages (e.g., "no such file" for /proc)
- Use `t.TempDir()` for isolated filesystem tests
- Verify file permissions with `os.Stat()` after creation
- Table-driven tests for parsing functions (config, meminfo)
- Integration tests for multi-step workflows (config → key → CSR)
- Assert error keywords, not exact error strings (brittle)

**Rationale:** These patterns enable fast, reliable tests that run on any platform while still validating Linux-specific behavior. Focus on critical paths (enrollment, signal collection) over 100% coverage. Tests document expected behavior and catch regressions.

### 2026-02-10: macOS support for local development
**By:** Kane
**What:** Refactored signal collection to support both Linux and macOS using Go build tags. Split implementation into `signals_linux.go` (uses `/proc` filesystem) and `signals_darwin.go` (uses `sysctl`, `vm_stat`, and `ps` commands). The `Snapshot` struct and `Collect()` API remain unchanged.

**Why:** Agent was failing on macOS with "no such file or directory" errors when trying to read from `/proc`. Adding macOS support enables local development and testing on Mac without requiring Linux VMs. Production deployment remains Linux-focused, but development workflow is now more flexible. Build tags ensure zero runtime overhead — only the appropriate OS implementation is compiled.

**Implementation:**
- Linux: `/proc/uptime`, `/proc/loadavg`, `/proc/meminfo`, `/proc/{pid}/`
- macOS: `sysctl kern.boottime`, `sysctl vm.loadavg`, `vm_stat`, `ps -A`
- Both: `syscall.Statfs` for disk usage (portable)

**Testing:** Verified all signal collectors work on macOS (uptime, load, memory, disk, process count).

### 2026-02-10: Use gopsutil for cross-platform signal collection
**By:** Jason Scherer (via Copilot)
**What:** Use github.com/shirou/gopsutil library for system signal collection instead of writing custom OS-specific collectors
**Why:** User preference — saves development time and maintenance burden, provides battle-tested cross-platform abstractions for system metrics

---

## Control Plane Development

### 2026-02-10: API route standardization
**By:** Dallas
**What:** Changed all control plane API routes from `/agents` to `/api/agents` prefix for consistency and alignment with UI expectations.

**Why:** 
- **Industry standard**: `/api` prefix is a widely adopted convention for REST APIs
- **UI compatibility**: Matches expectations from UI client code
- **Clear separation**: Distinguishes API routes from potential future non-API routes
- **Future-proof**: Makes it easier to add API versioning later (e.g., `/api/v1/agents`)

**Implementation:**
Updated `AgentsController.cs`:
- Changed route attribute from `[Route("agents")]` to `[Route("api/agents")]`
- All endpoints now accessible at:
  - `GET /api/agents` - List all agents
  - `GET /api/agents/{id}` - Get single agent detail
  - `POST /api/agents/register` - Agent enrollment
  - `POST /api/agents/{id}/revoke` - Revoke agent

### 2026-02-10: Response DTO field naming conventions
**By:** Dallas
**What:** Standardized response DTO property names to match UI expectations and follow consistent naming patterns:
- `AgentId` → `Id`
- `HealthState` → `CurrentHealth`
- `Signals` → `LatestSignals`
- `LastHeartbeat` → `LastHeartbeatAt`
- Added `RegisteredAt`, `RevokedAt`, `SampleCount` fields

**Why:** 
- **Clarity**: `CurrentHealth` is clearer than `HealthState` (state vs status ambiguity)
- **Consistency**: All timestamps use `*At` suffix (`RegisteredAt`, `LastHeartbeatAt`, `RevokedAt`)
- **Precision**: `LatestSignals` clarifies we're returning the most recent signals, not all signals
- **Uniqueness**: `Id` is sufficient in context; no need for `AgentId` in agent-specific responses
- **Completeness**: Added fields required by UI

### 2026-02-10: Control plane health evaluation strategy
**By:** Dallas
**What:** Implemented asynchronous health evaluation with 24-hour learning period, baseline computation, and anomaly detection using 3-sigma threshold.

**Why:** 
- Health evaluation must not block heartbeat responses to ensure agent throughput
- 24-hour learning window provides enough samples to establish reliable baselines
- Statistical approach (mean, std dev, 3-sigma) is simple, explainable, and requires no ML dependencies
- Fire-and-forget pattern decouples ingestion from evaluation for better scalability
- Baselines per metric enable granular anomaly detection without false positives from correlated metrics

### 2026-02-10: Database flexibility with SQLite and PostgreSQL support
**By:** Dallas
**What:** Control plane supports both SQLite (default, dev) and PostgreSQL (production) via configuration flag in appsettings.json.

**Why:**
- SQLite enables zero-config local development and testing
- PostgreSQL required for production per project spec
- EF Core abstractions make database switching seamless
- Single codebase reduces maintenance overhead
- Configuration-driven approach allows easy deployment environment customization

### 2026-02-10: DbContext scoping for fire-and-forget background tasks
**By:** Ripley
**What:** Fixed Entity Framework "disposed context" error by using IServiceScopeFactory for background tasks

**Why:** The heartbeat endpoint was firing off health evaluation in a fire-and-forget `Task.Run()` that outlived the HTTP request scope. The scoped DbContext was disposed when the request finished, but the background task still needed it. Solution: Inject `IServiceScopeFactory` into the controller and create a new scope with its own DbContext instance for the background work. This is the standard pattern for any background processing that needs DI services in ASP.NET Core.

### 2026-02-10: Fixed database schema mismatch
**By:** Dallas
**What:** Synchronized SQLite database schema with Agent entity model by adding missing columns (CertificatePem, RevokedAt, LastHeartbeatAt, CurrentHealth) and established EF Core migrations tracking

**Why:** The control plane was crashing with "SQLite Error 1: 'no such column: a.CertificatePem'" because the database was created before all Agent model properties were defined. The database had only Id, Hostname, Token, and RegisteredAt columns, but the model expected four additional properties. Fixed by adding missing columns via ALTER TABLE commands and creating an InitialCreate migration to track the schema going forward. This ensures future schema changes will be properly managed through EF Core migrations.

### 2026-02-10: Control plane listens on port 5001
**By:** Dallas
**What:** Changed control plane from default port 5000 to port 5001

**Why:** macOS ControlCenter system process occupies port 5000, preventing Kestrel from binding. Port 5001 avoids this conflict and is the standard ASP.NET Core alternative port. Updated Program.cs with explicit Kestrel configuration, documentation in README.md and README-CRYPTO.md to reflect new port, and added https://localhost:5173 to CORS for UI development.

### 2026-02-10: Use StatusCode(403) instead of Forbid() in custom mTLS authentication
**By:** Kane
**What:** Changed HeartbeatController to return `StatusCode(StatusCodes.Status403Forbidden)` instead of `Forbid()` for authentication failures

**Why:** The control plane uses custom mTLS middleware for authentication rather than ASP.NET Core's built-in authentication framework. The `Forbid()` method requires an authentication scheme to be registered (via `AddAuthentication()` in `Program.cs`), which we don't have since we handle auth in middleware. Using `StatusCode(403)` directly avoids the "No authenticationScheme was specified" exception while maintaining the same HTTP semantics.

---

## mTLS / Security

### 2026-02-10: mTLS Certificate Authority Implementation
**By:** Ash
**What:** Implemented internal CA and certificate signing infrastructure for Smidr control plane

**Why:** 
- Auto-generated CA on startup provides zero-config enrollment for agents
- X.509 certificate-based authentication is cryptographically strong and standards-based
- mTLS middleware centralizes authentication logic, keeping controllers clean
- RSA 2048-bit chosen for Go/C# compatibility (agent uses crypto/rsa, control plane uses System.Security.Cryptography)
- CustomRootTrust model isolates internal CA from system trust store
- Database-backed revocation simpler than CRL infrastructure for v0 scope

### 2026-02-10: Security Hardening Implementation
**By:** Ash
**What:** Implemented security hardening for CA and mTLS components

**Why:**
- TLS 1.2 minimum enforced (1.3 preferred) to prevent downgrade attacks
- ClientCertificateMode.AllowCertificate permits enrollment without cert, but enforces cert for protected endpoints
- Certificate chain validation uses CustomRootTrust to isolate internal CA from system trust store
- CA key permissions set to 600 (owner-only), CA cert to 644 (world-readable)
- Agent private keys generated on-device, never transmitted (CSR model)
- Certificate expiry checked on every request (NotBefore/NotAfter validation)
- Database-backed revocation simpler and sufficient for v0 (CRL/OCSP deferred to v1)
- ca-tool.sh provides secure CA operations (init, rotate, verify, export)

### 2026-02-10: Kestrel mTLS must be configured per-endpoint (consolidated)
**By:** Kane
**What:** Configure mTLS settings directly on the `UseHttps()` call for each endpoint using `ClientCertificateMode.AllowCertificate`, not via `ConfigureHttpsDefaults()`. Use `ClientCertificateValidation = (cert, chain, errors) => true` at the TLS layer, with actual validation in middleware.

**Why:** 
- When Kestrel's `UseHttps()` is called without parameters, it uses the ASP.NET Core development certificate and ignores global HTTPS defaults
- `AllowCertificate` mode accepts connections with or without client certs, delegating authentication decisions to middleware
- `RequireCertificate` would reject public endpoints (like `/api/agents/register`) at the TLS layer before requests reach middleware
- Registration endpoint must be accessible without client certificates so new agents can enroll
- Validation callback must return `true` to accept all certificates at the TLS layer; actual certificate chain validation against our internal CA happens in `MtlsAuthenticationMiddleware` where we have access to the CA service

**Evolution:** Initially used `RequireCertificate`, then discovered it blocked public endpoints. Changed to `AllowCertificate` to support both authenticated and public routes.

---

## Frontend Development

### 2026-02-10: React UI implementation with mock data
**By:** Lambert
**What:** Built React UI in `ui/` directory with TypeScript, Vite, React Router, and Axios. Implemented SystemList and SystemDetail pages with HealthBadge and MetricsCard components. API client uses mock data with toggle flag for easy transition to real API.

**Why:** User requested "yolo" build approach to get the UI up and running in parallel with agent and control plane development. Mock data allows frontend development to proceed independently while Dallas builds the API. Component structure follows the v0 spec: system list view, system detail view, health state visualization, and current vs baseline metrics display.

### 2026-02-10: UI Framework - Tailwind CSS + shadcn/ui
**By:** Lambert, Jason Scherer
**What:** Use Tailwind CSS and shadcn/ui component library for UI styling. Use data grids for multi-item views (e.g., agents dashboard).

**Why:** Modern CSS framework with component library. User preference for consistency and maintainability. Tailwind provides maintainable styling with responsive utilities, and shadcn/ui offers customizable components that live in the codebase. Data grid pattern scales better than card grid for many agents. This improves development velocity and UI consistency.

### 2026-02-10: UI-Control Plane Integration Complete
**By:** Lambert, Dallas, Ripley
**What:** UI successfully integrated with control plane REST API - replaced all mock data with real API calls

**Why:** User requested UI connect to real data instead of fake data

**Implementation:**
- Removed USE_MOCK flag and all mock data from ui/src/api/client.ts
- Dallas added missing GET /api/agents/{id} endpoint with baselines and recent heartbeats
- Dallas added CORS configuration for localhost:3000, 3001, 5173
- Made /api/agents endpoints public (no mTLS) for v0 UI development
- Updated client.ts to map control plane DTOs (camelCase: agentId, healthState, signals, baselines, recentHeartbeats)
- API base URL: http://localhost:5000/api (later updated to https://localhost:5001)

**Testing:**
- UI successfully calls control plane
- Returns empty array when no agents registered (correct behavior)
- Error handling working (connection refused, 404, 500 errors)

### 2026-02-10: UI endpoints made public for v0 development
**By:** Lambert
**What:** Added /agents and /ca/certificate to public endpoints in MtlsAuthenticationMiddleware

**Why:** UI runs in browser and cannot present mTLS client certificates. For v0 development, agents list/detail and CA certificate endpoints need to be accessible without mTLS.

**Security note:** 
This is acceptable for v0 local development but NOT for production. In production:
- UI should go through an authenticated API gateway/backend-for-frontend
- Or implement session-based auth separate from mTLS agent auth
- Agent endpoints (heartbeat) remain protected by mTLS

**Public endpoints for v0:**
- `/api/agents` (GET list)
- `/api/agents/{id}` (GET detail)
- `/ca/certificate` (GET CA cert for agent registration)
- `/api/agents/register` (POST - already public)

**Protected endpoints:**
- `/v0/agents/heartbeat` (POST - requires agent mTLS cert)
- `/api/agents/{id}/revoke` (POST - requires mTLS cert)

### 2026-02-11: Control Plane API Endpoint Configuration
**By:** Lambert, Dallas
**What:** UI updated to connect to control plane on HTTPS port 5001 (was HTTP port 5000). Fixed API response field name mismatches.

**Why:** Control plane Program.cs only configures HTTPS listener on port 5001 with mTLS support. No HTTP listener on port 5000 exists. UI was using stale configuration from earlier development. Also fixed API response field mapping: C# DTOs use `Id`, `CurrentHealth`, `LatestSignals` (serialized to camelCase as `id`, `currentHealth`, `latestSignals`) but UI client expected `agentId`, `healthState`, `signals`. Updated client.ts to correctly map API responses to UI types.

### 2026-02-11: UI API contract verification and Swagger status
**By:** Dallas
**What:** Verified control plane REST API endpoints work correctly and Swagger is still configured

**Why:** Jason reported UI failing to fetch agents and questioned Swagger status

**Investigation Results:**
- Control plane GET /api/agents endpoint works correctly, returns JSON with proper field names
- Swagger UI accessible at https://localhost:5001/swagger with full OpenAPI spec
- UI TypeScript client interfaces updated to match current control plane DTO field names:
  - Added `revokedAt: string | null` to both response interfaces
  - Added `sampleCount: number` to baseline interface
- Both endpoints confirmed functional via curl testing

**Field Mapping (Control Plane → UI):**
- `id` → `agentId`
- `currentHealth` → `healthState`
- `lastHeartbeatAt` → `lastHeartbeat`
- `latestSignals` → `signals`

**No Changes Required:**
- Control plane endpoints already correct
- Swagger already configured in Program.cs (lines 59-70)
- UI client.ts mapping functions already correct
- Only needed to add missing `revokedAt` and `sampleCount` fields to TypeScript interfaces

---

## Team Process

### 2026-02-10: Copilot directive files (archived)
**By:** System
**What:** Two copilot-directive files from early development (timestamps 1739217036, 1739221425)
**Why:** Historical record of initial setup directives
