# ⚒️ Smidr

Smidr is a CI/Build orchestration system designed for reliability in enterprise environments and clarity in operations. It separates a C# control plane from Go executors, with versioned contracts between them and an optional TypeScript/React Web UI.

- Control Plane (C#): schedules jobs, owns state machine, enforces policy, exposes REST/gRPC
- Executors (Go): pull-based, outbound-only connectivity; run jobs natively or in Docker; stream logs and status
- Contracts (protobuf): versioned Job, Executor Capabilities, and Events/Logs; executors declare supported versions
- Web UI (TypeScript/React): job submission, status dashboards, and administration

See the architecture overview in [docs/architecture/overview.md](docs/architecture/overview.md) and the diagram at [docs/image/Architecture.png](docs/image/Architecture.png).

NOTE: This project is a work in progress. It is being worked on actively and is not yet production-ready. This is a side project by Jason Scherer to explore reliable CI/build orchestration with clear separation of concerns and to learn new technologies.

## Why Smidr

- Reduce environment fragility and waste with isolation and shared caching
- Improve reliability with a pull-based executor model and clear lifecycle
- Gain observability via structured events/logs and artifact metadata
- Scale across heterogeneous hosts with capability-based scheduling

## Architecture

Core components and data flows are documented across:

- [Overview](docs/architecture/overview.md)
- [Control Plane API Design](docs/architecture/api-design.md)
- [Executor/Daemon Design](docs/architecture/daemon-design.md)
- [Providers](docs/architecture/providers.md)
- [Runtime](docs/architecture/Runtime.md)
- [Task Lifecycle](docs/architecture/TaskLifeCycle.md)

Key behaviors

- Pull-based executors (firewall-friendly, NAT-safe)
- Capability negotiation and versioned contracts
- Isolation modes: Docker when available, native as fallback
- Status/log streaming and artifact publishing

## Repository Layout (target)

The target layout for v0.1.x is:

```bash
Smidr/
├── api/                     # Control Plane (C#)
│   ├── Smidr.API.sln
│   ├── src/
│   │   ├── Controllers/     # REST / gRPC controllers
│   │   ├── Services/        # Scheduling, capability matching
│   │   ├── Models/          # DTOs for Job, Executor, Events
│   │   ├── Repositories/    # Persistence abstraction
│   │   └── Program.cs
│   ├── tests/
│   │   └── UnitTests/
│   └── README.md
│
├── executor/                # Executor / Daemon (Go)
│   ├── cmd/
│   │   └── smidrexec/       # main entry point
│   ├── pkg/
│   │   ├── executor/        # Executor loop, job runner
│   │   ├── capabilities/    # Detection of runtime + Docker
│   │   ├── docker/          # Docker utilities / isolation
│   │   ├── jobs/            # Job pulling & execution
│   │   ├── protocols/       # gRPC / HTTP client for control plane
│   │   └── logging/         # Log streaming
│   └── README.md
│
├── contracts/               # Shared contract definitions
│   ├── job.proto            # Job definition schema
│   ├── executor.proto       # Executor capability schema
│   ├── events.proto         # Job state & log streaming
│   └── README.md
│
├── web/                     # Web UI (TypeScript / React)
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── services/        # API clients
│   │   └── App.tsx
│   ├── package.json
│   └── README.md
│
├── docs/                    # Architecture & developer docs
│   ├── architecture/
│   └── image/
│
└── README.md
```

Note: Legacy implementations live under [archive](archive). They are not part of the current architecture.

## Security & Identity

- Executors authenticate to the control plane and report identity and capabilities on registration
- Mutual TLS or token-based auth; outbound-only executor connectivity
- Least-privilege isolation, preferring rootless Docker where possible

## Roadmap

v0.1.0 (MVP)

- Control Plane API: submit/list/cancel jobs; lifecycle enforcement; capability-based scheduling
- Go Executor: pull loop, Docker-first execution with native fallback; logs/events streaming
- Versioned Contracts: Job/Executor/Event schemas with version negotiation
- Persistence: minimal DB + object storage for artifacts
- Basic UI flows: submit job and view status

v0.2.0 (Yocto-focused)

- Yocto/BitBake-aware job types and artifact handling
- Improved log streaming and search; layer cache helpers
- Retry/backoff policies and clearer failure semantics

v0.3.0 (Reliability & multi-tenant)

- Multi-tenant queueing and quotas; metrics/monitoring
- Cancellation, timeouts, and resumable uploads
- Advanced caching strategies (sstate, downloads)

## Contributing

Smidr is under active development. Contributions are welcome in architecture, executors, control plane, and UI. Please start by reviewing [docs/architecture/overview.md](docs/architecture/overview.md) and related design pages.

## License

Smidr is licensed under the [MIT License](LICENSE).

## Links

- Repository: <https://github.com/schererja/smidr>
- Issues: <https://github.com/schererja/smidr/issues>
