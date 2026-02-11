# Ash Work History

## 2026-02-10: Team formation
- Assigned to crypto/mTLS engineering
- Responsibilities: Internal CA, certificate lifecycle, mTLS handshake, enrollment protocol, security review
- Work order: Collaborate with Kane (agent mTLS client) and Dallas (control plane CA + mTLS server)

## Learnings

### CA Implementation
- Implemented CaService.cs using System.Security.Cryptography for RSA key generation and X.509 certificate operations
- CA auto-initializes on first control plane startup, generates 2048-bit RSA key and self-signed root certificate
- Certificate signing uses X509SignatureGenerator.CreateForRSA with proper issuer DN and serial number
- CA files stored in control-plane/data/ca/ with proper permissions (key: 600, cert: 644)

### mTLS Server Configuration
- ASP.NET Core Kestrel configured for TLS 1.2+ with ClientCertificateMode.AllowCertificate
- MtlsValidationService performs X.509 chain validation against internal CA using CustomRootTrust
- MtlsAuthenticationMiddleware handles certificate validation for protected endpoints
- Public endpoints (/agents/register, /ca/certificate, /swagger) bypass certificate requirements
- Agent ID extracted from certificate CN field and stored in HttpContext.Items for controller access

### Certificate Lifecycle
- Enrollment: Agent submits CSR → Control plane signs → Returns certificate PEM
- Validation: Chain verification + expiry check + revocation check (database-backed, no CRL in v0)
- Revocation: Database flag (Agent.RevokedAt), checked during heartbeat validation
- Agent certificates valid for 365 days, CA valid for 10 years

### Key File Paths
- control-plane/Services/CaService.cs - CA operations (generate, sign, load)
- control-plane/Services/MtlsValidationService.cs - Certificate validation logic
- control-plane/Middleware/MtlsAuthenticationMiddleware.cs - Request authentication
- control-plane/Controllers/HeartbeatController.cs - mTLS-protected heartbeat endpoint
- control-plane/Controllers/CaController.cs - CA certificate distribution
- control-plane/Program.cs - Kestrel mTLS configuration
- scripts/ca-tool.sh - CLI tool for CA management (init, info, verify, export, secure)
- control-plane/docs/MTLS_SECURITY.md - Security architecture documentation

### Security Decisions
- RSA 2048-bit chosen for compatibility (Go agent uses RSA 2048)
- SHA-256 hash algorithm for all certificate signatures
- TLS 1.2 minimum, TLS 1.3 preferred
- No certificate revocation lists (CRLs) in v0, database-backed revocation only
- Client certificate validation uses CustomRootTrust (no system trust store)
- Agent private keys never leave the agent system (CSR model)

### ASP.NET Core Patterns
- Singleton CaService instantiated once, loads CA key into memory
- Scoped MtlsValidationService injected into middleware via DI
- Middleware runs before controllers, validates certificate and extracts agent ID
- Controllers access agent ID via HttpContext.Items["AgentId"]
- Kestrel ConfigureHttpsDefaults used for TLS settings (not app.UseHttps)
