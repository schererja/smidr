package main

import (
	cli "github.com/schererja/smidr/internal/cli/controlplane"
	"github.com/schererja/smidr/pkg/logger"
)

func main() {
	// config, err := config.Load("config.yaml")
	// if err != nil {
	// 	fmt.Printf("Failed to load configuration: %v\n", err)
	// 	return
	// }
	log := logger.NewLogger()

	err := cli.Execute(log)
	if err != nil {
		log.Fatal("Error running Smidr", err)
	}

}
