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
