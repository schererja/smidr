package main

import (
	"context"
	"log/slog"

	"github.com/intrik8-labs/smidr/internal/cli"
	"github.com/intrik8-labs/smidr/internal/logging"
)

func main() {
	logging.Init(logging.Config{
		Service:   "smidr-agent", // Binary name
		Component: "executor",    // Component within binary
		Level:     logging.LevelInfo,
		JSON:      true, // Always true for production
	})
	ctx := context.Background()

	log := logging.FromContext(ctx)
	err := cli.Execute(log)
	if err != nil {
		log.Error("Error running Smidr", slog.String("error", err.Error()))
	}
}
