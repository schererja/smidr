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
