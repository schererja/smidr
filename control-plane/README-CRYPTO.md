# Cryptography & mTLS Implementation

This document describes the crypto components in the Smidr control plane.

## Components

### CaService (Services/CaService.cs)
- Auto-generates internal CA on first startup
- Signs agent CSRs using X.509 certificate operations
- Loads existing CA from disk if present
- Exports CA certificate for distribution

### MtlsValidationService (Services/MtlsValidationService.cs)
- Validates client certificates against internal CA
- Builds X.509 certificate chains with CustomRootTrust
- Checks certificate expiry
- Extracts agent ID from certificate CN

### MtlsAuthenticationMiddleware (Middleware/MtlsAuthenticationMiddleware.cs)
- Intercepts requests and validates client certificates
- Allows public endpoints (enrollment, CA cert, swagger)
- Requires client cert for protected endpoints (/v0/*)
- Stores agent ID in HttpContext for controllers

### HeartbeatController (Controllers/HeartbeatController.cs)
- Protected endpoint requiring mTLS authentication
- Validates agent ID matches certificate CN
- Checks for revocation status
- Processes agent telemetry

### CaController (Controllers/CaController.cs)
- Public endpoint to download CA certificate
- Agents can fetch CA cert for verification

## Usage

### Start Control Plane
```bash
cd control-plane
dotnet run
```

CA auto-initializes at `data/ca/` on first run.

### Manage CA
```bash
# View CA info
../scripts/ca-tool.sh info

# Export CA certificate
../scripts/ca-tool.sh export > ca-cert.pem

# Verify agent certificate
../scripts/ca-tool.sh verify /path/to/agent.crt.pem
```

### Agent Enrollment
```bash
# Agent submits CSR
curl -X POST https://localhost:5001/agents/register \
  -H 'Content-Type: application/json' \
  -d '{"agentId":"UUID","hostname":"host","csrPem":"PEM"}'

# Returns signed certificate
{"certPem":"-----BEGIN CERTIFICATE-----\n..."}
```

### mTLS Heartbeat
```bash
# Requires client certificate
curl -X POST https://localhost:5001/v0/agents/heartbeat \
  --cert agent.crt.pem \
  --key agent.key.pem \
  -H 'Content-Type: application/json' \
  -d '{"agentId":"UUID","timestamp":"...","signals":...}'
```

## Security

- **CA Key**: 2048-bit RSA, 600 permissions, never transmitted
- **Agent Certs**: 365 day validity, signed by CA
- **TLS**: 1.2+ only, system cipher suites
- **Revocation**: Database-backed (no CRL in v0)

See `docs/MTLS_SECURITY.md` for full security architecture.
