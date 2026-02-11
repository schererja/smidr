# Kane Work History

## 2026-02-10: Team formation
- Assigned to agent development (Go)
- Responsibilities: signal collection, systemd service, mTLS client, enrollment flow, heartbeat loop
- Work order: Write agent first, Dallas integrates control plane after

## Learnings

### Architecture
- Signal collection abstracted into OS-specific implementations using Go build tags
- `signals.go` contains shared interface and structs
- `signals_linux.go` uses `/proc` filesystem (tagged `//go:build linux`)
- `signals_darwin.go` uses `sysctl` and `vm_stat` commands (tagged `//go:build darwin`)
- `syscall.Statfs` works portably across Linux and macOS for disk usage
- Keeps same API contract (`Snapshot` struct) regardless of OS
- **mTLS client configuration:** Agent loads client cert+key via `tls.LoadX509KeyPair()`, sets `Certificates` field in `tls.Config`, and optionally loads CA cert for server verification
- **Localhost development:** When connecting to localhost without CA cert, `InsecureSkipVerify` is enabled automatically

### File Locations
- `agent/internal/signals/signals.go` — shared signal collection interface
- `agent/internal/signals/signals_linux.go` — Linux-specific collectors using /proc
- `agent/internal/signals/signals_darwin.go` — macOS-specific collectors using sysctl/vm_stat
- `agent/cmd/agent/main.go` — agent entry point
- `agent/internal/heartbeat/client.go` — heartbeat client with mTLS support
- `agent/internal/agent/config.go` — configuration loading and defaults
- `agent/internal/agent/daemon.go` — enrollment and daemon logic
- `control-plane/Program.cs` — Kestrel server configuration with mTLS endpoint setup

### Patterns
- Go build tags enable OS-specific implementations without runtime checks
- Same function signatures across OS implementations allow transparent switching
- Error aggregation in `Collect()` reports all failures, not just first
- **Default config for localhost development:** Use `https://localhost:5001` instead of example domain
- **Config field naming:** Use snake_case in YAML (e.g., `ca_cert_path`), CamelCase in Go structs (e.g., `CACertPath`)
- **mTLS troubleshooting:** Client cert won't be sent unless server requests it during TLS handshake (requires `ClientCertificateMode.RequireCertificate` on Kestrel)
- **Kestrel mTLS configuration:** HTTPS options must be set on the specific endpoint via `UseHttps(httpsOptions => {...})`, not globally via `ConfigureHttpsDefaults()`, otherwise the dev certificate's defaults take precedence
- **Kestrel certificate validation:** Use `ClientCertificateValidation = (cert, chain, errors) => true` to accept any client certificate during TLS handshake; actual validation happens in middleware layer using internal CA

### macOS Signal Collection Methods
- **Uptime:** Parse `sysctl kern.boottime` output, calculate seconds since boot
- **Load Average:** Parse `sysctl vm.loadavg` output (format: `{ 1.23 4.56 7.89 }`)
- **Memory:** Parse `vm_stat` output for active/wired/compressed pages, get total from `sysctl hw.memsize`
- **Disk Usage:** Use `syscall.Statfs` (portable across Linux/macOS)
- **Process Count:** Count lines from `ps -A` output (excluding header)

### ASP.NET Core MVC
- **Controller Forbid() method:** `Forbid()` requires an authentication scheme to be registered in `Program.cs` via `AddAuthentication()`
- **Custom mTLS middleware pattern:** When using custom middleware for authentication instead of ASP.NET's built-in authentication, return `StatusCode(403)` directly from controllers instead of `Forbid()`
- **File locations:** `control-plane/Controllers/HeartbeatController.cs` handles heartbeat endpoint, checks agent ID from cert vs payload

### mTLS Certificate Lifecycle
- **Agent ID in certificate:** Control plane validates that agent ID in certificate CN matches agent ID in heartbeat payload (HeartbeatController.cs line 34)
- **CSR regeneration:** `EnsureKeyAndCSR()` only regenerates CSR if key was just generated OR CSR doesn't exist; changing agent_id in config requires deleting CSR+cert to trigger regeneration
- **Certificate issuance:** `CaService.IssueAgentCertificate()` preserves the subject from the CSR; agent ID must be correct in CSR for certificate to have correct CN
- **Agent enrollment flow:** Agent generates key+CSR → POSTs to `/api/agents/register` with agentId+csrPem → Control plane issues cert → Agent saves cert → Agent uses cert for mTLS
- **ClientCertificateMode.RequireCertificate vs AllowCertificate:** `RequireCertificate` rejects connections without client cert at TLS layer (before middleware), preventing registration endpoint from working; use `AllowCertificate` to allow public endpoints
- **Registration endpoint must be public:** `/api/agents/register` must be in middleware's public endpoints list AND server must use `ClientCertificateMode.AllowCertificate` so agents without certs can enroll

### Certificate Validation Architecture
- **Two-layer validation:** TLS layer accepts any client cert (`ClientCertificateValidation = (cert, chain, errors) => true`), then middleware validates against internal CA
- **Middleware extracts agent ID:** `MtlsValidationService.ExtractAgentId()` parses certificate subject's CN field
- **Agent ID validation flow:** Middleware extracts CN → stores in `HttpContext.Items["AgentId"]` → HeartbeatController compares to request payload's agentId
