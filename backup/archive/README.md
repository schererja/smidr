# ⚒️ Smidr

**WORK IN PROGRESS**
This project is still a work in progress, but after spending a year and half working on yocto and bitbake projects I wanted to make things easier and start to try and make ways to make it a simple process for more people.  Some of the "Nice to haves" will be implemented later on.

**The Digital Forge for Embedded Linux**

Smidr is a modern, container-based build system for embedded Linux that eliminates the pain points of traditional Yocto/BitBake workflows. Built by developers, for developers who are tired of cryptic errors, build state corruption, and massive storage requirements.

---

## 🔥 The Problem

Traditional embedded Linux development with Yocto/BitBake suffers from fundamental issues:

- **Massive storage waste**: 120GB+ per build environment, multiplied by every image variant
- **Build state corruption**: Mysterious failures requiring full rebuilds
- **Cryptic error messages**: Hours spent debugging parser failures and dependency hell
- **No parallelization**: Can't work on multiple images without massive resource overhead
- **Fragile environments**: One misconfigured layer breaks everything

## ⚡ The Smidr Solution

Smidr reimagines embedded Linux builds using modern container technology:

### 🏗️ Container-Native Architecture

- **Isolated builds**: Every build runs in a clean container environment
- **Shared caching**: Intelligent layer sharing dramatically reduces storage requirements
- **Parallel execution**: Build multiple images simultaneously without conflicts
- **Reproducible results**: Same input always produces the same output

### 💾 Intelligent Storage Management

- **Deduplicated storage**: Common base layers (Toradex, BSPs) shared across all builds
- **90%+ space savings**: Turn 360GB of duplicated builds into 40GB of smart cache
- **Incremental builds**: Only rebuild what actually changed

Note: Smidr defaults the BitBake shared state (sstate) cache to `${WORKDIR}/sstate-cache` on the host. When running builds in containers, Smidr bind-mounts the host SSTATE directory into the container at `/home/builder/sstate-cache` so containerized builds can use the shared cache transparently.

### 🛠️ Developer Experience First

- **Clear error messages**: Actionable feedback instead of cryptic stack traces
- **Simple configuration**: YAML-based build definitions that make sense
- **Fast iteration**: Quick builds for development, thorough builds for release

---

## 🚀 Quick Start

### Run the Daemon

Start the Smidr gRPC server to handle build requests:

```bash
# Start daemon on default port (50051)
smidr daemon

# Or specify a custom address and database path
smidr daemon --address :8080 --db-path ~/.smidr/builds.db
```

The daemon will print its listening address and database path at startup. Set `DEBUG=1` for detailed logging:

```bash
DEBUG=1 smidr daemon --address localhost:50051 --db-path ~/.smidr/builds.db
```

### Client Commands

Once the daemon is running, use the client subcommands to interact with it:

```bash
# Start a new build
smidr client start --config smidr.yaml --target core-image-minimal

# Start a build and follow logs immediately
smidr client start --config smidr.yaml --target core-image-minimal --follow

# Check build status
smidr client status --build-id build-123

# Stream logs from a running build
smidr client logs --build-id build-123 --follow

# List all builds
smidr client list

# List artifacts from a completed build
smidr client artifacts build-123

# Cancel a running build
smidr client cancel --build-id build-123

# Connect to a remote daemon
smidr client start --address remote-host:50051 --config smidr.yaml --target my-image
```

### Use an alternate config file

By default Smidr loads `smidr.yaml` in the current directory. You can point to any config file with `--config`:

```bash
smidr --config smidr-toradex-ci.yaml build --customer acme
```

Common per-build overrides are available via environment variables in the CI configs:

- `YOCTO_LAYERS_DIR` — cache of cloned layer repositories (default: `~/.smidr/layers`)
- `YOCTO_DL_DIR` — shared downloads (default: `~/.smidr/downloads`)
- `YOCTO_SSTATE_DIR` — shared sstate cache (default: `~/.smidr/sstate`)
- `YOCTO_TMP_DIR` — build tmpdir (default: per-build under `~/.smidr/builds/.../tmp`)
- `YOCTO_DEPLOY_DIR` — deploy artifacts (default: per-build under `~/.smidr/builds/.../deploy`)

### Example Configuration

```yaml
name: toradex-custom-image
description: "Custom Toradex image with our application stack"

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

custom:
  image: core-image-weston
  extra_packages:
    - python3
    - docker
    - nodejs

artifacts:
  - "*.wic"
  - "*.tar.bz2"
  - "*-sdk-*.sh"

build:
  # Control parallelism (defaults fall back to 2 if not set)
  bb_number_threads: 8
  parallel_make: 8

directories:
  # Explicitly wiring the layers dir is recommended for CI and local caching
  layers: ${YOCTO_LAYERS_DIR:-~/.smidr/layers}
```

---

## 🏛️ Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   smidr.yaml    │───▶│  Smidr Engine    │───▶│  Build Results  │
│  Configuration  │    │   (Go Runtime)   │    │   & Artifacts   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │ Container Pool  │
                       │ ┌─────────────┐ │
                       │ │ Build Env 1 │ │
                       │ ├─────────────┤ │
                       │ │ Build Env 2 │ │
                       │ ├─────────────┤ │
                       │ │ Build Env N │ │
                       │ └─────────────┘ │
                       └─────────────────┘
                                │
                                ▼
                    ┌──────────────────────┐
                    │   Shared Cache       │
                    │ ┌─────────────────┐  │
                    │ │ Source Downloads│  │
                    │ ├─────────────────┤  │
                    │ │ sstate-cache    │  │
                    │ ├─────────────────┤  │
                    │ │ Layer Cache     │  │
                    │ └─────────────────┘  │
                    └──────────────────────┘
```

---

## 🎯 Project Status

**Current Phase**: MVP Development

### ✅ Completed

- [x] Project architecture design
- [x] Core concept validation
- [x] Technology stack selection

### 🚧 In Progress

- [ ] CLI framework implementation
- [ ] Container orchestration system
- [ ] Configuration parser
- [ ] Build execution engine

### 📋 Planned Features

**MVP (v0.1.0)**

- Container-based build isolation
- YAML configuration system
- CLI tools for common operations
- Toradex BSP support
- Basic artifact management

**Post-MVP**

- Web UI for build management
- Advanced caching strategies
- CI/CD integration
- Multi-vendor BSP support
- Cloud build service
- Smidr daemon (gRPC server)

---

## 🏢 Built by Jason Scherer

---

## 🛰️ Smidr Daemon (gRPC Server)

The Smidr daemon is a gRPC server that exposes Smidr's build and artifact management capabilities over a network API. This enables remote orchestration, integration with CI/CD systems, and future web UI or automation tools.

### Running the Daemon

```bash
# Start the daemon on default port (:50051)
smidr daemon

# Or specify a custom address and database path
smidr daemon --address :8080 --db-path ~/.smidr/builds.db
```

For verbose logging, set `DEBUG=1`:

```bash
DEBUG=1 smidr daemon
```

### gRPC Services

The daemon exposes three main service APIs:

- **BuildService**:
  - `StartBuild` — Launch a new build with config and parameters
  - `GetBuildStatus` — Query status and logs for a build
  - `ListBuilds` — List all builds (active and completed)
  - `CancelBuild` — Stop a running build

- **LogService**:
  - `StreamBuildLogs` — Real-time log streaming for active builds

- **ArtifactService**:
  - `ListArtifacts` — Enumerate available build outputs and artifacts

All requests use structured types (e.g., `BuildIdentifier`, `Timestamps`) and return detailed responses with state, exit codes, error messages, and artifact metadata.

### Security & Deployment

- Runs as a system daemon or container
- Auth via mTLS or token (planned)
- Designed for local or remote use in CI/CD, developer workstations, or build farms

### Implementation

- Written in Go, using the same Smidr engine as the CLI
- Uses protobuf for API definition (see `docs/daemon.md` for proto outline)

See [docs/daemon.md](docs/daemon.md) for deeper details and API examples.

Smidr is developed by Jason Scherer, a software engineer focused on solving real problems in embedded systems development and DevOps automation.

**Team Background:**

- 20+ years in embedded systems and automation
- Real-world experience with Yocto's pain points

---

## 📖 Documentation

- [Installation Guide](docs/installation.md)
- [Configuration Reference](docs/configuration.md)
- [CLI Commands](docs/cli-reference.md)
- [Architecture Overview](docs/architecture.md)
- [Contributing Guide](CONTRIBUTING.md)
- [Cache & Source Management](docs/cache.md)
- [Concurrent Builds](docs/concurrent-builds.md) — Run multiple builds with shared caches

### Fast Yocto CI tiers

To validate container wiring and Yocto integration quickly without long builds:

- `make yocto-smoke` — starts a container and runs a lightweight smoke (no build)
- `make yocto-fetch` — runs fetch-only using your restored DL_DIR
- `make yocto-sstate` — attempts a build that should restore from sstate (fast when mirrors hit)

Use the provided `smidr-ci.yaml` with environment variables to steer cache locations:

- `YOCTO_DL_DIR` — path to pre-restored downloads
- `YOCTO_SSTATE_DIR` — path to pre-restored sstate

Advanced knobs (set in `advanced:` or via env in `smidr-ci.yaml`):

- `advanced.sstate_mirrors` — prefer mirrors inside containers (default: `file://.* file:///home/builder/sstate-cache/PATH`)
- `advanced.premirrors` — premirror mapping to your artifact store (e.g. `https?://.*/.* http://your-mirror/`)
- `advanced.bb_no_network` — set to `true` in fully offline CI
- `advanced.bb_fetch_premirroronly` — only download from premirrors

Nightly or scheduled jobs can perform a full build once and publish sstate/downloads to speed up PR runs.

---

## � Troubleshooting

- Duplicate BBFILE_COLLECTIONS or meta-layer collisions
  - Cause: auto-discovered test or skeleton layers polluting `BBLAYERS`.
  - Fix: Smidr now filters `/tests/`, `/testdata/`, `meta-selftest`, and `meta-skeleton` automatically. If you still see duplicates, wipe any hand-edited `conf/bblayers.conf` and re-run, or clean your `directories.layers` cache of stale repos.

- EULA required for NXP/Freescale components
  - Symptom: build errors referencing FSL EULA or i.MX GPU/VPU recipes.
  - Fix: set `advanced.accept_fsl_eula: true` in your config (or use the Toradex CI sample). Smidr writes `ACCEPT_FSL_EULA = "1"` to `local.conf`.

- Threads stuck at 2 in local.conf
  - Cause: wrong config loaded or unset parallelism.
  - Fix: pass the intended file with `--config`, and set `build.bb_number_threads` and `build.parallel_make`. Smidr logs the parsed values in container setup for easy verification.

- Permission errors under TMPDIR or DEPLOY in containers
  - Fix: set `directories.tmp` and `directories.deploy` to host-writable paths (see env overrides in Quick Start). Smidr maps these into the container (`/home/builder/tmp`, `.../deploy`) and writes `TMPDIR` accordingly.

- Sstate-only validation produces no new artifacts
  - Intentional: restore-from-sstate runs are optimized to validate builds quickly.
  - To generate fresh artifacts: disable or change `advanced.sstate_mirrors` for that job, or run a cleansstate/full build tier. Nightly jobs often publish sstate/downloads for the fast PR tiers.

---

## �🤝 Contributing

We welcome contributions! Smidr is built by embedded developers, for embedded developers.

1. Check our [Contributing Guide](CONTRIBUTING.md)
2. Look at [open issues](https://github.com/schereja/smidr/issues)
3. Join the discussion in [GitHub Discussions](https://github.com/schereja/smidr/discussions)

## ⚙️ Configuration essentials

A few key settings make Smidr builds fast, reproducible, and CI-friendly:

- Base selection
  - Set `base.provider`, `base.machine`, `base.distro`, and `base.version` to match your BSP.
  - If you select a Toradex machine like `verdin-imx8mp` without the required BSP layers present, Smidr safely falls back to `qemux86-64` in container preflights to keep CI green while you wire the layers.

- Layers discovery
  - Smidr recurses each repo under `directories.layers` to auto-include sublayers with a valid `conf/layer.conf`.
  - It skips test/example layers automatically: `/tests/`, `/testdata/`, `meta-selftest`, `meta-skeleton`.
  - If you set `yocto_series` (e.g. `kirkstone`), Smidr includes only layers compatible with that series via `LAYERSERIES_COMPAT_*` checks.

- Parallelism
  - Control with `build.bb_number_threads` and `build.parallel_make`. Values <= 0 default to 2 for safety in constrained runners.
  - Recommended: 8 on modern CI executors; tune to your runner’s CPU/memory.

- Package format
  - Default is `package_rpm`. For smaller artifacts and common OE setups, use IPK: `packages.classes: package_ipk` (the CI configs already do this).

- Licenses and EULA
  - For NXP/Freescale GPU/VPU components (e.g., Toradex i.MX), set `advanced.accept_fsl_eula: true` to write `ACCEPT_FSL_EULA = "1"`.
  - This is opt‑in; Smidr never auto‑accepts on your behalf.

- Directories and mounts
  - Pin these to speed up CI and share caches across builds:
    - `directories.layers` — shared clone cache of layer repos
    - `directories.downloads` — BitBake `DL_DIR`
    - `directories.sstate` — shared state cache (`SSTATE_DIR` or `SSTATE_MIRRORS`)
    - `directories.tmp` — `TMPDIR`; mount as a writable host path to avoid container permission issues
    - `directories.deploy` — where artifacts are written
  - In containers, Smidr prefers `SSTATE_MIRRORS` to a bind-mounted `SSTATE_DIR`, mapping to `/home/builder/sstate-cache` by default for fast restores.

### Validation vs. artifact builds

- Fast validation: rely on `SSTATE_MIRRORS` to hit caches and complete quickly.
- Clean artifact run: perform a cleansstate or use a job that doesn’t have the sstate available to force full builds and produce fresh artifacts.

---

## 📄 License

Smidr is licensed under the [MIT License](LICENSE).

---

## 🔗 Links

- **Website**: [smidr.dev](https://smidr.dev)
- **Documentation**: [docs.smidr.dev](https://docs.smidr.dev)
- **Issues**: [GitHub Issues](https://github.com/schereja/smidr/issues)

---

<div align="center">
<strong>Transform your embedded Linux builds from painful to powerful.</strong>
<br><br>
⚒️ <strong>Smidr - Forge Better Builds</strong> ⚒️
</div>
