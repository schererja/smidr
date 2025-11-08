package cli

import (
	"testing"
)

func TestShowBuildLogs_EmptyArgs(t *testing.T) {
	err := showBuildLogs("", "", "", false)
	if err != nil {
		t.Logf("showBuildLogs returned error: %v", err)
	}
}

func TestShowBuildLogs_JSONMode(t *testing.T) {
	err := showBuildLogs("", "", "", true)
	if err != nil {
		t.Logf("showBuildLogs (json mode) returned error: %v", err)
	}
}
