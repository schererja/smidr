# ⚒️ Smidr

**The Digital Forge for Embedded Linux**

Smidr is a modern, cloud-native build orchestration system for embedded Linux development. Built with .NET/C# and gRPC, it provides a robust API-first approach to managing Yocto/BitBake builds with container isolation, intelligent caching, and real-time monitoring.

> **Note**: This project is actively being refactored to a unified C# codebase. Documentation and examples are being updated to reflect the new architecture.

---

## 🔥 The Problem

Traditional embedded Linux development with Yocto/BitBake suffers from fundamental issues:

- **Massive storage waste**: 120GB+ per build environment, multiplied by every image variant
- **Build state corruption**: Mysterious failures requiring full rebuilds
- **Cryptic error messages**: Hours spent debugging parser failures and dependency hell
- **No parallelization**: Can't work on multiple images without massive resource overhead
- **Fragile environments**: One misconfigured layer breaks everything

## ⚡ The Smidr Solution

Smidr reimagines embedded Linux builds using modern container technology and cloud-native patterns:

### 🏗️ API-First Architecture

- **gRPC services**: Type-safe, efficient Protocol Buffer-based APIs
- **REST API**: HTTP/JSON wrapper for easy integration
- **Real-time streaming**: Live log tailing and status updates
- **Multi-client support**: CLI, web UI, or custom integrations

### 🐳 Container-Native Builds

- **Isolated builds**: Every build runs in a clean Docker container
- **Shared caching**: Intelligent layer sharing reduces storage by 90%+
- **Parallel execution**: Build multiple images simultaneously without conflicts
- **Reproducible results**: Same input always produces the same output

### 💾 Intelligent Storage Management

- **Deduplicated storage**: Common base layers shared across all builds
- **Smart caching**: Downloads, sstate, and layer repositories shared
- **Incremental builds**: Only rebuild what actually changed

---

## 🚀 Quick Start

### Prerequisites

- .NET 8.0+ SDK
- Docker (for containerized Yocto builds)
- Supported platforms: Linux, macOS, Windows

### Run the API Server

Start the Smidr Web API server (REST wrapper around gRPC daemon):

```bash
cd api/src/Schererja.Smidr.WebAPI
dotnet run
```

The API will start on `http://localhost:5000` (or `https://localhost:5001`) with Swagger UI at the root.

### Using the REST API

Once the API is running, interact with it via HTTP:

```bash
# Start a new build
curl -X POST http://localhost:5000/api/builds \
  -H "Content-Type: application/json" \
  -d '{
    "configPath": "/path/to/smidr.yaml",
    "target": "core-image-minimal",
    "customer": "demo"
  }'

# Get build status
curl http://localhost:5000/api/builds/{buildId}

# List all builds
curl http://localhost:5000/api/builds

# Stream logs (Server-Sent Events)
curl -N http://localhost:5000/api/builds/{buildId}/logs?follow=true

# List artifacts
curl http://localhost:5000/api/builds/{buildId}/artifacts

# Cancel a build
curl -X DELETE http://localhost:5000/api/builds/{buildId}
```

### gRPC Daemon (Background Service)

The gRPC daemon runs as a background service, handling build orchestration:

```bash
# Configured via appsettings.json or environment variables
export Smidr__DaemonUrl=http://localhost:50051
dotnet run
```

### Configuration

Smidr uses YAML configuration files to define build parameters:

```yaml
name: toradex-custom-image
description: "Custom Toradex image with application stack"

base:
  provider: toradex
  machine: verdin-imx8mp
  distro: tdx-xwayland
  version: "6.0.0"

layers:
  - name: meta-toradex-bsp-common
    git: https://git.toradex.com/meta-toradex-bsp-common
    branch: kirkstone-6.x.y

  - name: meta-mycompany
    path: ./layers/meta-mycompany

build:
  image: core-image-weston
  extra_packages:
    - python3
    - docker
    - nodejs
  bb_number_threads: 8
  parallel_make: 8

artifacts:
  - "*.wic"
  - "*.tar.bz2"
  - "*-sdk-*.sh"

directories:
  layers: ${SMIDR_LAYERS_DIR:-~/.smidr/layers}
  downloads: ${SMIDR_DL_DIR:-~/.smidr/downloads}
  sstate: ${SMIDR_SSTATE_DIR:-~/.smidr/sstate}
```

Common environment variable overrides:

- `SMIDR_LAYERS_DIR` — cache of cloned layer repositories
- `SMIDR_DL_DIR` — shared downloads
- `SMIDR_SSTATE_DIR` — shared sstate cache
- `SMIDR_TMP_DIR` — build tmpdir
- `SMIDR_DEPLOY_DIR` — deploy artifacts

---

## 🏛️ Architecture

```
┌──────────────────┐
│   Web UI / CLI   │
│   (REST Client)  │
└────────┬─────────┘
         │ HTTP/REST
         ▼
┌──────────────────────────┐
│   ASP.NET Web API        │
│   (REST → gRPC Bridge)   │
└────────┬─────────────────┘
         │ gRPC
         ▼
┌──────────────────────────┐
│   .NET gRPC Daemon       │
│   - BuildService         │
│   - LogService           │
│   - ArtifactService      │
└────────┬─────────────────┘
         │
         ├─▶ SQLite DB (EF Core)
         │   - Build tracking
         │   - Artifact metadata
         │
         ├─▶ Docker Engine
         │   - Container orchestration
         │   - Isolated build environments
         │
         └─▶ Shared Cache
             - Layer repositories
             - Downloads (DL_DIR)
             - Shared state (sstate-cache)
```

### Technology Stack

- **API Layer**: ASP.NET Core 8.0+ (REST API with Swagger)
- **Service Layer**: gRPC (.NET implementation)
- **Database**: SQLite with Entity Framework Core
- **Container Runtime**: Docker via Docker.DotNet
- **Configuration**: YamlDotNet for YAML parsing
- **Logging**: Serilog structured logging
- **Protocol**: Protocol Buffers (protobuf) for service contracts

---

## 🎯 Project Status

**Current Phase**: C# Refactoring & Consolidation

### ✅ Completed

### 🚧 In Progress

- [ ] Protocol buffer service definitions (BuildService, LogService, ArtifactService)
- [ ] ASP.NET Core Web API with Swagger
- [ ] REST → gRPC client integration
- [ ] Project architecture and documentation
- [ ] .NET gRPC daemon implementation (replacing Go daemon)
- [ ] Entity Framework Core database layer
- [ ] Docker.DotNet container orchestration
- [ ] YAML configuration parsing with YamlDotNet
- [ ] Build execution pipeline

### 📋 Planned Features

**v0.1.0 (Core Daemon MVP)**

- gRPC server implementation in C#
- BuildService: `StartBuild`, `GetBuildStatus`, `ListBuilds`
- SQLite persistence with EF Core (builds table)
- Basic Docker container execution
- YAML config parsing (minimal: name, target, layers)
- REST API wrapper (existing)
- End-to-end: API → Daemon → Docker → Database

**v0.2.0 (Yocto Integration)**

- Full BitBake/Yocto build support
- Real-time log streaming (LogService)
- Artifact extraction and management (ArtifactService)
- Layer fetching and caching
- Build state tracking (QUEUED → BUILDING → COMPLETED/FAILED)

**v0.3.0 (Production Ready)**

- Multi-tenant build queuing
- Advanced caching strategies (sstate, downloads)
- Enhanced error handling and recovery
- Build cancellation support
- Metrics and monitoring

**Future**

- Web UI (Blazor or React)
- CI/CD integration templates
- Multi-vendor BSP support (Toradex, NXP, Qualcomm)
- Distributed build execution
- Cloud-native deployment options
- Hosted/SaaS mode (build orchestration without local execution)

---

## 🛰️ gRPC Services

Smidr provides a comprehensive gRPC API for build management:

### BuildService

- `StartBuild` — Launch a new Yocto build with configuration
- `GetBuildStatus` — Query real-time build status
- `GetBuild` — Retrieve detailed build information
- `ListBuilds` — List all builds with filtering
- `CancelBuild` — Abort a running build
- `DeleteBuild` — Soft-delete completed builds
- `PurgeBuilds` — Cleanup old builds

### LogService

- `StreamBuildLogs` — Real-time log streaming with follow mode
- Supports both historical and live log tailing

### ArtifactService

- `ListArtifacts` — Enumerate build outputs (WIC images, tarballs, SDKs)
- `GetArtifact` — Retrieve artifact metadata
- `DownloadArtifact` — Get artifact download URLs
- `DeleteArtifact` — Remove specific artifacts

All services use Protocol Buffers for type-safe, efficient communication. See `protos/smidr/v1/` for complete service definitions.

---

## 📖 Documentation

- [.NET Daemon Architecture](docs/smidr-dotnet.md) — Complete C# implementation guide
- [API Reference](http://localhost:5000) — Interactive Swagger UI when running
- [Protocol Buffers](protos/smidr/v1/) — gRPC service definitions

### Development

```bash
# Clone repository
git clone https://github.com/schererja/smidr.git
cd smidr

# Build API
cd api/src/Schererja.Smidr.WebAPI
dotnet build

# Run API
dotnet run

# Generate C# from protos
cd ../../..
buf generate
```

---

## 🤝 Contributing

Contributions are welcome! This project is being actively developed with a focus on C#/.NET modernization.

Areas of interest:

- .NET gRPC daemon implementation
- Entity Framework Core database layer
- Docker integration improvements
- Web UI development (Blazor/React)
- Documentation and examples

See [docs/smidr-dotnet.md](docs/smidr-dotnet.md) for implementation details.

---

## 👨‍💻 Author

**Jason Scherer**

- Embedded systems engineer with 20+ years experience
- Focused on solving real problems in Yocto/BitBake workflows
- Building tools that developers actually want to use

---

## 📄 License

Smidr is licensed under the [MIT License](LICENSE).

---

## 🔗 Links

- **Repository**: [github.com/schererja/smidr](https://github.com/schererja/smidr)
- **Issues**: [GitHub Issues](https://github.com/schererja/smidr/issues)

---

<div align="center">
<strong>Transform your embedded Linux builds from painful to powerful.</strong>
<br><br>
⚒️ <strong>Smidr - Forge Better Builds</strong> ⚒️
</div>
