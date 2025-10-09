# Smidr MVP - TODO List

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

- [ ] Implement bitbake command execution in containers
- [ ] Add real-time log streaming from container
- [ ] Handle build process signals (interrupt, kill)
- [ ] Implement build state tracking
- [ ] Add error detection and reporting
- [ ] Create build timeout handling

## Phase 6: Artifact Management

- [ ] Implement artifact extraction from containers
- [ ] Create artifact storage directory structure
- [ ] Add artifact metadata tracking (build time, size, config used)
- [ ] Implement artifact listing functionality
- [ ] Add artifact cleanup/retention policies
- [ ] Create artifact download/copy utilities

## Phase 7: CLI Development

- [ ] Set up CLI framework (cobra or similar)
- [ ] Implement `smidr init` command
- [ ] Implement `smidr build` command
- [ ] Implement `smidr status` command
- [ ] Implement `smidr logs` command
- [ ] Implement `smidr artifacts` command
- [ ] Add global flags (verbose, config path, etc.)
- [ ] Create help text and usage examples

## Phase 8: Testing & Validation

- [ ] Write unit tests for core functionality
- [ ] Create integration tests with real Yocto builds
- [ ] Test with Toradex layers specifically
- [ ] Test with multiple custom layer combinations
- [ ] Validate disk space savings vs traditional approach
- [ ] Benchmark build times
- [ ] Test error handling and recovery

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
