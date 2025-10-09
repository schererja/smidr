package docker

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/intrik8-labs/smidr/internal/container"
)

// Note: These tests require Docker to be running and accessible.
// They are basic smoke tests for the DockerManager implementation.

func TestDockerManager_PullCreateStartStopRemove(t *testing.T) {
	// Check Docker availability
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("Docker not available or not running")
	}

	ctx := context.Background()
	dm, err := NewDockerManager()
	if err != nil {
		t.Fatalf("Failed to create DockerManager: %v", err)
	}
	image := "busybox:latest"
	containerName := "smidr-test-container-" + strings.ReplaceAll(time.Now().Format("150405.000"), ".", "-")
	var id string

	// Ensure cleanup even if test fails
	t.Cleanup(func() {
		if id != "" {
			_ = dm.StopContainer(ctx, id, 2*time.Second)
			_ = dm.RemoveContainer(ctx, id, true)
		}
	})

	// Pull image
	if err := dm.PullImage(ctx, image); err != nil {
		t.Fatalf("PullImage failed: %v", err)
	}

	// Create container
	cfg := container.ContainerConfig{
		Image: image,
		Name:  containerName,
		Cmd:   []string{"sleep", "2"},
	}
	id, err = dm.CreateContainer(ctx, cfg)
	if err != nil {
		t.Fatalf("CreateContainer failed: %v", err)
	}

	// Start container
	if err := dm.StartContainer(ctx, id); err != nil {
		t.Fatalf("StartContainer failed: %v", err)
	}

	// Exec in container
	res, err := dm.Exec(ctx, id, []string{"echo", "hello"}, 2*time.Second)
	if err != nil {
		t.Errorf("Exec failed: %v", err)
	} else if !strings.Contains(string(res.Stdout), "hello") {
		t.Errorf("Exec output missing: got %q", res.Stdout)
	}

	// Stop container
	if err := dm.StopContainer(ctx, id, 2*time.Second); err != nil {
		t.Errorf("StopContainer failed: %v", err)
	}

	// Remove container
	if err := dm.RemoveContainer(ctx, id, true); err != nil {
		t.Errorf("RemoveContainer failed: %v", err)
	}
}

// Optional: Clean up any leftover containers from failed runs
func TestDockerManager_Cleanup(t *testing.T) {
	cmd := exec.Command("docker", "ps", "-a", "--filter", "name=smidr-test-container", "--format", "{{.ID}}")
	out, err := cmd.Output()
	if err != nil {
		t.Skip("docker not available or not running")
	}
	ids := strings.Fields(string(out))
	for _, id := range ids {
		exec.Command("docker", "rm", "-f", id).Run()
	}
}

func TestDockerManager_VolumeMounts(t *testing.T) {
	// Check Docker availability
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("Docker not available or not running")
	}

	ctx := context.Background()
	dm, err := NewDockerManager()
	if err != nil {
		t.Fatalf("Failed to create DockerManager: %v", err)
	}

	downloadsDir := t.TempDir()
	sstateDir := t.TempDir()

	config := container.ContainerConfig{
		Image:          "busybox:latest",
		Cmd:            []string{"echo testfile > /home/builder/downloads/hello.txt && echo sstate > /home/builder/sstate-cache/ss.txt && sleep 1"},
		DownloadsDir:   downloadsDir,
		SstateCacheDir: sstateDir,
	}

	// Pull image first
	if err := dm.PullImage(ctx, config.Image); err != nil {
		t.Fatalf("PullImage failed: %v", err)
	}

	id, err := dm.CreateContainer(ctx, config)
	if err != nil {
		t.Fatalf("CreateContainer failed: %v", err)
	}
	defer func() {
		_ = dm.StopContainer(ctx, id, 2*time.Second)
		_ = dm.RemoveContainer(ctx, id, true)
	}()

	if err := dm.StartContainer(ctx, id); err != nil {
		t.Fatalf("StartContainer failed: %v", err)
	}

	// Wait for the container to finish writing files
	time.Sleep(3 * time.Second)

	// Check that the files exist on the host
	helloFile := downloadsDir + "/hello.txt"
	if _, err := exec.Command("ls", helloFile).Output(); err != nil {
		t.Errorf("File not found in downloads mount: %s", helloFile)
	}

	sstateFile := sstateDir + "/ss.txt"
	if _, err := exec.Command("ls", sstateFile).Output(); err != nil {
		t.Errorf("File not found in sstate-cache mount: %s", sstateFile)
	}
}

func TestDockerManager_LayerInjection(t *testing.T) {
	// Check Docker availability
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("Docker not available or not running")
	}

	ctx := context.Background()
	dm, err := NewDockerManager()
	if err != nil {
		t.Fatalf("Failed to create DockerManager: %v", err)
	}

	// Create temporary layer directories with sample content
	layer1Dir := t.TempDir()
	layer2Dir := t.TempDir()
	workspaceDir := t.TempDir()

	// Create sample layer content
	layer1File := layer1Dir + "/conf.conf"
	layer2File := layer2Dir + "/recipe.bb"
	if err := exec.Command("sh", "-c", "echo 'LAYER_NAME = \"test-layer-1\"' > "+layer1File).Run(); err != nil {
		t.Fatalf("Failed to create layer1 content: %v", err)
	}
	if err := exec.Command("sh", "-c", "echo 'DESCRIPTION = \"Test recipe\"' > "+layer2File).Run(); err != nil {
		t.Fatalf("Failed to create layer2 content: %v", err)
	}

	config := container.ContainerConfig{
		Image:        "busybox:latest",
		Cmd:          []string{"ls -la /home/builder/layers/layer-0/ && ls -la /home/builder/layers/layer-1/ && sleep 1"},
		WorkspaceDir: workspaceDir,
		LayerDirs:    []string{layer1Dir, layer2Dir},
	}

	// Pull image first
	if err := dm.PullImage(ctx, config.Image); err != nil {
		t.Fatalf("PullImage failed: %v", err)
	}

	id, err := dm.CreateContainer(ctx, config)
	if err != nil {
		t.Fatalf("CreateContainer failed: %v", err)
	}
	defer func() {
		_ = dm.StopContainer(ctx, id, 2*time.Second)
		_ = dm.RemoveContainer(ctx, id, true)
	}()

	if err := dm.StartContainer(ctx, id); err != nil {
		t.Fatalf("StartContainer failed: %v", err)
	}

	// Verify layers are accessible in container
	res, err := dm.Exec(ctx, id, []string{"ls", "/home/builder/layers/layer-0/conf.conf"}, 5*time.Second)
	if err != nil {
		t.Errorf("Failed to access layer-0 in container: %v", err)
	}
	if res.ExitCode != 0 {
		t.Errorf("Layer-0 file not found in container, exit code: %d", res.ExitCode)
	}

	res, err = dm.Exec(ctx, id, []string{"ls", "/home/builder/layers/layer-1/recipe.bb"}, 5*time.Second)
	if err != nil {
		t.Errorf("Failed to access layer-1 in container: %v", err)
	}
	if res.ExitCode != 0 {
		t.Errorf("Layer-1 file not found in container, exit code: %d", res.ExitCode)
	}
}
