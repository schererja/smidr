package cli

import (
	"github.com/spf13/cobra"
)

var (
	clientDaemonAddress string
)

var clientCmd = &cobra.Command{
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

func init() {
	rootCmd.AddCommand(clientCmd)
	
	// Global flag for all client commands
	clientCmd.PersistentFlags().StringVar(&clientDaemonAddress, "address", "localhost:50051", "Daemon address to connect to")
}
