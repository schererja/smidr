package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	"github.com/schererja/smidr/pkg/logger"
)

// lineCallbackWriter buffers bytes until newline and invokes a callback per line while mirroring to a buffer
type lineCallbackWriter struct {
	buf    bytes.Buffer
	onLine func(string)
	mirror *bytes.Buffer
}

// writeLine delivers a line to callback and mirror
func writeLine(w *lineCallbackWriter, line string) {
	if w.onLine != nil {
		w.onLine(strings.TrimRight(line, "\r"))
	}
	if w.mirror != nil {
		w.mirror.WriteString(line)
		w.mirror.WriteByte('\n')
	}
}

// Write implements io.Writer and splits input by newlines
func (w *lineCallbackWriter) Write(p []byte) (int, error) {
	n := len(p)
	for len(p) > 0 {
		if i := bytes.IndexByte(p, '\n'); i >= 0 {
			// write up to newline
			w.buf.Write(p[:i])
			line := w.buf.String()
			writeLine(w, line)
			w.buf.Reset()
			p = p[i+1:]
		} else {
			// no newline, buffer remainder
			w.buf.Write(p)
			break
		}
	}
	return n, nil
}

// DockerManager implements container.ContainerManager using the Docker Engine API
// (moby/moby Go client)

type DockerManager struct {
	cli    *client.Client
	logger *logger.Logger
}

// NewDockerManager creates a new DockerManager, returning error if Docker is unavailable
func NewDockerManager(log *logger.Logger) (*DockerManager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Error("failed to create Docker client", err)
		return nil, err
	}
	return &DockerManager{cli: cli, logger: log}, nil
}

func (d *DockerManager) PullImage(ctx context.Context, imageName string) error {
	d.logger.Info("pulling docker image", slog.String("image", imageName))
	out, err := d.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		d.logger.Error("failed to pull image", err, slog.String("image", imageName))
		return err
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
	d.logger.Debug("copying from container",
		slog.String("container_id", containerID),
		slog.String("container_path", containerPath),
		slog.String("host_path", hostPath))

	reader, _, err := d.cli.CopyFromContainer(ctx, containerID, containerPath)
	if err != nil {
		d.logger.Error("failed to copy from container", err,
			slog.String("container_id", containerID),
			slog.String("container_path", containerPath))
		return err
	}
	defer reader.Close()

	// Create a tar reader to extract the file
	tarReader := tar.NewReader(reader)
	_, err = tarReader.Next()
	if err != nil {
		d.logger.Error("failed to read tar header", err)
		return err
	}

	// Open the destination file
	outFile, err := os.Create(hostPath)
	if err != nil {
		d.logger.Error("failed to create destination file", err, slog.String("path", hostPath))
		return err
	}
	defer outFile.Close()

	// Copy the file content
	_, err = io.Copy(outFile, tarReader)
	if err != nil {
		d.logger.Error("failed to copy file content", err)
		return err
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
			d.logger.Warn("could not parse memory limit", slog.String("limit", cfg.MemoryLimit), slog.Any("error", err))
		} else {
			hostConfig.Memory = memBytes
			d.logger.Info("setting container memory limit", slog.String("limit", cfg.MemoryLimit), slog.Int64("bytes", memBytes))
		}
	}

	// Set CPU count if specified
	if cfg.CPUCount > 0 {
		// Get system info to check available CPUs
		info, err := d.cli.Info(context.Background())
		if err != nil {
			// If we can't get system info, proceed with warning
			d.logger.Warn("could not get system info to validate CPU count", slog.Any("error", err))
			hostConfig.NanoCPUs = int64(cfg.CPUCount) * 1_000_000_000 // 1 CPU = 1e9
			d.logger.Info("setting container CPU count (unchecked)", slog.Int("cpus", cfg.CPUCount))
		} else {
			maxCPUs := info.NCPU
			requestedCPUs := cfg.CPUCount

			// Cap the CPU count to available CPUs
			if requestedCPUs > maxCPUs {
				d.logger.Warn("requested CPU count exceeds available CPUs",
					slog.Int("requested", requestedCPUs),
					slog.Int("available", maxCPUs),
					slog.Int("using", maxCPUs))
				requestedCPUs = maxCPUs
			}

			hostConfig.NanoCPUs = int64(requestedCPUs) * 1_000_000_000 // 1 CPU = 1e9
			d.logger.Info("setting container CPU count", slog.Int("cpus", requestedCPUs))
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
			d.logger.Error("failed to create downloads dir", err, slog.String("dir", cfg.DownloadsDir))
			return "", err
		}
		// Try to set ownership to builder user (UID 1000, GID 1000); fallback to chmod 0777 so non-owners can write
		if err := os.Chown(cfg.DownloadsDir, 1000, 1000); err != nil {
			if chmodErr := os.Chmod(cfg.DownloadsDir, 0777); chmodErr != nil {
				d.logger.Warn("could not set writable permissions",
					slog.String("dir", cfg.DownloadsDir),
					slog.Any("chown_error", err),
					slog.Any("chmod_error", chmodErr))
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
			d.logger.Error("failed to create sstate cache dir", err, slog.String("dir", cfg.SstateCacheDir))
			return "", err
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
			d.logger.Error("failed to create build dir", err, slog.String("dir", cfg.BuildDir))
			return "", err
		}
		// Try to set ownership to builder user (UID 1000, GID 1000), fallback to chmod 0777 if chown fails
		chownErr := os.Chown(cfg.BuildDir, 1000, 1000)
		if chownErr != nil {
			// Not root or chown failed, fallback to chmod 0777
			chmodErr := os.Chmod(cfg.BuildDir, 0777)
			if chmodErr != nil {
				d.logger.Warn("could not set permissions on build dir", slog.String("dir", cfg.BuildDir), slog.Any("error", chmodErr))
			}
		}
		// Also ensure deploy subdir is writable
		deployDir := filepath.Join(cfg.BuildDir, "deploy")
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
			d.logger.Error("failed to create tmp dir", err, slog.String("dir", cfg.TmpDir))
			return "", err
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
				d.logger.Error("failed to create layer dir", err, slog.String("dir", layerDir))
				return "", err
			}
			// Mount each layer using its proper name (or layer-N as fallback)
			var target string
			if i < len(cfg.LayerNames) && cfg.LayerNames[i] != "" {
				target = "/home/builder/layers/" + cfg.LayerNames[i]
			} else {
				target = "/home/builder/layers/layer-" + strconv.Itoa(i)
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
		d.logger.Error("failed to create container", err, slog.String("name", cfg.Name), slog.String("image", cfg.Image))
		return "", err
	}
	d.logger.Info("container created", slog.String("container_id", resp.ID), slog.String("name", cfg.Name))
	return resp.ID, nil
}

func (d *DockerManager) StartContainer(ctx context.Context, containerID string) error {
	d.logger.Info("starting container", slog.String("container_id", containerID))
	if err := d.cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		d.logger.Error("failed to start container", err, slog.String("container_id", containerID))
		return err
	}

	// Create writable workspace for container operations
	// NOTE: All /home/builder/* paths may be bind-mounted and not writable by container user
	// Use /tmp for workspace that we know is writable, and create a symlink if needed
	userCheckRes, err := d.Exec(ctx, containerID, []string{"sh", "-c", "id builder >/dev/null 2>&1"}, 5*time.Second)
	if err != nil {
		d.logger.Warn("could not check for builder user", slog.Any("error", err))
	} else if userCheckRes.ExitCode == 0 {
		// Builder user exists, create workspace in writable location
		setupRes, err := d.Exec(ctx, containerID, []string{"sh", "-c", "mkdir -p /tmp/builder-workspace && chown builder:builder /tmp/builder-workspace"}, 10*time.Second)
		if err != nil {
			d.logger.Warn("setup command exec error", slog.Any("error", err))
		}
		if setupRes.ExitCode != 0 {
			d.logger.Warn("setup command failed",
				slog.Int("exit_code", setupRes.ExitCode),
				slog.String("stdout", string(setupRes.Stdout)),
				slog.String("stderr", string(setupRes.Stderr)))
		}
	} else {
		// No builder user, just ensure basic workspace exists in writable location
		setupRes, err := d.Exec(ctx, containerID, []string{"sh", "-c", "mkdir -p /tmp/workspace"}, 5*time.Second)
		if err != nil {
			d.logger.Warn("basic workspace setup error", slog.Any("error", err))
		}
		if setupRes.ExitCode != 0 {
			d.logger.Warn("basic workspace setup failed", slog.String("stderr", string(setupRes.Stderr)))
		}
	}

	// Ensure the bind-mounted /home/builder/tmp exists and is writable (sticky tmp semantics)
	// This helps BitBake create HOSTTOOLS under TMPDIR without PermissionError.
	tmpSetup := []string{"sh", "-c", "mkdir -p /home/builder/tmp && chmod 1777 /home/builder/tmp || chmod 0777 /home/builder/tmp"}
	if res, err := d.Exec(ctx, containerID, tmpSetup, 5*time.Second); err != nil {
		d.logger.Warn("could not ensure /home/builder/tmp permissions", slog.Any("error", err))
	} else if res.ExitCode != 0 {
		d.logger.Warn("ensuring /home/builder/tmp perms failed",
			slog.String("stdout", string(res.Stdout)),
			slog.String("stderr", string(res.Stderr)))
	}

	return nil
}

func (d *DockerManager) StopContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	d.logger.Info("stopping container", slog.String("container_id", containerID), slog.Duration("timeout", timeout))
	opts := container.StopOptions{}
	if timeout > 0 {
		secs := int(timeout.Seconds())
		opts.Timeout = &secs
	}
	if err := d.cli.ContainerStop(ctx, containerID, opts); err != nil {
		d.logger.Error("failed to stop container", err, slog.String("container_id", containerID))
		return err
	}
	return nil
}

func (d *DockerManager) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	d.logger.Debug("removing container", slog.String("container_id", containerID), slog.Bool("force", force))
	opts := container.RemoveOptions{Force: force}
	if err := d.cli.ContainerRemove(ctx, containerID, opts); err != nil {
		d.logger.Error("failed to remove container", err, slog.String("container_id", containerID))
		return err
	}
	return nil
}

func (d *DockerManager) Exec(ctx context.Context, containerID string, cmd []string, timeout time.Duration) (smidrContainer.ExecResult, error) {
	d.logger.Debug("executing command in container",
		slog.String("container_id", containerID),
		slog.Any("cmd", cmd),
		slog.Duration("timeout", timeout))

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}
	execIDResp, err := d.cli.ContainerExecCreate(execCtx, containerID, execConfig)
	if err != nil {
		d.logger.Error("failed to create exec", err, slog.String("container_id", containerID))
		return smidrContainer.ExecResult{}, err
	}
	resp, err := d.cli.ContainerExecAttach(execCtx, execIDResp.ID, container.ExecAttachOptions{})
	if err != nil {
		d.logger.Error("failed to attach to exec", err, slog.String("exec_id", execIDResp.ID))
		return smidrContainer.ExecResult{}, err
	}
	defer resp.Close()
	outBytes, errBytes := new(bytes.Buffer), new(bytes.Buffer)
	_, err = stdcopy.StdCopy(outBytes, errBytes, resp.Reader)
	if err != nil {
		d.logger.Error("failed to copy exec output", err)
		return smidrContainer.ExecResult{}, err
	}
	inspect, err := d.cli.ContainerExecInspect(execCtx, execIDResp.ID)
	if err != nil {
		d.logger.Error("failed to inspect exec", err, slog.String("exec_id", execIDResp.ID))
		return smidrContainer.ExecResult{}, err
	}
	return smidrContainer.ExecResult{
		Stdout:   outBytes.Bytes(),
		Stderr:   errBytes.Bytes(),
		ExitCode: inspect.ExitCode,
	}, nil
}

// ExecStream runs a command in the container with real-time output streaming to stdout/stderr
func (d *DockerManager) ExecStream(ctx context.Context, containerID string, cmd []string, timeout time.Duration) (smidrContainer.ExecResult, error) {
	d.logger.Debug("executing command in container with streaming",
		slog.String("container_id", containerID),
		slog.Any("cmd", cmd),
		slog.Duration("timeout", timeout))

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	execIDResp, err := d.cli.ContainerExecCreate(execCtx, containerID, execConfig)
	if err != nil {
		d.logger.Error("failed to create exec for streaming", err, slog.String("container_id", containerID))
		return smidrContainer.ExecResult{}, err
	}

	resp, err := d.cli.ContainerExecAttach(execCtx, execIDResp.ID, container.ExecAttachOptions{})
	if err != nil {
		d.logger.Error("failed to attach to exec for streaming", err, slog.String("exec_id", execIDResp.ID))
		return smidrContainer.ExecResult{}, err
	}
	defer resp.Close()

	// Stream output to stdout/stderr in real-time while also collecting it
	outBytes, errBytes := new(bytes.Buffer), new(bytes.Buffer)
	outWriter := io.MultiWriter(os.Stdout, outBytes)
	errWriter := io.MultiWriter(os.Stderr, errBytes)

	_, err = stdcopy.StdCopy(outWriter, errWriter, resp.Reader)
	if err != nil {
		d.logger.Error("failed to copy streaming exec output", err)
		return smidrContainer.ExecResult{}, err
	}

	inspect, err := d.cli.ContainerExecInspect(execCtx, execIDResp.ID)
	if err != nil {
		d.logger.Error("failed to inspect streaming exec", err, slog.String("exec_id", execIDResp.ID))
		return smidrContainer.ExecResult{}, err
	}

	return smidrContainer.ExecResult{
		Stdout:   outBytes.Bytes(),
		Stderr:   errBytes.Bytes(),
		ExitCode: inspect.ExitCode,
	}, nil
}

// ExecStreamLines runs a command in the container and invokes callbacks for each stdout/stderr line in real-time.
// It also collects the full output and returns it in ExecResult.
func (d *DockerManager) ExecStreamLines(ctx context.Context, containerID string, cmd []string, timeout time.Duration, onStdout func(string), onStderr func(string)) (smidrContainer.ExecResult, error) {
	d.logger.Debug("executing command in container with line streaming",
		slog.String("container_id", containerID),
		slog.Any("cmd", cmd),
		slog.Duration("timeout", timeout))

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	execIDResp, err := d.cli.ContainerExecCreate(execCtx, containerID, execConfig)
	if err != nil {
		d.logger.Error("failed to create exec for line streaming", err, slog.String("container_id", containerID))
		return smidrContainer.ExecResult{}, err
	}

	resp, err := d.cli.ContainerExecAttach(execCtx, execIDResp.ID, container.ExecAttachOptions{})
	if err != nil {
		d.logger.Error("failed to attach to exec for line streaming", err, slog.String("exec_id", execIDResp.ID))
		return smidrContainer.ExecResult{}, err
	}
	defer resp.Close()

	// Prepare writers for stdout and stderr
	outBytes, errBytes := new(bytes.Buffer), new(bytes.Buffer)
	stdoutLW := &lineCallbackWriter{onLine: onStdout, mirror: outBytes}
	stderrLW := &lineCallbackWriter{onLine: onStderr, mirror: errBytes}

	// Demultiplex Docker's multiplexed stream to our writers
	if _, err := stdcopy.StdCopy(stdoutLW, stderrLW, resp.Reader); err != nil {
		d.logger.Error("failed to copy line streaming exec output", err)
		return smidrContainer.ExecResult{}, err
	}

	// Flush any remaining partial lines
	if stdoutLW.buf.Len() > 0 {
		writeLine(stdoutLW, stdoutLW.buf.String())
		stdoutLW.buf.Reset()
	}
	if stderrLW.buf.Len() > 0 {
		writeLine(stderrLW, stderrLW.buf.String())
		stderrLW.buf.Reset()
	}

	inspect, err := d.cli.ContainerExecInspect(execCtx, execIDResp.ID)
	if err != nil {
		d.logger.Error("failed to inspect line streaming exec", err, slog.String("exec_id", execIDResp.ID))
		return smidrContainer.ExecResult{}, err
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
