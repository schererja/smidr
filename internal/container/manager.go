package container

import (
	"context"
	"time"
)

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
	MemoryLimit    string   `yaml:"memory"`    // e.g. "2g"
	CPUCount       int      `yaml:"cpu_count"` // Number of CPUs to allocate
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
	PullImage(ctx context.Context, image string) error
	CreateContainer(ctx context.Context, cfg ContainerConfig) (containerID string, err error)
	StartContainer(ctx context.Context, containerID string) error
	StopContainer(ctx context.Context, containerID string, timeout time.Duration) error
	RemoveContainer(ctx context.Context, containerID string, force bool) error
	Exec(ctx context.Context, containerID string, cmd []string, timeout time.Duration) (ExecResult, error)
	ExecStream(ctx context.Context, containerID string, cmd []string, timeout time.Duration) (ExecResult, error)
	ImageExists(ctx context.Context, imageName string) bool
	CopyFromContainer(ctx context.Context, containerID, containerPath, hostPath string) error
}
