# Executor/Daemon Design (v0.1.0)

## Role

Pull-based worker written in Go. Runs jobs, enforces isolation, streams logs/events, and reports state to the control plane. Operates with outbound-only connectivity (friendly to firewalls/NAT).

## Startup Flow

1) Load config (labels, resources, isolation preferences, credentials)
2) Detect capabilities (Docker availability, toolchains, OS/arch, resource limits)
3) Register capabilities + supported contract versions with control plane
4) Enter pull loop (with backoff when idle)

## Pull Loop (simplified)

```go
for {
 job := pullJob(capabilities)
 if job == nil { sleep(backoff); continue }
 runJob(job)
 reportCompletion(job)
}
```

## Subsystems

- Capability detector: collects host facts and runtime features
- Provider layer: pluggable execution backends (Docker, native process, future runtimes)
- Runtime layer: applies isolation, resources, lifecycle, cleanup
- Job runner: drives task lifecycle, invokes providers, captures signals
- Observability: streams logs/events/metrics; persists minimal local state for recovery
- Health/heartbeat: periodic liveness to control plane; recovers orphaned work on restart

## Isolation Modes

- Docker-first when available; native process fallback
- Jobs declare requirements; executor rejects when unsupported
- Resource limits applied via containers or OS primitives where available

## Failure & Recovery

- Idempotent reporting; duplicate sends tolerated
- On restart, reconcile any in-flight executions and report ORPHANED or RUNNING as appropriate
- Backoff on repeated pull/report failures; do not thrash the control plane

## Security

- Outbound TLS to control plane; executor identity and tokens kept local
- Minimal privileges; prefer rootless Docker when possible

## Mock Executors

- Optional demo mode that simulates jobs and logs; useful for UI and API exercises without real builds
