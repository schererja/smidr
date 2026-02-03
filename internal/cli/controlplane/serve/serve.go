package serve

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	config "github.com/schererja/smidr/internal/config/controlplane"
	"github.com/schererja/smidr/internal/controlplane"
	"github.com/schererja/smidr/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	daemonAddress string
	daemonDBPath  string
	log           *logger.Logger
	svrConfig     *config.ServerConfig
)

var daemonCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Smidr gRPC  server",
	Long: `Start the Smidr server to accept remote build requests via gRPC.

The server exposes a gRPC API that allows clients to:
- Start and monitor builds
- Stream build logs in real-time
- List and manage artifacts
- Cancel running builds

Example usage:
  smidr serve --address :50051
  smidr serve --address localhost:8080
  smidr serve --db-path ~/.smidr/builds.db`,
	RunE: runDaemon,
}

// New returns the daemon command for registration with the root command
func New(logger *logger.Logger, serverConfig *config.ServerConfig) *cobra.Command {
	log = logger
	svrConfig = serverConfig
	daemonCmd.Flags().StringVar(&daemonAddress, "address", ":50051", "Address to listen on (e.g., ':50051' or 'localhost:8080')")
	daemonCmd.Flags().StringVar(&daemonDBPath, "db-path", "", "Path to SQLite database for build persistence (e.g., ~/.smidr/builds.db). If not set, builds are not persisted.")
	return daemonCmd
}

func runDaemon(cmd *cobra.Command, args []string) error {
	// Print a direct console line to guarantee visible startup feedback even if logging is misconfigured
	// fmt.Printf("Smidr daemon starting on %s\n", daemonAddress)
	// if daemonDBPath != "" {
	// 	fmt.Printf("Using database: %s\n", daemonDBPath)
	// }

	log.Info("Starting Smidr daemon... Listening", slog.String("address", svrConfig.Address), slog.String("port", svrConfig.Port))

	// // Initialize database if --db-path is provided
	// var database *db.DB
	// if daemonDBPath != "" {
	// 	log.Info("Enabling build persistence", slog.String("db_path", daemonDBPath))

	// 	var err error
	// 	database, err = db.Open(daemonDBPath)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to open database: %w", err)
	// 	}
	// 	defer database.Close()

	// 	// Log resolved DB path and a quick count for observability
	// 	// This helps detect mismatched paths or empty databases on restart
	// 	if database != nil {
	// 		builds, berr := database.ListBuilds("", true, 0)
	// 		if berr != nil {
	// 			log.Warn("Database opened but list builds failed", slog.String("error", berr.Error()), slog.String("db_path", database.Path()))
	// 		} else {
	// 			log.Info("Database initialized successfully", slog.String("resolved_db_path", database.Path()), slog.Int("existing_builds", len(builds)))
	// 		}
	// 	} else {
	// 		log.Info("Database initialized successfully")
	// 	}

	// 	// TODO: Add recovery logic here to handle stale builds
	// 	// staleBuilds, err := database.ListStaleBuilds(time.Hour * 24)
	// 	// if err == nil && len(staleBuilds) > 0 {
	// 	//     log.Warn("Found stale builds from previous daemon run", slog.Int("count", len(staleBuilds)))
	// 	//     // Mark them as failed or handle recovery
	// 	// }
	// } else {
	// 	log.Info("Build persistence disabled (no --db-path provided)")
	// 	database = nil
	// }

	cps, err := controlplane.NewControlPlaneServer(svrConfig, log)
	if err != nil {
		return err
	}
	// Create the gRPC server
	// server := daemonpkg.NewServer(daemonAddress, log, database)

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		if err := cps.Start(); err != nil {
			errCh <- err
		}
	}()

	// // Wait for shutdown signal or error
	select {
	case <-sigCh:
		log.Info("\nReceived shutdown signal")
		cps.Stop()
		return nil
	case err := <-errCh:
		return fmt.Errorf("daemon error: %w", err)
	case <-ctx.Done():
		cps.Stop()
		return nil
	}
}
