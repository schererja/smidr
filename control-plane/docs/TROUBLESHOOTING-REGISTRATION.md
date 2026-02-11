# Agent Registration Troubleshooting Guide

This guide helps diagnose and fix issues with agent registration and heartbeat flow.

## Registration Flow Overview

1. **Agent starts** (`smidr-agent daemon`)
2. **Check for certificate**: If `agent.crt.pem` exists, skip to heartbeat
3. **Register with control plane**:
   - POST to `/api/agents/register` with CSR
   - Control plane signs certificate and stores agent in database
   - Agent saves certificate locally
4. **Start heartbeat loop**:
   - POST to `/v0/agents/heartbeat` with mTLS
   - Control plane verifies agent exists, stores heartbeat, triggers health evaluation

## Common Issues

### 1. "Agent not registered" (404) on Heartbeat

**Symptoms:**
```
initial heartbeat failed","error":"heartbeat failed with status 404: Agent faf726fd-ccda-4cf6-971f-eafd039d65f3 not registered
```

**Root causes:**
- Agent never completed registration (network failure, invalid CSR, etc.)
- Database missing agent record (database corruption, manual deletion)
- Agent ID mismatch between registration and heartbeat

**Diagnosis:**
```bash
cd control-plane
./check-database.sh
```

Check:
- Is the agent ID in the database?
- Does the agent have a certificate file?
- Check agent logs for registration success message

**Fix:**
1. Delete agent's certificate: `rm /etc/smidr/agent.crt.pem`
2. Restart agent to trigger re-registration: `systemctl restart smidr-agent`
3. Watch logs: `journalctl -u smidr-agent -f`

### 2. "no such column: a.OS" Database Error

**Symptoms:**
```
SQLite Error 1: 'no such column: a.OS'
```

**Root cause:**
Database was created with `EnsureCreated()` before AddOSToAgent migration existed, so OS column is missing.

**Fix:**
```bash
cd control-plane
./fix-migrations.sh
```

This deletes and recreates the database with all migrations applied.

**Note:** Only safe in development. Production requires backup and manual migration.

### 3. Registration Fails with Invalid CSR

**Symptoms:**
```
registration attempt failed: register agent failed: csrPem invalid: <error details>
```

**Root cause:**
- Corrupted CSR file
- Wrong CSR format
- Key/CSR mismatch

**Fix:**
```bash
# Delete old key/CSR to regenerate
rm /etc/smidr/agent.key.pem /etc/smidr/agent.csr.pem
systemctl restart smidr-agent
```

### 4. Connection Refused / Timeout

**Symptoms:**
```
register agent: Post "https://localhost:5001/api/agents/register": dial tcp [::1]:5001: connect: connection refused
```

**Root cause:**
Control plane not running or wrong URL

**Fix:**
1. Verify control plane is running:
   ```bash
   cd control-plane
   dotnet run
   ```
2. Check control plane URL in agent config:
   ```bash
   cat /etc/smidr/agent.toml
   # Should show: control_plane_url = "https://localhost:5001"
   ```
3. Test connectivity:
   ```bash
   curl -k https://localhost:5001/ca/certificate
   ```

## Manual Testing

### Test Registration Endpoint

```bash
cd control-plane
./test-registration.sh
```

This script:
- Generates a test CSR
- Calls the registration endpoint
- Verifies the agent is in the database
- Cleans up test data

### Check Database State

```bash
cd control-plane
./check-database.sh
```

This script:
- Verifies database exists
- Checks table schema
- Lists migration history
- Shows registered agents

### Verify Control Plane Is Running

```bash
# Check process
ps aux | grep -i dotnet

# Check port
lsof -i :5001

# Test API
curl -k https://localhost:5001/ca/certificate
```

### Check Agent State

```bash
# View agent config
cat /etc/smidr/agent.toml

# Check if certificate exists
ls -la /etc/smidr/*.pem

# View agent logs
journalctl -u smidr-agent -n 50

# Follow agent logs
journalctl -u smidr-agent -f
```

## Database Schema

### Agents Table
```sql
CREATE TABLE "Agents" (
    "Id" TEXT NOT NULL PRIMARY KEY,
    "Hostname" TEXT NOT NULL,
    "Token" TEXT,
    "CertificatePem" TEXT,
    "RevokedAt" TEXT,
    "RegisteredAt" TEXT NOT NULL,
    "LastHeartbeatAt" TEXT,
    "CurrentHealth" TEXT NOT NULL,
    "OS" TEXT
);
```

Key fields:
- `Id`: Agent UUID
- `RegisteredAt`: When agent first enrolled
- `LastHeartbeatAt`: Last successful heartbeat
- `CurrentHealth`: Learning/Healthy/Degraded/Attention
- `OS`: Operating system (darwin/linux/windows)

## Endpoint Reference

### Registration
- **URL**: `POST /api/agents/register`
- **Auth**: None (pre-enrollment)
- **Body**:
  ```json
  {
    "agentId": "uuid",
    "hostname": "my-server",
    "csrPem": "-----BEGIN CERTIFICATE REQUEST-----\n...",
    "os": "linux"
  }
  ```
- **Response**:
  ```json
  {
    "certPem": "-----BEGIN CERTIFICATE-----\n..."
  }
  ```

### Heartbeat
- **URL**: `POST /v0/agents/heartbeat`
- **Auth**: mTLS (requires agent certificate)
- **Body**:
  ```json
  {
    "agentId": "uuid",
    "timestamp": "2026-02-11T10:00:00Z",
    "uptimeSeconds": 86400.0,
    "loadAverage1m": 0.5,
    "memoryUsedPct": 45.2,
    "diskUsedPct": 60.1,
    "processCount": 120
  }
  ```
- **Response**: `200 OK` (empty body)

## Development Notes

### EnsureCreated vs Migrate

**Before:**
```csharp
db.Database.EnsureCreated(); // WRONG: Doesn't apply migrations
```

**After:**
```csharp
db.Database.Migrate(); // CORRECT: Applies all pending migrations
```

`EnsureCreated()` creates the database with the current schema but **ignores migrations**. This causes issues when schema changes are added via migrations after the database already exists.

`Migrate()` applies all pending migrations in order, ensuring the database schema matches the current model.

### Migration Workflow

1. Make model changes (e.g., add OS property to Agent)
2. Create migration:
   ```bash
   dotnet ef migrations add AddOSToAgent
   ```
3. Apply migration:
   ```bash
   dotnet ef database update
   ```
4. Verify:
   ```bash
   dotnet ef migrations list
   ```

### Testing in Development

1. Start control plane:
   ```bash
   cd control-plane
   dotnet run
   ```

2. In another terminal, check database:
   ```bash
   cd control-plane
   ./check-database.sh
   ```

3. Test registration:
   ```bash
   cd control-plane
   ./test-registration.sh
   ```

4. Start agent:
   ```bash
   cd agent
   go build -o bin/smidr-agent ./cmd/agent
   sudo ./bin/smidr-agent daemon --config /etc/smidr/agent.toml
   ```

5. Watch agent logs for registration and heartbeat success

## See Also

- `README-migrations.md` - Migration management guide
- `fix-migrations.sh` - Database reset script
- `check-database.sh` - Database inspection script
- `test-registration.sh` - Registration endpoint test
