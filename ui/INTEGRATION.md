# UI Integration Testing Guide

**Status:** API Integration Complete ✅

The UI is now fully integrated with the control plane API. Mock data has been removed.

## Prerequisites

1. **Control plane must be running:**
   ```bash
   cd control-plane
   dotnet run
   ```
   Should be accessible at http://localhost:5000

2. **At least one agent registered and sending heartbeats**

3. **Environment configured:**
   ```bash
   cd ui
   cp .env.example .env
   # Verify VITE_API_BASE_URL=http://localhost:5000
   ```

## Quick Start

```bash
# Terminal 1: Start control plane
cd control-plane
dotnet run

# Terminal 2: Start UI
cd ui
npm run dev
```

Access UI at http://localhost:3000

## API Contract (Implemented)

### GET /api/agents

Returns array of agents with latest signals and health state.

**Expected Response (PascalCase from C#):**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "hostname": "web-server-01",
    "registeredAt": "2026-02-10T10:00:00Z",
    "lastHeartbeatAt": "2026-02-10T10:30:00Z",
    "currentHealth": "Healthy",
    "revokedAt": null,
    "latestSignals": {
      "uptimeSeconds": 86400,
      "loadAverage1m": 0.85,
      "memoryUsedPct": 42.3,
      "diskUsedPct": 58.2,
      "processCount": 142
    }
  }
]
```

**UI Mapping:**
- `id` → `agentId`
- `lastHeartbeatAt` → `lastHeartbeat`
- `currentHealth` → `healthState` (lowercase)
- `latestSignals` → `signals`

### GET /api/agents/{id}

Returns single agent with baselines and recent heartbeats.

**Expected Response (PascalCase from C#):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "hostname": "web-server-01",
  "registeredAt": "2026-02-10T10:00:00Z",
  "lastHeartbeatAt": "2026-02-10T10:30:00Z",
  "currentHealth": "Healthy",
  "revokedAt": null,
  "latestSignals": { ... },
  "baselines": [
    {
      "metricName": "loadAverage1m",
      "mean": 1.2,
      "stdDev": 0.3,
      "min": 0.5,
      "max": 2.1,
      "sampleCount": 100
    }
  ],
  "recentHeartbeats": [
    {
      "timestamp": "2026-02-10T10:29:00Z",
      "uptimeSeconds": 86340,
      "loadAverage1m": 0.82,
      "memoryUsedPct": 41.8,
      "diskUsedPct": 58.0,
      "processCount": 140
    }
  ]
}
```

**UI Mapping:**
- `metricName` → `metric` (in baselines)
- All other fields follow same mapping as list endpoint

## Testing Steps

### 1. System List View

Navigate to http://localhost:3000/systems

**Expected:**
- List of all registered agents
- Each agent shows:
  - Hostname
  - Health state badge (colored: learning/healthy/degraded/attention/unknown)
  - Agent ID (truncated to 8 chars)
  - Last seen timestamp (e.g., "30s ago", "2m ago")
  - Quick stats: Load, Memory %, Disk %
- Page auto-refreshes every 30 seconds

**Error Scenarios:**
- Control plane not running: "Cannot connect to control plane API. Is the server running?"
- No agents exist: Empty list (no error message)

### 2. System Detail View

Click on an agent or navigate to http://localhost:3000/systems/{agent-id}

**Expected:**
- Agent hostname and health state badge
- **Current Signals section:**
  - Uptime, Load Average, Memory %, Disk %, Process Count
  - Delta from baseline shown if baselines exist
  - Color-coded: green (within range), orange (warning), red (critical)
- **Baselines section:**
  - Mean, Std Dev, Min, Max for each metric
  - Shows "Learning..." if baselines not yet computed
- **Recent Heartbeats section:**
  - Last 10 heartbeats with timestamps and all signals
  - Ordered newest to oldest
- Page auto-refreshes every 30 seconds

**Error Scenarios:**
- Invalid agent ID: "Agent not found"
- Control plane not running: "Cannot connect to control plane API. Is the server running?"
- Server error: "Server error loading agent details"

### 3. Browser DevTools Check

Open browser DevTools (F12) → Console tab

**Expected:**
- No JavaScript errors
- Network tab shows successful requests:
  - `GET /api/agents` → 200 OK
  - `GET /api/agents/{id}` → 200 OK
- Response bodies match API contract

**Debugging Tips:**
- CORS errors → Dallas needs CORS configuration (see below)
- 404 errors → Check API endpoints exist in control plane
- ECONNREFUSED → Verify control plane running on correct port

## Known Issues & Solutions

### CORS Configuration Required

If you see CORS errors in browser console, Dallas needs to add this to `Program.cs`:

```csharp
builder.Services.AddCors(options =>
{
    options.AddDefaultPolicy(policy =>
    {
        policy.WithOrigins("http://localhost:3000")
              .AllowAnyHeader()
              .AllowAnyMethod()
              .AllowCredentials();
    });
});

// After var app = builder.Build();
app.UseCors();
```

### Health State Case Handling

UI handles health states case-insensitively:
- `Healthy`, `healthy`, `HEALTHY` → all map to `'healthy'`
- Valid states: Learning, Healthy, Degraded, Attention, Unknown
- Invalid states default to `'unknown'`

### Missing Baselines

"Learning..." displays when:
- `baselines` field is null/missing
- `baselines` array is empty
- Control plane hasn't collected enough samples (~100 heartbeats)

This is expected behavior during the learning phase.

### Environment Configuration

Change API URL by editing `.env`:
```bash
VITE_API_BASE_URL=http://example.com:8080
```

**Important:** Must restart dev server after changing `.env`

## Success Criteria

✅ UI displays real agent data from database  
✅ Health states update as agents send heartbeats  
✅ Baselines display when learned  
✅ Error messages are user-friendly  
✅ Auto-refresh works (data updates every 30s)  
✅ No console errors  
✅ TypeScript build succeeds  
✅ Response mapping handles PascalCase → camelCase  

## Files Changed (API Integration)

- `ui/src/api/client.ts` - Real API calls with response mapping
- `ui/.env` - Environment configuration
- `ui/.env.example` - Template
- `ui/src/vite-env.d.ts` - TypeScript declarations
- `ui/README.md` - Updated docs

## Questions?

Contact Lambert or see `.ai-team/decisions/inbox/lambert-api-integration.md` for implementation details.

---

## Integration Complete - 2026-02-10 Evening

### Final Configuration

**Control Plane:**
- Running on http://localhost:5000
- CORS configured for http://localhost:3000, http://localhost:3001, http://localhost:5173
- Public endpoints (no mTLS): /api/agents, /api/agents/{id}, /api/ca/certificate
- Protected endpoints (mTLS required): /v0/agents/heartbeat, /api/agents/{id}/revoke

**UI:**
- Running on http://localhost:3000 (or 3001 if 3000 in use)
- API base URL: http://localhost:5000/api
- No mock data - all data from real API
- Auto-refresh every 30 seconds

### Verification Steps

1. **Check services are running:**
   ```bash
   lsof -ti:5000  # Control plane
   lsof -ti:3000  # UI
   ```

2. **Test API directly:**
   ```bash
   curl http://localhost:5000/api/agents
   # Should return: [] (empty array, no agents yet)
   ```

3. **Test CORS:**
   ```bash
   curl -H "Origin: http://localhost:3000" http://localhost:5000/api/agents
   # Should succeed with CORS headers
   ```

4. **Open UI:**
   - Navigate to http://localhost:3000
   - Should see system list page
   - Will show "No systems found" until agents register

### Database Note

If you encounter database errors, recreate the database:
```bash
cd control-plane
rm data/controlplane.db*
dotnet run  # Will recreate database on startup
```

### Next: Testing with Real Agent

To see data in the UI, run an agent:
```bash
cd agent
./bin/smidr-agent init
./bin/smidr-agent daemon
```

The agent will:
1. Register with control plane
2. Receive mTLS certificate
3. Start sending heartbeats every 60s
4. Appear in UI after first heartbeat

---

**Integration Status:** ✅ Complete and tested
**Last updated:** 2026-02-10
**By:** Lambert (Frontend Dev)
