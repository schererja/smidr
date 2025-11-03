package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/schererja/smidr/internal/daemon"
	"github.com/schererja/smidr/internal/db"
	"github.com/spf13/cobra"
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
  smidr daemon --address localhost:8080`,
	RunE: runDaemon,
}

var (
	daemonAddress string
)

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.Flags().StringVar(&daemonAddress, "address", ":50051", "Address to listen on (e.g., ':50051' or 'localhost:8080')")
}

func runDaemon(cmd *cobra.Command, args []string) error {
	log := GetLogger()
	log.Info("Starting Smidr daemon...")
	log.Info("📡 Listening", slog.String("address", daemonAddress))

	// TODO: Initialize database for build persistence
	// For now, pass nil to maintain backward compatibility
	// Future: add --db-path flag and initialize DB here
	var database *db.DB = nil

	// Create the gRPC server
	server := daemon.NewServer(daemonAddress, log, database)

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		if err := server.Start(); err != nil {
			errCh <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-sigCh:
		log.Info("\nReceived shutdown signal")
		server.Stop()
		return nil
	case err := <-errCh:
		return fmt.Errorf("daemon error: %w", err)
	case <-ctx.Done():
		server.Stop()
		return nil
	}
}
