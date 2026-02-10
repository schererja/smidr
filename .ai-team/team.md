# Smidr Squad — Team Roster

## Project Context

**Owner:** Jason Scherer (schereja@gmail.com)

**Project:** Smidr — Secure, Managed Infrastructure Delivery & Response

**Description:** A quiet, continuous assurance platform for Linux x86_64 systems. Answers the question: "Is this system behaving normally right now?" Features:
- Lightweight Go agent running as systemd service
- C# control plane with REST API (future: gRPC/WebSockets)
- React frontend with user registration
- PostgreSQL storage
- mTLS authentication via internal CA
- Baseline learning and anomaly detection
- Minimal system signals (uptime, load, memory, disk, process count)

**Tech Stack:**
- Agent: Go (Linux → Windows/Mac later)
- Control Plane: C# REST API
- Frontend: React
- Storage: PostgreSQL
- Auth: mTLS (Certificate Authority, enrollment flow)
- Build: GoReleaser (agent), GitHub Actions (CI/CD)

**Work Order:**
1. Agent dev writes first (Go)
2. Control plane dev integrates (C#)
3. Back-and-forth as needed
4. Crypto/mTLS specialist handles enrollment
5. Unit tests throughout
6. Documentation for all components

---

## Active Team

| Name | Role | Charter | Status |
|------|------|---------|--------|
| Ripley | Lead | .ai-team/agents/ripley/charter.md | ✅ Active |
| Kane | Agent Dev | .ai-team/agents/kane/charter.md | ✅ Active |
| Dallas | Control Plane Dev | .ai-team/agents/dallas/charter.md | ✅ Active |
| Ash | Crypto/mTLS Engineer | .ai-team/agents/ash/charter.md | ✅ Active |
| Lambert | Frontend Dev | .ai-team/agents/lambert/charter.md | ✅ Active |
| Parker | Tester | .ai-team/agents/parker/charter.md | ✅ Active |
| Brett | Technical Writer | .ai-team/agents/brett/charter.md | ✅ Active |
| Scribe | Memory Manager | .ai-team/agents/scribe/charter.md | ✅ Active |

---

## Team Formation

**Created:** 2026-02-10
**Assignment ID:** 2026-02-10T19:54:23-smidr
**Universe:** Alien
