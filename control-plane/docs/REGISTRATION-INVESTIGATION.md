# Registration Flow Investigation - Summary

## Problem
Agent was failing with error:
```
"initial heartbeat failed","error":"heartbeat failed with status 404: Agent faf726fd-ccda-4cf6-971f-eafd039d65f3 not registered"
```

## Investigation Findings

### 1. Registration Flow is Correctly Implemented

**Agent side (daemon.go):**
- On startup, agent checks if certificate exists
- If no cert, calls `POST /api/agents/register` with CSR
- Receives signed certificate and saves to disk
- Then starts heartbeat loop using mTLS

**Control plane side (AgentsController.cs):**
- Registration endpoint exists at `POST /api/agents/register`
- Accepts: agentId, hostname, csrPem, optional token, optional OS
- Signs CSR using CaService
- Saves agent to database (Agents table)
- Returns signed certificate

**Heartbeat side (HeartbeatController.cs):**
- Heartbeat endpoint at `POST /v0/agents/heartbeat`
- Requires mTLS authentication
- Checks if agent exists in database
- Returns 404 if agent not found

### 2. Root Cause: Database Initialization Method

**Problem found in Program.cs line 66:**
```csharp
db.Database.EnsureCreated();  // WRONG
```

**What EnsureCreated() does:**
- Creates database if it doesn't exist
- Uses current schema snapshot
- **Ignores all migrations**

**Why this breaks:**
1. Database created before AddOSToAgent migration exists → no OS column
2. Migration added later for OS field
3. Code expects OS column, but database doesn't have it
4. **More critically**: If database gets corrupted or deleted, registration might not persist correctly

**Fix applied:**
```csharp
db.Database.Migrate();  // CORRECT
```

**What Migrate() does:**
- Creates database if needed
- Applies all pending migrations in order
- Updates schema to match current model
- Idempotent (safe to run multiple times)

## Files Changed

### 1. Program.cs
- **Changed**: Line 66 from `EnsureCreated()` to `Migrate()`
- **Why**: Ensures database schema always matches code expectations
- **Impact**: Auto-applies migrations on startup, no manual intervention needed

### 2. Created Diagnostic Tools

**check-database.sh**
- Verifies database exists
- Shows table schemas
- Lists migration history
- Displays registered agents
- Checks for OS column

**test-registration.sh**
- Generates test CSR
- Calls registration endpoint
- Verifies agent in database
- Tests full registration flow

**docs/TROUBLESHOOTING-REGISTRATION.md**
- Comprehensive troubleshooting guide
- Documents registration flow
- Lists common issues and fixes
- Provides manual testing procedures
- Database schema reference

## Solution Steps

### For Existing Installation

If agent is failing with "not registered" error:

1. **Check database state:**
   ```bash
   cd control-plane
   ./check-database.sh
   ```

2. **If OS column is missing, reset database:**
   ```bash
   cd control-plane
   ./fix-migrations.sh
   ```
   (This deletes database and recreates with all migrations)

3. **Or, update to new Program.cs and restart control plane:**
   ```bash
   cd control-plane
   dotnet run
   ```
   (New code auto-applies migrations)

4. **Delete agent certificate to trigger re-registration:**
   ```bash
   sudo rm /etc/smidr/agent.crt.pem
   sudo systemctl restart smidr-agent
   ```

5. **Monitor agent logs:**
   ```bash
   journalctl -u smidr-agent -f
   ```
   Should see "registration successful" then heartbeats

### For Fresh Installation

With Program.cs fix:
1. Start control plane: `dotnet run`
2. Database created with all migrations applied automatically
3. Start agent: `smidr-agent daemon`
4. Agent registers and starts heartbeat loop
5. No manual intervention needed

## Testing Registration

### Quick Test
```bash
cd control-plane
./test-registration.sh
```

### Manual Test with Real Agent
```bash
# Terminal 1: Start control plane
cd control-plane
dotnet run

# Terminal 2: Check database before agent starts
cd control-plane
./check-database.sh

# Terminal 3: Start agent
cd agent
go build -o bin/smidr-agent ./cmd/agent
sudo ./bin/smidr-agent daemon --config /etc/smidr/agent.toml

# Terminal 4: Watch agent logs
journalctl -u smidr-agent -f

# Terminal 2: Check database after agent starts
cd control-plane
./check-database.sh
```

Expected outcome:
- Agent logs show "registration successful"
- Agent logs show "heartbeat sent" every 60 seconds
- Database shows agent record with LastHeartbeatAt updating
- No "404 not registered" errors

## Verification

After fix, verify:

1. **Control plane starts without errors**
2. **Database has Agents table with OS column**
3. **Agent can register successfully**
4. **Agent can send heartbeats after registration**
5. **Control plane stores heartbeats in database**

## Additional Notes

### Why Registration Might Still Fail

Even with this fix, registration can fail if:
- Control plane not reachable (network/firewall)
- Invalid CSR format
- Certificate directory not writable by agent
- Control plane database locked/corrupted

Use the troubleshooting guide and diagnostic scripts to diagnose these issues.

### When to Use fix-migrations.sh

Only use in development when:
- Database schema is corrupted
- Migration history out of sync
- Testing clean slate scenarios

**DO NOT** use in production - you'll lose all data. Production requires careful migration planning and backups.

## References

- `control-plane/Program.cs` - Database initialization
- `control-plane/Controllers/AgentsController.cs` - Registration endpoint
- `control-plane/Controllers/HeartbeatController.cs` - Heartbeat endpoint
- `agent/internal/agent/daemon.go` - Agent registration logic
- `agent/internal/heartbeat/client.go` - Heartbeat client logic
- `control-plane/docs/TROUBLESHOOTING-REGISTRATION.md` - Full troubleshooting guide
