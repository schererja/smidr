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

### 2026-02-11: Agent enrollment and certificate validation (consolidated)
**By:** Kane, Dallas
**What:** Implemented comprehensive agent enrollment flow with automatic stale certificate detection, 404 error handling, and reset-enrollment command. Agent validates certificates on first heartbeat, detects database resets, and provides clear recovery instructions.

**Why:** After control plane database resets, agents with valid certificates receive 404 errors because their agent IDs no longer exist in the database. This required systematic handling across multiple failure modes:

**Problem Context:**
- Cryptographically valid certificate (signed by CA, not expired)
- Agent record deleted from database (e.g., dev database reset)
- Two-stage authentication: Stage 1 (crypto validation) passes, Stage 2 (database lookup) fails
- Agent must detect this condition and guide user to recovery

**Implementation Components:**

**1. Agent Startup Logic (daemon.go):**
- Checks for certificate file existence to determine enrollment status
- Does NOT validate certificate recognition by control plane at startup
- First validation occurs when initial heartbeat is sent
- On 404 error, provides clear remediation steps

**2. Heartbeat Error Detection (heartbeat/client.go):**
- Added `ErrAgentNotRegistered` exported error variable
- Modified `sendOnce()` to detect 404 status and wrap error
- Modified `Run()` to fail fast on `ErrAgentNotRegistered` instead of retrying
- Clear logging: "agent certificate is invalid - this usually happens when the control plane database was reset"

**3. Reset-Enrollment Command:**
- Added `reset-enrollment` command to clear all enrollment state
- Deletes cert, csr, key files and clears agent_id from config
- Preserves other config (control_plane_url, heartbeat_interval)
- Idempotent design - safe to run multiple times
- Safe file deletion pattern: check existence before deleting

**4. Registration Error Handling:**
- Distinguish between fatal (4xx) and retryable (5xx) HTTP errors
- Exit enrollment loop on 4xx errors instead of infinite retry
- Log registration errors at ERROR level, not WARN
- Improved error messages showing HTTP status codes

**Two-Stage Authentication Analysis:**

**Stage 1: Cryptographic Certificate Validation** (Middleware)
- Certificate is signed by our CA
- Certificate chain builds successfully with CustomRootTrust
- Certificate is not expired (NotBefore/NotAfter checks)
- **Does NOT check if agent exists in database**

**Stage 2: Database Registration Check** (Controller)
- Agent ID exists in `Agents` table
- **Happens AFTER mTLS validation succeeds**

**HTTP Status Code Decision:**
- **404 Not Found (chosen):** Semantically correct - the resource (agent) does not exist in database
- **401 Unauthorized:** Incorrect - agent DOES have valid authentication credentials
- **403 Forbidden:** Incorrect - already used for revoked certs and agent ID mismatches

**User Experience:**

Before:
```
INFO agent already enrolled cert_path=/path/to/cert.pem
INFO heartbeat loop started
WARN heartbeat failed error="heartbeat failed with status 404"
... (continues forever)
```

After:
```
INFO certificate found cert_path=/path/to/cert.pem
INFO heartbeat loop started
ERROR agent certificate is invalid - this usually happens when the control plane database was reset
ERROR to fix this issue, run: smidr-agent reset-enrollment
ERROR then restart the daemon to re-enroll with a new certificate
(daemon exits with clear error)
```

**Future Enhancements (v1):**
1. Automatic re-enrollment on 404 for self-healing behavior
2. Startup validation heartbeat for fail-fast detection
3. Safety limits: max 3 re-enrollment attempts to prevent loops
4. Differentiate 404 vs 403: Return 403 for revoked agents, 404 for unknown agents
5. Production hardening: CRL including pre-reset certificates, certificate serial tracking

**Root Cause Summary:**
- Certificate signing and database registration are separate operations
- Certificate has longer lifetime (365 days) than database persistence
- mTLS validation is stateless - only checks crypto, not database presence
- No built-in mechanism to detect database reset

**Key Takeaway:** This is not a bug - it's expected behavior when database and certificate lifetimes diverge. The agent's `reset-enrollment` command is the correct recovery mechanism for v0.

**Files Changed:**
- `agent/internal/heartbeat/client.go` - Error detection and validation
- `agent/internal/agent/daemon.go` - Error handling and user guidance
- `agent/cmd/reset_enrollment.go` - New reset command
- `TESTING-RESET.md` - Documentation

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

### 2026-02-11: Added Diagnostic Scripts for Registration Troubleshooting

**By:** Dallas

**What:** Created three diagnostic scripts to help troubleshoot agent registration and database issues:
1. `check-database.sh` - Inspects database schema, migration history, and agent records
2. `test-registration.sh` - Tests registration endpoint with synthetic agent
3. `docs/TROUBLESHOOTING-REGISTRATION.md` - Comprehensive guide for diagnosing registration failures

**Why:**
- "Agent not registered" errors are confusing without visibility into database state
- Manual SQL queries are tedious and error-prone
- Need quick way to test registration without running full agent
- Team members (especially Kane working on agent) need to understand registration flow

**Usage:**
```bash
cd control-plane
./check-database.sh           # Inspect database state
./test-registration.sh        # Test registration endpoint
cat docs/TROUBLESHOOTING-REGISTRATION.md  # Read troubleshooting guide
```

**Files created:**
- `control-plane/check-database.sh`
- `control-plane/test-registration.sh`
- `control-plane/docs/TROUBLESHOOTING-REGISTRATION.md`

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

## Testing

### 2026-02-10: Test infrastructure established for all components
**By:** Parker
**What:** Set up unit test infrastructure for Agent (Go), Control Plane (C#), and UI (React) with 59 total tests covering critical paths. Established testing patterns: Go standard library with platform-specific build tags for agent, xUnit + InMemory EF for control plane, Vitest + Testing Library for UI. All tests runnable via standard commands.

**Why:** Testing is essential for confidence in code changes and catching regressions early. Each language/platform needed appropriate test tooling:
- **Agent (Go):** Standard library `testing` package with platform-specific build tags for Linux /proc filesystem tests
- **Control Plane (C#):** xUnit test project with EntityFramework InMemory database and real CaService for integration-style tests
- **UI (React):** Vitest with Testing Library for component and API client tests

**Details:**
- Agent: 50+ tests for config, certificates, UUID, signal collection (fixed Linux build tag issue)
- Control Plane: 22 tests for HealthEvaluationService, MtlsValidationService, AgentsController
- UI: 22 tests for API client, HealthBadge component, MetricsCard component
- No integration/E2E tests per v0 scope decision — unit tests only

---

## Documentation

### 2026-02-10: Documentation structure and standards
**By:** Brett
**What:** Created centralized documentation structure with three core documents (ARCHITECTURE.md, API.md, DEVELOPMENT.md) in `docs/` directory at repository root. Updated README.md to serve as entry point with links to detailed documentation.

**Why:** Establishes a clear documentation hierarchy that scales as the project grows. Separating architecture, API reference, and development guide allows different audiences (users, integrators, contributors) to find relevant information quickly. Centralizing docs in `docs/` directory follows industry convention and makes documentation discoverable. This structure prevents documentation sprawl and ensures all team members know where to add new docs as features are delivered.

**Standards Established:**
- Use markdown for all documentation
- Include runnable code examples
- Document all API endpoints with full request/response schemas
- Provide troubleshooting sections for common issues
- Cross-link between related documents
- Keep README.md concise with links to detailed docs

---

## Cross-Component Features

### 2026-02-11: OS field implementation (consolidated)
**By:** Kane, Dallas, Lambert
**What:** Implemented end-to-end OS detection from agent to control plane to UI. Agent reports operating system using `runtime.GOOS` during registration and heartbeat. Control plane stores OS in Agent entity and returns it in API responses. UI displays OS-specific icons.

**Why:** UI requires OS information to display appropriate icons for each agent. The complete data flow enables platform identification:

**Agent (Kane):**
- Added `OS string json:"os"` field to registration and heartbeat payloads
- Populates with `runtime.GOOS` value ("linux", "darwin", "windows")
- Imported `runtime` package in daemon.go and heartbeat client

**Control Plane (Dallas):**
- Added optional `string? OS` property to Agent entity model
- Created migration 20260211000000_AddOSToAgent to add database column
- Updated RegisterAgentRequest to accept optional OS parameter
- Included OS in both AgentListDto and AgentDetailDto API responses
- Nullable field ensures backward compatibility with existing agents

**UI (Lambert):**
- Created OSIcon component that renders platform-appropriate icons
- Agent type includes optional OS field
- Agent cards and detail pages display OS icons
- Falls back gracefully when OS data is unavailable (existing agents)

**Migration:** Existing agents without OS data have null values, allowing graceful degradation.

### 2026-02-11: Multi-drive disk metrics architecture (deferred to v1)
**By:** Ripley
**What:** Reviewed agent disk collection architecture in response to multi-drive requirements. Current implementation tracks root filesystem (`/`) only. Recommendation: Keep single aggregate disk metric for v0, defer per-drive tracking to v1.

**Why:** 
- **v0 scope:** Single-filesystem tracking is simple, explainable, and sufficient for initial release
- **Most servers have simple disk topologies:** Cloud VMs and containers typically use single root filesystem or root + data mount
- **Baseline complexity:** Per-drive baselines complicate health evaluation (which drives matter? how to aggregate?)
- **API versioning required:** Changing heartbeat payload from single float to array breaks agent/control-plane contract
- **UI display challenges:** Showing 5+ drives in compact metrics cards requires UX redesign

**Current State:**
- Agent: `collectDiskUsage("/")` hardcoded to root filesystem using `syscall.Statfs()`
- Control Plane: Single `DiskUsedPct` field in Heartbeat model
- API: `/v0/agents/heartbeat` expects single `diskUsedPct` field
- UI: Displays single disk percentage value

**v1 Implementation Plan:**
- Agent: Parse `/proc/mounts`, collect all real filesystems (ext4, xfs, btrfs)
- Control Plane: New `DiskMetric` table with per-mount baselines
- API: Increment to `/v1/agents/heartbeat` with disk array
- UI: Expandable disks section with per-mount details

**Fallback if required for v0:**
- Option A: Agent-side aggregation (weighted average across all drives, single float)
- Option B: Array field with backward compat (add `disks[]` array, keep `diskUsedPct`)

**Decision:** Ship v0 with root-only disk tracking. Gather user feedback. Implement per-drive in v1 if users report missed full-disk issues on non-root mounts.

---

## UI Design & UX

### 2026-02-11: Modern SaaS Dashboard UI Design
**By:** Lambert
**What:** Redesigned Smidr UI from basic table layout to modern SaaS dashboard with sidebar navigation, summary metrics, card-based layouts, and enhanced data visualization. Implemented grid/table view toggle, stat cards, and improved visual hierarchy throughout the application.

**Why:** 
- **User Request:** Jason explicitly requested a modern professional SaaS dashboard design, referencing TailAdmin examples
- **UX Improvements:** Persistent sidebar provides clear navigation hierarchy, summary stat cards give instant fleet health overview, grid vs table views accommodate different user preferences, icons and color coding reduce cognitive load
- **Professional Appearance:** Modern design builds trust and confidence in the monitoring platform

**Design Decisions:**
1. **Sidebar Navigation:** Standard left sidebar with logo, nav items, and user profile follows established SaaS patterns (GitHub, Stripe, AWS Console)
2. **Stat Cards:** Four key metrics (total agents, healthy, issues, avg uptime) provide fleet-wide visibility before agent-level detail
3. **Purple Brand Color:** Maintained existing purple gradient for brand consistency, used as primary accent throughout
4. **Grid View Default:** Cards are more visual and user-friendly for typical monitoring tasks (< 20 agents)
5. **Icon-First Design:** Every section, card, and metric has an icon for faster visual identification
6. **Multi-Level Severity:** Baseline deltas use 3 levels (normal/warning/critical) not binary good/bad

**Technical Implementation:**
- Created 4 new reusable components: Sidebar, Header, StatCard, AgentCard
- Redesigned SystemList with dashboard layout and view toggle
- Enhanced SystemDetail with hero section and improved data visualization
- Improved MetricsCard with icon-based card layout and severity colors
- Added recharts dependency for future chart features
- Updated documentation to reflect new design system

**Future Considerations:**
- Charts/graphs for metric trends over time (recharts ready to use)
- User settings page for customization
- Dark mode support (CSS variables already set up for theming)
- Mobile responsive improvements (current design works on tablets, needs optimization for phones)

### 2026-02-11: Removed duplicate navigation routes
**By:** Lambert
**What:** Removed the `/agents` route and "Agents" sidebar navigation item. Dashboard at `/systems` is now the single view for monitoring all agents.

**Why:** Both "Dashboard" and "Agents" navigation items pointed to the same `<SystemList />` component, creating user confusion. The Dashboard label better describes the page's purpose (overview with stats, search, and agent list). If we need a distinct "Agents" page in the future for operational tasks (add/remove/configure), we can add it back with different functionality.

**Files Changed:**
- `ui/src/components/Sidebar.tsx` - Removed `/agents` nav item from navItems array
- `ui/src/App.tsx` - Removed `<Route path="/agents" ...>` definition

### 2026-02-11: shadcn/ui requires CSS variable definitions in App.css
**By:** Lambert
**What:** When using shadcn/ui components, you must define CSS custom properties for all semantic theme tokens in App.css within a `@layer base` block, AND extend the Tailwind config to map those variables to utility class names.

**Why:** 
shadcn/ui components use semantic class names like `bg-background`, `border-input`, `text-muted-foreground` that reference CSS variables via `hsl(var(--variable))`. If these variables aren't defined, all component styling is invalid and the UI renders completely unstyled. 

The two-part setup is required:
1. Define CSS variables in App.css: `--background: 0 0% 100%;` (HSL space-separated format)
2. Map to Tailwind utilities in config: `background: "hsl(var(--background))"`

This architecture enables runtime theme switching (light/dark) by changing CSS variable values without rebuilding. It's a common mistake to copy shadcn components without completing the theme setup - the `npx shadcn@latest init` CLI handles this automatically, but manual setup must include both pieces.

**Files affected:**
- `ui/src/App.css` - CSS variable definitions
- `ui/tailwind.config.js` - Theme color mappings

---

## Database & Migrations

### 2026-02-11: Database Migration Recovery Strategy
**By:** Dallas
**What:** Established approach for recovering from corrupted EF Core migration state: delete database and re-run all migrations from scratch.
**Why:** When `__EFMigrationsHistory` is out of sync with actual schema (tables exist but aren't marked as migrated), attempting to fix manually is error-prone. For development environments, the fastest and safest solution is to delete the SQLite database and let EF Core recreate it cleanly. Created `control-plane/fix-migrations.sh` script to automate this. For production, would require backup and careful manual reconciliation.

### 2026-02-11: Use Migrate() Instead of EnsureCreated() for Database Initialization

**By:** Dallas

**What:** Changed Program.cs to use `db.Database.Migrate()` instead of `db.Database.EnsureCreated()` for database initialization on startup.

**Why:** 
- `EnsureCreated()` creates the database with the current schema but completely ignores EF Core migrations
- When new migrations are added (like AddOSToAgent), existing databases don't get updated
- This causes "no such column" errors when code expects fields added by migrations
- `Migrate()` applies all pending migrations automatically on startup, keeping schema in sync
- Eliminates need for manual `dotnet ef database update` commands during development

**Impact:**
- Control plane now auto-applies migrations on every start
- Developers don't need to remember to run migration commands
- Reduces "my code works but yours doesn't" issues from schema drift

**Files changed:**
- `control-plane/Program.cs`: Line 66 changed from `EnsureCreated()` to `Migrate()`

**Note:** This is safe for development and production. Migrations are idempotent and only apply once.

---

## Team Process

### 2026-02-10: Copilot directive files (archived)
**By:** System
**What:** Two copilot-directive files from early development (timestamps 1739217036, 1739221425)
**Why:** Historical record of initial setup directives

---
