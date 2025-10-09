package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	smidrContainer "github.com/intrik8-labs/smidr/internal/container"
)

// DockerManager implements container.ContainerManager using the Docker Engine API
// (moby/moby Go client)

type DockerManager struct {
	cli *client.Client
}

// NewDockerManager creates a new DockerManager, returning error if Docker is unavailable
func NewDockerManager() (*DockerManager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	return &DockerManager{cli: cli}, nil
}

func (d *DockerManager) PullImage(ctx context.Context, imageName string) error {
	out, err := d.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("pull image %q: %w", imageName, err)
	}
	defer out.Close()
	io.Copy(io.Discard, out)
	return nil
}

func (d *DockerManager) CreateContainer(ctx context.Context, cfg smidrContainer.ContainerConfig) (string, error) {
	mounts := []mount.Mount{}
	for _, m := range cfg.Mounts {
		mounts = append(mounts, mount.Mount{
			Type:     mount.TypeBind,
			Source:   m.Source,
			Target:   m.Target,
			ReadOnly: m.ReadOnly,
		})
	}

	// Add downloads directory mount if specified
	if cfg.DownloadsDir != "" {
		// Ensure the host directory exists so Docker can bind mount it
		if err := os.MkdirAll(cfg.DownloadsDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create downloads dir %s: %w", cfg.DownloadsDir, err)
		}
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: cfg.DownloadsDir,
			Target: "/home/builder/downloads",
		})
	}

	// Add sstate-cache directory mount if specified
	if cfg.SstateCacheDir != "" {
		if err := os.MkdirAll(cfg.SstateCacheDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create sstate cache dir %s: %w", cfg.SstateCacheDir, err)
		}
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: cfg.SstateCacheDir,
			Target: "/home/builder/sstate-cache",
		})
	}

	// Add workspace directory mount if specified
	if cfg.WorkspaceDir != "" {
		if err := os.MkdirAll(cfg.WorkspaceDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create workspace dir %s: %w", cfg.WorkspaceDir, err)
		}
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: cfg.WorkspaceDir,
			Target: "/home/builder/work",
		})
	}

	// Add layer directories (meta-layers) if specified
	for i, layerDir := range cfg.LayerDirs {
		if layerDir != "" {
			// Ensure layer dir exists on host so bind mount succeeds (read-only)
			if err := os.MkdirAll(layerDir, 0755); err != nil {
				return "", fmt.Errorf("failed to create layer dir %s: %w", layerDir, err)
			}
			// Mount each layer to /home/builder/layers/layer-N for easy access
			target := fmt.Sprintf("/home/builder/layers/layer-%d", i)
			mounts = append(mounts, mount.Mount{
				Type:     mount.TypeBind,
				Source:   layerDir,
				Target:   target,
				ReadOnly: true, // Layers are typically read-only during builds
			})
		}
	}

	// Ensure commands passed as a single string are executed via a POSIX shell
	// inside the container. Setting Entrypoint to /bin/sh -c is safe for most
	// images (busybox, ubuntu, etc.). If an image requires a different entry,
	// this can later be made configurable.
	// Use provided Entrypoint when set, otherwise default to /bin/sh -c so
	// single-string cmds are executed via a POSIX shell inside the container.
	entrypoint := strslice.StrSlice{}
	if len(cfg.Entrypoint) > 0 {
		entrypoint = strslice.StrSlice(cfg.Entrypoint)
	} else {
		entrypoint = strslice.StrSlice{"/bin/sh", "-c"}
	}

	resp, err := d.cli.ContainerCreate(ctx, &container.Config{
		Image:      cfg.Image,
		Env:        cfg.Env,
		Entrypoint: entrypoint,
		Cmd:        strslice.StrSlice(cfg.Cmd),
	}, &container.HostConfig{
		Mounts: mounts,
	}, &network.NetworkingConfig{}, nil, cfg.Name)
	if err != nil {
		return "", fmt.Errorf("create container: %w", err)
	}
	return resp.ID, nil
}

func (d *DockerManager) StartContainer(ctx context.Context, containerID string) error {
	if err := d.cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("start container %s: %w", containerID, err)
	}
	return nil
}

func (d *DockerManager) StopContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	opts := container.StopOptions{}
	if timeout > 0 {
		secs := int(timeout.Seconds())
		opts.Timeout = &secs
	}
	if err := d.cli.ContainerStop(ctx, containerID, opts); err != nil {
		return fmt.Errorf("stop container %s: %w", containerID, err)
	}
	return nil
}

func (d *DockerManager) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	opts := container.RemoveOptions{Force: force}
	if err := d.cli.ContainerRemove(ctx, containerID, opts); err != nil {
		return fmt.Errorf("remove container %s: %w", containerID, err)
	}
	return nil
}

func (d *DockerManager) Exec(ctx context.Context, containerID string, cmd []string, timeout time.Duration) (smidrContainer.ExecResult, error) {
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}
	execIDResp, err := d.cli.ContainerExecCreate(execCtx, containerID, execConfig)
	if err != nil {
		return smidrContainer.ExecResult{}, fmt.Errorf("exec create: %w", err)
	}
	resp, err := d.cli.ContainerExecAttach(execCtx, execIDResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return smidrContainer.ExecResult{}, fmt.Errorf("exec attach: %w", err)
	}
	defer resp.Close()
	outBytes, errBytes := new(bytes.Buffer), new(bytes.Buffer)
	_, err = stdcopy.StdCopy(outBytes, errBytes, resp.Reader)
	if err != nil {
		return smidrContainer.ExecResult{}, fmt.Errorf("exec output copy: %w", err)
	}
	inspect, err := d.cli.ContainerExecInspect(execCtx, execIDResp.ID)
	if err != nil {
		return smidrContainer.ExecResult{}, fmt.Errorf("exec inspect: %w", err)
	}
	return smidrContainer.ExecResult{
		Stdout:   outBytes.Bytes(),
		Stderr:   errBytes.Bytes(),
		ExitCode: inspect.ExitCode,
	}, nil
}
