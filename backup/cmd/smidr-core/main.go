package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/intrik8-labs/smidr/internal/core/config"
	"github.com/intrik8-labs/smidr/internal/core/registry"
	"github.com/intrik8-labs/smidr/internal/core/server"
	"github.com/intrik8-labs/smidr/internal/logging"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "config.yaml", "path to configuration file")
	flag.Parse()

	// Initialize logging
	logging.Init(logging.Config{
		Service:   "smidr-core",
		Component: "server",
		Level:     logging.LevelDebug,
		JSON:      false,
		Pretty:    true,
		AddSource: false,
	})
	ctx := context.Background()
	log := logging.FromContext(ctx)

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Error("Failed to load config", logging.Err(err))
		os.Exit(1)
	}

	log.Info("Configuration loaded",
		"agent_addr", cfg.AgentServer.Address(),
		"web_addr", cfg.WebServer.Address(),
	)

	// Create shared store based on configuration
	storeConfig := registry.Config{
		Type:       cfg.Store.Type,
		SQLitePath: cfg.Store.SQLitePath,
	}
	store, err := registry.NewStore(ctx, storeConfig)
	if err != nil {
		log.Error("Failed to create store", logging.Err(err))
		os.Exit(1)
	}
	defer store.Close(ctx)

	log.Info("Store initialized", "type", cfg.Store.Type)

	// Create both servers
	agentServer := server.NewAgentServer(cfg.AgentServer.Address(), store, log)
	webServer := server.NewWebServer(cfg.WebServer.Address(), store, log)

	// Start both servers in goroutines
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := agentServer.Start(); err != nil {
			errChan <- fmt.Errorf("agent server error: %w", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := webServer.Start(); err != nil {
			errChan <- fmt.Errorf("web server error: %w", err)
		}
	}()

	// Wait for interrupt signal or server errors
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		log.Info("Shutdown signal received")
	case err := <-errChan:
		log.Error("Server error", logging.Err(err))
	}

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Info("Gracefully stopping servers...")

	// Stop both servers
	if err := agentServer.Stop(shutdownCtx); err != nil {
		log.Error("Error stopping agent server", logging.Err(err))
	}
	if err := webServer.Stop(shutdownCtx); err != nil {
		log.Error("Error stopping web server", logging.Err(err))
	}

	log.Info("Server shutdown complete")
}
