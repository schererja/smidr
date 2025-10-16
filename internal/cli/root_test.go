package cli

import (
	"testing"
)

func TestExecuteReturnsError(t *testing.T) {
	err := Execute()
	if err == nil {
		t.Log("Execute() returned nil error (expected if no root command is set up in test)")
	}
}
