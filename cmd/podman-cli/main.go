package main

import (
	"os"

	"github.com/alexjch/podman-cli/internal/cli"
)

func main() {
	remoteCLI := cli.NewRemoteCLI()
	exitCode := remoteCLI.Run(os.Args[1:])
	os.Exit(exitCode)
}
