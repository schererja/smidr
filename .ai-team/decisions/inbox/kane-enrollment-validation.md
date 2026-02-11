### 2026-02-11: Agent enrollment validation and automatic stale certificate detection

**By:** Kane  
**What:** Agent now validates certificate on first heartbeat and detects stale certificates (404 errors)  
**Why:** Prevent silent failures when control plane database is reset and agent has old certificate

## Problem

After a control plane database reset, agents with existing certificates believe they're enrolled but get 404 errors on heartbeat because their agent ID no longer exists in the database. The daemon would continuously retry heartbeats without detecting the root cause.

## Solution

1. **Heartbeat validation:** First heartbeat attempt now validates that the certificate is recognized by the control plane
2. **404 detection:** When control plane returns 404, the heartbeat client wraps it in `ErrAgentNotRegistered` error
3. **Clear error messages:** Daemon catches this error and provides clear guidance to run `reset-enrollment` command
4. **Fail fast:** Daemon exits immediately on 404 instead of retrying indefinitely

## Implementation

**heartbeat/client.go:**
- Added `ErrAgentNotRegistered` exported error variable
- Modified `sendOnce()` to detect 404 status and wrap error
- Modified `Run()` to fail fast on `ErrAgentNotRegistered` instead of logging and continuing

**agent/daemon.go:**
- Added error check after `client.Run()` to detect `ErrAgentNotRegistered`
- Logs clear error messages with remediation steps
- Returns descriptive error instead of raw 404

## User Experience

**Before:**
```
INFO agent already enrolled cert_path=/path/to/cert.pem
INFO heartbeat loop started
WARN heartbeat failed error="heartbeat failed with status 404: Agent abc-123 not registered"
WARN heartbeat failed error="heartbeat failed with status 404: Agent abc-123 not registered"
... (continues forever)
```

**After:**
```
INFO certificate found cert_path=/path/to/cert.pem
INFO heartbeat loop started
ERROR agent certificate is invalid - this usually happens when the control plane database was reset
ERROR to fix this issue, run: smidr-agent reset-enrollment
ERROR then restart the daemon to re-enroll with a new certificate
(daemon exits with clear error)
```

## Testing

User should:
1. Stop agent daemon
2. Run `smidr-agent reset-enrollment` to clear stale state
3. Restart daemon - will generate new ID and re-enroll
4. Verify heartbeats succeed

## Files Changed

- `agent/internal/heartbeat/client.go` - Added error detection and validation
- `agent/internal/agent/daemon.go` - Added error handling and user guidance
- `TESTING-RESET.md` - Updated documentation with new behavior
