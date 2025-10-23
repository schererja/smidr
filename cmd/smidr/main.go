package main

import (
	"fmt"

	"github.com/schererja/smidr/internal/cli"
)

func main() {
	err := cli.Execute()
	if err != nil {
		fmt.Printf("Error running Smidr: %v\n", err)
	}
}
