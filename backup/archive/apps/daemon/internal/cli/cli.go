package cli

import (
	"github.com/schererja/smidr/internal/cli/root"
	"github.com/schererja/smidr/pkg/logger"
)

// Execute initializes and executes the root command
func Execute(logger *logger.Logger) error {
	return root.Execute(logger)
}
