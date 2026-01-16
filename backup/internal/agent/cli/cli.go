package cli

import (
	"github.com/intrik8-labs/smidr/internal/agent/cli/root"
	"github.com/intrik8-labs/smidr/internal/logging"
)

// Execute initializes and executes the root command
func Execute(logger *logging.Logger) error {
	return root.Execute(logger)
}
