# Orphaned Certificates - Understanding Database Reset Behavior

## Overview

When the control plane database is reset during development, agent certificates become "orphaned." This document explains why this happens, how the system handles it, and why the current behavior is correct.

## What is an Orphaned Certificate?

An orphaned certificate is a certificate that is:
- ✅ **Cryptographically valid** (signed by the control plane CA, not expired, chain validates)
- ❌ **Not registered in the database** (agent record was deleted)

## When Does This Happen?

**Common Development Scenario:**
1. Agent enrolls with control plane
2. Control plane creates agent record in database
3. Agent receives signed certificate (valid for 365 days)
4. **Developer recreates database** (schema changes, testing, etc.)
5. Agent record is deleted but certificate file still exists on agent
6. Agent tries to heartbeat with orphaned certificate

**Production Risk:** Minimal (databases should never be recreated from scratch in production)

## System Behavior

### Control Plane Flow

**Stage 1: mTLS Validation** (`MtlsAuthenticationMiddleware.cs`)
```
1. Extract client certificate from TLS connection
2. Validate certificate chain against internal CA
3. Check certificate expiry (NotBefore/NotAfter)
4. Extract agent ID from certificate CN field
5. Set HttpContext.Items["AgentId"] for downstream use
→ Result: PASS (certificate is cryptographically valid)
```

**Stage 2: Database Lookup** (`HeartbeatController.cs`)
```
6. Query database for agent record by ID
7. Check if agent exists
→ Result: FAIL (agent record not found)
8. Return: 404 Not Found "Agent {id} not registered"
```

### Agent Behavior

**Detection:** `internal/heartbeat/client.go:156-158`
```go
if resp.StatusCode == http.StatusNotFound {
    return fmt.Errorf("%w: %s", ErrAgentNotRegistered, snippet)
}
```

**Error Handling:** `internal/agent/daemon.go:64-68`
```go
if err != nil && errors.Is(err, heartbeat.ErrAgentNotRegistered) {
    log.Error("agent certificate is invalid - this usually happens when the control plane database was reset")
    log.Error("to fix this issue, run: smidr-agent reset-enrollment")
    log.Error("then restart the daemon to re-enroll with a new certificate")
    return fmt.Errorf("agent not registered in control plane (certificate is stale)")
}
```

**Result:** Agent daemon terminates with clear instructions

## Why Return 404 (Not 401 or 403)?

### HTTP Semantics

**404 Not Found** (current implementation)
- **Meaning:** The requested resource does not exist
- **Match:** ✅ The agent record literally doesn't exist in the database
- **Client behavior:** Agent detects orphaned cert, instructs user to reset enrollment

**401 Unauthorized** (not used)
- **Meaning:** The request lacks valid authentication credentials
- **Mismatch:** ❌ The certificate IS valid - mTLS authentication succeeded
- **Client behavior:** Would mislead agent into thinking cert is cryptographically invalid

**403 Forbidden** (already used for other cases)
- **Meaning:** Server understood request but refuses to fulfill it
- **Mismatch:** ❌ Server isn't refusing; the resource doesn't exist
- **Current use:** Revoked certificates, agent ID mismatch

**Conclusion:** 404 is semantically correct per RFC 9110

## Recovery Process

**User Command:**
```bash
smidr-agent reset-enrollment
```

**What it does:**
1. Deletes certificate files (`agent.crt.pem`, `agent.key.pem`, `agent.csr.pem`)
2. Clears agent ID from config file
3. Preserves other config settings

**Next Steps:**
```bash
systemctl restart smidr-agent  # or: smidr-agent daemon
```

**What happens:**
1. Agent detects missing certificate
2. Agent generates new key pair and CSR
3. Agent calls `/api/agents/register`
4. Control plane creates new agent record
5. Agent receives new certificate
6. Heartbeat loop starts successfully

## Architecture Insights

### Two-Stage Authentication

The control plane uses two independent checks:

**Stage 1: Cryptographic Validation**
- **Purpose:** Verify certificate is signed by our CA and not expired
- **Stateless:** No database access required
- **Fast:** Chain validation only
- **Location:** `MtlsValidationService.ValidateClientCertificate()`

**Stage 2: Registration Check**
- **Purpose:** Verify agent exists in our system
- **Stateful:** Database lookup required
- **Authoritative:** Database is source of truth for active agents
- **Location:** `HeartbeatController.Heartbeat()` line 39-45

**Why separate?**
- Certificate lifetime (365 days) vs database persistence (can be reset)
- Certificate is proof of identity, database is proof of registration
- Allows detection of orphaned certificates (valid cert, no registration)

### Agent ID Flow

**Generation:** Agent generates UUID v4 on first run
```go
// agent/internal/agent/uuid.go
id := uuid.New().String()
```

**CSR Creation:** Agent uses UUID as Common Name
```go
// agent/internal/agent/certificates.go:136-138
req := &x509.CertificateRequest{
    Subject: pkix.Name{CommonName: cfg.AgentID},  // CN=<UUID>
}
```

**Certificate Signing:** Control plane signs CSR, preserves CN
```csharp
// control-plane/Services/CaService.cs:54
var cert = csr.Create(
    _issuerCert.SubjectName,
    signatureGenerator,
    notBefore,
    notAfter,
    serial);  // Signed cert has CN=<UUID>
```

**Database Storage:** Control plane stores agent ID from request body
```csharp
// control-plane/Controllers/AgentsController.cs:159
existing = new Agent {
    Id = request.AgentId,  // UUID from agent
    ...
};
```

**mTLS Extraction:** Middleware extracts agent ID from certificate CN
```csharp
// control-plane/Services/MtlsValidationService.cs:59-62
return clientCert.Subject.Split(',')
    .Select(s => s.Trim())
    .FirstOrDefault(s => s.StartsWith("CN=", StringComparison.OrdinalIgnoreCase))
    ?.Substring(3);  // Returns UUID
```

**Result:** Agent ID is consistent across CSR → certificate → database → heartbeat validation

## Production Considerations

### Prevention Strategies

1. **Never recreate production databases**
   - Use migrations for schema changes
   - Use backups for disaster recovery
   - Never drop and recreate tables

2. **Certificate Revocation List (CRL)**
   - Maintain CRL with serials of all pre-reset certificates
   - Check CRL during mTLS validation
   - Automatically revoke orphaned certificates

3. **Certificate Serial Tracking**
   - Store certificate serials in separate audit table
   - Persist audit table across schema changes
   - Detect orphaned certificates by serial mismatch

4. **Database Backup/Restore**
   - Regular automated backups
   - Test restore procedures
   - Coordinate certificate lifetime with backup retention

### Monitoring

Add alerts for:
- High rate of 404 responses on heartbeat endpoint
- Agents repeatedly failing enrollment
- Database restoration events

### Documentation

Ensure ops teams understand:
- Database reset requires all agents to re-enroll
- Certificate files on agents must be deleted before restart
- Use `reset-enrollment` command, not manual file deletion

## Testing

### Reproduce Orphaned Certificate

```bash
# 1. Start control plane
cd control-plane
dotnet run

# 2. Enroll agent
cd agent
smidr-agent daemon &
# Wait for enrollment to complete

# 3. Delete database
cd control-plane
rm data/controlplane.db*

# 4. Restart control plane (runs migrations, creates new empty DB)
dotnet run

# 5. Agent next heartbeat will return 404
# Check agent logs for "agent not registered" error

# 6. Recover
smidr-agent reset-enrollment
smidr-agent daemon
# Agent re-enrolls successfully
```

### Verify Recovery

```bash
# Check agent logs
journalctl -u smidr-agent -f

# Expected output:
# "registration successful"
# "heartbeat sent"
```

## Related Documentation

- [Control Plane README](../README.md) - Setup and configuration
- [API Documentation](../../docs/API.md) - Endpoint specifications
- [mTLS Security](../README-CRYPTO.md) - Certificate architecture
- [Agent Documentation](../../agent/README.md) - Agent setup

## Summary

**Orphaned certificates are expected behavior when database and certificate lifetimes diverge.**

- **Detection:** Agent receives 404 on heartbeat
- **Recovery:** Run `smidr-agent reset-enrollment` and restart
- **Architecture:** Two-stage authentication separates cryptographic validity from registration status
- **Correctness:** 404 is semantically accurate (agent resource doesn't exist)
- **User Experience:** Clear error messages guide users to recovery command

**No code changes required** - current implementation handles this scenario correctly.
