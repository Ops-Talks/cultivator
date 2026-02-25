package main

import (
	"fmt"
	"os"

	"github.com/cultivator-dev/cultivator/pkg/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := cmd.NewRootCommand(version, commit, date)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
