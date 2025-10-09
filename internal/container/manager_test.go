package container

import (
	"testing"
	"time"
)

// DummyContainerManager is a mock implementation for testing
// (You can replace this with a real mock framework if desired)
type DummyContainerManager struct{}

func (d *DummyContainerManager) PullImage(image string) error { return nil }
func (d *DummyContainerManager) CreateContainer(cfg ContainerConfig) (string, error) {
	return "dummy-id", nil
}
func (d *DummyContainerManager) StartContainer(containerID string) error { return nil }
func (d *DummyContainerManager) StopContainer(containerID string, timeout time.Duration) error {
	return nil
}
func (d *DummyContainerManager) RemoveContainer(containerID string, force bool) error { return nil }
func (d *DummyContainerManager) Exec(containerID string, cmd []string, timeout time.Duration) (ExecResult, error) {
	return ExecResult{Stdout: []byte("ok"), Stderr: nil, ExitCode: 0}, nil
}

func TestContainerManagerInterface(t *testing.T) {
	var cm ContainerManager = &DummyContainerManager{}

	if err := cm.PullImage("busybox"); err != nil {
		t.Errorf("PullImage failed: %v", err)
	}
	id, err := cm.CreateContainer(ContainerConfig{Image: "busybox"})
	if err != nil || id == "" {
		t.Errorf("CreateContainer failed: %v, id=%s", err, id)
	}
	if err := cm.StartContainer(id); err != nil {
		t.Errorf("StartContainer failed: %v", err)
	}
	if err := cm.StopContainer(id, 2*time.Second); err != nil {
		t.Errorf("StopContainer failed: %v", err)
	}
	if err := cm.RemoveContainer(id, true); err != nil {
		t.Errorf("RemoveContainer failed: %v", err)
	}
	res, err := cm.Exec(id, []string{"echo", "hi"}, 1*time.Second)
	if err != nil || res.ExitCode != 0 {
		t.Errorf("Exec failed: %v, result=%+v", err, res)
	}
}
