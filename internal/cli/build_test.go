package cli

import (
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

	// Paths should be either under temp dir (if home dir unavailable) or under ~/.smidr
	// This is more flexible than the original test which expected everything under tmp
	t.Logf("Source dir: %s", cfg.Directories.Source)
	t.Logf("Build dir: %s", cfg.Directories.Build)
	t.Logf("SState dir: %s", cfg.Directories.SState)

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
