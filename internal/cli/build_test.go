package cli

import (
	"strings"
	"testing"

	config "github.com/intrik8-labs/smidr/internal/config"
)

func TestSetDefaultDirs(t *testing.T) {
	tmp := t.TempDir()
	cfg := &config.Config{}
	setDefaultDirs(cfg, tmp)

	if cfg.Directories.Source == "" {
		t.Fatalf("Source should be defaulted")
	}
	if cfg.Directories.Build == "" {
		t.Fatalf("Build should be defaulted")
	}
	if cfg.Directories.SState == "" {
		t.Fatalf("SState should be defaulted")
	}

	// Ensure paths are under tmp
	if !strings.HasPrefix(cfg.Directories.Source, tmp) {
		t.Fatalf("Source not under tmp: %s", cfg.Directories.Source)
	}
	if !strings.HasPrefix(cfg.Directories.Build, tmp) {
		t.Fatalf("Build not under tmp: %s", cfg.Directories.Build)
	}
	if !strings.HasPrefix(cfg.Directories.SState, tmp) {
		t.Fatalf("SState not under tmp: %s", cfg.Directories.SState)
	}

	// Now set explicit values and ensure they are preserved
	cfg2 := &config.Config{}
	cfg2.Directories.Source = "/custom/sources"
	cfg2.Directories.Build = "/custom/build"
	cfg2.Directories.SState = "/custom/sstate"
	setDefaultDirs(cfg2, tmp)
	if cfg2.Directories.Source != "/custom/sources" {
		t.Fatalf("Source should not be overridden")
	}
	if cfg2.Directories.Build != "/custom/build" {
		t.Fatalf("Build should not be overridden")
	}
	if cfg2.Directories.SState != "/custom/sstate" {
		t.Fatalf("SState should not be overridden")
	}
}
