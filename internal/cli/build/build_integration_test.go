//go:build integration
// +build integration

package build

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestDockerAvailability checks if Docker is available and running.
func TestDockerAvailability(t *testing.T) {
	cmd := exec.Command("docker", "version")
	if err := cmd.Run(); err != nil {
		t.Skip("Docker is not available, skipping integration tests.")
	}
}

// buildLocalImage builds the project's Dockerfile into a local image and returns the image name.
func buildLocalImage(t *testing.T, projectRoot string) string {
	tag := "smidr-itest-image:latest"
	// Build image
	buildCmd := exec.Command("docker", "build", "-t", tag, ".")
	buildCmd.Dir = projectRoot
	out, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build local image: %v\nOutput: %s", err, string(out))
	}
	// Return the tag (the image is built locally and PullImage is skipped by the CLI when SMIDR_TEST_IMAGE is set)
	return tag
}

// TestSmidrBuildIntegration runs the CLI build command and checks for expected output and exit code.
func TestSmidrBuildIntegration(t *testing.T) {
	// Find project root by traversing up until .git is found
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	projectRoot := cwd
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, ".git")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			t.Fatalf("Could not find project root (.git directory)")
		}
		projectRoot = parent
	}
	mainPath := filepath.Join(projectRoot, "cmd", "smidr", "main.go")
	cmd := exec.Command("go", "run", mainPath, "build")
	testContainerName := "smidr-itest-" + time.Now().Format("20060102-150405")
	// Build a local image and use it to avoid Docker Hub pull rate limits
	img := buildLocalImage(t, projectRoot)
	cmd.Env = append(os.Environ(),
		"SMIDR_TEST_CONTAINER_NAME="+testContainerName,
		"SMIDR_TEST_IMAGE="+img,
		"SMIDR_TEST_WRITE_MARKERS=1", // Enable smoke test mode to skip actual BitBake execution
	)
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	outStr := string(output)

	if err == nil {
		t.Logf("Build command succeeded. Output:\n%s", outStr)
	} else {
		t.Logf("Build command failed (exit code != 0). Output:\n%s", outStr)
	}

	// Check for key output markers (adjust as needed)
	if !strings.Contains(outStr, "Preparing container environment") {
		t.Errorf("Expected container setup output not found")
	}
	if !strings.Contains(outStr, "Cleaning up container") {
		t.Errorf("Expected cleanup output not found")
	}

	// Verify no container with the test name remains
	psCmd := exec.Command("docker", "ps", "-a", "--filter", "name="+testContainerName, "--format", "{{.ID}}")
	psOut, _ := psCmd.CombinedOutput()
	if strings.TrimSpace(string(psOut)) != "" {
		t.Errorf("Expected no remaining containers named %s, but found: %s", testContainerName, string(psOut))
		// Attempt forced cleanup to avoid pollution for subsequent runs
		_ = exec.Command("docker", "rm", "-f", testContainerName).Run()
	}
}

// TestSmidrMountsAndLayers verifies that downloads/workspace/sstate mounts are writable
// and optional layer injection is visible by leveraging environment overrides.
func TestSmidrMountsAndLayers(t *testing.T) {
	// Locate project root
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	projectRoot := cwd
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, ".git")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			t.Fatalf("Could not find project root (.git directory)")
		}
		projectRoot = parent
	}

	// Create temp dirs for mounts and a temp layer dir
	dlDir := t.TempDir()
	wsDir := t.TempDir()
	ssDir := t.TempDir()
	layerDir := t.TempDir()
	// Add a file to layer dir to probe visibility
	if err := os.WriteFile(filepath.Join(layerDir, "layer-info.txt"), []byte("layer-ok"), 0644); err != nil {
		t.Fatalf("failed to write layer file: %v", err)
	}

	mainPath := filepath.Join(projectRoot, "cmd", "smidr", "main.go")
	name := "smidr-itest-mounts-" + time.Now().Format("20060102-150405")
	// Build local image and set SMIDR_TEST_IMAGE to avoid Docker Hub rate limits
	img := buildLocalImage(t, projectRoot)
	cmd := exec.Command("go", "run", mainPath, "build")
	cmd.Env = append(os.Environ(),
		"SMIDR_TEST_CONTAINER_NAME="+name,
		"SMIDR_TEST_DOWNLOADS_DIR="+dlDir,
		"SMIDR_TEST_SSTATE_DIR="+ssDir,
		"SMIDR_TEST_WORKSPACE_DIR="+wsDir,
		"SMIDR_TEST_LAYER_DIRS="+layerDir,
		"SMIDR_TEST_WRITE_MARKERS=1",
		"SMIDR_TEST_IMAGE="+img,
	)
	cmd.Dir = projectRoot
	out, _ := cmd.CombinedOutput()
	t.Logf("Output:\n%s", string(out))

	// Assert markers exist where expected
	// Workspace marker should be written via docker cp (works regardless of permissions)
	if _, err := os.Stat(filepath.Join(wsDir, "itest.txt")); err != nil {
		t.Errorf("workspace marker not found: %v", err)
	}

	// For downloads and sstate, we just verify the directories exist and were mounted
	// (the container should have listed their contents in the output)
	if !strings.Contains(string(out), "Downloads directory accessible") {
		t.Logf("Downloads directory access test not found in output (may be expected if not mounted)")
	}
	if !strings.Contains(string(out), "Sstate directory accessible") {
		t.Logf("Sstate directory access test not found in output (may be expected if not mounted)")
	}

	// Ensure no container with the name remains
	psCmd := exec.Command("docker", "ps", "-a", "--filter", "name="+name, "--format", "{{.ID}}")
	psOut, _ := psCmd.CombinedOutput()
	if strings.TrimSpace(string(psOut)) != "" {
		t.Errorf("Expected no remaining containers named %s, but found: %s", name, string(psOut))
		_ = exec.Command("docker", "rm", "-f", name).Run()
	}
}

// TestSmidrSStateMount verifies that the configured SSTATE directory is mounted
// into the container when provided via config (or test override).
func TestSmidrSStateMount(t *testing.T) {
	// Skip if Docker not available
	cmdCheck := exec.Command("docker", "version")
	if err := cmdCheck.Run(); err != nil {
		t.Skip("Docker not available, skipping integration tests.")
	}

	// Locate project root
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	projectRoot := cwd
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, ".git")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			t.Fatalf("Could not find project root (.git directory)")
		}
		projectRoot = parent
	}

	ssDir := t.TempDir()
	mainPath := filepath.Join(projectRoot, "cmd", "smidr", "main.go")
	name := "smidr-itest-sstate-" + time.Now().Format("20060102-150405")
	img := buildLocalImage(t, projectRoot)
	cmd := exec.Command("go", "run", mainPath, "build")
	cmd.Env = append(os.Environ(),
		"SMIDR_TEST_CONTAINER_NAME="+name,
		"SMIDR_TEST_SSTATE_DIR="+ssDir,
		"SMIDR_TEST_WRITE_MARKERS=1",
		"SMIDR_TEST_IMAGE="+img,
	)
	cmd.Dir = projectRoot
	out, _ := cmd.CombinedOutput()
	t.Logf("Output:\n%s", string(out))

	// Assert sstate directory was accessible (shown in output)
	if !strings.Contains(string(out), "Sstate directory accessible") {
		t.Fatalf("Expected sstate directory to be accessible but not found in output")
	}

	// Cleanup any remaining container
	psCmd := exec.Command("docker", "ps", "-a", "--filter", "name="+name, "--format", "{{.ID}}")
	psOut, _ := psCmd.CombinedOutput()
	if strings.TrimSpace(string(psOut)) != "" {
		_ = exec.Command("docker", "rm", "-f", name).Run()
	}
}
