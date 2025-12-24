# Smidr .NET/C# Daemon Implementation Guide

## Overview

This document outlines the plan to create a .NET/C# implementation of the Smidr build daemon, matching the functionality of the existing Go and TypeScript implementations. The goal is to provide a unified C# codebase for the entire Smidr project while maintaining protocol buffer-based interoperability and gRPC service contracts.

### Motivation

- **Unified Language Stack**: Consolidate project to primarily C# for maintainability
- **Learning Opportunity**: Deep dive into C# ecosystem and best practices
- **Full Feature Parity**: Implement all daemon services available in Go version
- **Protocol Buffer Alignment**: Leverage existing proto definitions for consistent API contracts

### Architecture Overview

The daemon architecture mirrors the existing Go and TypeScript implementations:

```
┌─────────────────────────────────────────────────────┐
│         gRPC Service Layer                          │
├─────────────────────────────────────────────────────┤
│  BuildService  │  LogService  │  ArtifactService   │
└────────┬────────────────────────────────┬───────────┘
         │                                │
    ┌────▼──────────────────────────┐    │
    │  Build Orchestration          │    │
    │  (Runner Pipeline)            │    │
    └────┬──────────────────────────┘    │
         │                                │
    ┌────▼─────────┐  ┌───────────────┐  │
    │  Container   │  │  Layer        │  │
    │  Manager     │  │  Fetcher      │  │
    │  (Docker)    │  │  (Git)        │  │
    └──────────────┘  └───────────────┘  │
         │                                │
    ┌────▼──────────────────────────┐    │
    │  BitBake Config Generator     │    │
    └────┬──────────────────────────┘    │
         │                                │
    ┌────▼──────────────────────────┐    │
    │  Artifact Manager             │◄───┘
    │  (Extraction & Metadata)      │
    └────┬──────────────────────────┘
         │
    ┌────▼──────────────────────────┐
    │  Database Layer (EF Core)     │
    │  (SQLite: builds, artifacts)  │
    └────┬──────────────────────────┘
         │
    ┌────▼──────────────────────────┐
    │  Supporting Services          │
    │  • Logging (Serilog)          │
    │  • Config (YamlDotNet)        │
    │  • Concurrency (Semaphore)    │
    └───────────────────────────────┘
```

---

## Step 1: Create .NET Project Structure

### Objectives

- Establish initial project layout with proper C# conventions
- Configure build system and dependencies
- Set up IDE integration and development workflow

### Tasks

#### 1.1 Create Project Directory

```bash
mkdir -p apps/daemon-cs
cd apps/daemon-cs
```

#### 1.2 Create .csproj File

**File**: `apps/daemon-cs/smidr-daemon.csproj`

```xml
<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
    <OutputType>Exe</OutputType>
    <TargetFramework>net8.0</TargetFramework>
    <Nullable>enable</Nullable>
    <ImplicitUsings>enable</ImplicitUsings>
    <LangVersion>latest</LangVersion>
    <RootNamespace>Smidr.Daemon</RootNamespace>
    <AssemblyName>smidr-daemon</AssemblyName>
    <Version>0.1.0</Version>
  </PropertyGroup>

  <ItemGroup>
    <!-- gRPC -->
    <PackageReference Include="Grpc.AspNetCore" Version="2.60.0" />
    <PackageReference Include="Google.Protobuf" Version="3.25.1" />
    <PackageReference Include="Grpc.Tools" Version="2.60.0">
      <PrivateAssets>all</PrivateAssets>
      <IncludeAssets>runtime; build; native; contentfiles; analyzers; buildtransitive</IncludeAssets>
    </PackageReference>

    <!-- Database -->
    <PackageReference Include="Microsoft.EntityFrameworkCore" Version="8.0.0" />
    <PackageReference Include="Microsoft.EntityFrameworkCore.Sqlite" Version="8.0.0" />

    <!-- Configuration -->
    <PackageReference Include="YamlDotNet" Version="14.1.1" />

    <!-- Container Management -->
    <PackageReference Include="Docker.DotNet" Version="3.125.15" />

    <!-- Logging -->
    <PackageReference Include="Serilog" Version="3.1.1" />
    <PackageReference Include="Serilog.Sinks.Console" Version="5.0.0" />
    <PackageReference Include="Serilog.Sinks.File" Version="5.0.0" />
    <PackageReference Include="Serilog.Sinks.Async" Version="2.0.0" />
    <PackageReference Include="Serilog.Formatting.Json" Version="3.0.0" />

    <!-- CLI / Hosting -->
    <PackageReference Include="Microsoft.Extensions.Hosting" Version="8.0.0" />
    <PackageReference Include="Microsoft.Extensions.DependencyInjection" Version="8.0.0" />
    <PackageReference Include="Microsoft.Extensions.Configuration" Version="8.0.0" />

    <!-- Utilities -->
    <PackageReference Include="System.Net.Http.Json" Version="8.0.0" />
  </ItemGroup>

  <!-- Proto Files -->
  <ItemGroup>
    <Protobuf Include="../../protos/smidr/v1/*.proto" ProtoRoot="../../protos" GrpcServices="Server" />
  </ItemGroup>

</Project>
```

#### 1.3 Create Solution File (Optional)

```bash
cd /Users/schererja/src/github.com/schererja/smidr
dotnet new sln --name Smidr
dotnet sln add apps/daemon-cs/smidr-daemon.csproj
```

#### 1.4 Create .gitignore

**File**: `apps/daemon-cs/.gitignore`

```
# Build artifacts
bin/
obj/
*.dll
*.exe
*.pdb
*.so
*.dylib

# User-specific files
.vs/
.vscode/
*.user
*.suo

# Dependencies
packages/
.nuget/

# Logs
logs/
*.log

# Database
*.db
*.db-journal

# Environment
.env
.env.local
appsettings.*.json

# OS
.DS_Store
Thumbs.db

# IDE
.idea/
*.swp
*.swo
*~
```

#### 1.5 Create Directory Structure

```bash
mkdir -p apps/daemon-cs/src/{Services,Models,Infrastructure,Data,Configuration,Utilities,Logging}
mkdir -p apps/daemon-cs/tests/Smidr.Daemon.Tests
mkdir -p apps/daemon-cs/docker
```

#### 1.6 Create Dockerfile

**File**: `apps/daemon-cs/docker/Dockerfile`

```dockerfile
# Build stage
FROM mcr.microsoft.com/dotnet/sdk:8.0-alpine AS build

WORKDIR /src

# Copy project file
COPY ["smidr-daemon.csproj", "./"]

# Restore dependencies
RUN dotnet restore "smidr-daemon.csproj"

# Copy source code
COPY ["src/", "src/"]
COPY ["../../protos/", "protos/"]

# Build
RUN dotnet build "smidr-daemon.csproj" -c Release -o /app/build

# Publish stage
FROM mcr.microsoft.com/dotnet/runtime:8.0-alpine AS runtime

WORKDIR /app

# Install runtime dependencies for Yocto container interaction
RUN apk add --no-cache \
    ca-certificates \
    docker-cli

# Copy built application
COPY --from=build /app/build .

# Create non-root user
RUN addgroup -g 1000 builder && \
    adduser -u 1000 -G builder -s /sbin/nologin -D builder && \
    chown -R builder:builder /app

USER builder

# gRPC port
EXPOSE 50051

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD dotnet health || exit 1

# Run the daemon
ENTRYPOINT ["dotnet", "smidr-daemon.dll"]
```

#### 1.7 Create Makefile

**File**: `apps/daemon-cs/Makefile`

```makefile
.PHONY: help restore build dev release clean test lint run docker

DOTNET ?= dotnet
PROJECT := smidr-daemon.csproj
CONFIGURATION ?= Debug

help:
 @echo "Smidr .NET Daemon - Available targets:"
 @echo "  make restore    - Restore NuGet packages"
 @echo "  make build      - Build Debug configuration"
 @echo "  make release    - Build Release configuration"
 @echo "  make test       - Run unit tests"
 @echo "  make lint       - Run code analysis"
 @echo "  make run        - Run daemon locally"
 @echo "  make clean      - Remove build artifacts"
 @echo "  make docker     - Build Docker image"
 @echo "  make watch      - Watch and rebuild on changes"

restore:
 $(DOTNET) restore $(PROJECT)

build: restore
 $(DOTNET) build $(PROJECT) -c $(CONFIGURATION) --no-restore

release: restore
 $(DOTNET) build $(PROJECT) -c Release --no-restore

test: build
 $(DOTNET) test --no-build -v normal

lint:
 $(DOTNET) build $(PROJECT) /p:TreatWarningsAsErrors=true --no-restore

run: build
 $(DOTNET) run --project $(PROJECT) --no-build

watch:
 $(DOTNET) watch --project $(PROJECT) run

clean:
 $(DOTNET) clean $(PROJECT)
 rm -rf bin obj *.db *.log

docker:
 docker build -f docker/Dockerfile -t smidr-daemon-cs:latest .
 @echo "Built Docker image: smidr-daemon-cs:latest"

docker-run: docker
 docker run -d --name smidr-daemon-cs \
  -p 50051:50051 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  smidr-daemon-cs:latest
```

#### 1.8 Create README

**File**: `apps/daemon-cs/README.md`

```markdown
# Smidr .NET/C# Daemon

Yocto build daemon implementation in .NET/C# with gRPC services, Docker-based build execution, and SQLite persistence.

## Features

- **gRPC Services** - BuildService, LogService, ArtifactService
- **Build Orchestration** - Yocto build pipeline with Docker containers
- **Database Persistence** - Entity Framework Core + SQLite
- **Configuration Management** - YAML-based build configs with variable expansion
- **Logging** - Structured logging with Serilog
- **Concurrency Control** - Per-customer build queues with global limits
- **Multi-Tenant** - Isolated builds per customer with independent queues

## Prerequisites

- .NET 8.0 SDK or later
- Docker (for running Yocto builds)
- Docker socket access (`/var/run/docker.sock`)

## Installation

```bash
cd apps/daemon-cs
dotnet restore
```

## Development

```bash
# Build
make build

# Run locally
make run

# Watch and rebuild on changes
make watch

# Run tests
make test

# Code analysis
make lint
```

## Production

```bash
# Build release
make release

# Run
bin/Release/net8.0/smidr-daemon

# Or use Docker
make docker
docker run -p 50051:50051 -v /var/run/docker.sock:/var/run/docker.sock smidr-daemon-cs:latest
```

## Configuration

### Environment Variables

- `GRPC_HOST` - gRPC server host (default: `0.0.0.0`)
- `GRPC_PORT` - gRPC server port (default: `50051`)
- `DB_PATH` - SQLite database file path (default: `smidr.db`)
- `PROTO_DIR` - Protocol buffer definitions directory (default: `../../protos/smidr/v1`)
- `LOG_LEVEL` - Serilog minimum level (default: `Information`)
- `DOCKER_HOST` - Docker daemon socket or TCP endpoint (default: `unix:///var/run/docker.sock`)

### appsettings.json

```json
{
  "Serilog": {
    "MinimumLevel": "Information",
    "WriteTo": [
      {
        "Name": "Console",
        "Args": {
          "theme": "Ansi"
        }
      },
      {
        "Name": "File",
        "Args": {
          "path": "logs/daemon-.log",
          "rollingInterval": "Day",
          "outputTemplate": "{Timestamp:yyyy-MM-dd HH:mm:ss.fff zzz} [{Level:u3}] {Message:lj}{NewLine}{Exception}"
        }
      }
    ]
  },
  "Grpc": {
    "Host": "0.0.0.0",
    "Port": 50051
  },
  "Database": {
    "ConnectionString": "Data Source=smidr.db"
  },
  "Docker": {
    "Endpoint": "unix:///var/run/docker.sock"
  }
}
```

## Architecture

### Project Structure

```
smidr-daemon-cs/
├── src/
│   ├── Services/           # gRPC service implementations
│   ├── Models/             # Data models and entities
│   ├── Infrastructure/     # Database, Docker, external service clients
│   ├── Data/               # Entity Framework DbContext and migrations
│   ├── Configuration/      # Build config parser and models
│   ├── Utilities/          # Helper functions and extensions
│   ├── Logging/            # Serilog configuration
│   ├── Program.cs          # Application entry point
│   └── HostedServices/     # Background services and gRPC listener
├── tests/
│   └── Smidr.Daemon.Tests/
│       ├── Services/       # Service unit tests
│       ├── Infrastructure/ # Docker, DB, fetcher tests
│       └── Configuration/  # Config parser tests
├── docker/
│   └── Dockerfile          # Container image definition
├── smidr-daemon.csproj
├── Makefile
└── README.md
```

### Key Components

#### Services (`src/Services/`)

- **BuildService** - gRPC BuildService implementation
  - `StartBuild()` - Initiate async build
  - `GetBuildStatus()` - Query build state
  - `ListBuilds()` - Paginated build listing
  - `CancelBuild()` - Abort in-progress build
  - `GetBuild()` - Detailed build info
  - `DeleteBuild()` / `PurgeBuilds()` - Artifact cleanup

- **LogService** - gRPC LogService implementation
  - `StreamBuildLogs()` - Real-time log streaming with follow support

- **ArtifactService** - gRPC ArtifactService implementation
  - `ListArtifacts()` - List build artifacts
  - `GetArtifact()` - Artifact metadata
  - `DownloadArtifact()` - Download URL
  - `DeleteArtifact()` - Remove artifact

#### Infrastructure (`src/Infrastructure/`)

- **ContainerManager** - Docker client abstraction
  - Image pulling with retry
  - Container lifecycle management
  - Mount and volume configuration
  - Command execution with streaming
  - Artifact file copying

- **LayerFetcher** - Git-based layer management
  - Parallel clone/fetch of Yocto layers
  - Layer validation and caching
  - Mirror support

- **ArtifactManager** - Build output handling
  - Artifact extraction from containers
  - Metadata JSON generation
  - Checksum calculation

#### Data (`src/Data/`)

- **SmidrDbContext** - Entity Framework Core DbContext
  - Build entity and relationships
  - Migration support
  - Query helpers

#### Configuration (`src/Configuration/`)

- **BuildConfig** - YAML config model and parser
- **ConfigManager** - Config loading, validation, and variable expansion
- **Variable Expansion** - ${TOPDIR}, ${MACHINE}, etc.

---

## Comparison: Go vs TypeScript vs .NET

| Feature | Go | TypeScript | .NET |
|---------|-----|-----------|-----|
| **Framework** | stdlib gRPC | @grpc/grpc-js | Grpc.AspNetCore |
| **Database** | mattn/go-sqlite3 | sql.js | EF Core + SQLite |
| **Config** | spf13/viper | YamlDotNet | YamlDotNet |
| **Container** | docker/docker | dockerode | Docker.DotNet |
| **Logging** | log/slog | Winston | Serilog |
| **Concurrency** | goroutines + sync.Semaphore | async/Promise + Semaphore | Task + SemaphoreSlim |
| **Status** | Production | Functional | In Development |

---

## Testing

### Unit Tests

```bash
dotnet test --filter "Category=Unit"
```

### Integration Tests (Docker Required)

```bash
dotnet test --filter "Category=Integration"
```

### Full Test Suite

```bash
make test
```

---

## Troubleshooting

### Docker Connection Issues

```bash
# Verify Docker socket access
ls -la /var/run/docker.sock

# Check daemon logs
dotnet run -- --log-level=Debug
```

### Database Errors

```bash
# Reset database (development only)
rm smidr.db
dotnet ef database update
```

### Proto Compilation

```bash
# Regenerate C# from proto files
dotnet build /p:RegenerateProtos=true
```

---

## Development Roadmap

### Phase 1: Core gRPC Services

- [ ] Project setup and dependencies
- [ ] gRPC service stubs from protos
- [ ] Basic service implementations
- [ ] Unit tests

### Phase 2: Database & Persistence

- [ ] Entity Framework Core + SQLite schema
- [ ] Build entity models
- [ ] Database queries and mutations
- [ ] Migration support

### Phase 3: Build Orchestration

- [ ] Configuration manager (YAML parsing)
- [ ] Container manager (Docker.DotNet)
- [ ] Build runner pipeline
- [ ] Layer fetcher (Git operations)

### Phase 4: Artifact & Logging

- [ ] Artifact extraction and metadata
- [ ] Serilog integration
- [ ] Real-time log streaming
- [ ] Build log dual format (plain + JSON)

### Phase 5: Concurrency & Testing

- [ ] Per-customer build queues
- [ ] Global concurrency limits
- [ ] Integration tests with Docker
- [ ] Performance testing

### Phase 6: Deployment & Documentation

- [ ] Docker image refinement
- [ ] Kubernetes manifests (optional)
- [ ] CLI tool for daemon management
- [ ] API documentation (Swagger/OpenAPI)

---

## Contributing

When adding features:

1. Follow C# naming conventions (PascalCase for public members)
2. Use `async/await` for I/O operations
3. Write unit tests with AAA pattern (Arrange-Act-Assert)
4. Document public APIs with XML comments
5. Use dependency injection for testability
6. Log important events with structured logging

---

## Related Documentation

- [Go Daemon Implementation](./daemon.md)
- [Protocol Buffers & gRPC](../protos/)
- [Build Configuration Format](./configuration.md)

---

## License

See root LICENSE file

```

### Deliverables

- ✅ Project directory structure created
- ✅ .csproj with all required dependencies
- ✅ Dockerfile for containerized deployment
- ✅ Makefile with convenient build targets
- ✅ Comprehensive README with setup and architecture overview
- ✅ .gitignore for build artifacts

---

## Next Steps

When Step 1 is complete:

1. Run `dotnet restore` to verify dependencies resolve
2. Run `dotnet build` to verify initial compilation
3. Proceed to Step 2: Implement gRPC Service Stubs

---

## Step 2: Implement gRPC Services

### Objectives

- Create service interfaces matching proto definitions
- Implement BuildService, LogService, ArtifactService
- Setup service registration in dependency injection
- Create initial message handling without business logic

### Tasks (To Be Documented in Step 2)

---

## Step 3: Build Database Layer with EF Core

(To be documented)

---

## Step 4: Configuration Management

(To be documented)

---

## Step 5: Build Orchestration & Container Integration

(To be documented)

---

## Step 6: Logging & Artifacts

(To be documented)

---

## Step 7: Concurrency & Testing

(To be documented)

---

## Additional Resources

- [Entity Framework Core Docs](https://learn.microsoft.com/en-us/ef/core/)
- [gRPC for .NET](https://grpc.io/docs/languages/csharp/)
- [Serilog Documentation](https://serilog.net/)
- [Docker.DotNet GitHub](https://github.com/dotnet/Docker.DotNet)
- [YamlDotNet GitHub](https://github.com/aaubry/YamlDotNet)
