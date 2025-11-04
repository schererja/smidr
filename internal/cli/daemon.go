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
  smidr daemon --address localhost:8080
  smidr daemon --db-path ~/.smidr/builds.db`,
	RunE: runDaemon,
}

var (
	daemonAddress string
	daemonDBPath  string
)

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.Flags().StringVar(&daemonAddress, "address", ":50051", "Address to listen on (e.g., ':50051' or 'localhost:8080')")
	daemonCmd.Flags().StringVar(&daemonDBPath, "db-path", "", "Path to SQLite database for build persistence (e.g., ~/.smidr/builds.db). If not set, builds are not persisted.")
}

func runDaemon(cmd *cobra.Command, args []string) error {
	log := GetLogger()
	log.Info("Starting Smidr daemon...")
	log.Info("📡 Listening", slog.String("address", daemonAddress))

	// Initialize database if --db-path is provided
	var database *db.DB
	if daemonDBPath != "" {
		log.Info("Enabling build persistence", slog.String("db_path", daemonDBPath))

		var err error
		database, err = db.Open(daemonDBPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer database.Close()

		log.Info("Database initialized successfully")

		// TODO: Add recovery logic here to handle stale builds
		// staleBuilds, err := database.ListStaleBuilds(time.Hour * 24)
		// if err == nil && len(staleBuilds) > 0 {
		//     log.Warn("Found stale builds from previous daemon run", slog.Int("count", len(staleBuilds)))
		//     // Mark them as failed or handle recovery
		// }
	} else {
		log.Info("Build persistence disabled (no --db-path provided)")
		database = nil
	}

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
