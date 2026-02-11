### 2026-02-11: Agent 404 error handling and stale certificate detection

**By:** Kane

**What:** Analysis of agent startup/validation flow after database reset caused 404 errors

**Problem Context:**
The agent was failing to heartbeat with "404: Agent faf726fd-ccda-4cf6-971f-eafd039d65f3 not registered" after the control plane database was recreated. The agent had a valid certificate file but the control plane no longer recognized the agent ID.

**Investigation Findings:**

## 1. Current Startup Logic

**File:** `agent/internal/agent/daemon.go` lines 26-38

The agent checks if it's enrolled by simply checking if a certificate file exists:

```go
enrolled := false
if _, err := LoadCert(cfg.CertPath); err == nil {
    enrolled = true
    log.Info("certificate found", "cert_path", cfg.CertPath)
}

if !enrolled {
    if err := enrollAgent(ctx, cfg, log); err != nil {
        return err
    }
}
```

**Problem:** The agent does NOT validate whether the certificate is still recognized by the control plane. It only checks for file existence. This means a stale certificate (from a wiped database) passes the enrollment check.

## 2. Certificate Validation

The agent performs NO validation of certificate validity with the control plane at startup. There is:
- No check if the agent ID in the cert matches what the control plane expects
- No verification that the control plane still recognizes this agent
- No expiration date checking
- No attempt to verify the certificate chain

The first validation happens when the heartbeat is sent, at which point the control plane returns 404.

## 3. Agent ID Source

**File:** `agent/internal/agent/certificates.go` lines 136-142

The agent ID comes from two sources:
1. **Config file** (`agent_id` field) - generated on first startup if empty (via `PopulateRuntimeConfig()`)
2. **Certificate CN** - when CSR is created, the agent ID is embedded in the Common Name field

The agent stores the agent ID persistently in `config.yaml` (or platform-specific config path). The certificate's CN field contains the same agent ID. The heartbeat client reads the agent ID from the config file, not from the certificate.

**Files:**
- `agent/internal/agent/runtime_config.go` lines 21-28: Generates UUID if `cfg.AgentID == ""`
- `agent/internal/agent/certificates.go` line 137: CSR embeds `cfg.AgentID` in CN
- `agent/internal/heartbeat/client.go` lines 46-48: Uses `cfg.AgentID` from config

## 4. Error Handling

**File:** `agent/internal/heartbeat/client.go` lines 83-84, 94-100, 109-113, 156-158

The heartbeat client DOES detect 404 errors and treats them specially:

```go
var ErrAgentNotRegistered = errors.New("agent not registered in control plane")

// In sendOnce():
if resp.StatusCode == http.StatusNotFound {
    return fmt.Errorf("%w: %s", ErrAgentNotRegistered, snippet)
}

// In Run():
if err := c.sendOnce(ctx); err != nil {
    if errors.Is(err, ErrAgentNotRegistered) {
        return err  // Fail fast - don't retry
    }
    c.log.Warn("initial heartbeat failed", "error", err)
}
```

When a 404 is detected, the error propagates back to `daemon.go`:

**File:** `agent/internal/agent/daemon.go` lines 63-70

```go
err = client.Run(ctx)

if err != nil && errors.Is(err, heartbeat.ErrAgentNotRegistered) {
    log.Error("agent certificate is invalid - this usually happens when the control plane database was reset")
    log.Error("to fix this issue, run: smidr-agent reset-enrollment")
    log.Error("then restart the daemon to re-enroll with a new certificate")
    return fmt.Errorf("agent not registered in control plane (certificate is stale)")
}
```

**Good:** The agent provides clear guidance to the user about how to fix the problem.

**Problem:** The agent exits and requires manual intervention. No automatic recovery.

## 5. Recommendation: Automatic Re-enrollment on 404

**Should the agent auto-re-enroll when it gets 404?**

### Option A: Automatic Re-enrollment (Recommended)

**Pros:**
- Self-healing behavior - agent recovers without manual intervention
- Better user experience - no downtime after database resets
- Matches expected behavior of a resilient monitoring agent
- Simple implementation - call `ResetEnrollment()` then `enrollAgent()` on 404

**Cons:**
- Could mask security issues (revoked certificate looks the same as database reset)
- Agent creates new identity without user awareness
- Potential for enrollment loops if control plane is rejecting enrollment

**Implementation:**
1. When 404 is detected in heartbeat, return to daemon with `ErrAgentNotRegistered`
2. Daemon catches this error, calls `ResetEnrollment()` to wipe cert/key/CSR
3. Daemon calls `enrollAgent()` to re-enroll with new identity
4. Log clearly that re-enrollment is happening and why
5. Add retry backoff to prevent enrollment loops

**Code changes needed:**
- `daemon.go`: Replace error return with re-enrollment logic after detecting `ErrAgentNotRegistered`
- Add backoff timer to prevent enrollment loops (e.g., max 3 re-enrollment attempts per daemon run)

### Option B: Startup Validation Check

**Alternative:** Add a validation heartbeat on startup before entering the main loop.

**Pros:**
- Detects stale certificates immediately at startup
- Fails fast rather than after N seconds of heartbeat interval
- Could still auto-re-enroll or provide clear user guidance

**Cons:**
- Doesn't help if certificate becomes invalid while daemon is running (revocation)
- Adds startup latency (extra HTTP round-trip)

**Implementation:**
1. After loading certificate, send a test heartbeat
2. If 404, trigger re-enrollment or exit with clear error
3. If success, proceed to main heartbeat loop

### Option C: Keep Current Behavior (Not Recommended)

Require manual `reset-enrollment` and daemon restart.

**Pros:**
- User is explicitly aware of the identity change
- No risk of enrollment loops

**Cons:**
- Poor user experience - requires manual intervention
- Doesn't align with "quiet, continuous assurance" goal
- Increases operational burden

## Recommended Implementation

**Combine Option A + Option B:**

1. **Startup validation:** Send initial heartbeat immediately (already implemented in `client.Run()` line 94)
2. **Auto-re-enrollment on 404:** When 404 is detected, automatically reset enrollment and re-register
3. **Safety limits:** Max 3 re-enrollment attempts per daemon run to prevent loops
4. **Clear logging:** Log when re-enrollment happens and why

**Why:** This provides self-healing behavior while preventing enrollment loops. The agent will recover from database resets automatically but won't loop forever if the control plane is rejecting enrollments.

**Coordination with Dallas:**
- Control plane should return 404 for unknown agent IDs (already implemented)
- Consider: Should revoked agents return 404 or 403? Currently both return 404, making them indistinguishable
- Recommend: Return 403 Forbidden for revoked agents, 404 Not Found for unknown agents
- This allows agent to differentiate "I should re-enroll" (404) from "I was explicitly revoked" (403)

**Files to modify:**
1. `agent/internal/agent/daemon.go` - Add re-enrollment logic after 404 detection
2. Add tests for re-enrollment flow in `daemon_test.go`
3. Update documentation to explain auto-recovery behavior

**Why:** This decision improves operational resilience and aligns with the project goal of "quiet, continuous assurance." The agent should recover from expected operational scenarios (database resets, control plane migrations) without manual intervention.
