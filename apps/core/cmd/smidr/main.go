package main

import (
	"github.com/schererja/smidr/internal/cli"
	"github.com/schererja/smidr/pkg/logger"
)

func main() {
	log := logger.NewLogger()

	err := cli.Execute(log)
	if err != nil {
		log.Fatal("Error running Smidr", err)
	}
}
