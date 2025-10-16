package container

import (
	"context"
	"testing"
	"time"
)

// DummyContainerManager is a mock implementation for testing
// (You can replace this with a real mock framework if desired)
type DummyContainerManager struct{}

func (d *DummyContainerManager) PullImage(ctx context.Context, image string) error { return nil }
func (d *DummyContainerManager) CreateContainer(ctx context.Context, cfg ContainerConfig) (string, error) {
	return "dummy-id", nil
}
func (d *DummyContainerManager) StartContainer(ctx context.Context, containerID string) error {
	return nil
}
func (d *DummyContainerManager) StopContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	return nil
}
func (d *DummyContainerManager) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	return nil
}
func (d *DummyContainerManager) Exec(ctx context.Context, containerID string, cmd []string, timeout time.Duration) (ExecResult, error) {
	return ExecResult{Stdout: []byte("ok"), Stderr: nil, ExitCode: 0}, nil
}
func (d *DummyContainerManager) ExecStream(ctx context.Context, containerID string, cmd []string, timeout time.Duration) (ExecResult, error) {
	return ExecResult{Stdout: []byte("ok"), Stderr: nil, ExitCode: 0}, nil
}
func (d *DummyContainerManager) ImageExists(ctx context.Context, imageName string) bool {
	return true
}
func (d *DummyContainerManager) CopyFromContainer(ctx context.Context, containerID, containerPath, hostPath string) error {
	return nil
}

func TestContainerManagerInterface(t *testing.T) {
	var cm ContainerManager = &DummyContainerManager{}
	ctx := context.Background()

	if err := cm.PullImage(ctx, "busybox"); err != nil {
		t.Errorf("PullImage failed: %v", err)
	}
	id, err := cm.CreateContainer(ctx, ContainerConfig{Image: "busybox"})
	if err != nil || id == "" {
		t.Errorf("CreateContainer failed: %v, id=%s", err, id)
	}
	if err := cm.StartContainer(ctx, id); err != nil {
		t.Errorf("StartContainer failed: %v", err)
	}
	if err := cm.StopContainer(ctx, id, 2*time.Second); err != nil {
		t.Errorf("StopContainer failed: %v", err)
	}
	if err := cm.RemoveContainer(ctx, id, true); err != nil {
		t.Errorf("RemoveContainer failed: %v", err)
	}
	res, err := cm.Exec(ctx, id, []string{"echo", "hi"}, 1*time.Second)
	if err != nil || res.ExitCode != 0 {
		t.Errorf("Exec failed: %v, result=%+v", err, res)
	}
}

func TestNewContainerConfig(t *testing.T) {
	cfg := NewContainerConfig("busybox:latest", "test-container")
	if cfg.Image != "busybox:latest" {
		t.Errorf("expected image busybox:latest, got %s", cfg.Image)
	}
	if cfg.Name != "test-container" {
		t.Errorf("expected name test-container, got %s", cfg.Name)
	}
	if cfg.Env == nil || cfg.Cmd == nil {
		t.Errorf("expected initialized slices")
	}
}

func TestContainerConfig_WithEnv(t *testing.T) {
	cfg := NewContainerConfig("img", "name").WithEnv("KEY=value", "FOO=bar")
	if len(cfg.Env) != 2 {
		t.Errorf("expected 2 env vars, got %d", len(cfg.Env))
	}
	if cfg.Env[0] != "KEY=value" || cfg.Env[1] != "FOO=bar" {
		t.Errorf("env vars not set correctly: %v", cfg.Env)
	}
}

func TestContainerConfig_WithCmd(t *testing.T) {
	cfg := NewContainerConfig("img", "name").WithCmd("echo", "hello")
	if len(cfg.Cmd) != 2 {
		t.Errorf("expected 2 cmd args, got %d", len(cfg.Cmd))
	}
	if cfg.Cmd[0] != "echo" || cfg.Cmd[1] != "hello" {
		t.Errorf("cmd not set correctly: %v", cfg.Cmd)
	}
}

func TestContainerConfig_WithMemoryLimit(t *testing.T) {
	cfg := NewContainerConfig("img", "name").WithMemoryLimit("2g")
	if cfg.MemoryLimit != "2g" {
		t.Errorf("expected memory limit 2g, got %s", cfg.MemoryLimit)
	}
}

func TestContainerConfig_WithCPUCount(t *testing.T) {
	cfg := NewContainerConfig("img", "name").WithCPUCount(4)
	if cfg.CPUCount != 4 {
		t.Errorf("expected CPU count 4, got %d", cfg.CPUCount)
	}
}

func TestContainerConfig_WithDownloadsDir(t *testing.T) {
	cfg := NewContainerConfig("img", "name").WithDownloadsDir("/tmp/downloads")
	if cfg.DownloadsDir != "/tmp/downloads" {
		t.Errorf("expected downloads dir /tmp/downloads, got %s", cfg.DownloadsDir)
	}
}

func TestContainerConfig_WithWorkspaceDir(t *testing.T) {
	cfg := NewContainerConfig("img", "name").WithWorkspaceDir("/workspace")
	if cfg.WorkspaceDir != "/workspace" {
		t.Errorf("expected workspace dir /workspace, got %s", cfg.WorkspaceDir)
	}
}

func TestContainerConfig_AddLayer(t *testing.T) {
	cfg := NewContainerConfig("img", "name").
		AddLayer("/path/to/layer1", "layer1").
		AddLayer("/path/to/layer2", "layer2")
	if len(cfg.LayerDirs) != 2 || len(cfg.LayerNames) != 2 {
		t.Errorf("expected 2 layers, got %d dirs and %d names", len(cfg.LayerDirs), len(cfg.LayerNames))
	}
	if cfg.LayerDirs[0] != "/path/to/layer1" || cfg.LayerNames[0] != "layer1" {
		t.Errorf("layer 0 not set correctly: %s, %s", cfg.LayerDirs[0], cfg.LayerNames[0])
	}
	if cfg.LayerDirs[1] != "/path/to/layer2" || cfg.LayerNames[1] != "layer2" {
		t.Errorf("layer 1 not set correctly: %s, %s", cfg.LayerDirs[1], cfg.LayerNames[1])
	}
}

func TestContainerConfig_Chaining(t *testing.T) {
	cfg := NewContainerConfig("alpine:latest", "my-container").
		WithEnv("ENV=prod").
		WithCmd("sh", "-c", "echo test").
		WithMemoryLimit("1g").
		WithCPUCount(2).
		WithDownloadsDir("/downloads").
		WithWorkspaceDir("/workspace").
		AddLayer("/layer1", "meta-layer1")

	if cfg.Image != "alpine:latest" {
		t.Errorf("image not set correctly")
	}
	if len(cfg.Env) != 1 || cfg.Env[0] != "ENV=prod" {
		t.Errorf("env not chained correctly")
	}
	if len(cfg.Cmd) != 3 {
		t.Errorf("cmd not chained correctly")
	}
	if cfg.MemoryLimit != "1g" {
		t.Errorf("memory limit not chained correctly")
	}
	if cfg.CPUCount != 2 {
		t.Errorf("CPU count not chained correctly")
	}
	if cfg.DownloadsDir != "/downloads" {
		t.Errorf("downloads dir not chained correctly")
	}
	if cfg.WorkspaceDir != "/workspace" {
		t.Errorf("workspace dir not chained correctly")
	}
	if len(cfg.LayerDirs) != 1 || cfg.LayerDirs[0] != "/layer1" {
		t.Errorf("layer not chained correctly")
	}
}

func TestExecResult_IsSuccess(t *testing.T) {
	successResult := ExecResult{ExitCode: 0}
	if !successResult.IsSuccess() {
		t.Errorf("expected IsSuccess() to return true for exit code 0")
	}

	failResult := ExecResult{ExitCode: 1}
	if failResult.IsSuccess() {
		t.Errorf("expected IsSuccess() to return false for exit code 1")
	}
}

func TestExecResult_GetStdoutString(t *testing.T) {
	result := ExecResult{Stdout: []byte("hello world")}
	if result.GetStdoutString() != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", result.GetStdoutString())
	}
}

func TestExecResult_GetStderrString(t *testing.T) {
	result := ExecResult{Stderr: []byte("error message")}
	if result.GetStderrString() != "error message" {
		t.Errorf("expected 'error message', got '%s'", result.GetStderrString())
	}
}
