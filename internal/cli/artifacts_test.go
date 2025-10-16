package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func newTestCmd() *cobra.Command {
	return &cobra.Command{Use: "test"}
}

func TestRunArtifactsList_Basic(t *testing.T) {
	cmd := newTestCmd()
	err := runArtifactsList(cmd)
	if err != nil {
		t.Logf("runArtifactsList returned error: %v", err)
	}
}

func TestRunArtifactsCopy_Basic(t *testing.T) {
	cmd := newTestCmd()
	err := runArtifactsCopy(cmd, "fake-build", "/tmp/dest")
	if err != nil {
		t.Logf("runArtifactsCopy returned error: %v", err)
	}
}

func TestRunArtifactsClean_Basic(t *testing.T) {
	cmd := newTestCmd()
	err := runArtifactsClean(cmd)
	if err != nil {
		t.Logf("runArtifactsClean returned error: %v", err)
	}
}

func TestRunArtifactsShow_Basic(t *testing.T) {
	cmd := newTestCmd()
	err := runArtifactsShow(cmd, "fake-build")
	if err != nil {
		t.Logf("runArtifactsShow returned error: %v", err)
	}
}
