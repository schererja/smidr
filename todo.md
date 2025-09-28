# Smidr MVP - TODO List

## Phase 1: Project Setup & Foundation

- [X] Initialize Go project structure
- [X] Set up Git repository under intrik8-labs
- [ ] Create basic README with project vision
- [ ] Define project directory structure (cmd/, internal/, pkg/)
- [ ] Set up Go modules and dependencies
- [ ] Create .gitignore for Go projects

## Phase 2: Configuration System

- [ ] Design YAML configuration schema
- [ ] Implement config file parser (using go-yaml)
- [ ] Add config validation logic
- [ ] Create `smidr init` command to generate template config
- [ ] Add support for environment variable substitution
- [ ] Write unit tests for config parsing

## Phase 3: Source Management

- [ ] Implement git repository cloning for layers
- [ ] Add caching logic for git repositories
- [ ] Implement source package download system
- [ ] Create persistent cache directory structure
- [ ] Add cache invalidation/update logic
- [ ] Handle download mirrors and fallbacks

## Phase 4: Container Orchestration

- [ ] Research Docker/Podman Go SDK integration
- [ ] Create base container image with Yocto dependencies
- [ ] Implement container lifecycle management (create/start/stop/destroy)
- [ ] Add volume mounting for downloads and sstate-cache
- [ ] Implement layer injection into containers
- [ ] Add container cleanup on build completion/failure

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
