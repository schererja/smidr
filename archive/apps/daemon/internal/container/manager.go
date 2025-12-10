package container

import (
	"context"
	"time"
)

// ContainerConfig holds configuration for creating a container
type ContainerConfig struct {
	Image                string
	Name                 string
	Env                  []string
	Cmd                  []string
	Entrypoint           []string
	Mounts               []Mount
	DownloadsDir         string   // Host path to mount as /home/builder/downloads
	SstateCacheDir       string   // Host path to mount as /home/builder/sstate-cache
	BuildDir             string   // Host path to mount as /home/builder/build (persistent Yocto build dir)
	WorkspaceDir         string   // Host path to mount as /home/builder/work (main workspace)
	WorkspaceMountTarget string   // Container path where WorkspaceDir/BuildDir should be mounted (defaults to /home/builder/build if empty)
	LayerDirs            []string // Host paths to Yocto meta-layers to inject into /home/builder/layers
	LayerNames           []string // Names corresponding to LayerDirs for proper mounting
	MemoryLimit          string   `yaml:"memory"`    // e.g. "2g"
	CPUCount             int      `yaml:"cpu_count"` // Number of CPUs to allocate
	TmpDir               string   // Host path to mount as /home/builder/tmp
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

// ContainerManagerStreamer is an optional extension that supports line-by-line streaming callbacks.
// Implementations may choose to provide this for real-time processing without buffering.
type ContainerManagerStreamer interface {
	ExecStreamLines(ctx context.Context, containerID string, cmd []string, timeout time.Duration, onStdout func(string), onStderr func(string)) (ExecResult, error)
}

// NewContainerConfig creates a new container configuration with defaults
func NewContainerConfig(image, name string) ContainerConfig {
	return ContainerConfig{
		Image: image,
		Name:  name,
		Env:   make([]string, 0),
		Cmd:   make([]string, 0),
	}
}

// WithEnv adds environment variables to the container config
func (c ContainerConfig) WithEnv(env ...string) ContainerConfig {
	c.Env = append(c.Env, env...)
	return c
}

// WithCmd sets the command for the container
func (c ContainerConfig) WithCmd(cmd ...string) ContainerConfig {
	c.Cmd = cmd
	return c
}

// WithMemoryLimit sets the memory limit for the container
func (c ContainerConfig) WithMemoryLimit(limit string) ContainerConfig {
	c.MemoryLimit = limit
	return c
}

// WithCPUCount sets the CPU count for the container
func (c ContainerConfig) WithCPUCount(count int) ContainerConfig {
	c.CPUCount = count
	return c
}

// WithDownloadsDir sets the downloads directory mount
func (c ContainerConfig) WithDownloadsDir(dir string) ContainerConfig {
	c.DownloadsDir = dir
	return c
}

// WithWorkspaceDir sets the workspace directory mount
func (c ContainerConfig) WithWorkspaceDir(dir string) ContainerConfig {
	c.WorkspaceDir = dir
	return c
}

// AddLayer adds a layer directory to the container config
func (c ContainerConfig) AddLayer(hostPath, name string) ContainerConfig {
	c.LayerDirs = append(c.LayerDirs, hostPath)
	c.LayerNames = append(c.LayerNames, name)
	return c
}

// IsSuccess checks if an ExecResult indicates success
func (r ExecResult) IsSuccess() bool {
	return r.ExitCode == 0
}

// GetStdoutString returns stdout as a string
func (r ExecResult) GetStdoutString() string {
	return string(r.Stdout)
}

// GetStderrString returns stderr as a string
func (r ExecResult) GetStderrString() string {
	return string(r.Stderr)
}
