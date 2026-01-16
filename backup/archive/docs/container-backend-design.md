# Container Backend Design (Phase 4)

Goal: provide a small, testable abstraction for container lifecycle management used by Smidr to run Yocto builds inside containers. Start with one backend (Docker Engine) and keep the implementation pluggable so other backends (Podman, containerd) can be added later.

Requirements

- Create/start/stop/destroy containers
- Exec commands inside containers with streaming logs
- Mount host directories into containers (downloads, sstate-cache, layers)
- Create and manage named volumes
- Timeouts and graceful shutdown
- Proper error surface to caller (wrapped errors, retryable vs fatal)

Options considered

1) Docker Engine (moby/docker client)

- Pros: mature Go client, widespread, working on Linux/macOS/Windows, can run with Docker Desktop or Docker Engine on Linux.
- Cons: requires Docker daemon, which may not be present in all environments. Docker Desktop licensing considerations for larger orgs.

1. Podman (REST API / varlink compatibility)

    - Pros: daemonless mode (rootless), works well in Linux server environments, gaining adoption.
    - Cons: Go SDK less mature; interaction often via CLI or REST socket; portability on macOS requires Podman Desktop.

1. containerd (pure runtime)

    - Pros: lightweight and powerful, used under Docker and container runtimes.
    - Cons: lower-level; building a fully-featured client is more complex.

Recommendation

- Implement an abstraction `ContainerManager` in `internal/container`:
  - Provide a Docker Engine backend first using the official moby client.
  - Design the interface small and well-tested so backends can be swapped later.

Proposed interface

```go
package container

import "time"

type ContainerConfig struct {
    Image string
    Name  string
    Env   []string
    Cmd   []string
    Mounts []Mount
    // Resource limits, etc. can be added later
}

type Mount struct {
    Source string // host path or volume name
    Target string // container path
    ReadOnly bool
}

type ExecResult struct {
    Stdout []byte
    Stderr []byte
    ExitCode int
}

// ContainerManager manages container lifecycle and exec
type ContainerManager interface {
    PullImage(image string) error
    CreateContainer(cfg ContainerConfig) (containerID string, err error)
    StartContainer(containerID string) error
    StopContainer(containerID string, timeout time.Duration) error
    RemoveContainer(containerID string, force bool) error
    Exec(containerID string, cmd []string, timeout time.Duration) (ExecResult, error)
    // Streamed exec/logging variants can be added
}
```

Implementation plan

- Add `internal/container/docker` package with a `NewDockerManager` constructor.
- Implement basic lifecycle operations with retries where applicable (image pull). Use context for timeouts.
- Add unit tests that mock the Docker client where possible (use interfaces); add a small integration test that requires Docker and runs conditionally.

Developer notes

- Keep operations idempotent where reasonable (e.g., CreateContainer returns existing container ID if already created with same name).
- Enforce cleanup in defer paths and provide a force-cleanup command.

## Entrypoint configuration

Smidr now supports configuring the container entrypoint used when creating
containers. This is useful for images that require a specific process to be
launched (for example `/bin/bash -l -c`). The YAML key is `container.entrypoint`
and takes a YAML sequence, for example:

```yaml
container:
    entrypoint: ["/bin/bash","-l","-c"]
```

By default Smidr will use `/bin/sh -c` as a safe fallback so simple single-
string commands (like `echo 'ready'; sleep 5`) execute reliably across images.
Tests and CI can override the entrypoint with the `SMIDR_TEST_ENTRYPOINT`
environment variable (comma-separated) for per-run overrides.

*** End of design doc
