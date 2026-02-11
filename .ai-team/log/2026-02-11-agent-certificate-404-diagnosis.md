# Session Log: Agent Certificate 404 Diagnosis

**Date:** 2026-02-11  
**Requested by:** schererja  
**Session ID:** agent-certificate-404-diagnosis

## Participants

- **Kane** (Agent Dev)
- **Dallas** (Control Plane Dev)

## Summary

Investigated agent startup flow, certificate validation, and 404 error handling after control plane database reset caused "orphaned certificates."

### Kane's Analysis

**Focus:** Agent startup, validation flow, error handling, auto-recovery options

**Key Findings:**
- Agent checks only for certificate file existence, not validity with control plane
- No startup validation of certificate recognition
- 404 errors correctly detected in heartbeat with clear user guidance
- Agent ID flows from config → CSR CN → certificate CN → heartbeat payload

**Recommendations:**
- Implement automatic re-enrollment on 404 detection (self-healing)
- Add startup validation heartbeat for fail-fast behavior
- Safety limits: max 3 re-enrollment attempts to prevent loops
- Differentiate 404 (unknown agent) from 403 (revoked agent) for proper recovery

### Dallas's Analysis

**Focus:** Heartbeat endpoint behavior, mTLS validation stages, certificate vs registration checks

**Key Findings:**
- Two-stage auth: cryptographic validation (middleware) → registration check (controller)
- Orphaned certs pass stage 1 but fail stage 2 (agent record doesn't exist)
- 404 is semantically correct per RFC 9110: resource doesn't exist
- Agent already has excellent 404 handling with clear recovery instructions

**Recommendations:**
- Keep current 404 response (semantically correct, well-handled by agent)
- Optional: Add `X-Smidr-Error` header for precise error signaling
- Document expected behavior for database reset scenarios
- Production hardening: CRL for pre-reset certs, separate audit table for cert serials

## Decisions Made

1. **HTTP Status Code:** Confirmed 404 is correct for orphaned certificates
2. **Agent Error Handling:** Current implementation (detect 404, log guidance, exit) is appropriate
3. **Documentation:** Add explanations to both agent and control plane READMEs
4. **Future Enhancement:** Consider auto-re-enrollment for self-healing behavior (requires safety limits)

## Files Referenced

**Agent:**
- `agent/internal/agent/daemon.go` - Startup logic, enrollment check, error handling
- `agent/internal/agent/certificates.go` - CSR generation with agent ID in CN
- `agent/internal/heartbeat/client.go` - 404 detection, ErrAgentNotRegistered

**Control Plane:**
- `control-plane/Controllers/HeartbeatController.cs` - Registration check, 404 response
- `control-plane/Middleware/MtlsAuthenticationMiddleware.cs` - Cryptographic validation
- `control-plane/Services/MtlsValidationService.cs` - Agent ID extraction from certificate CN

## Outcome

Root cause confirmed: orphaned certificates after database reset. Current implementation is correct and well-designed. No urgent changes required. Documentation improvements and optional auto-recovery enhancement identified for future work.
