# SMIDR Daemon

Go-based core daemon and CLI for SMIDR (Simplified, Modern Image Delivery Runtime).

## Overview

The daemon is the core of SMIDR, handling:

- Yocto/BitBake build orchestration
- Source code fetching and caching
- Artifact extraction and management
- Container-based build isolation
- gRPC API server for client communication

## Directory Structure

- `cmd/smidr/` - Main CLI application entry point
- `internal/` - Internal packages (not importable externally)
  - `artifacts/` - Build artifact extraction and cleanup
  - `bitbake/` - BitBake command generation and execution
  - `build/` - Build orchestration and persistence
  - `cli/` - CLI commands and client implementations
  - `config/` - Configuration file parsing
  - `container/` - Docker container management
  - `daemon/` - gRPC daemon server
  - `db/` - SQLite database for build state
  - `source/` - Source fetching and download management
- `pkg/` - Public packages (importable externally)
  - `logger/` - Structured logging utilities

## Building

```bash
make build
```

## Running

```bash
# Start daemon
./smidr daemon

# Run a build
./smidr build -c smidr.yaml
```

## Testing

```bash
make test
```
