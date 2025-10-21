package main

import (
	"fmt"

	"github.com/intrik8-labs/smidr/internal/cli"
)

func main() {
	err := cli.Execute()
	if err != nil {
		fmt.Printf("Error running Smidr: %v\n", err)
	}
}
