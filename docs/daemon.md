# Smidr Daemon (gRPC Server)

Smidr is the build/agent + control-plane component of the larger Yggdrasil project. This repository will evolve into Yggdrasil; other Norse‑named components and the business application layer (C# WebAPI + Next.js) are planned to live alongside Smidr in the same repo and integrate via internal APIs.

The Smidr daemon is a planned gRPC service that exposes Smidr’s build orchestration and artifact management over a network API. This enables remote automation, CI/CD integration, and the Yggdrasil business application (C# WebAPI + Next.js) to consume Smidr capabilities.

## Overview

- Runs as a persistent process (systemd service or container)
- Accepts build requests, streams logs, manages artifacts
- Uses the same Go engine as the CLI

## Example API Methods (proto outline)

```proto
service Smidr {
  rpc StartBuild(StartBuildRequest) returns (BuildStatus);
  rpc GetBuildStatus(BuildStatusRequest) returns (BuildStatus);
  rpc ListArtifacts(ListArtifactsRequest) returns (ArtifactsList);
  rpc StreamLogs(StreamLogsRequest) returns (stream LogLine);
  rpc CancelBuild(CancelBuildRequest) returns (CancelResult);
}
```

## Typical Workflow

1. Client sends `StartBuild` with config and parameters
2. Daemon launches build, streams logs via `StreamLogs`
3. Client polls or subscribes to `GetBuildStatus`
4. On completion, client calls `ListArtifacts` to fetch outputs

## Security

- Planned: mTLS or token-based authentication
- Designed for local or remote CI/CD, developer workstations, or build farms

## Configuration

- Uses standard Smidr YAML config files
- Can be run with custom config via CLI or environment

## Example Usage

```bash
# Start the daemon
smidr daemon --config smidr.yaml

# Client (future):
smidr client build --target core-image-minimal --follow
```

## Roadmap

- Initial release: local builds, log streaming, artifact listing
- Future: remote layer management, build queueing, user authentication, Yggdrasil web application integration
