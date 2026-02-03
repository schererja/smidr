package daemon

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/schererja/smidr/internal/agent"
	config "github.com/schererja/smidr/internal/config/agent"
	"github.com/schererja/smidr/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	daemonAddress string
	daemonDBPath  string
	log           *logger.Logger
	agentCfg      *config.AgentConfig
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the Smidr gRPC daemon server",
	Long: `Start the Smidr daemon to accept remote build requests via gRPC.

The daemon exposes a gRPC API that allows clients to:
- Start and monitor builds
- Stream build logs in real-time
- List and manage artifacts
- Cancel running builds

Example usage:
  smidr daemon --address :50051
  smidr daemon --address localhost:8080
  smidr daemon --db-path ~/.smidr/builds.db`,
	RunE: runDaemon,
}

// New returns the daemon command for registration with the root command
func New(logger *logger.Logger, agentConfig *config.AgentConfig) *cobra.Command {
	log = logger
	agentCfg = agentConfig
	daemonCmd.Flags().StringVar(&daemonAddress, "address", ":50051", "Address to listen on (e.g., ':50051' or 'localhost:8080')")
	daemonCmd.Flags().StringVar(&daemonDBPath, "db-path", "", "Path to SQLite database for build persistence (e.g., ~/.smidr/builds.db). If not set, builds are not persisted.")
	return daemonCmd
}

func runDaemon(cmd *cobra.Command, args []string) error {
	log.Info("Starting agent daemon",
		slog.String("listen_address", agentCfg.ListenAddress),
		slog.String("control_plane", agentCfg.ControlPlaneAddress))

	// Create and start the agent
	a := agent.NewAgent(agentCfg)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start agent in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- a.Start()
	}()

	// Wait for shutdown signal or error
	select {
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("agent error: %w", err)
		}
	case sig := <-sigChan:
		log.Info("Received signal", slog.String("signal", sig.String()))
		return a.Stop()
	}
	return nil
}
