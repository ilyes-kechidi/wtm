package main

import (
	"os"

	"github.com/ilyes-kechidi/wtm/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
