package main

import (
	"os"

	"github.com/alexjch/podman-cli/internal/cli"
)

func main() {
	c := cli.NewRemoteCLI()
	exitCode := c.Run(os.Args[1:])
	os.Exit(exitCode)
}
