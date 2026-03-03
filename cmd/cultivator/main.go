// Package main is the entry point for the cultivator binary.
package main

import (
	"os"

	"github.com/Ops-Talks/cultivator/internal/cli"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	os.Exit(run())
}

func run() int {
	return cli.Run(os.Args, cli.VersionInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	})
}
