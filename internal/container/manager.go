package container

import "time"

// ContainerConfig holds configuration for creating a container
type ContainerConfig struct {
	Image          string
	Name           string
	Env            []string
	Cmd            []string
	Entrypoint     []string
	Mounts         []Mount
	DownloadsDir   string   // Host path to mount as /home/builder/downloads
	SstateCacheDir string   // Host path to mount as /home/builder/sstate-cache
	WorkspaceDir   string   // Host path to mount as /home/builder/work (main workspace)
	LayerDirs      []string // Host paths to Yocto meta-layers to inject into /home/builder/layers
	// Resource limits, etc. can be added later
}

type Mount struct {
	Source   string // host path or volume name
	Target   string // container path
	ReadOnly bool
}

type ExecResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

// ContainerManager manages container lifecycle and exec
// This is the main abstraction for container orchestration backends
// (Docker, Podman, containerd, etc.)
type ContainerManager interface {
	PullImage(image string) error
	CreateContainer(cfg ContainerConfig) (containerID string, err error)
	StartContainer(containerID string) error
	StopContainer(containerID string, timeout time.Duration) error
	RemoveContainer(containerID string, force bool) error
	Exec(containerID string, cmd []string, timeout time.Duration) (ExecResult, error)
	// Streamed exec/logging variants can be added
}
