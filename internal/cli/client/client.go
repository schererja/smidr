package client

import (
	"github.com/spf13/cobra"
)

var (
	clientDaemonAddress string
)

// New creates and returns the client command with all subcommands
func New() *cobra.Command {
	clientCmd := &cobra.Command{
		Use:   "client",
		Short: "Interact with a running Smidr daemon",
		Long: `The client commands allow you to interact with a running Smidr daemon.

You can start builds, monitor their status, stream logs, and manage artifacts.

Examples:
  smidr client start --config config.yaml --target core-image-minimal
  smidr client status --build-id build-123
  smidr client logs --build-id build-123 --follow
  smidr client list
  smidr client cancel --build-id build-123`,
	}

	// Global flag for all client commands
	clientCmd.PersistentFlags().StringVar(&clientDaemonAddress, "address", "localhost:50051", "Daemon address to connect to")

	// Add subcommands
	clientCmd.AddCommand(clientStartCmd)
	clientCmd.AddCommand(clientStatusCmd)
	clientCmd.AddCommand(clientLogsCmd)
	clientCmd.AddCommand(clientCancelCmd)
	clientCmd.AddCommand(clientListCmd)

	return clientCmd
}
