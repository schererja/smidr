# Smidr Daemon

.NET 8.0 gRPC daemon for orchestrating containerized Yocto/BitBake builds.

## Overview

The Smidr daemon is the core build orchestration service that:

- Exposes gRPC APIs for build management (BuildService, LogService, ArtifactService)
- Manages Docker containers for isolated build execution
- Persists build state and metadata in SQLite via Entity Framework Core
- Handles concurrent build queuing and execution
- Streams real-time build logs to clients

## Project Structure

```shell
daemon/
  src/
    Smidr.Daemon/
      Core/
        Configuration/
        Scheduling/
        Diagnostics/
        Security/
      Providers/
        Containers/
          Docker/
          Podman/
        Virtualization/
          LXC/
          QEMU/
        Networking/
      Services/
        Grpc/
        ApiModels/
      State/
        TaskState/
        SystemState/
      DaemonHost.cs
      Program.cs
  tests/
    Smidr.Daemon.Tests/
      Core/
      Providers/
      Services/
  docs/
    architecture/
      overview.md
      providers.md
      scheduler.md
      runtime.md
    setup/
      install-linux.md
      install-macos.md
      install-windows.md
  build/
    scripts/
    service-config/
```

## Getting Started

### Prerequisites

- .NET 8.0 SDK or later
- Docker (for build execution)
- Protocol buffer compiler (for proto generation)

### Create the Project

```bash
cd daemon

# Create solution
dotnet new sln -n Smidr.Daemon

# Create main project (console, not web!)
dotnet new console -n Smidr.Daemon -o src/Smidr.Daemon -f net8.0
dotnet sln add src/Smidr.Daemon

# Create test project
dotnet new xunit -n Smidr.Daemon.Tests -o tests/Smidr.Daemon.Tests -f net8.0
dotnet sln add tests/Smidr.Daemon.Tests
dotnet add tests/Smidr.Daemon.Tests reference src/Smidr.Daemon
```

### Install Dependencies

```bash
cd src/Smidr.Daemon

# gRPC (standalone server, not ASP.NET Core)
dotnet add package Grpc.Core
dotnet add package Grpc.Tools
dotnet add package Google.Protobuf

# Database
dotnet add package Microsoft.EntityFrameworkCore.Sqlite
dotnet add package Microsoft.EntityFrameworkCore.Design

# Docker
dotnet add package Docker.DotNet

# Configuration
dotnet add package YamlDotNet
dotnet add package Microsoft.Extensions.Configuration
dotnet add package Microsoft.Extensions.Configuration.Json
dotnet add package Microsoft.Extensions.Configuration.EnvironmentVariables

# Logging
dotnet add package Serilog
dotnet add package Serilog.Sinks.Console
dotnet add package Serilog.Sinks.File

# Hosting (for background services/graceful shutdown)
dotnet add package Microsoft.Extensions.Hosting
```

### Configuration

Environment variables:

- `SMIDR_HOST` - gRPC server host (default: `0.0.0.0`)
- `SMIDR_PORT` - gRPC server port (default: `50051`)
- `SMIDR_DB_PATH` - SQLite database path (default: `smidr.db`)
- `SMIDR_LOG_LEVEL` - Logging level (default: `Information`)

### Build & Run

```bash
# Build
dotnet build

# Run daemon
dotnet run --project src/Smidr.Daemon

# Run with custom settings
SMIDR_PORT=8080 dotnet run --project src/Smidr.Daemon

# Run tests
dotnet test
```

## Development Roadmap

### Phase 1: v0.1.0 MVP

- [x] Project structure setup
- [ ] gRPC service skeleton (BuildService only)
- [ ] SQLite database with EF Core
- [ ] Basic Docker container execution
- [ ] YAML config parsing (minimal)
- [ ] StartBuild, GetBuildStatus, ListBuilds implementation

### Phase 2: v0.2.0 Yocto Integration

- [ ] Full BitBake/Yocto build support
- [ ] LogService with real-time streaming
- [ ] ArtifactService implementation
- [ ] Layer fetching and Git integration
- [ ] Build state machine (QUEUED → BUILDING → COMPLETED/FAILED)

### Phase 3: v0.3.0 Production

- [ ] Multi-tenant build queuing
- [ ] Advanced caching (sstate, downloads)
- [ ] Build cancellation
- [ ] Error recovery and retry
- [ ] Metrics and health checks

## Architecture

### gRPC Services

The daemon implements three main gRPC services defined in `../../protos/smidr/v1/`:

1. **BuildService** - Build lifecycle management
   - `StartBuild` - Initiate new build
   - `GetBuildStatus` - Query build state
   - `GetBuild` - Detailed build info
   - `ListBuilds` - Query builds with filters
   - `CancelBuild` - Abort running build
   - `DeleteBuild` - Soft delete
   - `PurgeBuilds` - Cleanup old builds

2. **LogService** - Real-time log streaming
   - `StreamBuildLogs` - Stream logs with follow mode

3. **ArtifactService** - Build output management
   - `ListArtifacts` - List build outputs
   - `GetArtifact` - Artifact metadata
   - `DownloadArtifact` - Get download URL
   - `DeleteArtifact` - Remove artifact

### Database Schema

SQLite database managed by EF Core:

**builds** table:

- `id` (string, PK) - Build identifier
- `customer_id` (string) - Customer/tenant
- `target` (string) - BitBake target
- `config_path` (string) - YAML config location
- `state` (string) - Build state enum
- `start_time` (long) - Unix timestamp
- `end_time` (long, nullable) - Unix timestamp
- `exit_code` (int, nullable) - Build exit code
- `error_message` (string, nullable) - Error details

Future tables:

- `build_artifacts` - Artifact metadata
- `build_logs` - Log entries

### Docker Integration

Uses Docker.DotNet for container orchestration:

- Pull base images
- Create containers with volume mounts
- Execute BitBake commands
- Stream logs via attach
- Extract artifacts from container
- Cleanup on completion

### Build Queue

Concurrent build management:

- Global semaphore limiting total builds
- Per-customer queues for serialization
- Async/await based execution
- Cancellation token support

## Testing

```bash
# Run all tests
dotnet test

# Run with coverage
dotnet test /p:CollectCoverage=true

# Run specific test
dotnet test --filter "FullyQualifiedName~BuildServiceTests"
```

## Deployment

### Docker

```bash
# Build image
docker build -t smidr-daemon:latest .

# Run container
docker run -d \
  --name smidr-daemon \
  -p 50051:50051 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v ~/.smidr:/data \
  smidr-daemon:latest
```

### Systemd Service

```ini
[Unit]
Description=Smidr Build Daemon
After=docker.service
Requires=docker.service

[Service]
Type=notify
ExecStart=/usr/bin/dotnet /opt/smidr/Smidr.Daemon.dll
Restart=on-failure
Environment=SMIDR_DB_PATH=/var/lib/smidr/smidr.db

[Install]
WantedBy=multi-user.target
```

## Documentation

- [Architecture Overview](../docs/architecture/overview.md)
- [Daemon Design](../docs/architecture/daemon-design.md)
- [ADRs](../docs/decisions/)
- [Protocol Definitions](../protos/smidr/v1/)
- [REST API](../api/README.md)
- [UI](../ui/README.md)

## License

See [LICENSE](../LICENSE)
