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
	code := cli.Run(os.Args, cli.VersionInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	})
	os.Exit(code)
}
