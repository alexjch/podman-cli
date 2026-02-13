package main

import (
	"fmt"
	"os"

	"github.com/alexjch/podman-cli/internal/cli"
)

func main() {
	remoteCLI, err := cli.NewRemoteCLI(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to initialize CLI:", err)
		os.Exit(1)
	}
	os.Exit(remoteCLI.Run())
}
