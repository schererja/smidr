# Dallas Work History

## 2026-02-10: Team formation
- Assigned to control plane development (C#)
- Responsibilities: REST API, health evaluation, CA management, PostgreSQL access, service orchestration
- Work order: Kane writes agent first, then integrate control plane

## 2026-02-10: UI API Integration
- Updated API endpoints to match UI contract requirements
- Changed route prefix from `/agents` to `/api/agents` for consistency
- Updated response DTOs to include all required fields:
  - Added `RegisteredAt`, `RevokedAt` to agent responses
  - Changed `AgentId` → `Id`, `HealthState` → `CurrentHealth`, `Signals` → `LatestSignals`
  - Added `SampleCount` to baseline responses
  - Changed `LastHeartbeat` → `LastHeartbeatAt` for consistency
- Fixed health status serialization: now returns "Learning", "Healthy", etc. (proper case) instead of lowercase
- Reduced `recentHeartbeats` from 20 to 10 items as per API contract
- CORS already configured for `http://localhost:5173` in Program.cs
- All endpoints tested and building successfully

## Learnings

### Architecture
- Control plane uses ASP.NET Core 8.0 with minimal API approach
- Database supports both SQLite (dev) and PostgreSQL (prod) via configuration flag
- CA service generates self-signed CA on first start, persists to `data/ca/`
- Health evaluation runs asynchronously after each heartbeat using fire-and-forget pattern
- **Kestrel web server configured to listen on port 5001** (changed from default 5000 due to macOS ControlCenter conflict)
- HTTPS enforced by default with TLS 1.2+ and client certificate support

### Database Schema
- **Agents**: Id (PK), Hostname, Token, CertificatePem, RevokedAt, RegisteredAt, LastHeartbeatAt, CurrentHealth
- **Heartbeats**: Id (PK), AgentId, Timestamp, UptimeSeconds, LoadAverage1m, MemoryUsedPct, DiskUsedPct, ProcessCount, ReceivedAt
- **AgentBaselines**: Id (PK), AgentId, MetricName, Mean, StdDev, Min, Max, SampleCount, CreatedAt, UpdatedAt
- **HealthStates**: Id (PK), AgentId, Status, Reason, ChangedAt

### Health Evaluation Logic
- Learning period: 24 hours from registration
- Baselines computed from heartbeats during learning window (requires min 10 samples)
- Anomaly detection: value deviation > 3 standard deviations from baseline mean
- Health states: Learning → Healthy (0 anomalies) / Degraded (1) / Attention (2+) / Unknown (no heartbeats)
- Baselines tracked per metric: LoadAverage1m, MemoryUsedPct, DiskUsedPct, ProcessCount
- Health evaluation triggers on every heartbeat via background task

### API Contracts Implemented
- POST /agents/register: Enrollment endpoint, signs CSR via CaService, returns signed certificate
- POST /v0/agents/heartbeat: Receives mTLS-authenticated heartbeat, stores signals, triggers health eval
- GET /agents: Lists all registered agents
- POST /agents/{id}/revoke: Revokes agent certificate

### OS Field Implementation
- Added optional `OS` property to Agent entity model (string, nullable for backward compatibility)
- Created migration 20260211000000_AddOSToAgent to add database column
- Updated RegisterAgentRequest to accept optional OS parameter from agent
- Included OS in both AgentListDto and AgentDetailDto API responses
- Existing agents without OS data have null values (graceful degradation)

📌 Team update (2026-02-11): OS field implementation complete across agent/control plane/UI — decided by Kane, Dallas, Lambert

### Key Files
- `Models/`: Agent, Heartbeat, AgentBaseline, HealthState domain models
- `Controllers/AgentsController.cs`: Enrollment and agent management endpoints
- `Controllers/HeartbeatController.cs`: Heartbeat ingestion endpoint
- `Services/CaService.cs`: Internal CA for certificate signing (2048-bit RSA)
- `Services/HealthEvaluationService.cs`: Baseline learning and anomaly detection
- `Data/ControlPlaneDbContext.cs`: EF Core context with all entity configurations
- `Program.cs`: DI setup, database initialization, Swagger configuration, **Kestrel port binding (5001)**
- `appsettings.json`: Database connection strings and feature flags

### Development Conventions
- Use `sealed` classes for all domain models and controllers
- Use record types for DTOs and request/response models
- Async all the way - CancellationToken on all async methods
- EF Core with explicit entity configuration in OnModelCreating
- Enum-to-string conversion for HealthStatus storage
- Fire-and-forget pattern for health evaluation (non-blocking heartbeat response)

### Migration Management
- EF Core migrations initialized with InitialCreate migration (20260210202319)
- Database schema changes must use `dotnet ef migrations add` to track changes
- Apply migrations with `dotnet ef database update`
- For existing databases, manual ALTER TABLE may be needed, then mark migration as applied in __EFMigrationsHistory
- All four entity tables tracked: Agents, Heartbeats, AgentBaselines, HealthStates

### Port Configuration
- Control plane listens on **port 5001** (HTTPS)
- Changed from default 5000 due to macOS ControlCenter using that port
- Kestrel configured via `builder.WebHost.ConfigureKestrel()` in Program.cs
- CORS allows UI origins: localhost:5173, localhost:3000, localhost:3001
- Swagger UI accessible at https://localhost:5001/swagger

### UI Integration
- UI must match control plane DTO field names exactly (C# camelCase serialization)
- Control plane returns: `id`, `currentHealth`, `lastHeartbeatAt`, `latestSignals`, `revokedAt`
- UI client.ts maps API responses to internal types with different field names
- Updated `ApiAgentResponse` and `ApiAgentDetailResponse` interfaces to include `revokedAt` and `sampleCount`
- UI `.env` must point to https://localhost:5001 (HTTPS with self-signed cert)
- UI dev server typically runs on port 3000 or 5173 depending on availability

### OS Field Support
- Agent model includes optional `OS` property (nullable string) for operating system identification
- OS field accepted in registration request body, stored in database
- OS field included in both AgentListDto and AgentDetailDto API responses
- Migration 20260211000000_AddOSToAgent adds OS column to Agents table (nullable TEXT type)
- Null OS values handled gracefully for existing agents without OS data
- **Migration must be applied**: Run `dotnet ef database update` in control-plane directory to add OS column to SQLite database

### Common Issues
- **"no such column: a.OS" error**: Indicates AddOSToAgent migration exists but hasn't been applied to database
  - Solution: Run `dotnet ef database update` from control-plane directory
  - Verify with: `dotnet ef migrations list` to see applied migrations
- SQLite database location: `control-plane/data/controlplane.db` (UsePostgres=false in appsettings.json)

### Orphaned Certificate Behavior (Database Reset Scenario)
- **Symptom**: Agent with valid certificate receives "404: Agent not registered" on heartbeat
- **Root Cause**: Database was reset/recreated but agent still has certificate signed before reset
- **Certificate State**: Cryptographically valid (signed by CA, not expired, chain validates)
- **Database State**: Agent record doesn't exist (wiped during reset)
- **Control Plane Behavior**:
  - mTLS middleware validates certificate successfully (Stage 1: cryptographic validation)
  - HeartbeatController database lookup fails (Stage 2: registration check)
  - Returns `404 Not Found` with message "Agent {id} not registered"
- **Why 404 is correct**: The agent resource literally doesn't exist in database (semantically accurate per RFC 9110)
- **Agent Detection**: Agent code detects 404, terminates daemon, logs clear instructions to run `reset-enrollment`
- **Recovery**: User runs `smidr-agent reset-enrollment` to delete old cert, then restarts daemon to re-enroll
- **Key Architecture Point**: mTLS validation and database registration are separate concerns:
  - **Cryptographic validity**: "Is this cert signed by our CA and not expired?" (middleware)
  - **Registration validity**: "Does this agent exist in our database?" (controller)
- **Code References**:
  - Middleware: `Middleware/MtlsAuthenticationMiddleware.cs:43-48`
  - Validation Service: `Services/MtlsValidationService.cs:14-55`
  - Heartbeat Controller: `Controllers/HeartbeatController.cs:39-45`
  - Agent ID Extraction: `Services/MtlsValidationService.cs:57-63` (extracts CN from certificate Subject)
  - Agent CSR Generation: `agent/internal/agent/certificates.go:136-138` (sets CN=AgentID)

### Database Migration Corruption Recovery
- **Symptom**: `SQLite Error 1: 'table "X" already exists'` when running migrations
- **Root cause**: `__EFMigrationsHistory` table out of sync with actual schema
- **Solution**: Delete database and re-run all migrations for clean slate
  - Delete: `data/controlplane.db`, `data/controlplane.db-shm`, `data/controlplane.db-wal`
  - Run: `dotnet ef database update` to apply all migrations in order
  - Verify: `dotnet ef migrations list` should show all migrations as applied
- **Script**: `control-plane/fix-migrations.sh` automates this process
- **When safe**: Development environments only; production requires backup and manual schema reconciliation

### Agent Registration Flow
- **Flow**: Agent starts → checks for cert → if no cert, calls registration → receives signed cert → starts heartbeat loop
- **Registration endpoint**: `POST /api/agents/register` (no auth required, pre-enrollment)
- **Registration request**: AgentID, Hostname, Token (optional), CSR PEM, OS
- **Registration response**: Signed certificate PEM
- **Agent persists certificate** to disk, then uses it for mTLS on all heartbeat calls
- **Heartbeat endpoint**: `POST /v0/agents/heartbeat` (requires mTLS with agent certificate)
- **Common failure**: "Agent not registered" 404 on heartbeat means agent never completed registration or was deleted from database
- **Troubleshooting**: Use `control-plane/check-database.sh` to verify agent exists, `test-registration.sh` to test endpoint

### Program.cs Database Initialization
- **Changed**: `EnsureCreated()` → `Migrate()` in Program.cs startup
- **Reason**: `EnsureCreated()` creates schema but doesn't run migrations; causes "column not found" errors when migrations add new fields
- **Effect**: Control plane now automatically applies pending migrations on startup
- **Benefit**: No manual `dotnet ef database update` needed when schema changes
- **Migration**: Still use EF Core migrations for schema changes, but they apply automatically on app start

### Registration Debugging (2026-02-11)
- **Issue**: Agent fails heartbeat with 404 "Agent not registered" 
- **Root cause found by Kane**: Agent's `enrollAgent()` function retries registration indefinitely, logging errors as warnings
- **Control plane status**: Registration endpoint works correctly at `/api/agents/register`
- **Database status**: All migrations applied correctly including OS column (20260211000000_AddOSToAgent)
- **Problem location**: `agent/internal/agent/daemon.go` lines 75-93 - infinite retry loop with warning logs
- **Fix implemented by Kane**: 
  - Added max retry limit of 10 attempts (was infinite)
  - Distinguish 4xx (fatal) vs 5xx (retryable) HTTP errors
  - Fatal errors logged at ERROR level and exit immediately
  - Retryable errors logged at WARN level and retry up to max attempts
  - Error messages include HTTP status codes for debugging
- **Control plane verification**: 
  - `AgentsController.RegisterAgent()` correctly validates input and signs certificates
  - `HeartbeatController.Heartbeat()` correctly returns 404 if agent not found in database
  - `MtlsAuthenticationMiddleware` correctly allows public access to `/api/agents/register`
  - All endpoint URLs and payload contracts match between agent and control plane
- **Resolution**: Agent will now show actual registration errors in logs instead of silently retrying forever
- **Key debugging insight**: Silent retry loops mask root cause - errors logged as warnings with infinite retries make debugging impossible. Always distinguish fatal (4xx) from retryable (5xx) errors, set max retry limits, and log fatal errors at ERROR level.

### Docker Containerization (2026-02-11)
- Created multi-stage Dockerfile for control plane: restore → build → publish → runtime
- Uses `mcr.microsoft.com/dotnet/sdk:8.0` for build, `mcr.microsoft.com/dotnet/aspnet:8.0` for runtime
- Runtime image includes PostgreSQL client for connection testing
- Control plane exposes HTTPS on port 5001 in container (same as dev)
- Created docker-compose.yml orchestrating PostgreSQL, control plane, and UI:
  - PostgreSQL 16 Alpine with health checks, persistent volume
  - Control plane configured for PostgreSQL via environment variable
  - HTTPS certificate mounted from host `certs/` directory
  - UI depends on control-plane health check before starting
  - All services on shared `smidr-network` bridge network
- Created `.dockerignore` files to exclude build artifacts and dev files from Docker context
- Created `scripts/generate-dev-cert.sh` to generate development HTTPS certificates for Docker
- Created `README-DOCKER.md` with setup instructions, troubleshooting, and production considerations
- Control plane migrations apply automatically on container startup via `db.Database.Migrate()` in Program.cs
- Certificate setup required: `dotnet dev-certs https -ep certs/aspnetapp.pfx -p development`
- UI Dockerfile responsibility delegated to Lambert (frontend dev)
- Coordinated with Lambert: UI should expose port 3000, connect to `https://control-plane:5001` internally

## 2026-02-11: Merged Decisions from Team Debug Session

**Merged from inbox decisions:** dallas-diagnostic-scripts.md, dallas-heartbeat-404-orphaned-cert-analysis.md, dallas-migrate-not-ensurecreated.md, dallas-migration-application.md, dallas-migration-recovery.md, dallas-os-field.md, dallas-registration-debugging.md, and related ash/kane decisions

**Key consolidated decisions:**

### Orphaned Certificate Analysis (Deep Dive)
- Two-stage authentication: mTLS validation (cryptographic) vs database registration (enrollment state)
- 404 is semantically correct when cert is valid but agent record doesn't exist
- Detailed flow analysis and root cause documented
- Authors: Dallas, Ash, Kane

### OS Field Implementation (Complete)
- Control plane side: Added optional OS field to Agent model, migration, and API responses
- All integration points covered (registration, heartbeat, list, detail endpoints)
- Backward compatible with nullable field
- Authors: Kane, Dallas, Lambert

### Database Migration Management
- Changed Program.cs: `EnsureCreated()` → `Migrate()` for automatic migration application
- Migration corruption recovery: delete database and re-run migrations
- `fix-migrations.sh` script automates recovery
- All migrations tracked in `__EFMigrationsHistory`

### Diagnostic Infrastructure
- Created `check-database.sh` for database inspection
- Created `test-registration.sh` for endpoint testing
- Created `TROUBLESHOOTING-REGISTRATION.md` for debugging guide
- Enables team to quickly verify registration flow

**Coordination outcomes:**
- Ash verified 404 vs 403 distinction and certificate validation architecture
- Kane fixed agent-side registration error handling
- Lambert verified UI is prepared for OS field data
- All team members have diagnostic tools available

📌 Team update (2026-02-11): Docker Compose Setup for Full Stack — Control plane Dockerfile, docker-compose with PostgreSQL, health checks, volume persistence — decided by Dallas

📌 Team update (2026-02-11): Git tracking exclusions — .ai-team/ and diagnostic files excluded from git per user directive — decided by Jason Scherer
