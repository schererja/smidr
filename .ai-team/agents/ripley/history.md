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
