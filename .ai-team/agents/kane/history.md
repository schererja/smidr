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

### OS Detection
- Agent sends OS information using `runtime.GOOS` in both registration and heartbeat payloads
- Control plane stores OS in Agent model (nullable string for backward compatibility)
- Field flows agent → control plane → UI for OS-specific icon display
- Files: `agent/internal/agent/daemon.go`, `agent/internal/heartbeat/client.go`

📌 Team update (2026-02-11): OS field implementation complete across agent/control plane/UI — decided by Kane, Dallas, Lambert

📌 Team update (2026-02-11): Multi-drive disk metrics deferred to v1 (keep root-only for v0) — decided by Ripley

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

### Payload Structure
- **OS detection:** Use `runtime.GOOS` to report operating system ("linux", "darwin", "windows") in registration and heartbeat payloads
- **Registration payload:** Includes agentId, hostname, token, csrPem, and os fields
- **Heartbeat payload:** Includes agentId, timestamp, signals (uptime, load, memory, disk, processes), and os field
- **JSON field naming:** Go struct fields use PascalCase, JSON tags use camelCase (e.g., `AgentID string json:"agentId"`)

### Enrollment Reset
- **Reset command:** `smidr-agent reset-enrollment` clears all enrollment state and allows re-enrollment
- **What it deletes:** Certificate file (cert_path), CSR file (csr_path), private key file (key_path)
- **Config updates:** Clears agent_id, key_path, csr_path, cert_path fields in config file
- **Use case:** When control plane database is reset and agent has stale certificate for non-existent agent ID
- **Re-enrollment flow:** After reset, agent daemon will auto-generate new ID, new keypair, new CSR, and re-enroll
- **Files:** `agent/internal/agent/reset.go`, `agent/internal/agent/reset_test.go`, `agent/cmd/agent/main.go`
- **Safety:** Only deletes enrollment artifacts, preserves control_plane_url, hostname, token, and heartbeat_interval

### Certificate Validation on Startup
- **Problem:** After database reset, agent with existing cert believes it's enrolled but gets 404 errors
- **Solution:** First heartbeat validates certificate; daemon detects 404 and provides clear remediation steps
- **Error detection:** `ErrAgentNotRegistered` exported from heartbeat package, wrapped around 404 responses
- **Fail fast:** Daemon exits immediately on certificate validation failure instead of retrying indefinitely
- **User guidance:** Clear error messages explain the issue and direct user to run `reset-enrollment` command
- **Files:** `agent/internal/heartbeat/client.go` (line 84, 96, 109, 157), `agent/internal/agent/daemon.go` (line 64-69)
- **Coordination required:** Dallas verifies 404 responses, Ash verifies certificate validation
- **Testing:** See `.ai-team/agents/kane/TEAM-VERIFICATION.md` for verification checklist

### Certificate Validation and Error Handling (2026-02-11)
- **Enrollment check:** Agent only checks if certificate file exists, does NOT validate with control plane at startup (`daemon.go` lines 26-38)
- **First validation:** Happens when first heartbeat is sent, control plane may return 404 if agent not registered
- **404 error handling:** Heartbeat client returns `ErrAgentNotRegistered`, daemon logs error and exits with guidance to run reset-enrollment
- **No auto-recovery:** Current implementation requires manual intervention (reset-enrollment + daemon restart)
- **Agent ID storage:** Persisted in config file (not extracted from certificate); certificate CN contains copy of agent ID
- **Recommendation:** Add automatic re-enrollment on 404 for self-healing behavior (see kane-agent-404-analysis.md decision)
- **Control plane coordination needed:** Distinguish 404 (unknown agent, should re-enroll) from 403 (revoked agent, should not re-enroll)

## 2026-02-11: Merged Decisions from Team Debug Session

**Merged from inbox decisions:** ash-certificate-investigation-summary.md, ash-orphaned-cert-analysis.md, kane-agent-404-analysis.md, kane-agent-os-field.md, kane-enrollment-validation.md, kane-reset-enrollment-command.md, and related Dallas/Lambert decisions

**Key consolidated decisions:**

### Agent Enrollment and Certificate Validation (Consolidated)
- Comprehensive flow covering orphaned certs, 404 handling, and reset command
- Authors: Kane, Dallas
- Validates certificates on first heartbeat
- Detects stale certificates and provides clear recovery instructions

### OS Field Implementation (Complete)
- Full end-to-end: Agent sends `runtime.GOOS` → Control Plane stores in DB → UI displays with icons
- Authors: Kane, Dallas, Lambert
- Agent implementation: Added OS field to both registration and heartbeat payloads
- Backward compatible: field is optional

### Multi-drive Disk Metrics Decision
- Deferred to v1; keep root-only for v0
- Author: Ripley
- Comprehensive analysis with fallback plan if needed later

**Coordination outcomes:**
- Dallas verified database schema and migration handling
- Lambert verified UI is prepared for OS icons when backend provides data
- Ash reviewed certificate validation architecture and 404 semantics
- Registration debugging identified silent retry loops (now fixed by Dallas)

### Automatic Certificate Re-enrollment (2026-02-10)
- **Problem:** Agent with stale certificate (from wiped control plane database) required manual reset-enrollment command and daemon restart
- **Solution:** Implemented automatic re-enrollment when agent detects 404 from control plane
- **Flow:** Agent detects 404 → deletes stale cert files → generates new agent ID + key + CSR → re-enrolls → restarts heartbeat loop
- **Safety:** Max 3 re-enrollment attempts per daemon run to prevent infinite loops
- **Config persistence:** New agent ID is persisted to config file after successful re-enrollment
- **User experience:** Agent recovers automatically from database resets without manual intervention
- **Files modified:** `agent/internal/agent/daemon.go` (lines 64-140), `agent/internal/agent/reset.go` (added DeleteEnrollmentFiles function), `agent/cmd/agent/main.go` (added configPath parameter)
- **Key insight:** Heartbeat config must be rebuilt with new agent ID after re-enrollment, not reused from initial setup
- **Testing:** Verified agent automatically re-enrolls when deleted from database, generates new ID, and continues heartbeat loop successfully

### Build System (2026-02-11)
- **Makefile:** Created comprehensive Makefile with 20+ targets for building, testing, linting, and development
- **Cross-compilation:** Support for Linux (amd64), macOS (amd64, arm64) via `build-linux`, `build-darwin`, `build-all` targets
- **Testing:** Multiple test targets: `test` (with race detector), `test-verbose`, `test-coverage`, `test-coverage-html`
- **Code quality:** Targets for `fmt`, `fmt-check`, `vet`, `lint` (requires golangci-lint)
- **Development workflows:** `dev` target (clean+deps+fmt+vet+test+build), `ci` target (deps+fmt-check+vet+test-coverage)
- **Version info:** Build embeds version, commit hash, and build date via ldflags (extractable from git)
- **Binary naming:** Default binary is OS-specific (`smidr-agent`), cross-compiled binaries include OS/arch suffix (`smidr-agent-linux-amd64`)
- **Helper targets:** `deps` (download/tidy), `clean` (remove artifacts), `install` (install to /usr/local/bin), `run-init`, `run-daemon`
- **Help system:** `make help` shows all targets with descriptions
- **File:** `agent/Makefile`

📌 Team update (2026-02-11): Makefile build system for agent — comprehensive build tooling with 20+ targets, cross-platform support — decided by Kane

📌 Team update (2026-02-11): Git tracking exclusions — .ai-team/ and diagnostic files excluded from git per user directive — decided by Jason Scherer
