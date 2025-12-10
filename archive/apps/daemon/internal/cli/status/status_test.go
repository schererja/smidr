package status

import (
	"testing"
)

func TestShowBuildStatus_EmptyArgs(t *testing.T) {
	err := showBuildStatus("", "", "", false)
	if err != nil {
		t.Logf("showBuildStatus returned error: %v", err)
	}
}

func TestShowBuildStatus_ListArtifacts(t *testing.T) {
	err := showBuildStatus("", "", "", true)
	if err != nil {
		t.Logf("showBuildStatus (listArtifacts) returned error: %v", err)
	}
}
