package bitbake

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/schererja/smidr/internal/config"
)

func minimalConfig() *config.Config {
	return &config.Config{
		Name:        "test",
		Description: "desc",
		Base: config.BaseConfig{
			Machine: "verdin-imx8mp",
			Distro:  "tdx-xwayland",
		},
		Layers: []config.Layer{
			{Name: "meta-a", Path: "./layers/meta-a"},
			{Name: "meta-b", Git: "https://example.com/meta-b", Branch: "main"},
		},
		Build: config.BuildConfig{
			Image:         "core-image-minimal",
			ExtraPackages: nil,
		},
	}
}

func configWithExtras() *config.Config {
	c := minimalConfig()
	c.Build.ExtraPackages = []string{"vim", "htop"}
	return c
}

func TestGenerate_WritesConfFiles(t *testing.T) {
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

	g := NewGenerator(minimalConfig(), filepath.Join(tmp, "build"))
	if err := g.Generate(); err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	// Check root conf directory
	rootConfDir := filepath.Join(tmp, "conf")
	for _, f := range []string{"local.conf", "bblayers.conf"} {
		p := filepath.Join(rootConfDir, f)
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected %s to exist in root conf: %v", p, err)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		if len(b) == 0 {
			t.Fatalf("expected %s to be non-empty", p)
		}
	}
}

func TestGenerate_CustomImageRecipeWhenExtraPackages(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	g := NewGenerator(configWithExtras(), filepath.Join(tmp, "build"))
	if err := g.Generate(); err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	recipe := filepath.Join(tmp, "meta-smidr-custom", "recipes-core", "images", "smidr-custom-image.bb")
	b, err := os.ReadFile(recipe)
	if err != nil {
		t.Fatalf("expected recipe at %s: %v", recipe, err)
	}
	s := string(b)
	if !strings.Contains(s, "IMAGE_INSTALL += \"") {
		t.Fatalf("expected IMAGE_INSTALL block in recipe, got: %s", s)
	}
}

func TestGetBitBakeCommand(t *testing.T) {
	t.Parallel()
	g1 := NewGenerator(minimalConfig(), ".")
	if cmd := g1.GetBitBakeCommand(); cmd != "bitbake core-image-minimal" {
		t.Fatalf("unexpected cmd: %q", cmd)
	}

	g2 := NewGenerator(configWithExtras(), ".")
	if cmd := g2.GetBitBakeCommand(); cmd != "bitbake smidr-custom-image" {
		t.Fatalf("unexpected cmd with extras: %q", cmd)
	}
}

func TestGenerateLocalConf_AdvancedOverrides(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	cfg := minimalConfig()
	// Ensure directories set so generator writes to predictable paths
	cfg.Directories.Downloads = filepath.Join(tmp, "downloads")
	cfg.Directories.SState = filepath.Join(tmp, "sstate-cache")
	// Set Advanced overrides
	cfg.Advanced.SStateMirrors = "file://.* file:///home/builder/sstate-cache/PATH"
	cfg.Advanced.PreMirrors = "git://.*/.* http://premirror.example.com/\nhttps?://.*/.* http://premirror.example.com/"
	cfg.Advanced.NoNetwork = true
	cfg.Advanced.FetchPremirrorOnly = true

	g := NewGenerator(cfg, filepath.Join(tmp, "build"))
	if err := g.Generate(); err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	// Read generated local.conf
	b, err := os.ReadFile(filepath.Join(tmp, "conf", "local.conf"))
	if err != nil {
		t.Fatalf("read local.conf: %v", err)
	}
	s := string(b)

	if !strings.Contains(s, "SSTATE_MIRRORS = \"file://.* file:///home/builder/sstate-cache/PATH\"") {
		t.Fatalf("expected SSTATE_MIRRORS override to be present, got:\n%s", s)
	}
	if !strings.Contains(s, "PREMIRRORS = \"") {
		t.Fatalf("expected PREMIRRORS to be present, got:\n%s", s)
	}
	if !strings.Contains(s, "BB_NO_NETWORK = \"1\"") {
		t.Fatalf("expected BB_NO_NETWORK to be present, got:\n%s", s)
	}
	if !strings.Contains(s, "BB_FETCH_PREMIRRORONLY = \"1\"") {
		t.Fatalf("expected BB_FETCH_PREMIRRORONLY to be present, got:\n%s", s)
	}

	// When SSTATE_MIRRORS is set, SSTATE_DIR may still be present earlier in file for diskmon,
	// but we ensure mirrors line exists which is what we care about.
}
