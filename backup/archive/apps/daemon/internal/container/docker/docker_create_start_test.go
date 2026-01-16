package docker

import (
	"context"
	"testing"
	"time"

	smidrContainer "github.com/schererja/smidr/internal/container"
	"github.com/schererja/smidr/pkg/logger"
)

// TestCreateStartContainer exercises CreateContainer and StartContainer success paths.
// Skips when Docker is unavailable in the environment.
func TestCreateStartContainer(t *testing.T) {
	log := logger.NewLogger()
	dm, err := NewDockerManager(log)
	if err != nil {
		t.Skipf("docker not available: %v", err)
	}
	ctx := context.Background()

	// Ensure busybox image is present
	const image = "busybox:latest"
	if !dm.ImageExists(ctx, image) {
		if err := dm.PullImage(ctx, image); err != nil {
			t.Skipf("could not pull busybox: %v", err)
		}
	}

	cfg := smidrContainer.NewContainerConfig(image, "").
		WithCmd("sh", "-c", "sleep 1").
		WithMemoryLimit("64m").
		WithCPUCount(1)

	id, err := dm.CreateContainer(ctx, cfg)
	if err != nil {
		t.Fatalf("CreateContainer error: %v", err)
	}
	defer func() {
		_ = dm.StopContainer(ctx, id, 2*time.Second)
		_ = dm.RemoveContainer(ctx, id, true)
	}()

	if err := dm.StartContainer(ctx, id); err != nil {
		t.Fatalf("StartContainer error: %v", err)
	}

	// Exec a quick command while container is running
	res, err := dm.Exec(ctx, id, []string{"sh", "-c", "echo ok"}, 5*time.Second)
	if err != nil {
		t.Fatalf("Exec error: %v", err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("Exec exit code = %d, want 0 (stderr=%s)", res.ExitCode, string(res.Stderr))
	}
}
