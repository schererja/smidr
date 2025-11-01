package docker

import (
	"archive/tar"
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
	"github.com/docker/go-units"

	smidrContainer "github.com/schererja/smidr/internal/container"
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

func (d *DockerManager) ImageExists(ctx context.Context, imageName string) bool {
	_, err := d.cli.ImageInspect(ctx, imageName)
	return err == nil
}

func (d *DockerManager) CopyFromContainer(ctx context.Context, containerID, containerPath, hostPath string) error {
	reader, _, err := d.cli.CopyFromContainer(ctx, containerID, containerPath)
	if err != nil {
		return fmt.Errorf("copy from container %s:%s: %w", containerID, containerPath, err)
	}
	defer reader.Close()

	// Create a tar reader to extract the file
	tarReader := tar.NewReader(reader)
	_, err = tarReader.Next()
	if err != nil {
		return fmt.Errorf("read tar header: %w", err)
	}

	// Open the destination file
	outFile, err := os.Create(hostPath)
	if err != nil {
		return fmt.Errorf("create destination file %s: %w", hostPath, err)
	}
	defer outFile.Close()

	// Copy the file content
	_, err = io.Copy(outFile, tarReader)
	if err != nil {
		return fmt.Errorf("copy file content: %w", err)
	}

	return nil
}

func (d *DockerManager) CreateContainer(ctx context.Context, cfg smidrContainer.ContainerConfig) (string, error) {
	mounts := []mount.Mount{}
	hostConfig := &container.HostConfig{}

	// Set memory limit if specified
	if cfg.MemoryLimit != "" {
		memBytes, err := units.RAMInBytes(cfg.MemoryLimit)
		if err != nil {
			fmt.Printf("[WARN] Could not parse memory limit %q: %v\n", cfg.MemoryLimit, err)
		} else {
			hostConfig.Memory = memBytes
			fmt.Printf("[INFO] Setting container memory limit: %s (%d bytes)\n", cfg.MemoryLimit, memBytes)
		}
	}

	// Set CPU count if specified
	if cfg.CPUCount > 0 {
		// Get system info to check available CPUs
		info, err := d.cli.Info(context.Background())
		if err != nil {
			// If we can't get system info, proceed with warning
			fmt.Printf("[WARNING] Could not get system info to validate CPU count: %v\n", err)
			hostConfig.NanoCPUs = int64(cfg.CPUCount) * 1_000_000_000 // 1 CPU = 1e9
			fmt.Printf("[INFO] Setting container CPU count: %d (unchecked)\n", cfg.CPUCount)
		} else {
			maxCPUs := info.NCPU
			requestedCPUs := cfg.CPUCount

			// Cap the CPU count to available CPUs
			if requestedCPUs > maxCPUs {
				fmt.Printf("[WARNING] Requested CPU count (%d) exceeds available CPUs (%d), using %d CPUs\n",
					requestedCPUs, maxCPUs, maxCPUs)
				requestedCPUs = maxCPUs
			}

			hostConfig.NanoCPUs = int64(requestedCPUs) * 1_000_000_000 // 1 CPU = 1e9
			fmt.Printf("[INFO] Setting container CPU count: %d\n", requestedCPUs)
		}
	}
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
		// Try to set ownership to builder user (UID 1000, GID 1000); fallback to chmod 0777 so non-owners can write
		if err := os.Chown(cfg.DownloadsDir, 1000, 1000); err != nil {
			if chmodErr := os.Chmod(cfg.DownloadsDir, 0777); chmodErr != nil {
				fmt.Printf("[WARN] Could not set writable permissions on %s: chown err=%v chmod err=%v\n", cfg.DownloadsDir, err, chmodErr)
			}
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

	// Add build directory mount if specified
	if cfg.BuildDir != "" {
		if err := os.MkdirAll(cfg.BuildDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create build dir %s: %w", cfg.BuildDir, err)
		}
		// Try to set ownership to builder user (UID 1000, GID 1000), fallback to chmod 0777 if chown fails
		chownErr := os.Chown(cfg.BuildDir, 1000, 1000)
		if chownErr != nil {
			// Not root or chown failed, fallback to chmod 0777
			chmodErr := os.Chmod(cfg.BuildDir, 0777)
			if chmodErr != nil {
				fmt.Printf("[WARN] Could not set permissions on %s: %v\n", cfg.BuildDir, chmodErr)
			}
		}
		// Also ensure deploy subdir is writable
		deployDir := fmt.Sprintf("%s/deploy", cfg.BuildDir)
		if err := os.MkdirAll(deployDir, 0755); err == nil {
			if err := os.Chown(deployDir, 1000, 1000); err != nil {
				_ = os.Chmod(deployDir, 0777)
			}
		}
		// Use unique workspace mount target if specified, otherwise default to /home/builder/build
		workspaceTarget := cfg.WorkspaceMountTarget
		if workspaceTarget == "" {
			workspaceTarget = "/home/builder/build"
		}
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: cfg.BuildDir,
			Target: workspaceTarget,
		})
	}

	// Add tmp directory mount if specified
	if cfg.TmpDir != "" {
		if err := os.MkdirAll(cfg.TmpDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create tmp dir %s: %w", cfg.TmpDir, err)
		}
		// Ensure tmp is writable by container user; use 1777 (sticky bit) if possible, else 0777
		if err := os.Chown(cfg.TmpDir, 1000, 1000); err != nil {
			// Not owner or not permitted; still try to set broad perms
			if chmodErr := os.Chmod(cfg.TmpDir, 01777); chmodErr != nil {
				// Fallback to 0777 if sticky not supported
				_ = os.Chmod(cfg.TmpDir, 0777)
			}
		} else {
			// If chown succeeded, set sticky perms as best practice for tmp directories
			_ = os.Chmod(cfg.TmpDir, 01777)
		}
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: cfg.TmpDir,
			Target: "/home/builder/tmp",
		})
	}

	// Add layer directories (meta-layers) if specified
	for i, layerDir := range cfg.LayerDirs {
		if layerDir != "" {
			// Ensure layer dir exists on host so bind mount succeeds (read-only)
			if err := os.MkdirAll(layerDir, 0755); err != nil {
				return "", fmt.Errorf("failed to create layer dir %s: %w", layerDir, err)
			}
			// Mount each layer using its proper name (or layer-N as fallback)
			var target string
			if i < len(cfg.LayerNames) && cfg.LayerNames[i] != "" {
				target = fmt.Sprintf("/home/builder/layers/%s", cfg.LayerNames[i])
			} else {
				target = fmt.Sprintf("/home/builder/layers/layer-%d", i)
			}
			mounts = append(mounts, mount.Mount{
				Type:     mount.TypeBind,
				Source:   layerDir,
				Target:   target,
				ReadOnly: true, // Layers are typically read-only during builds
			})
		}
	}

	// Determine Entrypoint/Cmd to use when creating the container.
	// If the caller specified an explicit Entrypoint, use it. If not, only
	// default to "/bin/sh -c" when the provided Cmd is a single string
	// (shell form). If the Cmd is a multi-element slice, we leave Entrypoint
	// nil so the image's own ENTRYPOINT will be used and Cmd will be applied
	// as the argv (exec form).
	var entrypoint strslice.StrSlice
	var cmd strslice.StrSlice
	if len(cfg.Entrypoint) > 0 {
		entrypoint = strslice.StrSlice(cfg.Entrypoint)
		cmd = strslice.StrSlice(cfg.Cmd)
	} else if len(cfg.Cmd) == 1 {
		// single-string command -> execute via shell
		entrypoint = strslice.StrSlice{"/bin/sh", "-c"}
		cmd = strslice.StrSlice{cfg.Cmd[0]}
	} else {
		// multi-element Cmd -> use image ENTRYPOINT and pass Cmd as argv
		entrypoint = nil
		cmd = strslice.StrSlice(cfg.Cmd)
	}
	// Remove old CpusetCpus logic (now handled by NanoCPUs)
	hostConfig.Mounts = mounts

	resp, err := d.cli.ContainerCreate(ctx, &container.Config{
		Image:      cfg.Image,
		Env:        cfg.Env,
		Entrypoint: entrypoint,
		Cmd:        cmd,
	}, hostConfig, &network.NetworkingConfig{}, nil, cfg.Name)
	if err != nil {
		return "", fmt.Errorf("create container: %w", err)
	}
	return resp.ID, nil
}

func (d *DockerManager) StartContainer(ctx context.Context, containerID string) error {
	if err := d.cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("start container %s: %w", containerID, err)
	}

	// Create writable workspace for container operations
	// NOTE: All /home/builder/* paths may be bind-mounted and not writable by container user
	// Use /tmp for workspace that we know is writable, and create a symlink if needed
	userCheckRes, err := d.Exec(ctx, containerID, []string{"sh", "-c", "id builder >/dev/null 2>&1"}, 5*time.Second)
	if err != nil {
		fmt.Printf("⚠️  Could not check for builder user: %v\n", err)
	} else if userCheckRes.ExitCode == 0 {
		// Builder user exists, create workspace in writable location
		setupRes, err := d.Exec(ctx, containerID, []string{"sh", "-c", "mkdir -p /tmp/builder-workspace && chown builder:builder /tmp/builder-workspace"}, 10*time.Second)
		if err != nil {
			fmt.Printf("⚠️  Setup command exec error: %v\n", err)
		}
		if setupRes.ExitCode != 0 {
			fmt.Printf("⚠️  Setup command failed (exit %d): stdout=%s stderr=%s\n", setupRes.ExitCode, string(setupRes.Stdout), string(setupRes.Stderr))
		}
	} else {
		// No builder user, just ensure basic workspace exists in writable location
		setupRes, err := d.Exec(ctx, containerID, []string{"sh", "-c", "mkdir -p /tmp/workspace"}, 5*time.Second)
		if err != nil {
			fmt.Printf("⚠️  Basic workspace setup error: %v\n", err)
		}
		if setupRes.ExitCode != 0 {
			fmt.Printf("⚠️  Basic workspace setup failed: %s\n", string(setupRes.Stderr))
		}
	}

	// Ensure the bind-mounted /home/builder/tmp exists and is writable (sticky tmp semantics)
	// This helps BitBake create HOSTTOOLS under TMPDIR without PermissionError.
	tmpSetup := []string{"sh", "-c", "mkdir -p /home/builder/tmp && chmod 1777 /home/builder/tmp || chmod 0777 /home/builder/tmp"}
	if res, err := d.Exec(ctx, containerID, tmpSetup, 5*time.Second); err != nil {
		fmt.Printf("⚠️  Could not ensure /home/builder/tmp permissions: %v\n", err)
	} else if res.ExitCode != 0 {
		fmt.Printf("⚠️  Ensuring /home/builder/tmp perms failed: stdout=%s stderr=%s\n", string(res.Stdout), string(res.Stderr))
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

// ExecStream runs a command in the container with real-time output streaming to stdout/stderr
func (d *DockerManager) ExecStream(ctx context.Context, containerID string, cmd []string, timeout time.Duration) (smidrContainer.ExecResult, error) {
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

	// Stream output to stdout/stderr in real-time while also collecting it
	outBytes, errBytes := new(bytes.Buffer), new(bytes.Buffer)
	outWriter := io.MultiWriter(os.Stdout, outBytes)
	errWriter := io.MultiWriter(os.Stderr, errBytes)

	_, err = stdcopy.StdCopy(outWriter, errWriter, resp.Reader)
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

func parseMemory(mem string) int64 {
	bytes, err := units.RAMInBytes(mem)
	if err != nil {
		// fallback or log error
		return 0
	}
	return bytes
}
