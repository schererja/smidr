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

func TestConfigValidation_RequiredFields(t *testing.T) {
	t.Parallel()

	// Test missing name
	cfg := &Config{
		Description: "test",
		Base:        BaseConfig{Machine: "test", Distro: "test"},
		Layers:      []Layer{{Name: "test", Git: "https://example.com"}},
		Build:       BuildConfig{Image: "test"},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for missing name")
	}

	// Test missing description
	cfg = &Config{
		Name:   "test",
		Base:   BaseConfig{Machine: "test", Distro: "test"},
		Layers: []Layer{{Name: "test", Git: "https://example.com"}},
		Build:  BuildConfig{Image: "test"},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for missing description")
	}
}

func TestBaseConfigValidation(t *testing.T) {
	t.Parallel()

	// Test missing machine
	base := BaseConfig{Distro: "test"}
	if err := base.Validate(); err == nil {
		t.Fatalf("expected validation error for missing machine")
	}

	// Test missing distro
	base = BaseConfig{Machine: "test"}
	if err := base.Validate(); err == nil {
		t.Fatalf("expected validation error for missing distro")
	}

	// Test invalid machine name
	base = BaseConfig{Machine: "test@invalid", Distro: "test"}
	if err := base.Validate(); err == nil {
		t.Fatalf("expected validation error for invalid machine name")
	}
}

func TestLayerValidation(t *testing.T) {
	t.Parallel()

	// Test missing name
	layer := Layer{Git: "https://example.com"}
	if err := layer.Validate(); err == nil {
		t.Fatalf("expected validation error for missing name")
	}

	// Test neither git nor path
	layer = Layer{Name: "test"}
	if err := layer.Validate(); err == nil {
		t.Fatalf("expected validation error for missing git/path")
	}

	// Test both git and path
	layer = Layer{Name: "test", Git: "https://example.com", Path: "./test"}
	if err := layer.Validate(); err == nil {
		t.Fatalf("expected validation error for both git and path")
	}

	// Test invalid git URL
	layer = Layer{Name: "test", Git: "invalid-url"}
	if err := layer.Validate(); err == nil {
		t.Fatalf("expected validation error for invalid git URL")
	}
}

func TestBuildConfigValidation(t *testing.T) {
	t.Parallel()

	// Test missing image
	build := BuildConfig{}
	if err := build.Validate(); err == nil {
		t.Fatalf("expected validation error for missing image")
	}

	// Test negative parallel_make
	build = BuildConfig{Image: "test", ParallelMake: -1}
	if err := build.Validate(); err == nil {
		t.Fatalf("expected validation error for negative parallel_make")
	}

	// Test invalid package name
	build = BuildConfig{Image: "test", ExtraPackages: []string{"invalid@package"}}
	if err := build.Validate(); err == nil {
		t.Fatalf("expected validation error for invalid package name")
	}
}

func TestPackageConfigValidation(t *testing.T) {
	t.Parallel()

	// Test invalid package class
	pkg := PackageConfig{Classes: "invalid_class"}
	if err := pkg.Validate(); err == nil {
		t.Fatalf("expected validation error for invalid package class")
	}

	// Test valid package class
	pkg = PackageConfig{Classes: "package_rpm"}
	if err := pkg.Validate(); err != nil {
		t.Fatalf("unexpected validation error for valid package class: %v", err)
	}
}

func TestFeatureConfigValidation(t *testing.T) {
	t.Parallel()

	// Test invalid extra image feature
	features := FeatureConfig{ExtraImageFeatures: []string{"invalid-feature"}}
	if err := features.Validate(); err == nil {
		t.Fatalf("expected validation error for invalid extra image feature")
	}

	// Test invalid user class
	features = FeatureConfig{UserClasses: []string{"invalid-class"}}
	if err := features.Validate(); err == nil {
		t.Fatalf("expected validation error for invalid user class")
	}

	// Test invalid inherit class
	features = FeatureConfig{InheritClasses: []string{"invalid-inherit"}}
	if err := features.Validate(); err == nil {
		t.Fatalf("expected validation error for invalid inherit class")
	}
}

func TestAdvancedConfigValidation(t *testing.T) {
	t.Parallel()

	// Test invalid BB_HASHSERVE
	advanced := AdvancedConfig{BBHashServe: "invalid"}
	if err := advanced.Validate(); err == nil {
		t.Fatalf("expected validation error for invalid BB_HASHSERVE")
	}

	// Test invalid BB_SIGNATURE_HANDLER
	advanced = AdvancedConfig{BBSignatureHandler: "InvalidHandler"}
	if err := advanced.Validate(); err == nil {
		t.Fatalf("expected validation error for invalid BB_SIGNATURE_HANDLER")
	}

	// Test invalid CONF_VERSION
	advanced = AdvancedConfig{ConfVersion: "invalid-version"}
	if err := advanced.Validate(); err == nil {
		t.Fatalf("expected validation error for invalid CONF_VERSION")
	}
}

func TestContainerConfigValidation(t *testing.T) {
	t.Parallel()

	// Test invalid memory format
	container := ContainerConfig{Memory: "invalid"}
	if err := container.Validate(); err == nil {
		t.Fatalf("expected validation error for invalid memory format")
	}

	// Test negative CPU count
	container = ContainerConfig{CPUCount: -1}
	if err := container.Validate(); err == nil {
		t.Fatalf("expected validation error for negative CPU count")
	}
}

func TestCacheConfigValidation(t *testing.T) {
	t.Parallel()

	// Test negative retention
	cache := CacheConfig{Retention: -1}
	if err := cache.Validate(); err == nil {
		t.Fatalf("expected validation error for negative retention")
	}
}

func TestEnvironmentVariableSubstitution(t *testing.T) {
	t.Parallel()

	// Set up test environment variables
	os.Setenv("TEST_MACHINE", "verdin-imx8mp")
	os.Setenv("TEST_DISTRO", "tdx-xwayland")
	os.Setenv("TEST_IMAGE", "core-image-weston")
	os.Setenv("TEST_THREADS", "8")
	os.Setenv("TEST_MEMORY", "16g")
	os.Setenv("EMPTY_VAR", "")
	defer func() {
		os.Unsetenv("TEST_MACHINE")
		os.Unsetenv("TEST_DISTRO")
		os.Unsetenv("TEST_IMAGE")
		os.Unsetenv("TEST_THREADS")
		os.Unsetenv("TEST_MEMORY")
		os.Unsetenv("EMPTY_VAR")
	}()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "smidr.yaml")

	// Test basic substitution
	yaml := []byte(`
name: test-project
description: "Test with env vars"
base:
  machine: ${TEST_MACHINE}
  distro: ${TEST_DISTRO}
layers:
  - name: meta-test
    git: https://example.com/meta-test
build:
  image: ${TEST_IMAGE}
  bb_number_threads: ${TEST_THREADS}
container:
  memory: ${TEST_MEMORY}
`)

	if err := os.WriteFile(cfgPath, yaml, 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Base.Machine != "verdin-imx8mp" {
		t.Errorf("expected machine 'verdin-imx8mp', got '%s'", cfg.Base.Machine)
	}
	if cfg.Base.Distro != "tdx-xwayland" {
		t.Errorf("expected distro 'tdx-xwayland', got '%s'", cfg.Base.Distro)
	}
	if cfg.Build.Image != "core-image-weston" {
		t.Errorf("expected image 'core-image-weston', got '%s'", cfg.Build.Image)
	}
	if cfg.Container.Memory != "16g" {
		t.Errorf("expected memory '16g', got '%s'", cfg.Container.Memory)
	}
}

func TestEnvironmentVariableSubstitutionWithDefaults(t *testing.T) {
	t.Parallel()

	// Set up test environment variables
	os.Setenv("SET_VAR", "set-value")
	os.Setenv("EMPTY_VAR", "")
	defer func() {
		os.Unsetenv("SET_VAR")
		os.Unsetenv("EMPTY_VAR")
	}()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "smidr.yaml")

	// Test substitution with defaults and alternatives
	yaml := []byte(`
name: test-project
description: "Test with env var defaults"
base:
  machine: ${UNSET_VAR:-qemux86-64}
  distro: ${EMPTY_VAR:-tdx-xwayland}
layers:
  - name: meta-test
    git: https://example.com/meta-test
build:
  image: ${SET_VAR:+core-image-minimal}
  bb_number_threads: 4
`)

	if err := os.WriteFile(cfgPath, yaml, 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	// Test default substitution (${UNSET_VAR:-default})
	if cfg.Base.Machine != "qemux86-64" {
		t.Errorf("expected machine 'qemux86-64' (default), got '%s'", cfg.Base.Machine)
	}

	// Test default substitution with empty var (${EMPTY_VAR:-default})
	if cfg.Base.Distro != "tdx-xwayland" {
		t.Errorf("expected distro 'tdx-xwayland' (default), got '%s'", cfg.Base.Distro)
	}

	// Test alternative substitution (${SET_VAR:+alternative})
	if cfg.Build.Image != "core-image-minimal" {
		t.Errorf("expected image 'core-image-minimal' (alternative), got '%s'", cfg.Build.Image)
	}
}

func TestEnvironmentVariableSubstitutionComplex(t *testing.T) {
	t.Parallel()

	// Set up test environment variables
	os.Setenv("PROJECT_NAME", "my-project")
	os.Setenv("BUILD_THREADS", "16")
	os.Setenv("CACHE_DIR", "/custom/cache")
	defer func() {
		os.Unsetenv("PROJECT_NAME")
		os.Unsetenv("BUILD_THREADS")
		os.Unsetenv("CACHE_DIR")
	}()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "smidr.yaml")

	// Test complex substitution with multiple variables
	yaml := []byte(`
name: ${PROJECT_NAME}
description: "Project ${PROJECT_NAME} with custom settings"
base:
  machine: ${MACHINE_TYPE:-verdin-imx8mp}
  distro: ${DISTRO_TYPE:-tdx-xwayland}
layers:
  - name: meta-${PROJECT_NAME}
    git: https://example.com/meta-${PROJECT_NAME}
build:
  image: core-image-${PROJECT_NAME}
  bb_number_threads: ${BUILD_THREADS}
directories:
  downloads: ${CACHE_DIR}/downloads
  sstate: ${CACHE_DIR}/sstate
cache:
  downloads: ${CACHE_DIR}/downloads
  sstate: ${CACHE_DIR}/sstate
`)

	if err := os.WriteFile(cfgPath, yaml, 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	// Test complex substitutions
	if cfg.Name != "my-project" {
		t.Errorf("expected name 'my-project', got '%s'", cfg.Name)
	}
	if cfg.Description != "Project my-project with custom settings" {
		t.Errorf("expected description with substitution, got '%s'", cfg.Description)
	}
	if cfg.Layers[0].Name != "meta-my-project" {
		t.Errorf("expected layer name 'meta-my-project', got '%s'", cfg.Layers[0].Name)
	}
	if cfg.Layers[0].Git != "https://example.com/meta-my-project" {
		t.Errorf("expected git URL with substitution, got '%s'", cfg.Layers[0].Git)
	}
	if cfg.Build.Image != "core-image-my-project" {
		t.Errorf("expected image 'core-image-my-project', got '%s'", cfg.Build.Image)
	}
	if cfg.Directories.Downloads != "/custom/cache/downloads" {
		t.Errorf("expected downloads path with substitution, got '%s'", cfg.Directories.Downloads)
	}
}
