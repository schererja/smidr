# Team Decisions

This file records scope, architecture, and process decisions made by the team.

Agents write decisions to `.ai-team/decisions/inbox/` and Scribe merges them here.

---

## Initial Decisions

### 2026-02-10: Project scope and tech stack
**By:** Squad (Coordinator)
**What:** Established project scope, tech stack, and work order for Smidr v0
**Why:** Team formation — agents need day-1 context

**Scope:**
- v0 system: quiet, continuous assurance for Linux x86_64 systems
- Agent (Go) → Control Plane (C#) → UI (React)
- mTLS authentication via internal CA
- Baseline learning + anomaly detection
- PostgreSQL storage

**Tech Stack:**
- Agent: Go, systemd service
- Control Plane: C# REST API (future: gRPC/WebSockets)
- Frontend: React with user registration
- Storage: PostgreSQL
- Auth: mTLS (internal CA)
- Build: GoReleaser (agent), GitHub Actions (CI/CD)

**Work Order:**
1. Kane (Agent Dev) writes first
2. Dallas (Control Plane) integrates
3. Back-and-forth as needed between agent ↔ control plane
4. Ash handles mTLS/crypto
5. Unit tests throughout (Parker)
6. Brett documents everything
