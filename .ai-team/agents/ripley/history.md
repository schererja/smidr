# Ripley Work History

## 2026-02-10: Team formation
- Squad hired: 8 agents from Alien universe
- Project: Smidr v0 — quiet, continuous assurance for Linux x86_64
- Tech stack: Go (agent), C# (control plane), React (UI), PostgreSQL (storage), mTLS (auth)
- Work order established: Kane → Dallas → back-and-forth, Ash handles mTLS, Parker tests, Brett documents

## 2026-02-10: UI-API integration coordination
- **Problem:** UI was using mock/fake data, needed to connect to real control plane API
- **Gap identified:** Control plane had `/agents` list but missing detail endpoint, no CORS configured
- **Contract defined:** Documented API endpoints, response formats, CORS requirements in decision file
- **Coordination:** Spawned Dallas and Lambert in parallel to implement integration
  - Dallas: Added `/api/agents/{id}` detail endpoint, updated list endpoint with latestSignals, configured CORS
  - Lambert: Replaced mock data with real API calls, added error handling, configured environment variables
- **Outcome:** UI now displays real agent data from control plane database, no more fake data

## Learnings

### Multi-agent coordination pattern
- When UI and API need to sync, define contract first (endpoints, request/response formats, CORS)
- Write detailed coordination decision before spawning agents
- Spawn frontend and backend agents in parallel with matching contracts
- Both agents succeeded on first attempt due to clear contract specification

### DbContext lifecycle in ASP.NET Core fire-and-forget scenarios
- Scoped DbContext instances are disposed when HTTP request completes
- Fire-and-forget background tasks (`Task.Run()`) outlive request scope
- Solution: Inject `IServiceScopeFactory` instead of scoped services, create new scope in background task
- Pattern: `using var scope = _scopeFactory.CreateScope(); var service = scope.ServiceProvider.GetRequiredService<T>();`
- File: `control-plane/Controllers/HeartbeatController.cs` - demonstrates proper scoping for background health evaluation

### .gitignore strategy for polyglot projects
- Root .gitignore: OS files (macOS, Windows, Linux), editor configs, local secrets, generic patterns
- Component-specific .gitignore files: Language-specific build artifacts, dependencies, test outputs
- Go (agent/): bin/, vendor/, *.coverprofile, compiled binaries
- C# (control-plane/): bin/, obj/, .vs/, *.user, NuGet packages/, SQLite dev databases
- React (ui/): node_modules/, dist/, .env variants
- Certificate exclusions at both root and component levels (dev vs production certs)

### Agent disk collection currently tracks root filesystem only
- `agent/internal/signals/signals.go` line 49: hardcoded `collectDiskUsage("/")` call
- No enumeration of multiple mount points or drives
- Returns single `DiskUsedPct` float in agent signal snapshot
- Control plane stores single `disk_used_pct` column in heartbeats table
- API contract: single `diskUsedPct` field in heartbeat payload

### Per-drive disk metrics deferred to v1
- v0 spec assumes single disk metric (root filesystem percentage)
- Multi-drive systems are common (separate /data, /var/log, database volumes)
- **Decision:** Keep root-only tracking for v0, defer per-drive metrics to v1
- Rationale: Scope containment, most VMs have simple disk topology, baseline complexity, API versioning required, UI display challenges
- v1 plan: Parse /proc/mounts, per-mount baselines, array field in API, expandable UI section

📌 Team update (2026-02-11): Multi-drive disk metrics deferred to v1 (keep root-only for v0) — decided by Ripley
- Current approach can miss full secondary drives if root is healthy
- Proper solution requires agent parsing `/proc/mounts`, control plane schema changes, per-mount baselines, UI updates
- Decision: Ship v0 with root-only, gather feedback, implement per-mount in v1 if needed
- See: `.ai-team/decisions/inbox/ripley-multi-drive-architecture.md` for full analysis

## 2026-02-11: Merged Decisions from Team Debug Session

**Merged from inbox decisions:** ripley-multi-drive-architecture.md

**Key consolidated decisions:**

### Multi-Drive Disk Metrics Architecture (Deferred to v1)
- Comprehensive analysis of disk metric architecture
- Current state: root filesystem only (/) via hardcoded collectDiskUsage("/")
- Production concern: Jason noted systems often have multiple mounted drives
- Decision: Keep root-only for v0, defer per-mount metrics to v1

### Rationale for v0 Deferral
- Scope containment: Adding multi-drive changes agent, control plane, database, and UI
- Single metric: Sufficient for typical development VMs
- Baseline complexity: How to handle anomalies on different drives with different patterns
- API versioning: Would require /v1/agents/heartbeat or feature flags
- UI challenges: How to display 5+ drives compactly in metric card

### v1 Implementation Plan (When Ready)
- **Agent:** Parse /proc/mounts, collect metrics for each non-tmpfs filesystem
- **Control Plane:** DiskMetric table with (heartbeat_id, mount_point, used_pct, total_gb, used_gb)
- **Baselines:** Per-mount baselines using metric names like "DiskUsedPct:/" and "DiskUsedPct:/data"
- **API:** Array of disk objects instead of single diskUsedPct float
- **UI:** Expandable "Disks" section showing table of mount points

### Fallback Option if Jason Insists on v0
- Agent-side aggregation: compute weighted average across all drives, still single API field
- No schema or API changes, captures multi-drive reality without full implementation

**Coordination outcomes:**
- Kane understands disk architecture for future implementation
- Dallas has schema design ready (DiskMetric table structure planned)
- Lambert prepared for future UI updates
- Decision documented with full analysis and v0/v1 tradeoffs
- Deferred with clear decision: "Ship v0 root-only, gather feedback for v1"

### Baseline model is metric-name-based
- `agent_baselines` table: `(agent_id, metric_name, mean, std_dev, min, max, sample_count)`
- Metric names hardcoded in HealthEvaluationService: "LoadAverage1m", "MemoryUsedPct", "DiskUsedPct", "ProcessCount"
- One baseline per metric name per agent (no composite keys, no dimensions)
- This design makes per-drive baselines straightforward: use metric names like "DiskUsedPct:/" and "DiskUsedPct:/data"
- But requires schema change: heartbeats would need per-mount storage (separate table or JSON column)

📌 Team update (2026-02-11): Git tracking exclusions — .ai-team/ and diagnostic files excluded from git per user directive — decided by Jason Scherer
