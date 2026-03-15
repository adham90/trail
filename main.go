package main

import (
	"os"

	"github.com/adhameldeeb/trail/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
