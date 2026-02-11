### 2026-02-11: Orphaned Certificate Analysis and Resolution

**By:** Ash (Crypto/mTLS Engineer)

**What:** Diagnosed "agent not registered" issue after database reset. Agent has valid certificate but control plane database was wiped, causing 404 errors on heartbeat.

**Why:** Understanding this failure mode is critical for operational resilience and guides implementation of auto-recovery features.

## Problem Statement

**Scenario:** Database was recreated (all agent registrations wiped). Agent has existing certificate at `~/.smidr/agent.crt.pem` and believes it's enrolled.

**Symptoms:**
- Agent heartbeat fails with `404 Not Found: Agent {id} not registered`
- Agent logs: "Run smidr-agent reset-enrollment, then restart"
- Agent terminates after detecting 404

## Root Cause Analysis

### mTLS Authentication Flow (Current)

1. **mTLS Validation (✅ SUCCEEDS)**
   - `MtlsAuthenticationMiddleware` validates certificate cryptographically
   - Certificate chain verification against internal CA: ✅ PASSES
   - Certificate expiry check (NotBefore/NotAfter): ✅ PASSES
   - Agent ID extracted from certificate CN field: `faf726fd-ccda-4cf6-971f-eafd039d65f3`

2. **Database Lookup (❌ FAILS)**
   - `HeartbeatController` queries database for agent ID
   - `FirstOrDefaultAsync(a => a.Id == request.AgentId)` returns `null`
   - Returns `404 Not Found: Agent {id} not registered`

3. **Agent Response (✅ CORRECT)**
   - Detects `404` status code
   - Returns `ErrAgentNotRegistered` error
   - Daemon logs clear instructions and terminates

### Why This Happens

**"Orphaned Certificates"** — Certificates are cryptographically valid but reference non-existent database records.

**Causes:**
- Database reset/migration without certificate revocation
- Agent enrollment failed partway (cert issued, DB write failed)
- Database backup restored from earlier state
- Multi-region deployment with DB replication lag

### Certificate vs. Database Consistency

The system has **two sources of truth**:
1. **Cryptographic Trust:** CA-signed certificate proves identity
2. **Database State:** Agent record proves registration

**Current behavior:** BOTH must be valid for heartbeat to succeed.

## Security Analysis

### Current Design is Correct

✅ **404 is semantically correct per RFC 9110** — Resource (agent) doesn't exist  
✅ **mTLS validates cryptographic integrity** — Certificate is genuine  
✅ **Database enforces enrollment state** — Only registered agents can heartbeat  
✅ **Agent handles 404 gracefully** — Provides clear error message and terminates

### Differentiation: Revoked vs. Orphaned

| Scenario | mTLS Validation | Database Lookup | HTTP Status | Meaning |
|----------|----------------|-----------------|-------------|---------|
| Valid enrollment | ✅ Pass | ✅ Found | 200 OK | Agent is registered and healthy |
| Orphaned cert | ✅ Pass | ❌ Not found | 404 Not Found | Agent not in database (safe to re-enroll) |
| Revoked cert | ✅ Pass | ✅ Found (RevokedAt set) | 403 Forbidden | Agent explicitly revoked (do NOT re-enroll) |
| Invalid cert | ❌ Fail | N/A | 403 Forbidden | Certificate cryptographically invalid |

**Key Insight:** 404 vs 403 distinction allows agent to know when re-enrollment is safe:
- **404:** "I'm not in the database" → Safe to re-enroll
- **403:** "I was explicitly revoked" → Do NOT re-enroll

## Recommended Solutions

### Option 1: Auto-Re-Enrollment (RECOMMENDED)

**Agent-side change:** When heartbeat receives 404, automatically re-enroll.

**Advantages:**
- Self-healing behavior aligns with "quiet, continuous assurance"
- No manual intervention required after database resets
- Maintains security (only 404 triggers re-enrollment, not 403)

**Implementation:**
```go
func RunDaemon(ctx context.Context, cfg Config, log *logging.Logger) error {
    // ... existing enrollment check ...
    
    err = client.Run(ctx)
    
    // If heartbeat failed due to 404, auto-re-enroll
    if err != nil && errors.Is(err, heartbeat.ErrAgentNotRegistered) {
        log.Warn("agent not registered in control plane, attempting re-enrollment")
        
        // Reset enrollment (delete old cert)
        if err := ResetEnrollment(cfg); err != nil {
            return fmt.Errorf("reset enrollment: %w", err)
        }
        
        // Re-enroll with new certificate
        if err := enrollAgent(ctx, cfg, log); err != nil {
            return fmt.Errorf("re-enrollment failed: %w", err)
        }
        
        log.Info("re-enrollment successful, restarting heartbeat")
        // Restart heartbeat (tail recursion or loop)
    }
    
    return err
}
```

**Safety measures:**
- Add retry counter (max 3 re-enrollment attempts per daemon run)
- Add exponential backoff between attempts
- Only trigger on 404, never on 403 (revoked agents must not re-enroll)

### Option 2: Manual Recovery (CURRENT)

**Status:** Already implemented  
**Behavior:** Agent logs clear instructions, user runs `reset-enrollment`, restarts daemon

**Advantages:**
- Simple, no new code required
- User maintains control over re-enrollment
- Already documented in error messages

**Disadvantages:**
- Requires manual intervention after database resets
- Not "quiet" — alerts operators unnecessarily
- Slower recovery time

### Option 3: Database Sync Check (FUTURE)

**Agent-side change:** Before entering heartbeat loop, verify agent exists in database.

**Advantages:**
- Catches orphaned certs before first heartbeat failure
- Could fetch agent metadata (last heartbeat, health status)

**Disadvantages:**
- Requires new control plane endpoint (`GET /api/agents/{id}/status`)
- Adds latency to daemon startup
- Doesn't eliminate need for auto-re-enrollment (race conditions still possible)

## Control Plane Enhancements (Optional)

### Custom Error Header

Add `X-Smidr-Error` header to 404 responses for more precise error signaling:

```csharp
if (agent == null)
{
    Response.Headers["X-Smidr-Error"] = "agent-not-registered";
    return NotFound($"Agent {request.AgentId} not registered");
}
```

**Agent-side:** Parse header to differentiate error types without string matching.

### No Other Changes Required

- ✅ mTLS validation is correct
- ✅ Database lookup is correct
- ✅ 404 status code is semantically correct
- ✅ Revocation handling (403) is correct

## Testing Recommendations

### Unit Tests

1. **Agent:** `TestRunDaemon_AutoReenrollOn404`
2. **Control Plane:** `TestHeartbeat_Returns404ForUnknownAgent`
3. **Control Plane:** `TestHeartbeat_Returns403ForRevokedAgent`

### Integration Tests

1. Enroll agent → wipe database → verify auto-re-enrollment succeeds
2. Enroll agent → revoke → verify agent does NOT re-enroll on 403
3. Enroll agent → expire certificate → verify mTLS rejects at TLS layer

### Manual Testing

```bash
# 1. Enroll agent
smidr-agent daemon

# 2. Wipe control plane database
rm control-plane/data/smidr.db
dotnet run --project control-plane

# 3. Verify agent auto-re-enrolls (if Option 1 implemented)
# OR verify agent logs clear error message (current behavior)
```

## Decision for v0

**Recommended:** Keep current manual recovery behavior for v0, document clearly.

**Rationale:**
- Current behavior is secure and correct
- v0 scope is single-user local development
- Database resets are rare in production
- Can add auto-re-enrollment in v1 after user feedback

**Documentation updates needed:**
- Add "Orphaned Certificates" section to MTLS_SECURITY.md
- Add troubleshooting guide to control plane README
- Update agent README with recovery procedure

## Implementation Plan (if auto-re-enrollment chosen for v1)

1. **Kane:** Implement auto-re-enrollment logic in `daemon.go`
2. **Kane:** Add retry counter and exponential backoff
3. **Kane:** Write unit tests for re-enrollment flow
4. **Dallas:** Add `X-Smidr-Error` header to 404 responses (optional)
5. **Parker:** Write integration tests for database reset scenario
6. **Brett:** Document auto-recovery behavior

**Estimated effort:** 2-4 hours (agent changes + tests)
