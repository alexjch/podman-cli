package main

import (
	"fmt"
	"os"

	"github.com/alexjch/podman-cli/internal/cli"
)

func main() {
	remoteCLI, err := cli.NewRemoteCLI(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize CLI: %v\n", err)
		os.Exit(1)
	}
	exitCode := remoteCLI.Run()
	os.Exit(exitCode)
}
