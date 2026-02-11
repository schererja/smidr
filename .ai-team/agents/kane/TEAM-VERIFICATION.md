# Agent Registration Issue - Team Verification Tasks

**Context:** Agent has stale certificate after database reset. Kane fixed agent-side detection. Need to verify control plane and mTLS components.

## Dallas - Control Plane Verification

### Check HeartbeatController
**File:** `control-plane/Controllers/HeartbeatController.cs`

```bash
# Verify the endpoint returns proper 404 when agent doesn't exist
curl -X POST https://localhost:5001/v0/agents/heartbeat \
  -H "Content-Type: application/json" \
  -d '{"agentId":"non-existent-id","timestamp":"2024-01-01T00:00:00Z",...}' \
  --cert /path/to/any-valid-cert.pem \
  --key /path/to/any-valid-key.pem

# Expected: 404 with message "Agent {id} not registered"
```

### Check AgentsController Registration
**File:** `control-plane/Controllers/AgentsController.cs`

```bash
# Verify registration endpoint works
curl -X POST https://localhost:5001/api/agents/register \
  -H "Content-Type: application/json" \
  -d '{
    "agentId":"test-123",
    "hostname":"test-host",
    "csrPem":"-----BEGIN CERTIFICATE REQUEST-----...",
    "os":"darwin"
  }'

# Expected: 200 with certPem in response
```

### Questions for Dallas
1. Does HeartbeatController return 404 when agent ID from cert doesn't exist in DB?
2. Does the response body include "Agent {id} not registered" message?
3. Is `/api/agents/register` accessible without client certificate?
4. Does registration create agent in database with proper ID?

---

## Ash - mTLS Verification

### Check Certificate Validation
**File:** `control-plane/Services/MtlsValidationService.cs`

### Check CA Service
**File:** `control-plane/Services/CaService.cs`

### Questions for Ash
1. Does `ExtractAgentId()` properly parse CN from certificate subject?
2. Does `IssueAgentCertificate()` set the CN to the agent ID from the CSR?
3. Is the CA certificate valid and loaded on startup?
4. Do newly issued certificates work for mTLS authentication?

### Test Certificate Issuance
```bash
# Generate test CSR with agent ID in CN
openssl req -new -key test.key -out test.csr -subj "/CN=test-agent-123"

# Submit to registration endpoint
curl -X POST https://localhost:5001/api/agents/register \
  -H "Content-Type: application/json" \
  -d "{\"agentId\":\"test-agent-123\",\"hostname\":\"test\",\"csrPem\":\"$(cat test.csr | base64)\"}"

# Verify returned certificate has CN=test-agent-123
openssl x509 -in returned-cert.pem -noout -subject
```

---

## Integration Test (All Team)

Once individual components are verified, run full flow:

### 1. Clean State
```bash
# Stop agent
pkill smidr-agent

# Reset agent enrollment
cd /Users/schererja/src/github.com/schererja/smidr/agent
../bin/smidr-agent reset-enrollment

# Verify DB is clean (Dallas)
# SELECT * FROM Agents; -- should be empty or not have this agent
```

### 2. Fresh Enrollment
```bash
# Start agent daemon
../bin/smidr-agent daemon

# Watch logs - should see:
# - Key generation
# - CSR creation
# - Registration attempt
# - Certificate received
# - Heartbeat loop started
# - Successful heartbeats
```

### 3. Verify in UI
```bash
# Check https://localhost:5173 (Lambert)
# New agent should appear with:
# - Fresh agent ID (different from old faf726fd-ccda-4cf6-971f-eafd039d65f3)
# - Hostname
# - Green health status
# - Live metrics updating
```

### 4. Test Stale Certificate Detection
```bash
# Simulate database reset (Dallas)
# DROP TABLE Agents; CREATE TABLE Agents...;

# Restart agent (should still have cert from step 2)
../bin/smidr-agent daemon

# Should see:
# ERROR agent certificate is invalid
# ERROR run: smidr-agent reset-enrollment
# (exits)

# Reset and re-enroll
../bin/smidr-agent reset-enrollment
../bin/smidr-agent daemon

# Should successfully enroll with new ID
```

---

## If Still Failing

Check these common issues:

### Network/Connectivity
```bash
# Can agent reach control plane?
curl -k https://localhost:5001/api/agents

# Is control plane running?
ps aux | grep control-plane
```

### Certificates
```bash
# Is CA cert valid?
openssl x509 -in /path/to/ca.crt.pem -noout -dates

# Is agent cert issued by correct CA?
openssl verify -CAfile /path/to/ca.crt.pem /path/to/agent.crt.pem
```

### Configuration
```bash
# Check agent config
cat ~/Library/Application\ Support/smidr/agent.yaml

# Verify control_plane_url is correct
# Verify paths are correct
```

### Logs
```bash
# Agent logs (if using systemd)
journalctl -u smidr-agent -f

# Control plane logs
# Check console output or log files
```

---

**Next Action:** Each team member verify their component, then run integration test together.
