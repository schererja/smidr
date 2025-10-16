package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test showBuildLogs error when no artifacts exist
func TestShowBuildLogs_NoArtifacts(t *testing.T) {
	// Point HOME to an empty temp dir
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// No artifacts exist
	err := showBuildLogs("", "", "", false)
	if err == nil || !strings.Contains(err.Error(), "no artifacts") {
		t.Fatalf("expected no artifacts error, got %v", err)
	}
}

// Test JSON mode formatting and customer scoping
func TestShowBuildLogs_JSONAndCustomer(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create artifact structure: ~/.smidr/artifacts/artifact-acme/build-123/myimage-something/build-log.jsonl
	base := filepath.Join(tmpHome, ".smidr", "artifacts", "artifact-acme", "build-123", "myimage-1")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatalf("mkdirs: %v", err)
	}
	// Write a simple JSONL file with two entries
	logPath := filepath.Join(base, "build-log.jsonl")
	f, err := os.Create(logPath)
	if err != nil {
		t.Fatalf("create jsonl: %v", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	_ = enc.Encode(map[string]any{"msg": "hello", "level": "info"})
	_ = enc.Encode(map[string]any{"msg": "world", "level": "info"})
	f.Close()

	// Request logs in JSON mode with customer scoping; should succeed
	if err := showBuildLogs("build-123", "myimage", "acme", true); err != nil {
		t.Fatalf("showBuildLogs json failed: %v", err)
	}
}
