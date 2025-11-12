package init

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateConfigTemplate_IncludesProjectName(t *testing.T) {
	t.Parallel()
	out := generateConfigTemplate("myproj")
	if !strings.Contains(out, "myproj") {
		t.Fatalf("template should include project name, got: %s", out)
	}
	if !strings.Contains(out, "# Smidr Build Configuration") {
		t.Fatalf("template header missing")
	}
}

func TestInitProject_WritesFileAndPreventsOverwrite(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldwd) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// First call should create smidr.yaml
	if err := initProject("proj"); err != nil {
		t.Fatalf("initProject returned error: %v", err)
	}
	p := filepath.Join(tmp, "smidr.yaml")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("expected smidr.yaml: %v", err)
	}
	if len(b) == 0 {
		t.Fatalf("smidr.yaml must not be empty")
	}

	// Second call should fail due to existing file
	if err := initProject("proj"); err == nil {
		t.Fatalf("expected error on overwrite, got nil")
	}
}
