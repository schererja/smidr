### 2026-02-11: Heartbeat 404 for Orphaned Certificates - Root Cause Analysis

**By:** Dallas  
**Context:** Agent with valid certificate receiving "404: Agent not registered" after control plane database reset  
**Investigation Date:** 2026-02-11

---

## Problem Statement

An agent with a cryptographically valid certificate (signed by the control plane CA) is unable to send heartbeats after the control plane database was recreated. The control plane returns:

```
404: Agent faf726fd-ccda-4cf6-971f-eafd039d65f3 not registered
```

The certificate is valid from a cryptographic perspective (properly signed, not expired, chain validates against CA), but the agent record no longer exists in the database. This creates an "orphaned certificate" scenario.

---

## Current Flow Analysis

### 1. Heartbeat Endpoint Behavior

**Location:** `control-plane/Controllers/HeartbeatController.cs:22-77`

**Flow when agent is not in database:**

```csharp
// Line 27: Extract agent ID from mTLS cert (via middleware)
var agentIdFromCert = HttpContext.Items["AgentId"] as string;

// Line 29-32: Validate request body contains agent ID
if (string.IsNullOrWhiteSpace(request.AgentId)) {
    return BadRequest("agentId is required");
}

// Line 34-37: Verify cert agent ID matches payload agent ID
if (!string.IsNullOrWhiteSpace(agentIdFromCert) && request.AgentId != agentIdFromCert) {
    return StatusCode(StatusCodes.Status403Forbidden, "Agent ID mismatch");
}

// Line 39-45: Database lookup - THIS IS WHERE IT FAILS
var agent = await _db.Agents
    .FirstOrDefaultAsync(a => a.Id == request.AgentId, cancellationToken);

if (agent == null) {
    return NotFound($"Agent {request.AgentId} not registered");  // 404
}

// Line 47-50: Check revocation status
if (agent.RevokedAt != null) {
    return StatusCode(StatusCodes.Status403Forbidden, "Agent certificate revoked");
}
```

**Key Insight:** The heartbeat endpoint returns `404 Not Found` when the database lookup fails, even though mTLS authentication succeeded. The endpoint never reaches the revocation check because the agent record doesn't exist.

---

### 2. mTLS Validation vs Registration Check

**Location:** `control-plane/Middleware/MtlsAuthenticationMiddleware.cs:24-57`

**Two-stage authentication:**

**Stage 1: Cryptographic Certificate Validation** (Middleware, Line 43-48)
```csharp
if (!mtlsValidation.ValidateClientCertificate(clientCert, out var errorMessage)) {
    context.Response.StatusCode = StatusCodes.Status403Forbidden;
    await context.Response.WriteAsync($"Certificate validation failed: {errorMessage}");
    return;
}
```

This validates:
- Certificate is signed by our CA (`MtlsValidationService.cs:24-45`)
- Certificate chain builds successfully with CustomRootTrust
- Certificate is not expired (`NotBefore` and `NotAfter` checks)
- **Does NOT check if agent exists in database**

**Stage 2: Database Registration Check** (Controller, Line 39-45)
```csharp
var agent = await _db.Agents
    .FirstOrDefaultAsync(a => a.Id == request.AgentId, cancellationToken);

if (agent == null) {
    return NotFound($"Agent {request.AgentId} not registered");
}
```

This checks:
- Agent ID exists in `Agents` table
- **Happens AFTER mTLS validation succeeds**

**Critical Distinction:**
- **Cryptographic validity:** "Is this certificate signed by our CA and not expired?"
- **Registration validity:** "Does this agent have a record in our database?"

An orphaned certificate passes Stage 1 but fails Stage 2.

---

### 3. Agent ID Extraction

**Location:** `control-plane/Services/MtlsValidationService.cs:57-63`

```csharp
public string? ExtractAgentId(X509Certificate2 clientCert)
{
    return clientCert.Subject.Split(',')
        .Select(s => s.Trim())
        .FirstOrDefault(s => s.StartsWith("CN=", StringComparison.OrdinalIgnoreCase))
        ?.Substring(3);
}
```

**Agent CSR Generation:** `agent/internal/agent/certificates.go:136-138`

```go
req := &x509.CertificateRequest{
    Subject: pkix.Name{CommonName: cfg.AgentID},  // Agent UUID goes here
}
```

**Flow:**
1. Agent generates UUID (e.g., `faf726fd-ccda-4cf6-971f-eafd039d65f3`)
2. Agent creates CSR with `CN=<UUID>`
3. Control plane signs CSR, certificate has `CN=faf726fd-ccda-4cf6-971f-eafd039d65f3`
4. Middleware extracts agent ID from certificate CN field
5. Controller uses extracted ID to query database

**Verification:** The agent ID in the certificate CN matches what's stored during enrollment (`AgentsController.cs:158-168`). The extraction logic is correct.

---

### 4. Registration Flow

**Location:** `control-plane/Controllers/AgentsController.cs:123-183`

**Registration stores:**
```csharp
existing = new Agent {
    Id = request.AgentId,              // From request body
    Hostname = request.Hostname,
    Token = request.Token,
    CertificatePem = certPem,          // Signed certificate
    RevokedAt = null,
    RegisteredAt = DateTime.UtcNow,
    OS = request.OS
};
_db.Agents.Add(existing);
await _db.SaveChangesAsync(cancellationToken);
```

**Key Fields:**
- `Id` (PK): Agent UUID, extracted from CSR CN during signing
- `CertificatePem`: The signed certificate (stored for audit purposes)
- `RevokedAt`: Revocation timestamp (null when active)

**Database Wipe Scenario:**
1. Agent registers → control plane creates `Agent` record with `Id = <UUID>`
2. Agent receives signed certificate with `CN=<UUID>`, saves to disk
3. **Database is recreated** → `Agent` record deleted
4. Agent tries to heartbeat with valid certificate
5. mTLS validation succeeds (cert is cryptographically valid)
6. Database lookup fails (no record with `Id = <UUID>`)
7. **404 returned**

---

## Current Error Code Analysis

### What Happens Now

**Error:** `404 Not Found` with message `"Agent {agentId} not registered"`

**Agent Response:** `agent/internal/heartbeat/client.go:94-100, 156-158`

```go
// First heartbeat immediately validates enrollment
if err := c.sendOnce(ctx); err != nil {
    if errors.Is(err, ErrAgentNotRegistered) {
        return err  // Fail fast - stops daemon
    }
    c.log.Warn("initial heartbeat failed", "error", err)
}

// 404 detection
if resp.StatusCode == http.StatusNotFound {
    return fmt.Errorf("%w: %s", ErrAgentNotRegistered, snippet)
}
```

**User-facing message:** `agent/internal/agent/daemon.go:64-68`

```go
if err != nil && errors.Is(err, heartbeat.ErrAgentNotRegistered) {
    log.Error("agent certificate is invalid - this usually happens when the control plane database was reset")
    log.Error("to fix this issue, run: smidr-agent reset-enrollment")
    log.Error("then restart the daemon to re-enroll with a new certificate")
    return fmt.Errorf("agent not registered in control plane (certificate is stale)")
}
```

**Good:** The agent already has excellent error handling for this scenario! It detects 404, logs clear instructions, and terminates gracefully.

---

## Semantic Correctness of HTTP Status Codes

### HTTP 404 Not Found

**RFC 9110 Definition:**  
> The 404 (Not Found) status code indicates that the origin server did not find a current representation for the target resource or is not willing to disclose that one exists.

**Semantic match:**  
✅ The resource `/v0/agents/heartbeat` for agent ID `faf726fd-...` does not exist in the database. This is technically correct.

### Alternative: HTTP 401 Unauthorized

**RFC 9110 Definition:**  
> The 401 (Unauthorized) status code indicates that the request has not been applied because it lacks valid authentication credentials for the target resource.

**Semantic match:**  
❌ The agent **does** have valid authentication credentials (certificate is cryptographically valid). Returning 401 would mislead the client into thinking the certificate itself is invalid.

### Alternative: HTTP 403 Forbidden

**RFC 9110 Definition:**  
> The 403 (Forbidden) status code indicates that the server understood the request but refuses to fulfill it.

**Semantic match:**  
❌ The server isn't refusing a valid request; the resource literally doesn't exist. Using 403 would suggest the agent is authenticated but not authorized for this action.

### Recommendation: Keep 404

**Rationale:**
- **Semantically correct:** The agent resource doesn't exist
- **Already handled by agent:** Agent code detects 404, logs clear error messages, and guides user to `reset-enrollment`
- **Distinct from auth failures:** 403 is already used for expired certs and agent ID mismatches
- **No confusion:** The agent doesn't retry on 404 - it terminates immediately with instructions

---

## Root Cause Summary

**Problem:** Orphaned certificates after database reset

**Why it happens:**
1. Certificate signing and database registration are **separate operations** with no transactional guarantee
2. Certificate has **longer lifetime** (365 days) than database persistence (can be reset anytime in dev)
3. mTLS validation is **stateless** - only checks cryptographic validity, not database presence
4. No built-in mechanism to detect database reset or invalidate old certificates

**Why 404 is correct:**
- Certificate is cryptographically valid → passes mTLS middleware
- Agent record doesn't exist → database lookup returns null
- Resource (agent) not found → 404 is semantically accurate
- Agent has excellent 404 handling → user gets clear instructions

---

## Recommendations

### 1. Keep Current Error Code (404) ✅

**No change needed.** The current implementation is semantically correct and the agent already handles it gracefully.

### 2. Consider Alternative Error Signaling (Optional)

If you want to explicitly distinguish "valid cert but not registered" from other 404 scenarios, consider adding a custom response header:

**Control plane change:**
```csharp
if (agent == null) {
    Response.Headers.Add("X-Smidr-Error", "agent-not-registered");
    return NotFound($"Agent {request.AgentId} not registered");
}
```

**Agent change:**
```go
if resp.StatusCode == http.StatusNotFound {
    errorType := resp.Header.Get("X-Smidr-Error")
    if errorType == "agent-not-registered" {
        return fmt.Errorf("%w: database reset detected", ErrAgentNotRegistered)
    }
    return fmt.Errorf("%w: %s", ErrAgentNotRegistered, snippet)
}
```

**Benefits:**
- More precise error reporting
- Enables different client behavior for different 404 scenarios
- Non-breaking (header is optional)

**Drawback:**
- Adds complexity for minimal benefit (current solution works fine)

### 3. Document Expected Behavior (Recommended) ✅

Add to `control-plane/README.md`:

```markdown
## Database Reset Handling

When the database is reset during development, agent certificates become "orphaned":

- **Certificate:** Still cryptographically valid (signed by CA, not expired)
- **Database:** Agent record no longer exists

**Behavior:**
- Heartbeat endpoint returns `404 Not Found`
- Agent detects orphaned certificate and terminates
- User must run `smidr-agent reset-enrollment` and restart daemon

**Why 404?** The agent resource literally doesn't exist in the database. The certificate is valid but references a non-existent agent.
```

### 4. Production Hardening (Future)

For production, consider:
- **Certificate revocation list (CRL):** Maintain a CRL that includes all certificates issued before the last database reset
- **Database backup/restore:** Never recreate production database from scratch
- **Certificate serial tracking:** Store certificate serials in a separate audit table that persists across resets
- **Health check endpoint:** Add `/v0/agents/{id}/ping` that returns 404 if agent not registered (allows pre-flight check)

---

## Coordination with Kane (Agent Dev)

**Kane's perspective:**
The agent already handles this scenario perfectly:
- Detects 404 on first heartbeat
- Returns `ErrAgentNotRegistered` 
- Daemon logs clear error with instructions
- User runs `reset-enrollment` command
- Agent re-enrolls with new certificate

**No agent changes needed** unless we want to implement custom header detection (optional enhancement).

**Recommended joint action:**
1. Add documentation to both agent and control plane READMEs explaining orphaned certificates
2. Consider adding a "health check" endpoint that agents can call on startup (separate from heartbeat)
3. Test the full reset-enrollment flow to ensure it works smoothly

---

## Conclusion

The current implementation is **correct and handles the orphaned certificate scenario gracefully**. The 404 error code is semantically accurate, and the agent provides excellent user guidance. No changes are required, but documentation improvements would help developers understand the expected behavior.

**Key Takeaway:** This is not a bug - it's expected behavior when database and certificate lifetimes diverge. The agent's `reset-enrollment` command is the correct recovery mechanism.
