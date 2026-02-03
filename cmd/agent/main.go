package main

import (
	cli "github.com/schererja/smidr/internal/cli/agent"
	"github.com/schererja/smidr/pkg/logger"
)

func main() {
	// Application entry point
	log := logger.NewLogger()
	log.Info("Starting agent application")

	err := cli.Execute(log)
	if err != nil {
		log.Fatal("Error running Smidr", err)
	}

}
