package main

import (
	"context"
	"log/slog"

	"github.com/intrik8-labs/smidr/internal/agent/cli"
	"github.com/intrik8-labs/smidr/internal/logging"
)

func main() {
	logging.Init(logging.Config{
		Service:   "smidr-agent", // Binary name
		Component: "executor",    // Component within binary
		Level:     logging.LevelDebug,
		JSON:      true, // Always true for production
		Pretty:    true, // Clean output for debugging
		AddSource: true, // Set to true to see file:line info
	})
	ctx := context.Background()

	log := logging.FromContext(ctx)
	err := cli.Execute(log)
	if err != nil {
		log.Error("Error running Smidr", slog.String("error", err.Error()))
	}
}
