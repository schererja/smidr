package config

import (
    "os"
    "path/filepath"
    "testing"
)

func TestLoad_ValidYAML(t *testing.T) {
    t.Parallel()

    tmpDir := t.TempDir()
    cfgPath := filepath.Join(tmpDir, "smidr.yaml")
    yaml := []byte(`
name: test-project
description: "Test description"
base:
  provider: toradex
  machine: verdin-imx8mp
  distro: tdx-xwayland
  version: "6.0.0"
layers:
  - name: meta-toradex-bsp-common
    git: https://git.toradex.com/meta-toradex-bsp-common
    branch: kirkstone-6.x.y
build:
  image: core-image-minimal
  machine: verdin-imx8mp
  extra_packages:
    - python3
container:
  base_image: "smidr/yocto-builder:kirkstone"
`)

    if err := os.WriteFile(cfgPath, yaml, 0644); err != nil {
        t.Fatalf("failed to write temp config: %v", err)
    }

    cfg, err := Load(cfgPath)
    if err != nil {
        t.Fatalf("Load returned error: %v", err)
    }

    if cfg.Name != "test-project" {
        t.Errorf("unexpected Name: %q", cfg.Name)
    }
    if cfg.Base.Machine != "verdin-imx8mp" {
        t.Errorf("unexpected Base.Machine: %q", cfg.Base.Machine)
    }
    if len(cfg.Layers) != 1 {
        t.Fatalf("expected 1 layer, got %d", len(cfg.Layers))
    }
    if cfg.Layers[0].Name != "meta-toradex-bsp-common" {
        t.Errorf("unexpected layer name: %q", cfg.Layers[0].Name)
    }
    if cfg.Build.Image != "core-image-minimal" {
        t.Errorf("unexpected Build.Image: %q", cfg.Build.Image)
    }
}

func TestLoad_FileNotFound(t *testing.T) {
    t.Parallel()
    _, err := Load("/non/existent/path/smidr.yaml")
    if err == nil {
        t.Fatalf("expected error for missing file, got nil")
    }
}

func TestLoad_InvalidYAML(t *testing.T) {
    t.Parallel()
    tmpDir := t.TempDir()
    cfgPath := filepath.Join(tmpDir, "smidr.yaml")
    bad := []byte("name: [unclosed")
    if err := os.WriteFile(cfgPath, bad, 0644); err != nil {
        t.Fatalf("failed to write temp config: %v", err)
    }
    if _, err := Load(cfgPath); err == nil {
        t.Fatalf("expected YAML unmarshal error, got nil")
    }
}


