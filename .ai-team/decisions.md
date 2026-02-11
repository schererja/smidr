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
**By:** Kane, Dallas
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

## Team Process

### 2026-02-10: Copilot directive files (archived)
**By:** System
**What:** Two copilot-directive files from early development (timestamps 1739217036, 1739221425)
**Why:** Historical record of initial setup directives

---

## Diagnostic & Troubleshooting

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

### 2026-02-11: Heartbeat 404 for Orphaned Certificates - Root Cause and Status Code Analysis (consolidated)

**By:** Dallas, Kane

**What:** Analyzed agent 404 error handling after control plane database reset causes "orphaned certificates." Confirmed HTTP 404 is the correct status code and agent handles the scenario appropriately.

**Context:** Agent with valid certificate receiving "404: Agent not registered" after control plane database reset.

#### Problem Statement

An agent with a cryptographically valid certificate (signed by the control plane CA) is unable to send heartbeats after the control plane database was recreated. The certificate is valid from a cryptographic perspective but the agent record no longer exists in the database.

#### Two-Stage Authentication Analysis

**Stage 1: Cryptographic Certificate Validation** (Middleware)
- Certificate is signed by our CA
- Certificate chain builds successfully with CustomRootTrust
- Certificate is not expired (NotBefore/NotAfter checks)
- **Does NOT check if agent exists in database**

**Stage 2: Database Registration Check** (Controller)
- Agent ID exists in `Agents` table
- **Happens AFTER mTLS validation succeeds**

**Critical Distinction:**
- **Cryptographic validity:** "Is this certificate signed by our CA and not expired?"
- **Registration validity:** "Does this agent have a record in our database?"

An orphaned certificate passes Stage 1 but fails Stage 2.

#### Agent Startup Logic Analysis

**Current Behavior:**
The agent checks if it's enrolled by simply checking if a certificate file exists. It does NOT validate whether the certificate is still recognized by the control plane. The first validation happens when the heartbeat is sent, at which point the control plane returns 404.

**Error Detection:**
The heartbeat client correctly detects 404 errors and treats them specially with `ErrAgentNotRegistered`. The daemon provides clear user guidance:
```
ERROR agent certificate is invalid - this usually happens when the control plane database was reset
ERROR to fix this issue, run: smidr-agent reset-enrollment
ERROR then restart the daemon to re-enroll with a new certificate
```

#### HTTP Status Code Evaluation

**404 Not Found (Current):**
✅ Semantically correct per RFC 9110: The resource (agent) does not exist in the database

**401 Unauthorized:**
❌ Incorrect: The agent DOES have valid authentication credentials (certificate is cryptographically valid)

**403 Forbidden:**
❌ Incorrect: The server isn't refusing a valid request; the resource literally doesn't exist. 403 is already used for revoked certs and agent ID mismatches.

#### Recommendations

**Immediate:**
1. **Keep 404 status code** - semantically correct and well-handled by agent
2. **Document expected behavior** in both agent and control plane READMEs
3. **Optional enhancement:** Add `X-Smidr-Error` header for precise error signaling

**Future (v1):**
1. **Automatic re-enrollment on 404** for self-healing behavior
   - Add startup validation heartbeat for fail-fast
   - Safety limits: max 3 re-enrollment attempts to prevent loops
   - Clear logging when re-enrollment happens
2. **Differentiate 404 vs 403:** Return 403 for revoked agents, 404 for unknown agents
3. **Production hardening:**
   - Certificate revocation list (CRL) including pre-reset certificates
   - Certificate serial tracking in separate audit table
   - Health check endpoint for pre-flight validation

#### Root Cause Summary

**Why it happens:**
1. Certificate signing and database registration are separate operations with no transactional guarantee
2. Certificate has longer lifetime (365 days) than database persistence (can be reset anytime in dev)
3. mTLS validation is stateless - only checks cryptographic validity, not database presence
4. No built-in mechanism to detect database reset or invalidate old certificates

**Why 404 is correct:**
- Certificate is cryptographically valid → passes mTLS middleware
- Agent record doesn't exist → database lookup returns null
- Resource (agent) not found → 404 is semantically accurate
- Agent has excellent 404 handling → user gets clear instructions

**Key Takeaway:** This is not a bug - it's expected behavior when database and certificate lifetimes diverge. The agent's `reset-enrollment` command is the correct recovery mechanism.

### 2026-02-11: Registration failure root cause and fix

**By:** Dallas, Kane  
**What:** Identified that agent registration failures are silently ignored due to infinite retry loop with only warning-level logging  
**Why:** Agent logs show "heartbeat failed with status 404: Agent not registered" but no preceding registration errors are visible. Investigation revealed registration errors are logged as warnings and retried indefinitely, masking the actual failure reason.

**Root Cause:**
- Agent's `enrollAgent()` function (daemon.go:75-93) retries registration indefinitely
- Registration errors from `registerOnce()` are logged as `Warn` but not returned
- No distinction between retryable (5xx) and fatal (4xx) errors
- Agent eventually continues to heartbeat loop even if registration never succeeded
- Heartbeat then fails with 404 because agent was never written to database

**Control Plane Status:**
- Registration endpoint `/api/agents/register` is correctly implemented
- Accepts: agentId, hostname, csrPem, token (optional), os (optional)
- Returns: signed certificate PEM
- Database schema includes all required columns (including OS column from migration 20260211000000)
- Program.cs correctly calls `Migrate()` on startup to apply pending migrations

**Agent Fix Required (Kane):**
1. Distinguish between fatal (4xx) and retryable (5xx) HTTP errors
2. Exit enrollment loop on 4xx errors instead of retrying
3. Add maximum retry limit (e.g., 10 attempts)
4. Log actual registration errors at ERROR level, not WARN
5. Improve error messages to show HTTP status codes

**Recommendation:**
```go
// In registerOnce():
if resp.StatusCode >= 400 && resp.StatusCode < 500 {
    return fmt.Errorf("registration failed permanently (status %d): %s", resp.StatusCode, snippet)
}

// In enrollAgent():
if strings.Contains(err.Error(), "permanently") {
    log.Error("registration failed permanently", "error", err)
    return err  // Don't retry
}
```

**Testing Required:**
1. Run control-plane/test-registration.sh to verify endpoint works
2. Check control-plane/check-database.sh to verify schema is correct
3. Test agent registration with control plane running
4. Verify registration errors are now visible in agent logs
5. Verify 4xx errors stop retry loop immediately

---

## Database & Migrations

### 2026-02-11: Database Migration Recovery Strategy
**By:** Dallas
**What:** Established approach for recovering from corrupted EF Core migration state: delete database and re-run all migrations from scratch.
**Why:** When `__EFMigrationsHistory` is out of sync with actual schema (tables exist but aren't marked as migrated), attempting to fix manually is error-prone. For development environments, the fastest and safest solution is to delete the SQLite database and let EF Core recreate it cleanly. Created `control-plane/fix-migrations.sh` script to automate this. For production, would require backup and careful manual reconciliation.

### 2026-02-11: Agent enrollment reset command
**By:** Kane
**What:** Added `reset-enrollment` command to agent that clears all enrollment state (cert, csr, key files and agent_id) and enables re-enrollment

**Why:** When control plane database is reset, agents with old certificates can't heartbeat because their agent ID no longer exists in the database. The reset command provides a clean way to clear stale enrollment state without manual file deletion. After reset, the agent daemon automatically generates new credentials and re-enrolls with the control plane. This is safer than manual `rm` commands and preserves important config like control_plane_url and heartbeat_interval.

**Implementation Pattern — Idempotent Reset Commands:**
This reset command follows an idempotent design pattern that's reusable for any cleanup/reset operation:

1. **Safe file deletion:** Check if file exists before deleting (no error if already deleted)
   ```go
   func deleteFileIfExists(path string) error {
       if _, err := os.Stat(path); err == nil {
           return os.Remove(path)
       } else if !errors.Is(err, os.ErrNotExist) {
           return err  // Report permission errors, etc.
       }
       return nil  // File doesn't exist, nothing to do
   }
   ```

2. **Selective config clearing:** Only clear reset-related fields, preserve others
   ```go
   updates := map[string]string{
       "agent_id":  "",
       "key_path":  "",
       "csr_path":  "",
       "cert_path": "",
   }
   // Preserves: control_plane_url, hostname, token, heartbeat_interval
   ```

3. **Clear user feedback:** Tell user what was done and next steps
4. **Test both scenarios:** Fresh reset and double reset (idempotency)

This pattern makes reset commands operator-friendly and automation-ready (can run multiple times during troubleshooting without errors).

---

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

### 2026-02-11: Migration Application Required for OS Column

**By:** Dallas

**What:** The AddOSToAgent migration (20260211000000) exists but needs to be manually applied to the database before the control plane can use the OS field.

**Why:** EF Core migrations create the schema change code but don't automatically apply it to the database. The migration file was created in a previous session but wasn't applied, causing the "no such column: a.OS" SQLite error at runtime.

**Action Required:**
```bash
cd control-plane
dotnet ef database update
```

**Technical Details:**
- Migration file: `Migrations/20260211000000_AddOSToAgent.cs`
- Adds nullable TEXT column "OS" to Agents table
- Database location: `control-plane/data/controlplane.db` (SQLite)
- Column is nullable to support existing agents without OS data

### 2026-02-11: Added OS field to Agent model and API

**By:** Dallas

**What:** Added optional OS field to Agent entity, database schema, registration endpoint, and API responses. Agents can now report their operating system (e.g., "linux", "windows", "darwin") during registration.

**Why:** UI requires OS information to display appropriate icons for each agent. The OS field flows from agent → control plane → UI:
- Kane's agent collects and sends OS during registration
- Control plane stores OS in database and returns it in API responses
- Lambert's UI displays OS-specific icons based on this data

**Implementation:**
- Added `string? OS` property to Agent model
- Created migration 20260211000000_AddOSToAgent to add OS column
- Updated RegisterAgentRequest to accept optional OS parameter
- Included OS in both AgentListDto and AgentDetailDto responses
- Nullable field ensures backward compatibility with existing agents

**Migration applied:** Existing agents without OS data will have null values, allowing graceful degradation in the UI.

---

## Agent Development

### 2026-02-11: Agent sends OS information to control plane

**By:** Kane

**What:** Added OS field to agent registration and heartbeat payloads. Agent now reports operating system using `runtime.GOOS` ("linux", "darwin", "windows").

**Why:** UI needs OS information to display appropriate icons for each agent. Registration payload captures OS during enrollment, and heartbeat payload includes OS for consistency. This enables the frontend to show platform-specific icons without requiring a separate API call or database migration.

**Implementation:**
- Added `OS string json:"os"` field to `registerAgentRequest` struct in `agent/internal/agent/daemon.go`
- Added `OS string json:"os"` field to `heartbeatRequest` struct in `agent/internal/heartbeat/client.go`
- Both payloads populate OS field with `runtime.GOOS` value
- Imported `runtime` package in both files

### 2026-02-11: Agent 404 error handling and stale certificate detection

**By:** Kane

**What:** Analysis of agent startup/validation flow after database reset caused 404 errors

**Problem Context:**
The agent was failing to heartbeat with "404: Agent faf726fd-ccda-4cf6-971f-eafd039d65f3 not registered" after the control plane database was recreated. The agent had a valid certificate file but the control plane no longer recognized the agent ID.

**Investigation Findings:**

## 1. Current Startup Logic

**File:** `agent/internal/agent/daemon.go` lines 26-38

The agent checks if it's enrolled by simply checking if a certificate file exists:

```go
enrolled := false
if _, err := LoadCert(cfg.CertPath); err == nil {
    enrolled = true
    log.Info("certificate found", "cert_path", cfg.CertPath)
}

if !enrolled {
    if err := enrollAgent(ctx, cfg, log); err != nil {
        return err
    }
}
```

**Problem:** The agent does NOT validate whether the certificate is still recognized by the control plane. It only checks for file existence. This means a stale certificate (from a wiped database) passes the enrollment check.

## 2. Certificate Validation

The agent performs NO validation of certificate validity with the control plane at startup. There is:
- No check if the agent ID in the cert matches what the control plane expects
- No verification that the control plane still recognizes this agent
- No expiration date checking
- No attempt to verify the certificate chain

The first validation happens when the heartbeat is sent, at which point the control plane returns 404.

## 3. Agent ID Source

**File:** `agent/internal/agent/certificates.go` lines 136-142

The agent ID comes from two sources:
1. **Config file** (`agent_id` field) - generated on first startup if empty (via `PopulateRuntimeConfig()`)
2. **Certificate CN** - when CSR is created, the agent ID is embedded in the Common Name field

The agent stores the agent ID persistently in `config.yaml` (or platform-specific config path). The certificate's CN field contains the same agent ID. The heartbeat client reads the agent ID from the config file, not from the certificate.

**Files:**
- `agent/internal/agent/runtime_config.go` lines 21-28: Generates UUID if `cfg.AgentID == ""`
- `agent/internal/agent/certificates.go` line 137: CSR embeds `cfg.AgentID` in CN
- `agent/internal/heartbeat/client.go` lines 46-48: Uses `cfg.AgentID` from config

## 4. Error Handling

**File:** `agent/internal/heartbeat/client.go` lines 83-84, 94-100, 109-113, 156-158

The heartbeat client DOES detect 404 errors and treats them specially:

```go
var ErrAgentNotRegistered = errors.New("agent not registered in control plane")

// In sendOnce():
if resp.StatusCode == http.StatusNotFound {
    return fmt.Errorf("%w: %s", ErrAgentNotRegistered, snippet)
}

// In Run():
if err := c.sendOnce(ctx); err != nil {
    if errors.Is(err, ErrAgentNotRegistered) {
        return err  // Fail fast - don't retry
    }
    c.log.Warn("initial heartbeat failed", "error", err)
}
```

When a 404 is detected, the error propagates back to `daemon.go`:

**File:** `agent/internal/agent/daemon.go` lines 63-70

```go
err = client.Run(ctx)

if err != nil && errors.Is(err, heartbeat.ErrAgentNotRegistered) {
    log.Error("agent certificate is invalid - this usually happens when the control plane database was reset")
    log.Error("to fix this issue, run: smidr-agent reset-enrollment")
    log.Error("then restart the daemon to re-enroll with a new certificate")
    return fmt.Errorf("agent not registered in control plane (certificate is stale)")
}
```

**Good:** The agent provides clear guidance to the user about how to fix the problem.

**Problem:** The agent exits and requires manual intervention. No automatic recovery.

## 5. Recommendation: Automatic Re-enrollment on 404

**Should the agent auto-re-enroll when it gets 404?**

### Option A: Automatic Re-enrollment (Recommended)

**Pros:**
- Self-healing behavior - agent recovers without manual intervention
- Better user experience - no downtime after database resets
- Matches expected behavior of a resilient monitoring agent
- Simple implementation - call `ResetEnrollment()` then `enrollAgent()` on 404

**Cons:**
- Could mask security issues (revoked certificate looks the same as database reset)
- Agent creates new identity without user awareness
- Potential for enrollment loops if control plane is rejecting enrollment

**Implementation:**
1. When 404 is detected in heartbeat, return to daemon with `ErrAgentNotRegistered`
2. Daemon catches this error, calls `ResetEnrollment()` to wipe cert/key/CSR
3. Daemon calls `enrollAgent()` to re-enroll with new identity
4. Log clearly that re-enrollment is happening and why
5. Add retry backoff to prevent enrollment loops

**Code changes needed:**
- `daemon.go`: Replace error return with re-enrollment logic after detecting `ErrAgentNotRegistered`
- Add backoff timer to prevent enrollment loops (e.g., max 3 re-enrollment attempts per daemon run)

### Option B: Startup Validation Check

**Alternative:** Add a validation heartbeat on startup before entering the main loop.

**Pros:**
- Detects stale certificates immediately at startup
- Fails fast rather than after N seconds of heartbeat interval
- Could still auto-re-enroll or provide clear user guidance

**Cons:**
- Doesn't help if certificate becomes invalid while daemon is running (revocation)
- Adds startup latency (extra HTTP round-trip)

**Implementation:**
1. After loading certificate, send a test heartbeat
2. If 404, trigger re-enrollment or exit with clear error
3. If success, proceed to main heartbeat loop

### Option C: Keep Current Behavior (Not Recommended)

Require manual `reset-enrollment` and daemon restart.

**Pros:**
- User is explicitly aware of the identity change
- No risk of enrollment loops

**Cons:**
- Poor user experience - requires manual intervention
- Doesn't align with "quiet, continuous assurance" goal
- Increases operational burden

## Recommended Implementation

**Combine Option A + Option B:**

1. **Startup validation:** Send initial heartbeat immediately (already implemented in `client.Run()` line 94)
2. **Auto-re-enrollment on 404:** When 404 is detected, automatically reset enrollment and re-register
3. **Safety limits:** Max 3 re-enrollment attempts per daemon run to prevent loops
4. **Clear logging:** Log when re-enrollment happens and why

**Why:** This provides self-healing behavior while preventing enrollment loops. The agent will recover from database resets automatically but won't loop forever if the control plane is rejecting enrollments.

**Coordination with Dallas:**
- Control plane should return 404 for unknown agent IDs (already implemented)
- Consider: Should revoked agents return 404 or 403? Currently both return 404, making them indistinguishable
- Recommend: Return 403 Forbidden for revoked agents, 404 Not Found for unknown agents
- This allows agent to differentiate "I should re-enroll" (404) from "I was explicitly revoked" (403)

**Files to modify:**
1. `agent/internal/agent/daemon.go` - Add re-enrollment logic after 404 detection
2. Add tests for re-enrollment flow in `daemon_test.go`
3. Update documentation to explain auto-recovery behavior

**Why:** This decision improves operational resilience and aligns with the project goal of "quiet, continuous assurance." The agent should recover from expected operational scenarios (database resets, control plane migrations) without manual intervention.

### 2026-02-11: Agent enrollment validation and automatic stale certificate detection

**By:** Kane  
**What:** Agent now validates certificate on first heartbeat and detects stale certificates (404 errors)  
**Why:** Prevent silent failures when control plane database is reset and agent has old certificate

## Problem

After a control plane database reset, agents with existing certificates believe they're enrolled but get 404 errors on heartbeat because their agent ID no longer exists in the database. The daemon would continuously retry heartbeats without detecting the root cause.

## Solution

1. **Heartbeat validation:** First heartbeat attempt now validates that the certificate is recognized by the control plane
2. **404 detection:** When control plane returns 404, the heartbeat client wraps it in `ErrAgentNotRegistered` error
3. **Clear error messages:** Daemon catches this error and provides clear guidance to run `reset-enrollment` command
4. **Fail fast:** Daemon exits immediately on 404 instead of retrying indefinitely

## Implementation

**heartbeat/client.go:**
- Added `ErrAgentNotRegistered` exported error variable
- Modified `sendOnce()` to detect 404 status and wrap error
- Modified `Run()` to fail fast on `ErrAgentNotRegistered` instead of logging and continuing

**agent/daemon.go:**
- Added error check after `client.Run()` to detect `ErrAgentNotRegistered`
- Logs clear error messages with remediation steps
- Returns descriptive error instead of raw 404

## User Experience

**Before:**
```
INFO agent already enrolled cert_path=/path/to/cert.pem
INFO heartbeat loop started
WARN heartbeat failed error="heartbeat failed with status 404: Agent abc-123 not registered"
WARN heartbeat failed error="heartbeat failed with status 404: Agent abc-123 not registered"
... (continues forever)
```

**After:**
```
INFO certificate found cert_path=/path/to/cert.pem
INFO heartbeat loop started
ERROR agent certificate is invalid - this usually happens when the control plane database was reset
ERROR to fix this issue, run: smidr-agent reset-enrollment
ERROR then restart the daemon to re-enroll with a new certificate
(daemon exits with clear error)
```

## Testing

User should:
1. Stop agent daemon
2. Run `smidr-agent reset-enrollment` to clear stale state
3. Restart daemon - will generate new ID and re-enroll
4. Verify heartbeats succeed

## Files Changed

- `agent/internal/heartbeat/client.go` - Added error detection and validation
- `agent/internal/agent/daemon.go` - Added error handling and user guidance
- `TESTING-RESET.md` - Updated documentation with new behavior

### 2026-02-11: Agent enrollment reset command
**By:** Kane
**What:** Added `reset-enrollment` command to agent that clears all enrollment state (cert, csr, key files and agent_id) and enables re-enrollment

**Why:** When control plane database is reset, agents with old certificates can't heartbeat because their agent ID no longer exists in the database. The reset command provides a clean way to clear stale enrollment state without manual file deletion. After reset, the agent daemon automatically generates new credentials and re-enrolls with the control plane. This is safer than manual `rm` commands and preserves important config like control_plane_url and heartbeat_interval.

**Implementation Pattern — Idempotent Reset Commands:**
This reset command follows an idempotent design pattern that's reusable for any cleanup/reset operation:

1. **Safe file deletion:** Check if file exists before deleting (no error if already deleted)
   ```go
   func deleteFileIfExists(path string) error {
       if _, err := os.Stat(path); err == nil {
           return os.Remove(path)
       } else if !errors.Is(err, os.ErrNotExist) {
           return err  // Report permission errors, etc.
       }
       return nil  // File doesn't exist, nothing to do
   }
   ```

2. **Selective config clearing:** Only clear reset-related fields, preserve others
   ```go
   updates := map[string]string{
       "agent_id":  "",
       "key_path":  "",
       "csr_path":  "",
       "cert_path": "",
   }
   // Preserves: control_plane_url, hostname, token, heartbeat_interval
   ```

3. **Clear user feedback:** Tell user what was done and next steps
4. **Test both scenarios:** Fresh reset and double reset (idempotency)

This pattern makes reset commands operator-friendly and automation-ready (can run multiple times during troubleshooting without errors).

---

## UI Design & Features

### 2026-02-11: Removed duplicate navigation routes

**By:** Lambert

**What:** Removed the `/agents` route and "Agents" sidebar navigation item. Dashboard at `/systems` is now the single view for monitoring all agents.

**Why:** Both "Dashboard" and "Agents" navigation items pointed to the same `<SystemList />` component, creating user confusion. The Dashboard label better describes the page's purpose (overview with stats, search, and agent list). If we need a distinct "Agents" page in the future for operational tasks (add/remove/configure), we can add it back with different functionality.

**Files Changed:**
- `ui/src/components/Sidebar.tsx` - Removed `/agents` nav item from navItems array
- `ui/src/App.tsx` - Removed `<Route path="/agents" ...>` definition

### 2026-02-11: Modern SaaS Dashboard UI Design

**By:** Lambert

**What:** Redesigned Smidr UI from basic table layout to modern SaaS dashboard with sidebar navigation, summary metrics, card-based layouts, and enhanced data visualization. Implemented grid/table view toggle, stat cards, and improved visual hierarchy throughout the application.

**Why:** 

**User Request:** Jason explicitly requested a modern professional SaaS dashboard design, referencing TailAdmin examples. The existing UI was functional but visually basic - just a table with minimal styling.

**UX Improvements:**
- **Navigation:** Persistent sidebar with clear visual hierarchy makes the app feel like a complete dashboard platform, not just a single-page utility
- **Information Architecture:** Summary stat cards at the top give instant fleet health overview before diving into individual agents
- **Flexibility:** Grid vs table views accommodate different user preferences and use cases (visual scan vs detailed search)
- **Scannability:** Icons, color coding, and card-based layouts reduce cognitive load and speed up status assessment
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

### 2026-02-11: UI prepared for OS icons and multiple drives
**By:** Lambert
**What:** UI now has OSIcon component and types ready for OS detection. Added placeholder for multiple drive support.

**Current State:**
- UI has OS field in Agent type (optional)
- OSIcon component renders OS-appropriate icons (Linux/Windows/macOS)
- Agent cards and detail pages display OS icons
- Single disk usage metric displayed with TODO comment

**What Needs Backend Changes:**
1. **OS Detection** (requires agent + control plane changes):
   - Agent: Send runtime.GOOS in RegisterAgentRequest
   - Control Plane: Add OS field to Agent model
   - Control Plane: Accept OS in RegisterAgentRequest
   - Control Plane: Return OS in API responses

2. **Multiple Drives** (requires agent + control plane changes):
   - Agent: Detect all mounted filesystems, send array of disk metrics
   - Agent: Each disk should include mount point, total space, used space, used %
   - Control Plane: Update Heartbeat model to support disk array
   - Control Plane: Calculate baselines per drive
   - UI: Display each drive separately in metrics card

**Why:** Jason requested OS icons for each system and support for multiple drives. UI is now prepared to receive this data, but backend changes are required before it can be implemented.

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

## Testing & Quality Assurance

### 2026-02-10: Test infrastructure established for all components
**By:** Parker
**What:** Set up unit test infrastructure for Agent (Go), Control Plane (C#), and UI (React) with 59 total tests covering critical paths.  

**Why:** Testing is essential for confidence in code changes and catching regressions early. Each language/platform needed appropriate test tooling:

- **Agent (Go):** Standard library `testing` package with platform-specific build tags for Linux /proc filesystem tests
- **Control Plane (C#):** xUnit test project with EntityFramework InMemory database and real CaService for integration-style tests
- **UI (React):** Vitest with Testing Library for component and API client tests

**Details:**
- Agent: 50+ tests for config, certificates, UUID, signal collection (fixed Linux build tag issue)
- Control Plane: 22 tests for HealthEvaluationService, MtlsValidationService, AgentsController
- UI: 22 tests for API client, HealthBadge component, MetricsCard component
- All tests runnable via standard commands: `go test ./...`, `dotnet test`, `npm test`
- No integration/E2E tests per v0 scope decision — unit tests only

---

## Architecture & Infrastructure

### 2026-02-11: Multi-drive disk metrics architecture review
**By:** Ripley
**What:** Reviewed agent disk collection architecture in response to Jason's observation that systems can have multiple drives. Confirmed current implementation only tracks root filesystem (`/`).  
**Recommendation:** Keep single aggregate disk metric for v0, defer per-drive tracking to v1.

---

## Current State

**Agent (Go):**
- `agent/internal/signals/signals.go` line 49: `collectDiskUsage("/")` — hardcoded to root filesystem
- Uses `syscall.Statfs()` on a single mount point
- Returns single `DiskUsedPct` float in `Snapshot` struct
- No enumeration of `/proc/mounts` or multiple filesystems

**Control Plane (C#):**
- `Heartbeat` model has single `DiskUsedPct` field (double)
- Health evaluation treats "DiskUsedPct" as one metric name (line 103, HealthEvaluationService.cs)
- Baseline computed per metric name — one baseline for aggregate disk usage
- Database schema: `heartbeats` table has single `disk_used_pct` column

**API Contract:**
- `/v0/agents/heartbeat` expects single `diskUsedPct` field (docs/API.md line 143)
- Documented as "Root disk usage percentage (0-100)"

**UI:**
- Displays single disk percentage value in metrics cards

---

## Architecture Decision: Root-Only for v0

**Recommendation:** Keep single-filesystem tracking for Smidr v0. Do NOT implement per-drive metrics now.

### Rationale

**1. Scope Creep Risk**
- v0 spec defined five signals: uptime, load, memory, disk, process count
- Adding per-drive tracking changes data model, API contract, UI, baselines, health evaluation
- Would require 4-component change (agent, control plane, database migration, UI)

**2. Most Servers Have Simple Disk Topologies**
- Target: Linux x86_64 systems (cloud VMs, containers, basic servers)
- Common pattern: single root filesystem, or root + separate /data mount
- For multi-drive systems, root filesystem health is still the critical signal (OS binaries, logs, temp space)

**3. Baseline Complexity**
- Per-drive baselines complicate evaluation: Do we alert if ANY drive is anomalous? Or majority?
- Different drives have different usage patterns (e.g., /var/log grows linearly, /home is bursty)
- Single root metric is simple, explainable, and sufficient for "is this system healthy?" question

**4. API Versioning Would Be Required**
- Changing heartbeat payload from single `diskUsedPct` to array of drives breaks existing agent/control-plane contract
- Would need `/v1/agents/heartbeat` endpoint or feature flag
- Not worth the versioning burden for v0

**5. UI Display Challenges**
- How do we show 5+ drives in a compact metrics card?
- Which drive do we show in the system list view health badge?
- Do we aggregate across drives (back to single metric)? If so, what did we gain?

### What Jason's Right About

Jason correctly identified that **production systems often have multiple drives:**
- Separate data volumes (`/data`, `/var/lib/docker`)
- Database volumes (`/var/lib/postgresql`, `/mnt/db`)
- Log aggregation mounts (`/var/log`)
- Ephemeral storage for temp files

If `/` is only 10% full but `/data` is at 98%, we'd miss a real problem.

### The Right Solution for v1

When we tackle this properly (v1+), here's the architecture:

**1. Agent Change:**
```go
type DiskSnapshot struct {
    MountPoint string  `json:"mountPoint"`
    UsedPct    float64 `json:"usedPct"`
    TotalGB    float64 `json:"totalGB"`
    UsedGB     float64 `json:"usedGB"`
}

type Snapshot struct {
    // ... existing fields
    Disks []DiskSnapshot `json:"disks"`  // NEW: array of disks
}
```

Parse `/proc/mounts`, filter to real filesystems (ext4, xfs, btrfs, not tmpfs/devtmpfs), run Statfs on each.

**2. Control Plane Changes:**
- New `DiskMetric` table: `(heartbeat_id, mount_point, used_pct, total_gb, used_gb)`
- Baselines: Store per-mount baseline (e.g., "DiskUsedPct:/data", "DiskUsedPct:/")
- Health evaluation: Treat each mount as independent metric, OR aggregate anomaly count across drives

**3. API Contract:**
- Increment to `/v1/agents/heartbeat`
- Change `diskUsedPct` from float to array of objects
- Maintain `/v0/` for backward compatibility (or deprecate)

**4. UI Changes:**
- System detail view: Expandable "Disks" section with table of mount points
- System list view: Show worst disk percentage or aggregate health across all disks
- Baselines chart: Per-mount sparklines

**5. Migration Path:**
- Ship v0 with root-only disk tracking, gather user feedback
- If users report "missed full disk on /data mount" issues, prioritize v1 multi-disk feature
- If root-only proves sufficient, defer indefinitely

---

## Recommendation Summary

**Keep it simple for v0:**
1. No changes to agent disk collection
2. No changes to control plane models or API
3. Document limitation in README: "v0 tracks root filesystem only"
4. Add to backlog: "v1: per-mount disk metrics" as future enhancement

**If Jason insists on multi-disk for v0:**
1. Spawn Kane to add `/proc/mounts` parsing and multi-drive collection
2. Spawn Dallas to add `DiskMetric` model and update heartbeat schema
3. Spawn Lambert to update UI for disk arrays
4. Requires database migration, API version bump, UI redesign
5. Estimated effort: 1-2 days (vs. 0 hours for deferral)

**My call as Lead:** Ship v0 with root-only disk. Gather feedback. Implement per-drive in v1 if users need it.

---

## If We Must Implement Now (Fallback Plan)

If Jason decides this is critical for v0, here's the minimal-viable approach:

**Option A: Agent-side aggregation**
- Agent reads `/proc/mounts`, collects all non-tmpfs filesystems
- Computes weighted average: `total_used_blocks / total_available_blocks` across all drives
- Still sends single `diskUsedPct` float
- **Pros:** No API/schema changes, captures multi-drive reality
- **Cons:** Loses per-drive granularity, can mask one full drive if others are empty

**Option B: Array field with backward compat**
- Add `disks: []` array to heartbeat request, keep `diskUsedPct` as deprecated field
- Agent sends both: single root metric + optional array
- Control plane stores array in JSON column or separate table
- v0 evaluation uses root only, v1 evaluation uses array
- **Pros:** Forward-compatible, no API version bump
- **Cons:** Doubles storage for disk metrics, UI doesn't show array yet

I recommend **Option A** if we must ship multi-disk in v0.

---

**Follow-up:** Jason, let me know your decision. If you want to proceed with multi-drive now, I'll coordinate Kane, Dallas, and Lambert for the implementation.
