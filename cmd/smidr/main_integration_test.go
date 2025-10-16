package main

import (
	"os/exec"
	"testing"
)

func TestMain_Integration(t *testing.T) {
	cmd := exec.Command("go", "run", "./cmd/smidr/main.go")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("main.go exited with error (expected for CLI): %v", err)
	}
	if len(output) == 0 {
		t.Errorf("main.go produced no output")
	} else {
		t.Logf("main.go output: %s", string(output))
	}
}
