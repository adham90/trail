package main

import (
	"os"

	"github.com/adham90/trail/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
