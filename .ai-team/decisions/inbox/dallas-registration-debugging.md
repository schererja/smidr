### 2026-02-11: Registration failure root cause and fix

**By:** Dallas, Kane  
**What:** Identified that agent registration failures are silently ignored due to infinite retry loop with only warning-level logging  
**Why:** Agent logs show "heartbeat failed with status 404: Agent not registered" but no preceding registration errors are visible. Investigation revealed registration errors are logged as warnings and retried indefinitely, masking the actual failure reason.

**Root Cause:**
- Agent's `enrollAgent()` function (daemon.go:75-93) retries registration indefinitely
- Registration errors from `registerOnce()` are logged as `Warn` but not returned
- No distinction between retryable (5xx) and fatal (4xx) errors
- Agent eventually continues to heartbeat loop even if registration never succeeded
- Heartbeat then fails with 404 because agent was never written to database

**Control Plane Status:**
- Registration endpoint `/api/agents/register` is correctly implemented
- Accepts: agentId, hostname, csrPem, token (optional), os (optional)
- Returns: signed certificate PEM
- Database schema includes all required columns (including OS column from migration 20260211000000)
- Program.cs correctly calls `Migrate()` on startup to apply pending migrations

**Agent Fix Required (Kane):**
1. Distinguish between fatal (4xx) and retryable (5xx) HTTP errors
2. Exit enrollment loop on 4xx errors instead of retrying
3. Add maximum retry limit (e.g., 10 attempts)
4. Log actual registration errors at ERROR level, not WARN
5. Improve error messages to show HTTP status codes

**Recommendation:**
```go
// In registerOnce():
if resp.StatusCode >= 400 && resp.StatusCode < 500 {
    return fmt.Errorf("registration failed permanently (status %d): %s", resp.StatusCode, snippet)
}

// In enrollAgent():
if strings.Contains(err.Error(), "permanently") {
    log.Error("registration failed permanently", "error", err)
    return err  // Don't retry
}
```

**Testing Required:**
1. Run control-plane/test-registration.sh to verify endpoint works
2. Check control-plane/check-database.sh to verify schema is correct
3. Test agent registration with control plane running
4. Verify registration errors are now visible in agent logs
5. Verify 4xx errors stop retry loop immediately
