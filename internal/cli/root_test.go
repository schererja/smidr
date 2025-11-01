package cli

import (
	"fmt"
	"testing"

	"github.com/schererja/smidr/pkg/logger"
	"github.com/spf13/cobra"
)

func TestExecuteReturnsError(t *testing.T) {
	log := logger.NewLogger()
	// Save original rootCmd
	orig := rootCmd
	defer func() { rootCmd = orig }()

	// Set up rootCmd to always error
	rootCmd = &cobra.Command{
		Use: "smidr",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("forced error")
		},
	}
	err := Execute(log)
	if err == nil {
		t.Error("Expected error from Execute, got nil")
	}
}

func TestExecuteSuccess(t *testing.T) {
	log := logger.NewLogger()
	// Save original rootCmd
	orig := rootCmd
	defer func() { rootCmd = orig }()

	// Set up rootCmd to succeed
	rootCmd = &cobra.Command{
		Use: "smidr",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	err := Execute(log)
	if err != nil {
		t.Errorf("Expected success from Execute, got error: %v", err)
	}
}
