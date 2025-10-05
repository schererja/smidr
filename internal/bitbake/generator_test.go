package bitbake

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/intrik8-labs/smidr/internal/config"
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
