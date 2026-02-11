# Smidr - Secure, Managed Infrastructure Delivery & Response – v0 Specification

## 1. Purpose

The v0 system provides **quiet, continuous assurance** for Linux x86_64 systems by answering one core question:

> **“Is this system behaving normally right now?”**

The platform:

- Runs a lightweight agent on Linux systems
- Collects minimal system signals
- Learns baseline behavior centrally
- Detects meaningful deviations
- Surfaces issues with context, not noise

v0 prioritizes:

- correctness over completeness
- clarity over flexibility
- shipping over generality

---

## Architecture

```
┌─────────────────┐         ┌─────────────────┐         ┌──────────┐
│  Linux Systems  │         │ Control Plane   │         │    UI    │
│   (Go Agent)    │─HTTPS──▶│   (C# API)      │◀─HTTPS──│  React   │
│                 │  mTLS   │                 │         │          │
└─────────────────┘         └────────┬────────┘         └──────────┘
                                     │
                                ┌────▼─────┐
                                │PostgreSQL│
                                └──────────┘
```

**Components:**
- **Agent**: Lightweight Go daemon collecting system signals
- **Control Plane**: ASP.NET Core API managing enrollment, learning, and evaluation
- **CA Service**: Internal certificate authority for mTLS
- **UI**: React web app for viewing system health
- **Database**: PostgreSQL (production) or SQLite (development)

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for complete architecture details.

---

## Tech Stack

- **Agent:** Go 1.21+, systemd integration
- **Control Plane:** C# / .NET 8, ASP.NET Core, Entity Framework Core
- **UI:** React 19, TypeScript, Vite, Tailwind CSS
- **Database:** PostgreSQL (production), SQLite (development)
- **Auth:** Mutual TLS (mTLS) with internal CA
- **Build:** GoReleaser (agent), GitHub Actions (CI/CD)

---

## v0 Scope and Non-Goals

### In Scope (v0)
- ✅ Agent enrollment and heartbeat over mTLS
- ✅ Baseline learning and anomaly detection
- ✅ Health state transitions
- ✅ Minimal UI for viewing system status
- ✅ Single-tenant deployment

### Out of Scope (v0)
- ❌ Integrations (Slack, email, PagerDuty)
- ❌ Inbound agent connections or remote commands
- ❌ Windows or non-x86 architectures
- ❌ User-configurable alert rules
- ❌ Multi-tenant support
- ❌ Dashboards, graphs, or charts
- ❌ User authentication (use reverse proxy)

---

## Development

### Prerequisites
- Go 1.21+ (agent)
- .NET 8.0 SDK (control plane)
- Node.js 18+ and npm (UI)
- PostgreSQL 14+ (production) or SQLite (development)

### Local Development

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for complete setup instructions.

**Quick commands:**

```bash
# Build agent
cd agent && go build -o bin/smidr-agent ./cmd/agent

# Run control plane
cd control-plane && dotnet run

# Run UI dev server
cd ui && npm install && npm run dev

# Run tests
cd agent && go test ./...
cd control-plane && dotnet test
```

---

## API Examples

### Enroll Agent

```bash
curl -X POST https://localhost:5001/agents/register \
  -H 'Content-Type: application/json' \
  -d '{
    "agentId": "550e8400-e29b-41d4-a716-446655440000",
    "hostname": "web-server-01",
    "csrPem": "-----BEGIN CERTIFICATE REQUEST-----\n..."
  }'
```

### Send Heartbeat (requires mTLS)

```bash
curl -X POST https://localhost:5001/v0/agents/heartbeat \
  --cert agent.crt.pem --key agent.key.pem \
  -H 'Content-Type: application/json' \
  -d '{
    "agentId": "550e8400-e29b-41d4-a716-446655440000",
    "timestamp": "2024-02-10T15:04:05Z",
    "uptimeSeconds": 86400.5,
    "loadAverage1m": 1.23,
    "memoryUsedPct": 45.6,
    "diskUsedPct": 67.8,
    "processCount": 156
  }'
```

See [docs/API.md](docs/API.md) for complete API reference.

---

## Security

- **mTLS Authentication:** All agent communication uses mutual TLS
- **Internal CA:** Self-signed CA for agent certificate issuance
- **Agent Isolation:** Runs as unprivileged user with systemd hardening
- **Key Management:** Private keys never leave their host systems
- **Certificate Rotation:** Agents can re-enroll to renew certificates

See [control-plane/README-CRYPTO.md](control-plane/README-CRYPTO.md) for cryptographic implementation details.

---

## Design Principles

1. **Agents report facts, not opinions** - All decision-making happens in the control plane
2. **Silence is a signal** - Missing heartbeats indicate problems
3. **Defaults must be safe and boring** - No surprises, no magic
4. **Behavior must be explainable** - Every health state change has a reason
5. **One tenant now, many later** - Architecture supports future multi-tenancy

---

## Roadmap

**v0 (Current):**
- ✅ Core agent, control plane, and UI
- ✅ mTLS enrollment and authentication
- ✅ Baseline learning and anomaly detection
- ✅ SQLite and PostgreSQL support

**v1 (Future):**
- Multi-tenant support with per-tenant CAs
- User authentication for UI
- Integrations (Slack, email, webhooks)
- Certificate rotation automation
- gRPC/WebSocket transport
- Extended signal collection

---

## Contributing

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for development setup and contribution guidelines.

**Key areas for contribution:**
- Unit and integration tests (Parker)
- Extended signal collectors (Kane)
- UI enhancements (Lambert)
- Security hardening (Ash)
- Documentation improvements (Brett)

---

## License

(Add license information here)

---

## Success Criteria

Smidr v0 succeeds if:

- Agent installs in under 5 minutes
- Runs unattended for days without intervention
- Detects deviations without alert noise
- Health decisions are explainable to operators
- Users feel less need to manually check system status
