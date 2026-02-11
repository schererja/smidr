# Certificate and Enrollment Investigation Summary

## Investigation Date
2026-02-11

## Issue Description
Agent fails to heartbeat with `404 Not Found: Agent faf726fd-ccda-4cf6-971f-eafd039d65f3 not registered` after database was reset. Agent has existing certificate and believes it's enrolled.

## Root Cause
**Orphaned Certificate:** The agent has a cryptographically valid certificate, but the control plane database was wiped, removing the agent's registration record. This creates a mismatch between the certificate (which is still valid) and the database (which has no record of the agent).

## Certificate Flow Analysis

### 1. Enrollment (Working Correctly)
```
Agent generates:
- RSA 2048-bit private key
- CSR with Subject CN=<AgentID> (UUID)
- Sends CSR to POST /api/agents/register

Control Plane:
- Signs CSR with internal CA
- Preserves Subject from CSR (CN=<AgentID>)
- Stores agent record in database with ID=<AgentID>
- Returns signed certificate PEM

Agent:
- Saves certificate to ~/.smidr/agent.crt.pem
- Now considers itself "enrolled"
```

### 2. Heartbeat Authentication (Two-Stage Validation)

#### Stage 1: mTLS Validation (✅ PASSES for orphaned cert)
```
MtlsAuthenticationMiddleware:
1. Extracts client certificate from TLS connection
2. Validates certificate chain against internal CA
3. Checks certificate expiry (NotBefore/NotAfter)
4. Extracts Agent ID from certificate CN field
5. Stores agent ID in HttpContext.Items["AgentId"]

Result: ✅ Certificate is cryptographically valid
```

#### Stage 2: Database Validation (❌ FAILS for orphaned cert)
```
HeartbeatController:
1. Queries database: Agents.FirstOrDefaultAsync(a => a.Id == request.AgentId)
2. If agent == null → return 404 "Agent not registered"
3. If agent.RevokedAt != null → return 403 "Agent certificate revoked"
4. Otherwise → accept heartbeat and store metrics

Result: ❌ Agent not found in database
```

### 3. Agent Error Handling (✅ CORRECT)
```
Agent heartbeat client:
- Detects 404 status code
- Returns ErrAgentNotRegistered
- Daemon logs clear recovery instructions
- Terminates gracefully

Error message:
"agent certificate is invalid - this usually happens when the control plane database was reset"
"to fix this issue, run: smidr-agent reset-enrollment"
"then restart the daemon to re-enroll with a new certificate"
```

## Security Analysis

### Status Code Semantics
| HTTP Status | Meaning | When It Occurs | Agent Should |
|-------------|---------|----------------|--------------|
| 200 OK | Heartbeat accepted | Agent registered, cert valid | Continue heartbeat |
| 401 Unauthorized | No client cert | Missing certificate | Error (should be enrolled) |
| 403 Forbidden (cert invalid) | mTLS validation failed | Cert expired, wrong CA | Error (should re-enroll) |
| 403 Forbidden (revoked) | Agent revoked | Database: RevokedAt != null | Terminate (do NOT re-enroll) |
| 404 Not Found | Agent not registered | Database: agent record missing | Re-enroll (safe to retry) |

### Key Security Properties
✅ **mTLS validates cryptographic integrity** — Only CA-signed certificates pass  
✅ **Database enforces enrollment state** — Only registered agents can heartbeat  
✅ **Revocation is respected** — Revoked agents get 403, not 404  
✅ **404 vs 403 distinction enables safe auto-recovery** — Agent knows when re-enrollment is safe

## Recommendations

### For v0 (Current Behavior - Keep As-Is)
**Manual Recovery:**
1. User runs `smidr-agent reset-enrollment` to delete old certificate
2. User restarts daemon to trigger fresh enrollment
3. Agent generates new key/CSR and registers with control plane

**Rationale:**
- Behavior is correct and secure
- Database resets are rare in single-user development
- Clear error messages guide users to recovery
- Minimizes code changes for v0 release

### For v1 (Optional Enhancement)
**Auto-Re-Enrollment:**
- When agent receives 404 on heartbeat, automatically call reset-enrollment and re-enroll
- Safety: Max 3 attempts per daemon run, exponential backoff
- Only trigger on 404, never on 403 (revoked agents must not re-enroll)
- Aligns with "quiet, continuous assurance" philosophy

**Implementation:** See decision document for code example

## Current Status
✅ **Certificate generation:** Working correctly (RSA 2048, CN=<AgentID>)  
✅ **CA signing:** Working correctly (preserves subject, 365-day expiry)  
✅ **mTLS validation:** Working correctly (chain + expiry validation)  
✅ **Agent ID extraction:** Working correctly (parses CN field)  
✅ **Database lookup:** Working correctly (returns 404 when not found)  
✅ **Error handling:** Working correctly (clear recovery instructions)

## No Bugs Found
The system is working as designed. The 404 error is semantically correct for an orphaned certificate scenario. The agent provides clear recovery instructions. No code changes are required for v0.

## Files Involved
- **Agent:** `agent/internal/agent/certificates.go` (CSR generation with CN=<AgentID>)
- **Agent:** `agent/internal/agent/daemon.go` (enrollment + error handling)
- **Agent:** `agent/internal/heartbeat/client.go` (404 detection)
- **Control Plane:** `control-plane/Services/CaService.cs` (CSR signing)
- **Control Plane:** `control-plane/Services/MtlsValidationService.cs` (cert validation + CN extraction)
- **Control Plane:** `control-plane/Middleware/MtlsAuthenticationMiddleware.cs` (request auth)
- **Control Plane:** `control-plane/Controllers/HeartbeatController.cs` (database lookup)
- **Control Plane:** `control-plane/Controllers/AgentsController.cs` (registration)
