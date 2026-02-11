# Agent Registration Issue - Root Cause Analysis and Fix

**Reporter:** Kane (Agent Developer)  
**Date:** 2026-02-11  
**Issue:** Agent not registering after control plane database reset

## Root Cause

The agent checks if a certificate file exists (`daemon.go:28`) and assumes it's enrolled if the file is present. However, after a control plane database reset, the certificate is **stale** — it references an agent ID that no longer exists in the database.

**Result:** Agent skips enrollment and goes straight to heartbeat loop, which fails with 404 errors indefinitely.

## The Fix

I've implemented **certificate validation on first use** to detect and handle stale certificates:

### Changes Made

**1. Heartbeat Validation (`agent/internal/heartbeat/client.go`)**
- Added `ErrAgentNotRegistered` exported error (line 84)
- Modified `sendOnce()` to detect 404 responses and wrap them in this error (line 157)
- Modified `Run()` to fail fast when receiving this error (lines 96, 109)

**2. Daemon Error Handling (`agent/internal/agent/daemon.go`)**
- Added error checking after `client.Run()` (line 64)
- Detects `ErrAgentNotRegistered` and logs clear remediation steps
- Exits with actionable error message instead of retrying indefinitely

**3. Documentation (`TESTING-RESET.md`)**
- Updated use case section to explain new behavior
- Documents error messages user will see

## User Experience

### Before (Broken)
```
INFO agent already enrolled cert_path=/path/to/cert.pem
INFO heartbeat loop started
WARN heartbeat failed error="heartbeat failed with status 404: Agent faf726fd-ccda-4cf6-971f-eafd039d65f3 not registered"
WARN heartbeat failed error="heartbeat failed with status 404: Agent faf726fd-ccda-4cf6-971f-eafd039d65f3 not registered"
... (continues forever)
```

### After (Fixed)
```
INFO certificate found cert_path=/path/to/cert.pem
INFO heartbeat loop started
ERROR agent certificate is invalid - this usually happens when the control plane database was reset
ERROR to fix this issue, run: smidr-agent reset-enrollment
ERROR then restart the daemon to re-enroll with a new certificate
(daemon exits with error code)
```

## Testing Required

Since I cannot execute bash commands in this environment, **please run these tests:**

### 1. Build the Agent
```bash
cd /Users/schererja/src/github.com/schererja/smidr/agent
go build -o ../bin/smidr-agent ./cmd/agent
```

**Expected:** Should compile without errors

### 2. Test Current State (Should Detect Stale Cert)
```bash
# Start daemon with existing stale certificate
../bin/smidr-agent daemon

# Expected output:
# INFO certificate found cert_path=...
# INFO heartbeat loop started  
# ERROR agent certificate is invalid - this usually happens when the control plane database was reset
# ERROR to fix this issue, run: smidr-agent reset-enrollment
# ERROR then restart the daemon to re-enroll with a new certificate
# (exits)
```

### 3. Reset and Re-enroll
```bash
# Clear stale state
../bin/smidr-agent reset-enrollment

# Should output:
# Enrollment state reset successfully
# Deleted certificate files and cleared agent ID from config
# Run 'smidr-agent daemon' to re-enroll

# Restart daemon - should enroll successfully
../bin/smidr-agent daemon

# Expected:
# INFO certificate not found, enrolling...
# INFO agent registered attempt=1
# INFO heartbeat loop started
# (starts sending successful heartbeats)
```

### 4. Verify Registration
```bash
# Check UI - new agent should appear with new ID
# Check logs - heartbeats should succeed (no 404 errors)
```

## Coordination Needed

### Dallas (Control Plane)
- Verify 404 response format matches what agent expects
- Confirm `/api/agents/register` endpoint is working
- Check HeartbeatController properly validates agent ID from certificate

### Ash (Security/mTLS)
- Verify certificate validation in MtlsValidationService
- Confirm certificate CN extraction works correctly
- Check that new enrollment flow issues valid certificates

## Files Changed

- ✅ `agent/internal/heartbeat/client.go` - Added stale cert detection
- ✅ `agent/internal/agent/daemon.go` - Added error handling and user guidance
- ✅ `TESTING-RESET.md` - Updated documentation
- ✅ `.ai-team/agents/kane/history.md` - Recorded learnings
- ✅ `.ai-team/decisions/inbox/kane-enrollment-validation.md` - Documented decision

## Next Steps

1. **Build and test** the updated agent (commands above)
2. **Verify** control plane behavior (Dallas)
3. **Confirm** mTLS validation (Ash)
4. **Report** test results back

If the agent still doesn't register after these changes, we'll need to investigate:
- Is the control plane reachable?
- Is `/api/agents/register` returning proper responses?
- Is the CA service issuing valid certificates?
- Are there any network/TLS issues?

---

**Status:** Code changes complete, awaiting build & test verification
