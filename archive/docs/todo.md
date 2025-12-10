# Smidr MVP - TODO List

## Daemon (gRPC) – Initial Interface Priorities

Tracked priorities for making the daemon interface production‑ready. Items reflect current status and next concrete steps.

- [ ] Priority 1: Artifact Management
  - [x] Extract artifacts after successful builds (copy from deploy dir into per‑build store)
  - [x] Save artifact metadata (sizes) alongside build metadata
  - [ ] Implement ListArtifacts RPC to read artifacts from store and return full details
    - [ ] Include size (from metadata) and checksum (compute on demand or precompute)
    - [ ] Error handling for non‑completed builds and missing artifact sets
  - [ ] CLI: add download support (DownloadArtifact RPC) and `smidr client download`
  - [ ] Web UI: add download links once RPC is available

- [ ] Priority 2: Improve Build Cancellation
  - [ ] Ensure context cancellation fully propagates through Runner and bitbake executor
  - [ ] Robust cleanup: stop/remove container, clean temp dirs, release resources
  - [ ] State transitions: set CANCELLED deterministically and emit terminal logs
  - [ ] Return clear CancelResult messages (already cancelled, not found, completed)

- [x] Priority 3: Persistence
  - [x] Persist build history (SQLite or lightweight bolt/Badger). Schema: builds, states, timings
  - [x] Persist logs to disk per build; stream from file for historical access
  - [x] Recovery on startup: reload recent builds, reconcile terminal states

- [ ] Priority 4: Authentication & Security
  - [ ] TLS/mTLS for gRPC transport (configurable cert/key paths)
  - [ ] Optional token‑based auth for clients
  - [ ] Configurable bind address; document firewalling and network exposure

### Nice‑to‑have next (post‑initial interface)

- [ ] Concurrent build queue with max‑parallel limit and fair scheduling
- [ ] Health/metrics endpoints (gRPC health; Prometheus metrics)
- [ ] Retention policy in daemon for artifact store (reuse artifacts package policies)
- [ ] Better error surfaces and log categorization in client and Web UI

## Phase 1: Project Setup & Foundation

- [X] Initialize Go project structure
- [X] Set up Git repository under intrik8-labs
- [X] Create basic README with project vision
- [X] Define project directory structure (cmd/, internal/, pkg/)
- [X] Set up Go modules and dependencies
- [X] Create .gitignore for Go projects

## Phase 2: Configuration System

- [X] Design YAML configuration schema
- [X] Implement config file parser (using go-yaml)
- [X] Add config validation logic
- [X] Create `smidr init` command to generate template config
- [X] Add support for environment variable substitution
- [X] Write unit tests for config parsing

## Phase 3: Source Management

- [X] Implement git repository cloning for layers
- [X] Add caching logic for git repositories
- [X] Implement source package download system
- [X] Create persistent cache directory structure
- [X] Add cache invalidation/update logic (cache metadata, TTL eviction, per-repo and per-download)
- [X] Handle download mirrors and fallbacks (mirror/retry logic, tests)

## Phase 4: Container Orchestration

- [X] Research Docker/Podman Go SDK integration

- [X] Create base container image with Yocto dependencies

- [X] Implement container lifecycle management (create/start/stop/destroy) and add unit tests for container and docker packages

- [X] Add volume mounting for downloads and sstate-cache
  - Implemented in `internal/container/docker/docker.go` (host dir creation and mounts) and validated by `internal/container/docker/docker_test.go`.

- [X] Implement layer injection into containers
  - Implemented in `internal/container/docker/docker.go` (mounts each configured layer into `/home/builder/layers/layer-N`). CLI wiring and tests live in `internal/cli/build.go` and `internal/cli/build_integration_test.go`.

- [X] Add container cleanup on build completion/failure
  - Implemented via `StopContainer` and `RemoveContainer` in `internal/container/docker/docker.go` and enforced in the build flow (`internal/cli/build.go` defer cleanup). Integration tests validate cleanup behavior.

## Phase 5: Build Execution

- [X] Implement bitbake command execution in containers
- [X] Add real-time log streaming from container
- [X] Handle build process signals (interrupt, kill)
- [X] Implement build state tracking
- [X] Add error detection and reporting
- [X] Create build timeout handling

## Phase 6: Artifact Management

- [x] Implement artifact extraction from containers *(done: robust extraction and copy logic, supports symlinks and all file types)*
- [x] Create artifact storage directory structure *(done: customer/image/timestamp scoped, nested support)*
- [x] Add artifact metadata tracking (build time, size, config used) *(done: metadata.json written for each build)*
- [x] Implement artifact listing functionality *(done: CLI lists all nested files, supports --customer)*
- [x] Add artifact cleanup/retention policies *(done: CLI clean command, retention by count/age)*
- [x] Create artifact download/copy utilities *(done: CLI copy command, supports nested and symlinks)*

## Phase 7: CLI Development

- [x] Set up CLI framework (cobra or similar) *(done: cobra used for all commands)*
- [x] Implement `smidr init` command *(done: see internal/cli/init.go)*
- [x] Implement `smidr build` command *(done: see internal/cli/build.go)*
- [x] Implement `smidr status` command *(done: full status logic, artifact summary, and --list-artifacts flag)*
- [x] Implement `smidr logs` command *(done: CLI log viewing, supports --customer, buildID, image)*
- [x] Implement `smidr artifacts` command *(done: see internal/cli/artifacts.go)*
- [x] Add global flags (verbose, config path, etc.) *(done: see internal/cli/root.go)*
- [X] Write detailed help text for each CLI command (init, build, status, logs, artifacts)
- [X] Add usage examples for common workflows (init, build, status, logs, artifacts)
- [X] Ensure --help output is clear and includes flag explanations and sample invocations

## Phase 8: Testing & Validation

- [x] Write unit tests for all core packages (artifacts, bitbake, cli, config, container, source)
- [x] Ensure >60% code coverage for core logic (current total ~66% as of 2025-10-16)
- [x] Create integration tests for CLI workflows (init, status, logs, artifacts)
- [x] Add entrypoint smoke test for main.go
- [X] Run integration tests with real Yocto builds (end-to-end)
- [X] Test with Toradex layers specifically (integration and artifact extraction)
- [ ] Test with multiple custom Yocto layer combinations
  - 🚧 In progress: Created test configs for Raspberry Pi, Intel BSP, Toradex
  - 🚧 Matrix CI workflow created to test all combinations
  - ⏳ Pending: Run and validate matrix builds
- [ ] Validate disk space savings vs traditional Yocto build approach
- [ ] Benchmark build times for typical and large builds
- [ ] Test error handling and recovery (simulate build/container failures)
- [X] Add automated test runs to CI (GitHub Actions)
- [ ] Generate and publish test coverage reports
- [ ] Document test strategy and how to run tests locally

## Phase 9: Documentation

- [ ] Write installation instructions
- [ ] Create quick start guide
- [ ] Document configuration file format
- [ ] Add CLI command reference
- [ ] Create troubleshooting guide
- [ ] Write architecture overview
- [ ] Add example configurations

## Phase 10: Polish & Release

- [ ] Add proper logging throughout application
- [ ] Implement progress indicators for long operations
- [ ] Add configuration validation with helpful error messages
- [ ] Create release scripts/automation
- [ ] Tag v0.1.0 MVP release
- [ ] Publish to GitHub releases
- [ ] Create announcement/blog post

## Future Enhancements (Post-MVP)

- [ ] Web UI for build management
- [ ] Advanced caching strategies (deduplication, compression)
- [ ] CI/CD integration for automated builds
- [ ] Support for additional BSPs and vendors
- [ ] Cloud-based build service option
- [ ] Support for multiple build configurations
- [ ] User authentication and access control
- [ ] Integration with external artifact repositories (e.g., Artifactory)
- [ ] Notifications (email, Slack) on build completion/failure
- [ ] Plugin system for extending functionality
- [ ] Localization and internationalization
- [ ] Performance optimizations based on usage patterns
- [ ] Support for alternative container runtimes (e.g., Podman)
- [ ] Enhanced error reporting with actionable suggestions
- [ ] Support for custom build scripts/hooks

## Look into

1. Should we offer the ability in the config to do shared or per-build sstate-cache directories?
2. Should we offer the ability in the config to do shared or per-build downloads directories?

## Optional things for elixir

Trim logger filters: If logs are quiet now, we can remove or guard the translator and Erlang primary filter behind dev-only config to keep production signal clean.
Fast health check: Add a tiny “ping” helper that does a cheap RPC (or checks channel state) for readiness probes and admin dashboards.
Resilience polish: Add a modest exponential backoff on reconnect and a couple of telemetry events so we can observe reconnects without noise.
Docs: Note SMIDR_CLIENT_TYPE=grpc and the dns:// normalization, plus a short “why one channel” rationale for future us.
Test: Add an integration test asserting the client reuses the same channel across calls (prevents regressions).
